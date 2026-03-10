// skill_register.go — loads the expanded skill set (120 new skills across 8 classes)
// into the global game.Skills map via init().
//
// The legacy 48 skills in data.go are preserved unchanged.  This file only ADDS
// new skills; it never overwrites existing ones (matching the same strategy used
// by rpgdata_loader.go).
package game

import "github.com/tormenta-bot/internal/game/skills"

func init() {
	for _, sk := range skills.AllSkills() {
		if _, exists := Skills[sk.ID]; !exists {
			Skills[sk.ID] = sk
		}
	}
}
