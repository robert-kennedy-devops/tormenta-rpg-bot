package models

import "time"

// ── Player ────────────────────────────────────────────────
type Player struct {
	ID           int64      `db:"id"`
	Username     string     `db:"username"`
	CreatedAt    time.Time  `db:"created_at"`
	IsVIP        bool       `db:"is_vip"`
	VIPExpiresAt *time.Time `db:"vip_expires_at"`
}

// IsVIPActive returns true if the player currently has active VIP.
func (p *Player) IsVIPActive() bool {
	if !p.IsVIP {
		return false
	}
	if p.VIPExpiresAt == nil {
		return true // permanent VIP
	}
	return time.Now().Before(*p.VIPExpiresAt)
}

// ── Character ─────────────────────────────────────────────
type Character struct {
	ID               int       `db:"id"`
	PlayerID         int64     `db:"player_id"`
	Name             string    `db:"name"`
	Race             string    `db:"race"`
	Class            string    `db:"class"`
	Level            int       `db:"level"`
	Experience       int       `db:"experience"`
	ExperienceNext   int       `db:"experience_next"`
	HP               int       `db:"hp"`
	HPMax            int       `db:"hp_max"`
	MP               int       `db:"mp"`
	MPMax            int       `db:"mp_max"`
	Energy           int       `db:"energy"`
	EnergyMax        int       `db:"energy_max"`
	EnergyRegenAt    time.Time `db:"energy_regen_at"`
	LastEnergyUpdate int64     `db:"last_energy_update"` // unix timestamp (seconds)
	Diamonds         int       `db:"diamonds"`
	Strength         int       `db:"strength"`
	Dexterity        int       `db:"dexterity"`
	Constitution     int       `db:"constitution"`
	Intelligence     int       `db:"intelligence"`
	Wisdom           int       `db:"wisdom"`
	Charisma         int       `db:"charisma"`
	Attack           int       `db:"attack"`
	Defense          int       `db:"defense"`
	MagicAttack      int       `db:"magic_attack"`
	MagicDefense     int       `db:"magic_defense"`
	Speed            int       `db:"speed"`
	Gold             int       `db:"gold"`
	CurrentMap       string    `db:"current_map"`
	State            string    `db:"state"` // idle | combat | dungeon | pvp
	CombatMonsterID  string    `db:"combat_monster_id"`
	CombatMonsterHP  int       `db:"combat_monster_hp"`
	SkillPoints      int       `db:"skill_points"`
	Deaths           int       `db:"deaths"`
	XPBoostExpiry    time.Time `db:"xp_boost_expiry"` // ativo se After(time.Now())
	// Veneno no player (aplicado por monstros)
	PoisonTurns int `db:"poison_turns"`
	PoisonDmg   int `db:"poison_dmg"`
	// Veneno no monstro (aplicado pelo player)
	CombatMonsterPoisonTurns int `db:"combat_monster_poison_turns"`
	CombatMonsterPoisonDmg   int `db:"combat_monster_poison_dmg"`
	// Calculados em runtime — não persistidos no banco
	EquipCABonus  int `db:"-"`
	EquipHitBonus int `db:"-"`
}

// ── Inventory ─────────────────────────────────────────────
type InventoryItem struct {
	ID          int    `db:"id"`
	CharacterID int    `db:"character_id"`
	ItemID      string `db:"item_id"`
	ItemType    string `db:"item_type"`
	Quantity    int    `db:"quantity"`
	Equipped    bool   `db:"equipped"`
	Slot        string `db:"slot"` // weapon|head|chest|hands|legs|feet|offhand|accessory1|accessory2
}

type CharacterSkill struct {
	ID          int    `db:"id"`
	CharacterID int    `db:"character_id"`
	SkillID     string `db:"skill_id"`
	Level       int    `db:"level"`
}

// ── Rarity ────────────────────────────────────────────────
type Rarity int

const (
	RarityCommon Rarity = iota
	RarityUncommon
	RarityRare
	RarityEpic
	RarityLegendary
)

func (r Rarity) Emoji() string {
	switch r {
	case RarityCommon:
		return "⚪"
	case RarityUncommon:
		return "🟢"
	case RarityRare:
		return "🔵"
	case RarityEpic:
		return "🟣"
	case RarityLegendary:
		return "🟡"
	}
	return "⚪"
}

