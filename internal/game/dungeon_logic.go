package game

import (
	"math/rand"

	"github.com/tormenta-bot/internal/models"
)

// ── DUNGEON TYPES ─────────────────────────────────────────

type DungeonFloor struct {
	Floor       int
	Description string
	MonsterIDs  []string
	IsBoss      bool
	BonusGold   int
	BonusXP     int
}

type Dungeon struct {
	ID              string
	Name            string
	Emoji           string
	Description     string
	MinLevel        int
	MaxLevel        int
	Floors          int
	Difficulty      string // easy | normal | hard | legendary
	FloorData       []DungeonFloor
	EnergyCostEntry int
	RewardGold      int
	RewardDiamonds  int
	RewardItem      string
}

// ── DUNGEON DEFINITIONS ───────────────────────────────────

var Dungeons = map[string]Dungeon{
	"crypt_novice": {
		ID: "crypt_novice", Name: "Cripta dos Novatos", Emoji: "\U0001f3da\ufe0f",
		Description:     "Uma cripta antiga nos arredores da vila. Ideal para iniciantes.",
		MinLevel: 1, MaxLevel: 5, Floors: 5, Difficulty: "easy", EnergyCostEntry: 10,
		RewardGold: 100, RewardDiamonds: 2, RewardItem: "potion_large",
		FloorData: []DungeonFloor{
			{Floor: 1, Description: "Sal\u00e3o de Entrada",         MonsterIDs: []string{"rat","goblin","mushroom","crow","slime"},       BonusGold: 10,  BonusXP: 20},
			{Floor: 2, Description: "C\u00e2mara dos Mortos",        MonsterIDs: []string{"goblin","slime","mushroom","crow","rat"},        BonusGold: 15,  BonusXP: 30},
			{Floor: 3, Description: "Passagem Sombria",               MonsterIDs: []string{"slime","wolf","crow","mushroom","goblin"},       BonusGold: 20,  BonusXP: 40},
			{Floor: 4, Description: "Sala do Tesouro",                MonsterIDs: []string{"wolf","orc","harpy","slime","crow"},             BonusGold: 30,  BonusXP: 60},
			{Floor: 5, Description: "\U0001f525 C\u00e2mara do Chefe", MonsterIDs: []string{"troll","werewolf","bandit_leader","orc","harpy"}, BonusGold: 60, BonusXP: 120, IsBoss: true},
		},
	},
	"dark_keep": {
		ID: "dark_keep", Name: "Fortaleza Sombria", Emoji: "\U0001f3f0",
		Description:     "Fortaleza tomada por for\u00e7as das trevas. Perigo elevado.",
		MinLevel: 6, MaxLevel: 12, Floors: 6, Difficulty: "normal", EnergyCostEntry: 15,
		RewardGold: 300, RewardDiamonds: 5, RewardItem: "elixir",
		FloorData: []DungeonFloor{
			{Floor: 1, Description: "Muralha Externa",        MonsterIDs: []string{"orc","troll","harpy","werewolf","bandit_leader"},            BonusGold: 30,  BonusXP: 60},
			{Floor: 2, Description: "P\u00e1tio Interno",    MonsterIDs: []string{"troll","bandit_leader","werewolf","harpy","orc"},            BonusGold: 45,  BonusXP: 90},
			{Floor: 3, Description: "Torre do Vigia",         MonsterIDs: []string{"bandit_leader","bat","werewolf","spider","troll"},           BonusGold: 60,  BonusXP: 120},
			{Floor: 4, Description: "Masmorra Profunda",      MonsterIDs: []string{"bat","spider","stone_golem_shard","crystal_wraith","golem"}, BonusGold: 80,  BonusXP: 160},
			{Floor: 5, Description: "C\u00e2mara de Torturas", MonsterIDs: []string{"spider","golem","crystal_wraith","stone_golem_shard","bat"}, BonusGold: 100, BonusXP: 200},
			{Floor: 6, Description: "\U0001f525 Trono das Trevas", MonsterIDs: []string{"undead_knight","crystal_wraith","golem","shadow_assassin","stone_golem_shard"}, BonusGold: 180, BonusXP: 360, IsBoss: true},
		},
	},
	"abyss_dungeon": {
		ID: "abyss_dungeon", Name: "Abismo Eterno", Emoji: "\U0001f311",
		Description:     "Portal para o abismo. Criaturas demon\u00edacas infestam cada canto.",
		MinLevel: 12, MaxLevel: 17, Floors: 7, Difficulty: "hard", EnergyCostEntry: 20,
		RewardGold: 600, RewardDiamonds: 10, RewardItem: "energy_elixir",
		FloorData: []DungeonFloor{
			{Floor: 1, Description: "Portal do Abismo",        MonsterIDs: []string{"undead_knight","demon","shadow_assassin","crystal_wraith","lich"},    BonusGold: 60,  BonusXP: 120},
			{Floor: 2, Description: "Plan\u00edcie Infernal",  MonsterIDs: []string{"demon","shadow_assassin","lich","necromancer","undead_knight"},       BonusGold: 90,  BonusXP: 180},
			{Floor: 3, Description: "Floresta de Almas",        MonsterIDs: []string{"demon","necromancer","lich","shadow_assassin","vampire_lord"},        BonusGold: 110, BonusXP: 220},
			{Floor: 4, Description: "Castelo da Morte",         MonsterIDs: []string{"necromancer","lich","shadow_assassin","demon","vampire_lord"},        BonusGold: 130, BonusXP: 260},
			{Floor: 5, Description: "C\u00e2mara do Sangue",   MonsterIDs: []string{"necromancer","vampire_lord","lich","shadow_assassin","demon"},        BonusGold: 160, BonusXP: 320},
			{Floor: 6, Description: "Trono do Abismo",          MonsterIDs: []string{"vampire_lord","lich","shadow_assassin","demon","necromancer"},        BonusGold: 200, BonusXP: 400},
			{Floor: 7, Description: "\U0001f525 N\u00facleo Demon\u00edaco", MonsterIDs: []string{"vampire_lord","lich","shadow_assassin","demon","necromancer"}, BonusGold: 400, BonusXP: 800, IsBoss: true},
		},
	},
	"dragon_lair": {
		ID: "dragon_lair", Name: "Covil dos Drag\u00f5es", Emoji: "\U0001f432",
		Description:     "A morada dos drag\u00f5es mais antigos. Apenas os mais poderosos sobrevivem.",
		MinLevel: 17, MaxLevel: 20, Floors: 8, Difficulty: "legendary", EnergyCostEntry: 30,
		RewardGold: 1500, RewardDiamonds: 25, RewardItem: "dragon_armor",
		FloorData: []DungeonFloor{
			{Floor: 1, Description: "Entrada do Covil",             MonsterIDs: []string{"vampire_lord","wyvern","phoenix","lich","shadow_assassin"},        BonusGold: 100, BonusXP: 200},
			{Floor: 2, Description: "Passagem de Fogo",             MonsterIDs: []string{"wyvern","dragon_young","phoenix","vampire_lord","lich"},            BonusGold: 150, BonusXP: 300},
			{Floor: 3, Description: "Sala dos Ovos",                MonsterIDs: []string{"dragon_young","wyvern","phoenix","vampire_lord","lich"},            BonusGold: 200, BonusXP: 400},
			{Floor: 4, Description: "C\u00e2mara do Drag\u00e3o", MonsterIDs: []string{"dragon_young","phoenix","wyvern","dragon_elder","vampire_lord"},    BonusGold: 250, BonusXP: 500},
			{Floor: 5, Description: "Tesouro Ancestral",            MonsterIDs: []string{"dragon_young","dragon_elder","wyvern","phoenix","vampire_lord"},    BonusGold: 300, BonusXP: 600},
			{Floor: 6, Description: "Sal\u00e3o Eterno",           MonsterIDs: []string{"dragon_elder","phoenix","wyvern","dragon_young","lich"},            BonusGold: 400, BonusXP: 800},
			{Floor: 7, Description: "Trono do Anci\u00e3o",        MonsterIDs: []string{"dragon_elder","phoenix","wyvern","dragon_young","lich"},            BonusGold: 500, BonusXP: 1000},
			{Floor: 8, Description: "\U0001f525 O Drag\u00e3o Primordial", MonsterIDs: []string{"dragon_elder","phoenix","wyvern","dragon_young","lich"},   BonusGold: 800, BonusXP: 2000, IsBoss: true},
		},
	},
}

