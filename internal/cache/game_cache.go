package cache

import "time"

type GameCache struct {
	ranking *TTLCache
	shop    *TTLCache
	dungeon *TTLCache
}

func NewGameCache() *GameCache {
	return &GameCache{
		ranking: NewTTLCache(),
		shop:    NewTTLCache(),
		dungeon: NewTTLCache(),
	}
}

func (c *GameCache) SetRanking(key string, value interface{}) {
	c.ranking.Set(key, value, 30*time.Second)
}
func (c *GameCache) GetRanking(key string) (interface{}, bool) { return c.ranking.Get(key) }

func (c *GameCache) SetShop(key string, value interface{})  { c.shop.Set(key, value, 30*time.Second) }
func (c *GameCache) GetShop(key string) (interface{}, bool) { return c.shop.Get(key) }

func (c *GameCache) SetDungeon(key string, value interface{}) {
	c.dungeon.Set(key, value, 60*time.Second)
}
func (c *GameCache) GetDungeon(key string) (interface{}, bool) { return c.dungeon.Get(key) }
