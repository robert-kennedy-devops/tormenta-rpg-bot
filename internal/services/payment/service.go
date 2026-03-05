package payment

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/tormenta-bot/internal/database"
)

type Service struct {
	mu       sync.Mutex
	inflight map[string]time.Time
}

func NewService() *Service {
	return &Service{
		inflight: make(map[string]time.Time),
	}
}

type ConfirmResult struct {
	CharacterID      int
	PlayerID         int64
	Diamonds         int
	AlreadyProcessed bool
}

func (s *Service) lockTx(txid string) bool {
	now := time.Now()
	cutoff := now.Add(-2 * time.Minute)

	s.mu.Lock()
	defer s.mu.Unlock()

	for k, t := range s.inflight {
		if t.Before(cutoff) {
			delete(s.inflight, k)
		}
	}
	if _, ok := s.inflight[txid]; ok {
		return false
	}
	s.inflight[txid] = now
	return true
}

func (s *Service) unlockTx(txid string) {
	s.mu.Lock()
	delete(s.inflight, txid)
	s.mu.Unlock()
}

// ConfirmByTxID provides idempotent confirmation with basic fraud guards.
func (s *Service) ConfirmByTxID(txid, source string) (ConfirmResult, error) {
	if txid == "" {
		return ConfirmResult{}, fmt.Errorf("empty txid")
	}
	if !s.lockTx(txid) {
		return ConfirmResult{AlreadyProcessed: true}, nil
	}
	defer s.unlockTx(txid)

	rec, err := database.GetPixPayment(txid)
	if err != nil {
		return ConfirmResult{}, err
	}
	if rec == nil {
		return ConfirmResult{}, fmt.Errorf("payment not found")
	}
	if rec.Diamonds <= 0 || rec.AmountBRL <= 0 {
		return ConfirmResult{}, fmt.Errorf("fraud check failed: invalid amount/diamonds")
	}
	if rec.Status != "pending" {
		return s.resultFromChar(rec.CharacterID, 0, true)
	}

	charID, diamonds, err := database.ConfirmPixPaymentByTxID(txid)
	if err != nil {
		return ConfirmResult{}, err
	}
	if diamonds > 0 {
		reason := "pix_confirmed"
		if source != "" {
			reason = "pix_" + source
		}
		database.LogDiamond(charID, diamonds, reason)
	}
	return s.resultFromChar(charID, diamonds, false)
}

func (s *Service) resultFromChar(charID, diamonds int, already bool) (ConfirmResult, error) {
	var playerID int64
	if charID > 0 {
		if err := database.DB.QueryRow(`
			SELECT p.id FROM players p
			JOIN characters c ON c.player_id = p.id
			WHERE c.id=$1
		`, charID).Scan(&playerID); err != nil {
			log.Printf("[payment] could not resolve player for char=%d: %v", charID, err)
		}
	}
	return ConfirmResult{
		CharacterID:      charID,
		PlayerID:         playerID,
		Diamonds:         diamonds,
		AlreadyProcessed: already,
	}, nil
}
