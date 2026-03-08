package guild

import (
	"errors"
	"sync"
	"time"
)

// ─── Territory ────────────────────────────────────────────────────────────────

var (
	ErrTerritoryNotFound   = errors.New("território não encontrado")
	ErrWarAlreadyActive    = errors.New("guerra de guilda já ativa para este território")
	ErrWarNotFound         = errors.New("guerra não encontrada")
	ErrWarNotActive        = errors.New("guerra não está ativa")
	ErrGuildNotParticipant = errors.New("sua guilda não participa desta guerra")
)

// Territory is a game world area that guilds can capture for bonuses.
type Territory struct {
	ID             string
	Name           string
	Emoji          string
	Description    string
	OwnerGuildID   int64  // 0 = uncontrolled
	OwnerGuildName string
	Bonus          TerritoryBonus
	CapturedAt     time.Time
}

// TerritoryBonus describes what a guild gains from holding a territory.
type TerritoryBonus struct {
	GoldIncomePerHour int // passive gold into guild bank every hour
	XPBonusPct        int // % XP bonus for all members while in this territory's map
	ResourceMultiplier float64 // crafting material drop rate multiplier
}

// Territories lists all capturable territories.
var Territories = map[string]*Territory{
	"darkwood": {
		ID: "darkwood", Name: "Floresta Sombria", Emoji: "🌲",
		Description: "Uma floresta antiga com recursos raros.",
		Bonus: TerritoryBonus{GoldIncomePerHour: 50, XPBonusPct: 10, ResourceMultiplier: 1.2},
	},
	"iron_keep": {
		ID: "iron_keep", Name: "Fortaleza de Ferro", Emoji: "🏰",
		Description: "Antiga fortaleza com minas de metal valioso.",
		Bonus: TerritoryBonus{GoldIncomePerHour: 100, XPBonusPct: 5, ResourceMultiplier: 1.5},
	},
	"sunken_temple": {
		ID: "sunken_temple", Name: "Templo Submerso", Emoji: "🏛️",
		Description: "Ruínas místicas que amplificam o poder mágico.",
		Bonus: TerritoryBonus{GoldIncomePerHour: 80, XPBonusPct: 15, ResourceMultiplier: 1.1},
	},
	"dragons_peak": {
		ID: "dragons_peak", Name: "Pico do Dragão", Emoji: "🐉",
		Description: "O território mais perigoso — e o mais lucrativo.",
		Bonus: TerritoryBonus{GoldIncomePerHour: 200, XPBonusPct: 20, ResourceMultiplier: 2.0},
	},
}

// ─── Guild War ────────────────────────────────────────────────────────────────

// WarStatus tracks a war's lifecycle.
type WarStatus string

const (
	WarScheduled WarStatus = "scheduled"
	WarActive    WarStatus = "active"
	WarEnded     WarStatus = "ended"
)

// GuildWar represents a scheduled territorial battle between two guilds.
type GuildWar struct {
	ID            int64
	TerritoryID   string
	AttackerID    int64 // guild ID
	AttackerName  string
	DefenderID    int64
	DefenderName  string
	AttackerScore int
	DefenderScore int
	Status        WarStatus
	ScheduledAt   time.Time
	StartsAt      time.Time
	EndsAt        time.Time
	WinnerGuildID int64
}

// ─── War store ────────────────────────────────────────────────────────────────

// WarStore handles guild war persistence.
type WarStore interface {
	CreateWar(w *GuildWar) error
	GetWar(id int64) (*GuildWar, error)
	ActiveWarForTerritory(territoryID string) (*GuildWar, error)
	UpdateWar(w *GuildWar) error
	ListActiveWars() ([]*GuildWar, error)
}

// MemWarStore is an in-memory war store.
type MemWarStore struct {
	mu   sync.RWMutex
	wars map[int64]*GuildWar
	seq  int64
}

func NewMemWarStore() *MemWarStore {
	return &MemWarStore{wars: make(map[int64]*GuildWar)}
}

