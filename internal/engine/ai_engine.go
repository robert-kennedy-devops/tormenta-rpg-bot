package engine

import "math/rand"

// ─── AI difficulty tiers ───────────────────────────────────────────────────────

// AITier controls how much an adaptive monster has evolved.
type AITier int

const (
	AITierBasic    AITier = 0 // vanilla — no adaptations
	AITierAdapted  AITier = 1 // has learned 1-2 counters
	AITierVeteran  AITier = 2 // has learned 3+ counters
	AITierElite    AITier = 3 // boss-grade intelligence
)

// ─── Adaptation flags ─────────────────────────────────────────────────────────

// MonsterAdaptation records what a monster type has learned from player behaviour.
// These are stored per-monster-type in the AI module and applied at combat start.
type MonsterAdaptation struct {
	MagicResistBonus   int // flat bonus to MagicDef from seeing lots of magic
	PhysResistBonus    int // flat bonus to Defense from seeing lots of melee
	CounterattackChance int // % chance to trigger a counterattack after being hit in melee
	HasEvasion         bool // gained dodge capability against ranged
	Tier               AITier
}

// ─── AI action selection ──────────────────────────────────────────────────────

// AIActionKind is what the monster AI decides to do this turn.
type AIActionKind string

const (
	AIActionBasicAttack   AIActionKind = "basic_attack"
	AIActionSkillAttack   AIActionKind = "skill_attack"
	AIActionCounterattack AIActionKind = "counterattack"
	AIActionHeal          AIActionKind = "heal"
	AIActionFlee          AIActionKind = "flee"
	AIActionEnrage        AIActionKind = "enrage"
)

// AIContext provides the monster decision-maker with combat state.
type AIContext struct {
	MonsterHPPct   float64 // current HP / max HP  (0.0–1.0)
	PlayerHPPct    float64
	TurnNumber     int
	Adaptation     MonsterAdaptation
	LastPlayerAction string // "magic" | "melee" | "ranged" | "skill"
	RandSource     *rand.Rand
}

// SelectAction returns the action the monster AI chooses this turn.
func SelectAction(ctx AIContext) AIActionKind {
	r := ctx.RandSource
	if r == nil {
		r = rand.New(rand.NewSource(rand.Int63()))
	}

	// Enrage when HP is very low — only for veteran/elite tier
	if ctx.MonsterHPPct < 0.15 && ctx.Adaptation.Tier >= AITierVeteran {
		return AIActionEnrage
	}

	// Healing: basic self-heal (if monster has it and HP < 30%)
	if ctx.MonsterHPPct < 0.30 && ctx.Adaptation.Tier >= AITierElite && r.Intn(100) < 25 {
		return AIActionHeal
	}

	// Counterattack when player used melee last turn and monster has adaptation
	if ctx.LastPlayerAction == "melee" &&
		ctx.Adaptation.CounterattackChance > 0 &&
		r.Intn(100) < ctx.Adaptation.CounterattackChance {
		return AIActionCounterattack
	}

	// Skill attack probability increases with tier
	skillChance := 10 + int(ctx.Adaptation.Tier)*10
	if ctx.TurnNumber%3 == 0 && r.Intn(100) < skillChance {
		return AIActionSkillAttack
	}

	return AIActionBasicAttack
}

// ─── Adaptation updater ───────────────────────────────────────────────────────

// UpdateAdaptation evolves a monster adaptation based on the player's latest action.
// Call this after each combat turn; persist the result back to the AI store.
func UpdateAdaptation(a MonsterAdaptation, playerAction string) MonsterAdaptation {
	switch playerAction {
	case "magic":
		a.MagicResistBonus += 1
		if a.MagicResistBonus > 20 {
			a.MagicResistBonus = 20
		}
	case "melee":
		a.PhysResistBonus += 1
		a.CounterattackChance += 2
		if a.PhysResistBonus > 15 {
			a.PhysResistBonus = 15
		}
		if a.CounterattackChance > 30 {
			a.CounterattackChance = 30
		}
	case "ranged":
		a.HasEvasion = true
	}

	// Promote tier when thresholds are reached
	score := a.MagicResistBonus + a.PhysResistBonus + a.CounterattackChance/2
	switch {
	case score >= 40:
		a.Tier = AITierElite
	case score >= 20:
		a.Tier = AITierVeteran
	case score >= 5:
		a.Tier = AITierAdapted
	}
	return a
}
