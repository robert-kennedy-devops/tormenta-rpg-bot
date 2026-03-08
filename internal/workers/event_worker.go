package workers

import (
	"log"

	"github.com/tormenta-bot/internal/eventbus"
)

// ─── Event worker ─────────────────────────────────────────────────────────────
//
// Subscribes to the global event bus and performs cross-cutting reactions:
//   - Logs notable events
//   - Updates leaderboards in cache
//   - Sends Telegram notifications for major events (boss spawns, war results)
//
// The notification callback is injected by the caller so this package remains
// independent of the Telegram bot API.

// NotifyFunc is a function that sends a message to a Telegram chat.
// chatID = 0 means broadcast to all online players (implementation-specific).
type NotifyFunc func(chatID int64, message string)

// EventWorker reacts to events from the global bus.
type EventWorker struct {
	notify NotifyFunc
	done   chan struct{}
}

// NewEventWorker creates a worker.  notify may be nil (no Telegram messages).
func NewEventWorker(notify NotifyFunc) *EventWorker {
	return &EventWorker{notify: notify, done: make(chan struct{})}
}

// Start registers all event subscriptions and begins processing.
func (w *EventWorker) Start() {
	bus := eventbus.Global
	bus.Subscribe(eventbus.KindPlayerLevelUp, w.onLevelUp)
	bus.Subscribe(eventbus.KindBossKilled, w.onBossKilled)
	bus.Subscribe(eventbus.KindWorldBossSpawned, w.onBossSpawn)
	bus.Subscribe(eventbus.KindGuildWarEnded, w.onWarEnded)
	bus.Subscribe(eventbus.KindSeasonEnded, w.onSeasonEnded)
	bus.Subscribe(eventbus.KindInflationCritical, w.onInflationCritical)
	bus.Subscribe(eventbus.KindRaidCompleted, w.onRaidCompleted)
}

// Stop deregisters all subscriptions.
func (w *EventWorker) Stop() {
	close(w.done)
}

func (w *EventWorker) onLevelUp(e eventbus.Event) {
	log.Printf("[EventWorker] 🎉 Player %d alcançou nível %d!", e.PlayerID, e.IntValue)
}

func (w *EventWorker) onBossKilled(e eventbus.Event) {
	msg := "🏆 *Boss derrotado!* " + e.EntityID + " foi eliminado pelos aventureiros de Arton!"
	log.Printf("[EventWorker] %s", msg)
	w.broadcast(msg)
}

func (w *EventWorker) onBossSpawn(e eventbus.Event) {
	msg := "⚠️ *ALERTA MUNDIAL!* " + e.StrValue + " surgiu no mundo! Corram para derrotá-lo!"
	log.Printf("[EventWorker] %s", msg)
	w.broadcast(msg)
}

func (w *EventWorker) onWarEnded(e eventbus.Event) {
	msg := "⚔️ *Guerra de Guilda encerrada!* " + e.StrValue
	log.Printf("[EventWorker] %s", msg)
	w.broadcast(msg)
}

func (w *EventWorker) onSeasonEnded(e eventbus.Event) {
	msg := "🏁 *Temporada encerrada!* Uma nova temporada começa agora. Verifique suas recompensas!"
	log.Printf("[EventWorker] %s", msg)
	w.broadcast(msg)
}

func (w *EventWorker) onInflationCritical(e eventbus.Event) {
	log.Printf("[EventWorker] 🚨 Inflação crítica detectada! gold_total=%d", e.IntValue)
	// Trigger automated gold sink: increase NPC repair costs, etc.
	// (Specific actions delegated to economy package at handler level)
}

func (w *EventWorker) onRaidCompleted(e eventbus.Event) {
	msg := "🌟 *Raid concluído!* Um grupo de aventureiros completou o raid " + e.EntityID + "!"
	log.Printf("[EventWorker] %s", msg)
	w.broadcast(msg)
}

// broadcast sends to chat ID 0 which callers interpret as "all players".
func (w *EventWorker) broadcast(msg string) {
	if w.notify != nil {
		w.notify(0, msg)
	}
}
