package events

import (
	"math/rand"
	"sync"
	"time"
)

type Kind string

const (
	EventNone          Kind = "none"
	EventBloodMoon     Kind = "blood_moon"
	EventTormentaStorm Kind = "tormenta_storm"
	EventDoubleDrop    Kind = "double_drop"
)

type ActiveEvent struct {
	Kind      Kind
	StartedAt time.Time
	EndsAt    time.Time
}

type Manager struct {
	mu     sync.RWMutex
	active ActiveEvent
}

var Global = NewManager()

func NewManager() *Manager {
	return &Manager{active: ActiveEvent{Kind: EventNone}}
}

func (m *Manager) Active(now time.Time) ActiveEvent {
	if now.IsZero() {
		now = time.Now()
	}
	m.mu.RLock()
	ev := m.active
	m.mu.RUnlock()
	if ev.Kind != EventNone && now.After(ev.EndsAt) {
		m.Clear()
		return ActiveEvent{Kind: EventNone}
	}
	return ev
}

func (m *Manager) Clear() {
	m.mu.Lock()
	m.active = ActiveEvent{Kind: EventNone}
	m.mu.Unlock()
}

func (m *Manager) Start(kind Kind, d time.Duration, now time.Time) {
	if d <= 0 {
		d = 30 * time.Minute
	}
	if now.IsZero() {
		now = time.Now()
	}
	m.mu.Lock()
	m.active = ActiveEvent{
		Kind:      kind,
		StartedAt: now,
		EndsAt:    now.Add(d),
	}
	m.mu.Unlock()
}

func (m *Manager) MaybeRoll(now time.Time) ActiveEvent {
	if now.IsZero() {
		now = time.Now()
	}
	if ev := m.Active(now); ev.Kind != EventNone {
		return ev
	}
	// Low frequency roll for background worker ticks.
	if rand.Intn(100) > 4 {
		return ActiveEvent{Kind: EventNone}
	}
	kinds := []Kind{EventBloodMoon, EventTormentaStorm, EventDoubleDrop}
	kind := kinds[rand.Intn(len(kinds))]
	m.Start(kind, 30*time.Minute, now)
	return m.Active(now)
}

func DropRateMultiplier(ev ActiveEvent) float64 {
	switch ev.Kind {
	case EventDoubleDrop:
		return 2.0
	case EventBloodMoon:
		return 1.25
	default:
		return 1.0
	}
}

func MonsterDifficultyMultiplier(ev ActiveEvent) float64 {
	switch ev.Kind {
	case EventTormentaStorm:
		return 1.20
	case EventBloodMoon:
		return 1.10
	default:
		return 1.0
	}
}
