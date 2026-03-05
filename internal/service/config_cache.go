package service

import (
	"fmt"

	"github.com/tormenta-bot/internal/cache"
	"github.com/tormenta-bot/internal/game"
	"github.com/tormenta-bot/internal/models"
)

type ConfigCacheService struct {
	cache *cache.GameCache
}

func NewConfigCacheService(c *cache.GameCache) *ConfigCacheService {
	if c == nil {
		c = cache.NewGameCache()
	}
	return &ConfigCacheService{cache: c}
}

func (s *ConfigCacheService) ShopItemsByType(itemType string, level int, classID string) []models.Item {
	key := fmt.Sprintf("shop:%s:%d:%s", itemType, level, classID)
	if v, ok := s.cache.GetShop(key); ok {
		if list, ok2 := v.([]models.Item); ok2 {
			return list
		}
	}
	var out []models.Item
	for _, it := range game.Items {
		if it.Type != itemType || it.Price <= 0 {
			continue
		}
		if it.MinLevel > level {
			continue
		}
		if it.ClassReq != "" && it.ClassReq != classID {
			continue
		}
		out = append(out, it)
	}
	s.cache.SetShop(key, out)
	return out
}

func (s *ConfigCacheService) DungeonConfig(id string) (game.Dungeon, bool) {
	key := "dungeon:" + id
	if v, ok := s.cache.GetDungeon(key); ok {
		if d, ok2 := v.(game.Dungeon); ok2 {
			return d, true
		}
	}
	d, ok := game.Dungeons[id]
	if !ok {
		return game.Dungeon{}, false
	}
	s.cache.SetDungeon(key, d)
	return d, true
}
