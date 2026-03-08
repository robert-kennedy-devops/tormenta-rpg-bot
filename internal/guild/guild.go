// Package guild implements the complete guild system: creation, membership,
// bank, perks, chat and guild wars for territory control.
package guild

import (
	"errors"
	"sync"
	"time"
)

// ─── Errors ───────────────────────────────────────────────────────────────────

var (
	ErrGuildNotFound      = errors.New("guilda não encontrada")
	ErrAlreadyInGuild     = errors.New("você já pertence a uma guilda")
	ErrNotInGuild         = errors.New("você não pertence a uma guilda")
	ErrNotGuildLeader     = errors.New("apenas o líder pode executar esta ação")
	ErrNotGuildOfficer    = errors.New("apenas oficiais podem executar esta ação")
	ErrGuildFull          = errors.New("a guilda está cheia")
	ErrInsufficientFunds  = errors.New("banco da guilda sem fundos suficientes")
	ErrGuildNameTaken     = errors.New("nome de guilda já em uso")
)

// ─── Rank ─────────────────────────────────────────────────────────────────────

// GuildRank defines a member's role within the guild.
type GuildRank string

const (
	RankLeader  GuildRank = "leader"
	RankOfficer GuildRank = "officer"
	RankMember  GuildRank = "member"
	RankRecruit GuildRank = "recruit"
)

// ─── Guild model ──────────────────────────────────────────────────────────────

// Guild is the core guild entity.
type Guild struct {
	ID          int64
	Name        string
	Tag         string // [TAG] short code, max 5 chars
	Description string
	Emoji       string
	LeaderID    int64
	Level       int    // 1–10
	XP          int
	XPNext      int
	MaxMembers  int    // grows with level
	BankGold    int
	TerritoryID string // ID of controlled territory (empty = none)
	CreatedAt   time.Time
	Perks       GuildPerks
	BuffExpiry  time.Time // when the active guild buff expires
	ActiveBuff  string    // buff key (empty = none)
}

// GuildXPForLevel returns the XP needed to reach the next level.
func GuildXPForLevel(level int) int {
	return 1000 * level * level
}

// MaxMembersForLevel returns the member cap at a guild level.
func MaxMembersForLevel(level int) int {
	return 10 + level*5 // lv1=15, lv10=60
}

// ─── Guild store interface ────────────────────────────────────────────────────

// Store abstracts guild persistence.
type Store interface {
	Create(g *Guild) error
	GetByID(id int64) (*Guild, error)
	GetByLeader(leaderID int64) (*Guild, error)
	GetByName(name string) (*Guild, error)
	Update(g *Guild) error
	Delete(id int64) error
	List(limit, offset int) ([]*Guild, error)
	// Member operations
	AddMember(guildID, playerID int64, rank GuildRank) error
	RemoveMember(guildID, playerID int64) error
	GetMember(playerID int64) (*Member, error)
	ListMembers(guildID int64) ([]*Member, error)
	UpdateMemberRank(guildID, playerID int64, rank GuildRank) error
}

// ─── In-memory store ──────────────────────────────────────────────────────────

// MemStore is an in-memory guild store.
type MemStore struct {
	mu      sync.RWMutex
	guilds  map[int64]*Guild
	members map[int64]*Member // keyed by playerID
	names   map[string]int64  // name → guildID
	seq     int64
}

// NewMemStore creates an empty store.
func NewMemStore() *MemStore {
	return &MemStore{
		guilds:  make(map[int64]*Guild),
		members: make(map[int64]*Member),
		names:   make(map[string]int64),
	}
}

func (s *MemStore) Create(g *Guild) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.names[g.Name]; exists {
		return ErrGuildNameTaken
	}
	s.seq++
	g.ID = s.seq
	s.guilds[g.ID] = g
	s.names[g.Name] = g.ID
	return nil
}

func (s *MemStore) GetByID(id int64) (*Guild, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	g, ok := s.guilds[id]
	if !ok {
		return nil, ErrGuildNotFound
	}
	return g, nil
}

func (s *MemStore) GetByLeader(leaderID int64) (*Guild, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, g := range s.guilds {
		if g.LeaderID == leaderID {
			return g, nil
		}
	}
	return nil, ErrGuildNotFound
}

func (s *MemStore) GetByName(name string) (*Guild, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	id, ok := s.names[name]
	if !ok {
		return nil, ErrGuildNotFound
	}
	return s.guilds[id], nil
}

func (s *MemStore) Update(g *Guild) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.guilds[g.ID]; !ok {
		return ErrGuildNotFound
	}
	s.guilds[g.ID] = g
	return nil
}

func (s *MemStore) Delete(id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	g, ok := s.guilds[id]
	if !ok {
		return ErrGuildNotFound
	}
	delete(s.names, g.Name)
	delete(s.guilds, id)
	return nil
}

func (s *MemStore) List(limit, offset int) ([]*Guild, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []*Guild
	for _, g := range s.guilds {
		out = append(out, g)
	}
	if offset >= len(out) {
		return nil, nil
	}
	out = out[offset:]
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

func (s *MemStore) AddMember(guildID, playerID int64, rank GuildRank) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.members[playerID] = &Member{
		GuildID:  guildID,
		PlayerID: playerID,
		Rank:     rank,
		JoinedAt: time.Now(),
	}
	return nil
}

func (s *MemStore) RemoveMember(guildID, playerID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.members, playerID)
	return nil
}

func (s *MemStore) GetMember(playerID int64) (*Member, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	m, ok := s.members[playerID]
	if !ok {
		return nil, ErrNotInGuild
	}
	return m, nil
}

func (s *MemStore) ListMembers(guildID int64) ([]*Member, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []*Member
	for _, m := range s.members {
		if m.GuildID == guildID {
			out = append(out, m)
		}
	}
	return out, nil
}

func (s *MemStore) UpdateMemberRank(guildID, playerID int64, rank GuildRank) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	m, ok := s.members[playerID]
	if !ok || m.GuildID != guildID {
		return ErrNotInGuild
	}
	m.Rank = rank
	return nil
}
