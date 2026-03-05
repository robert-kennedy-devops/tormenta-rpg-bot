package explore

import (
	"math/rand"
	"testing"
)

func TestRollEventReturnsValidType(t *testing.T) {
	d := DefaultDistribution()
	rng := rand.New(rand.NewSource(7))
	ev := RollEvent(d, rng)
	switch ev {
	case EventMonster, EventTreasure, EventRare, EventNothing:
	default:
		t.Fatalf("invalid event type: %s", ev)
	}
}
