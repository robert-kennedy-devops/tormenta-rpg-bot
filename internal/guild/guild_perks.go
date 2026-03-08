package guild

import "time"

// ─── Guild perks ──────────────────────────────────────────────────────────────

// GuildPerks holds permanent passive upgrades unlocked at the guild level.
type GuildPerks struct {
	XPBonusPct     int // % bonus XP for all members
	GoldBonusPct   int // % bonus gold drops for all members
	DropRateBonus  int // % bonus to item drop rate
	MaxBankGold    int // gold storage cap
	TerritorySlots int // max territories the guild can hold
}

// PerkForLevel returns the perk set unlocked at a given guild level.
func PerkForLevel(level int) GuildPerks {
	return GuildPerks{
		XPBonusPct:     level * 3,   // +3% XP per level (max +30% at lv10)
		GoldBonusPct:   level * 2,   // +2% gold per level (max +20%)
		DropRateBonus:  level,        // +1% drop rate per level (max +10%)
		MaxBankGold:    10000 * level, // 10k * level
		TerritorySlots: level / 3,    // first slot at lv3, max 3 slots at lv9+
	}
}

// ─── Guild buff ───────────────────────────────────────────────────────────────

// GuildBuffDef describes a purchasable temporary guild-wide buff.
type GuildBuffDef struct {
	Key         string
	Name        string
	Emoji       string
	Description string
	Cost        int           // gold from guild bank
	Duration    time.Duration
	XPBonus     float64 // multiplier bonus (0.2 = +20%)
	GoldBonus   float64
	DropBonus   float64
	MinGuildLvl int
}

// Buffs lists all available guild buffs.
var Buffs = []GuildBuffDef{
	{
		Key: "exp_boost", Name: "Bênção do Sábio", Emoji: "📚",
		Description: "+30% XP para todos os membros por 2 horas.",
		Cost: 500, Duration: 2 * time.Hour, XPBonus: 0.30, MinGuildLvl: 1,
	},
	{
		Key: "gold_rush", Name: "Febre do Ouro", Emoji: "💰",
		Description: "+40% gold em drops por 2 horas.",
		Cost: 700, Duration: 2 * time.Hour, GoldBonus: 0.40, MinGuildLvl: 3,
	},
	{
		Key: "loot_frenzy", Name: "Pilhagem Frenética", Emoji: "🎁",
		Description: "+25% drop rate de itens por 1 hora.",
		Cost: 800, Duration: time.Hour, DropBonus: 0.25, MinGuildLvl: 5,
	},
	{
		Key: "warlords_blessing", Name: "Bênção do Senhor da Guerra", Emoji: "⚔️",
		Description: "+50% XP e +30% gold por 1 hora. Apenas guildas lv8+.",
		Cost: 2000, Duration: time.Hour, XPBonus: 0.50, GoldBonus: 0.30, MinGuildLvl: 8,
	},
}

// GetBuff returns a buff definition by key.
func GetBuff(key string) (GuildBuffDef, bool) {
	for _, b := range Buffs {
		if b.Key == key {
			return b, true
		}
	}
	return GuildBuffDef{}, false
}

// ActivateBuff purchases and activates a guild buff (officer/leader only).
func (s *GuildService) ActivateBuff(activatorID int64, buffKey string) error {
	m, err := s.store.GetMember(activatorID)
	if err != nil {
		return ErrNotInGuild
	}
	if !m.CanManageBank() {
		return ErrNotGuildOfficer
	}
	buff, ok := GetBuff(buffKey)
	if !ok {
		return ErrGuildNotFound
	}
	g, err := s.store.GetByID(m.GuildID)
	if err != nil {
		return err
	}
	if g.Level < buff.MinGuildLvl {
		return ErrNotGuildOfficer
	}
	if g.BankGold < buff.Cost {
		return ErrInsufficientFunds
	}
	g.BankGold -= buff.Cost
	g.ActiveBuff = buffKey
	g.BuffExpiry = time.Now().Add(buff.Duration)
	return s.store.Update(g)
}

// ActiveBuffFor returns the active buff definition for a guild (if any).
func (s *GuildService) ActiveBuffFor(guildID int64) (*GuildBuffDef, bool) {
	g, err := s.store.GetByID(guildID)
	if err != nil || g.ActiveBuff == "" || time.Now().After(g.BuffExpiry) {
		return nil, false
	}
	b, ok := GetBuff(g.ActiveBuff)
	if !ok {
		return nil, false
	}
	return &b, true
}
