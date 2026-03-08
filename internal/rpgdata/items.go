package rpgdata

import (
	"fmt"

	"github.com/tormenta-bot/internal/models"
)

// ─── Item templates ───────────────────────────────────────────────────────────

// itemTemplate is the blueprint for a single item type before tier/rarity
// scaling is applied.  One template generates up to 30 final items
// (5 tiers × 6 rarities).
type itemTemplate struct {
	idBase   string // prefix; final ID = "<idBase>_t<tier>_<rarity>"
	name     string // base name; tier prefix is prepended automatically
	emoji    string
	slot     string // weapon / armor / helmet / boots / ring / necklace / accessory
	itemType string // matches models.Item.ItemType (string field)
	baseAtk  int    // attack bonus at tier 1 / Common
	baseDef  int    // defense bonus at tier 1 / Common
	baseHP   int
	baseMP   int
	basePrice int // gold at tier 1 / Common
	// rarities controls which rarities are generated for this template.
	// nil = all 6 rarities (Common … Mythic)
	rarities []models.Rarity
}

// tierPrefix is prepended to item names to indicate progression band.
var tierPrefix = [6]string{"", "", "Veterano", "Mestre", "Élite", "Lendário"}

// rarityLabel is appended to item names for rarities above Common.
var rarityLabel = [6]string{"", " (Incomum)", " (Raro)", " (Épico)", " (Lendário)", " (Mítico)"}

// ─── Weapon templates (7 archetypes × 5 tiers × 6 rarities = 210 weapon items)

var weaponTemplates = []itemTemplate{
	// Swords — melee weapon for warrior / paladin / bard
	{
		idBase: "sword", name: "Espada", emoji: "⚔️",
		slot: "weapon", itemType: "weapon",
		baseAtk: 10, baseDef: 2, baseHP: 0, baseMP: 0, basePrice: 50,
	},
	// Axes — melee weapon for warrior / barbarian
	{
		idBase: "axe", name: "Machado", emoji: "🪓",
		slot: "weapon", itemType: "weapon",
		baseAtk: 13, baseDef: 0, baseHP: 0, baseMP: 0, basePrice: 55,
	},
	// Staffs — magic weapon for mage / cleric / druid / necromancer
	{
		idBase: "staff", name: "Cajado", emoji: "🔮",
		slot: "weapon", itemType: "weapon",
		baseAtk: 6, baseDef: 0, baseHP: 0, baseMP: 15, basePrice: 60,
	},
	// Daggers — weapon for rogue / archer / bard
	{
		idBase: "dagger", name: "Adaga", emoji: "🗡️",
		slot: "weapon", itemType: "weapon",
		baseAtk: 8, baseDef: 0, baseHP: 0, baseMP: 0, basePrice: 40,
	},
	// Bows — ranged weapon for archer
	{
		idBase: "bow", name: "Arco", emoji: "🏹",
		slot: "weapon", itemType: "weapon",
		baseAtk: 11, baseDef: 0, baseMP: 0, basePrice: 55,
	},
	// Maces — weapon for cleric / paladin
	{
		idBase: "mace", name: "Maça", emoji: "🪨",
		slot: "weapon", itemType: "weapon",
		baseAtk: 9, baseDef: 1, baseHP: 5, baseMP: 0, basePrice: 50,
	},
	// Scythes — weapon for necromancer / druid
	{
		idBase: "scythe", name: "Foice", emoji: "💀",
		slot: "weapon", itemType: "weapon",
		baseAtk: 12, baseDef: 0, baseHP: 0, baseMP: 10, basePrice: 65,
	},
}

// ─── Armor templates (5 archetypes × 5 tiers × 6 rarities = 150 armor items)

