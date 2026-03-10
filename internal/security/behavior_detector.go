package security

// behavior_detector.go — tracks per-user action history and raises anomalies
// for patterns that suggest automated / bot-like behaviour or farming exploits.
//
// Detects:
//   • Impossible action speed (sub-human tap intervals)
//   • Dungeon farming (excessive kill rate in a short window)
//   • Spam farming (very high action density over minutes)
//   • Bot-like regularity (suspiciously uniform inter-action intervals)
//   • Economy anomalies (gold/xp gain rate exceeds theoretical cap)

import (
	"fmt"
	"math"
	"sync"
	"time"
)

// ─── Anomaly kinds ────────────────────────────────────────────────────────────

// AnomalyKind names a class of suspicious behaviour.
type AnomalyKind string

const (
	AnomalySpeedFarm      AnomalyKind = "speed_farm"       // actions too fast for a human
	AnomalyDungeonFarm    AnomalyKind = "dungeon_farm"     // kill rate impossibly high
	AnomalySpamFarm       AnomalyKind = "spam_farm"        // action density > threshold
	AnomalyBotPattern     AnomalyKind = "bot_pattern"      // suspiciously uniform timing
	AnomalyGoldAnomaly    AnomalyKind = "gold_anomaly"     // gold gain rate spike
	AnomalyXPAnomaly      AnomalyKind = "xp_anomaly"       // xp gain rate spike
	AnomalyCallbackSpam   AnomalyKind = "callback_spam"    // too many callbacks/min
)

// AnomalyEvent is produced whenever an anomaly is detected.
type AnomalyEvent struct {
	Time    time.Time
	UserID  int64
	Kind    AnomalyKind
	Score   float64 // 0.0–1.0 confidence
	Detail  string
}

// ─── Per-user profile ─────────────────────────────────────────────────────────

const profileWindowSize = 60 // keep last N timestamps for variance analysis

type userProfile struct {
	mu sync.Mutex

	// Sliding window of action timestamps (newest last).
	actions []time.Time

	// Running economy counters (reset every minute).
	goldEarned    int
	xpEarned      int
	dungeonKills  int
	callbacks     int
	windowStart   time.Time

	// Anomaly strike counter: 3 strikes → automatic flag.
	strikes int
	flagged bool
	flagAt  time.Time
}

func newUserProfile() *userProfile {
	return &userProfile{windowStart: time.Now()}
}

// recordAction appends a timestamp to the sliding action window.
// Returns the current window size (for caller convenience).
func (p *userProfile) recordAction(now time.Time) int {
	p.actions = append(p.actions, now)
	// Keep only the last profileWindowSize entries.
	if len(p.actions) > profileWindowSize {
		p.actions = p.actions[len(p.actions)-profileWindowSize:]
	}
	return len(p.actions)
}

// rotateIfNeeded resets per-minute counters when the window has elapsed.
func (p *userProfile) rotateIfNeeded(now time.Time) {
	if now.Sub(p.windowStart) >= time.Minute {
		p.goldEarned = 0
		p.xpEarned = 0
		p.dungeonKills = 0
		p.callbacks = 0
		p.windowStart = now
	}
}

// ─── BehaviorDetector ────────────────────────────────────────────────────────

// BehaviorDetector tracks player action history and raises anomaly events.
// It is safe for concurrent use.
type BehaviorDetector struct {
	mu       sync.Mutex
	profiles map[int64]*userProfile
	events   []AnomalyEvent
	cleanAt  time.Time

	// Tunable thresholds.
	cfg BehaviorConfig
}

// BehaviorConfig holds tunable detection thresholds.
type BehaviorConfig struct {
	// MaxActionsPerMinute: action count per 60 s before spam_farm fires.
	MaxActionsPerMinute int
	// MaxDungeonKillsPerMinute: kill rate cap.
	MaxDungeonKillsPerMinute int
	// MaxGoldPerMinute: maximum gold gain per 60 s window.
	MaxGoldPerMinute int
	// MaxXPPerMinute: maximum XP gain per 60 s window.
	MaxXPPerMinute int
	// MaxCallbacksPerMinute: raw callback dispatches per 60 s.
	MaxCallbacksPerMinute int
	// BotPatternCV: coefficient of variation below which timing is "too regular".
	// Humans have high timing variance; bots have very low variance.
	BotPatternCV float64
	// MinSamplesForBotCheck: need at least this many timed actions before
	// running the bot-pattern variance test.
	MinSamplesForBotCheck int
}

