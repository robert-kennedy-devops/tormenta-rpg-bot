package repository

import (
	"github.com/tormenta-bot/internal/database"
	"github.com/tormenta-bot/internal/models"
)

type PlayerRepository interface {
	GetPlayer(playerID int64) (*models.Player, error)
	UpsertPlayer(playerID int64, username string) error
}

type CharacterRepository interface {
	GetCharacter(playerID int64) (*models.Character, error)
	GetCharacterByID(charID int) (*models.Character, error)
	SaveCharacter(char *models.Character) error
}

type PixRepository interface {
	GetPixPayment(txid string) (*database.PixPaymentRecord, error)
	ConfirmPixPaymentByTxID(txid string) (int, int, error)
	GetPendingPixPayments(charID int) ([]database.PixPaymentRecord, error)
	GetAllPendingPixPayments() ([]database.PixPaymentRecord, error)
}

type AutoHuntRepository interface {
	GetAutoHuntSession(charID int) (*database.AutoHuntSession, error)
	StartAutoHunt(charID int, mapID string, cfg database.AutoHuntSkillConfig) (*database.AutoHuntSession, error)
	StopAutoHunt(charID int, reason string) error
	UpdateAutoHuntTick(sessionID, xpGain, goldGain int) error
}
