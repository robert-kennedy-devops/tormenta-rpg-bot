package market

import "github.com/tormenta-bot/internal/economy"

// Global singletons — backed by in-memory stores for now.
var (
	GlobalStore        ListingStore  = NewMemListingStore()
	GlobalAuctionStore AuctionStore  = NewMemAuctionStore()
	GlobalService                    = NewMarketService(GlobalStore, GlobalAuctionStore, economy.Global)
)
