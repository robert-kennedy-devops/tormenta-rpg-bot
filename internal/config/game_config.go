// Package config centralises every hardcoded game-balance and infrastructure
// constant so that tuning never requires hunting through business-logic files.
package config

import (
	"os"
	"strconv"
	"time"
)

// ── Database ──────────────────────────────────────────────────────────────────

// DB returns database connection settings derived from the environment,
// falling back to safe development defaults.
type DBConfig struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

func DB() DBConfig {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://tormenta:tormenta123@localhost:5432/tormenta_rpg?sslmode=disable"
	}
	return DBConfig{
		DSN:             dsn,
		MaxOpenConns:    envInt("DB_MAX_OPEN", 25),
		MaxIdleConns:    envInt("DB_MAX_IDLE", 10),
		ConnMaxLifetime: envDur("DB_CONN_LIFETIME", 5*time.Minute),
	}
}

// ── Energy ────────────────────────────────────────────────────────────────────

const (
	EnergyMaxFree    = 100
	EnergyMaxVIP     = 200
	EnergyRegenFree  = 10 * time.Minute
	EnergyRegenVIP   = 5 * time.Minute
)

// ── Economy ───────────────────────────────────────────────────────────────────

const (
	MarketTaxPct    = 5    // percent taken on every market sale
	AuctionTaxPct   = 8    // percent taken on auction settlement
	GuildCreateCost = 1000 // gold to found a guild

	// Inflation thresholds (economy_manager)
	InflationWarnPct  = 15
	InflationHighPct  = 30
	DeflationWarnPct  = -15
)

// ── Market ────────────────────────────────────────────────────────────────────

const (
	ListingDuration = 48 * time.Hour
	AuctionDuration = 24 * time.Hour
	MaxListingsPerPlayer = 10
)

// ── World Boss ────────────────────────────────────────────────────────────────

const (
	BossSpawnInterval    = 12 * time.Hour
	BossWindowDuration   = 30 * time.Minute
	BossMinParticipants  = 1
)

// ── Seasons ───────────────────────────────────────────────────────────────────

const (
	SeasonDuration = 90 * 24 * time.Hour // 90 days
)

// ── PvP ───────────────────────────────────────────────────────────────────────

const (
	EloK             = 32  // Elo K-factor
	EloDefault       = 1000
	MatchmakingRange = 200 // initial Elo window
)

// ── Forge / Crafting ──────────────────────────────────────────────────────────

const (
	ForgeMaxLevel      = 5
	ForgeBreakThreshold = 4 // items can break on attempts beyond this level
)

// ── Character progression ─────────────────────────────────────────────────────

const (
	MaxLevel    = 100
	StartGold   = 50
	StartDiamonds = 5
)

// ── Dungeons ──────────────────────────────────────────────────────────────────

const (
	DungeonRoomsMin = 3
	DungeonRoomsMax = 8
	DungeonBossChance = 25 // percent chance of a boss room
)

// ── Concurrency & Workers ─────────────────────────────────────────────────────

const (
	CombatWorkers     = 16
	CombatQueueSize   = 1000
	EventBusWorkers   = 8
	EventBusQueueSize = 10_000
	EconomyTickEvery  = 10 * time.Minute
)

// ── Rate Limits ───────────────────────────────────────────────────────────────

const (
	CallbackRateLimit    = 36            // max callbacks per window
	CallbackRateWindow   = 2 * time.Second
	CallbackDedupWindow  = 900 * time.Millisecond
)

// ── helpers ───────────────────────────────────────────────────────────────────

func envInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func envDur(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}
