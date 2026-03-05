package menu

import "sync"

// StateStore centralizes per-user navigation state and is safe for concurrent use.
type StateStore struct {
	mu      sync.RWMutex
	current map[int64]string
	back    map[int64]map[string]string
}

func NewStateStore() *StateStore {
	return &StateStore{
		current: make(map[int64]string),
		back:    make(map[int64]map[string]string),
	}
}

func (s *StateStore) SetCurrent(userID int64, screen string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.current[userID] = screen
}

func (s *StateStore) GetCurrent(userID int64) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.current[userID]
}

func (s *StateStore) SetBack(userID int64, menuDest, previous string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.back[userID] == nil {
		s.back[userID] = make(map[string]string)
	}
	s.back[userID][menuDest] = previous
}

func (s *StateStore) GetBack(userID int64, menuDest, fallback string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if byMenu, ok := s.back[userID]; ok {
		if v, ok := byMenu[menuDest]; ok && v != "" {
			return v
		}
	}
	return fallback
}
