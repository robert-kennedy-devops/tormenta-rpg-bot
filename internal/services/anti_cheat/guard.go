package anti_cheat

import (
	"sync"
	"time"
)

type CallbackGuard struct {
	mu   sync.Mutex
	seen map[string]time.Time
}

func NewCallbackGuard() *CallbackGuard {
	return &CallbackGuard{seen: make(map[string]time.Time)}
}

// Allow returns false if the same key was seen within ttl.
func (g *CallbackGuard) Allow(key string, ttl time.Duration) bool {
	if key == "" {
		return true
	}
	now := time.Now()
	cutoff := now.Add(-ttl * 2)

	g.mu.Lock()
	defer g.mu.Unlock()

	for k, t := range g.seen {
		if t.Before(cutoff) {
			delete(g.seen, k)
		}
	}

	if ttl > 0 {
		if t, ok := g.seen[key]; ok && now.Sub(t) <= ttl {
			return false
		}
	}
	g.seen[key] = now
	return true
}

type TransitionGuard struct {
	allowed map[string]map[string]bool
}

func NewTransitionGuard() *TransitionGuard {
	return &TransitionGuard{
		allowed: map[string]map[string]bool{
			"idle":           {"combat": true, "dungeon_combat": true, "auto_hunt": true, "dungeon": true},
			"combat":         {"idle": true, "dungeon_combat": true},
			"dungeon":        {"dungeon_combat": true, "idle": true},
			"dungeon_combat": {"idle": true, "dungeon": true},
			"auto_hunt":      {"idle": true},
		},
	}
}

func (g *TransitionGuard) IsValid(from, to string) bool {
	if from == to {
		return true
	}
	next, ok := g.allowed[from]
	if !ok {
		// Unknown state: do not block for backward compatibility.
		return true
	}
	return next[to]
}
