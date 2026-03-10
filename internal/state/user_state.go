// Package state provides a concurrency-safe per-user session store.
//
// All handler state that was previously stored in plain, un-mutexed package-level
// maps (creationState, shopCart, sellCart, navCurrent, navBack, shopQtyState)
// is managed here behind a single RWMutex, eliminating data-race conditions
// that occur when multiple Telegram updates are dispatched concurrently.
package state

import (
	"sync"

	"github.com/tormenta-bot/internal/models"
)

// ── UserSession holds all ephemeral state for one Telegram user. ──────────────

// UserSession holds every transient piece of UI/flow state for a single user.
// Fields are intentionally exported so callers can pattern-match on nil.
type UserSession struct {
	// Character-creation wizard key→value bag ("awaiting_name", "race", "class", …)
	Creation map[string]string

	// Legacy single-item buy-quantity selection (shop flow)
	ShopQty *models.ShopQtyState

	// Multi-item shop cart
	Cart *models.ShopCart

	// Multi-item sell selection
	SellCart *models.SellCart

	// Current menu screen identifier (for navigation back stack)
	NavCurrent string

	// Dynamic back-button targets keyed by the menu that set them
	NavBack map[string]string
}

// ── Manager is the global, mutex-safe per-user store. ────────────────────────

// Manager is safe for concurrent use by multiple goroutines.
type Manager struct {
	mu       sync.RWMutex
	sessions map[int64]*UserSession
}

// New returns an initialised Manager ready to use.
func New() *Manager {
	return &Manager{sessions: make(map[int64]*UserSession)}
}

// Global is the process-wide singleton used by all handler code.
var Global = New()

// ── Session helpers ───────────────────────────────────────────────────────────

// get returns the session for userID, creating it if absent.
// Caller MUST hold at least a write lock.
func (m *Manager) get(userID int64) *UserSession {
	s := m.sessions[userID]
	if s == nil {
		s = &UserSession{
			Creation: map[string]string{},
			NavBack:  map[string]string{},
		}
		m.sessions[userID] = s
	}
	return s
}

// ── Creation state ────────────────────────────────────────────────────────────

// CreationGet returns the creation-flow value for key (empty string if absent).
func (m *Manager) CreationGet(userID int64, key string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s := m.sessions[userID]
	if s == nil {
		return ""
	}
	return s.Creation[key]
}

// CreationSet sets a key in the creation flow for userID.
func (m *Manager) CreationSet(userID int64, key, value string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.get(userID).Creation[key] = value
}

// CreationGetAll returns a copy of the entire creation map for userID.
func (m *Manager) CreationGetAll(userID int64) map[string]string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s := m.sessions[userID]
	if s == nil {
		return nil
	}
	out := make(map[string]string, len(s.Creation))
	for k, v := range s.Creation {
		out[k] = v
	}
	return out
}

// CreationDelete removes a single key from the creation map.
func (m *Manager) CreationDelete(userID int64, key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	s := m.sessions[userID]
	if s != nil {
		delete(s.Creation, key)
	}
}

// CreationClear removes the entire creation state for userID (after character is made).
func (m *Manager) CreationClear(userID int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	s := m.sessions[userID]
	if s != nil {
		s.Creation = map[string]string{}
	}
}

// ── Shop qty (legacy single-item) ────────────────────────────────────────────

func (m *Manager) ShopQtyGet(userID int64) *models.ShopQtyState {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s := m.sessions[userID]
	if s == nil {
		return nil
	}
	return s.ShopQty
}

func (m *Manager) ShopQtySet(userID int64, v *models.ShopQtyState) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.get(userID).ShopQty = v
}

func (m *Manager) ShopQtyClear(userID int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	s := m.sessions[userID]
	if s != nil {
		s.ShopQty = nil
	}
}

// ── Shop cart (multi-item) ────────────────────────────────────────────────────

func (m *Manager) CartGet(userID int64) *models.ShopCart {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s := m.sessions[userID]
	if s == nil {
		return nil
	}
	return s.Cart
}

func (m *Manager) CartSet(userID int64, c *models.ShopCart) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.get(userID).Cart = c
}

func (m *Manager) CartClear(userID int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	s := m.sessions[userID]
	if s != nil {
		s.Cart = nil
	}
}

// ── Sell cart ─────────────────────────────────────────────────────────────────

func (m *Manager) SellCartGet(userID int64) *models.SellCart {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s := m.sessions[userID]
	if s == nil {
		return nil
	}
	return s.SellCart
}

func (m *Manager) SellCartSet(userID int64, c *models.SellCart) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.get(userID).SellCart = c
}

func (m *Manager) SellCartClear(userID int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	s := m.sessions[userID]
	if s != nil {
		s.SellCart = nil
	}
}

// ── Navigation ────────────────────────────────────────────────────────────────

func (m *Manager) NavCurrentGet(userID int64) string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s := m.sessions[userID]
	if s == nil {
		return ""
	}
	return s.NavCurrent
}

func (m *Manager) NavCurrentSet(userID int64, screen string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.get(userID).NavCurrent = screen
}

func (m *Manager) NavBackGet(userID int64, menu string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s := m.sessions[userID]
	if s == nil {
		return ""
	}
	return s.NavBack[menu]
}

func (m *Manager) NavBackSet(userID int64, menu, dest string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.get(userID).NavBack[menu] = dest
}

// ── Session cleanup ───────────────────────────────────────────────────────────

// Clear removes the entire session for userID (call on character deletion, logout, etc.).
func (m *Manager) Clear(userID int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sessions, userID)
}

// ActiveCount returns the number of sessions currently in memory (for monitoring).
func (m *Manager) ActiveCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.sessions)
}
