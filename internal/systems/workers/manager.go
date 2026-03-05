package workers

import (
	"log"
	"time"

	"github.com/tormenta-bot/internal/database"
	"github.com/tormenta-bot/internal/handlers"
	"github.com/tormenta-bot/internal/services/payment"
	"github.com/tormenta-bot/internal/systems/events"
)

type Manager struct {
	payments *payment.Service
	events   *events.Manager
}

func NewManager() *Manager {
	return &Manager{
		payments: payment.NewService(),
		events:   events.Global,
	}
}

func (m *Manager) Start(enablePixPolling bool) {
	go m.energyWorker(2 * time.Minute)
	go m.dungeonCleanupWorker(15 * time.Minute)
	go m.auctionCleanupWorker(20 * time.Minute)
	go m.eventWorker(1 * time.Minute)
	if enablePixPolling {
		go m.pixWorker(15 * time.Second)
	}
}

func (m *Manager) energyWorker(interval time.Duration) {
	t := time.NewTicker(interval)
	defer t.Stop()
	for range t.C {
		if err := database.ClampCharacterEnergy(); err != nil {
			log.Printf("[workers.energy] clamp error: %v", err)
		}
	}
}

func (m *Manager) dungeonCleanupWorker(interval time.Duration) {
	t := time.NewTicker(interval)
	defer t.Stop()
	for range t.C {
		n, err := database.CleanupExpiredDungeonRuns(1)
		if err != nil {
			log.Printf("[workers.dungeon_cleanup] error: %v", err)
			continue
		}
		if n > 0 {
			log.Printf("[workers.dungeon_cleanup] expired runs=%d", n)
		}
	}
}

// Placeholder for future auction system without breaking architecture.
func (m *Manager) auctionCleanupWorker(interval time.Duration) {
	t := time.NewTicker(interval)
	defer t.Stop()
	for range t.C {
		_, _ = database.DB.Exec(`SELECT 1`)
	}
}

func (m *Manager) eventWorker(interval time.Duration) {
	t := time.NewTicker(interval)
	defer t.Stop()
	for range t.C {
		ev := m.events.MaybeRoll(time.Now())
		if ev.Kind != events.EventNone {
			log.Printf("[workers.events] active=%s ends=%s", ev.Kind, ev.EndsAt.Format(time.RFC3339))
		}
	}
}

func (m *Manager) pixWorker(interval time.Duration) {
	t := time.NewTicker(interval)
	defer t.Stop()
	for range t.C {
		pending, err := database.GetAllPendingPixPayments()
		if err != nil || len(pending) == 0 {
			continue
		}
		for _, p := range pending {
			if p.TxID == "" {
				continue
			}
			status, err := handlers.CheckAbacateStatusForWorker(p.TxID)
			if err != nil || status != "approved" {
				continue
			}
			res, err := m.payments.ConfirmByTxID(p.TxID, "worker")
			if err != nil || res.Diamonds <= 0 {
				continue
			}
			handlers.NotifyPixConfirmedWorker(res.CharacterID, res.Diamonds, p.PackageID)
			handlers.NotifyGMPixPaid(res.CharacterID, res.Diamonds, p.PackageID, p.AmountBRL)
		}
	}
}
