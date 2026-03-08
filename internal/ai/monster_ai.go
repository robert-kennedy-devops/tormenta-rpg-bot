package ai

import (
	"sync"

	"github.com/tormenta-bot/internal/engine"
)

// ─── Monster adaptation store ─────────────────────────────────────────────────
//
// Adaptations are stored per monster-TYPE (not per instance).  Every time a
// monster type fights a player it can learn.  The adaptation is shared across
// all players — a magic-heavy server will cause ALL instances of that monster
// to build magic resistance over time.

// AdaptationStore manages adaptation records for monster types.
type AdaptationStore struct {
	mu      sync.RWMutex
	records map[string]engine.MonsterAdaptation // monsterTypeID → adaptation
}

// GlobalAdaptations is the singleton adaptation store.
var GlobalAdaptations = &AdaptationStore{
	records: make(map[string]engine.MonsterAdaptation),
}

// Get returns the adaptation for a monster type (or a blank adaptation if none exists yet).
func (s *AdaptationStore) Get(monsterTypeID string) engine.MonsterAdaptation {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.records[monsterTypeID]
}

// ApplyLearning observes a player action against a monster type and evolves its adaptation.
// Call this after every combat turn.
func (s *AdaptationStore) ApplyLearning(monsterTypeID string, playerAction ActionKind) {
	s.mu.Lock()
	defer s.mu.Unlock()
	current := s.records[monsterTypeID]
	updated := engine.UpdateAdaptation(current, string(playerAction))
	s.records[monsterTypeID] = updated
}

// ─── Monster AI advisor ───────────────────────────────────────────────────────

// MonsterAI ties together the player behaviour tracker and the engine AI layer
// to give a complete decision for one monster on one turn.

// TurnAdvice is what the AI recommends for a monster's turn.
type TurnAdvice struct {
	Action         engine.AIActionKind
	Adaptation     engine.MonsterAdaptation
	MagicResBonus  int // extra magic defense from adaptation
	PhysResBonus   int // extra physical defense from adaptation
	CounterChance  int // % chance to counterattack
}

// Advise produces a turn recommendation for a monster facing a player.
func Advise(
	monsterTypeID string,
	playerID int64,
	monsterHPPct float64,
	playerHPPct float64,
	turnNumber int,
) TurnAdvice {
	adaptation := GlobalAdaptations.Get(monsterTypeID)
	profile := GlobalTracker.Summary(playerID)

	dominant := string(profile.Dominant())

	ctx := engine.AIContext{
		MonsterHPPct:     monsterHPPct,
		PlayerHPPct:      playerHPPct,
		TurnNumber:       turnNumber,
		Adaptation:       adaptation,
		LastPlayerAction: dominant,
	}

	action := engine.SelectAction(ctx)

	return TurnAdvice{
		Action:        action,
		Adaptation:    adaptation,
		MagicResBonus: adaptation.MagicResistBonus,
		PhysResBonus:  adaptation.PhysResistBonus,
		CounterChance: adaptation.CounterattackChance,
	}
}

// RecordPlayerAction observes a player action and updates both the behaviour
// profile and the monster adaptation.  Call once per combat turn.
func RecordPlayerAction(playerID int64, monsterTypeID string, action ActionKind) {
	GlobalTracker.Observe(playerID, action)
	GlobalAdaptations.ApplyLearning(monsterTypeID, action)
}
