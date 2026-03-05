package worker

import (
	"github.com/tormenta-bot/internal/systems/workers"
)

// Manager is a compatibility layer for projects expecting internal/worker package.
type Manager struct {
	base *workers.Manager
}

func NewManager() *Manager {
	return &Manager{base: workers.NewManager()}
}

func (m *Manager) Start(enablePixWorker bool) {
	m.base.Start(enablePixWorker)
}
