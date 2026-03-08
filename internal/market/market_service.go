package market

import (
	"fmt"
	"time"

	"github.com/tormenta-bot/internal/economy"
)

const (
	ListingDuration = 48 * time.Hour // listings expire after 48h
	AuctionDuration = 24 * time.Hour // auctions last 24h by default
	MaxListingsPerPlayer = 10
)

// ─── MarketService ────────────────────────────────────────────────────────────

// MarketService orchestrates all marketplace operations.
type MarketService struct {
	listings ListingStore
	auctions AuctionStore
	economy  *economy.EconomyManager
	taxCfg   economy.TaxConfig
}

// NewMarketService creates a fully initialised service.
func NewMarketService(ls ListingStore, as AuctionStore, em *economy.EconomyManager) *MarketService {
	return &MarketService{
		listings: ls,
		auctions: as,
		economy:  em,
		taxCfg:   economy.DefaultTaxConfig(),
	}
}

// ─── Listing operations ───────────────────────────────────────────────────────

// PostListing creates a new market listing.
// The caller must have already verified the player owns the item.
func (s *MarketService) PostListing(sellerID int64, sellerName, itemID, itemName, itemEmoji string, qty, unitPrice int) (*Listing, error) {
	if qty <= 0 {
		return nil, fmt.Errorf("quantidade inválida")
	}
	if unitPrice <= 0 {
		return nil, fmt.Errorf("preço inválido")
	}
	// Check listing cap
	existing, err := s.listings.ListBySeller(sellerID)
	if err != nil {
		return nil, err
	}
	active := 0
	for _, l := range existing {
		if l.Status == ListingActive {
			active++
		}
	}
	if active >= MaxListingsPerPlayer {
		return nil, fmt.Errorf("limite de %d listagens ativas atingido", MaxListingsPerPlayer)
	}

	now := time.Now()
	l := &Listing{
		SellerID:   sellerID,
		SellerName: sellerName,
		ItemID:     itemID,
		ItemName:   itemName,
		ItemEmoji:  itemEmoji,
		Quantity:   qty,
		UnitPrice:  unitPrice,
		TotalPrice: qty * unitPrice,
		Status:     ListingActive,
		CreatedAt:  now,
		ExpiresAt:  now.Add(ListingDuration),
	}

	// Listing fee (gold sink)
	fee := economy.CalculateTax(l.TotalPrice, 0.01) // 1% listing fee
	economy.RecordSink(s.economy, fee, economy.SinkMarketListing)

	if err := s.listings.Save(l); err != nil {
		return nil, err
	}
	return l, nil
}

// BuyListing completes a purchase.
// Returns (sellerProceeds, taxAmount, error).
// Caller must deduct gold from buyer, add sellerProceeds to seller, remove item from seller.
func (s *MarketService) BuyListing(listingID, buyerID int64) (sellerProceeds, taxAmount int, listing *Listing, err error) {
	listing, err = s.listings.GetByID(listingID)
	if err != nil {
		return
	}
	if listing.Status != ListingActive {
		err = ErrListingNotActive
		return
	}
	if time.Now().After(listing.ExpiresAt) {
		listing.Status = ListingExpired
		_ = s.listings.Update(listing)
		err = ErrListingExpired
		return
	}
	if listing.SellerID == buyerID {
		err = ErrSelfPurchase
		return
	}

	cfg := economy.InflationAdjustedConfig(s.taxCfg, s.economy)
	sellerProceeds, taxAmount = economy.ApplyMarketTax(listing.TotalPrice, cfg)

	now := time.Now()
	listing.Status = ListingSold
	listing.SoldAt = &now
	listing.BuyerID = &buyerID
	if saveErr := s.listings.Update(listing); saveErr != nil {
		err = saveErr
		return
	}

	// Update gold flows in economy manager
	s.economy.AddGold(int64(sellerProceeds))  // seller gains
	economy.RecordSink(s.economy, taxAmount, economy.SinkTax) // tax destroyed
	return
}

// CancelListing withdraws a listing (no refund of listing fee).
func (s *MarketService) CancelListing(listingID, sellerID int64) error {
	l, err := s.listings.GetByID(listingID)
	if err != nil {
		return err
	}
	if l.SellerID != sellerID {
		return fmt.Errorf("você não é o dono desta listagem")
	}
	if l.Status != ListingActive {
		return ErrListingNotActive
	}
	l.Status = ListingCancelled
	return s.listings.Update(l)
}

// SearchListings returns active listings filtered by item ID.
func (s *MarketService) SearchListings(itemID string, page int) ([]*Listing, error) {
	const pageSize = 10
	offset := page * pageSize
	return s.listings.ListActive(itemID, pageSize, offset)
}

// ─── Auction operations ───────────────────────────────────────────────────────

// CreateAuction opens a new auction.
func (s *MarketService) CreateAuction(sellerID int64, sellerName, itemID, itemName, itemEmoji string, qty, startPrice, buyItNow int) (*Auction, error) {
	if startPrice <= 0 {
		return nil, fmt.Errorf("preço inicial inválido")
	}
	now := time.Now()
	a := &Auction{
		SellerID:   sellerID,
		SellerName: sellerName,
		ItemID:     itemID,
		ItemName:   itemName,
		ItemEmoji:  itemEmoji,
		Quantity:   qty,
		StartPrice: startPrice,
		BuyItNow:   buyItNow,
		Status:     AuctionOpen,
		CreatedAt:  now,
		EndsAt:     now.Add(AuctionDuration),
	}
	if err := s.auctions.Save(a); err != nil {
		return nil, err
	}
	return a, nil
}

// Bid places a bid on an auction.
// Returns the BidResult (caller must handle OldBidder refund and reserve new bid from buyer).
func (s *MarketService) Bid(auctionID, bidderID int64, bidderName string, amount int) (BidResult, error) {
	a, err := s.auctions.GetByID(auctionID)
	if err != nil {
		return BidResult{}, err
	}
	result, err := PlaceBid(a, bidderID, bidderName, amount)
	if err != nil {
		return BidResult{}, err
	}
	if saveErr := s.auctions.Update(a); saveErr != nil {
		return BidResult{}, saveErr
	}
	return result, nil
}

// SettleExpiredAuctions finds auctions past their end time and settles them.
// Returns a list of settled auctions with winner info.
type SettledAuction struct {
	Auction       *Auction
	WinnerID      int64
	WinnerAmount  int
	SellerProceeds int
	TaxAmount     int
	NoBids        bool
}

func (s *MarketService) SettleExpiredAuctions() ([]SettledAuction, error) {
	open, err := s.auctions.ListOpen(100, 0)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	var settled []SettledAuction
	for _, a := range open {
		if now.Before(a.EndsAt) {
			continue
		}
		winnerID, winnerAmount, noBids := Settle(a)
		if err := s.auctions.Update(a); err != nil {
			continue
		}
		sa := SettledAuction{Auction: a, NoBids: noBids}
		if !noBids {
			cfg := economy.InflationAdjustedConfig(s.taxCfg, s.economy)
			sa.WinnerID = winnerID
			sa.WinnerAmount = winnerAmount
			sa.SellerProceeds, sa.TaxAmount = economy.ApplyAuctionTax(winnerAmount, cfg)
			s.economy.AddGold(int64(sa.SellerProceeds))
			economy.RecordSink(s.economy, sa.TaxAmount, economy.SinkAuctionFee)
		}
		settled = append(settled, sa)
	}
	return settled, nil
}
