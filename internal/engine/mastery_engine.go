package engine

import (
	"fmt"
	"sync"
)

// ─── Mastery levels ───────────────────────────────────────────────────────────

// MasteryLevel represents how proficient a character is with a specific skill.
type MasteryLevel int

const (
	MasteryNovice      MasteryLevel = 0 // 0–9 uses
	MasteryApprentice  MasteryLevel = 1 // 10–29 uses
	MasteryAdept       MasteryLevel = 2 // 30–74 uses
	MasteryExpert      MasteryLevel = 3 // 75–149 uses
	MasteryMaster      MasteryLevel = 4 // 150–299 uses
	MasteryGrandmaster MasteryLevel = 5 // 300+ uses
)

// MasteryThresholds defines the use-count threshold for each mastery level.
var MasteryThresholds = [6]int{0, 10, 30, 75, 150, 300}

// MasteryNames maps mastery level to a display name.
var MasteryNames = [6]string{
	"Novato",
	"Aprendiz",
	"Adepto",
	"Especialista",
	"Mestre",
	"Grão-Mestre",
}

// MasteryEmoji maps mastery level to an emoji badge.
var MasteryEmoji = [6]string{"⚪", "🟢", "🔵", "🟣", "🟡", "🔴"}

// MasteryFromUses returns the mastery level for a given use count.
func MasteryFromUses(uses int) MasteryLevel {
	level := MasteryNovice
	for i, threshold := range MasteryThresholds {
		if uses >= threshold {
			level = MasteryLevel(i)
		}
	}
	return level
}

// ─── Mastery record ───────────────────────────────────────────────────────────

// SkillMastery tracks how many times a character has used a specific skill.
type SkillMastery struct {
	SkillID  string
	UseCount int
}

// Level computes the mastery level from the current use count.
func (s *SkillMastery) Level() MasteryLevel {
	return MasteryFromUses(s.UseCount)
}

// UsesUntilNext returns the number of remaining uses needed to reach the next level.
// Returns 0 if already at Grandmaster.
func (s *SkillMastery) UsesUntilNext() int {
	next := int(s.Level()) + 1
	if next >= len(MasteryThresholds) {
		return 0
	}
	return MasteryThresholds[next] - s.UseCount
}

// ─── Mastery bonus ────────────────────────────────────────────────────────────

// MasteryBonus describes the bonus a character gains from mastering a skill.
type MasteryBonus struct {
	// DamageMultiplier is applied to the skill's base damage (1.0 = no change).
	DamageMultiplier float64
	// MPCostMultiplier reduces MP cost (1.0 = no change, 0.8 = 20% cheaper).
	MPCostMultiplier float64
	// ExtraCritChance is added to the base critical hit chance.
	ExtraCritChance int
	// HealBonus is added to healing effects.
	HealBonus int
	// Description is shown to the player.
	Description string
}

// GetMasteryBonus returns the bonuses granted at a given mastery level.
func GetMasteryBonus(level MasteryLevel) MasteryBonus {
	switch level {
	case MasteryNovice:
		return MasteryBonus{
			DamageMultiplier: 1.00,
			MPCostMultiplier: 1.00,
			ExtraCritChance:  0,
			HealBonus:        0,
			Description:      "Sem bônus.",
		}
	case MasteryApprentice:
		return MasteryBonus{
			DamageMultiplier: 1.08,
			MPCostMultiplier: 0.97,
			ExtraCritChance:  1,
			HealBonus:        3,
			Description:      "+8% dano, -3% custo de MP, +1% crítico.",
		}
	case MasteryAdept:
		return MasteryBonus{
			DamageMultiplier: 1.18,
			MPCostMultiplier: 0.92,
			ExtraCritChance:  3,
			HealBonus:        8,
			Description:      "+18% dano, -8% custo de MP, +3% crítico.",
		}
	case MasteryExpert:
		return MasteryBonus{
			DamageMultiplier: 1.32,
			MPCostMultiplier: 0.85,
			ExtraCritChance:  6,
			HealBonus:        15,
			Description:      "+32% dano, -15% custo de MP, +6% crítico.",
		}
	case MasteryMaster:
		return MasteryBonus{
			DamageMultiplier: 1.50,
			MPCostMultiplier: 0.75,
			ExtraCritChance:  10,
			HealBonus:        25,
			Description:      "+50% dano, -25% custo de MP, +10% crítico.",
		}
	case MasteryGrandmaster:
		return MasteryBonus{
			DamageMultiplier: 1.75,
			MPCostMultiplier: 0.60,
			ExtraCritChance:  15,
			HealBonus:        40,
			Description:      "+75% dano, -40% custo de MP, +15% crítico.",
		}
	}
	return MasteryBonus{DamageMultiplier: 1.0, MPCostMultiplier: 1.0}
}

// ApplyMasteryToSkillRequest applies mastery bonuses to a SkillUseRequest,
// returning a modified copy with reduced MP cost.
func ApplyMasteryToSkillRequest(req SkillUseRequest, bonus MasteryBonus) SkillUseRequest {
	if bonus.MPCostMultiplier > 0 && bonus.MPCostMultiplier < 1.0 {
		reduced := int(float64(req.MPCost) * bonus.MPCostMultiplier)
		if reduced < 1 && req.MPCost > 0 {
			reduced = 1
		}
		req.MPCost = reduced
	}
	return req
}

