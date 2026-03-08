// Package workers provides scalable background processing for the game.
// Each worker runs in its own goroutine pool and communicates via the event bus
// and shared state.  Workers are designed to scale horizontally — multiple
// instances can run concurrently without data races.
package workers

import (
	"log"
	"time"

	"github.com/tormenta-bot/internal/cache"
	"github.com/tormenta-bot/internal/economy"
	"github.com/tormenta-bot/internal/eventbus"
)

// ─── Economy worker ───────────────────────────────────────────────────────────
//
// Runs every 10 minutes.  Responsibilities:
//   - Take an economy snapshot
//   - Update the cache with the latest snapshot
//   - Publish inflation events if thresholds are exceeded
//   - Trigger passive gold sink (territory income, repair fees, etc.)

// EconomyWorker periodically monitors and adjusts the game economy.
type EconomyWorker struct {
	em       *economy.EconomyManager
	interval time.Duration
	done     chan struct{}
}

// NewEconomyWorker creates a new worker.
func NewEconomyWorker(em *economy.EconomyManager, interval time.Duration) *EconomyWorker {
	if interval <= 0 {
		interval = 10 * time.Minute
	}
	return &EconomyWorker{em: em, interval: interval, done: make(chan struct{})}
}

// Start runs the worker in a background goroutine.
func (w *EconomyWorker) Start() {
	go w.loop()
}

// Stop signals the worker to shut down gracefully.
func (w *EconomyWorker) Stop() {
	close(w.done)
}

func (w *EconomyWorker) loop() {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			w.tick()
		case <-w.done:
			return
		}
	}
}

func (w *EconomyWorker) tick() {
	snap := w.em.Snapshot()

	// Cache the snapshot for handlers/UI
	_ = cache.Global.Set(cache.EconomySnapshotKey, snap, 15*time.Minute)

	// Publish inflation events
	switch snap.InflationLevel {
	case economy.InflationCritical:
		eventbus.Pub(eventbus.NewEvent(eventbus.KindInflationCritical).
			WithInt(int(snap.TotalGold)).
			WithStr("gold_per_player"))
		log.Printf("[EconomyWorker] ⚠️ CRITICAL INFLATION: %d gold/player", int(snap.GoldPerPlayer))
	case economy.InflationWarning:
		eventbus.Pub(eventbus.NewEvent(eventbus.KindInflationWarning).
			WithInt(int(snap.GoldPerPlayer)))
	}
}
