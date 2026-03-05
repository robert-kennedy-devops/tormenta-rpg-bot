package crafting

import "testing"

func TestCanCraftBlackSword(t *testing.T) {
	r := DefaultRecipes[RecipeBlackSword]
	inv := map[string]int{
		"mat_black_metal":    3,
		"mat_arcane_essence": 1,
	}
	if !CanCraft(inv, r) {
		t.Fatal("expected craftable recipe")
	}
}

func TestConsumeMaterials(t *testing.T) {
	r := DefaultRecipes[RecipeBlackSword]
	inv := map[string]int{
		"mat_black_metal":    4,
		"mat_arcane_essence": 1,
	}
	next, err := ConsumeMaterials(inv, r)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if next["mat_black_metal"] != 1 {
		t.Fatalf("expected remaining metal=1, got %d", next["mat_black_metal"])
	}
	if _, ok := next["mat_arcane_essence"]; ok {
		t.Fatal("expected arcane essence to be consumed")
	}
}
