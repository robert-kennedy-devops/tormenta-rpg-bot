package drops

import (
	"math/rand"
)

type Mode string

const (
	ModeNormal   Mode = "normal"
	ModeDungeon  Mode = "dungeon"
	ModeExplore  Mode = "explore"
	ModeAutoHunt Mode = "auto_hunt"
)

var modeMultiplier = map[Mode]float64{
	ModeNormal:   1.0,
	ModeDungeon:  1.0,
	ModeExplore:  1.0,
	ModeAutoHunt: 0.30, // ex.: 5% -> 1.5%
}

type Entry struct {
	ItemID     string
	BaseChance float64 // 0.0 to 1.0
	MinQty     int
	MaxQty     int
}

type LootTable struct {
	Entries []Entry
}

type Drop struct {
	ItemID string
	Qty    int
}

func ModeMultiplier(mode Mode) float64 {
	m, ok := modeMultiplier[mode]
	if !ok {
		return 1.0
	}
	return m
}

func SetModeMultiplier(mode Mode, mult float64) {
	if mult < 0 {
		mult = 0
	}
	modeMultiplier[mode] = mult
}

func Roll(table LootTable, mode Mode, rng *rand.Rand) []Drop {
	if rng == nil {
		rng = rand.New(rand.NewSource(rand.Int63()))
	}
	mult := ModeMultiplier(mode)
	if mult <= 0 {
		return nil
	}

	var out []Drop
	for _, e := range table.Entries {
		if e.ItemID == "" || e.BaseChance <= 0 {
			continue
		}
		chance := e.BaseChance * mult
		if chance > 1 {
			chance = 1
		}
		if rng.Float64() >= chance {
			continue
		}
		minQty, maxQty := e.MinQty, e.MaxQty
		if minQty < 1 {
			minQty = 1
		}
		if maxQty < minQty {
			maxQty = minQty
		}
		qty := minQty
		if maxQty > minQty {
			qty = minQty + rng.Intn((maxQty-minQty)+1)
		}
		out = append(out, Drop{ItemID: e.ItemID, Qty: qty})
	}
	return out
}