func (s *MemWarStore) CreateWar(w *GuildWar) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.seq++
	w.ID = s.seq
	s.wars[w.ID] = w
	return nil
}

func (s *MemWarStore) GetWar(id int64) (*GuildWar, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	w, ok := s.wars[id]
	if !ok {
		return nil, ErrWarNotFound
	}
	return w, nil
}

func (s *MemWarStore) ActiveWarForTerritory(territoryID string) (*GuildWar, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, w := range s.wars {
		if w.TerritoryID == territoryID && (w.Status == WarActive || w.Status == WarScheduled) {
			return w, nil
		}
	}
	return nil, ErrWarNotFound
}

func (s *MemWarStore) UpdateWar(w *GuildWar) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.wars[w.ID]; !ok {
		return ErrWarNotFound
	}
	s.wars[w.ID] = w
	return nil
}

func (s *MemWarStore) ListActiveWars() ([]*GuildWar, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []*GuildWar
	for _, w := range s.wars {
		if w.Status == WarActive {
			out = append(out, w)
		}
	}
	return out, nil
}

// ─── War service ──────────────────────────────────────────────────────────────

// WarService manages guild war scheduling and scoring.
type WarService struct {
	wars  WarStore
	store Store
}

func NewWarService(store Store, wars WarStore) *WarService {
	return &WarService{wars: wars, store: store}
}

const (
	WarNoticePeriod = 30 * time.Minute // declaration to start
	WarDuration     = 1 * time.Hour    // battle window
)

// DeclareWar schedules a guild war for a territory.
func (s *WarService) DeclareWar(attackerGuildID int64, territoryID string) (*GuildWar, error) {
	t, ok := Territories[territoryID]
	if !ok {
		return nil, ErrTerritoryNotFound
	}
	// Check for existing war
	if _, err := s.wars.ActiveWarForTerritory(territoryID); err == nil {
		return nil, ErrWarAlreadyActive
	}
	attacker, err := s.store.GetByID(attackerGuildID)
	if err != nil {
		return nil, err
	}
	defenderID := t.OwnerGuildID
	defenderName := t.OwnerGuildName
	if defenderID == 0 {
		defenderName = "(Território Neutro)"
	}

	now := time.Now()
	w := &GuildWar{
		TerritoryID:  territoryID,
		AttackerID:   attackerGuildID,
		AttackerName: attacker.Name,
		DefenderID:   defenderID,
		DefenderName: defenderName,
		Status:       WarScheduled,
		ScheduledAt:  now,
		StartsAt:     now.Add(WarNoticePeriod),
		EndsAt:       now.Add(WarNoticePeriod + WarDuration),
	}
	if err := s.wars.CreateWar(w); err != nil {
		return nil, err
	}
	return w, nil
}

// RecordAttack records a guild member's combat contribution to a war.
// Points are scaled to damage dealt.
func (s *WarService) RecordAttack(warID, guildID int64, points int) error {
	w, err := s.wars.GetWar(warID)
	if err != nil {
		return err
	}
	if w.Status != WarActive {
		return ErrWarNotActive
	}
	if guildID == w.AttackerID {
		w.AttackerScore += points
	} else if guildID == w.DefenderID {
		w.DefenderScore += points
	} else {
		return ErrGuildNotParticipant
	}
	return s.wars.UpdateWar(w)
}

// SettleWar closes a war and awards territory to the winner.
func (s *WarService) SettleWar(warID int64) (*GuildWar, error) {
	w, err := s.wars.GetWar(warID)
	if err != nil {
		return nil, err
	}
	w.Status = WarEnded
	if w.AttackerScore >= w.DefenderScore {
		w.WinnerGuildID = w.AttackerID
		t := Territories[w.TerritoryID]
		t.OwnerGuildID = w.AttackerID
		t.OwnerGuildName = w.AttackerName
		t.CapturedAt = time.Now()
	} else {
		w.WinnerGuildID = w.DefenderID
	}
	return w, s.wars.UpdateWar(w)
}
