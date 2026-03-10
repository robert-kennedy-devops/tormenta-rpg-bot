package security

// logger.go — structured security event logger.
//
// All security-relevant operations emit a SecurityLogEntry with:
//   • timestamp, user ID, event kind, action, detail, source IP (if available)
//
// The logger writes to the standard Go log package by default (compatible with
// any slog/zerolog/zap wrapper that captures stdlib log output).  The output
// is intentionally line-structured JSON so it can be ingested directly by
// log-shipping agents (Loki, Datadog, CloudWatch, etc.).
//
// Usage:
//   security.Log.Info(userID, "shop_buy", "user bought sword")
//   security.Log.SecurityEvent(userID, security.EventExploit, "callback_dup", detail, "")
//   entries := security.Log.Recent(50)

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"
)

// ─── Event kinds ──────────────────────────────────────────────────────────────

// EventKind classifies a security log entry.
type EventKind string

const (
	EventInfo             EventKind = "info"
	EventWarn             EventKind = "warn"
	EventExploit          EventKind = "exploit"
	EventRateLimit        EventKind = "rate_limit"
	EventAnomaly          EventKind = "anomaly"
	EventBlocked          EventKind = "blocked"
	EventPermissionDenied EventKind = "permission_denied"
	EventInvalidInput     EventKind = "invalid_input"
	EventEconomyViolation EventKind = "economy_violation"
	EventPayment          EventKind = "payment"
	EventCombat           EventKind = "combat"
	EventShop             EventKind = "shop"
	EventDrop             EventKind = "drop"
	EventEnergy           EventKind = "energy"
	EventGuild            EventKind = "guild"
	EventGM               EventKind = "gm"
	EventMarket           EventKind = "market"
)

// ─── Log entry ────────────────────────────────────────────────────────────────

// SecurityLogEntry is one structured log record.
type SecurityLogEntry struct {
	Time      time.Time `json:"t"`
	UserID    int64     `json:"uid"`
	Kind      EventKind `json:"kind"`
	Action    string    `json:"action"`
	Detail    string    `json:"detail,omitempty"`
	Extra     string    `json:"extra,omitempty"`
}

// String returns a JSON representation of the entry.
func (e SecurityLogEntry) String() string {
	b, _ := json.Marshal(e)
	return string(b)
}

// ─── SecLogger ────────────────────────────────────────────────────────────────

const maxLogBuffer = 20_000

// SecLogger holds a ring-buffer of recent security events and forwards them to
// the stdlib logger.
type SecLogger struct {
	mu      sync.Mutex
	entries []SecurityLogEntry

	// If set, entries are also forwarded to this channel (for live monitoring).
	feed chan SecurityLogEntry

	// stdLog controls whether entries are forwarded to log.Printf.
	stdLog bool
}

// NewSecLogger creates a logger. If stdLog is true, every entry is also
// forwarded to the stdlib log package.
func NewSecLogger(stdLog bool) *SecLogger {
	return &SecLogger{
		entries: make([]SecurityLogEntry, 0, 1024),
		stdLog:  stdLog,
	}
}

// WithFeed attaches a channel that receives every new entry.
// The channel must be buffered or entries are dropped (never blocks the logger).
func (l *SecLogger) WithFeed(ch chan SecurityLogEntry) *SecLogger {
	l.mu.Lock()
	l.feed = ch
	l.mu.Unlock()
	return l
}

// ─── Logging methods ──────────────────────────────────────────────────────────

// SecurityEvent records a security-relevant event with full context.
func (l *SecLogger) SecurityEvent(userID int64, kind EventKind, action, detail, extra string) {
	e := SecurityLogEntry{
		Time:   time.Now(),
		UserID: userID,
		Kind:   kind,
		Action: action,
		Detail: detail,
		Extra:  extra,
	}
	l.write(e)
}

// Info logs an informational game event.
func (l *SecLogger) Info(userID int64, action, detail string) {
	l.SecurityEvent(userID, EventInfo, action, detail, "")
}

// Warn logs a warning.
func (l *SecLogger) Warn(userID int64, action, detail string) {
	l.SecurityEvent(userID, EventWarn, action, detail, "")
}

// Combat logs a combat action.
func (l *SecLogger) Combat(userID int64, detail string) {
	l.SecurityEvent(userID, EventCombat, "combat_action", detail, "")
}

