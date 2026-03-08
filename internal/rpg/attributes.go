package rpg

import "math"

// ─── Core attributes ──────────────────────────────────────────────────────────

// Attributes holds the six D&D-style core stats for a character.
type Attributes struct {
	Strength     int
	Dexterity    int
	Constitution int
	Intelligence int
	Wisdom       int
	Charisma     int
}

// Modifier returns the standard D&D modifier: (stat-10)/2, floored.
func Modifier(stat int) int {
	return int(math.Floor(float64(stat-10) / 2.0))
}

// ─── Derived stats ────────────────────────────────────────────────────────────

// DerivedStats are computed from core attributes + class + level.
type DerivedStats struct {
	MaxHP      int
	MaxMP      int
	Attack     int
	Defense    int
	MagicAtk   int
	MagicDef   int
	Speed      int
	Initiative int
	CritChance int // percent (0-100)
	Evasion    int // flat bonus to CA against physical
}

// Compute calculates DerivedStats for a character given their attributes, class and level.
func Compute(attr Attributes, class ClassDef, level int) DerivedStats {
	hp, mp, atk, def := class.StatsAtLevel(level)
	conMod := Modifier(attr.Constitution)
	intMod := Modifier(attr.Intelligence)
	strMod := Modifier(attr.Strength)
	dexMod := Modifier(attr.Dexterity)
	wisMod := Modifier(attr.Wisdom)

	return DerivedStats{
		MaxHP:      hp + conMod*level,
		MaxMP:      mp + intMod*level,
		Attack:     atk + strMod,
		Defense:    def + conMod,
		MagicAtk:   class.BaseMagAtk + intMod + level/2,
		MagicDef:   class.BaseMagDef + wisMod + level/3,
		Speed:      class.BaseSpd + dexMod,
		Initiative: dexMod + wisMod,
		CritChance: 5 + dexMod,                    // base 5%, +dex mod
		Evasion:    dexMod,
	}
}

// ─── Attribute point allocation ───────────────────────────────────────────────

// PointBuy validates a set of proposed attributes against the point-buy budget.
// Budget: 27 points; each stat starts at 8; costs: 8→13 = 1pt/stat, 14 = 2pt, 15 = 3pt.
func PointBuyCost(stat int) int {
	switch {
	case stat <= 13:
		return stat - 8
	case stat == 14:
		return 7
	case stat == 15:
		return 9
	default:
		return 99 // illegal above 15 at creation
	}
}

// TotalPointBuyCost returns the total cost of the proposed attribute set.
func TotalPointBuyCost(a Attributes) int {
	return PointBuyCost(a.Strength) +
		PointBuyCost(a.Dexterity) +
		PointBuyCost(a.Constitution) +
		PointBuyCost(a.Intelligence) +
		PointBuyCost(a.Wisdom) +
		PointBuyCost(a.Charisma)
}

// ApplyRaceBonuses adds race bonuses to attributes.
func ApplyRaceBonuses(a Attributes, r RaceDef) Attributes {
	a.Strength += r.BonusStr
	a.Dexterity += r.BonusDex
	a.Constitution += r.BonusCon
	a.Intelligence += r.BonusInt
	a.Wisdom += r.BonusWis
	a.Charisma += r.BonusCha
	return a
}

// LevelUpAttributeGain returns bonus attribute points at the given level.
// Every 4 levels → +1 to two primary stats; every 10 levels → +1 to all stats.
func LevelUpAttributeGain(level int, primaryStats []string) map[string]int {
	gains := make(map[string]int)
	if level%10 == 0 {
		for _, s := range []string{"strength", "dexterity", "constitution", "intelligence", "wisdom", "charisma"} {
			gains[s] = 1
		}
		return gains
	}
	if level%4 == 0 {
		for _, s := range primaryStats {
			gains[s] = 1
		}
	}
	return gains
}