var armorTemplates = []itemTemplate{
	// Heavy chest
	{
		idBase: "plate", name: "Armadura de Placas", emoji: "🛡️",
		slot: "armor", itemType: "armor",
		baseAtk: 0, baseDef: 15, baseHP: 20, baseMP: 0, basePrice: 70,
	},
	// Medium chest
	{
		idBase: "chainmail", name: "Cota de Malha", emoji: "⛓️",
		slot: "armor", itemType: "armor",
		baseAtk: 0, baseDef: 10, baseHP: 12, baseMP: 5, basePrice: 60,
	},
	// Light chest (robes)
	{
		idBase: "robe", name: "Manto Arcano", emoji: "🎭",
		slot: "armor", itemType: "armor",
		baseAtk: 0, baseDef: 5, baseHP: 5, baseMP: 25, basePrice: 65,
	},
	// Leather chest
	{
		idBase: "leather", name: "Couro Reforçado", emoji: "🦺",
		slot: "armor", itemType: "armor",
		baseAtk: 1, baseDef: 8, baseHP: 8, baseMP: 0, basePrice: 55,
	},
	// Druidic / nature armor
	{
		idBase: "nature_armor", name: "Armadura Natural", emoji: "🌿",
		slot: "armor", itemType: "armor",
		baseAtk: 0, baseDef: 7, baseHP: 10, baseMP: 12, basePrice: 60,
	},
}

// ─── Accessory templates (6 archetypes × 5 tiers × 6 rarities = 180 acc items)

var accessoryTemplates = []itemTemplate{
	// Attack ring
	{
		idBase: "ring_atk", name: "Anel de Força", emoji: "💍",
		slot: "ring", itemType: "accessory",
		baseAtk: 5, baseDef: 0, baseHP: 5, baseMP: 0, basePrice: 30,
	},
	// Defense ring
	{
		idBase: "ring_def", name: "Anel de Proteção", emoji: "💍",
		slot: "ring", itemType: "accessory",
		baseAtk: 0, baseDef: 6, baseHP: 8, baseMP: 0, basePrice: 30,
	},
	// Magic necklace
	{
		idBase: "necklace_mp", name: "Colar Místico", emoji: "📿",
		slot: "necklace", itemType: "accessory",
		baseAtk: 0, baseDef: 0, baseHP: 0, baseMP: 20, basePrice: 35,
	},
	// HP necklace
	{
		idBase: "necklace_hp", name: "Amuleto Vital", emoji: "📿",
		slot: "necklace", itemType: "accessory",
		baseAtk: 0, baseDef: 2, baseHP: 20, baseMP: 0, basePrice: 35,
	},
	// Balanced amulet
	{
		idBase: "amulet", name: "Amuleto do Herói", emoji: "🏅",
		slot: "necklace", itemType: "accessory",
		baseAtk: 3, baseDef: 3, baseHP: 10, baseMP: 10, basePrice: 45,
	},
	// Elemental orb (off-hand accessory)
	{
		idBase: "orb", name: "Orbe Elemental", emoji: "🔮",
		slot: "accessory", itemType: "accessory",
		baseAtk: 4, baseDef: 0, baseHP: 0, baseMP: 18, basePrice: 40,
	},
}

// ─── Helmet templates (4 archetypes × 5 tiers × 6 rarities = 120 helmet items)

var helmetTemplates = []itemTemplate{
	{
		idBase: "helm_heavy", name: "Elmo de Ferro", emoji: "⛑️",
		slot: "helmet", itemType: "armor",
		baseAtk: 0, baseDef: 8, baseHP: 10, baseMP: 0, basePrice: 40,
	},
	{
		idBase: "helm_light", name: "Capuz de Couro", emoji: "🎩",
		slot: "helmet", itemType: "armor",
		baseAtk: 1, baseDef: 4, baseHP: 6, baseMP: 0, basePrice: 30,
	},
	{
		idBase: "helm_arcane", name: "Chapéu Arcano", emoji: "🧙",
		slot: "helmet", itemType: "armor",
		baseAtk: 0, baseDef: 2, baseHP: 0, baseMP: 15, basePrice: 35,
	},
	{
		idBase: "helm_nature", name: "Coroa de Folhas", emoji: "🌿",
		slot: "helmet", itemType: "armor",
		baseAtk: 0, baseDef: 4, baseHP: 5, baseMP: 8, basePrice: 35,
	},
}

// ─── Boot templates (3 archetypes × 5 tiers × 6 rarities = 90 boot items)

