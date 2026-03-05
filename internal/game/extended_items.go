package game

import "github.com/tormenta-bot/internal/models"

func init() {
	// Materiais para forja/crafting
	Items["mat_forge_stone"] = models.Item{
		ID:          "mat_forge_stone",
		Name:        "Pedra de Forja",
		Description: "Material básico para aprimoramento.",
		Emoji:       "🪨",
		Type:        "material",
		Rarity:      models.RarityCommon,
		Price:       0,
		SellPrice:   8,
	}
	Items["mat_refined_stone"] = models.Item{
		ID:          "mat_refined_stone",
		Name:        "Pedra Refinada",
		Description: "Catalisador de forjas avançadas.",
		Emoji:       "🔷",
		Type:        "material",
		Rarity:      models.RarityUncommon,
		Price:       0,
		SellPrice:   20,
	}
	Items["mat_arcane_essence"] = models.Item{
		ID:          "mat_arcane_essence",
		Name:        "Essência Arcana",
		Description: "Núcleo mágico raro para craft e refinamento.",
		Emoji:       "✨",
		Type:        "material",
		Rarity:      models.RarityRare,
		Price:       0,
		SellPrice:   45,
	}
	Items["mat_monster_shard"] = models.Item{
		ID:          "mat_monster_shard",
		Name:        "Fragmento de Monstro",
		Description: "Fragmento residual de criaturas derrotadas.",
		Emoji:       "🧩",
		Type:        "material",
		Rarity:      models.RarityCommon,
		Price:       0,
		SellPrice:   12,
	}
	Items["mat_black_metal"] = models.Item{
		ID:          "mat_black_metal",
		Name:        "Metal Negro",
		Description: "Liga rara usada em equipamentos especiais.",
		Emoji:       "⛓️",
		Type:        "material",
		Rarity:      models.RarityEpic,
		Price:       0,
		SellPrice:   90,
	}

	// Resultado inicial de crafting
	Items["sword_black"] = models.Item{
		ID:          "sword_black",
		Name:        "Espada Negra",
		Description: "Lâmina sombria forjada com metal negro e essência arcana.",
		Emoji:       "🗡️",
		Type:        "weapon",
		Rarity:      models.RarityEpic,
		Price:       0,
		SellPrice:   780,
		AttackBonus: 52,
		HitBonus:    2,
		MinLevel:    14,
		ClassReq:    "warrior",
		DropWeight:  0,
		Slot:        "weapon",
	}
}
