package explore

import "math/rand"

type EventType string

const (
	EventMonster  EventType = "monster"
	EventTreasure EventType = "treasure"
	EventRare     EventType = "rare"
	EventNothing  EventType = "nothing"
)

type Distribution struct {
	Monster  float64
	Treasure float64
	Rare     float64
	Nothing  float64
}

func DefaultDistribution() Distribution {
	return Distribution{
		Monster:  0.55,
		Treasure: 0.20,
		Rare:     0.05,
		Nothing:  0.20,
	}
}

func RollEvent(d Distribution, rng *rand.Rand) EventType {
	if rng == nil {
		rng = rand.New(rand.NewSource(rand.Int63()))
	}
	r := rng.Float64()
	if r < d.Monster {
		return EventMonster
	}
	r -= d.Monster
	if r < d.Treasure {
		return EventTreasure
	}
	r -= d.Treasure
	if r < d.Rare {
		return EventRare
	}
	return EventNothing
}