var bootTemplates = []itemTemplate{
	{
		idBase: "boots_heavy", name: "Botas de Ferro", emoji: "🥾",
		slot: "boots", itemType: "armor",
		baseAtk: 0, baseDef: 5, baseHP: 8, baseMP: 0, basePrice: 30,
	},
	{
		idBase: "boots_light", name: "Botas de Couro", emoji: "👢",
		slot: "boots", itemType: "armor",
		baseAtk: 2, baseDef: 2, baseHP: 4, baseMP: 0, basePrice: 25,
	},
	{
		idBase: "boots_arcane", name: "Sapatilhas Mágicas", emoji: "✨",
		slot: "boots", itemType: "armor",
		baseAtk: 0, baseDef: 1, baseHP: 0, baseMP: 10, basePrice: 28,
	},
}

// ─── Generator ────────────────────────────────────────────────────────────────

// AllItems holds all generated items.  Populated by init().
var AllItems map[string]models.Item

func init() {
	// Rough capacity: (7+5+6+4+3) templates × 5 tiers × 6 rarities = ~750
	AllItems = make(map[string]models.Item, 750)

	allTemplates := make([]itemTemplate, 0, 25)
	allTemplates = append(allTemplates, weaponTemplates...)
	allTemplates = append(allTemplates, armorTemplates...)
	allTemplates = append(allTemplates, accessoryTemplates...)
	allTemplates = append(allTemplates, helmetTemplates...)
	allTemplates = append(allTemplates, bootTemplates...)

	rarities := []models.Rarity{
		models.RarityCommon,
		models.RarityUncommon,
		models.RarityRare,
		models.RarityEpic,
		models.RarityLegendary,
		RarityMythic,
	}

	for _, tmpl := range allTemplates {
		allowedRarities := rarities
		if tmpl.rarities != nil {
			allowedRarities = tmpl.rarities
		}
		for tier := 1; tier <= 5; tier++ {
			for _, r := range allowedRarities {
				item := buildItem(tmpl, tier, r)
				AllItems[item.ID] = item
			}
		}
	}
}

func buildItem(tmpl itemTemplate, tier int, r models.Rarity) models.Item {
	id := fmt.Sprintf("%s_t%d_r%d", tmpl.idBase, tier, int(r))

	// Name construction
	name := tmpl.name
	if tier >= 3 && tier <= 5 {
		name = tierPrefix[tier] + " " + name
	}
	name += rarityLabel[int(r)]

	return models.Item{
		ID:          id,
		Name:        name,
		Emoji:       tmpl.emoji,
		Type:        tmpl.itemType,
		Rarity:      r,
		Slot:        tmpl.slot,
		MinLevel:    ItemTiers[tier-1].MinLevel,
		AttackBonus: ScaleItemStat(tmpl.baseAtk, tier, r),
		DefenseBonus: ScaleItemStat(tmpl.baseDef, tier, r),
		HPBonus:     ScaleItemStat(tmpl.baseHP, tier, r),
		MPBonus:     ScaleItemStat(tmpl.baseMP, tier, r),
		Price:       ScaleItemPrice(tmpl.basePrice, tier, r),
		DropWeight:  DropWeightForRarity(r),
	}
}

// ─── Lookup helpers ───────────────────────────────────────────────────────────

// ItemsForSlot returns all items matching the given slot string.
func ItemsForSlot(slot string) []models.Item {
	out := make([]models.Item, 0, 30)
	for _, it := range AllItems {
		if it.Slot == slot {
			out = append(out, it)
		}
	}
	return out
}

// ItemsForTier returns all items whose MinLevel falls within the given tier.
func ItemsForTier(tier int) []models.Item {
	if tier < 1 || tier > 5 {
		return nil
	}
	t := ItemTiers[tier-1]
	out := make([]models.Item, 0, 30)
	for _, it := range AllItems {
		if it.MinLevel >= t.MinLevel && it.MinLevel <= t.MaxLevel {
			out = append(out, it)
		}
	}
	return out
}

// ItemCount returns the total number of generated items.
func ItemCount() int { return len(AllItems) }
