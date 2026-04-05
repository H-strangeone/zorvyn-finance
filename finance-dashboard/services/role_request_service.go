package services

import (
	"finance-dashboard/models"
	"finance-dashboard/store"
	"finance-dashboard/utils"
	"time"

	"github.com/google/uuid"
)

type RoleRequestService struct {
	roleRequestStore store.RoleRequestStore
	userStore        store.UserStore
}

func NewRoleRequestService(
	roleRequestStore store.RoleRequestStore,
	userStore store.UserStore,
) *RoleRequestService {
	return &RoleRequestService{
		roleRequestStore: roleRequestStore,
		userStore:        userStore,
	}
}

type CreateRoleRequestInput struct {
	RequestedRole string `json:"requestedRole"`
	Reason        string `json:"reason"`
}

type ProcessRoleRequestInput struct {
	Action     string `json:"action"`     // "approve" or "reject"
	ReviewNote string `json:"reviewNote"` // optional
}

func (s *RoleRequestService) Create(userID string, input CreateRoleRequestInput) (*models.RoleRequest, *utils.AppError) {
	if !utils.IsValidRequestedRole(input.RequestedRole) {
		return nil, utils.NewValidationError("requestedRole must be analyst or admin")
	}

	user := s.userStore.GetByID(userID)
	if user == nil {
		return nil, utils.NewNotFoundError("user")
	}

	if string(user.Role) == input.RequestedRole {
		return nil, utils.NewConflictError("you already have this role")
	}

	if user.Role == models.RoleAdmin {
		return nil, utils.NewForbiddenError("admins cannot submit role requests")
	}

	existing := s.roleRequestStore.GetPendingByUserID(userID)
	if existing != nil {
		return nil, utils.NewConflictError("you already have a pending role request")
	}

	now := time.Now().UTC()
	req := &models.RoleRequest{
		ID:            uuid.New().String(),
		UserID:        userID,
		RequestedRole: models.Role(input.RequestedRole),
		Status:        models.StatusPending,
		Reason:        input.Reason,
		ReviewedBy:    "",
		ReviewNote:    "",
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := s.roleRequestStore.Create(req); err != nil {
		return nil, utils.NewInternalError("failed to create role request")
	}

	return req, nil
}

func (s *RoleRequestService) GetAll(status string) ([]*models.RoleRequest, *utils.AppError) {
	if status != "" {
		validStatuses := map[string]bool{
			"pending":  true,
			"approved": true,
			"rejected": true,
		}
		if !validStatuses[status] {
			return nil, utils.NewValidationError("status must be pending, approved, or rejected")
		}
	}

	return s.roleRequestStore.GetAll(status), nil
}

func (s *RoleRequestService) GetByID(id string) (*models.RoleRequest, *utils.AppError) {
	if !utils.IsValidUUID(id) {
		return nil, utils.NewValidationError("invalid request ID format")
	}

	req := s.roleRequestStore.GetByID(id)
	if req == nil {
		return nil, utils.NewNotFoundError("role request")
	}

	return req, nil
}

func (s *RoleRequestService) GetMine(userID string) ([]*models.RoleRequest, *utils.AppError) {
	return s.roleRequestStore.GetByUserID(userID), nil
}

func (s *RoleRequestService) Process(id string, adminID string, input ProcessRoleRequestInput) (*models.RoleRequest, *utils.AppError) {
	if !utils.IsValidRoleRequestAction(input.Action) {
		return nil, utils.NewValidationError("action must be approve or reject")
	}

	if !utils.IsValidUUID(id) {
		return nil, utils.NewValidationError("invalid request ID format")
	}

	req := s.roleRequestStore.GetByID(id)
	if req == nil {
		return nil, utils.NewNotFoundError("role request")
	}

	if req.Status != models.StatusPending {
		return nil, utils.NewConflictError("role request has already been processed")
	}

	user := s.userStore.GetByID(req.UserID)
	if user == nil {
		return nil, utils.NewNotFoundError("user")
	}

	now := time.Now().UTC()

	// Update request
	if input.Action == "approve" {
		req.Status = models.StatusApproved
	} else {
		req.Status = models.StatusRejected
	}
	req.ReviewedBy = adminID
	req.ReviewNote = input.ReviewNote
	req.UpdatedAt = now

	if err := s.roleRequestStore.Update(req); err != nil {
		return nil, utils.NewInternalError("failed to process role request")
	}

	if input.Action == "approve" {
		user.Role = req.RequestedRole
		user.UpdatedAt = now
		if err := s.userStore.Update(user); err != nil {
			return nil, utils.NewInternalError("role request approved but failed to update user role")
		}
	}

	return req, nil
}
