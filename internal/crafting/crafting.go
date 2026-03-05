package crafting

import "fmt"

type Recipe struct {
	ID           string
	Name         string
	ResultItemID string
	ResultQty    int
	Materials    map[string]int // itemID -> qty
}

func (r Recipe) Validate() error {
	if r.ID == "" || r.ResultItemID == "" {
		return fmt.Errorf("recipe id/result cannot be empty")
	}
	if r.ResultQty < 1 {
		return fmt.Errorf("invalid result qty: %d", r.ResultQty)
	}
	if len(r.Materials) == 0 {
		return fmt.Errorf("recipe has no materials")
	}
	for itemID, qty := range r.Materials {
		if itemID == "" || qty < 1 {
			return fmt.Errorf("invalid material %q qty=%d", itemID, qty)
		}
	}
	return nil
}

func CanCraft(inv map[string]int, recipe Recipe) bool {
	if err := recipe.Validate(); err != nil {
		return false
	}
	for itemID, need := range recipe.Materials {
		if inv[itemID] < need {
			return false
		}
	}
	return true
}

// ConsumeMaterials returns a copied map with recipe materials removed.
func ConsumeMaterials(inv map[string]int, recipe Recipe) (map[string]int, error) {
	if !CanCraft(inv, recipe) {
		return nil, fmt.Errorf("insufficient materials for recipe %s", recipe.ID)
	}
	next := make(map[string]int, len(inv))
	for k, v := range inv {
		next[k] = v
	}
	for itemID, need := range recipe.Materials {
		next[itemID] -= need
		if next[itemID] <= 0 {
			delete(next, itemID)
		}
	}
	return next, nil
}
