package security

// rate_limiter.go — per-user token-bucket rate limiter.
//
// Architecture note:
//   This implementation is intentionally dependency-free (stdlib only) so that
//   the security package never forces new entries in go.mod.  The algorithm is
//   a standard token-bucket identical in semantics to golang.org/x/time/rate.
//
// Usage:
//   rl := security.NewUserRateLimiter(security.DefaultLimits)
//   if !rl.Allow(userID, security.ActionShopBuy) {
//       // reject request
//   }

import (
	"sync"
	"time"
)

// ─── Action keys ─────────────────────────────────────────────────────────────

// ActionKey identifies a category of user action for rate-limiting purposes.
type ActionKey string

const (
	ActionGeneral    ActionKey = "general"    // default bucket
	ActionShopBuy    ActionKey = "shop_buy"   // commerce operations
	ActionShopSell   ActionKey = "shop_sell"
	ActionDungeon    ActionKey = "dungeon"    // dungeon entry / combat turn
	ActionCombat     ActionKey = "combat"
	ActionPVP        ActionKey = "pvp"
	ActionMarket     ActionKey = "market"
	ActionAuction    ActionKey = "auction"
	ActionGuildBank  ActionKey = "guild_bank"
	ActionEnergy     ActionKey = "energy"
	ActionForge      ActionKey = "forge"
	ActionCallback   ActionKey = "callback"   // raw callback dispatch
)

// ─── Limit definition ─────────────────────────────────────────────────────────

// Limit defines how many tokens a bucket holds and how fast they refill.
type Limit struct {
	// Burst is the maximum number of actions allowed in a single instant.
	Burst int
	// Rate is the number of tokens replenished per second (can be fractional
	// via Window: Rate tokens are added every Window duration).
	Rate   int
	Window time.Duration
}

// DefaultLimits is the recommended production configuration.
var DefaultLimits = map[ActionKey]Limit{
	ActionGeneral:   {Burst: 30, Rate: 10, Window: time.Second},
	ActionShopBuy:   {Burst: 5, Rate: 2, Window: time.Second},
	ActionShopSell:  {Burst: 5, Rate: 2, Window: time.Second},
	ActionDungeon:   {Burst: 3, Rate: 1, Window: 2 * time.Second},
	ActionCombat:    {Burst: 10, Rate: 3, Window: time.Second},
	ActionPVP:       {Burst: 3, Rate: 1, Window: 3 * time.Second},
	ActionMarket:    {Burst: 6, Rate: 2, Window: time.Second},
	ActionAuction:   {Burst: 4, Rate: 1, Window: time.Second},
	ActionGuildBank: {Burst: 3, Rate: 1, Window: 2 * time.Second},
	ActionEnergy:    {Burst: 5, Rate: 2, Window: time.Second},
	ActionForge:     {Burst: 4, Rate: 1, Window: time.Second},
	ActionCallback:  {Burst: 20, Rate: 8, Window: time.Second},
}

// ─── Token bucket ─────────────────────────────────────────────────────────────

type bucket struct {
	tokens   float64
	lastFill time.Time
	mu       sync.Mutex
}

func newBucket(burst int) *bucket {
	return &bucket{
		tokens:   float64(burst),
		lastFill: time.Now(),
	}
}

// allow consumes one token. Returns true if the action is permitted.
func (b *bucket) allow(lim Limit) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	// Refill: add (elapsed / Window) * Rate tokens, capped at Burst.
	elapsed := now.Sub(b.lastFill)
	if elapsed > 0 && lim.Window > 0 {
		refill := elapsed.Seconds() / lim.Window.Seconds() * float64(lim.Rate)
		b.tokens += refill
		if b.tokens > float64(lim.Burst) {
			b.tokens = float64(lim.Burst)
		}
		b.lastFill = now
	}

	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

// ─── Per-user limiter ─────────────────────────────────────────────────────────

type userBuckets map[ActionKey]*bucket

// UserRateLimiter manages independent token buckets per (user, action).
type UserRateLimiter struct {
	mu      sync.RWMutex
	users   map[int64]userBuckets
	limits  map[ActionKey]Limit
	cleanAt time.Time
}

// NewUserRateLimiter creates a limiter using the provided limit map.
// Pass DefaultLimits for the recommended production configuration.
func NewUserRateLimiter(limits map[ActionKey]Limit) *UserRateLimiter {
	if limits == nil {
		limits = DefaultLimits
	}
	return &UserRateLimiter{
		users:   make(map[int64]userBuckets),
		limits:  limits,
		cleanAt: time.Now().Add(10 * time.Minute),
	}
}

// Allow returns true when the action is within the rate limit for the given user.
// It is safe to call from multiple goroutines.
func (rl *UserRateLimiter) Allow(userID int64, action ActionKey) bool {
	lim, ok := rl.limits[action]
	if !ok {
		lim = rl.limits[ActionGeneral]
	}

	b := rl.getBucket(userID, action, lim.Burst)
	allowed := b.allow(lim)

	// Periodic cleanup of idle user maps (runs at most every 10 min).
	if time.Now().After(rl.cleanAt) {
		go rl.prune()
	}

	return allowed
}

// getBucket returns (or lazily creates) the bucket for (user, action).
func (rl *UserRateLimiter) getBucket(userID int64, action ActionKey, burst int) *bucket {
	rl.mu.RLock()
	if ub, ok := rl.users[userID]; ok {
		if b, ok := ub[action]; ok {
			rl.mu.RUnlock()
			return b
		}
	}
	rl.mu.RUnlock()

	rl.mu.Lock()
	defer rl.mu.Unlock()

	ub, ok := rl.users[userID]
	if !ok {
		ub = make(userBuckets)
		rl.users[userID] = ub
	}
	if b, ok := ub[action]; ok {
		return b
	}
	b := newBucket(burst)
	ub[action] = b
	return b
}

// prune removes users whose buckets have not been touched for > 30 minutes,
// preventing unbounded memory growth in long-running processes.
func (rl *UserRateLimiter) prune() {
	cutoff := time.Now().Add(-30 * time.Minute)
	rl.mu.Lock()
	defer rl.mu.Unlock()
	for uid, ub := range rl.users {
		idle := true
		for _, b := range ub {
			b.mu.Lock()
			if b.lastFill.After(cutoff) {
				idle = false
			}
			b.mu.Unlock()
			if !idle {
				break
			}
		}
		if idle {
			delete(rl.users, uid)
		}
	}
	rl.cleanAt = time.Now().Add(10 * time.Minute)
}

// Reset clears all buckets for a user (e.g. after a GM pardon).
func (rl *UserRateLimiter) Reset(userID int64) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	delete(rl.users, userID)
}

// ─── Package-level singleton ──────────────────────────────────────────────────

// RL is the default package-level UserRateLimiter. Use security.RL.Allow().
var RL = NewUserRateLimiter(DefaultLimits)
