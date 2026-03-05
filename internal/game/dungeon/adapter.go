package dungeon

import core "github.com/tormenta-bot/internal/dungeon"

type RoomType = core.RoomType
type Room = core.Room
type Dungeon = core.Dungeon

const (
	RoomMonster  = core.RoomMonster
	RoomTreasure = core.RoomTreasure
	RoomTrap     = core.RoomTrap
	RoomElite    = core.RoomElite
	RoomBoss     = core.RoomBoss
)

func Generate(difficulty int, seed int64) Dungeon { return core.Generate(difficulty, seed) }
func LootMultiplier(difficulty int, roomType RoomType) float64 {
	return core.LootMultiplier(difficulty, roomType)
}
