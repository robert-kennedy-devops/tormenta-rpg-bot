package guild

// ─── Guild bank ───────────────────────────────────────────────────────────────

// BankResult is the outcome of a bank operation.
type BankResult struct {
	NewBalance int
	TxAmount   int
	TxType     string // "deposit" | "withdraw"
}

// Deposit adds gold to the guild bank.
// Any member can deposit; the deposit counts as contribution.
func (s *GuildService) Deposit(depositorID int64, amount int) (BankResult, error) {
	if amount <= 0 {
		return BankResult{}, ErrInsufficientFunds
	}
	m, err := s.store.GetMember(depositorID)
	if err != nil {
		return BankResult{}, ErrNotInGuild
	}
	g, err := s.store.GetByID(m.GuildID)
	if err != nil {
		return BankResult{}, err
	}
	g.BankGold += amount
	m.Contribution += amount
	if saveErr := s.store.Update(g); saveErr != nil {
		return BankResult{}, saveErr
	}
	return BankResult{NewBalance: g.BankGold, TxAmount: amount, TxType: "deposit"}, nil
}

// Withdraw removes gold from the guild bank (officer/leader only).
func (s *GuildService) Withdraw(withdrawerID int64, amount int) (BankResult, error) {
	if amount <= 0 {
		return BankResult{}, ErrInsufficientFunds
	}
	m, err := s.store.GetMember(withdrawerID)
	if err != nil {
		return BankResult{}, ErrNotInGuild
	}
	if !m.CanManageBank() {
		return BankResult{}, ErrNotGuildOfficer
	}
	g, err := s.store.GetByID(m.GuildID)
	if err != nil {
		return BankResult{}, err
	}
	if g.BankGold < amount {
		return BankResult{}, ErrInsufficientFunds
	}
	g.BankGold -= amount
	if saveErr := s.store.Update(g); saveErr != nil {
		return BankResult{}, saveErr
	}
	return BankResult{NewBalance: g.BankGold, TxAmount: amount, TxType: "withdraw"}, nil
}

// AddGuildXP adds XP to the guild and handles level-up.
// Typically called when guild members kill bosses, complete dungeons, etc.
func (s *GuildService) AddGuildXP(guildID int64, xp int) (levelled bool, newLevel int, err error) {
	g, err := s.store.GetByID(guildID)
	if err != nil {
		return
	}
	g.XP += xp
	if g.XP >= g.XPNext && g.Level < 10 {
		g.Level++
		g.XP -= g.XPNext
		g.XPNext = GuildXPForLevel(g.Level)
		g.MaxMembers = MaxMembersForLevel(g.Level)
		levelled = true
		newLevel = g.Level
	}
	err = s.store.Update(g)
	return
}
