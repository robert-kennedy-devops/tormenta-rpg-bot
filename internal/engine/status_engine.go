package engine

import "time"

// ─── Status Effect types ───────────────────────────────────────────────────────

// StatusKind identifies a status effect category.
type StatusKind string

const (
	StatusPoison     StatusKind = "poison"
	StatusBurn       StatusKind = "burn"
	StatusFreeze     StatusKind = "freeze"
	StatusStun       StatusKind = "stun"
	StatusBlind      StatusKind = "blind"
	StatusBerserk    StatusKind = "berserk"
	StatusShield     StatusKind = "shield"
	StatusHaste      StatusKind = "haste"
	StatusRegen      StatusKind = "regen"
	StatusCurse      StatusKind = "curse"
	StatusSilence    StatusKind = "silence"
	StatusProtect    StatusKind = "protect"
)

// StatusEffect is an active effect on a combatant.
type StatusEffect struct {
	Kind       StatusKind
	TurnsLeft  int
	DamagePerTurn int    // for DoTs (poison, burn)
	StatMod    int       // +/- modifier to a stat while active
	ExpiresAt  time.Time // absolute expiry (for out-of-combat effects)
}

// IsExpired returns true if the effect has no turns left or has passed its wall-clock expiry.
func (s *StatusEffect) IsExpired(now time.Time) bool {
	if !s.ExpiresAt.IsZero() && now.After(s.ExpiresAt) {
		return true
	}
	return s.TurnsLeft <= 0
}

// ─── StatusSet — collection of effects on one combatant ───────────────────────

// StatusSet manages multiple concurrent status effects.
type StatusSet struct {
	effects map[StatusKind]*StatusEffect
}

// NewStatusSet constructs an empty set.
func NewStatusSet() *StatusSet {
	return &StatusSet{effects: make(map[StatusKind]*StatusEffect)}
}

// Apply adds or overwrites an effect (last writer wins, allowing stronger sources
// to overwrite weaker ones if needed by callers).
func (s *StatusSet) Apply(e StatusEffect) {
	s.effects[e.Kind] = &e
}

// Has returns true if the effect kind is present and not expired.
func (s *StatusSet) Has(kind StatusKind) bool {
	e, ok := s.effects[kind]
	if !ok {
		return false
	}
	return !e.IsExpired(time.Now())
}

// Get returns the effect or nil.
func (s *StatusSet) Get(kind StatusKind) *StatusEffect {
	e := s.effects[kind]
	return e
}

// TickResult records what happened to a combatant during a status tick.
type TickResult struct {
	TotalDamage int
	TotalHeal   int
	Expired     []StatusKind
	Messages    []string
}

// Tick advances all effects by one turn, applies DoT/HoT, and prunes expired entries.
func (s *StatusSet) Tick() TickResult {
	now := time.Now()
	result := TickResult{}
	for kind, e := range s.effects {
		if e.IsExpired(now) {
			result.Expired = append(result.Expired, kind)
			delete(s.effects, kind)
			continue
		}
		switch kind {
		case StatusPoison, StatusBurn:
			dmg := e.DamagePerTurn
			result.TotalDamage += dmg
			result.Messages = append(result.Messages, statusDoTMsg(kind, dmg, e.TurnsLeft))
		case StatusRegen:
			heal := e.DamagePerTurn // reusing field for heal amount
			result.TotalHeal += heal
		}
		e.TurnsLeft--
	}
	return result
}

// Clear removes all effects.
func (s *StatusSet) Clear() {
	s.effects = make(map[StatusKind]*StatusEffect)
}

// All returns a snapshot of current effects (for serialisation / display).
func (s *StatusSet) All() []StatusEffect {
	out := make([]StatusEffect, 0, len(s.effects))
	for _, e := range s.effects {
		out = append(out, *e)
	}
	return out
}

func statusDoTMsg(kind StatusKind, dmg, turnsLeft int) string {
	switch kind {
	case StatusPoison:
		return "☠️ Veneno causou " + itoa(dmg) + " de dano (" + itoa(turnsLeft-1) + " turnos restantes)."
	case StatusBurn:
		return "🔥 Queimadura causou " + itoa(dmg) + " de dano (" + itoa(turnsLeft-1) + " turnos restantes)."
	}
	return ""
}

// ─── Status effect combat modifiers ───────────────────────────────────────────

// AttackPenalty returns a multiplier (0.0–1.0) applied to attacker accuracy when
// the set contains accuracy-reducing effects.
func (s *StatusSet) AttackPenalty() float64 {
	if s.Has(StatusBlind) {
		return 0.5
	}
	if s.Has(StatusStun) {
		return 0.0 // stunned = can't attack
	}
	return 1.0
}

// DamageBonus returns a multiplier added on top of base damage.
func (s *StatusSet) DamageBonus() float64 {
	mult := 1.0
	if s.Has(StatusBerserk) {
		mult *= 1.5
	}
	if s.Has(StatusHaste) {
		mult *= 1.2
	}
	return mult
}

// DamageTaken returns a multiplier for incoming damage.
func (s *StatusSet) DamageTaken() float64 {
	mult := 1.0
	if s.Has(StatusShield) || s.Has(StatusProtect) {
		mult *= 0.6
	}
	if s.Has(StatusBerserk) {
		mult *= 1.2 // berserk: more damage taken
	}
	if s.Has(StatusCurse) {
		mult *= 1.25
	}
	return mult
}

// tiny helper so we don't import fmt in this package
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	buf := make([]byte, 0, 12)
	for n > 0 {
		buf = append([]byte{byte('0' + n%10)}, buf...)
		n /= 10
	}
	if neg {
		buf = append([]byte{'-'}, buf...)
	}
	return string(buf)
}
