package repository

import (
	"github.com/tormenta-bot/internal/database"
	"github.com/tormenta-bot/internal/models"
)

type SQLRepository struct{}

func NewSQLRepository() *SQLRepository { return &SQLRepository{} }

func (r *SQLRepository) GetPlayer(playerID int64) (*models.Player, error) {
	return database.GetPlayer(playerID)
}

func (r *SQLRepository) UpsertPlayer(playerID int64, username string) error {
	return database.UpsertPlayer(playerID, username)
}

func (r *SQLRepository) GetCharacter(playerID int64) (*models.Character, error) {
	return database.GetCharacter(playerID)
}

func (r *SQLRepository) GetCharacterByID(charID int) (*models.Character, error) {
	return database.GetCharacterByID(charID)
}

func (r *SQLRepository) SaveCharacter(char *models.Character) error {
	return database.SaveCharacter(char)
}

func (r *SQLRepository) GetPixPayment(txid string) (*database.PixPaymentRecord, error) {
	return database.GetPixPayment(txid)
}

func (r *SQLRepository) ConfirmPixPaymentByTxID(txid string) (int, int, error) {
	return database.ConfirmPixPaymentByTxID(txid)
}

func (r *SQLRepository) GetPendingPixPayments(charID int) ([]database.PixPaymentRecord, error) {
	return database.GetPendingPixPayments(charID)
}

func (r *SQLRepository) GetAllPendingPixPayments() ([]database.PixPaymentRecord, error) {
	return database.GetAllPendingPixPayments()
}

func (r *SQLRepository) GetAutoHuntSession(charID int) (*database.AutoHuntSession, error) {
	return database.GetAutoHuntSession(charID)
}

func (r *SQLRepository) StartAutoHunt(charID int, mapID string, cfg database.AutoHuntSkillConfig) (*database.AutoHuntSession, error) {
	return database.StartAutoHunt(charID, mapID, cfg)
}

func (r *SQLRepository) StopAutoHunt(charID int, reason string) error {
	return database.StopAutoHunt(charID, reason)
}

func (r *SQLRepository) UpdateAutoHuntTick(sessionID, xpGain, goldGain int) error {
	return database.UpdateAutoHuntTick(sessionID, xpGain, goldGain)
}
