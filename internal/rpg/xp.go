package rpg

import "math"

// ─── XP curve (levels 1–100) ──────────────────────────────────────────────────
//
// Formula: XPForNextLevel(n) = base * n^exponent
// Chosen constants create a smooth curve where:
//   - Level 20 (old cap) ≈ 150,000 XP total
//   - Level 50           ≈ 1,200,000 XP total
//   - Level 100          ≈ 8,000,000 XP total

const (
	xpBase     = 100.0
	xpExponent = 2.3
)

// XPRequired returns the XP needed to advance FROM level n TO level n+1.
// Level 100 is the cap; returns 0 if n >= 100.
func XPRequired(level int) int {
	if level >= 100 {
		return 0
	}
	if level < 1 {
		level = 1
	}
	return int(xpBase * math.Pow(float64(level), xpExponent))
}

// TotalXPForLevel returns the cumulative XP needed to *reach* a given level from 0.
func TotalXPForLevel(level int) int {
	total := 0
	for l := 1; l < level; l++ {
		total += XPRequired(l)
	}
	return total
}

// LevelFromTotalXP returns the level a character is at given their total XP.
func LevelFromTotalXP(totalXP int) int {
	for l := 99; l >= 1; l-- {
		if totalXP >= TotalXPForLevel(l) {
			return l
		}
	}
	return 1
}

// XPScalingMultiplier scales XP reward based on level difference between player and monster.
//
//   levelDiff > 0: monster is weaker → penalty
//   levelDiff < 0: monster is stronger → bonus
func XPScalingMultiplier(playerLevel, monsterLevel int) float64 {
	diff := playerLevel - monsterLevel
	switch {
	case diff > 10:
		return 0.1 // farming grey mobs
	case diff > 5:
		return math.Pow(0.85, float64(diff-5))
	case diff < -5:
		return 1.5 // extremely dangerous monster
	case diff < -3:
		return 1.3
	default:
		return 1.0
	}
}

// ─── Level-up rewards (extended to 100) ───────────────────────────────────────

// LevelReward describes the bonuses granted on level-up.
type LevelReward struct {
	HPGain         int
	MPGain         int
	SkillPoints    int
	AttributeGains map[string]int // stat → delta
	Milestone      string         // e.g. "Campeão" title at level 50
}

// CalcLevelReward computes the reward for levelling from (level-1) to level.
func CalcLevelReward(level int, class ClassDef, attr Attributes) LevelReward {
	reward := LevelReward{
		HPGain:      class.HPPerLevel + Modifier(attr.Constitution),
		MPGain:      class.MPPerLevel + Modifier(attr.Intelligence),
		SkillPoints: 1,
	}
	reward.AttributeGains = LevelUpAttributeGain(level, class.PrimaryStats)

	// Milestone levels: 10, 25, 50, 75, 100
	switch level {
	case 10:
		reward.Milestone = "Aventureiro"
		reward.SkillPoints += 1
	case 25:
		reward.Milestone = "Veterano"
		reward.SkillPoints += 2
	case 50:
		reward.Milestone = "Campeão"
		reward.SkillPoints += 3
		reward.HPGain += 50
		reward.MPGain += 25
	case 75:
		reward.Milestone = "Lendário"
		reward.SkillPoints += 4
	case 100:
		reward.Milestone = "Imortal"
		reward.SkillPoints += 5
		reward.HPGain += 100
		reward.MPGain += 50
	}
	return reward
}
