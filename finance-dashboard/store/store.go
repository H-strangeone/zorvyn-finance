package store

import (
	"errors"
	"finance-dashboard/models"
)

var (
	ErrEmailAlreadyExists  = errors.New("email already exists")
	ErrUserNotFound        = errors.New("user not found")
	ErrTransactionNotFound = errors.New("transaction not found")
	ErrRoleRequestNotFound = errors.New("role request not found")
)

type UserStore interface {
	Create(user *models.User) error
	GetByID(id string) *models.User
	GetByEmail(email string) *models.User
	GetAll() []*models.User
	Update(user *models.User) error
	UpdateStatus(id string, isActive bool) error
	HasActiveAdmin() bool
}

type TransactionStore interface {
	Create(tx *models.Transaction) error
	GetByID(id string, includeDeleted bool) *models.Transaction
	GetAll(filter models.TransactionFilter) ([]*models.Transaction, int)
	Update(tx *models.Transaction) error
	SoftDelete(id string) error
	FetchAll() []*models.Transaction
}

type RoleRequestStore interface {
	Create(req *models.RoleRequest) error
	GetByID(id string) *models.RoleRequest
	GetAll(status string) []*models.RoleRequest
	GetByUserID(userID string) []*models.RoleRequest
	GetPendingByUserID(userID string) *models.RoleRequest
	Update(req *models.RoleRequest) error
}
