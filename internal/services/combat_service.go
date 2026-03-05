package services

import (
	"math/rand"

	"github.com/tormenta-bot/internal/game"
	"github.com/tormenta-bot/internal/models"
)

type CombatService struct{}

func NewCombatService() *CombatService { return &CombatService{} }

func (s *CombatService) RandomMonsterForMap(mapID string) (*models.Monster, bool) {
	monsters := game.GetMonstersForMap(mapID)
	if len(monsters) == 0 {
		return nil, false
	}
	m := monsters[rand.Intn(len(monsters))]
	return &m, true
}

func (s *CombatService) Attack(char *models.Character, monster *models.Monster) game.CombatResult {
	return game.PlayerAttack(char, monster)
}

func (s *CombatService) SkillAttack(char *models.Character, skill *models.Skill, monster *models.Monster) game.CombatResult {
	return game.PlayerSkillAttack(char, skill, monster)
}
