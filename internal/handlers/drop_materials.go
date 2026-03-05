package handlers

import (
	"fmt"

	"github.com/tormenta-bot/internal/database"
	"github.com/tormenta-bot/internal/drops"
	"github.com/tormenta-bot/internal/game"
	"github.com/tormenta-bot/internal/models"
	"github.com/tormenta-bot/internal/services"
)

var materialDropService = services.NewDropService()

func applyMaterialDrops(char *models.Character, monster *models.Monster, mode drops.Mode) []string {
	if char == nil || monster == nil {
		return nil
	}
	results := materialDropService.RollMaterialDrops(monster, mode)
	if len(results) == 0 {
		return nil
	}

	lines := make([]string, 0, len(results))
	for _, d := range results {
		if d.ItemID == "" || d.Qty < 1 {
			continue
		}
		_ = database.AddItem(char.ID, d.ItemID, "material", d.Qty)

		if item, ok := game.Items[d.ItemID]; ok {
			if d.Qty > 1 {
				lines = append(lines, fmt.Sprintf("🧱 %dx %s *%s*", d.Qty, item.Emoji, item.Name))
			} else {
				lines = append(lines, fmt.Sprintf("🧱 %s *%s*", item.Emoji, item.Name))
			}
		} else {
			if d.Qty > 1 {
				lines = append(lines, fmt.Sprintf("🧱 %dx *%s*", d.Qty, d.ItemID))
			} else {
				lines = append(lines, fmt.Sprintf("🧱 *%s*", d.ItemID))
			}
		}
	}
	return lines
}
