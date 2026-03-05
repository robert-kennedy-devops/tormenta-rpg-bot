package services

import (
	"time"

	"github.com/tormenta-bot/internal/drops"
	"github.com/tormenta-bot/internal/models"
	"github.com/tormenta-bot/internal/systems/events"
)

type DropService struct {
	registry *drops.Registry
}

func NewDropService() *DropService {
	r := drops.NewRegistry()
	for k, v := range drops.DefaultMaterialTables {
		r.Register(k, v)
	}
	return &DropService{registry: r}
}

func (s *DropService) RollMaterialDrops(monster *models.Monster, mode drops.Mode) []drops.Drop {
	if monster == nil {
		return nil
	}
	tableKey := s.tableKey(monster, mode)
	table, ok := s.registry.Get(tableKey)
	if !ok {
		return nil
	}
	ev := events.Global.Active(time.Time{})
	m := events.DropRateMultiplier(ev)
	if m != 1.0 {
		scaled := table
		scaled.Entries = make([]drops.Entry, 0, len(table.Entries))
		for _, e := range table.Entries {
			e.BaseChance *= m
			scaled.Entries = append(scaled.Entries, e)
		}
		table = scaled
	}
	return drops.Roll(table, mode, nil)
}

func (s *DropService) tableKey(monster *models.Monster, mode drops.Mode) string {
	if monster.Level >= 16 {
		return "boss_generic"
	}
	switch mode {
	case drops.ModeDungeon:
		return "dungeon_generic"
	case drops.ModeExplore, drops.ModeNormal, drops.ModeAutoHunt:
		return "explore_generic"
	default:
		return "explore_generic"
	}
}
