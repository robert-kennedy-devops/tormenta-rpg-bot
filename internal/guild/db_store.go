// db_store.go — PostgreSQL-backed implementation of guild.Store.
//
// Replaces MemStore so guild data survives server restarts.
// The schema is created by database.Migrate() (guilds + guild_members tables).
package guild

import (
	"database/sql"
	"time"
)

// DBStore implements Store using a *sql.DB.
type DBStore struct {
	db *sql.DB
}

// NewDBStore wraps an existing database connection.
func NewDBStore(db *sql.DB) *DBStore {
	return &DBStore{db: db}
}

// ── Guild CRUD ────────────────────────────────────────────────────────────────

func (s *DBStore) Create(g *Guild) error {
	err := s.db.QueryRow(`
		INSERT INTO guilds
			(name, tag, description, emoji, leader_id, level, xp, xp_next, max_members, bank_gold,
			 territory_id, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,NOW(),NOW())
		RETURNING id`,
		g.Name, g.Tag, g.Description, g.Emoji, g.LeaderID,
		g.Level, g.XP, g.XPNext, g.MaxMembers, g.BankGold,
		g.TerritoryID,
	).Scan(&g.ID)
	if err != nil {
		if isUniqueViolation(err) {
			return ErrGuildNameTaken
		}
		return err
	}
	return nil
}

func (s *DBStore) GetByID(id int64) (*Guild, error) {
	return s.scanGuild(s.db.QueryRow(`
		SELECT id,name,tag,description,emoji,leader_id,level,xp,xp_next,max_members,
		       bank_gold,territory_id,active_buff,buff_expiry,created_at
		FROM guilds WHERE id=$1`, id))
}

func (s *DBStore) GetByLeader(leaderID int64) (*Guild, error) {
	return s.scanGuild(s.db.QueryRow(`
		SELECT id,name,tag,description,emoji,leader_id,level,xp,xp_next,max_members,
		       bank_gold,territory_id,active_buff,buff_expiry,created_at
		FROM guilds WHERE leader_id=$1`, leaderID))
}

func (s *DBStore) GetByName(name string) (*Guild, error) {
	return s.scanGuild(s.db.QueryRow(`
		SELECT id,name,tag,description,emoji,leader_id,level,xp,xp_next,max_members,
		       bank_gold,territory_id,active_buff,buff_expiry,created_at
		FROM guilds WHERE name=$1`, name))
}

func (s *DBStore) Update(g *Guild) error {
	_, err := s.db.Exec(`
		UPDATE guilds SET
			name=$1, tag=$2, description=$3, emoji=$4, leader_id=$5,
			level=$6, xp=$7, xp_next=$8, max_members=$9, bank_gold=$10,
			territory_id=$11, active_buff=$12, buff_expiry=$13, updated_at=NOW()
		WHERE id=$14`,
		g.Name, g.Tag, g.Description, g.Emoji, g.LeaderID,
		g.Level, g.XP, g.XPNext, g.MaxMembers, g.BankGold,
		g.TerritoryID, g.ActiveBuff, g.BuffExpiry,
		g.ID,
	)
	return err
}

func (s *DBStore) Delete(id int64) error {
	_, err := s.db.Exec(`DELETE FROM guilds WHERE id=$1`, id)
	return err
}

