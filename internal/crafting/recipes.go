package crafting

import "github.com/tormenta-bot/internal/items"

const RecipeBlackSword = "recipe_black_sword"

var DefaultRecipes = map[string]Recipe{
	RecipeBlackSword: {
		ID:           RecipeBlackSword,
		Name:         "Espada Negra",
		ResultItemID: "sword_black",
		ResultQty:    1,
		Materials: map[string]int{
			items.MaterialBlackMetal:    3,
			items.MaterialArcaneEssence: 1,
		},
	},
}
