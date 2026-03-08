package rpgdata

import (
	"fmt"

	"github.com/tormenta-bot/internal/models"
)

// ─── Monster archetype ────────────────────────────────────────────────────────

// MonsterArchetype is the blueprint for one family of monsters.
// For each level band the generator spawns one concrete Monster entry.
type MonsterArchetype struct {
	idBase      string
	name        string
	emoji       string
	description string
	weakness    string
	// Base stats at the archetype's native level
	nativeLevel int
	baseHP      int
	baseCA      int
	baseAtk     int
	baseDef     int
	baseMagAtk  int
	baseMagDef  int
	baseSpeed   int
	baseExp     int
	baseGold    int
	poisonChance int
	poisonDmg   int
	poisonTurns int
	// levelBands is the set of target levels for which variants are generated.
	// nil = use allBandLevels
	levelBands []int
}

// allBandLevels are the representative levels for each of the 5 bands.
var allBandLevels = []int{5, 10, 20, 30, 40, 50, 60, 70, 80, 90, 100}

// ─── Monster archetypes ───────────────────────────────────────────────────────

// monsterArchetypes defines every unique monster family.
// 13 archetypes × ~11 level variants ≈ 143 generated monsters.
var monsterArchetypes = []MonsterArchetype{
	// ── Tier 1 (lv 1–20) ──────────────────────────────────────────────────
	{
		idBase: "goblin", name: "Goblin", emoji: "👺",
		description: "Criatura verde e traiçoeira que espreita nas sombras.",
		weakness: "fire",
		nativeLevel: 5,
		baseHP: 30, baseCA: 10, baseAtk: 4, baseDef: 2,
		baseMagAtk: 0, baseMagDef: 1, baseSpeed: 6,
		baseExp: 40, baseGold: 10,
		levelBands: []int{2, 5, 8, 12, 16, 20},
	},
	{
		idBase: "slime", name: "Slime", emoji: "🟢",
		description: "Massa viscosa que absorve tudo ao seu redor.",
		weakness: "fire",
		nativeLevel: 3,
		baseHP: 25, baseCA: 8, baseAtk: 2, baseDef: 3,
		baseMagAtk: 0, baseMagDef: 2, baseSpeed: 2,
		baseExp: 25, baseGold: 5,
		levelBands: []int{1, 3, 6, 10, 15},
	},
	{
		idBase: "wolf", name: "Lobo", emoji: "🐺",
		description: "Predador veloz das florestas sombrias.",
		weakness: "fire",
		nativeLevel: 8,
		baseHP: 45, baseCA: 12, baseAtk: 7, baseDef: 3,
		baseMagAtk: 0, baseMagDef: 1, baseSpeed: 9,
		baseExp: 60, baseGold: 8,
		levelBands: []int{5, 10, 15, 20},
	},

	// ── Tier 2 (lv 21–40) ─────────────────────────────────────────────────
	{
		idBase: "orc", name: "Orc Guerreiro", emoji: "👹",
		description: "Guerreiro brutal de pele verde-escura.",
		weakness: "holy",
		nativeLevel: 25,
		baseHP: 120, baseCA: 14, baseAtk: 16, baseDef: 8,
		baseMagAtk: 0, baseMagDef: 3, baseSpeed: 5,
		baseExp: 180, baseGold: 35,
		levelBands: []int{22, 25, 30, 35, 40},
	},
	{
		idBase: "skeleton", name: "Esqueleto", emoji: "💀",
		description: "Morto-vivo que guarda antigas ruínas.",
		weakness: "holy",
		nativeLevel: 22,
		baseHP: 80, baseCA: 11, baseAtk: 12, baseDef: 5,
		baseMagAtk: 5, baseMagDef: 8, baseSpeed: 4,
		baseExp: 140, baseGold: 28,
		levelBands: []int{21, 25, 30, 38},
	},
	{
		idBase: "harpy", name: "Harpia", emoji: "🦅",
		description: "Criatura alada com grito ensurdecedor.",
		weakness: "ice",
		nativeLevel: 30,
		baseHP: 100, baseCA: 13, baseAtk: 18, baseDef: 5,
		baseMagAtk: 8, baseMagDef: 5, baseSpeed: 11,
		baseExp: 220, baseGold: 42,
		levelBands: []int{25, 30, 35, 40},
	},

	// ── Tier 3 (lv 41–60) ─────────────────────────────────────────────────
	{
		idBase: "troll", name: "Troll das Cavernas", emoji: "🧌",
		description: "Gigante regenerador das profundezas.",
		weakness: "fire",
		nativeLevel: 45,
		baseHP: 280, baseCA: 16, baseAtk: 28, baseDef: 15,
		baseMagAtk: 0, baseMagDef: 6, baseSpeed: 4,
		baseExp: 600, baseGold: 90,
		levelBands: []int{42, 45, 50, 55, 60},
	},
	{
		idBase: "vampire", name: "Vampiro", emoji: "🧛",
		description: "Nobre das trevas que drena a força vital.",
		weakness: "holy",
		nativeLevel: 50,
		baseHP: 220, baseCA: 15, baseAtk: 32, baseDef: 12,
		baseMagAtk: 30, baseMagDef: 20, baseSpeed: 8,
		baseExp: 750, baseGold: 120,
		levelBands: []int{45, 50, 55, 60},
	},
	{
		idBase: "golem", name: "Golem de Pedra", emoji: "🗿",
		description: "Construto mágico de pedra com força descomunal.",
		weakness: "lightning",
		nativeLevel: 48,
		baseHP: 400, baseCA: 18, baseAtk: 22, baseDef: 25,
		baseMagAtk: 5, baseMagDef: 8, baseSpeed: 2,
		baseExp: 680, baseGold: 110,
		levelBands: []int{42, 48, 55, 60},
	},

	// ── Tier 4 (lv 61–80) ─────────────────────────────────────────────────
	{
		idBase: "wyvern", name: "Viverna", emoji: "🐉",
		description: "Dragão menor de veneno mortal.",
		weakness: "ice",
		nativeLevel: 65,
		baseHP: 600, baseCA: 18, baseAtk: 45, baseDef: 22,
		baseMagAtk: 20, baseMagDef: 15, baseSpeed: 10,
		baseExp: 1800, baseGold: 250,
		poisonChance: 30, poisonDmg: 15, poisonTurns: 3,
		levelBands: []int{62, 65, 70, 75, 80},
	},
	{
		idBase: "lich", name: "Lich", emoji: "🧟",
		description: "Mago supremo que transcendeu a morte.",
		weakness: "holy",
		nativeLevel: 70,
		baseHP: 450, baseCA: 17, baseAtk: 35, baseDef: 18,
		baseMagAtk: 60, baseMagDef: 40, baseSpeed: 5,
		baseExp: 2200, baseGold: 320,
		levelBands: []int{65, 70, 75, 80},
	},

	// ── Tier 5 (lv 81–100) ────────────────────────────────────────────────
	{
		idBase: "ancient_dragon", name: "Dragão Ancião", emoji: "🐲",
		description: "Um dos mais antigos seres de Arton, poder primordial.",
		weakness: "holy",
		nativeLevel: 90,
		baseHP: 1500, baseCA: 22, baseAtk: 80, baseDef: 45,
		baseMagAtk: 70, baseMagDef: 50, baseSpeed: 12,
		baseExp: 8000, baseGold: 800,
		levelBands: []int{82, 87, 92, 97, 100},
	},
	{
		idBase: "void_titan", name: "Titã do Vazio", emoji: "🌑",
		description: "Entidade dimensional que corrói a realidade.",
		weakness: "fire",
		nativeLevel: 95,
		baseHP: 2000, baseCA: 24, baseAtk: 90, baseDef: 55,
		baseMagAtk: 85, baseMagDef: 60, baseSpeed: 8,
		baseExp: 12000, baseGold: 1200,
		levelBands: []int{85, 90, 95, 100},
	},
}

