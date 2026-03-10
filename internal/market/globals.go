package market

import (
	"database/sql"

	"github.com/tormenta-bot/internal/economy"
)

// Global singletons — start with in-memory fallback, switched to DB by InitDB.
var (
	GlobalStore        ListingStore = NewMemListingStore()
	GlobalAuctionStore AuctionStore = NewMemAuctionStore()
	GlobalService                   = NewMarketService(GlobalStore, GlobalAuctionStore, economy.Global)
)

// InitDB replaces the in-memory stores with DB-backed implementations.
// Must be called once after database.Connect().
func InitDB(db *sql.DB) {
	GlobalStore = NewDBListingStore(db)
	GlobalAuctionStore = NewDBAuctionStore(db)
	GlobalService = NewMarketService(GlobalStore, GlobalAuctionStore, economy.Global)
}
