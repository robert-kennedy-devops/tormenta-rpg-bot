// Package rpgdata is the central data library for Tormenta-bot's level-1-to-100
// RPG progression system. It provides generated datasets for items, monsters,
// skills, races and classes, all expressed in terms of the existing models
// package so that callers do not need to convert between types.
//
// # Architecture
//
//   rpgdata (pure data, no business logic)
//       ↓  imported by
//   game/rpgdata_loader.go   (merges maps into game.Items / game.Skills / game.Monsters)
//       ↓  consumed by
//   handlers / combat / drops / dungeon  (unchanged callers)
//
// # Scaling philosophy
//
// Three orthogonal axes control every number in this package:
//
//   Tier   (1–5) — the progression band, driven by level.
//   Rarity (0–5) — how exceptional the item/skill is within its tier.
//   Level  (1–100) — fine-grained power inside a tier.
//
// Monsters follow only Tier and Level.  Skills follow Tier + Class only.
//
// # Adding new content
//
// Items:   add a new entry to weaponTemplates / armorTemplates / accessoryTemplates
//          in items.go — the generator creates all tier/rarity variants automatically.
//
// Monsters: add a new MonsterArchetype to monsterArchetypes in monsters.go.
//
// Skills:  add a new SkillDef to classDefs[<class>].skills in skills.go.
package rpgdata

import "github.com/tormenta-bot/internal/models"

// ─── Extra rarity constant (extends models.Rarity without modifying models) ────

// RarityMythic is the sixth rarity tier, above Legendary.
// Emoji: 🔴  Name: "Mítico"
const RarityMythic models.Rarity = 5

// RarityEmoji returns the coloured circle emoji for any rarity value,
// including the Mythic extension defined in this package.
func RarityEmoji(r models.Rarity) string {
	if r == RarityMythic {
		return "🔴"
	}
	return r.Emoji()
}

// RarityName returns the Portuguese display name for any rarity value.
func RarityName(r models.Rarity) string {
	if r == RarityMythic {
		return "Mítico"
	}
	return r.Name()
}

// ─── Tier definitions ─────────────────────────────────────────────────────────

// Tier groups levels into progression bands.  Items, monsters and skills each
// carry a Tier number so systems can quickly determine appropriate content.
type Tier struct {
	Number   int    // 1–5 for items/monsters; 1–4 for skills
	Name     string
	Emoji    string
	MinLevel int
	MaxLevel int
}

// ItemTiers are the five item progression bands.
var ItemTiers = [5]Tier{
	{1, "Aprendiz", "🪨", 1, 20},
	{2, "Veterano", "🔩", 21, 40},
	{3, "Mestre", "⚙️", 41, 60},
	{4, "Élite", "💎", 61, 80},
	{5, "Lendário", "🌌", 81, 100},
}

// SkillTiers are the four skill progression bands (Tormenta-inspired).
var SkillTiers = [4]Tier{
	{1, "Iniciante", "📖", 1, 20},
	{2, "Adepto", "📚", 21, 40},
	{3, "Especialista", "🎓", 41, 70},
	{4, "Mestre", "🏆", 71, 100},
}

// TierFor returns the ItemTier that covers the given level (1–100).
// Returns ItemTiers[4] (Tier 5) for levels above 80.
func TierFor(level int) Tier {
	for _, t := range ItemTiers {
		if level >= t.MinLevel && level <= t.MaxLevel {
			return t
		}
	}
	return ItemTiers[4]
}

// SkillTierFor returns the SkillTier that covers the given level.
func SkillTierFor(level int) Tier {
	for _, t := range SkillTiers {
		if level >= t.MinLevel && level <= t.MaxLevel {
			return t
		}
	}
	return SkillTiers[3]
}

// ─── Scaling helpers ──────────────────────────────────────────────────────────

// tierItemMult maps tier number (1–5) to a base stat multiplier.
// Chosen so that Tier-5 gear is ~14× stronger than Tier-1.
var tierItemMult = [6]float64{0, 1.0, 2.5, 5.0, 9.0, 14.0}

// rarityItemMult maps models.Rarity (0–5) to a stat bonus multiplier.
var rarityItemMult = [6]float64{1.0, 1.30, 1.70, 2.30, 3.20, 4.50}

// rarityDropWeight maps models.Rarity to a weighted drop probability.
// Common drops most often; Mythic almost never.
var rarityDropWeight = [6]int{60, 25, 10, 4, 1, 0} // Mythic is hand-placed only

// ScaleItemStat applies tier and rarity multipliers to a base stat value.
func ScaleItemStat(base, tier int, rarity models.Rarity) int {
	if tier < 1 || tier > 5 {
		tier = 1
	}
	rarityIdx := int(rarity)
	if rarityIdx > 5 {
		rarityIdx = 5
	}
	v := float64(base) * tierItemMult[tier] * rarityItemMult[rarityIdx]
	if v < 1 && base > 0 {
		return 1
	}
	return int(v)
}

// ScaleItemPrice computes the gold price of an item given base price, tier, rarity.
// Price grows faster than stats to create economic tension.
func ScaleItemPrice(basePrice, tier int, rarity models.Rarity) int {
	if tier < 1 {
		tier = 1
	}
	rarityIdx := int(rarity)
	if rarityIdx > 5 {
		rarityIdx = 5
	}
	// Quadratic tier scaling + rarity premium
	tierFactor := float64(tier * tier)
	v := float64(basePrice) * tierFactor * rarityItemMult[rarityIdx]
	if v < 1 {
		return 1
	}
	return int(v)
}

// ScaleMonsterStat scales a monster's base stat to a target level.
// Uses a smooth power curve so lv100 monsters are ~20× lv1 monsters.
func ScaleMonsterStat(baseStat, baseLevel, targetLevel int) int {
	if baseLevel < 1 {
		baseLevel = 1
	}
	if targetLevel < 1 {
		targetLevel = 1
	}
	ratio := float64(targetLevel) / float64(baseLevel)
	// Exponent 1.5 gives a moderate super-linear growth (not as steep as quadratic)
	scaled := float64(baseStat) * pow15(ratio)
	if scaled < 1 {
		return 1
	}
	return int(scaled)
}

// pow15 computes x^1.5 using integer-friendly approximation.
func pow15(x float64) float64 {
	return x * sqrtApprox(x)
}

// sqrtApprox uses Babylonian iteration for a fast float sqrt.
func sqrtApprox(n float64) float64 {
	if n <= 0 {
		return 0
	}
	x := n
	for i := 0; i < 8; i++ {
		x = (x + n/x) / 2
	}
	return x
}

// DropWeightForRarity returns the weighted drop probability for a rarity class.
// Used by the item generator to fill DropWeight on generated items.
func DropWeightForRarity(r models.Rarity) int {
	idx := int(r)
	if idx > 5 {
		idx = 5
	}
	return rarityDropWeight[idx]
}