// Shop logs a shop transaction.
func (l *SecLogger) Shop(userID int64, action, detail string) {
	l.SecurityEvent(userID, EventShop, action, detail, "")
}

// Drop logs a loot drop event.
func (l *SecLogger) Drop(userID int64, detail string) {
	l.SecurityEvent(userID, EventDrop, "loot_drop", detail, "")
}

// Energy logs an energy usage event.
func (l *SecLogger) Energy(userID int64, action string, before, after int) {
	l.SecurityEvent(userID, EventEnergy, action,
		fmt.Sprintf("before=%d after=%d", before, after), "")
}

// Payment logs a payment / diamond-crediting event.
func (l *SecLogger) Payment(userID int64, txID string, diamonds int) {
	l.SecurityEvent(userID, EventPayment, "diamond_credit",
		fmt.Sprintf("txID=%s diamonds=%d", txID, diamonds), "")
}

// EconomyChange logs any change to the in-game economy (gold/diamonds).
func (l *SecLogger) EconomyChange(userID int64, currency string, before, delta, after int) {
	l.SecurityEvent(userID, EventEconomyViolation, "economy_change",
		fmt.Sprintf("currency=%s before=%d delta=%+d after=%d", currency, before, delta, after), "")
}

// GM logs a game-master command execution.
func (l *SecLogger) GM(gmUserID int64, cmd, detail string) {
	l.SecurityEvent(gmUserID, EventGM, cmd, detail, "")
}

// Market logs a market / auction event.
func (l *SecLogger) Market(userID int64, action, detail string) {
	l.SecurityEvent(userID, EventMarket, action, detail, "")
}

// Exploit logs a detected exploit.
func (l *SecLogger) Exploit(userID int64, kind ExploitKind, detail string) {
	l.SecurityEvent(userID, EventExploit, string(kind), detail, "")
}

// Anomaly logs a behaviour anomaly.
func (l *SecLogger) Anomaly(ev AnomalyEvent) {
	l.SecurityEvent(ev.UserID, EventAnomaly, string(ev.Kind),
		fmt.Sprintf("score=%.2f %s", ev.Score, ev.Detail), "")
}

// ─── Retrieval ────────────────────────────────────────────────────────────────

// Recent returns a snapshot of the most recent n entries (newest last).
func (l *SecLogger) Recent(n int) []SecurityLogEntry {
	l.mu.Lock()
	defer l.mu.Unlock()
	if n <= 0 || n > len(l.entries) {
		n = len(l.entries)
	}
	out := make([]SecurityLogEntry, n)
	copy(out, l.entries[len(l.entries)-n:])
	return out
}

// RecentByKind returns the most recent n entries whose Kind matches k.
func (l *SecLogger) RecentByKind(kind EventKind, n int) []SecurityLogEntry {
	l.mu.Lock()
	defer l.mu.Unlock()
	var out []SecurityLogEntry
	for i := len(l.entries) - 1; i >= 0 && len(out) < n; i-- {
		if l.entries[i].Kind == kind {
			out = append(out, l.entries[i])
		}
	}
	// Reverse so newest is last.
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return out
}

// RecentByUser returns the most recent n entries for a specific user.
func (l *SecLogger) RecentByUser(userID int64, n int) []SecurityLogEntry {
	l.mu.Lock()
	defer l.mu.Unlock()
	var out []SecurityLogEntry
	for i := len(l.entries) - 1; i >= 0 && len(out) < n; i-- {
		if l.entries[i].UserID == userID {
			out = append(out, l.entries[i])
		}
	}
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return out
}

// Count returns the total number of entries stored.
func (l *SecLogger) Count() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.entries)
}

// ─── Internal write ───────────────────────────────────────────────────────────

func (l *SecLogger) write(e SecurityLogEntry) {
	l.mu.Lock()
	if len(l.entries) >= maxLogBuffer {
		l.entries = l.entries[maxLogBuffer/2:]
	}
	l.entries = append(l.entries, e)
	feed := l.feed
	l.mu.Unlock()

	if l.stdLog {
		log.Printf("[SEC] %s", e.String())
	}
	if feed != nil {
		select {
		case feed <- e:
		default: // never block
		}
	}
}

// ─── Package-level singleton ──────────────────────────────────────────────────

// Log is the default package-level SecLogger (writes to stdlib log).
var Log = NewSecLogger(true)
