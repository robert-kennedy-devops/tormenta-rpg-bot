// Package engine — CombatEngine orchestrates a full turn-based combat exchange
// between two Combatants using the modular damage, skill, status and AI layers.
// It is completely stateless and side-effect-free: all mutations are applied to
// the Combatant values that the caller provides, and the caller is responsible
// for persisting them.
package engine

import (
	"fmt"
	"math/rand"
)

// ─── Turn result ──────────────────────────────────────────────────────────────

// TurnResult holds the full narrative + numeric outcome of one combat turn.
type TurnResult struct {
	// Attacker side
	AttackerDmg    int
	AttackerMsg    string
	AttackerCrit   bool
	AttackerMiss   bool
	AttackerFumble bool

	// Defender side (counter-attack / monster attack)
	DefenderDmg    int
	DefenderMsg    string
	DefenderCrit   bool
	DefenderMiss   bool

	// Status ticks
	AttackerStatusTick TickResult
	DefenderStatusTick TickResult

	// Resource cost
	MPConsumed     int
	EnergyConsumed int

	// Extra events
	Killed    bool   // defender dropped to 0 HP
	StatusMsg string // any status effect messages
}

// ─── CombatEngine ─────────────────────────────────────────────────────────────

// CombatEngine holds configuration for a fight session.
type CombatEngine struct {
	Rng *rand.Rand
}

// NewCombatEngine creates an engine seeded with the given value.
func NewCombatEngine(seed int64) *CombatEngine {
	return &CombatEngine{Rng: rand.New(rand.NewSource(seed))}
}

// TurnInput carries everything needed to process one turn.
type TurnInput struct {
	// Attacker
	Attacker    Combatant
	AtkBonus    int     // d20 attack bonus
	AtkStatMod  int     // primary stat modifier for damage
	AtkLevel    int
	AtkElement  Element
	AtkSkills   []Effect // nil = basic attack
	MPAvail     int
	MPCost      int
	EnergyCost  int

	// Defender
	Defender    Combatant
	DefCA       int     // armour class of target
	DefArmor    int     // flat damage reduction

	// AI context (set when defender is a monster)
	AICtx       *AIContext
}

