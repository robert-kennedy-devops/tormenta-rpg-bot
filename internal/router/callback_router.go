package router

import (
	"strings"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type CallbackHandler func(*tgbotapi.CallbackQuery)

type CallbackRouter struct {
	mu       sync.RWMutex
	exact    map[string]CallbackHandler
	prefix   []prefixRoute
	fallback CallbackHandler
}

type prefixRoute struct {
	prefix  string
	handler CallbackHandler
}

func NewCallbackRouter() *CallbackRouter {
	return &CallbackRouter{
		exact:  make(map[string]CallbackHandler),
		prefix: make([]prefixRoute, 0),
	}
}

func (r *CallbackRouter) RegisterExact(key string, h CallbackHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.exact[key] = h
}

func (r *CallbackRouter) RegisterPrefix(prefix string, h CallbackHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.prefix = append(r.prefix, prefixRoute{prefix: prefix, handler: h})
}

func (r *CallbackRouter) RegisterFallback(h CallbackHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.fallback = h
}

func (r *CallbackRouter) Dispatch(cb *tgbotapi.CallbackQuery) bool {
	if cb == nil {
		return false
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	if h, ok := r.exact[cb.Data]; ok && h != nil {
		h(cb)
		return true
	}
	for _, pr := range r.prefix {
		if strings.HasPrefix(cb.Data, pr.prefix) && pr.handler != nil {
			pr.handler(cb)
			return true
		}
	}
	if r.fallback != nil {
		r.fallback(cb)
		return true
	}
	return false
}
