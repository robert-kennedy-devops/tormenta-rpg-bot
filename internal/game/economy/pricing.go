package economy

import (
	"math"

	"github.com/tormenta-bot/internal/database"
)

// DynamicPrice applies a bounded multiplier using usage statistics.
// Keeps backward compatibility by returning base when no stats exist.
func DynamicPrice(itemID string, base int) int {
	if base <= 0 || itemID == "" {
		return base
	}
	usage, err := database.GetItemUsage(itemID)
	if err != nil || usage == 0 {
		return base
	}

	// Smooth growth: every 100 purchases => +1%, capped at +30%.
	growth := math.Min(0.30, float64(usage)/10000.0)
	mult := 1.0 + growth
	price := int(math.Round(float64(base) * mult))
	if price < 1 {
		price = 1
	}
	return price
}
