package dungeon

import (
	"math/rand"
	"time"
)

type RoomType string

const (
	RoomMonster  RoomType = "monster"
	RoomTreasure RoomType = "treasure"
	RoomTrap     RoomType = "trap"
	RoomElite    RoomType = "elite"
	RoomBoss     RoomType = "boss"
)

type Room struct {
	Index int
	Type  RoomType
}

type Dungeon struct {
	Seed       int64
	Difficulty int
	Rooms      []Room
}

// Generate creates a procedural dungeon with 5-10 rooms.
// Distribution target:
// monster 50%, treasure 20%, trap 15%, elite 10%, boss 5%.
func Generate(difficulty int, seed int64) Dungeon {
	if difficulty < 1 {
		difficulty = 1
	}
	if seed == 0 {
		seed = time.Now().UnixNano()
	}
	rng := rand.New(rand.NewSource(seed))
	roomCount := 5 + rng.Intn(6) // 5..10

	types := make([]RoomType, 0, roomCount)
	weights := []struct {
		t RoomType
		w int
	}{
		{RoomMonster, 50},
		{RoomTreasure, 20},
		{RoomTrap, 15},
		{RoomElite, 10},
		{RoomBoss, 5},
	}

	for len(types) < roomCount {
		roll := rng.Intn(100)
		acc := 0
		for _, e := range weights {
			acc += e.w
			if roll < acc {
				types = append(types, e.t)
				break
			}
		}
	}

	rooms := make([]Room, 0, roomCount)
	for i, t := range types {
		rooms = append(rooms, Room{Index: i + 1, Type: t})
	}

	return Dungeon{
		Seed:       seed,
		Difficulty: difficulty,
		Rooms:      rooms,
	}
}

// LootMultiplier scales rewards by difficulty and room type.
func LootMultiplier(difficulty int, roomType RoomType) float64 {
	if difficulty < 1 {
		difficulty = 1
	}
	base := 1.0 + (float64(difficulty-1) * 0.12)
	switch roomType {
	case RoomTreasure:
		return base * 1.25
	case RoomElite:
		return base * 1.5
	case RoomBoss:
		return base * 2.0
	case RoomTrap:
		return base * 0.7
	default:
		return base
	}
}
