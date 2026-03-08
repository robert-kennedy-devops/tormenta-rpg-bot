// Package ai provides adaptive monster intelligence.
// Monsters learn from player behaviour patterns and evolve counter-strategies
// over time, making repeat encounters more challenging.
package ai

import (
	"sync"
)

// ─── Player behaviour fingerprint ─────────────────────────────────────────────

// ActionKind is the category of a player action observed by the AI.
type ActionKind string

const (
	ActionMelee  ActionKind = "melee"
	ActionMagic  ActionKind = "magic"
	ActionRanged ActionKind = "ranged"
	ActionSkill  ActionKind = "skill"
	ActionFlee   ActionKind = "flee"
	ActionHeal   ActionKind = "heal"
)

// BehaviourProfile tracks the aggregate behaviour of ONE player across ALL fights.
// Used to drive per-monster-type adaptation.
type BehaviourProfile struct {
	PlayerID  int64
	Counts    map[ActionKind]int // total uses
	Total     int                // total actions observed
}

// NewProfile creates an empty profile for a player.
func NewProfile(playerID int64) *BehaviourProfile {
	return &BehaviourProfile{
		PlayerID: playerID,
		Counts:   make(map[ActionKind]int),
	}
}

// Record adds one observed action to the profile.
func (p *BehaviourProfile) Record(action ActionKind) {
	p.Counts[action]++
	p.Total++
}

// Dominant returns the most frequently used action kind.
func (p *BehaviourProfile) Dominant() ActionKind {
	best := ActionMelee
	max := 0
	for k, v := range p.Counts {
		if v > max {
			max = v
			best = k
		}
	}
	return best
}

// Pct returns what percentage of actions fall into a given category (0–100).
func (p *BehaviourProfile) Pct(action ActionKind) int {
	if p.Total == 0 {
		return 0
	}
	return p.Counts[action] * 100 / p.Total
}

// ─── Behaviour tracker (global) ───────────────────────────────────────────────

// Tracker maintains behaviour profiles for all players.
// In production, these would be persisted to Redis/DB and loaded on demand.
type Tracker struct {
	mu       sync.RWMutex
	profiles map[int64]*BehaviourProfile
}

// GlobalTracker is the singleton.
var GlobalTracker = &Tracker{profiles: make(map[int64]*BehaviourProfile)}

// Get returns (and lazily creates) a profile for a player.
func (t *Tracker) Get(playerID int64) *BehaviourProfile {
	t.mu.Lock()
	defer t.mu.Unlock()
	p, ok := t.profiles[playerID]
	if !ok {
		p = NewProfile(playerID)
		t.profiles[playerID] = p
	}
	return p
}

// Observe records one action for a player.
func (t *Tracker) Observe(playerID int64, action ActionKind) {
	p := t.Get(playerID)
	t.mu.Lock()
	defer t.mu.Unlock()
	p.Record(action)
}

// Summary returns a copy of the profile for read-only access.
func (t *Tracker) Summary(playerID int64) BehaviourProfile {
	t.mu.RLock()
	defer t.mu.RUnlock()
	p, ok := t.profiles[playerID]
	if !ok {
		return BehaviourProfile{PlayerID: playerID, Counts: make(map[ActionKind]int)}
	}
	// shallow copy
	copy := BehaviourProfile{
		PlayerID: p.PlayerID,
		Counts:   make(map[ActionKind]int, len(p.Counts)),
		Total:    p.Total,
	}
	for k, v := range p.Counts {
		copy.Counts[k] = v
	}
	return copy
}
