package models

import "time"

type TransactionType string

const (
	TypeIncome  TransactionType = "income"
	TypeExpense TransactionType = "expense"
)

type Transaction struct {
	ID          string          `json:"id"`
	UserID      string          `json:"userId"`
	Amount      float64         `json:"amount"`
	Type        TransactionType `json:"type"`
	Category    string          `json:"category"`
	Date        time.Time       `json:"date"`
	Description string          `json:"description"`
	IsDeleted   bool            `json:"isDeleted"`
	CreatedAt   time.Time       `json:"createdAt"`
	UpdatedAt   time.Time       `json:"updatedAt"`
}

type TransactionFilter struct {
	Type     string
	Category string
	From     time.Time
	To       time.Time
	Search   string
	Sort     string
	Limit    int
	Offset   int
}
