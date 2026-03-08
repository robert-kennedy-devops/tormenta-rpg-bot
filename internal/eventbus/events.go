// Package eventbus provides a lightweight in-process publish/subscribe event bus.
// All game systems publish events here; any number of subscribers react
// asynchronously without coupling the publisher to the consumer.
//
// Events are processed in a dedicated goroutine pool so that slow subscribers
// never block the game loop.
package eventbus

import "time"

// ─── Event kinds ──────────────────────────────────────────────────────────────

// Kind identifies an event type.
type Kind string

const (
	// Player lifecycle
	KindPlayerRegistered Kind = "PLAYER_REGISTERED"
	KindPlayerLevelUp    Kind = "PLAYER_LEVEL_UP"
	KindPlayerDeath      Kind = "PLAYER_DEATH"
	KindPlayerLogin      Kind = "PLAYER_LOGIN"

	// Combat
	KindMonsterKilled  Kind = "MONSTER_KILLED"
	KindBossKilled     Kind = "BOSS_KILLED"
	KindCriticalHit    Kind = "CRITICAL_HIT"
	KindPlayerFled     Kind = "PLAYER_FLED"

	// Items & Economy
	KindItemDropped       Kind = "ITEM_DROPPED"
	KindItemCrafted       Kind = "ITEM_CRAFTED"
	KindTradeCompleted    Kind = "TRADE_COMPLETED"
	KindAuctionSold       Kind = "AUCTION_SOLD"
	KindMarketListing     Kind = "MARKET_LISTING"
	KindGoldSink          Kind = "GOLD_SINK"

	// Guilds
	KindGuildCreated    Kind = "GUILD_CREATED"
	KindGuildDisbanded  Kind = "GUILD_DISBANDED"
	KindGuildLevelUp    Kind = "GUILD_LEVEL_UP"
	KindGuildWarStarted Kind = "GUILD_WAR_STARTED"
	KindGuildWarEnded   Kind = "GUILD_WAR_ENDED"
	KindTerritoryCapture Kind = "TERRITORY_CAPTURE"

	// World events
	KindWorldBossSpawned  Kind = "WORLD_BOSS_SPAWNED"
	KindWorldBossKilled   Kind = "WORLD_BOSS_KILLED"
	KindRaidCompleted     Kind = "RAID_COMPLETED"
	KindSeasonEnded       Kind = "SEASON_ENDED"
	KindSeasonStarted     Kind = "SEASON_STARTED"

	// PVP
	KindPVPMatchStarted Kind = "PVP_MATCH_STARTED"
	KindPVPMatchEnded   Kind = "PVP_MATCH_ENDED"
	KindRankingUpdate   Kind = "RANKING_UPDATE"

	// System
	KindInflationWarning Kind = "INFLATION_WARNING"
	KindInflationCritical Kind = "INFLATION_CRITICAL"
)

// ─── Event payload ────────────────────────────────────────────────────────────

// Event is the envelope published to the bus.
type Event struct {
	Kind      Kind
	OccurredAt time.Time
	PlayerID  int64  // primary player (0 if system event)
	GuildID   int64  // guild involved (0 if none)
	EntityID  string // item ID, boss ID, territory ID etc.
	IntValue  int    // generic int payload (xp, gold, damage, level...)
	StrValue  string // generic string payload (name, message, etc.)
	Extra     map[string]any // arbitrary extra data (avoid hot paths)
}

// NewEvent constructs a minimal event.
func NewEvent(kind Kind) Event {
	return Event{Kind: kind, OccurredAt: time.Now()}
}

// WithPlayer sets the primary player ID.
func (e Event) WithPlayer(id int64) Event { e.PlayerID = id; return e }

// WithGuild sets the guild ID.
func (e Event) WithGuild(id int64) Event { e.GuildID = id; return e }

// WithEntity sets the entity ID (boss, item, territory).
func (e Event) WithEntity(id string) Event { e.EntityID = id; return e }

// WithInt sets a generic integer value.
func (e Event) WithInt(v int) Event { e.IntValue = v; return e }

// WithStr sets a generic string value.
func (e Event) WithStr(v string) Event { e.StrValue = v; return e }
