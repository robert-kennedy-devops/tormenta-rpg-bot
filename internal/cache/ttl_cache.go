package cache

import (
	"sync"
	"time"
)

type entry struct {
	value     interface{}
	expiresAt time.Time
}

type TTLCache struct {
	mu    sync.RWMutex
	items map[string]entry
}

func NewTTLCache() *TTLCache {
	return &TTLCache{items: make(map[string]entry)}
}

func (c *TTLCache) Set(key string, value interface{}, ttl time.Duration) {
	if key == "" {
		return
	}
	if ttl <= 0 {
		ttl = 10 * time.Second
	}
	c.mu.Lock()
	c.items[key] = entry{value: value, expiresAt: time.Now().Add(ttl)}
	c.mu.Unlock()
}

func (c *TTLCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	e, ok := c.items[key]
	c.mu.RUnlock()
	if !ok || time.Now().After(e.expiresAt) {
		if ok {
			c.mu.Lock()
			delete(c.items, key)
			c.mu.Unlock()
		}
		return nil, false
	}
	return e.value, true
}

func (c *TTLCache) Delete(key string) {
	c.mu.Lock()
	delete(c.items, key)
	c.mu.Unlock()
}
