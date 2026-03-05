package services

import (
	"sync"
	"time"

	"github.com/tormenta-bot/internal/database"
)

type AutoHuntService struct {
	interval time.Duration
	locks    sync.Map
}

func NewAutoHuntService(interval time.Duration) *AutoHuntService {
	if interval <= 0 {
		interval = 60 * time.Second
	}
	return &AutoHuntService{interval: interval}
}

func (s *AutoHuntService) lock(charID int) func() {
	l, _ := s.locks.LoadOrStore(charID, &sync.Mutex{})
	mu := l.(*sync.Mutex)
	mu.Lock()
	return mu.Unlock
}

// ProcessOffline executes pending cycles based on time elapsed since last tick.
// runCycle should return false when the session must stop.
func (s *AutoHuntService) ProcessOffline(
	charID int,
	maxCycles int,
	loadSession func(int) (*database.AutoHuntSession, error),
	runCycle func(int, *database.AutoHuntSession) bool,
) (int, error) {
	unlock := s.lock(charID)
	defer unlock()

	session, err := loadSession(charID)
	if err != nil || session == nil || session.Status != "running" {
		return 0, err
	}

	cycles := int(time.Since(session.LastTickAt) / s.interval)
	if cycles <= 0 {
		return 0, nil
	}
	if maxCycles > 0 && cycles > maxCycles {
		cycles = maxCycles
	}

	done := 0
	for i := 0; i < cycles; i++ {
		current, e := loadSession(charID)
		if e != nil || current == nil || current.Status != "running" {
			break
		}
		if !runCycle(charID, current) {
			break
		}
		done++
	}
	return done, nil
}
