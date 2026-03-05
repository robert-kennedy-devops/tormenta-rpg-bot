package items

import "github.com/tormenta-bot/internal/models"

const (
	MaterialForgeStone    = "mat_forge_stone"
	MaterialRefinedStone  = "mat_refined_stone"
	MaterialArcaneEssence = "mat_arcane_essence"
	MaterialMonsterShard  = "mat_monster_shard"
	MaterialBlackMetal    = "mat_black_metal"
)

// MaterialTemplates are base material definitions for forge/crafting systems.
var MaterialTemplates = map[string]ItemTemplate{
	MaterialForgeStone: {
		ID:          MaterialForgeStone,
		Name:        "Pedra de Forja",
		Description: "Material básico para aprimoramento de equipamentos.",
		Emoji:       "🪨",
		Type:        "material",
		Rarity:      models.RarityCommon,
	},
	MaterialRefinedStone: {
		ID:          MaterialRefinedStone,
		Name:        "Pedra Refinada",
		Description: "Catalisa forjas avançadas.",
		Emoji:       "🔷",
		Type:        "material",
		Rarity:      models.RarityUncommon,
	},
	MaterialArcaneEssence: {
		ID:          MaterialArcaneEssence,
		Name:        "Essência Arcana",
		Description: "Núcleo de energia mágica para craft e reforço.",
		Emoji:       "✨",
		Type:        "material",
		Rarity:      models.RarityRare,
	},
	MaterialMonsterShard: {
		ID:          MaterialMonsterShard,
		Name:        "Fragmento de Monstro",
		Description: "Fragmento residual obtido de criaturas.",
		Emoji:       "🧩",
		Type:        "material",
		Rarity:      models.RarityCommon,
	},
	MaterialBlackMetal: {
		ID:          MaterialBlackMetal,
		Name:        "Metal Negro",
		Description: "Liga densa usada em armas e armaduras especiais.",
		Emoji:       "⛓️",
		Type:        "material",
		Rarity:      models.RarityEpic,
	},
}
