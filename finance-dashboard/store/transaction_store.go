package store

import (
	"finance-dashboard/models"
	"sort"
	"strings"
	"sync"
	"time"
)

type InMemoryTransactionStore struct {
	mu           sync.RWMutex
	transactions map[string]*models.Transaction // key: transaction ID
}

func NewInMemoryTransactionStore() *InMemoryTransactionStore {
	return &InMemoryTransactionStore{
		transactions: make(map[string]*models.Transaction),
	}
}

func (s *InMemoryTransactionStore) Create(tx *models.Transaction) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.transactions[tx.ID] = tx
	return nil
}

func (s *InMemoryTransactionStore) GetByID(id string, includeDeleted bool) *models.Transaction {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tx, exists := s.transactions[id]
	if !exists {
		return nil
	}

	if !includeDeleted && tx.IsDeleted {
		return nil
	}

	return tx
}

func (s *InMemoryTransactionStore) GetAll(filter models.TransactionFilter) ([]*models.Transaction, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Step 1: collect all non-deleted transactions that match filters
	var filtered []*models.Transaction

	for _, tx := range s.transactions {
		if tx.IsDeleted {
			continue
		}

		if filter.Type != "" && string(tx.Type) != filter.Type {
			continue
		}

		if filter.Category != "" &&
			!strings.EqualFold(tx.Category, filter.Category) {
			continue
		}

		if !filter.From.IsZero() && tx.Date.Before(filter.From) {
			continue
		}
		if !filter.To.IsZero() && tx.Date.After(filter.To) {
			continue
		}

		if filter.Search != "" {
			searchLower := strings.ToLower(filter.Search)
			categoryMatch := strings.Contains(
				strings.ToLower(tx.Category), searchLower,
			)
			descriptionMatch := strings.Contains(
				strings.ToLower(tx.Description), searchLower,
			)
			if !categoryMatch && !descriptionMatch {
				continue
			}
		}

		filtered = append(filtered, tx)
	}

	total := len(filtered)

	sortTransactions(filtered, filter.Sort)

	limit := filter.Limit
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	if offset >= total {
		return []*models.Transaction{}, total
	}

	end := offset + limit
	if end > total {
		end = total
	}

	return filtered[offset:end], total
}
func (s *InMemoryTransactionStore) Update(tx *models.Transaction) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.transactions[tx.ID]; !exists {
		return ErrTransactionNotFound
	}

	tx.UpdatedAt = time.Now().UTC()
	s.transactions[tx.ID] = tx
	return nil
}

func (s *InMemoryTransactionStore) SoftDelete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tx, exists := s.transactions[id]
	if !exists {
		return ErrTransactionNotFound
	}

	// Idempotent — deleting already deleted is fine
	// Why? Double delete should not error — the desired state is already achieved
	tx.IsDeleted = true
	tx.UpdatedAt = time.Now().UTC()
	return nil
}

func sortTransactions(txs []*models.Transaction, sortOption string) {
	if sortOption == "" || !isValidSort(sortOption) {
		sortOption = "date_desc"
	}

	sort.Slice(txs, func(i, j int) bool {
		switch sortOption {
		case "date_asc":
			return txs[i].Date.Before(txs[j].Date)
		case "date_desc":
			return txs[i].Date.After(txs[j].Date)
		case "amount_asc":
			return txs[i].Amount < txs[j].Amount
		case "amount_desc":
			return txs[i].Amount > txs[j].Amount
		default:
			return txs[i].Date.After(txs[j].Date) // fallback: date_desc
		}
	})
}

func isValidSort(s string) bool {
	switch s {
	case "date_asc", "date_desc", "amount_asc", "amount_desc":
		return true
	}
	return false
}
func (s *InMemoryTransactionStore) FetchAll() []*models.Transaction {
    s.mu.RLock()
    defer s.mu.RUnlock()

    var result []*models.Transaction
    for _, tx := range s.transactions {
        if !tx.IsDeleted {
            result = append(result, tx)
        }
    }
    return result
}