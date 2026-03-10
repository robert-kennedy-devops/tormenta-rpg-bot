// Package skills defines the full expanded skill set for every class.
//
// It exposes AllSkills() which returns every models.Skill across all 8 classes
// (120 new skills total: 15 per class).  These are registered into game.Skills
// via an init() in game/skill_register.go so the Telegram skill menus and
// combat system can find them without touching existing code.
//
// Legacy skills (48 hardcoded in game/data.go) are preserved unchanged; new
// skills complement them giving 27 skills for the 4 existing classes and 15
// skills for the 4 new classes (barbarian, paladin, cleric, bard).
package skills

import "github.com/tormenta-bot/internal/models"

// SkillRole constants — mirror rpg.SkillRole but live here to avoid import cycles.
const (
	RoleDirect  = "DIRECT_DAMAGE"
	RoleAoE     = "AOE"
	RoleDoT     = "DOT"
	RoleBuff    = "BUFF"
	RoleDebuff  = "DEBUFF"
	RoleControl = "CONTROL"
	RoleHeal    = "HEAL"
	RoleUtility = "UTILITY"
	RoleSummon  = "SUMMON"
	RolePassive = "PASSIVE"
	RoleUlt     = "ULTIMATE"
)

// AllSkills returns the full new-skill list for all 8 classes.
func AllSkills() []models.Skill {
	var out []models.Skill
	out = append(out, WarriorSkills()...)
	out = append(out, MageSkills()...)
	out = append(out, RogueSkills()...)
	out = append(out, ArcherSkills()...)
	out = append(out, BarbarianSkills()...)
	out = append(out, PaladinSkills()...)
	out = append(out, ClericSkills()...)
	out = append(out, BardSkills()...)
	return out
}
