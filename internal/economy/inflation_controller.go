package economy

// ─── Inflation controller ─────────────────────────────────────────────────────
//
// The InflationController calculates dynamic multipliers that are applied to:
//   - gold drop amounts     (reduced during inflation)
//   - gold shop item costs  (increased during inflation — makes sinks more effective)
//   - NPC repair costs      (increased during inflation)
//
// Callers must multiply their base gold values by the multipliers returned here.

// DropRateMultiplier returns a [0.0–1.0] multiplier to apply to gold drops.
// During inflation, drops are reduced; normal = 1.0.
func DropRateMultiplier(m *EconomyManager) float64 {
	switch m.CurrentInflation() {
	case InflationCritical:
		return 0.40 // heavy cut — only 40% of gold dropped
	case InflationWarning:
		return 0.70
	default:
		return 1.0
	}
}

// ShopCostMultiplier returns a [1.0–2.0] multiplier for NPC shop prices.
// Higher prices during inflation act as a gold sink.
func ShopCostMultiplier(m *EconomyManager) float64 {
	switch m.CurrentInflation() {
	case InflationCritical:
		return 1.8
	case InflationWarning:
		return 1.3
	default:
		return 1.0
	}
}

// RepairCostMultiplier returns a multiplier for equipment repair costs.
func RepairCostMultiplier(m *EconomyManager) float64 {
	switch m.CurrentInflation() {
	case InflationCritical:
		return 2.0
	case InflationWarning:
		return 1.5
	default:
		return 1.0
	}
}

// ApplyDropMultiplier applies the current multiplier to a raw gold amount.
// Returns the adjusted amount (minimum 0).
func ApplyDropMultiplier(rawGold int, m *EconomyManager) int {
	adjusted := int(float64(rawGold) * DropRateMultiplier(m))
	if adjusted < 0 {
		return 0
	}
	return adjusted
}

// ─── Gold sink events ─────────────────────────────────────────────────────────

// GoldSinkKind identifies a gold sink mechanism.
type GoldSinkKind string

const (
	SinkRepair      GoldSinkKind = "repair"
	SinkTax         GoldSinkKind = "tax"
	SinkAuctionFee  GoldSinkKind = "auction_fee"
	SinkCraftingCost GoldSinkKind = "crafting"
	SinkInnRest     GoldSinkKind = "inn_rest"
	SinkMarketListing GoldSinkKind = "market_listing"
)

// RecordSink tells the EconomyManager that gold was destroyed by a sink.
// This should be called whenever players spend gold on non-player targets.
func RecordSink(m *EconomyManager, amount int, kind GoldSinkKind) {
	if amount <= 0 {
		return
	}
	m.RemoveGold(int64(amount))
}
