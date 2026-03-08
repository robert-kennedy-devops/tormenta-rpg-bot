package rpg

// ─── Talent system ────────────────────────────────────────────────────────────
//
// Talents are bonus specializations unlocked at levels 10, 25, 50, 75, 100.
// Unlike skill tree nodes they are not gated by prerequisites — each talent
// is a standalone bonus chosen from a class-specific list.

// Talent describes a single talent choice.
type Talent struct {
	ID          string
	Name        string
	Emoji       string
	Description string
	Class       string // "" = universal
	UnlockLevel int
	Bonus       TalentBonus
}

// TalentBonus is the mechanical effect of a talent.
type TalentBonus struct {
	FlatHP       int
	FlatMP       int
	FlatATK      int
	FlatDEF      int
	CritChance   int    // %
	XPMultiplier float64 // additive bonus (0.1 = +10% XP)
	GoldMultiplier float64
	DropRateBonus float64
	SkillCostReduction int // MP reduction for skills
}

// ─── Talent registry ──────────────────────────────────────────────────────────

// Talents is the master list of all available talents.
var Talents = []Talent{
	// ── Universal talents (available to all classes) ───────────────────────
	{
		ID: "tough", Name: "Durão", Emoji: "💪", UnlockLevel: 10,
		Description: "+50 HP máximo.",
		Bonus: TalentBonus{FlatHP: 50},
	},
	{
		ID: "sharp_mind", Name: "Mente Afiada", Emoji: "🧠", UnlockLevel: 10,
		Description: "+40 MP máximo.",
		Bonus: TalentBonus{FlatMP: 40},
	},
	{
		ID: "treasure_hunter", Name: "Caçador de Tesouros", Emoji: "💎", UnlockLevel: 25,
		Description: "+10% drop rate de todos os itens.",
		Bonus: TalentBonus{DropRateBonus: 0.10},
	},
	{
		ID: "swift_learner", Name: "Aprendizado Rápido", Emoji: "📚", UnlockLevel: 10,
		Description: "+15% XP ganho.",
		Bonus: TalentBonus{XPMultiplier: 0.15},
	},
	{
		ID: "gold_sense", Name: "Sentido do Ouro", Emoji: "💰", UnlockLevel: 25,
		Description: "+20% gold encontrado.",
		Bonus: TalentBonus{GoldMultiplier: 0.20},
	},
	{
		ID: "iron_will", Name: "Vontade de Ferro", Emoji: "🔩", UnlockLevel: 50,
		Description: "+100 HP, +50 MP máximo.",
		Bonus: TalentBonus{FlatHP: 100, FlatMP: 50},
	},
	{
		ID: "legendary_prowess", Name: "Proeza Lendária", Emoji: "🌟", UnlockLevel: 75,
		Description: "+5% crit chance, +30 ATK, +20 DEF.",
		Bonus: TalentBonus{CritChance: 5, FlatATK: 30, FlatDEF: 20},
	},
	{
		ID: "paragon", Name: "Paragão", Emoji: "👑", UnlockLevel: 100,
		Description: "+200 HP, +100 MP, +50 ATK, +40 DEF, +10% XP, +20% gold.",
		Bonus: TalentBonus{FlatHP: 200, FlatMP: 100, FlatATK: 50, FlatDEF: 40, XPMultiplier: 0.10, GoldMultiplier: 0.20},
	},

	// ── Warrior / Barbarian / Paladin ──────────────────────────────────────
	{
		ID: "berserker_mastery", Name: "Domínio do Berserk", Emoji: "💢",
		Class: "barbarian", UnlockLevel: 25,
		Description: "Aumenta a duração de Berserk em +2 turnos.",
		Bonus: TalentBonus{},
	},
	{
		ID: "holy_warrior", Name: "Guerreiro Sagrado", Emoji: "⚜️",
		Class: "paladin", UnlockLevel: 25,
		Description: "Golpe Divino causa +2d8 adicional de dano sagrado.",
		Bonus: TalentBonus{},
	},

	// ── Mage / Cleric / Bard ───────────────────────────────────────────────
	{
		ID: "mana_siphon", Name: "Sifão de Mana", Emoji: "🔮",
		Class: "mage", UnlockLevel: 25,
		Description: "Ataques mágicos têm 15% de chance de recuperar 10 MP.",
		Bonus: TalentBonus{},
	},
	{
		ID: "divine_herald", Name: "Arauto Divino", Emoji: "✝️",
		Class: "cleric", UnlockLevel: 25,
		Description: "Curas curam +25% de HP.",
		Bonus: TalentBonus{},
	},

	// ── Rogue / Archer ─────────────────────────────────────────────────────
	{
		ID: "shadow_master", Name: "Mestre das Sombras", Emoji: "👤",
		Class: "rogue", UnlockLevel: 25,
		Description: "+10% crit chance; críticos aplicam Blind por 1 turno.",
		Bonus: TalentBonus{CritChance: 10},
	},
	{
		ID: "hunters_focus", Name: "Foco do Caçador", Emoji: "🏹",
		Class: "archer", UnlockLevel: 25,
		Description: "Primeiro ataque de cada combate é sempre crítico.",
		Bonus: TalentBonus{},
	},
}

// TalentsForLevel returns all talents available at a given level (universal + class specific).
func TalentsForLevel(level int, classID string) []Talent {
	var out []Talent
	for _, t := range Talents {
		if t.UnlockLevel != level {
			continue
		}
		if t.Class == "" || t.Class == classID {
			out = append(out, t)
		}
	}
	return out
}

// GetTalent returns a talent by ID.
func GetTalent(id string) (Talent, bool) {
	for _, t := range Talents {
		if t.ID == id {
			return t, true
		}
	}
	return Talent{}, false
}
