package services

import (
	"finance-dashboard/config"
	"finance-dashboard/models"
	"finance-dashboard/store"
	"finance-dashboard/utils"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userStore store.UserStore
}

func NewAuthService(userStore store.UserStore) *AuthService {
	return &AuthService{userStore: userStore}
}

type RegisterInput struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string              `json:"token"`
	User  models.UserResponse `json:"user"`
}

func (s *AuthService) Register(input RegisterInput) (*AuthResponse, *utils.AppError) {
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
		Role:      models.RoleViewer,
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

	token, appErr := generateToken(user)
	if appErr != nil {
		return nil, appErr
	}

	return &AuthResponse{
		Token: token,
		User:  user.ToResponse(),
	}, nil
}

func (s *AuthService) Login(input LoginInput) (*AuthResponse, *utils.AppError) {
	input.Email = utils.NormalizeEmail(input.Email)
	if input.Email == "" || input.Password == "" {
		return nil, utils.NewValidationError("email and password are required")
	}

	user := s.userStore.GetByEmail(input.Email)
	if user == nil {
		return nil, utils.NewUnauthorizedError("invalid credentials")
	}

	if !user.IsActive {
		return nil, utils.NewUnauthorizedError("account has been deactivated")
	}

	err := bcrypt.CompareHashAndPassword(
		[]byte(user.Password), []byte(input.Password),
	)
	if err != nil {
		return nil, utils.NewUnauthorizedError("invalid credentials")
	}

	token, appErr := generateToken(user)
	if appErr != nil {
		return nil, appErr
	}

	return &AuthResponse{
		Token: token,
		User:  user.ToResponse(),
	}, nil
}

func generateToken(user *models.User) (string, *utils.AppError) {
	now := time.Now().UTC()

	claims := jwt.MapClaims{
		"userId": user.ID,
		"role":   string(user.Role),
		"jti":    uuid.New().String(),
		"iat":    now.Unix(),
		"exp": now.Add(
			time.Duration(config.App.JWTExpiryHours) * time.Hour,
		).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signed, err := token.SignedString([]byte(config.App.JWTSecret))
	if err != nil {
		return "", utils.NewInternalError("failed to generate token")
	}

	return signed, nil
}
