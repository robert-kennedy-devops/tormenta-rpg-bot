package economy

// ─── Tax system ───────────────────────────────────────────────────────────────
//
// Every player-to-player transaction (market, auction, direct trade) is taxed.
// Tax gold is destroyed (gold sink), never redistributed.

// TaxConfig configures tax rates per transaction type.
type TaxConfig struct {
	MarketTaxRate  float64 // fraction of sale price taken as tax (e.g. 0.05 = 5%)
	AuctionTaxRate float64 // on top of final bid
	TradeTaxRate   float64 // direct player trades
}

// DefaultTaxConfig returns the base tax configuration.
func DefaultTaxConfig() TaxConfig {
	return TaxConfig{
		MarketTaxRate:  0.05,
		AuctionTaxRate: 0.08,
		TradeTaxRate:   0.03,
	}
}

// InflationAdjustedConfig scales tax rates up during inflation to destroy gold faster.
func InflationAdjustedConfig(base TaxConfig, m *EconomyManager) TaxConfig {
	switch m.CurrentInflation() {
	case InflationCritical:
		return TaxConfig{
			MarketTaxRate:  base.MarketTaxRate * 2.0,
			AuctionTaxRate: base.AuctionTaxRate * 2.0,
			TradeTaxRate:   base.TradeTaxRate * 2.0,
		}
	case InflationWarning:
		return TaxConfig{
			MarketTaxRate:  base.MarketTaxRate * 1.5,
			AuctionTaxRate: base.AuctionTaxRate * 1.5,
			TradeTaxRate:   base.TradeTaxRate * 1.5,
		}
	default:
		return base
	}
}

// CalculateTax returns the tax amount (rounded down) for a given transaction.
func CalculateTax(price int, rate float64) int {
	if rate <= 0 || price <= 0 {
		return 0
	}
	tax := int(float64(price) * rate)
	if tax < 1 {
		tax = 1
	}
	return tax
}

// ApplyMarketTax computes seller proceeds and tax for a market sale.
// Returns (sellerReceives, taxAmount).
func ApplyMarketTax(salePrice int, cfg TaxConfig) (sellerReceives int, taxAmount int) {
	taxAmount = CalculateTax(salePrice, cfg.MarketTaxRate)
	sellerReceives = salePrice - taxAmount
	if sellerReceives < 0 {
		sellerReceives = 0
	}
	return
}

// ApplyAuctionTax computes proceeds and tax for a finalised auction.
// Returns (sellerReceives, taxAmount).
func ApplyAuctionTax(finalBid int, cfg TaxConfig) (sellerReceives int, taxAmount int) {
	taxAmount = CalculateTax(finalBid, cfg.AuctionTaxRate)
	sellerReceives = finalBid - taxAmount
	if sellerReceives < 0 {
		sellerReceives = 0
	}
	return
}
