package drops

import "sync"

type Registry struct {
	mu     sync.RWMutex
	tables map[string]LootTable
}

func NewRegistry() *Registry {
	return &Registry{tables: make(map[string]LootTable)}
}

func (r *Registry) Register(monsterID string, table LootTable) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tables[monsterID] = table
}

func (r *Registry) Get(monsterID string) (LootTable, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.tables[monsterID]
	return t, ok
}
