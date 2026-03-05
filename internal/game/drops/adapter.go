package drops

import core "github.com/tormenta-bot/internal/drops"

type Mode = core.Mode
type Entry = core.Entry
type LootTable = core.LootTable
type Drop = core.Drop

const (
	ModeNormal   = core.ModeNormal
	ModeDungeon  = core.ModeDungeon
	ModeExplore  = core.ModeExplore
	ModeAutoHunt = core.ModeAutoHunt
)

func ModeMultiplier(mode Mode) float64          { return core.ModeMultiplier(mode) }
func SetModeMultiplier(mode Mode, mult float64) { core.SetModeMultiplier(mode, mult) }
func Roll(table LootTable, mode Mode) []Drop    { return core.Roll(table, mode, nil) }
