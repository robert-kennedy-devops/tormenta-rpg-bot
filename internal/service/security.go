package service

import (
	"sync"
	"time"

	"github.com/tormenta-bot/internal/repository"
)

type SecurityService struct {
	repo     repository.PlayerRepository
	locks    sync.Map
	rlMu     sync.Mutex
	rateData map[int64]rateEntry
}

type rateEntry struct {
	windowStart time.Time
	count       int
}

func NewSecurityService(repo repository.PlayerRepository) *SecurityService {
	return &SecurityService{
		repo:     repo,
		rateData: make(map[int64]rateEntry),
	}
}

func (s *SecurityService) LockPlayer(playerID int64) func() {
	l, _ := s.locks.LoadOrStore(playerID, &sync.Mutex{})
	mu := l.(*sync.Mutex)
	mu.Lock()
	return mu.Unlock
}

// AllowUserAction returns false when user exceeds max actions in the given window.
func (s *SecurityService) AllowUserAction(userID int64, max int, window time.Duration) bool {
	if max <= 0 {
		max = 20
	}
	if window <= 0 {
		window = 2 * time.Second
	}
	now := time.Now()

	s.rlMu.Lock()
	defer s.rlMu.Unlock()

	cur := s.rateData[userID]
	if cur.windowStart.IsZero() || now.Sub(cur.windowStart) > window {
		s.rateData[userID] = rateEntry{windowStart: now, count: 1}
		return true
	}
	cur.count++
	s.rateData[userID] = cur
	return cur.count <= max
}

func (s *SecurityService) UpsertPlayer(userID int64, username string) error {
	return s.repo.UpsertPlayer(userID, username)
}
