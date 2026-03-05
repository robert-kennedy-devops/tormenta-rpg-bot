package service

import (
	"fmt"

	"github.com/tormenta-bot/internal/cache"
	"github.com/tormenta-bot/internal/database"
)

type RankingService struct {
	cache *cache.GameCache
}

func NewRankingService(c *cache.GameCache) *RankingService {
	if c == nil {
		c = cache.NewGameCache()
	}
	return &RankingService{cache: c}
}

func (s *RankingService) Top(limit int) ([]database.RankEntry, error) {
	if limit <= 0 {
		limit = 10
	}
	key := fmt.Sprintf("top:%d", limit)
	if v, ok := s.cache.GetRanking(key); ok {
		if list, ok2 := v.([]database.RankEntry); ok2 {
			return list, nil
		}
	}
	list, err := database.GetRanking(limit)
	if err != nil {
		return nil, err
	}
	s.cache.SetRanking(key, list)
	return list, nil
}
