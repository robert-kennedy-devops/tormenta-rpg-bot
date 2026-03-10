// db_store.go — PostgreSQL-backed implementations of ListingStore and AuctionStore.
package market

import (
	"database/sql"
	"time"
)

// ── DB-backed ListingStore ────────────────────────────────────────────────────

type DBListingStore struct{ db *sql.DB }

func NewDBListingStore(db *sql.DB) *DBListingStore { return &DBListingStore{db: db} }

func (s *DBListingStore) Save(l *Listing) error {
	return s.db.QueryRow(`
		INSERT INTO market_listings
			(seller_id,seller_name,item_id,item_type,quantity,unit_price,expires_at,sold)
		VALUES ($1,$2,$3,'',$4,$5,$6,false) RETURNING id`,
		l.SellerID, l.SellerName, l.ItemID, l.Quantity, l.UnitPrice, l.ExpiresAt,
	).Scan(&l.ID)
}

func (s *DBListingStore) GetByID(id int64) (*Listing, error) {
	l := &Listing{}
	var sold bool
	err := s.db.QueryRow(`
		SELECT id,seller_id,seller_name,item_id,quantity,unit_price,created_at,expires_at,sold
		FROM market_listings WHERE id=$1`, id).
		Scan(&l.ID, &l.SellerID, &l.SellerName, &l.ItemID, &l.Quantity, &l.UnitPrice,
			&l.CreatedAt, &l.ExpiresAt, &sold)
	if err == sql.ErrNoRows {
		return nil, ErrListingNotFound
	}
	if err != nil {
		return nil, err
	}
	l.TotalPrice = l.UnitPrice * l.Quantity
	l.Status = listingStatus(sold, l.ExpiresAt)
	return l, nil
}

func (s *DBListingStore) ListActive(itemID string, limit, offset int) ([]*Listing, error) {
	if limit <= 0 {
		limit = 20
	}
	var rows *sql.Rows
	var err error
	if itemID == "" {
		rows, err = s.db.Query(`
			SELECT id,seller_id,seller_name,item_id,quantity,unit_price,created_at,expires_at
			FROM market_listings WHERE NOT sold AND expires_at>NOW()
			ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	} else {
		rows, err = s.db.Query(`
			SELECT id,seller_id,seller_name,item_id,quantity,unit_price,created_at,expires_at
			FROM market_listings WHERE NOT sold AND expires_at>NOW() AND item_id=$1
			ORDER BY unit_price ASC LIMIT $2 OFFSET $3`, itemID, limit, offset)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanListings(rows)
}

func (s *DBListingStore) ListBySeller(sellerID int64) ([]*Listing, error) {
	rows, err := s.db.Query(`
		SELECT id,seller_id,seller_name,item_id,quantity,unit_price,created_at,expires_at
		FROM market_listings WHERE seller_id=$1 AND NOT sold ORDER BY created_at DESC`, sellerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanListings(rows)
}

func (s *DBListingStore) Update(l *Listing) error {
	_, err := s.db.Exec(`UPDATE market_listings SET sold=$1 WHERE id=$2`, l.Status == ListingSold, l.ID)
	return err
}

func listingStatus(sold bool, expires time.Time) ListingStatus {
	if sold {
		return ListingSold
	}
	if time.Now().After(expires) {
		return ListingExpired
	}
	return ListingActive
}

func scanListings(rows *sql.Rows) ([]*Listing, error) {
	var out []*Listing
	for rows.Next() {
		l := &Listing{}
		if err := rows.Scan(&l.ID, &l.SellerID, &l.SellerName, &l.ItemID,
			&l.Quantity, &l.UnitPrice, &l.CreatedAt, &l.ExpiresAt); err != nil {
			return nil, err
		}
		l.TotalPrice = l.UnitPrice * l.Quantity
		l.Status = listingStatus(false, l.ExpiresAt)
		out = append(out, l)
	}
	return out, rows.Err()
}

// ── DB-backed AuctionStore ────────────────────────────────────────────────────

type DBAuctionStore struct{ db *sql.DB }

func NewDBAuctionStore(db *sql.DB) *DBAuctionStore { return &DBAuctionStore{db: db} }

func (s *DBAuctionStore) Save(a *Auction) error {
	bidderID := int64(0)
	if a.CurrentBidder != nil {
		bidderID = *a.CurrentBidder
	}
	return s.db.QueryRow(`
		INSERT INTO market_auctions
			(seller_id,seller_name,item_id,item_type,quantity,
			 start_price,current_bid,bidder_id,bidder_name,expires_at,settled)
		VALUES ($1,$2,$3,'',$4,$5,$6,$7,$8,$9,false) RETURNING id`,
		a.SellerID, a.SellerName, a.ItemID, a.Quantity,
		a.StartPrice, a.CurrentBid, bidderID, a.CurrentBidderName, a.EndsAt,
	).Scan(&a.ID)
}

func (s *DBAuctionStore) GetByID(id int64) (*Auction, error) {
	a := &Auction{}
	var settled bool
	var bidderID int64
	err := s.db.QueryRow(`
		SELECT id,seller_id,seller_name,item_id,quantity,start_price,
		       current_bid,bidder_id,bidder_name,expires_at,settled,created_at
		FROM market_auctions WHERE id=$1`, id).
		Scan(&a.ID, &a.SellerID, &a.SellerName, &a.ItemID, &a.Quantity, &a.StartPrice,
			&a.CurrentBid, &bidderID, &a.CurrentBidderName, &a.EndsAt, &settled, &a.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrAuctionNotFound
	}
	if err != nil {
		return nil, err
	}
	if bidderID != 0 {
		v := bidderID; a.CurrentBidder = &v
	}
	a.Status = auctionStatus(settled, a.EndsAt)
	return a, nil
}

func (s *DBAuctionStore) ListOpen(limit, offset int) ([]*Auction, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := s.db.Query(`
		SELECT id,seller_id,seller_name,item_id,quantity,start_price,
		       current_bid,bidder_id,bidder_name,expires_at,settled,created_at
		FROM market_auctions WHERE NOT settled AND expires_at>NOW()
		ORDER BY expires_at ASC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanAuctions(rows)
}

func (s *DBAuctionStore) Update(a *Auction) error {
	bidderID := int64(0)
	if a.CurrentBidder != nil {
		bidderID = *a.CurrentBidder
	}
	_, err := s.db.Exec(`
		UPDATE market_auctions SET current_bid=$1,bidder_id=$2,bidder_name=$3,settled=$4 WHERE id=$5`,
		a.CurrentBid, bidderID, a.CurrentBidderName, a.Status == AuctionClosed, a.ID)
	return err
}

func auctionStatus(settled bool, ends time.Time) AuctionStatus {
	if settled {
		return AuctionClosed
	}
	if time.Now().After(ends) {
		return AuctionEnded
	}
	return AuctionOpen
}

func scanAuctions(rows *sql.Rows) ([]*Auction, error) {
	var out []*Auction
	for rows.Next() {
		a := &Auction{}
		var settled bool
		var bidderID int64
		if err := rows.Scan(&a.ID, &a.SellerID, &a.SellerName, &a.ItemID, &a.Quantity, &a.StartPrice,
			&a.CurrentBid, &bidderID, &a.CurrentBidderName, &a.EndsAt, &settled, &a.CreatedAt); err != nil {
			return nil, err
		}
		if bidderID != 0 {
			v := bidderID; a.CurrentBidder = &v
		}
		a.Status = auctionStatus(settled, a.EndsAt)
		out = append(out, a)
	}
	return out, rows.Err()
}
