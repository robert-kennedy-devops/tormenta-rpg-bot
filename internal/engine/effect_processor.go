package engine

// ─── Effect definitions ────────────────────────────────────────────────────────

// EffectType categorises what a skill or item effect does.
type EffectType string

const (
	EffectDamage       EffectType = "damage"
	EffectHeal         EffectType = "heal"
	EffectApplyStatus  EffectType = "apply_status"
	EffectRemoveStatus EffectType = "remove_status"
	EffectStatBuff     EffectType = "stat_buff"
	EffectStatDebuff   EffectType = "stat_debuff"
	EffectSummon       EffectType = "summon"
	EffectAoE          EffectType = "aoe"
)

// Effect describes a single outcome produced by a skill/ability/item.
type Effect struct {
	Type       EffectType
	Element    Element
	BasePower  int        // raw power value (damage, heal amount, etc.)
	StatusKind StatusKind // for apply/remove status effects
	StatusTurns int
	StatusDmgPT int      // DoT damage per turn
	StatName   string    // "attack","defense", etc. for buff/debuff
	StatDelta  int       // amount to change the stat
	Radius     int       // for AoE: number of targets
}

// ─── EffectProcessor ──────────────────────────────────────────────────────────

// Combatant abstracts any entity that can receive effects (player or monster).
type Combatant interface {
	GetHP() int
	SetHP(int)
	GetStatusSet() *StatusSet
	GetLevel() int
	GetResistances() []Element
	GetWeaknesses() []Element
}

// ProcessResult records the aggregate outcome of processing all effects.
type ProcessResult struct {
	DamageDealt  int
	HealingDone  int
	StatusApplied []StatusKind
	Messages     []string
}

// ProcessEffect applies a single Effect from source to target and returns the result.
// It intentionally does NOT commit any DB writes — callers are responsible for
// persisting changed state.
func ProcessEffect(e Effect, attacker Combatant, target Combatant) ProcessResult {
	result := ProcessResult{}

	switch e.Type {
	case EffectDamage, EffectAoE:
		params := ScalingParams{
			BaseDamage:    e.BasePower,
			AttackerLevel: attacker.GetLevel(),
			ElementMult:   ElementMultiplier(e.Element, target.GetWeaknesses(), target.GetResistances()),
			StatWeight:    0.3,
			LevelScaling:  0.5,
			CritMult:      1.0,
			SkillMult:     attacker.GetStatusSet().DamageBonus(),
		}
		dmg := CalculateDamage(params)
		dmg = int(float64(dmg) * target.GetStatusSet().DamageTaken())
		target.SetHP(target.GetHP() - dmg)
		result.DamageDealt = dmg
		result.Messages = append(result.Messages, "💥 "+itoa(dmg)+" de dano ("+string(e.Element)+").")

	case EffectHeal:
		heal := e.BasePower
		target.SetHP(target.GetHP() + heal)
		result.HealingDone = heal
		result.Messages = append(result.Messages, "💚 +"+itoa(heal)+" HP recuperado.")

	case EffectApplyStatus:
		target.GetStatusSet().Apply(StatusEffect{
			Kind:          e.StatusKind,
			TurnsLeft:     e.StatusTurns,
			DamagePerTurn: e.StatusDmgPT,
		})
		result.StatusApplied = append(result.StatusApplied, e.StatusKind)
		result.Messages = append(result.Messages, statusApplyMsg(e.StatusKind))

	case EffectRemoveStatus:
		// Remove a specific debuff (e.g. cure poison)
		target.GetStatusSet().Clear()
		result.Messages = append(result.Messages, "✨ Efeito removido.")
	}

	return result
}

// ProcessEffects applies a list of effects sequentially and merges results.
func ProcessEffects(effects []Effect, attacker, target Combatant) ProcessResult {
	merged := ProcessResult{}
	for _, e := range effects {
		r := ProcessEffect(e, attacker, target)
		merged.DamageDealt += r.DamageDealt
		merged.HealingDone += r.HealingDone
		merged.StatusApplied = append(merged.StatusApplied, r.StatusApplied...)
		merged.Messages = append(merged.Messages, r.Messages...)
	}
	return merged
}

func statusApplyMsg(k StatusKind) string {
	switch k {
	case StatusPoison:
		return "☠️ Envenenado!"
	case StatusBurn:
		return "🔥 Em chamas!"
	case StatusFreeze:
		return "❄️ Congelado!"
	case StatusStun:
		return "⚡ Atordoado!"
	case StatusBlind:
		return "👁️ Cegado!"
	case StatusBerserk:
		return "💢 Berserk ativado!"
	case StatusShield:
		return "🛡️ Escudo ativo!"
	case StatusHaste:
		return "💨 Pressa ativada!"
	case StatusRegen:
		return "💚 Regeneração ativa!"
	case StatusCurse:
		return "💀 Amaldiçoado!"
	case StatusSilence:
		return "🔇 Silenciado!"
	case StatusProtect:
		return "🛡️ Proteção ativa!"
	}
	return "✨ Efeito aplicado."
}