// ── DUNGEON HELPERS ───────────────────────────────────────

// GetDungeonFloor returns floor data (1-indexed). Returns nil if out of range.
func GetDungeonFloor(dungeonID string, floor int) *DungeonFloor {
	d, ok := Dungeons[dungeonID]
	if !ok || floor < 1 || floor > len(d.FloorData) {
		return nil
	}
	f := d.FloorData[floor-1]
	return &f
}

// RollDungeonMonster picks a random monster for the given floor.
// Boss floors get boosted stats (+50% HP, +30% ATK, 2x XP/gold rewards).
func RollDungeonMonster(dungeonID string, floor int) *models.Monster {
	f := GetDungeonFloor(dungeonID, floor)
	if f == nil || len(f.MonsterIDs) == 0 {
		return nil
	}
	id := f.MonsterIDs[rand.Intn(len(f.MonsterIDs))]
	m, ok := Monsters[id]
	if !ok {
		return nil
	}
	if f.IsBoss {
		boosted := m
		boosted.HP            = int(float64(m.HP) * 1.5)
		boosted.Attack        = int(float64(m.Attack) * 1.3)
		boosted.Defense       = int(float64(m.Defense) * 1.2)
		boosted.Name          = "\U0001f451 " + m.Name + " (Chefe)"
		boosted.ExpReward     = int(float64(m.ExpReward) * 2)
		boosted.GoldReward    = int(float64(m.GoldReward) * 2)
		boosted.DiamondChance = m.DiamondChance + 30
		return &boosted
	}
	return &m
}

