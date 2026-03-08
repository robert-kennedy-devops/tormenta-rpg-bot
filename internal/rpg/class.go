package rpg

// ─── Class definition ─────────────────────────────────────────────────────────

// Role describes the combat function of a class.
type Role string

const (
	RoleTank   Role = "tank"
	RoleHealer Role = "healer"
	RoleDPS    Role = "dps"
	RoleMage   Role = "mage"
	RoleRanged Role = "ranged"
	RoleSupport Role = "support"
)

// ClassDef is the extended class template.
type ClassDef struct {
	ID          string
	Name        string
	Emoji       string
	Description string
	Role        Role

	// Base stats at level 1.
	BaseHP  int
	BaseMP  int
	BaseAtk int
	BaseDef int
	BaseMagAtk int
	BaseMagDef int
	BaseSpd int

	// Per-level growth.
	HPPerLevel  int
	MPPerLevel  int
	AtkPerLevel int
	DefPerLevel int

	// Primary stats that receive bonus points every 5 levels.
	PrimaryStats []string

	// Allowed races (empty = all races allowed).
	AllowedRaces []string

	// Trait key for class-specific passive.
	TraitKey  string
	TraitDesc string
}

// ─── Class registry ───────────────────────────────────────────────────────────

// Classes is the master map of all playable classes.
var Classes = map[string]ClassDef{

	// ── Legacy classes (IDs must match game/data.go) ───────────────────────
	"warrior": {
		ID: "warrior", Name: "Guerreiro", Emoji: "⚔️", Role: RoleTank,
		Description: "Mestre das armas e armaduras pesadas. Linha de frente implacável.",
		BaseHP: 80, BaseMP: 20, BaseAtk: 12, BaseDef: 10, BaseMagAtk: 2, BaseMagDef: 4, BaseSpd: 8,
		HPPerLevel: 12, MPPerLevel: 3, AtkPerLevel: 3, DefPerLevel: 2,
		PrimaryStats: []string{"strength", "constitution"},
		TraitKey: "warrior_weapon_mastery", TraitDesc: "+5% dano com armas por 5 níveis de especialização.",
	},
	"mage": {
		ID: "mage", Name: "Arcanista", Emoji: "🧙", Role: RoleMage,
		Description: "Manipula as forças arcanas com maestria e poder devastador.",
		BaseHP: 45, BaseMP: 80, BaseAtk: 4, BaseDef: 3, BaseMagAtk: 18, BaseMagDef: 12, BaseSpd: 7,
		HPPerLevel: 5, MPPerLevel: 14, AtkPerLevel: 1, DefPerLevel: 1,
		PrimaryStats: []string{"intelligence", "wisdom"},
		TraitKey: "mage_arcane_surge", TraitDesc: "A cada 4 feitiços, o próximo não consome MP.",
	},
	"rogue": {
		ID: "rogue", Name: "Ladino", Emoji: "🗡️", Role: RoleDPS,
		Description: "Ágil e letal nas sombras. Veneno, traição e precisão cirúrgica.",
		BaseHP: 55, BaseMP: 40, BaseAtk: 10, BaseDef: 5, BaseMagAtk: 5, BaseMagDef: 5, BaseSpd: 13,
		HPPerLevel: 7, MPPerLevel: 6, AtkPerLevel: 2, DefPerLevel: 1,
		PrimaryStats: []string{"dexterity", "charisma"},
		TraitKey: "rogue_backstab", TraitDesc: "Ataques pelo flanco causam +40% de dano.",
	},
	"archer": {
		ID: "archer", Name: "Caçador", Emoji: "🏹", Role: RoleRanged,
		Description: "Precisão e velocidade a longa distância. Perito em rastreio e armadilhas.",
		BaseHP: 60, BaseMP: 35, BaseAtk: 9, BaseDef: 6, BaseMagAtk: 3, BaseMagDef: 6, BaseSpd: 11,
		HPPerLevel: 8, MPPerLevel: 5, AtkPerLevel: 2, DefPerLevel: 1,
		PrimaryStats: []string{"dexterity", "wisdom"},
		TraitKey: "archer_eagle_eye", TraitDesc: "+10% chance de crítico com ataques à distância.",
	},

	// ── New Tormenta classes ───────────────────────────────────────────────
	"paladin": {
		ID: "paladin", Name: "Paladino", Emoji: "⚜️", Role: RoleTank,
		Description: "Guerreiro sagrado — combina força marcial com magia divina de cura.",
		BaseHP: 75, BaseMP: 50, BaseAtk: 10, BaseDef: 10, BaseMagAtk: 8, BaseMagDef: 10, BaseSpd: 7,
		HPPerLevel: 10, MPPerLevel: 7, AtkPerLevel: 2, DefPerLevel: 2,
		PrimaryStats: []string{"strength", "wisdom"},
		TraitKey: "paladin_holy_aura", TraitDesc: "Aura sagrada: companheiros adjacentes recebem +10% de HP em cura.",
	},
	"cleric": {
		ID: "cleric", Name: "Clérigo", Emoji: "✝️", Role: RoleHealer,
		Description: "Canal dos deuses — curandeiro incomparável e destruidor de mortos-vivos.",
		BaseHP: 55, BaseMP: 75, BaseAtk: 5, BaseDef: 7, BaseMagAtk: 14, BaseMagDef: 14, BaseSpd: 6,
		HPPerLevel: 7, MPPerLevel: 12, AtkPerLevel: 1, DefPerLevel: 2,
		PrimaryStats: []string{"wisdom", "constitution"},
		TraitKey: "cleric_divine_grace", TraitDesc: "Curas têm 10% de chance de curar 2x; imune a Curse.",
	},
	"barbarian": {
		ID: "barbarian", Name: "Bárbaro", Emoji: "🪓", Role: RoleDPS,
		Description: "Guerreiro primitivo movido pela fúria — dano bruto e resistência sobre-humana.",
		BaseHP: 90, BaseMP: 10, BaseAtk: 14, BaseDef: 8, BaseMagAtk: 1, BaseMagDef: 3, BaseSpd: 9,
		HPPerLevel: 14, MPPerLevel: 2, AtkPerLevel: 4, DefPerLevel: 1,
		PrimaryStats: []string{"strength", "constitution"},
		TraitKey: "barbarian_rage", TraitDesc: "Ao ativar Berserk, ganha +60% ataque e -30% dano recebido por 3 turnos.",
	},
	"bard": {
		ID: "bard", Name: "Bardo", Emoji: "🎵", Role: RoleSupport,
		Description: "Artista e guerreiro — usa música e magia para buff aliados e debuff inimigos.",
		BaseHP: 58, BaseMP: 60, BaseAtk: 7, BaseDef: 6, BaseMagAtk: 9, BaseMagDef: 8, BaseSpd: 10,
		HPPerLevel: 7, MPPerLevel: 9, AtkPerLevel: 2, DefPerLevel: 1,
		PrimaryStats: []string{"charisma", "dexterity"},
		TraitKey: "bard_inspiration", TraitDesc: "Canção de Inspiração: toda a party ganha +15% XP por 10 minutos após a batalha.",
	},
}

// GetClass returns a ClassDef by ID, plus a found flag.
func GetClass(id string) (ClassDef, bool) {
	c, ok := Classes[id]
	return c, ok
}

// StatsAtLevel returns approximate base combat stats at a given level (1–100).
func (c ClassDef) StatsAtLevel(level int) (hp, mp, atk, def int) {
	if level < 1 {
		level = 1
	}
	lvl := level - 1
	hp = c.BaseHP + c.HPPerLevel*lvl
	mp = c.BaseMP + c.MPPerLevel*lvl
	atk = c.BaseAtk + c.AtkPerLevel*lvl
	def = c.BaseDef + c.DefPerLevel*lvl
	return
}