func (s *DBStore) List(limit, offset int) ([]*Guild, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.db.Query(`
		SELECT id,name,tag,description,emoji,leader_id,level,xp,xp_next,max_members,
		       bank_gold,territory_id,active_buff,buff_expiry,created_at
		FROM guilds ORDER BY level DESC, xp DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*Guild
	for rows.Next() {
		g := &Guild{}
		if err := s.scan(rows, g); err != nil {
			return nil, err
		}
		out = append(out, g)
	}
	return out, rows.Err()
}

// ── Member operations ─────────────────────────────────────────────────────────

func (s *DBStore) AddMember(guildID, playerID int64, rank GuildRank) error {
	_, err := s.db.Exec(`
		INSERT INTO guild_members (guild_id, player_id, rank, joined_at, last_online)
		VALUES ($1,$2,$3,NOW(),NOW())
		ON CONFLICT (guild_id, player_id) DO UPDATE SET rank=$3, last_online=NOW()`,
		guildID, playerID, string(rank))
	return err
}

func (s *DBStore) RemoveMember(guildID, playerID int64) error {
	_, err := s.db.Exec(`DELETE FROM guild_members WHERE guild_id=$1 AND player_id=$2`, guildID, playerID)
	return err
}

func (s *DBStore) GetMember(playerID int64) (*Member, error) {
	m := &Member{}
	var rank string
	err := s.db.QueryRow(`
		SELECT guild_id, player_id, player_name, rank, contribution, joined_at, last_online
		FROM guild_members WHERE player_id=$1`, playerID).
		Scan(&m.GuildID, &m.PlayerID, &m.PlayerName, &rank, &m.Contribution, &m.JoinedAt, &m.LastOnline)
	if err == sql.ErrNoRows {
		return nil, ErrNotInGuild
	}
	if err != nil {
		return nil, err
	}
	m.Rank = GuildRank(rank)
	return m, nil
}

func (s *DBStore) ListMembers(guildID int64) ([]*Member, error) {
	rows, err := s.db.Query(`
		SELECT guild_id, player_id, player_name, rank, contribution, joined_at, last_online
		FROM guild_members WHERE guild_id=$1 ORDER BY
			CASE rank WHEN 'leader' THEN 0 WHEN 'officer' THEN 1 WHEN 'member' THEN 2 ELSE 3 END`,
		guildID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*Member
	for rows.Next() {
		m := &Member{}
		var rank string
		if err := rows.Scan(&m.GuildID, &m.PlayerID, &m.PlayerName, &rank, &m.Contribution, &m.JoinedAt, &m.LastOnline); err != nil {
			return nil, err
		}
		m.Rank = GuildRank(rank)
		out = append(out, m)
	}
	return out, rows.Err()
}

func (s *DBStore) UpdateMemberRank(guildID, playerID int64, rank GuildRank) error {
	_, err := s.db.Exec(`
		UPDATE guild_members SET rank=$1 WHERE guild_id=$2 AND player_id=$3`,
		string(rank), guildID, playerID)
	return err
}

// ── scan helpers ──────────────────────────────────────────────────────────────

type rower interface {
	Scan(dest ...any) error
}

func (s *DBStore) scanGuild(row *sql.Row) (*Guild, error) {
	g := &Guild{}
	if err := row.Scan(
		&g.ID, &g.Name, &g.Tag, &g.Description, &g.Emoji, &g.LeaderID,
		&g.Level, &g.XP, &g.XPNext, &g.MaxMembers,
		&g.BankGold, &g.TerritoryID, &g.ActiveBuff, &g.BuffExpiry, &g.CreatedAt,
	); err == sql.ErrNoRows {
		return nil, ErrGuildNotFound
	} else if err != nil {
		return nil, err
	}
	return g, nil
}

func (s *DBStore) scan(rows *sql.Rows, g *Guild) error {
	return rows.Scan(
		&g.ID, &g.Name, &g.Tag, &g.Description, &g.Emoji, &g.LeaderID,
		&g.Level, &g.XP, &g.XPNext, &g.MaxMembers,
		&g.BankGold, &g.TerritoryID, &g.ActiveBuff, &g.BuffExpiry, &g.CreatedAt,
	)
}

// ── postgres error helpers ────────────────────────────────────────────────────

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	// lib/pq error code 23505 = unique_violation
	return len(err.Error()) > 5 && err.Error()[:5] == "pq: d" ||
		contains(err.Error(), "23505") || contains(err.Error(), "unique")
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsStr(s, sub))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

// UpdateMemberName persists the player name on the membership row (called after char rename).
func (s *DBStore) UpdateMemberName(playerID int64, name string) error {
	_, err := s.db.Exec(`UPDATE guild_members SET player_name=$1 WHERE player_id=$2`, name, playerID)
	return err
}

// UpdateMemberOnline refreshes the last_online timestamp.
func (s *DBStore) UpdateMemberOnline(playerID int64) error {
	_, err := s.db.Exec(`UPDATE guild_members SET last_online=$1 WHERE player_id=$2`, time.Now(), playerID)
	return err
}
