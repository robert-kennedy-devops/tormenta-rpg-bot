// Package market implements a global player-driven marketplace with listings,
// direct trades and an auction house.  All gold flows are routed through the
// economy package so the EconomyManager always has accurate circulation data.
package market

import (
	"errors"
	"sync"
	"time"
)

// ─── Listing ──────────────────────────────────────────────────────────────────

// ListingStatus tracks the lifecycle of a market listing.
type ListingStatus string

const (
	ListingActive    ListingStatus = "active"
	ListingCancelled ListingStatus = "cancelled"
	ListingSold      ListingStatus = "sold"
	ListingExpired   ListingStatus = "expired"
)

// Listing represents one item (or stack) posted on the marketplace.
type Listing struct {
	ID          int64
	SellerID    int64  // player_id of seller
	SellerName  string
	ItemID      string
	ItemName    string
	ItemEmoji   string
	Quantity    int
	UnitPrice   int    // gold per unit
	TotalPrice  int    // UnitPrice * Quantity
	Status      ListingStatus
	CreatedAt   time.Time
	ExpiresAt   time.Time
	SoldAt      *time.Time
	BuyerID     *int64
}

var (
	ErrListingNotFound  = errors.New("listagem não encontrada")
	ErrListingExpired   = errors.New("listagem expirada")
	ErrListingNotActive = errors.New("listagem não está ativa")
	ErrInsufficientFunds = errors.New("ouro insuficiente")
	ErrSelfPurchase     = errors.New("você não pode comprar seu próprio item")
)

// ─── In-memory listing store ──────────────────────────────────────────────────
// In production this would be backed by PostgreSQL.  The store interface is
// kept minimal so a DB-backed implementation can be swapped in later.

// ListingStore is the storage interface for market listings.
type ListingStore interface {
	Save(l *Listing) error
	GetByID(id int64) (*Listing, error)
	ListActive(itemID string, limit, offset int) ([]*Listing, error)
	ListBySeller(sellerID int64) ([]*Listing, error)
	Update(l *Listing) error
}

// MemListingStore is a simple in-memory implementation (for testing / MVP).
type MemListingStore struct {
	mu       sync.RWMutex
	listings map[int64]*Listing
	seq      int64
}

// NewMemListingStore creates an empty in-memory store.
func NewMemListingStore() *MemListingStore {
	return &MemListingStore{listings: make(map[int64]*Listing)}
}

func (s *MemListingStore) Save(l *Listing) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.seq++
	l.ID = s.seq
	s.listings[l.ID] = l
	return nil
}

func (s *MemListingStore) GetByID(id int64) (*Listing, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	l, ok := s.listings[id]
	if !ok {
		return nil, ErrListingNotFound
	}
	return l, nil
}

func (s *MemListingStore) ListActive(itemID string, limit, offset int) ([]*Listing, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	now := time.Now()
	var out []*Listing
	for _, l := range s.listings {
		if l.Status != ListingActive {
			continue
		}
		if now.After(l.ExpiresAt) {
			continue
		}
		if itemID != "" && l.ItemID != itemID {
			continue
		}
		out = append(out, l)
	}
	// simple offset/limit
	if offset >= len(out) {
		return nil, nil
	}
	out = out[offset:]
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

func (s *MemListingStore) ListBySeller(sellerID int64) ([]*Listing, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []*Listing
	for _, l := range s.listings {
		if l.SellerID == sellerID {
			out = append(out, l)
		}
	}
	return out, nil
}

func (s *MemListingStore) Update(l *Listing) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.listings[l.ID]; !ok {
		return ErrListingNotFound
	}
	s.listings[l.ID] = l
	return nil
}
