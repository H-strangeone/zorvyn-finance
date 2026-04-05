package services

import (
	"finance-dashboard/models"
	"finance-dashboard/store"
	"finance-dashboard/utils"
	"time"

	"github.com/google/uuid"
)

type TransactionService struct {
	transactionStore store.TransactionStore
}

func NewTransactionService(transactionStore store.TransactionStore) *TransactionService {
	return &TransactionService{transactionStore: transactionStore}
}

type CreateTransactionInput struct {
	Amount      float64 `json:"amount"`
	Type        string  `json:"type"`
	Category    string  `json:"category"`
	Date        string  `json:"date"`
	Description string  `json:"description"`
}

type UpdateTransactionInput struct {
	Amount      float64 `json:"amount"`
	Type        string  `json:"type"`
	Category    string  `json:"category"`
	Date        string  `json:"date"`
	Description string  `json:"description"`
}

type TransactionListResponse struct {
	Transactions []*models.Transaction `json:"transactions"`
	Total        int                   `json:"total"`
	Limit        int                   `json:"limit"`
	Offset       int                   `json:"offset"`
}

func (s *TransactionService) GetAll(filter models.TransactionFilter) (*TransactionListResponse, *utils.AppError) {
	if filter.Sort == "" || !utils.IsValidSortOption(filter.Sort) {
		filter.Sort = "date_desc"
	}
	if filter.Limit <= 0 {
		filter.Limit = 10
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}

	txs, total := s.transactionStore.GetAll(filter)

	return &TransactionListResponse{
		Transactions: txs,
		Total:        total,
		Limit:        filter.Limit,
		Offset:       filter.Offset,
	}, nil
}

func (s *TransactionService) GetByID(id string) (*models.Transaction, *utils.AppError) {
	if !utils.IsValidUUID(id) {
		return nil, utils.NewValidationError("invalid transaction ID format")
	}

	tx := s.transactionStore.GetByID(id, false) // false = exclude deleted
	if tx == nil {
		return nil, utils.NewNotFoundError("transaction")
	}

	return tx, nil
}

func (s *TransactionService) Create(userID string, input CreateTransactionInput) (*models.Transaction, *utils.AppError) {
	if input.Amount <= 0 {
		return nil, utils.NewValidationError("amount must be greater than 0")
	}

	if !utils.IsValidTransactionType(input.Type) {
		return nil, utils.NewValidationError("type must be income or expense")
	}

	if input.Category == "" {
		return nil, utils.NewValidationError("category is required")
	}

	parsedDate, err := parseDate(input.Date)
	if err != nil {
		return nil, utils.NewValidationError("invalid date format, use YYYY-MM-DD or RFC3339")
	}

	now := time.Now().UTC()
	tx := &models.Transaction{
		ID:          uuid.New().String(),
		UserID:      userID,
		Amount:      input.Amount,
		Type:        models.TransactionType(input.Type),
		Category:    input.Category,
		Date:        parsedDate,
		Description: input.Description,
		IsDeleted:   false,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.transactionStore.Create(tx); err != nil {
		return nil, utils.NewInternalError("failed to create transaction")
	}

	return tx, nil
}

func (s *TransactionService) Update(id string, input UpdateTransactionInput) (*models.Transaction, *utils.AppError) {
	if !utils.IsValidUUID(id) {
		return nil, utils.NewValidationError("invalid transaction ID format")
	}

	tx := s.transactionStore.GetByID(id, false)
	if tx == nil {
		return nil, utils.NewNotFoundError("transaction")
	}

	if input.Amount != 0 {
		if input.Amount <= 0 {
			return nil, utils.NewValidationError("amount must be greater than 0")
		}
		tx.Amount = input.Amount
	}
	if input.Type != "" {
		if !utils.IsValidTransactionType(input.Type) {
			return nil, utils.NewValidationError("type must be income or expense")
		}
		tx.Type = models.TransactionType(input.Type)
	}
	if input.Category != "" {
		tx.Category = input.Category
	}
	if input.Date != "" {
		parsedDate, err := parseDate(input.Date)
		if err != nil {
			return nil, utils.NewValidationError("invalid date format, use YYYY-MM-DD or RFC3339")
		}
		tx.Date = parsedDate
	}
	if input.Description != "" {
		tx.Description = input.Description
	}

	if err := s.transactionStore.Update(tx); err != nil {
		if err == store.ErrTransactionNotFound {
			return nil, utils.NewNotFoundError("transaction")
		}
		return nil, utils.NewInternalError("failed to update transaction")
	}

	return tx, nil
}

func (s *TransactionService) Delete(id string) *utils.AppError {
	if !utils.IsValidUUID(id) {
		return utils.NewValidationError("invalid transaction ID format")
	}

	tx := s.transactionStore.GetByID(id, false)
	if tx == nil {
		return utils.NewNotFoundError("transaction")
	}

	if err := s.transactionStore.SoftDelete(id); err != nil {
		return utils.NewInternalError("failed to delete transaction")
	}

	return nil
}

func parseDate(dateStr string) (time.Time, *utils.AppError) {
    if dateStr == "" {
        return time.Time{}, utils.NewValidationError("date is required")
    }
    if t, err := time.Parse(time.RFC3339, dateStr); err == nil {
        return t.UTC(), nil
    }
    if t, err := time.Parse("2006-01-02", dateStr); err == nil {
        return t.UTC(), nil
    }
    return time.Time{}, utils.NewValidationError("invalid date format, use YYYY-MM-DD or RFC3339")// Returns zero time on error — callers must check error before using time value
}
