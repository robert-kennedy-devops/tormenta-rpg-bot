package market

import (
	"errors"
	"sync"
	"time"
)

// ─── Auction ──────────────────────────────────────────────────────────────────

var (
	ErrAuctionNotFound   = errors.New("leilão não encontrado")
	ErrAuctionClosed     = errors.New("leilão encerrado")
	ErrBidTooLow         = errors.New("lance abaixo do mínimo")
	ErrOwnAuction        = errors.New("você não pode dar lance no seu próprio leilão")
)

// AuctionStatus tracks the lifecycle of an auction.
type AuctionStatus string

const (
	AuctionOpen   AuctionStatus = "open"
	AuctionClosed AuctionStatus = "closed"
	AuctionEnded  AuctionStatus = "ended" // time expired, pending settlement
)

// Bid is one bid placed on an auction.
type Bid struct {
	BidderID   int64
	BidderName string
	Amount     int
	PlacedAt   time.Time
}

// Auction is an active rare-item auction.
type Auction struct {
	ID          int64
	SellerID    int64
	SellerName  string
	ItemID      string
	ItemName    string
	ItemEmoji   string
	Quantity    int
	StartPrice  int     // minimum opening bid
	BuyItNow   int     // optional: instantly ends auction
	CurrentBid  int
	CurrentBidder *int64
	CurrentBidderName string
	Bids        []Bid
	Status      AuctionStatus
	CreatedAt   time.Time
	EndsAt      time.Time
}

// ─── In-memory auction store ──────────────────────────────────────────────────

// AuctionStore is the storage interface.
type AuctionStore interface {
	Save(a *Auction) error
	GetByID(id int64) (*Auction, error)
	ListOpen(limit, offset int) ([]*Auction, error)
	Update(a *Auction) error
}

// MemAuctionStore is a simple in-memory store.
type MemAuctionStore struct {
	mu       sync.RWMutex
	auctions map[int64]*Auction
	seq      int64
}

// NewMemAuctionStore creates an empty store.
func NewMemAuctionStore() *MemAuctionStore {
	return &MemAuctionStore{auctions: make(map[int64]*Auction)}
}

func (s *MemAuctionStore) Save(a *Auction) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.seq++
	a.ID = s.seq
	s.auctions[a.ID] = a
	return nil
}

func (s *MemAuctionStore) GetByID(id int64) (*Auction, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	a, ok := s.auctions[id]
	if !ok {
		return nil, ErrAuctionNotFound
	}
	return a, nil
}

func (s *MemAuctionStore) ListOpen(limit, offset int) ([]*Auction, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	now := time.Now()
	var out []*Auction
	for _, a := range s.auctions {
		if a.Status != AuctionOpen {
			continue
		}
		if now.After(a.EndsAt) {
			continue
		}
		out = append(out, a)
	}
	if offset >= len(out) {
		return nil, nil
	}
	out = out[offset:]
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

func (s *MemAuctionStore) Update(a *Auction) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.auctions[a.ID]; !ok {
		return ErrAuctionNotFound
	}
	s.auctions[a.ID] = a
	return nil
}

// ─── PlaceBid validates and records a bid (pure logic, no DB) ─────────────────

// BidResult is what happens when a bid is placed.
type BidResult struct {
	Accepted    bool
	OldBidder   *int64 // previous bidder to refund
	OldAmount   int
	NewBid      int
	IsBuyItNow  bool
}

// PlaceBid validates a bid and updates the auction in memory.
// Callers must persist the updated Auction and handle the OldBidder refund.
func PlaceBid(a *Auction, bidderID int64, bidderName string, amount int) (BidResult, error) {
	if a.Status != AuctionOpen {
		return BidResult{}, ErrAuctionClosed
	}
	if a.SellerID == bidderID {
		return BidResult{}, ErrOwnAuction
	}
	minBid := a.StartPrice
	if a.CurrentBid > 0 {
		minBid = a.CurrentBid + 1
	}
	if amount < minBid {
		return BidResult{}, ErrBidTooLow
	}

	result := BidResult{
		Accepted:  true,
		OldBidder: a.CurrentBidder,
		OldAmount: a.CurrentBid,
		NewBid:    amount,
	}

	// Buy-it-now?
	if a.BuyItNow > 0 && amount >= a.BuyItNow {
		result.IsBuyItNow = true
		a.Status = AuctionClosed
		amount = a.BuyItNow
	}

	bid := Bid{BidderID: bidderID, BidderName: bidderName, Amount: amount, PlacedAt: time.Now()}
	a.Bids = append(a.Bids, bid)
	a.CurrentBid = amount
	a.CurrentBidder = &bidderID
	a.CurrentBidderName = bidderName

	return result, nil
}

// Settle marks an ended auction as closed and returns the winner info.
// Returns (winnerID, winnerAmount, false) if there's a winner, or (0, 0, true) if no bids.
func Settle(a *Auction) (winnerID int64, winnerAmount int, noBids bool) {
	if len(a.Bids) == 0 {
		a.Status = AuctionClosed
		return 0, 0, true
	}
	a.Status = AuctionClosed
	return *a.CurrentBidder, a.CurrentBid, false
}
