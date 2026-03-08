package rpgdata

import "github.com/tormenta-bot/internal/models"

// ─── Classes ──────────────────────────────────────────────────────────────────
//
// Ten classes cover the Tormenta RPG spectrum from melee bruiser to arcane
// spellcaster.  Each class entry extends or replaces the game.Classes map
// entry of the same ID so existing characters keep their class without changes
// to the DB schema.
//
// Scaling law (per level beyond 1):
//
//   HPPerLevel / MPPerLevel drive the raw survivability trajectory.
//   A Bárbaro reaches ~1 400 HP at level 100; a Arcanista reaches ~650 HP.
//   Stats at milestone levels 20/50/100 are printed in the class description.
//
// Role strings: "tank" | "dps" | "support" | "mage" | "ranged" | "hybrid"

// AllClasses contains all playable classes keyed by their canonical ID.
// Merged into game.Classes by the loader in game/rpgdata_loader.go.
var AllClasses = map[string]models.Class{
	// ── Existing classes (updated stats for lv100 balance) ────────────────────
	"warrior": {
		ID: "warrior", Name: "Guerreiro", Emoji: "⚔️",
		Description: "Mestre das armas e armaduras pesadas. Tanque inigualável.",
		BaseHP: 80, BaseMP: 20, HPPerLevel: 14, MPPerLevel: 3,
		BaseAttack: 12, BaseDefense: 10,
		PrimaryStats: []string{"strength", "constitution"}, Role: "tank",
	},
	"mage": {
		ID: "mage", Name: "Arcanista", Emoji: "🧙",
		Description: "Manipula forças arcanas com maestria. Maior dano mágico do jogo.",
		BaseHP: 45, BaseMP: 90, HPPerLevel: 6, MPPerLevel: 14,
		BaseAttack: 4, BaseDefense: 3,
		PrimaryStats: []string{"intelligence", "wisdom"}, Role: "mage",
	},
	"rogue": {
		ID: "rogue", Name: "Ladino", Emoji: "🗡️",
		Description: "Ágil e letal nas sombras. Especialista em ataques críticos.",
		BaseHP: 55, BaseMP: 45, HPPerLevel: 8, MPPerLevel: 6,
		BaseAttack: 10, BaseDefense: 5,
		PrimaryStats: []string{"dexterity", "charisma"}, Role: "dps",
	},
	"archer": {
		ID: "archer", Name: "Caçador", Emoji: "🏹",
		Description: "Precisão e velocidade a longa distância. Mestre do rastreamento.",
		BaseHP: 60, BaseMP: 40, HPPerLevel: 8, MPPerLevel: 5,
		BaseAttack: 9, BaseDefense: 6,
		PrimaryStats: []string{"dexterity", "wisdom"}, Role: "ranged",
	},
	"paladin": {
		ID: "paladin", Name: "Paladino", Emoji: "🛡️",
		Description: "Guerreiro sagrado que equilibra força e proteção divina.",
		BaseHP: 75, BaseMP: 55, HPPerLevel: 11, MPPerLevel: 8,
		BaseAttack: 10, BaseDefense: 9,
		PrimaryStats: []string{"strength", "charisma"}, Role: "tank",
	},
	"cleric": {
		ID: "cleric", Name: "Clérigo", Emoji: "✝️",
		Description: "Canal divino de cura e proteção. Suporte imprescindível.",
		BaseHP: 55, BaseMP: 75, HPPerLevel: 7, MPPerLevel: 11,
		BaseAttack: 6, BaseDefense: 7,
		PrimaryStats: []string{"wisdom", "constitution"}, Role: "support",
	},
	"barbarian": {
		ID: "barbarian", Name: "Bárbaro", Emoji: "🪓",
		Description: "Guerreiro selvagem movido por fúria incontrolável. Mais HP do jogo.",
		BaseHP: 100, BaseMP: 15, HPPerLevel: 16, MPPerLevel: 2,
		BaseAttack: 14, BaseDefense: 7,
		PrimaryStats: []string{"strength", "constitution"}, Role: "dps",
	},
	"bard": {
		ID: "bard", Name: "Bardo", Emoji: "🎵",
		Description: "Artista versátil que inspira aliados e perturba inimigos.",
		BaseHP: 52, BaseMP: 60, HPPerLevel: 7, MPPerLevel: 9,
		BaseAttack: 7, BaseDefense: 5,
		PrimaryStats: []string{"charisma", "dexterity"}, Role: "support",
	},

	// ── New classes (lv100 expansion) ─────────────────────────────────────────
	"druid": {
		ID: "druid", Name: "Druida", Emoji: "🌿",
		Description: "Guardião da natureza. Transforma-se em animais e controla elementos.",
		BaseHP: 60, BaseMP: 70, HPPerLevel: 9, MPPerLevel: 10,
		BaseAttack: 7, BaseDefense: 6,
		PrimaryStats: []string{"wisdom", "constitution"}, Role: "hybrid",
	},
	"necromancer": {
		ID: "necromancer", Name: "Necromante", Emoji: "💀",
		Description: "Mestre da morte que anima mortos-vivos e drena a vida dos inimigos.",
		BaseHP: 48, BaseMP: 95, HPPerLevel: 6, MPPerLevel: 15,
		BaseAttack: 5, BaseDefense: 4,
		PrimaryStats: []string{"intelligence", "wisdom"}, Role: "mage",
	},
}

// PrimaryStatBonus returns the per-level attribute gain deltas for a class's
// primary stats, used by ApplyLevelUp-style functions.
// Called at milestones (every 5 levels) to grow the class's defining stats.
func PrimaryStatBonus(classID string) map[string]int {
	c, ok := AllClasses[classID]
	if !ok {
		return nil
	}
	gains := map[string]int{}
	for _, stat := range c.PrimaryStats {
		switch c.Role {
		case "tank":
			gains[stat] = 2
		case "dps":
			gains[stat] = 3
		case "mage":
			gains[stat] = 2
		default:
			gains[stat] = 2
		}
	}
	return gains
}
