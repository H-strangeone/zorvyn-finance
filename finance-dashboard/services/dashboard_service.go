package services

import (
	"finance-dashboard/models"
	"finance-dashboard/store"
	"sort"
)

type DashboardService struct {
	transactionStore store.TransactionStore
}

func NewDashboardService(transactionStore store.TransactionStore) *DashboardService {
	return &DashboardService{transactionStore: transactionStore}
}

// SummaryResponse — top level dashboard numbers
type SummaryResponse struct {
	TotalIncome   float64 `json:"totalIncome"`
	TotalExpenses float64 `json:"totalExpenses"`
	NetBalance    float64 `json:"netBalance"`
	TotalRecords  int     `json:"totalRecords"`
}

// CategoryTotal — one category's aggregated total
type CategoryTotal struct {
	Category string  `json:"category"`
	Total    float64 `json:"total"`
	Type     string  `json:"type"`
}

// MonthlyTrend — one month's income and expense summary
type MonthlyTrend struct {
	Month    string  `json:"month"` // "2024-01"
	Income   float64 `json:"income"`
	Expenses float64 `json:"expenses"`
	Net      float64 `json:"net"`
}

// CategoryBreakdown wraps income and expense categories
type CategoryBreakdown struct {
	Income   []CategoryTotal `json:"income"`
	Expenses []CategoryTotal `json:"expenses"`
}

func (s *DashboardService) GetSummary() *SummaryResponse {
	txs, _ := s.transactionStore.GetAll(models.TransactionFilter{
		Limit: 100000,
	})

	summary := &SummaryResponse{}

	for _, tx := range txs {
		summary.TotalRecords++
		if tx.Type == models.TypeIncome {
			summary.TotalIncome += tx.Amount
		} else {
			summary.TotalExpenses += tx.Amount
		}
	}

	summary.NetBalance = summary.TotalIncome - summary.TotalExpenses

	return summary
}

func (s *DashboardService) GetByCategory() *CategoryBreakdown {
	txs, _ := s.transactionStore.GetAll(models.TransactionFilter{
		Limit: 100000,
	})

	incomeTotals := make(map[string]float64)
	expenseTotals := make(map[string]float64)

	for _, tx := range txs {
		if tx.Type == models.TypeIncome {
			incomeTotals[tx.Category] += tx.Amount
		} else {
			expenseTotals[tx.Category] += tx.Amount
		}
	}

	breakdown := &CategoryBreakdown{
		Income:   make([]CategoryTotal, 0),
		Expenses: make([]CategoryTotal, 0),
	}

	for category, total := range incomeTotals {
		breakdown.Income = append(breakdown.Income, CategoryTotal{
			Category: category,
			Total:    total,
			Type:     "income",
		})
	}

	for category, total := range expenseTotals {
		breakdown.Expenses = append(breakdown.Expenses, CategoryTotal{
			Category: category,
			Total:    total,
			Type:     "expense",
		})
	}

	return breakdown
}

func (s *DashboardService) GetTrends() []MonthlyTrend {
	txs, _ := s.transactionStore.GetAll(models.TransactionFilter{
		Limit: 100000,
	})

	type monthData struct {
		income   float64
		expenses float64
	}

	monthMap := make(map[string]*monthData)

	for _, tx := range txs {
		key := tx.Date.Format("2006-01")

		if _, exists := monthMap[key]; !exists {
			monthMap[key] = &monthData{}
		}

		if tx.Type == models.TypeIncome {
			monthMap[key].income += tx.Amount
		} else {
			monthMap[key].expenses += tx.Amount
		}
	}

	// Convert to sorted slice
	trends := make([]MonthlyTrend, 0, len(monthMap))
	for month, data := range monthMap {
		trends = append(trends, MonthlyTrend{
			Month:    month,
			Income:   data.income,
			Expenses: data.expenses,
			Net:      data.income - data.expenses,
		})
	}

	sortTrendsByMonth(trends)

	return trends
}

func (s *DashboardService) GetRecent(limit int) []*models.Transaction {
	if limit <= 0 {
		limit = 5
	}
	if limit > 20 {
		limit = 20
	}

	txs, _ := s.transactionStore.GetAll(models.TransactionFilter{
		Sort:  "date_desc",
		Limit: limit,
	})

	return txs
}


func sortTrendsByMonth(trends []MonthlyTrend) {
    sort.Slice(trends, func(i, j int) bool {
        return trends[i].Month > trends[j].Month // descending
    })
}