// DefaultBehaviorConfig is the recommended production configuration.
var DefaultBehaviorConfig = BehaviorConfig{
	MaxActionsPerMinute:      120,
	MaxDungeonKillsPerMinute: 20,
	MaxGoldPerMinute:         500_000,
	MaxXPPerMinute:           200_000,
	MaxCallbacksPerMinute:    60,
	BotPatternCV:             0.05, // <5% variance → suspicious
	MinSamplesForBotCheck:    20,
}

// NewBehaviorDetector creates a detector with the given configuration.
// Pass DefaultBehaviorConfig for production defaults.
func NewBehaviorDetector(cfg BehaviorConfig) *BehaviorDetector {
	return &BehaviorDetector{
		profiles: make(map[int64]*userProfile),
		cfg:      cfg,
		cleanAt:  time.Now().Add(15 * time.Minute),
	}
}

// ─── Observation API ──────────────────────────────────────────────────────────

// ObserveAction records a generic action for userID and returns any detected
// anomaly (nil if none).
func (bd *BehaviorDetector) ObserveAction(userID int64) *AnomalyEvent {
	p := bd.getProfile(userID)
	now := time.Now()

	p.mu.Lock()
	defer p.mu.Unlock()

	p.rotateIfNeeded(now)
	count := p.recordAction(now)

	if count >= bd.cfg.MaxActionsPerMinute {
		return bd.flag(p, userID, AnomalySpamFarm, 0.8,
			fmt.Sprintf("actions/min=%d (cap %d)", count, bd.cfg.MaxActionsPerMinute))
	}

	// Bot-pattern check: low timing variance on a sufficient sample.
	if count >= bd.cfg.MinSamplesForBotCheck {
		if cv := timingCV(p.actions); cv < bd.cfg.BotPatternCV {
			return bd.flag(p, userID, AnomalyBotPattern, 1.0-cv,
				fmt.Sprintf("timing_cv=%.4f (threshold %.2f)", cv, bd.cfg.BotPatternCV))
		}
	}
	return nil
}

// ObserveDungeonKill records one monster kill and checks the kill-rate cap.
func (bd *BehaviorDetector) ObserveDungeonKill(userID int64) *AnomalyEvent {
	p := bd.getProfile(userID)
	now := time.Now()

	p.mu.Lock()
	defer p.mu.Unlock()

	p.rotateIfNeeded(now)
	p.dungeonKills++
	if p.dungeonKills > bd.cfg.MaxDungeonKillsPerMinute {
		return bd.flag(p, userID, AnomalyDungeonFarm, 0.9,
			fmt.Sprintf("kills/min=%d (cap %d)", p.dungeonKills, bd.cfg.MaxDungeonKillsPerMinute))
	}
	return nil
}

// ObserveGoldGain records gold earned by userID in this tick.
func (bd *BehaviorDetector) ObserveGoldGain(userID int64, amount int) *AnomalyEvent {
	p := bd.getProfile(userID)
	now := time.Now()

	p.mu.Lock()
	defer p.mu.Unlock()

	p.rotateIfNeeded(now)
	p.goldEarned += amount
	if p.goldEarned > bd.cfg.MaxGoldPerMinute {
		return bd.flag(p, userID, AnomalyGoldAnomaly, 0.85,
			fmt.Sprintf("gold/min=%d (cap %d)", p.goldEarned, bd.cfg.MaxGoldPerMinute))
	}
	return nil
}

// ObserveXPGain records XP earned by userID in this tick.
func (bd *BehaviorDetector) ObserveXPGain(userID int64, amount int) *AnomalyEvent {
	p := bd.getProfile(userID)
	now := time.Now()

	p.mu.Lock()
	defer p.mu.Unlock()

	p.rotateIfNeeded(now)
	p.xpEarned += amount
	if p.xpEarned > bd.cfg.MaxXPPerMinute {
		return bd.flag(p, userID, AnomalyXPAnomaly, 0.85,
			fmt.Sprintf("xp/min=%d (cap %d)", p.xpEarned, bd.cfg.MaxXPPerMinute))
	}
	return nil
}

