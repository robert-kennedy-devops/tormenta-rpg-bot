// safe_state.go — mutex-protected per-user session maps.
//
// The original raw maps (creationState, shopCart, sellCart, …) were package-level
// variables with no synchronisation. Telegram updates for different users are
// processed concurrently, so concurrent reads+writes on the same Go map triggers
// the race detector and can panic at runtime.
//
// This file replaces every bare map with mutex-guarded accessors whose names
// exactly mirror the original access patterns, keeping call-site diffs minimal.
package handlers

import (
	"sync"

	"github.com/tormenta-bot/internal/models"
)

// ── Underlying storage (never accessed directly outside this file) ────────────

var stateMu sync.RWMutex

var (
	_creationState = map[int64]map[string]string{}
	_shopQtyState  = map[int64]*models.ShopQtyState{}
	_shopCart      = map[int64]*models.ShopCart{}
	_sellCart      = map[int64]*models.SellCart{}
	_navCurrent    = map[int64]string{}
	_navBack       = map[int64]map[string]string{}
)

// ── creationState helpers ─────────────────────────────────────────────────────

// csGet returns the creation map for a user (nil if absent). Read-only use.
func csGet(userID int64) map[string]string {
	stateMu.RLock()
	defer stateMu.RUnlock()
	return _creationState[userID]
}

// csInit creates (or resets) the creation map for a user.
func csInit(userID int64) {
	stateMu.Lock()
	defer stateMu.Unlock()
	_creationState[userID] = map[string]string{}
}

// csSet sets a single key in the creation map; creates the map if needed.
func csSet(userID int64, key, val string) {
	stateMu.Lock()
	defer stateMu.Unlock()
	if _creationState[userID] == nil {
		_creationState[userID] = map[string]string{}
	}
	_creationState[userID][key] = val
}

// csField returns one field from the creation map (empty string if absent).
func csField(userID int64, key string) string {
	stateMu.RLock()
	defer stateMu.RUnlock()
	m := _creationState[userID]
	if m == nil {
		return ""
	}
	return m[key]
}

// csDelete removes one key from the creation map.
func csDelete(userID int64, key string) {
	stateMu.Lock()
	defer stateMu.Unlock()
	if m := _creationState[userID]; m != nil {
		delete(m, key)
	}
}

// csClear removes the whole creation entry for a user.
func csClear(userID int64) {
	stateMu.Lock()
	defer stateMu.Unlock()
	delete(_creationState, userID)
}

// ── shopQtyState helpers ──────────────────────────────────────────────────────

func sqGet(userID int64) *models.ShopQtyState {
	stateMu.RLock()
	defer stateMu.RUnlock()
	return _shopQtyState[userID]
}

func sqSet(userID int64, v *models.ShopQtyState) {
	stateMu.Lock()
	defer stateMu.Unlock()
	_shopQtyState[userID] = v
}

func sqClear(userID int64) {
	stateMu.Lock()
	defer stateMu.Unlock()
	delete(_shopQtyState, userID)
}

// ── shopCart helpers ──────────────────────────────────────────────────────────

func scGet(userID int64) *models.ShopCart {
	stateMu.RLock()
	defer stateMu.RUnlock()
	return _shopCart[userID]
}

func scSet(userID int64, c *models.ShopCart) {
	stateMu.Lock()
	defer stateMu.Unlock()
	_shopCart[userID] = c
}

func scClear(userID int64) {
	stateMu.Lock()
	defer stateMu.Unlock()
	delete(_shopCart, userID)
}

// ── sellCart helpers ──────────────────────────────────────────────────────────

func sellGet(userID int64) *models.SellCart {
	stateMu.RLock()
	defer stateMu.RUnlock()
	return _sellCart[userID]
}

func sellSet(userID int64, c *models.SellCart) {
	stateMu.Lock()
	defer stateMu.Unlock()
	_sellCart[userID] = c
}

func sellClear(userID int64) {
	stateMu.Lock()
	defer stateMu.Unlock()
	delete(_sellCart, userID)
}

// ── navCurrent helpers ────────────────────────────────────────────────────────

func navGet(userID int64) string {
	stateMu.RLock()
	defer stateMu.RUnlock()
	return _navCurrent[userID]
}

func navSet(userID int64, screen string) {
	stateMu.Lock()
	defer stateMu.Unlock()
	_navCurrent[userID] = screen
}

// ── navBack helpers ───────────────────────────────────────────────────────────

func nbGet(userID int64, menu string) string {
	stateMu.RLock()
	defer stateMu.RUnlock()
	m := _navBack[userID]
	if m == nil {
		return ""
	}
	return m[menu]
}

func nbSet(userID int64, menu, dest string) {
	stateMu.Lock()
	defer stateMu.Unlock()
	if _navBack[userID] == nil {
		_navBack[userID] = map[string]string{}
	}
	_navBack[userID][menu] = dest
}

// nbGetAll returns the full navBack sub-map for a user (read-only snapshot).
func nbGetAll(userID int64) map[string]string {
	stateMu.RLock()
	defer stateMu.RUnlock()
	src := _navBack[userID]
	if src == nil {
		return nil
	}
	out := make(map[string]string, len(src))
	for k, v := range src {
		out[k] = v
	}
	return out
}
