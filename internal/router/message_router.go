package router

import "sync"

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

type MessageHandler func(*tgbotapi.Message)

type MessageRouter struct {
	mu       sync.RWMutex
	commands map[string]MessageHandler
	fallback MessageHandler
}

func NewMessageRouter() *MessageRouter {
	return &MessageRouter{
		commands: make(map[string]MessageHandler),
	}
}

func (r *MessageRouter) RegisterCommand(cmd string, h MessageHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.commands[cmd] = h
}

func (r *MessageRouter) RegisterFallback(h MessageHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.fallback = h
}

func (r *MessageRouter) Dispatch(msg *tgbotapi.Message) bool {
	if msg == nil {
		return false
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	if msg.IsCommand() {
		if h, ok := r.commands[msg.Command()]; ok && h != nil {
			h(msg)
			return true
		}
	}
	if r.fallback != nil {
		r.fallback(msg)
		return true
	}
	return false
}
