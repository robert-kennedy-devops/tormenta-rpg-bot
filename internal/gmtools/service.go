package gmtools

import (
	"fmt"

	"github.com/tormenta-bot/internal/database"
	"github.com/tormenta-bot/internal/game"
)

type Service struct{}

func New() *Service { return &Service{} }

func (s *Service) Teleport(charID int, mapID string) error {
	char, err := database.GetCharacterByID(charID)
	if err != nil || char == nil {
		return fmt.Errorf("character not found")
	}
	if _, ok := game.Maps[mapID]; !ok {
		return fmt.Errorf("invalid map")
	}
	char.CurrentMap = mapID
	return database.SaveCharacter(char)
}

func (s *Service) AddEnergy(charID int, delta int) error {
	char, err := database.GetCharacterByID(charID)
	if err != nil || char == nil {
		return fmt.Errorf("character not found")
	}
	char.Energy += delta
	if char.Energy < 0 {
		char.Energy = 0
	}
	if char.Energy > char.EnergyMax {
		char.Energy = char.EnergyMax
	}
	return database.SaveCharacter(char)
}

func (s *Service) SpawnItem(charID int, itemID string, qty int) error {
	if qty <= 0 {
		qty = 1
	}
	item, ok := game.Items[itemID]
	if !ok {
		return fmt.Errorf("item not found")
	}
	return database.AddItem(charID, itemID, item.Type, qty)
}

func (s *Service) ResetDungeon(charID int) error {
	run, err := database.GetActiveDungeonRun(charID)
	if err != nil {
		return err
	}
	if run != nil {
		if err := database.FinishDungeonRun(run.ID, "reset_by_gm"); err != nil {
			return err
		}
	}
	char, err := database.GetCharacterByID(charID)
	if err == nil && char != nil {
		char.State = "idle"
		char.CombatMonsterID = ""
		char.CombatMonsterHP = 0
		_ = database.SaveCharacter(char)
	}
	return nil
}