// ─── Mastery store ────────────────────────────────────────────────────────────

// MasteryStore keeps an in-memory registry of every character's skill mastery data.
// Keys are formatted as "charID:skillID".
type MasteryStore struct {
	mu   sync.RWMutex
	data map[string]*SkillMastery
}

// GlobalMasteryStore is the package-level singleton.
var GlobalMasteryStore = &MasteryStore{
	data: make(map[string]*SkillMastery),
}

func masteryKey(charID int, skillID string) string {
	return fmt.Sprintf("%d:%s", charID, skillID)
}

// RecordUse increments the use count for a skill and returns the (potentially new)
// mastery level. If a new level was reached, levelled is true.
func (ms *MasteryStore) RecordUse(charID int, skillID string) (level MasteryLevel, levelled bool) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	key := masteryKey(charID, skillID)
	entry, ok := ms.data[key]
	if !ok {
		entry = &SkillMastery{SkillID: skillID, UseCount: 0}
		ms.data[key] = entry
	}

	prev := entry.Level()
	entry.UseCount++
	next := entry.Level()

	return next, next > prev
}

// Get returns the mastery record for a character+skill. Returns a zero record if
// the skill has never been used.
func (ms *MasteryStore) Get(charID int, skillID string) SkillMastery {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	key := masteryKey(charID, skillID)
	if entry, ok := ms.data[key]; ok {
		return *entry
	}
	return SkillMastery{SkillID: skillID, UseCount: 0}
}

// GetBonus returns the mastery bonus for a character's skill.
func (ms *MasteryStore) GetBonus(charID int, skillID string) MasteryBonus {
	m := ms.Get(charID, skillID)
	return GetMasteryBonus(m.Level())
}

// GetLevel returns just the mastery level.
func (ms *MasteryStore) GetLevel(charID int, skillID string) MasteryLevel {
	m := ms.Get(charID, skillID)
	return m.Level()
}

// Summary returns a displayable string summarising mastery across all learned skills.
func (ms *MasteryStore) Summary(charID int, learnedSkillIDs []string) string {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	out := ""
	for _, sid := range learnedSkillIDs {
		key := masteryKey(charID, sid)
		entry, ok := ms.data[key]
		if !ok {
			continue
		}
		lv := entry.Level()
		if lv == MasteryNovice {
			continue // skip novice to keep output short
		}
		out += fmt.Sprintf("%s %s: %s (%d usos)\n", MasteryEmoji[lv], sid, MasteryNames[lv], entry.UseCount)
	}
	if out == "" {
		return "Nenhuma maestria desenvolvida ainda."
	}
	return out
}

// BulkLoad loads mastery data from persistent storage into the store.
// Callers pass a slice of (charID, skillID, useCount) tuples loaded from DB.
func (ms *MasteryStore) BulkLoad(entries []MasteryEntry) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	for _, e := range entries {
		key := masteryKey(e.CharID, e.SkillID)
		ms.data[key] = &SkillMastery{SkillID: e.SkillID, UseCount: e.UseCount}
	}
}

// Snapshot returns all mastery entries for persistence (DB save).
func (ms *MasteryStore) Snapshot() []MasteryEntry {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	out := make([]MasteryEntry, 0, len(ms.data))
	for _, entry := range ms.data {
		// We can't extract charID from entry alone; we store it in the key.
		// A proper solution would store charID in the struct — for now this is a
		// simplified snapshot that can be extended when DB persistence is added.
		out = append(out, MasteryEntry{SkillID: entry.SkillID, UseCount: entry.UseCount})
	}
	return out
}

// MasteryEntry is a flat DTO used for bulk DB persistence.
type MasteryEntry struct {
	CharID   int
	SkillID  string
	UseCount int
}

// ─── Mastery-aware skill resolution ─────────────────────────────────────────────

// ResolveSkillWithMastery resolves a skill use applying mastery bonuses.
// It records the use, checks for level-ups, and applies damage/cost bonuses.
func ResolveSkillWithMastery(req SkillUseRequest, attacker, target Combatant, charID int) (SkillUseResult, MasteryLevel, bool) {
	level, levelled := GlobalMasteryStore.RecordUse(charID, req.SkillID)
	bonus := GetMasteryBonus(level)

	// Apply MP cost reduction from mastery.
	req = ApplyMasteryToSkillRequest(req, bonus)

	// Scale effect base power with damage multiplier.
	if bonus.DamageMultiplier != 1.0 {
		scaled := make([]Effect, len(req.SkillEffects))
		for i, e := range req.SkillEffects {
			if e.Type == EffectDamage || e.Type == EffectAoE {
				e.BasePower = int(float64(e.BasePower) * bonus.DamageMultiplier)
			}
			if e.Type == EffectHeal {
				e.BasePower += bonus.HealBonus
			}
			scaled[i] = e
		}
		req.SkillEffects = scaled
	}

	result := ResolveSkillUse(req, attacker, target)
	return result, level, levelled
}
