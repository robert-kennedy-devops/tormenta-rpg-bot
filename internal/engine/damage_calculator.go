// Package engine provides the core RPG computation layer — combat math, skill
// resolution, status effects and AI heuristics.  All functions are pure (no
// side-effects, no DB calls) so they can be unit-tested in isolation and reused
// from any handler, worker or service.
package engine

import "math"

// ─── Element system ───────────────────────────────────────────────────────────

// Element represents a damage element type.
type Element string

const (
	ElementNone    Element = "none"
	ElementFire    Element = "fire"
	ElementIce     Element = "ice"
	ElementLightning Element = "lightning"
	ElementPoison  Element = "poison"
	ElementHoly    Element = "holy"
	ElementDark    Element = "dark"
	ElementPhysical Element = "physical"
	ElementMagic   Element = "magic"
)

// Affinity describes how strong/weak a creature is against an element.
type Affinity float64

const (
	AffinityImmune     Affinity = 0.0
	AffinityResistant  Affinity = 0.5
	AffinityNeutral    Affinity = 1.0
	AffinityVulnerable Affinity = 1.5
	AffinityAbsorb     Affinity = -0.5 // heals instead of hurts
)

// ElementMatrix maps (attackElement, defenderWeakness) → Affinity multiplier.
// Weaknesses are declared on Monster/Player structs via a []Element slice.
func ElementMultiplier(attackElem Element, weaknesses, resistances []Element) float64 {
	for _, w := range weaknesses {
		if w == attackElem {
			return float64(AffinityVulnerable)
		}
	}
	for _, r := range resistances {
		if r == attackElem {
			return float64(AffinityResistant)
		}
	}
	return float64(AffinityNeutral)
}

// ─── Critical hit system ───────────────────────────────────────────────────────

// CritResult holds the outcome of a critical hit calculation.
type CritResult struct {
	IsCrit     bool
	Multiplier float64
}

// CalcCrit determines whether an attack is critical and its damage multiplier.
//
//   - baseChance: flat % chance (e.g. 5 = 5%)
//   - critMod:    additional % from skills/gear (additive)
//   - roll:       d100 roll (0-99)
func CalcCrit(baseChance, critMod, roll int) CritResult {
	total := baseChance + critMod
	if total > 95 {
		total = 95 // cap at 95%
	}
	if roll < total {
		return CritResult{IsCrit: true, Multiplier: 2.0}
	}
	return CritResult{IsCrit: false, Multiplier: 1.0}
}

// ─── Scaling damage formula ────────────────────────────────────────────────────

// ScalingParams bundles all inputs needed by the universal damage formula.
type ScalingParams struct {
	BaseDamage    int     // flat base (skill value, weapon damage dice result)
	AttackerStat  int     // primary offensive stat (STR / INT / DEX)
	StatWeight    float64 // how much the stat contributes (0.0–1.0)
	DefenderArmor int     // flat damage reduction
	LevelScaling  float64 // attacker_level^x exponent (typically 0.5)
	AttackerLevel int
	ElementMult   float64 // from ElementMultiplier()
	CritMult      float64 // from CritResult.Multiplier
	SkillMult     float64 // extra multiplier from skills (1.0 = none)
}

// CalculateDamage returns the final damage after all modifiers.
// Formula: ((base + stat*weight) * level_scale * element * crit * skill) - armor
func CalculateDamage(p ScalingParams) int {
	if p.StatWeight <= 0 {
		p.StatWeight = 0.3
	}
	if p.LevelScaling <= 0 {
		p.LevelScaling = 0.5
	}
	if p.ElementMult <= 0 {
		p.ElementMult = 1.0
	}
	if p.CritMult <= 0 {
		p.CritMult = 1.0
	}
	if p.SkillMult <= 0 {
		p.SkillMult = 1.0
	}

	base := float64(p.BaseDamage) + float64(p.AttackerStat)*p.StatWeight
	levelFactor := 1.0 + math.Pow(float64(p.AttackerLevel), p.LevelScaling)*0.05
	raw := base * levelFactor * p.ElementMult * p.CritMult * p.SkillMult
	final := int(raw) - p.DefenderArmor
	if final < 1 {
		final = 1
	}
	return final
}

// ─── Armour Class helper ───────────────────────────────────────────────────────

// D20HitCheck returns true if the attacker hits given their roll + bonus vs target CA.
// Natural 20 always hits; natural 1 always misses.
func D20HitCheck(roll, atkBonus, targetCA int) (hits bool, isCrit bool, isFumble bool) {
	if roll == 20 {
		return true, true, false
	}
	if roll == 1 {
		return false, false, true
	}
	return (roll + atkBonus) >= targetCA, false, false
}
