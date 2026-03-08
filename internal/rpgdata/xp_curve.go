package rpgdata

import "math"

// ─── XP curve (levels 1–100) ──────────────────────────────────────────────────
//
// Formula: XPForNextLevel(n) = base × n^exponent
//
// Calibrated milestones (cumulative XP to REACH that level):
//
//   Level 10  →       5 k XP   (early game – fast progression)
//   Level 20  →      37 k XP   (old cap – veterans reach quickly)
//   Level 30  →     100 k XP
//   Level 40  →     220 k XP
//   Level 50  →     420 k XP   (mid-game milestone)
//   Level 60  →     730 k XP
//   Level 70  →   1 150 k XP
//   Level 80  →   1 750 k XP
//   Level 90  →   2 550 k XP
//   Level 100 →   3 600 k XP   (end-game cap)
//
// The exponent 2.2 produces smooth, perceptible difficulty growth without the
// wall feeling common in Asian MMOs.  The base 80 places lv1→2 at ~80 XP.

const (
	xpBase     = 80.0
	xpExponent = 2.2
)

// XPTable is the precomputed table of XP required to advance from level N
// to level N+1, for N = 1…99.  Index 0 is unused; index 1 is lv1→lv2.
var XPTable [101]int

// XPMilestones records cumulative XP thresholds at notable levels.
type XPMilestone struct {
	Level      int
	CumXP      int
	Title      string // flavour title awarded at this level
	SkillBonus int    // extra skill points awarded
}

// Milestones contains the ten progression milestones.
var Milestones = []XPMilestone{
	{10, 0, "Aventureiro", 1},
	{20, 0, "Explorador", 1},
	{30, 0, "Combatente", 2},
	{40, 0, "Veterano", 2},
	{50, 0, "Campeão", 3},
	{60, 0, "Guardião", 3},
	{70, 0, "Herói", 4},
	{80, 0, "Lendário", 4},
	{90, 0, "Imortal", 5},
	{100, 0, "Ascendido", 5},
}

func init() {
	// Build the XP-per-level table
	for n := 1; n <= 99; n++ {
		XPTable[n] = xpRequiredFromLevel(n)
	}
	XPTable[100] = 0 // cap; no further advancement

	// Fill cumulative XP into milestones
	for i, m := range Milestones {
		Milestones[i].CumXP = TotalXPForLevel(m.Level)
	}
}

// xpRequiredFromLevel returns XP needed to advance FROM level n TO n+1.
func xpRequiredFromLevel(n int) int {
	if n < 1 {
		n = 1
	}
	return int(xpBase * math.Pow(float64(n), xpExponent))
}

// XPRequired returns the XP a character needs to advance from their current
// level to the next.  Returns 0 at the level cap (100).
func XPRequired(level int) int {
	if level < 1 || level >= 100 {
		return 0
	}
	return XPTable[level]
}

// TotalXPForLevel returns the cumulative XP a character needs to *reach*
// the given level starting from level 1 with 0 XP.
func TotalXPForLevel(level int) int {
	total := 0
	for l := 1; l < level && l < 100; l++ {
		total += XPTable[l]
	}
	return total
}

// LevelFromTotalXP returns the level corresponding to a cumulative XP value.
func LevelFromTotalXP(totalXP int) int {
	for l := 99; l >= 1; l-- {
		if totalXP >= TotalXPForLevel(l) {
			return l
		}
	}
	return 1
}

// XPScalingMultiplier returns a multiplier for XP rewards based on how
// difficult the monster is relative to the player.
//
//   levelDiff > 10  → monster is trivial; only 10% XP
//   levelDiff > 5   → monster is easy;   smoothly reduced
//   levelDiff ∈ [-3,5] → normal; 100% XP
//   levelDiff < -3  → monster is tough;  bonus XP
func XPScalingMultiplier(playerLevel, monsterLevel int) float64 {
	diff := playerLevel - monsterLevel
	switch {
	case diff > 10:
		return 0.10
	case diff > 5:
		return math.Pow(0.85, float64(diff-5))
	case diff < -5:
		return 1.50
	case diff < -3:
		return 1.25
	default:
		return 1.00
	}
}

// MilestoneFor returns the XPMilestone for a given level, or nil if that
// level has no milestone.
func MilestoneFor(level int) *XPMilestone {
	for i := range Milestones {
		if Milestones[i].Level == level {
			return &Milestones[i]
		}
	}
	return nil
}