// ─── Generator ────────────────────────────────────────────────────────────────

// AllMonsters holds all generated monsters.  Populated by init().
var AllMonsters map[string]models.Monster

func init() {
	AllMonsters = make(map[string]models.Monster, 160)
	for _, arch := range monsterArchetypes {
		bands := arch.levelBands
		if bands == nil {
			bands = allBandLevels
		}
		for _, lvl := range bands {
			m := buildMonster(arch, lvl)
			AllMonsters[m.ID] = m
		}
	}
}

func buildMonster(a MonsterArchetype, targetLevel int) models.Monster {
	id := fmt.Sprintf("%s_lv%d", a.idBase, targetLevel)

	scale := func(base int) int {
		return ScaleMonsterStat(base, a.nativeLevel, targetLevel)
	}

	// Exp and gold scale faster (quadratic-ish via 1.5 power)
	return models.Monster{
		ID:          id,
		Name:        fmt.Sprintf("%s (Lv.%d)", a.name, targetLevel),
		Emoji:       a.emoji,
		Description: a.description,
		Level:       targetLevel,
		HP:          scale(a.baseHP),
		CA:          a.baseCA + (targetLevel-a.nativeLevel)/10,
		Attack:      scale(a.baseAtk),
		Defense:     scale(a.baseDef),
		MagicAtk:    scale(a.baseMagAtk),
		MagicDef:    scale(a.baseMagDef),
		Speed:       a.baseSpeed,
		ExpReward:   scale(a.baseExp),
		GoldReward:  scale(a.baseGold),
		PoisonChance: a.poisonChance,
		PoisonDmg:   scale(a.poisonDmg),
		PoisonTurns: a.poisonTurns,
		Weakness:    a.weakness,
		DropTable:   defaultDropTable(targetLevel),
	}
}

