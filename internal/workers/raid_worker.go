package workers

import (
	"log"
	"time"

	"github.com/tormenta-bot/internal/eventbus"
	"github.com/tormenta-bot/internal/world"
)

// ─── Raid worker ──────────────────────────────────────────────────────────────
//
// Responsibilities:
//   - Check if the world boss window has expired and mark as escaped
//   - Attempt to spawn a new world boss every BossSpawnInterval
//   - Settle expired auctions, wars and seasons (delegated to respective services)

// RaidWorker monitors world boss spawns and expiry.
type RaidWorker struct {
	bossManager *world.BossManager
	interval    time.Duration
	done        chan struct{}
}

// NewRaidWorker creates the raid worker.
func NewRaidWorker(bm *world.BossManager, interval time.Duration) *RaidWorker {
	if interval <= 0 {
		interval = 5 * time.Minute
	}
	return &RaidWorker{bossManager: bm, interval: interval, done: make(chan struct{})}
}

// Start runs the worker.
func (w *RaidWorker) Start() {
	go w.loop()
}

// Stop signals shutdown.
func (w *RaidWorker) Stop() {
	close(w.done)
}

func (w *RaidWorker) loop() {
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

func (w *RaidWorker) tick() {
	// Try to spawn boss
	spawned := w.bossManager.SpawnIfReady()
	if spawned != nil {
		log.Printf("[RaidWorker] 🌍 World boss spawned: %s (%s)", spawned.Def.Name, spawned.BossID)
		eventbus.Pub(
			eventbus.NewEvent(eventbus.KindWorldBossSpawned).
				WithEntity(spawned.BossID).
				WithStr(spawned.Def.Name),
		)
	}

	// Check if active boss expired
	active := w.bossManager.Active()
	if active != nil && time.Now().After(active.ExpiresAt) {
		log.Printf("[RaidWorker] 💨 World boss escaped: %s", active.BossID)
	}
}