func (r Rarity) Name() string {
	switch r {
	case RarityCommon:
		return "Comum"
	case RarityUncommon:
		return "Incomum"
	case RarityRare:
		return "Raro"
	case RarityEpic:
		return "Épico"
	case RarityLegendary:
		return "Lendário"
	}
	return "Comum"
}

// ── Item ──────────────────────────────────────────────────
type Item struct {
	ID             string
	Name           string
	Description    string
	Emoji          string
	Type           string // weapon | armor | consumable | chest
	Rarity         Rarity
	Price          int // gold (0 = not in shop)
	SellPrice      int
	DiamondPrice   int
	AttackBonus    int
	MagicAtkBonus  int
	DefenseBonus   int
	MagicDefBonus  int
	SpeedBonus     int
	CABonus        int // bônus direto na Classe de Armadura (armaduras/escudos)
	HitBonus       int // bônus no teste de ataque d20 (armas)
	HealHP         int
	HealMP         int
	RestoreEnergy  int
	XPBoostMinutes int  // se >0, ativa bônus de +50% XP por N minutos
	CurePoison     bool // se true, cura envenenamento do player
	MinLevel       int
	ClassReq       string // empty = all classes
	DropWeight     int
	Slot           string // weapon|head|chest|hands|legs|feet|accessory (for equip slot logic)
	HPBonus        int    // bônus flat de HP máximo
	MPBonus        int    // bônus flat de MP máximo
}

// ── Game world ────────────────────────────────────────────
type Race struct {
	ID          string
	Name        string
	Description string
	Emoji       string
	BonusHP     int
	BonusMP     int
	BonusStr    int
	BonusDex    int
	BonusCon    int
	BonusInt    int
	BonusWis    int
	BonusCha    int
	Trait       string
}

type Class struct {
	ID           string
	Name         string
	Description  string
	Emoji        string
	BaseHP       int
	BaseMP       int
	HPPerLevel   int
	MPPerLevel   int
	BaseAttack   int
	BaseDefense  int
	PrimaryStats []string
	Role         string
}

type Skill struct {
	ID               string
	Name             string
	Description      string
	Class            string
	Branch           string // ramo de build: "berserker","protector","champion", etc.
	Tier             int    // 1-4
	PointCost        int    // pontos para aprender: T1=1, T2=1, T3=2, T4=3
	MPCost           int
	Damage           int
	DamageType       string
	RequiredLevel    int
	Requires         string // skillID pré-requisito
	Passive          bool
	Emoji            string
	PoisonDmgPerTurn int // dano de veneno por turno (DoT)
	PoisonTurnsCount int // quantos turnos o veneno dura
}

type Monster struct {
	ID            string
	Name          string
	Description   string
	Emoji         string
	Level         int
	HP            int
	CA            int // Classe de Armadura (dificuldade para acertar)
	Attack        int // bônus de ataque no d20
	Defense       int // mantido para compatibilidade de stats
	MagicAtk      int
	MagicDef      int
	Speed         int
	ExpReward     int
	GoldReward    int
	DiamondChance int
	PoisonChance  int // chance (%) de aplicar veneno ao acertar
	PoisonDmg     int // dano por turno do veneno aplicado ao player
	PoisonTurns   int // duração do veneno aplicado ao player
	MapID         string
	Weakness      string
	DropTable     map[string]int
}

type GameMap struct {
	ID          string
	Name        string
	Description string
	Emoji       string
	MinLevel    int
	MaxLevel    int
	ConnectsTo  []string
	Monsters    []string
	HasShop     bool
	HasInn      bool
}

// ── State models ──────────────────────────────────────────
type ShopQtyState struct {
	ItemID   string
	Quantity int
	PayWith  string // "gold" | "diamonds"
}

// ShopCartItem é um item no carrinho de compra.
type ShopCartItem struct {
	ItemID  string
	Qty     int
	Diamond bool // true = pagar com diamantes
}

// ShopCart é o carrinho multi-item da loja.
type ShopCart struct {
	Items   []ShopCartItem
	TabType string // "consumable" | "weapon" | "armor"
}

// SellCartItem é um item selecionado para venda.
type SellCartItem struct {
	ItemID string
	Qty    int
}

// SellCart é a seleção multi-item de venda.
type SellCart struct {
	Items []SellCartItem
}

// ── Pix / Diamond packages ────────────────────────────────
type DiamondPackage struct {
	ID       string
	Name     string
	Emoji    string
	Amount   int // diamonds granted
	Bonus    int // bonus diamonds
	PriceBRL float64
	Price    string // formatted display "R$ X,XX"
}
