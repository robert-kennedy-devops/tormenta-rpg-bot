package services

import (
	"github.com/tormenta-bot/internal/game"
	"github.com/tormenta-bot/internal/models"
)

type ShopService struct{}

func NewShopService() *ShopService { return &ShopService{} }

func (s *ShopService) CanBuyWithGold(char *models.Character, item models.Item, qty int) bool {
	if char == nil || qty < 1 || item.Price <= 0 {
		return false
	}
	return char.Gold >= item.Price*qty
}

func (s *ShopService) CanBuyWithDiamonds(char *models.Character, diamondCost int) bool {
	if char == nil || diamondCost <= 0 {
		return false
	}
	return char.Diamonds >= diamondCost
}

func (s *ShopService) AvailableItemsForTab(tab string, level int) []models.Item {
	var out []models.Item
	for _, item := range game.Items {
		if item.Type != tab || item.MinLevel > level {
			continue
		}
		out = append(out, item)
	}
	return out
}
