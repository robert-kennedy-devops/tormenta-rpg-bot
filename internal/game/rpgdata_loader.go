package game

// rpgdata_loader.go merges all data from the internal/rpgdata package into the
// live game maps (Items, Skills, Monsters, Races, Classes) used by the rest of
// the engine.
//
// Merge strategy:
//   - rpgdata entries are written first; any key that already exists in the
//     hard-coded map is NOT overwritten (legacy data always wins).
//   - This lets hand-crafted items / quest items remain unchanged while the
//     generated 500+ items are added alongside them.
//
// The init() function runs automatically when the game package is imported.

import "github.com/tormenta-bot/internal/rpgdata"

func init() {
	// ── Races ─────────────────────────────────────────────────────────────
	for id, race := range rpgdata.AllRaces {
		if _, exists := Races[id]; !exists {
			Races[id] = race
		}
	}

	// ── Classes ───────────────────────────────────────────────────────────
	for id, class := range rpgdata.AllClasses {
		if _, exists := Classes[id]; !exists {
			Classes[id] = class
		}
	}

	// ── Skills ────────────────────────────────────────────────────────────
	for id, skill := range rpgdata.AllSkills {
		if _, exists := Skills[id]; !exists {
			Skills[id] = skill
		}
	}

	// ── Items ─────────────────────────────────────────────────────────────
	for id, item := range rpgdata.AllItems {
		if _, exists := Items[id]; !exists {
			Items[id] = item
		}
	}

	// ── Monsters ──────────────────────────────────────────────────────────
	for id, monster := range rpgdata.AllMonsters {
		if _, exists := Monsters[id]; !exists {
			Monsters[id] = monster
		}
	}
}
