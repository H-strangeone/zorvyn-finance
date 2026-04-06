package store

import (
	"finance-dashboard/models"
	"sort"
	"sync"
	"time"
)

type InMemoryRoleRequestStore struct {
	mu       sync.RWMutex
	requests map[string]*models.RoleRequest
}

func NewInMemoryRoleRequestStore() *InMemoryRoleRequestStore {
	return &InMemoryRoleRequestStore{
		requests: make(map[string]*models.RoleRequest),
	}
}

func (s *InMemoryRoleRequestStore) Create(req *models.RoleRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.requests[req.ID] = req
	return nil
}

func (s *InMemoryRoleRequestStore) GetByID(id string) *models.RoleRequest {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.requests[id]
}

func (s *InMemoryRoleRequestStore) GetAll(status string) []*models.RoleRequest {
	s.mu.RLock()
	defer s.mu.RUnlock()

	requests := make([]*models.RoleRequest, 0, len(s.requests))
	for _, r := range s.requests {
		if status == "" || string(r.Status) == status {
			requests = append(requests, r)
		}
	}
	sort.Slice(requests, func(i, j int) bool {
		return requests[i].CreatedAt.After(requests[j].CreatedAt)
	})
	return requests
}

func (s *InMemoryRoleRequestStore) GetByUserID(userID string) []*models.RoleRequest {
	s.mu.RLock()
	defer s.mu.RUnlock()

	requests := make([]*models.RoleRequest, 0)
	for _, r := range s.requests {
		if r.UserID == userID {
			requests = append(requests, r)
		}
	}
	sort.Slice(requests, func(i, j int) bool {
		return requests[i].CreatedAt.After(requests[j].CreatedAt)
	})
	return requests
}

func (s *InMemoryRoleRequestStore) GetPendingByUserID(userID string) *models.RoleRequest {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, r := range s.requests {
		if r.UserID == userID && r.Status == models.StatusPending {
			return r
		}
	}
	return nil
}

func (s *InMemoryRoleRequestStore) Update(req *models.RoleRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.requests[req.ID]; !exists {
		return ErrRoleRequestNotFound
	}

	req.UpdatedAt = time.Now().UTC()
	s.requests[req.ID] = req
	return nil
}

func (s *InMemoryRoleRequestStore) ProcessRequest(
	id string,
	status models.RequestStatus,
	reviewedBy string,
	reviewNote string,
) (*models.RoleRequest, error) {
	// Full Lock not RLock — we are reading AND writing
	// Both operations happen under the same lock
	// This is what makes it atomic — check and update cannot be separated
	s.mu.Lock()
	defer s.mu.Unlock()

	req, exists := s.requests[id]
	if !exists {
		return nil, ErrRoleRequestNotFound
	}

	// State machine check happens while lock is held
	// Second admin hitting this simultaneously will block on Lock()
	// above, then reach here and find status is no longer pending
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