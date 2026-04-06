package services

import (
	"finance-dashboard/models"
	"finance-dashboard/store"
	"finance-dashboard/utils"
	"time"
	"sync"
	"errors"
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
	status:= models.RequestStatus(models.StatusApproved)
	if input.Action == "reject" {
		status = models.RequestStatus(models.StatusRejected)
	}
	req, err := s.roleRequestStore.ProcessRequest(id, status, adminID, input.ReviewNote)
	if err != nil {
		return nil, utils.NewInternalError("failed to process role request")
	}
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
	if input.Action == "approve" {
    user := s.userStore.GetByID(req.UserID)
    if user == nil {
        return nil, utils.NewNotFoundError("user")
    }
    user.Role = req.RequestedRole
    user.UpdatedAt = time.Now().UTC()
    if err := s.userStore.Update(user); err != nil {
        return nil, utils.NewInternalError(
            "approved but failed to update user role — contact administrator",
        )
    }
}

	if err := s.roleRequestStore.Update(req); err != nil {
		return nil, utils.NewInternalError("failed to process role request")
	}

	

	return req, nil
}
type InMemoryRoleRequestStore struct {
	mu       sync.RWMutex
	requests map[string]*models.RoleRequest // key: request ID
}
var (
	ErrEmailAlreadyExists  = errors.New("email already exists")
	ErrUserNotFound        = errors.New("user not found")
	ErrTransactionNotFound = errors.New("transaction not found")
	ErrRoleRequestNotFound = errors.New("role request not found")
	ErrRequestAlreadyProcessed = errors.New("request already processed")
)
func (s *InMemoryRoleRequestStore) ProcessRequest(
	id string,
	status models.RequestStatus,
	reviewedBy string,
	reviewNote string,
) (*models.RoleRequest, error) {
	// Single lock covers both the check AND the update
	// This is what makes it atomic — no other goroutine can
	// read or write between our status check and our update
	s.mu.Lock()
	defer s.mu.Unlock()

	req, exists := s.requests[id]
	if !exists {
		return nil, ErrRoleRequestNotFound
	}

	// Finite state machine enforcement under lock
	// Both the check and the mutation happen while lock is held
	// Two admins approving simultaneously — second one hits this and gets error
	if req.Status != models.StatusPending {
		return nil, ErrRequestAlreadyProcessed
	}

	req.Status = status
	req.ReviewedBy = reviewedBy
	req.ReviewNote = reviewNote
	req.UpdatedAt = time.Now().UTC()
	s.requests[id] = req

	return req, nil
}