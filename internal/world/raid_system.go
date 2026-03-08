package world

import (
	"sync"
	"time"
)

// ─── Raid difficulty ──────────────────────────────────────────────────────────

// RaidDifficulty sets the scaling modifier for a raid.
type RaidDifficulty string

const (
	RaidNormal    RaidDifficulty = "normal"
	RaidHard      RaidDifficulty = "hard"
	RaidLegendary RaidDifficulty = "legendary"
)

// ─── Raid definition ──────────────────────────────────────────────────────────

// RaidDef is a multi-stage boss encounter requiring a coordinated group.
type RaidDef struct {
	ID          string
	Name        string
	Emoji       string
	Description string
	MinPlayers  int
	MaxPlayers  int
	Stages      []RaidStage
	Rewards     BossRewardTable
}

// RaidStage is one phase of a raid (boss or mechanic).
type RaidStage struct {
	StageNum    int
	BossID      string // references world boss definition
	HPMultiplier float64
	Special     string // mechanic description
}

// Raids is the registry of all available raids.
var Raids = map[string]RaidDef{
	"tormenta_raid": {
		ID: "tormenta_raid", Name: "Raid: Núcleo da Tormenta", Emoji: "🌀",
		Description: "Penetre no coração da Tormenta para destruí-la de dentro.",
		MinPlayers: 5, MaxPlayers: 20,
		Stages: []RaidStage{
			{StageNum: 1, BossID: "scorpion_king", HPMultiplier: 1.5, Special: "Veneno em área — curem os aliados!"},
			{StageNum: 2, BossID: "frost_dragon", HPMultiplier: 1.8, Special: "Onda de gelo — role para esquivar!"},
			{StageNum: 3, BossID: "tormenta", HPMultiplier: 2.5, Special: "Fase final — todo o dano é dobrado."},
		},
		Rewards: BossRewardTable{
			GoldPerRank: []int{10000, 6000, 4000, 2000, 1000},
			XPPerRank:   []int{20000, 14000, 9000, 5000, 2000},
			LegendaryItems: []string{"tormenta_blade", "tormenta_crown", "void_heart"},
			GuaranteedDrop: "raid_token",
		},
	},
}

// ─── Active raid ──────────────────────────────────────────────────────────────

// RaidSession tracks an ongoing raid run.
type RaidSession struct {
	ID          int64
	RaidID      string
	Difficulty  RaidDifficulty
	CurrentStage int
	Status      string // "recruiting" | "in_progress" | "completed" | "failed"
	Members     map[int64]string // playerID → name
	StartedAt   time.Time
	FinishedAt  *time.Time
}

// RaidManager manages active raid sessions.
type RaidManager struct {
	mu       sync.RWMutex
	sessions map[int64]*RaidSession
	seq      int64
}

// GlobalRaids is the singleton raid manager.
var GlobalRaids = &RaidManager{sessions: make(map[int64]*RaidSession)}

// CreateSession starts a new raid recruitment session.
func (m *RaidManager) CreateSession(raidID string, difficulty RaidDifficulty, leaderID int64, leaderName string) (*RaidSession, error) {
	def, ok := Raids[raidID]
	if !ok {
		return nil, ErrBossNotFound
	}
	_ = def // validate exists
	m.mu.Lock()
	defer m.mu.Unlock()
	m.seq++
	s := &RaidSession{
		ID:         m.seq,
		RaidID:     raidID,
		Difficulty: difficulty,
		Status:     "recruiting",
		Members:    map[int64]string{leaderID: leaderName},
		StartedAt:  time.Now(),
	}
	m.sessions[s.ID] = s
	return s, nil
}

// JoinSession adds a player to a recruiting raid.
func (m *RaidManager) JoinSession(sessionID, playerID int64, playerName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.sessions[sessionID]
	if !ok {
		return ErrBossNotFound
	}
	if s.Status != "recruiting" {
		return ErrBossNotFound
	}
	def := Raids[s.RaidID]
	if len(s.Members) >= def.MaxPlayers {
		return ErrBossNotFound
	}
	s.Members[playerID] = playerName
	return nil
}

// StartSession transitions a raid from recruiting to in_progress.
func (m *RaidManager) StartSession(sessionID int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.sessions[sessionID]
	if !ok {
		return ErrBossNotFound
	}
	def := Raids[s.RaidID]
	if len(s.Members) < def.MinPlayers {
		return ErrBossNotFound
	}
	s.Status = "in_progress"
	s.CurrentStage = 1
	return nil
}

// AdvanceStage moves the raid to the next stage.
func (m *RaidManager) AdvanceStage(sessionID int64) (completed bool, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.sessions[sessionID]
	if !ok {
		err = ErrBossNotFound
		return
	}
	def := Raids[s.RaidID]
	s.CurrentStage++
	if s.CurrentStage > len(def.Stages) {
		now := time.Now()
		s.Status = "completed"
		s.FinishedAt = &now
		completed = true
	}
	return
}

// ErrBossNotFound is returned when a boss or raid is not in the registry.
var ErrBossNotFound = &raidError{"boss ou raid não encontrado"}

type raidError struct{ msg string }

func (e *raidError) Error() string { return e.msg }
