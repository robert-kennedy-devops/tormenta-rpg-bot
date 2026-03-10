// safe_lookup.go — nil-safe accessors for the global game data maps.
//
// The raw maps (Races, Classes, Skills, Items, Monsters) are populated at
// startup and treated as read-only at runtime.  These helpers replace direct
// map indexing (which returns a zero-value struct silently) with explicit
// error returns so callers can handle missing data gracefully.
package game

import (
	"fmt"

	"github.com/tormenta-bot/internal/models"
)

// ─── Race ──────────────────────────────────────────────────────────────────────

// GetRace returns the Race for the given ID or an error if it doesn't exist.
func GetRace(id string) (models.Race, error) {
	r, ok := Races[id]
	if !ok {
		return models.Race{}, fmt.Errorf("race %q not found", id)
	}
	return r, nil
}

// RaceOrDefault returns the Race or a sensible fallback (human).
func RaceOrDefault(id string) models.Race {
	if r, ok := Races[id]; ok {
		return r
	}
	return Races["human"]
}

// ─── Class ────────────────────────────────────────────────────────────────────

// GetClass returns the Class for the given ID or an error.
func GetClass(id string) (models.Class, error) {
	c, ok := Classes[id]
	if !ok {
		return models.Class{}, fmt.Errorf("class %q not found", id)
	}
	return c, nil
}

// ClassOrDefault returns the Class or a sensible fallback (warrior).
func ClassOrDefault(id string) models.Class {
	if c, ok := Classes[id]; ok {
		return c
	}
	return Classes["warrior"]
}

// ─── Monster ──────────────────────────────────────────────────────────────────

// GetMonster returns the Monster for the given ID or an error.
// Use this instead of Monsters[id] to avoid nil-pointer panics when the ID
// is stale (e.g. a saved character whose monster was removed from data.go).
func GetMonster(id string) (models.Monster, error) {
	m, ok := Monsters[id]
	if !ok {
		return models.Monster{}, fmt.Errorf("monster %q not found", id)
	}
	return m, nil
}

// ─── Skill ────────────────────────────────────────────────────────────────────

// GetSkill returns the legacy Skill struct for the given ID or an error.
func GetSkill(id string) (models.Skill, error) {
	s, ok := Skills[id]
	if !ok {
		return models.Skill{}, fmt.Errorf("skill %q not found", id)
	}
	return s, nil
}

// ─── Item ─────────────────────────────────────────────────────────────────────

// GetItem returns the Item template for the given ID or an error.
func GetItem(id string) (models.Item, error) {
	it, ok := Items[id]
	if !ok {
		return models.Item{}, fmt.Errorf("item %q not found", id)
	}
	return it, nil
}
