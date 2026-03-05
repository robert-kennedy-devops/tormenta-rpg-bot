package services

import (
	"github.com/tormenta-bot/internal/database"
	"github.com/tormenta-bot/internal/game"
	"github.com/tormenta-bot/internal/models"
)

type PlayerService struct{}

func NewPlayerService() *PlayerService { return &PlayerService{} }

func (s *PlayerService) GetCharacterByPlayerID(playerID int64) (*models.Character, error) {
	return database.GetCharacter(playerID)
}

func (s *PlayerService) TickAndSaveEnergy(char *models.Character, isVIP bool) error {
	if char == nil {
		return nil
	}
	game.TickEnergyVIP(char, isVIP)
	return database.SaveCharacter(char)
}

func (s *PlayerService) IsBanned(playerID int64) bool {
	return database.IsPlayerBanned(playerID)
}