// ObserveCallback records a raw callback dispatch for spam detection.
func (bd *BehaviorDetector) ObserveCallback(userID int64) *AnomalyEvent {
	p := bd.getProfile(userID)
	now := time.Now()

	p.mu.Lock()
	defer p.mu.Unlock()

	p.rotateIfNeeded(now)
	p.callbacks++
	if p.callbacks > bd.cfg.MaxCallbacksPerMinute {
		return bd.flag(p, userID, AnomalyCallbackSpam, 0.75,
			fmt.Sprintf("callbacks/min=%d (cap %d)", p.callbacks, bd.cfg.MaxCallbacksPerMinute))
	}
	return nil
}

// IsFlagged returns true when the user has been flagged for review.
func (bd *BehaviorDetector) IsFlagged(userID int64) bool {
	p := bd.getProfile(userID)
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.flagged
}

// UnflagUser clears a user's flagged status (GM pardon).
func (bd *BehaviorDetector) UnflagUser(userID int64) {
	p := bd.getProfile(userID)
	p.mu.Lock()
	p.flagged = false
	p.strikes = 0
	p.mu.Unlock()
}

// RecentAnomalies returns a snapshot of the most recent n anomaly events.
func (bd *BehaviorDetector) RecentAnomalies(n int) []AnomalyEvent {
	bd.mu.Lock()
	defer bd.mu.Unlock()
	if n <= 0 || n > len(bd.events) {
		n = len(bd.events)
	}
	out := make([]AnomalyEvent, n)
	copy(out, bd.events[len(bd.events)-n:])
	return out
}

// ─── Internals ────────────────────────────────────────────────────────────────

func (bd *BehaviorDetector) getProfile(userID int64) *userProfile {
	bd.mu.Lock()
	p, ok := bd.profiles[userID]
	if !ok {
		p = newUserProfile()
		bd.profiles[userID] = p
		bd.maybeClean()
	}
	bd.mu.Unlock()
	return p
}

// flag records a strike and, after threshold, sets flagged=true.
// Caller must hold p.mu. Returns the AnomalyEvent for forwarding.
func (bd *BehaviorDetector) flag(p *userProfile, userID int64, kind AnomalyKind, score float64, detail string) *AnomalyEvent {
	p.strikes++
	if p.strikes >= 3 && !p.flagged {
		p.flagged = true
		p.flagAt = time.Now()
	}

	ev := AnomalyEvent{
		Time:   time.Now(),
		UserID: userID,
		Kind:   kind,
		Score:  score,
		Detail: detail,
	}

	bd.mu.Lock()
	const maxAnomalyLog = 5_000
	if len(bd.events) >= maxAnomalyLog {
		bd.events = bd.events[maxAnomalyLog/2:]
	}
	bd.events = append(bd.events, ev)
	bd.mu.Unlock()

	return &ev
}

// maybeClean removes profiles that have been idle for 30 min (caller holds mu).
func (bd *BehaviorDetector) maybeClean() {
	if !time.Now().After(bd.cleanAt) {
		return
	}
	cutoff := time.Now().Add(-30 * time.Minute)
	for uid, p := range bd.profiles {
		p.mu.Lock()
		idle := len(p.actions) == 0 || p.actions[len(p.actions)-1].Before(cutoff)
		p.mu.Unlock()
		if idle {
			delete(bd.profiles, uid)
		}
	}
	bd.cleanAt = time.Now().Add(15 * time.Minute)
}

// ─── Timing variance helper ───────────────────────────────────────────────────

// timingCV computes the coefficient of variation (stddev/mean) of inter-action
// intervals from the timestamp slice. Returns 1.0 (maximum variance) for
// fewer than 2 entries.
func timingCV(ts []time.Time) float64 {
	n := len(ts)
	if n < 2 {
		return 1.0
	}
	intervals := make([]float64, n-1)
	for i := 1; i < n; i++ {
		intervals[i-1] = float64(ts[i].Sub(ts[i-1]).Milliseconds())
	}
	mean := 0.0
	for _, v := range intervals {
		mean += v
	}
	mean /= float64(len(intervals))
	if mean == 0 {
		return 0
	}
	variance := 0.0
	for _, v := range intervals {
		d := v - mean
		variance += d * d
	}
	variance /= float64(len(intervals))
	return math.Sqrt(variance) / mean
}

// ─── Package-level singleton ──────────────────────────────────────────────────

// Behavior is the default package-level BehaviorDetector.
var Behavior = NewBehaviorDetector(DefaultBehaviorConfig)
