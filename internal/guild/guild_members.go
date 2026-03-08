package guild

import "time"

// ─── Member model ─────────────────────────────────────────────────────────────

// Member represents a player's membership in a guild.
type Member struct {
	GuildID      int64
	PlayerID     int64
	PlayerName   string
	Rank         GuildRank
	JoinedAt     time.Time
	Contribution int // gold contributed to bank
	LastOnline   time.Time
}

// CanInvite returns true if the member's rank allows inviting new members.
func (m *Member) CanInvite() bool {
	return m.Rank == RankLeader || m.Rank == RankOfficer
}

// CanKick returns true if the member can kick someone of the target rank.
func (m *Member) CanKick(target GuildRank) bool {
	if m.Rank == RankLeader {
		return target != RankLeader
	}
	if m.Rank == RankOfficer {
		return target == RankMember || target == RankRecruit
	}
	return false
}

// CanManageBank returns true if the member can deposit/withdraw from the guild bank.
func (m *Member) CanManageBank() bool {
	return m.Rank == RankLeader || m.Rank == RankOfficer
}

// ─── GuildService — member operations ─────────────────────────────────────────

// GuildService provides all guild management operations.
type GuildService struct {
	store Store
}

// NewGuildService creates a service backed by the given store.
func NewGuildService(store Store) *GuildService {
	return &GuildService{store: store}
}

// CreateGuild creates a new guild with the caller as leader.
func (s *GuildService) CreateGuild(leaderID int64, leaderName, guildName, tag, description string) (*Guild, error) {
	// Check player not already in a guild
	if _, err := s.store.GetMember(leaderID); err == nil {
		return nil, ErrAlreadyInGuild
	}

	now := time.Now()
	g := &Guild{
		Name:        guildName,
		Tag:         tag,
		Description: description,
		Emoji:       "⚔️",
		LeaderID:    leaderID,
		Level:       1,
		XPNext:      GuildXPForLevel(1),
		MaxMembers:  MaxMembersForLevel(1),
		CreatedAt:   now,
	}
	if err := s.store.Create(g); err != nil {
		return nil, err
	}
	if err := s.store.AddMember(g.ID, leaderID, RankLeader); err != nil {
		_ = s.store.Delete(g.ID)
		return nil, err
	}
	return g, nil
}

// Invite adds a player to a guild (inviter must be officer or leader).
func (s *GuildService) Invite(inviterID, inviteeID int64, inviteeName string) error {
	inviter, err := s.store.GetMember(inviterID)
	if err != nil {
		return ErrNotInGuild
	}
	if !inviter.CanInvite() {
		return ErrNotGuildOfficer
	}
	if _, err := s.store.GetMember(inviteeID); err == nil {
		return ErrAlreadyInGuild
	}
	g, err := s.store.GetByID(inviter.GuildID)
	if err != nil {
		return err
	}
	members, err := s.store.ListMembers(g.ID)
	if err != nil {
		return err
	}
	if len(members) >= g.MaxMembers {
		return ErrGuildFull
	}
	return s.store.AddMember(g.ID, inviteeID, RankRecruit)
}

// Leave removes a player from their guild. Leaders must transfer leadership first.
func (s *GuildService) Leave(playerID int64) error {
	m, err := s.store.GetMember(playerID)
	if err != nil {
		return ErrNotInGuild
	}
	if m.Rank == RankLeader {
		return ErrNotGuildLeader // leader can't just leave — must disband or transfer
	}
	return s.store.RemoveMember(m.GuildID, playerID)
}

// Kick removes a member by rank authority.
func (s *GuildService) Kick(kickerID, targetID int64) error {
	kicker, err := s.store.GetMember(kickerID)
	if err != nil {
		return ErrNotInGuild
	}
	target, err := s.store.GetMember(targetID)
	if err != nil {
		return ErrNotInGuild
	}
	if kicker.GuildID != target.GuildID {
		return ErrNotInGuild
	}
	if !kicker.CanKick(target.Rank) {
		return ErrNotGuildOfficer
	}
	return s.store.RemoveMember(target.GuildID, targetID)
}

// Promote promotes a member to officer (leader only).
func (s *GuildService) Promote(leaderID, targetID int64) error {
	leader, err := s.store.GetMember(leaderID)
	if err != nil || leader.Rank != RankLeader {
		return ErrNotGuildLeader
	}
	return s.store.UpdateMemberRank(leader.GuildID, targetID, RankOfficer)
}

// Demote demotes an officer to member (leader only).
func (s *GuildService) Demote(leaderID, targetID int64) error {
	leader, err := s.store.GetMember(leaderID)
	if err != nil || leader.Rank != RankLeader {
		return ErrNotGuildLeader
	}
	return s.store.UpdateMemberRank(leader.GuildID, targetID, RankMember)
}

// Disband deletes the guild (leader only, removes all members).
func (s *GuildService) Disband(leaderID int64) error {
	leader, err := s.store.GetMember(leaderID)
	if err != nil || leader.Rank != RankLeader {
		return ErrNotGuildLeader
	}
	members, err := s.store.ListMembers(leader.GuildID)
	if err != nil {
		return err
	}
	for _, m := range members {
		_ = s.store.RemoveMember(leader.GuildID, m.PlayerID)
	}
	return s.store.Delete(leader.GuildID)
}
