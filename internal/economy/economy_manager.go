// Package economy provides a fully controlled economic layer to prevent inflation
// in the game world.  It tracks total gold in circulation, applies dynamic gold
// sinks and adjusts drop rates automatically when inflation spikes.
//
// All public methods are safe for concurrent use.
package economy

import (
	"sync"
	"sync/atomic"
	"time"
)

// ─── Thresholds ───────────────────────────────────────────────────────────────

const (
	// InflationWarnThreshold: gold per player that triggers a mild penalty.
	InflationWarnThreshold = 50_000
	// InflationCritThreshold: gold per player that triggers aggressive sinks.
	InflationCritThreshold = 200_000
)

// ─── EconomyManager ───────────────────────────────────────────────────────────

// EconomyManager is the central authority on gold flow.
type EconomyManager struct {
	mu sync.RWMutex

	totalGoldInCirculation int64 // atomic via methods
	totalPlayers           int64 // set by census worker
	lastSnapshot           time.Time
	snapshotHistory        []GoldSnapshot
}

// GoldSnapshot records the state of the economy at a point in time.
type GoldSnapshot struct {
	Time        time.Time
	TotalGold   int64
	TotalPlayers int64
	GoldPerPlayer float64
	InflationLevel InflationLevel
}

// InflationLevel categorises the current inflation state.
type InflationLevel int

const (
	InflationNormal   InflationLevel = 0
	InflationWarning  InflationLevel = 1
	InflationCritical InflationLevel = 2
)

// Global is the singleton EconomyManager used by all game systems.
var Global = &EconomyManager{lastSnapshot: time.Now()}

// ─── Gold flow ────────────────────────────────────────────────────────────────

// AddGold records gold entering the economy (monster drops, quest rewards, etc.).
func (m *EconomyManager) AddGold(amount int64) {
	atomic.AddInt64(&m.totalGoldInCirculation, amount)
}

// RemoveGold records gold leaving the economy (repairs, taxes, NPC purchases, etc.).
func (m *EconomyManager) RemoveGold(amount int64) {
	for {
		cur := atomic.LoadInt64(&m.totalGoldInCirculation)
		next := cur - amount
		if next < 0 {
			next = 0
		}
		if atomic.CompareAndSwapInt64(&m.totalGoldInCirculation, cur, next) {
			return
		}
	}
}

// SetPlayerCount updates the total online player census.
func (m *EconomyManager) SetPlayerCount(n int64) {
	atomic.StoreInt64(&m.totalPlayers, n)
}

// TotalGold returns the current gold in circulation.
func (m *EconomyManager) TotalGold() int64 {
	return atomic.LoadInt64(&m.totalGoldInCirculation)
}

// GoldPerPlayer returns average gold per player.
func (m *EconomyManager) GoldPerPlayer() float64 {
	p := atomic.LoadInt64(&m.totalPlayers)
	if p == 0 {
		return 0
	}
	return float64(m.TotalGold()) / float64(p)
}

// CurrentInflation returns the current inflation level.
func (m *EconomyManager) CurrentInflation() InflationLevel {
	gpp := m.GoldPerPlayer()
	switch {
	case gpp >= float64(InflationCritThreshold):
		return InflationCritical
	case gpp >= float64(InflationWarnThreshold):
		return InflationWarning
	default:
		return InflationNormal
	}
}

// ─── Snapshot ────────────────────────────────────────────────────────────────

// Snapshot records the current economic state.  Call from the economy worker
// every N minutes.
func (m *EconomyManager) Snapshot() GoldSnapshot {
	snap := GoldSnapshot{
		Time:           time.Now(),
		TotalGold:      m.TotalGold(),
		TotalPlayers:   atomic.LoadInt64(&m.totalPlayers),
		GoldPerPlayer:  m.GoldPerPlayer(),
		InflationLevel: m.CurrentInflation(),
	}
	m.mu.Lock()
	m.snapshotHistory = append(m.snapshotHistory, snap)
	if len(m.snapshotHistory) > 144 { // keep last 24h at 10min interval
		m.snapshotHistory = m.snapshotHistory[1:]
	}
	m.lastSnapshot = time.Now()
	m.mu.Unlock()
	return snap
}

// History returns a copy of the snapshot history.
func (m *EconomyManager) History() []GoldSnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]GoldSnapshot, len(m.snapshotHistory))
	copy(out, m.snapshotHistory)
	return out
}
