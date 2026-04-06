package store

import (
	"finance-dashboard/models"
	"sync"
	"time"
	"sort"
)

type InMemoryUserStore struct {
	mu         sync.RWMutex
	users      map[string]*models.User
	emailIndex map[string]string
}

func NewInMemoryUserStore() *InMemoryUserStore {
	return &InMemoryUserStore{
		users:      make(map[string]*models.User),
		emailIndex: make(map[string]string),
	}
}

func (s *InMemoryUserStore) Create(user *models.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.emailIndex[user.Email]; exists {
		return ErrEmailAlreadyExists
	}

	s.users[user.ID] = user
	s.emailIndex[user.Email] = user.ID
	return nil
}

func (s *InMemoryUserStore) GetByID(id string) *models.User {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.users[id]
}

func (s *InMemoryUserStore) GetByEmail(email string) *models.User {
	s.mu.RLock()
	defer s.mu.RUnlock()

	userID, exists := s.emailIndex[email]
	if !exists {
		return nil
	}
	return s.users[userID]
}

func (s *InMemoryUserStore) GetAll() []*models.User {
	s.mu.RLock()
	defer s.mu.RUnlock()

	users := make([]*models.User, 0, len(s.users))
	for _, u := range s.users {
		users = append(users, u)
	}
	sort.Slice(users, func(i, j int) bool {
        return users[i].CreatedAt.Before(users[j].CreatedAt)
    })
	return users
}

func (s *InMemoryUserStore) Update(user *models.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	existing, exists := s.users[user.ID]
	if !exists {
		return ErrUserNotFound
	}

	if existing.Email != user.Email {
		if _, taken := s.emailIndex[user.Email]; taken {
			return ErrEmailAlreadyExists
		}
		delete(s.emailIndex, existing.Email)
		s.emailIndex[user.Email] = user.ID
	}

	user.UpdatedAt = time.Now().UTC()
	s.users[user.ID] = user
	return nil
}

func (s *InMemoryUserStore) UpdateStatus(id string, isActive bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, exists := s.users[id]
	if !exists {
		return ErrUserNotFound
	}

	user.IsActive = isActive
	user.UpdatedAt = time.Now().UTC()
	s.users[id] = user
	return nil
}

func (s *InMemoryUserStore) HasActiveAdmin() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, u := range s.users {
		if u.Role == models.RoleAdmin && u.IsActive {
			return true
		}
	}
	return false
}
