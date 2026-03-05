package dungeon

import "testing"

func TestGenerateRoomCount(t *testing.T) {
	d := Generate(3, 123)
	if len(d.Rooms) < 5 || len(d.Rooms) > 10 {
		t.Fatalf("room count out of range: %d", len(d.Rooms))
	}
}

func TestGenerateDeterministicBySeed(t *testing.T) {
	a := Generate(2, 777)
	b := Generate(2, 777)
	if len(a.Rooms) != len(b.Rooms) {
		t.Fatalf("different room count for same seed")
	}
	for i := range a.Rooms {
		if a.Rooms[i].Type != b.Rooms[i].Type {
			t.Fatalf("different room type at idx %d", i)
		}
	}
}

func TestLootMultiplier(t *testing.T) {
	if LootMultiplier(1, RoomBoss) <= LootMultiplier(1, RoomMonster) {
		t.Fatalf("boss should have higher multiplier than monster")
	}
}