// defaultDropTable returns a generic drop table keyed by item IDs with weights.
// Weights are approximate; actual resolution is done by the drop system.
func defaultDropTable(level int) map[string]int {
	tier := TierFor(level).Number
	// Add common and uncommon loot for the appropriate tier.
	t := map[string]int{
		fmt.Sprintf("sword_t%d_r0", tier):        8,
		fmt.Sprintf("leather_t%d_r0", tier):      6,
		fmt.Sprintf("ring_atk_t%d_r0", tier):     4,
		fmt.Sprintf("ring_atk_t%d_r1", tier):     2,
		fmt.Sprintf("necklace_hp_t%d_r0", tier):  3,
	}
	// Add rare loot for tier ≥ 3
	if tier >= 3 {
		t[fmt.Sprintf("sword_t%d_r2", tier)] = 1
		t[fmt.Sprintf("plate_t%d_r2", tier)] = 1
	}
	return t
}

// ─── Lookup helpers ───────────────────────────────────────────────────────────

// MonstersForLevel returns all monsters whose Level equals targetLevel.
func MonstersForLevel(targetLevel int) []models.Monster {
	out := make([]models.Monster, 0, 4)
	for _, m := range AllMonsters {
		if m.Level == targetLevel {
			out = append(out, m)
		}
	}
	return out
}

// MonstersForTier returns all monsters whose level falls within the given tier.
func MonstersForTier(tier int) []models.Monster {
	if tier < 1 || tier > 5 {
		return nil
	}
	t := ItemTiers[tier-1]
	out := make([]models.Monster, 0, 20)
	for _, m := range AllMonsters {
		if m.Level >= t.MinLevel && m.Level <= t.MaxLevel {
			out = append(out, m)
		}
	}
	return out
}

// MonsterCount returns the total number of generated monsters.
func MonsterCount() int { return len(AllMonsters) }
