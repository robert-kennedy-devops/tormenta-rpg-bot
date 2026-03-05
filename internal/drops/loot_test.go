package drops

import (
	"math/rand"
	"testing"
)

func TestAutoHuntMultiplier(t *testing.T) {
	if got := ModeMultiplier(ModeAutoHunt); got != 0.30 {
		t.Fatalf("expected 0.30, got %v", got)
	}
}

func TestRollDeterministicDrop(t *testing.T) {
	table := LootTable{
		Entries: []Entry{
			{ItemID: "mat_forge_stone", BaseChance: 1.0, MinQty: 1, MaxQty: 1},
		},
	}
	rng := rand.New(rand.NewSource(42))
	got := Roll(table, ModeNormal, rng)
	if len(got) != 1 || got[0].ItemID != "mat_forge_stone" || got[0].Qty != 1 {
		t.Fatalf("unexpected drops: %#v", got)
	}
}