// GetAvailableDungeons returns all dungeons accessible at the given character level.
func GetAvailableDungeons(level int) []Dungeon {
	order := []string{"crypt_novice", "dark_keep", "abyss_dungeon", "dragon_lair"}
	var result []Dungeon
	for _, id := range order {
		d := Dungeons[id]
		if level >= d.MinLevel {
			result = append(result, d)
		}
	}
	return result
}

// DifficultyEmoji returns a colored circle emoji for the given difficulty.
func DifficultyEmoji(diff string) string {
	switch diff {
	case "easy":      return "\U0001f7e2"
	case "normal":    return "\U0001f7e1"
	case "hard":      return "\U0001f534"
	case "legendary": return "\U0001f7e3"
	}
	return "\u26aa"
}

// DungeonCompleteRewards returns the gold, diamonds, and bonus item for finishing
// a dungeon. Partial completions receive a proportional gold/diamond reward.
func DungeonCompleteRewards(dungeonID string, floorsCleared int, _ interface{}) (gold, diamonds int, item string) {
	d, ok := Dungeons[dungeonID]
	if !ok {
		return
	}
	if floorsCleared >= d.Floors {
		return d.RewardGold, d.RewardDiamonds, d.RewardItem
	}
	if floorsCleared <= 0 {
		return
	}
	pct := float64(floorsCleared) / float64(d.Floors)
	gold = int(float64(d.RewardGold) * pct)
	diamonds = int(float64(d.RewardDiamonds) * pct)
	return gold, diamonds, ""
}
