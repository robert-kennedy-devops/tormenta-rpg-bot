package pvp

import (
	"fmt"
	"sync"
	"time"
)

// ─── Arena match ──────────────────────────────────────────────────────────────

// MatchStatus tracks a live PVP match.
type MatchStatus string

const (
	MatchInProgress MatchStatus = "in_progress"
	MatchFinished   MatchStatus = "finished"
	MatchAbandoned  MatchStatus = "abandoned"
)

// ArenaMatch represents an active 1v1 PVP battle.
type ArenaMatch struct {
	ID       int64
	PlayerA  MatchPlayer
	PlayerB  MatchPlayer
	Turn     int
	Status   MatchStatus
	WinnerID int64
	StartedAt time.Time
	FinishedAt *time.Time
	Log      []string // combat narrative
}

// MatchPlayer holds combat state for one player during a PVP match.
type MatchPlayer struct {
	PlayerID   int64
	PlayerName string
	CharLevel  int
	Rating     int
	MaxHP      int
	CurrentHP  int
	MaxMP      int
	CurrentMP  int
	Attack     int
	Defense    int
	MagicAtk   int
	Speed      int
}

// ArenaManager manages all live PVP matches.
type ArenaManager struct {
	mu      sync.RWMutex
	matches map[int64]*ArenaMatch
	// playerID → matchID (active match index)
	playerMatch map[int64]int64
	seq         int64
}

// GlobalArena is the singleton arena manager.
var GlobalArena = &ArenaManager{
	matches:     make(map[int64]*ArenaMatch),
	playerMatch: make(map[int64]int64),
}

// StartMatch creates a new match from a MatchResult.
func (a *ArenaManager) StartMatch(result *MatchResult, playerA, playerB MatchPlayer) (*ArenaMatch, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Check neither player is already in a match
	if _, ok := a.playerMatch[result.PlayerA.PlayerID]; ok {
		return nil, fmt.Errorf("jogador %d já está em uma partida de PVP", result.PlayerA.PlayerID)
	}
	if _, ok := a.playerMatch[result.PlayerB.PlayerID]; ok {
		return nil, fmt.Errorf("jogador %d já está em uma partida de PVP", result.PlayerB.PlayerID)
	}

	a.seq++
	match := &ArenaMatch{
		ID:        a.seq,
		PlayerA:   playerA,
		PlayerB:   playerB,
		Status:    MatchInProgress,
		StartedAt: time.Now(),
	}
	match.PlayerA.CurrentHP = playerA.MaxHP
	match.PlayerA.CurrentMP = playerA.MaxMP
	match.PlayerB.CurrentHP = playerB.MaxHP
	match.PlayerB.CurrentMP = playerB.MaxMP

	a.matches[match.ID] = match
	a.playerMatch[playerA.PlayerID] = match.ID
	a.playerMatch[playerB.PlayerID] = match.ID
	return match, nil
}

// GetMatchForPlayer returns the active match for a player.
func (a *ArenaManager) GetMatchForPlayer(playerID int64) (*ArenaMatch, bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	mid, ok := a.playerMatch[playerID]
	if !ok {
		return nil, false
	}
	m, ok := a.matches[mid]
	return m, ok
}

// PVPAttackResult is the outcome of one attack in a PVP match.
type PVPAttackResult struct {
	AttackerID  int64
	AttackerMsg string
	Damage      int
	IsCrit      bool
	IsMiss      bool
	TargetHP    int
	MatchOver   bool
	WinnerID    int64
}

// Attack processes one player's attack action in a PVP match.
func (a *ArenaManager) Attack(attackerID int64) (PVPAttackResult, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	mid, ok := a.playerMatch[attackerID]
	if !ok {
		return PVPAttackResult{}, fmt.Errorf("você não está em uma partida de PVP")
	}
	match := a.matches[mid]
	if match.Status != MatchInProgress {
		return PVPAttackResult{}, fmt.Errorf("a partida já terminou")
	}

	attacker, defender := a.resolveParticipants(match, attackerID)

	// Simple damage calc — uses existing game layer formula approximation
	baseDmg := attacker.Attack + attacker.CharLevel/5
	defense := defender.Defense / 2
	dmg := baseDmg - defense
	if dmg < 1 {
		dmg = 1
	}

	// Crit: 10% chance
	isCrit := match.Turn%10 == 0
	if isCrit {
		dmg *= 2
	}

	defender.CurrentHP -= dmg
	msg := fmt.Sprintf("⚔️ %s causou *%d* de dano em %s.", attacker.PlayerName, dmg, defender.PlayerName)
	if isCrit {
		msg = fmt.Sprintf("⭐ CRÍTICO! %s causou *%d* de dano em %s!", attacker.PlayerName, dmg, defender.PlayerName)
	}
	match.Log = append(match.Log, msg)
	match.Turn++

	result := PVPAttackResult{
		AttackerID:  attackerID,
		AttackerMsg: msg,
		Damage:      dmg,
		IsCrit:      isCrit,
		TargetHP:    defender.CurrentHP,
	}

	if defender.CurrentHP <= 0 {
		now := time.Now()
		match.Status = MatchFinished
		match.WinnerID = attacker.PlayerID
		match.FinishedAt = &now
		result.MatchOver = true
		result.WinnerID = attacker.PlayerID

		// Cleanup
		delete(a.playerMatch, match.PlayerA.PlayerID)
		delete(a.playerMatch, match.PlayerB.PlayerID)

		// Update rankings
		GlobalRankStore.ApplyResult(attacker.PlayerID, defender.PlayerID, attacker.PlayerName, defender.PlayerName)
	}
	return result, nil
}

func (a *ArenaManager) resolveParticipants(match *ArenaMatch, attackerID int64) (*MatchPlayer, *MatchPlayer) {
	if match.PlayerA.PlayerID == attackerID {
		return &match.PlayerA, &match.PlayerB
	}
	return &match.PlayerB, &match.PlayerA
}

// Forfeit allows a player to concede the match.
func (a *ArenaManager) Forfeit(playerID int64) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	mid, ok := a.playerMatch[playerID]
	if !ok {
		return fmt.Errorf("você não está em uma partida de PVP")
	}
	match := a.matches[mid]
	attacker, defender := a.resolveParticipants(match, playerID)
	now := time.Now()
	match.Status = MatchFinished
	match.WinnerID = defender.PlayerID
	match.FinishedAt = &now
	match.Log = append(match.Log, fmt.Sprintf("🏳️ %s desistiu da partida.", attacker.PlayerName))
	delete(a.playerMatch, match.PlayerA.PlayerID)
	delete(a.playerMatch, match.PlayerB.PlayerID)
	GlobalRankStore.ApplyResult(defender.PlayerID, attacker.PlayerID, defender.PlayerName, attacker.PlayerName)
	return nil
}