// ResolveTurn executes one full turn (attacker acts → defender responds).
func (ce *CombatEngine) ResolveTurn(in TurnInput) TurnResult {
	result := TurnResult{}

	// ── 1. Status ticks on attacker ────────────────────────────────────────
	result.AttackerStatusTick = in.Attacker.GetStatusSet().Tick()
	if result.AttackerStatusTick.TotalDamage > 0 {
		hp := in.Attacker.GetHP() - result.AttackerStatusTick.TotalDamage
		in.Attacker.SetHP(hp)
		for _, m := range result.AttackerStatusTick.Messages {
			result.StatusMsg += m + "\n"
		}
	}

	// ── 2. Can attacker act? ───────────────────────────────────────────────
	attackPenalty := in.Attacker.GetStatusSet().AttackPenalty()
	if attackPenalty == 0.0 {
		result.AttackerMsg = "⚡ Você está atordoado e não pode agir!"
		// Defender still gets to attack
		result = ce.resolveDefenderTurn(in, result)
		return result
	}

	// ── 3. d20 hit check ──────────────────────────────────────────────────
	roll := ce.Rng.Intn(20) + 1
	adjustedBonus := int(float64(in.AtkBonus) * attackPenalty)
	hits, isCrit, isFumble := D20HitCheck(roll, adjustedBonus, in.DefCA)

	result.AttackerCrit = isCrit
	result.AttackerFumble = isFumble
	result.AttackerMiss = !hits

	if isFumble {
		result.AttackerMsg = fmt.Sprintf("🎲 [1] 💨 *Fumble!* Você escorregou e errou feio!")
	} else if !hits {
		result.AttackerMsg = fmt.Sprintf("🎲 [%d]+%d vs CA %d — ❌ *Errou!*", roll, adjustedBonus, in.DefCA)
	} else {
		// ── 4. Calculate damage ───────────────────────────────────────────
		critMult := 1.0
		if isCrit {
			critMult = 2.0
		}
		dmgBonus := in.Attacker.GetStatusSet().DamageBonus()

		var dmg int
		var dmgDesc string

		if len(in.AtkSkills) > 0 && in.MPAvail >= in.MPCost {
			// Skill attack — process each effect
			proc := ProcessEffects(in.AtkSkills, in.Attacker, in.Defender)
			dmg = proc.DamageDealt
			result.MPConsumed = in.MPCost
			result.EnergyConsumed = in.EnergyCost
			for _, m := range proc.Messages {
				dmgDesc += m + " "
			}
		} else {
			// Basic attack
			params := ScalingParams{
				BaseDamage:    ce.Rng.Intn(8) + 1, // 1d8 base
				AttackerStat:  in.AtkStatMod,
				StatWeight:    0.4,
				DefenderArmor: in.DefArmor,
				LevelScaling:  0.5,
				AttackerLevel: in.AtkLevel,
				ElementMult:   ElementMultiplier(in.AtkElement, in.Defender.GetWeaknesses(), in.Defender.GetResistances()),
				CritMult:      critMult,
				SkillMult:     dmgBonus,
			}
			dmg = CalculateDamage(params)
		}

		in.Defender.SetHP(in.Defender.GetHP() - dmg)
		result.AttackerDmg = dmg

		if isCrit {
			result.AttackerMsg = fmt.Sprintf("🎲 [20] ⭐ *CRÍTICO!* %s → *%d* de dano! (dobrado)", dmgDesc, dmg)
		} else {
			result.AttackerMsg = fmt.Sprintf("🎲 [%d]+%d vs CA %d ✅ — %s*%d* de dano.", roll, adjustedBonus, in.DefCA, dmgDesc, dmg)
		}

		if in.Defender.GetHP() <= 0 {
			result.Killed = true
			return result
		}
	}

	// ── 5. Defender counterattacks ─────────────────────────────────────────
	result = ce.resolveDefenderTurn(in, result)
	return result
}

func (ce *CombatEngine) resolveDefenderTurn(in TurnInput, result TurnResult) TurnResult {
	// Status tick on defender
	result.DefenderStatusTick = in.Defender.GetStatusSet().Tick()
	if result.DefenderStatusTick.TotalDamage > 0 {
		hp := in.Defender.GetHP() - result.DefenderStatusTick.TotalDamage
		in.Defender.SetHP(hp)
	}

	// Defender attacks back (simplified basic attack)
	roll := ce.Rng.Intn(20) + 1
	defBonus := in.DefCA / 3 // rough approximation for monster attack bonus
	hits, isCrit, isFumble := D20HitCheck(roll, defBonus, in.AtkBonus+10)

	if isFumble || !hits {
		result.DefenderMsg = fmt.Sprintf("🎲 [%d] ❌ *Defender errou!*", roll)
		result.DefenderMiss = true
	} else {
		critMult := 1.0
		if isCrit {
			critMult = 2.0
			result.DefenderCrit = true
		}
		params := ScalingParams{
			BaseDamage:    ce.Rng.Intn(6) + 1,
			DefenderArmor: 0,
			CritMult:      critMult,
			ElementMult:   1.0,
			SkillMult:     1.0,
			LevelScaling:  0.4,
		}
		dmg := CalculateDamage(params)
		dmg = int(float64(dmg) * in.Attacker.GetStatusSet().DamageTaken())
		in.Attacker.SetHP(in.Attacker.GetHP() - dmg)
		result.DefenderDmg = dmg
		if isCrit {
			result.DefenderMsg = fmt.Sprintf("🎲 [20] 💥 *CRÍTICO!* Defensor causou *%d* de dano!", dmg)
		} else {
			result.DefenderMsg = fmt.Sprintf("🎲 [%d] Defensor causou *%d* de dano.", roll, dmg)
		}
	}
	return result
}
