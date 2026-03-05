package menu

import "sync"

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

type Screen struct {
	ImageKey string
	Caption  string
	Keyboard *tgbotapi.InlineKeyboardMarkup
}

type Builder func(userID int64) (Screen, error)

// Router maps logical menu names to screen builders.
type Router struct {
	mu       sync.RWMutex
	builders map[string]Builder
}

func NewRouter() *Router {
	return &Router{builders: make(map[string]Builder)}
}

func (r *Router) Register(menuID string, b Builder) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.builders[menuID] = b
}

func (r *Router) Build(menuID string, userID int64) (Screen, bool, error) {
	r.mu.RLock()
	b, ok := r.builders[menuID]
	r.mu.RUnlock()
	if !ok || b == nil {
		return Screen{}, false, nil
	}
	s, err := b(userID)
	return s, true, err
}
