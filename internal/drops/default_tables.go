package drops

import "github.com/tormenta-bot/internal/items"

// DefaultMaterialTables is a starter table set for material progression.
// Chances are base rates for normal/dungeon/explore; auto-hunt uses mode multiplier.
var DefaultMaterialTables = map[string]LootTable{
	// Example: dungeon mobs can drop Pedra de Forja at 5% base.
	"dungeon_generic": {
		Entries: []Entry{
			{ItemID: items.MaterialIronOre, BaseChance: 0.08, MinQty: 1, MaxQty: 1},
			{ItemID: items.MaterialSilverOre, BaseChance: 0.08, MinQty: 1, MaxQty: 1},
			{ItemID: items.MaterialGoldOre, BaseChance: 0.08, MinQty: 1, MaxQty: 1},
			{ItemID: items.MaterialMagicEssence, BaseChance: 0.08, MinQty: 1, MaxQty: 1},
			{ItemID: items.MaterialForgeStone, BaseChance: 0.05, MinQty: 1, MaxQty: 1},
			{ItemID: items.MaterialMonsterShard, BaseChance: 0.12, MinQty: 1, MaxQty: 2},
		},
	},
	"explore_generic": {
		Entries: []Entry{
			{ItemID: items.MaterialIronOre, BaseChance: 0.04, MinQty: 1, MaxQty: 1},
			{ItemID: items.MaterialSilverOre, BaseChance: 0.04, MinQty: 1, MaxQty: 1},
			{ItemID: items.MaterialGoldOre, BaseChance: 0.04, MinQty: 1, MaxQty: 1},
			{ItemID: items.MaterialMagicEssence, BaseChance: 0.04, MinQty: 1, MaxQty: 1},
			{ItemID: items.MaterialForgeStone, BaseChance: 0.03, MinQty: 1, MaxQty: 1},
			{ItemID: items.MaterialMonsterShard, BaseChance: 0.08, MinQty: 1, MaxQty: 1},
		},
	},
	"auto_generic": {
		Entries: []Entry{
			{ItemID: items.MaterialIronOre, BaseChance: 0.0333334, MinQty: 1, MaxQty: 1},
			{ItemID: items.MaterialSilverOre, BaseChance: 0.0333334, MinQty: 1, MaxQty: 1},
			{ItemID: items.MaterialGoldOre, BaseChance: 0.0333334, MinQty: 1, MaxQty: 1},
			{ItemID: items.MaterialMagicEssence, BaseChance: 0.0333334, MinQty: 1, MaxQty: 1},
		},
	},
	"boss_generic": {
		Entries: []Entry{
			{ItemID: items.MaterialRefinedStone, BaseChance: 0.20, MinQty: 1, MaxQty: 2},
			{ItemID: items.MaterialArcaneEssence, BaseChance: 0.10, MinQty: 1, MaxQty: 1},
			{ItemID: items.MaterialBlackMetal, BaseChance: 0.05, MinQty: 1, MaxQty: 1},
		},
	},
}
