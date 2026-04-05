package store

import (
	"finance-dashboard/models"
	"sync"
	"time"
)

type InMemoryRoleRequestStore struct {
	mu       sync.RWMutex
	requests map[string]*models.RoleRequest // key: request ID
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
		// Empty status means return all — no filter applied
		if status == "" || string(r.Status) == status {
			requests = append(requests, r)
		}
	}
	return requests
}

func (s *InMemoryRoleRequestStore) GetByUserID(userID string) []*models.RoleRequest {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var requests []*models.RoleRequest
	for _, r := range s.requests {
		if r.UserID == userID {
			requests = append(requests, r)
		}
	}
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
