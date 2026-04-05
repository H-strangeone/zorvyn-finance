package services

import (
	"finance-dashboard/models"
	"finance-dashboard/store"
	"finance-dashboard/utils"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	userStore store.UserStore
}

func NewUserService(userStore store.UserStore) *UserService {
	return &UserService{userStore: userStore}
}

type CreateUserInput struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type UpdateUserInput struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

func (s *UserService) GetAll() []*models.UserResponse {
	users := s.userStore.GetAll()

	responses := make([]*models.UserResponse, 0, len(users))
	for _, u := range users {
		r := u.ToResponse()
		responses = append(responses, &r)
	}
	return responses
}

func (s *UserService) GetByID(id string) (*models.UserResponse, *utils.AppError) {
	if !utils.IsValidUUID(id) {
		return nil, utils.NewValidationError("invalid user ID format")
	}

	user := s.userStore.GetByID(id)
	if user == nil {
		return nil, utils.NewNotFoundError("user")
	}

	r := user.ToResponse()
	return &r, nil
}

func (s *UserService) Create(input CreateUserInput) (*models.UserResponse, *utils.AppError) {
	input.Email = utils.NormalizeEmail(input.Email)

	if input.Name == "" {
		return nil, utils.NewValidationError("name is required")
	}
	if !utils.IsValidEmail(input.Email) {
		return nil, utils.NewValidationError("invalid email format")
	}
	if len(input.Password) < 8 {
		return nil, utils.NewValidationError("password must be at least 8 characters")
	}
	if !utils.IsValidRole(input.Role) {
		return nil, utils.NewValidationError("role must be viewer, analyst, or admin")
	}

	if existing := s.userStore.GetByEmail(input.Email); existing != nil {
		return nil, utils.NewConflictError("email already registered")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(input.Password), 12,
	)
	if err != nil {
		return nil, utils.NewInternalError("failed to process password")
	}

	now := time.Now().UTC()
	user := &models.User{
		ID:        uuid.New().String(),
		Name:      input.Name,
		Email:     input.Email,
		Password:  string(hashedPassword),
		Role:      models.Role(input.Role),
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.userStore.Create(user); err != nil {
		if err == store.ErrEmailAlreadyExists {
			return nil, utils.NewConflictError("email already registered")
		}
		return nil, utils.NewInternalError("failed to create user")
	}

	r := user.ToResponse()
	return &r, nil
}

func (s *UserService) Update(id string, input UpdateUserInput) (*models.UserResponse, *utils.AppError) {
	if !utils.IsValidUUID(id) {
		return nil, utils.NewValidationError("invalid user ID format")
	}

	user := s.userStore.GetByID(id)
	if user == nil {
		return nil, utils.NewNotFoundError("user")
	}

	if input.Name != "" {
		user.Name = input.Name
	}
	if input.Email != "" {
		input.Email = utils.NormalizeEmail(input.Email)
		if !utils.IsValidEmail(input.Email) {
			return nil, utils.NewValidationError("invalid email format")
		}
		user.Email = input.Email
	}
	if input.Role != "" {
		if !utils.IsValidRole(input.Role) {
			return nil, utils.NewValidationError("role must be viewer, analyst, or admin")
		}
		user.Role = models.Role(input.Role)
	}

	if err := s.userStore.Update(user); err != nil {
		if err == store.ErrEmailAlreadyExists {
			return nil, utils.NewConflictError("email already taken")
		}
		return nil, utils.NewInternalError("failed to update user")
	}

	r := user.ToResponse()
	return &r, nil
}

func (s *UserService) UpdateStatus(id string, isActive bool) (*models.UserResponse, *utils.AppError) {
	if !utils.IsValidUUID(id) {
		return nil, utils.NewValidationError("invalid user ID format")
	}

	user := s.userStore.GetByID(id)
	if user == nil {
		return nil, utils.NewNotFoundError("user")
	}

	if user.IsActive == isActive {
		r := user.ToResponse()
		return &r, nil
	}

	if err := s.userStore.UpdateStatus(id, isActive); err != nil {
		return nil, utils.NewInternalError("failed to update user status")
	}

	user = s.userStore.GetByID(id)
	r := user.ToResponse()
	return &r, nil
}

func (s *UserService) SeedAdmin(name, email, password string) error {
	concreteStore, ok := s.userStore.(*store.InMemoryUserStore)
	if !ok {
		return nil
	}

	if concreteStore.HasActiveAdmin() {
		return nil // active admin exists, nothing to do
	}

	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(password), 12,
	)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	admin := &models.User{
		ID:        uuid.New().String(),
		Name:      name,
		Email:     utils.NormalizeEmail(email),
		Password:  string(hashedPassword),
		Role:      models.RoleAdmin,
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	return s.userStore.Create(admin)
}
