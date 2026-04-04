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
	Amount      float64         `json:"amount"`      // always > 0, sign comes from Type
	Type        TransactionType `json:"type"`        // income or expense
	Category    string          `json:"category"`
	Date        time.Time       `json:"date"`
	Description string          `json:"description"` 
	IsDeleted   bool            `json:"isDeleted"`   // soft delete flag
	CreatedAt   time.Time       `json:"createdAt"`
	UpdatedAt   time.Time       `json:"updatedAt"`
}


type TransactionFilter struct {
	Type     string
	Category string
	From     time.Time
	To       time.Time
	Search   string
	Sort     string // default: date_desc
	Limit    int    // default: 10, max: 100
	Offset   int    // default: 0
}