package services

import (
	"testing"

	"github.com/tormenta-bot/internal/drops"
	"github.com/tormenta-bot/internal/models"
)

func TestAutoGenericEffectiveRateNearOnePercent(t *testing.T) {
	svc := NewDropService()

	monster := &models.Monster{
		ID:    "rat",
		Name:  "Rato",
		Level: 1,
	}

	const trials = 200000
	var hits int
	for i := 0; i < trials; i++ {
		ds := svc.RollMaterialDrops(monster, drops.ModeAutoHunt)
		for _, d := range ds {
			if d.ItemID == "mat_iron_ore" {
				hits++
				break
			}
		}
	}

	rate := float64(hits) / float64(trials)
	// Expected ~1.0% (0.0333334 * 0.30 ~= 0.01000002).
	// Keep a practical tolerance for pseudo-random variance.
	if rate < 0.0085 || rate > 0.0115 {
		t.Fatalf("unexpected auto-hunt effective rate: got %.6f want near 0.010000", rate)
	}
}
