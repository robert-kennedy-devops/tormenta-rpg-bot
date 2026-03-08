// Package cache provides a unified caching layer.
// The RedisCache wraps a real Redis connection when available; it falls back
// gracefully to the existing in-memory TTL cache so that the game works without
// Redis during development / low-scale deployment.
//
// Redis is used for:
//   - Player session data (cooldowns, state)
//   - PVP / boss leaderboards
//   - Economy snapshots
//   - Distributed locks (for atomic operations across multiple instances)
//
// To enable Redis: set the REDIS_ADDR environment variable (e.g. "localhost:6379").
// When REDIS_ADDR is not set the cache silently uses the in-memory fallback.
package cache

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// ─── Interface ────────────────────────────────────────────────────────────────

// Cache is the generic key/value cache interface used throughout the game.
// All keys are strings; values are JSON-serialised Go values.
type Cache interface {
	Set(key string, value any, ttl time.Duration) error
	Get(key string, dest any) (bool, error)
	Delete(key string) error
	Exists(key string) bool
	// Atomic increment (useful for counters, cooldown tracking).
	Incr(key string) (int64, error)
	// Leaderboard operations (sorted set).
	ZSet(board, member string, score float64) error
	ZTopN(board string, n int) ([]ZEntry, error)
	// Lock / unlock for distributed mutual exclusion.
	Lock(key string, ttl time.Duration) (bool, error)
	Unlock(key string) error
}

// ZEntry is one element of a sorted-set leaderboard result.
type ZEntry struct {
	Member string
	Score  float64
}

// ─── In-memory implementation (no-Redis fallback) ─────────────────────────────

// memItem stores one cached value with its expiry.
type memItem struct {
	data      []byte
	expiresAt time.Time
}

// MemCache is a simple in-memory cache implementing the Cache interface.
// It is safe for concurrent use and serves as the Redis fallback.
type MemCache struct {
	mu     sync.RWMutex
	items  map[string]memItem
	boards map[string]map[string]float64 // sorted sets
	locks  map[string]time.Time
}

// NewMemCache creates an empty MemCache.
func NewMemCache() *MemCache {
	return &MemCache{
		items:  make(map[string]memItem),
		boards: make(map[string]map[string]float64),
		locks:  make(map[string]time.Time),
	}
}

// Global is the singleton cache (starts as MemCache; replaced by Redis if configured).
var Global Cache = NewMemCache()

// ─── MemCache implementation ──────────────────────────────────────────────────

func (c *MemCache) Set(key string, value any, ttl time.Duration) error {
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}
	c.mu.Lock()
	c.items[key] = memItem{data: b, expiresAt: time.Now().Add(ttl)}
	c.mu.Unlock()
	return nil
}

func (c *MemCache) Get(key string, dest any) (bool, error) {
	c.mu.RLock()
	item, ok := c.items[key]
	c.mu.RUnlock()
	if !ok || time.Now().After(item.expiresAt) {
		return false, nil
	}
	return true, json.Unmarshal(item.data, dest)
}

func (c *MemCache) Delete(key string) error {
	c.mu.Lock()
	delete(c.items, key)
	c.mu.Unlock()
	return nil
}

func (c *MemCache) Exists(key string) bool {
	c.mu.RLock()
	item, ok := c.items[key]
	c.mu.RUnlock()
	return ok && time.Now().Before(item.expiresAt)
}

func (c *MemCache) Incr(key string) (int64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	var n int64
	if item, ok := c.items[key]; ok && time.Now().Before(item.expiresAt) {
		_ = json.Unmarshal(item.data, &n)
	}
	n++
	b, _ := json.Marshal(n)
	ttl := time.Minute // default TTL for counters
	c.items[key] = memItem{data: b, expiresAt: time.Now().Add(ttl)}
	return n, nil
}

func (c *MemCache) ZSet(board, member string, score float64) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.boards[board] == nil {
		c.boards[board] = make(map[string]float64)
	}
	c.boards[board][member] = score
	return nil
}

func (c *MemCache) ZTopN(board string, n int) ([]ZEntry, error) {
	c.mu.RLock()
	b := c.boards[board]
	c.mu.RUnlock()
	// Build and sort
	entries := make([]ZEntry, 0, len(b))
	for m, s := range b {
		entries = append(entries, ZEntry{Member: m, Score: s})
	}
	// Simple insertion sort (small N expected)
	for i := 0; i < len(entries); i++ {
		for j := i + 1; j < len(entries); j++ {
			if entries[j].Score > entries[i].Score {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}
	if n > 0 && len(entries) > n {
		entries = entries[:n]
	}
	return entries, nil
}

func (c *MemCache) Lock(key string, ttl time.Duration) (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if exp, ok := c.locks[key]; ok && time.Now().Before(exp) {
		return false, nil // already locked
	}
	c.locks[key] = time.Now().Add(ttl)
	return true, nil
}

func (c *MemCache) Unlock(key string) error {
	c.mu.Lock()
	delete(c.locks, key)
	c.mu.Unlock()
	return nil
}

// ─── Cache key helpers ────────────────────────────────────────────────────────

// PlayerCooldownKey returns the cache key for a player's skill cooldown.
func PlayerCooldownKey(playerID int64, skillID string) string {
	return fmt.Sprintf("cooldown:%d:%s", playerID, skillID)
}

// PlayerSessionKey returns the cache key for player session data.
func PlayerSessionKey(playerID int64) string {
	return fmt.Sprintf("session:%d", playerID)
}

// PVPLeaderboardKey is the sorted set key for PVP ratings.
const PVPLeaderboardKey = "pvp:leaderboard"

// BossLeaderboardKey returns a key for the damage leaderboard of an active boss.
func BossLeaderboardKey(bossInstanceID string) string {
	return fmt.Sprintf("boss:leaderboard:%s", bossInstanceID)
}

// EconomySnapshotKey is the cache key for the latest economy snapshot.
const EconomySnapshotKey = "economy:snapshot"
