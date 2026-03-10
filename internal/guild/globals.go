package guild

import "database/sql"

// Global singletons.
// Call InitDB(db) at startup (after database.Connect()) to switch from the
// in-memory fallback to the persistent DB-backed store.
var (
	globalStore    Store    = NewMemStore() // replaced by InitDB
	globalWarStore          = NewMemWarStore()

	// GlobalService exposes all guild membership and bank operations.
	GlobalService = NewGuildService(globalStore)

	// GlobalWarService exposes war scheduling and scoring.
	GlobalWarService = NewWarService(globalStore, globalWarStore)
)

// InitDB replaces the in-memory store with a DB-backed store.
// Must be called once after the database connection is established.
func InitDB(db *sql.DB) {
	s := NewDBStore(db)
	globalStore = s
	GlobalService = NewGuildService(s)
	GlobalWarService = NewWarService(s, globalWarStore)
}

// ─── Accessor helpers ─────────────────────────────────────────────────────────

// GetMember returns the membership record for a player.
func (s *GuildService) GetMember(playerID int64) (*Member, error) {
	return s.store.GetMember(playerID)
}

// GetGuild returns a guild by ID.
func (s *GuildService) GetGuild(guildID int64) (*Guild, error) {
	return s.store.GetByID(guildID)
}

// ListMembers returns all members of a guild.
func (s *GuildService) ListMembers(guildID int64) ([]*Member, error) {
	return s.store.ListMembers(guildID)
}

// ListGuilds returns a paginated list of all guilds.
func ListGuilds(limit, offset int) ([]*Guild, error) {
	return globalStore.List(limit, offset)
}
