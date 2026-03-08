// Package season manages the 3-month competitive season cycle.
// Each season has a unique theme, new bosses and exclusive rewards.
// At the end of a season, rankings reset, rewards are distributed and a new
// season begins automatically.
package season

import (
	"sync"
	"time"
)

// ─── Season definition ────────────────────────────────────────────────────────

// Season represents one competitive cycle.
type Season struct {
	ID          int
	Name        string
	Emoji       string
	Description string
	ThemeBoss   string // world boss ID that is prominent this season
	StartedAt   time.Time
	EndsAt      time.Time
	IsActive    bool
}

// SeasonDuration is the standard length of one season.
const SeasonDuration = 90 * 24 * time.Hour // 3 months

// ─── Season manager ───────────────────────────────────────────────────────────

// Manager handles season lifecycle.
type Manager struct {
	mu      sync.RWMutex
	current *Season
	history []*Season
	nextID  int
}

// Global is the singleton season manager.
var Global = &Manager{nextID: 1}

// CurrentSeason returns the active season (or nil if none started yet).
func (m *Manager) CurrentSeason() *Season {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.current
}

// StartSeason initialises a new season.
func (m *Manager) StartSeason(name, emoji, description, themeBoss string) *Season {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.current != nil && m.current.IsActive {
		m.current.IsActive = false
		m.history = append(m.history, m.current)
	}

	now := time.Now()
	s := &Season{
		ID:          m.nextID,
		Name:        name,
		Emoji:       emoji,
		Description: description,
		ThemeBoss:   themeBoss,
		StartedAt:   now,
		EndsAt:      now.Add(SeasonDuration),
		IsActive:    true,
	}
	m.nextID++
	m.current = s
	return s
}

// CheckAndRollOver ends the current season if its time has elapsed and begins
// a new one.  Returns (ended, new season) — should be called by a background
// worker on a regular interval.
func (m *Manager) CheckAndRollOver() (ended bool, newSeason *Season) {
	m.mu.RLock()
	cur := m.current
	m.mu.RUnlock()

	if cur == nil || !cur.IsActive {
		return false, nil
	}
	if time.Now().Before(cur.EndsAt) {
		return false, nil
	}

	// Rotate to the next season definition
	idx := (cur.ID - 1) % len(SeasonDefs)
	def := SeasonDefs[idx]
	newSeason = m.StartSeason(def.Name, def.Emoji, def.Description, def.ThemeBoss)
	return true, newSeason
}

// History returns all past seasons.
func (m *Manager) History() []*Season {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]*Season, len(m.history))
	copy(out, m.history)
	return out
}

// ─── Predefined season rotation ───────────────────────────────────────────────

// SeasonDef is a template for a recurring season.
type SeasonDef struct {
	Name        string
	Emoji       string
	Description string
	ThemeBoss   string
}

// SeasonDefs is the rotation of season themes (cycles through).
var SeasonDefs = []SeasonDef{
	{
		Name: "A Tormenta Surge", Emoji: "🌀",
		Description: "A Tormenta desperta novamente. Defeite-a antes que consuma Arton.",
		ThemeBoss:   "tormenta",
	},
	{
		Name: "O Inverno Eterno", Emoji: "❄️",
		Description: "O Dragão de Gelo avança sobre as cidades. O frio congela toda a esperança.",
		ThemeBoss:   "frost_dragon",
	},
	{
		Name: "Praga do Deserto", Emoji: "🦂",
		Description: "O Rei Escorpião emerge das areias com exércitos de veneno.",
		ThemeBoss:   "scorpion_king",
	},
}
