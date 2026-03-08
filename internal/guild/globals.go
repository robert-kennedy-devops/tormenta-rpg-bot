package guild

// Global singletons — backed by in-memory stores for now.
// Replace with DB-backed stores when persistence is added.
var (
	globalStore    = NewMemStore()
	globalWarStore = NewMemWarStore()

	// GlobalService exposes all guild membership and bank operations.
	GlobalService = NewGuildService(globalStore)

	// GlobalWarService exposes war scheduling and scoring.
	GlobalWarService = NewWarService(globalStore, globalWarStore)
)

// ─── Accessor helpers ────────────────────────────────────────────────────────

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
