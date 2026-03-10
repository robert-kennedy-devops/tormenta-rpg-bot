// Package security provides input validation, rate limiting, anti-exploit
// detection, economy integrity checks, behaviour anomaly detection and
// structured security logging for the Tormenta bot.
package security

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

// ─── Sentinel errors ──────────────────────────────────────────────────────────

var (
	ErrEmptyInput        = errors.New("security: empty input")
	ErrInputTooLong      = errors.New("security: input too long")
	ErrInvalidCharacters = errors.New("security: invalid characters in input")
	ErrInvalidID         = errors.New("security: invalid ID")
	ErrInvalidAmount     = errors.New("security: invalid amount")
	ErrNegativeAmount    = errors.New("security: amount must be positive")
	ErrAmountTooLarge    = errors.New("security: amount exceeds maximum allowed")
	ErrInvalidItemID     = errors.New("security: invalid item ID")
	ErrInvalidSlot       = errors.New("security: invalid equipment slot")
	ErrInvalidCallback   = errors.New("security: invalid callback data")
	ErrInvalidCommand    = errors.New("security: invalid command")
)

// ─── Limits ───────────────────────────────────────────────────────────────────

const (
	MaxCallbackDataLen = 64    // Telegram platform limit
	MaxCommandLen      = 64
	MaxItemIDLen       = 64
	MaxNameLen         = 32
	MaxGoldAmount      = 10_000_000
	MaxDiamondAmount   = 100_000
	MaxItemQuantity    = 9_999
	MaxInventorySlots  = 200
)

// ─── Allowed sets ─────────────────────────────────────────────────────────────

var allowedEquipSlots = map[string]bool{
	"weapon": true, "head": true, "chest": true, "hands": true,
	"legs": true, "feet": true, "offhand": true,
	"accessory1": true, "accessory2": true,
}

var allowedPayWith = map[string]bool{
	"gold": true, "diamonds": true,
}

var allowedState = map[string]bool{
	"idle": true, "combat": true, "dungeon": true,
	"dungeon_combat": true, "pvp": true, "auto_hunt": true,
}

// safeIDRegexp matches alphanumeric, underscores and hyphens only (no path
// separators, angle brackets or SQL special characters).
var safeIDRegexp = regexp.MustCompile(`^[A-Za-z0-9_\-]+$`)

// safeCallbackRegexp: same charset plus colon and dot for compound keys.
var safeCallbackRegexp = regexp.MustCompile(`^[A-Za-z0-9_\-:.]+$`)

// ─── Validator ────────────────────────────────────────────────────────────────

// Validator groups all input-validation helpers. Instantiate once and share.
type Validator struct{}

// NewValidator returns a ready-to-use Validator.
func NewValidator() *Validator { return &Validator{} }

// ── Generic helpers ───────────────────────────────────────────────────────────

// SafeString checks that s is non-empty, within maxLen runes and contains only
// printable UTF-8 characters. Returns the trimmed string on success.
func (v *Validator) SafeString(s string, maxLen int) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", ErrEmptyInput
	}
	if utf8.RuneCountInString(s) > maxLen {
		return "", ErrInputTooLong
	}
	return s, nil
}

// SafeID validates that s looks like a safe identifier (alphanumeric / _ / -).
func (v *Validator) SafeID(s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", ErrEmptyInput
	}
	if len(s) > MaxItemIDLen {
		return "", ErrInputTooLong
	}
	if !safeIDRegexp.MatchString(s) {
		return "", ErrInvalidID
	}
	return s, nil
}

// SafeCallback validates Telegram callback_data: length ≤ 64 bytes and safe
// charset only (no shell metacharacters or SQL injection vectors).
func (v *Validator) SafeCallback(data string) (string, error) {
	if data == "" {
		return "", ErrEmptyInput
	}
	if len(data) > MaxCallbackDataLen {
		return "", ErrInputTooLong
	}
	if !safeCallbackRegexp.MatchString(data) {
		return "", ErrInvalidCallback
	}
	return data, nil
}

// SafeCommand strips the leading "/" and validates the command string.
func (v *Validator) SafeCommand(cmd string) (string, error) {
	cmd = strings.TrimPrefix(strings.TrimSpace(cmd), "/")
	if cmd == "" {
		return "", ErrInvalidCommand
	}
	if len(cmd) > MaxCommandLen {
		return "", ErrInputTooLong
	}
	if !safeIDRegexp.MatchString(cmd) {
		return "", ErrInvalidCommand
	}
	return cmd, nil
}

// ── Numeric helpers ───────────────────────────────────────────────────────────

// ParsePositiveInt parses s as a decimal integer and requires 1 ≤ n ≤ max.
// Returns the parsed value or an error (never silently returns 0 for bad input).
func (v *Validator) ParsePositiveInt(s string, max int) (int, error) {
	s = strings.TrimSpace(s)
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, ErrInvalidAmount
	}
	if n <= 0 {
		return 0, ErrNegativeAmount
	}
	if n > max {
		return 0, ErrAmountTooLarge
	}
	return n, nil
}

// ParseNonNegativeInt parses s as a decimal integer ≥ 0.
func (v *Validator) ParseNonNegativeInt(s string, max int) (int, error) {
	s = strings.TrimSpace(s)
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, ErrInvalidAmount
	}
	if n < 0 {
		return 0, ErrNegativeAmount
	}
	if n > max {
		return 0, ErrAmountTooLarge
	}
	return n, nil
}

// ParseInt64 parses s as a 64-bit decimal integer with no sign restriction.
func (v *Validator) ParseInt64(s string) (int64, error) {
	s = strings.TrimSpace(s)
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, ErrInvalidID
	}
	return n, nil
}

// ── Domain helpers ────────────────────────────────────────────────────────────

// GoldAmount validates a gold value: must be in [1, MaxGoldAmount].
func (v *Validator) GoldAmount(n int) error {
	if n <= 0 {
		return ErrNegativeAmount
	}
	if n > MaxGoldAmount {
		return ErrAmountTooLarge
	}
	return nil
}

// DiamondAmount validates a diamond value: must be in [1, MaxDiamondAmount].
func (v *Validator) DiamondAmount(n int) error {
	if n <= 0 {
		return ErrNegativeAmount
	}
	if n > MaxDiamondAmount {
		return ErrAmountTooLarge
	}
	return nil
}

// ItemQuantity validates a purchase/sell quantity: must be in [1, MaxItemQuantity].
func (v *Validator) ItemQuantity(n int) error {
	if n <= 0 {
		return ErrNegativeAmount
	}
	if n > MaxItemQuantity {
		return ErrAmountTooLarge
	}
	return nil
}

// EquipSlot returns an error if slot is not one of the recognised equipment slots.
func (v *Validator) EquipSlot(slot string) error {
	if !allowedEquipSlots[slot] {
		return ErrInvalidSlot
	}
	return nil
}

// PayWith validates the payment method string.
func (v *Validator) PayWith(method string) error {
	if !allowedPayWith[method] {
		return errors.New("security: invalid payment method")
	}
	return nil
}

// CharacterState validates a state string against the known FSM states.
func (v *Validator) CharacterState(state string) error {
	if !allowedState[state] {
		return errors.New("security: unknown character state")
	}
	return nil
}

// UserID validates that a Telegram user ID is positive.
func (v *Validator) UserID(id int64) error {
	if id <= 0 {
		return ErrInvalidID
	}
	return nil
}

// ─── Package-level singleton ──────────────────────────────────────────────────

// V is the default package-level Validator. Callers can use security.V.SafeID()
// directly without constructing their own instance.
var V = NewValidator()
