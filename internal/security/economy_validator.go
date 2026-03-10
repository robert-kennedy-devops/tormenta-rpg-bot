package security

// economy_validator.go — integrity checks for all economy transactions.
//
// Every function is pure (no I/O) and returns an EconomyViolation on failure,
// which carries machine-readable fields suitable for structured logging.
//
// Integration points:
//   • Call ValidatePurchase before applying a shop buy.
//   • Call ValidateSale before crediting gold from a sell.
//   • Call ValidateDrop before adding loot to the inventory.
//   • Call ValidateGoldChange after any gold mutation to detect inconsistencies.

import (
	"errors"
	"fmt"
)

// ─── Violation ────────────────────────────────────────────────────────────────

// EconomyViolation wraps an economy validation failure with structured context.
type EconomyViolation struct {
	UserID    int64
	CharID    int
	Kind      string
	Detail    string
	Suggested string // human-readable remediation hint
}

func (e *EconomyViolation) Error() string {
	return fmt.Sprintf("economy violation [%s] user=%d char=%d: %s", e.Kind, e.UserID, e.CharID, e.Detail)
}

func violation(userID int64, charID int, kind, detail, suggested string) *EconomyViolation {
	return &EconomyViolation{
		UserID: userID, CharID: charID,
		Kind: kind, Detail: detail, Suggested: suggested,
	}
}

// ─── Constants ────────────────────────────────────────────────────────────────

const (
	// MaxGoldBalance is the maximum gold a single character may hold.
	MaxGoldBalance = 50_000_000
	// MaxDiamondBalance is the maximum diamonds a single character may hold.
	MaxDiamondBalance = 500_000
	// MaxCartItems is the maximum number of distinct items in a single purchase.
	MaxCartItems = 20
	// MaxSellItems is the maximum distinct items in a single sell operation.
	MaxSellItems = 20
	// MaxSingleDropQty is the maximum quantity of a single item per drop roll.
	MaxSingleDropQty = 50
	// MaxInventorySize is the hard cap on total inventory slots.
	MaxInventorySize = 200
)

// ─── EconomyValidator ─────────────────────────────────────────────────────────

// EconomyValidator provides stateless economy integrity checks.
// All methods return nil on success or *EconomyViolation on failure.
type EconomyValidator struct{}

// NewEconomyValidator returns a ready EconomyValidator.
func NewEconomyValidator() *EconomyValidator { return &EconomyValidator{} }

// ── Shop purchase ─────────────────────────────────────────────────────────────

// CartItem represents one line in a purchase cart (used by ValidatePurchase).
type CartItem struct {
	ItemID   string
	Qty      int
	UnitCost int  // in gold or diamonds depending on Diamond flag
	Diamond  bool // true → pay with diamonds
}

// ValidatePurchase checks a proposed shop purchase for integrity:
//   - cart must be non-empty and ≤ MaxCartItems items
//   - every item has a valid ID and qty in [1, MaxItemQuantity]
//   - every unit cost is non-negative
//   - total gold cost does not exceed current balance
//   - total diamond cost does not exceed current balance
//   - totals do not overflow int
func (ev *EconomyValidator) ValidatePurchase(
	userID int64, charID int,
	cart []CartItem,
	currentGold, currentDiamonds int,
) error {
	if len(cart) == 0 {
		return violation(userID, charID, "empty_cart", "no items in cart", "add items before purchasing")
	}
	if len(cart) > MaxCartItems {
		return violation(userID, charID, "cart_overflow",
			fmt.Sprintf("cart has %d items (max %d)", len(cart), MaxCartItems), "")
	}

	totalGold, totalDiam := 0, 0
	seen := make(map[string]bool, len(cart))

	for i, ci := range cart {
		if ci.ItemID == "" {
			return violation(userID, charID, "invalid_item_id",
				fmt.Sprintf("cart[%d]: empty itemID", i), "")
		}
		if !safeIDRegexp.MatchString(ci.ItemID) {
			return violation(userID, charID, "invalid_item_id",
				fmt.Sprintf("cart[%d]: malformed itemID %q", i, ci.ItemID), "")
		}
		if ci.Qty <= 0 || ci.Qty > MaxItemQuantity {
			return violation(userID, charID, "invalid_qty",
				fmt.Sprintf("cart[%d] %s qty=%d out of range [1,%d]", i, ci.ItemID, ci.Qty, MaxItemQuantity), "")
		}
		if ci.UnitCost < 0 {
			return violation(userID, charID, "negative_price",
				fmt.Sprintf("cart[%d] %s unitCost=%d", i, ci.ItemID, ci.UnitCost), "")
		}

		lineTotal := ci.UnitCost * ci.Qty
		if lineTotal < 0 { // overflow guard
			return violation(userID, charID, "cost_overflow",
				fmt.Sprintf("cart[%d] %s lineTotal overflow", i, ci.ItemID), "")
		}

		if ci.Diamond {
			totalDiam += lineTotal
		} else {
			totalGold += lineTotal
		}
		// Overflow guards on running totals.
		if totalGold < 0 || totalDiam < 0 {
			return violation(userID, charID, "cost_overflow", "running total overflow", "")
		}

		// Flag duplicate line items (potential duplication attempt).
		dupKey := fmt.Sprintf("%s:%v", ci.ItemID, ci.Diamond)
		if seen[dupKey] {
			return violation(userID, charID, "duplicate_line_item",
				fmt.Sprintf("item %s appears more than once in cart", ci.ItemID), "merge quantities into single line")
		}
		seen[dupKey] = true
	}

	if totalGold > currentGold {
		return violation(userID, charID, "insufficient_gold",
			fmt.Sprintf("need %d gold, have %d", totalGold, currentGold), "")
	}
	if totalDiam > currentDiamonds {
		return violation(userID, charID, "insufficient_diamonds",
			fmt.Sprintf("need %d diamonds, have %d", totalDiam, currentDiamonds), "")
	}
	return nil
}

// ── Sell operation ────────────────────────────────────────────────────────────

// SellItem represents one line in a sell operation.
type SellItem struct {
	ItemID    string
	Qty       int
	SellPrice int // gold per unit
}

// ValidateSale checks a proposed sell operation for integrity.
func (ev *EconomyValidator) ValidateSale(
	userID int64, charID int,
	items []SellItem,
	currentGold int,
) error {
	if len(items) == 0 {
		return violation(userID, charID, "empty_sell", "no items to sell", "")
	}
	if len(items) > MaxSellItems {
		return violation(userID, charID, "sell_overflow",
			fmt.Sprintf("%d items (max %d)", len(items), MaxSellItems), "")
	}

	proceeds := 0
	for i, si := range items {
		if si.ItemID == "" || !safeIDRegexp.MatchString(si.ItemID) {
			return violation(userID, charID, "invalid_item_id",
				fmt.Sprintf("sell[%d] malformed itemID %q", i, si.ItemID), "")
		}
		if si.Qty <= 0 || si.Qty > MaxItemQuantity {
			return violation(userID, charID, "invalid_qty",
				fmt.Sprintf("sell[%d] %s qty=%d", i, si.ItemID, si.Qty), "")
		}
		if si.SellPrice < 0 {
			return violation(userID, charID, "negative_sell_price",
				fmt.Sprintf("sell[%d] %s price=%d", i, si.ItemID, si.SellPrice), "")
		}
		line := si.SellPrice * si.Qty
		if line < 0 {
			return violation(userID, charID, "proceeds_overflow", "sell proceeds overflow", "")
		}
		proceeds += line
		if proceeds < 0 {
			return violation(userID, charID, "proceeds_overflow", "running proceeds overflow", "")
		}
	}

	// Resulting balance must remain within the hard cap.
	if currentGold+proceeds > MaxGoldBalance {
		return violation(userID, charID, "gold_cap_exceeded",
			fmt.Sprintf("would reach %d gold (cap %d)", currentGold+proceeds, MaxGoldBalance),
			"excess gold will be capped")
	}
	return nil
}

// ── Drop / loot ───────────────────────────────────────────────────────────────

// DropEntry represents one item in a drop roll.
type DropEntry struct {
	ItemID string
	Qty    int
}

// ValidateDrop ensures that a loot drop is within sane bounds.
func (ev *EconomyValidator) ValidateDrop(
	userID int64, charID int,
	drops []DropEntry,
	currentInventorySize int,
) error {
	if len(drops) == 0 {
		return nil // empty drop is fine
	}

	totalNewSlots := 0
	for i, d := range drops {
		if d.ItemID == "" || !safeIDRegexp.MatchString(d.ItemID) {
			return violation(userID, charID, "invalid_drop_item",
				fmt.Sprintf("drop[%d] malformed itemID %q", i, d.ItemID), "")
		}
		if d.Qty <= 0 || d.Qty > MaxSingleDropQty {
			return violation(userID, charID, "invalid_drop_qty",
				fmt.Sprintf("drop[%d] %s qty=%d (max %d)", i, d.ItemID, d.Qty, MaxSingleDropQty), "")
		}
		totalNewSlots++
	}

	if currentInventorySize+totalNewSlots > MaxInventorySize {
		return violation(userID, charID, "inventory_full",
			fmt.Sprintf("inventory %d+%d > %d", currentInventorySize, totalNewSlots, MaxInventorySize),
			"sell or discard items first")
	}
	return nil
}

// ── Gold / diamond mutations ──────────────────────────────────────────────────

// ValidateGoldChange verifies that applying delta to currentGold yields a sane
// result. Use this AFTER every gold mutation as a post-condition check.
func (ev *EconomyValidator) ValidateGoldChange(
	userID int64, charID int,
	currentGold, delta int,
) error {
	newBalance := currentGold + delta
	if newBalance < 0 {
		return violation(userID, charID, "negative_gold_balance",
			fmt.Sprintf("%d + %d = %d", currentGold, delta, newBalance), "")
	}
	if newBalance > MaxGoldBalance {
		return violation(userID, charID, "gold_cap_exceeded",
			fmt.Sprintf("balance would be %d (cap %d)", newBalance, MaxGoldBalance), "")
	}
	return nil
}

// ValidateDiamondChange is the diamond equivalent of ValidateGoldChange.
func (ev *EconomyValidator) ValidateDiamondChange(
	userID int64, charID int,
	currentDiamonds, delta int,
) error {
	newBalance := currentDiamonds + delta
	if newBalance < 0 {
		return violation(userID, charID, "negative_diamond_balance",
			fmt.Sprintf("%d + %d = %d", currentDiamonds, delta, newBalance), "")
	}
	if newBalance > MaxDiamondBalance {
		return violation(userID, charID, "diamond_cap_exceeded",
			fmt.Sprintf("balance would be %d (cap %d)", newBalance, MaxDiamondBalance), "")
	}
	return nil
}

// ── Payment ───────────────────────────────────────────────────────────────────

// ValidatePaymentCredit checks a diamond-crediting operation from a payment
// confirmation. Requires amount > 0 and txID to be a non-empty safe string.
func (ev *EconomyValidator) ValidatePaymentCredit(
	userID int64, charID int,
	txID string, diamondAmount int,
) error {
	if txID == "" {
		return violation(userID, charID, "empty_tx_id", "txID must not be empty", "")
	}
	// txID from payment providers may contain hyphens and alphanumeric chars.
	if len(txID) > 128 {
		return violation(userID, charID, "tx_id_too_long",
			fmt.Sprintf("txID length %d > 128", len(txID)), "")
	}
	if diamondAmount <= 0 {
		return violation(userID, charID, "invalid_diamond_credit",
			fmt.Sprintf("diamondAmount=%d must be > 0", diamondAmount), "")
	}
	if diamondAmount > 10_000 { // largest sane package
		return violation(userID, charID, "diamond_credit_too_large",
			fmt.Sprintf("diamondAmount=%d exceeds max package size", diamondAmount), "")
	}
	return nil
}

// ─── Sentinel errors (returned as plain errors from helpers) ─────────────────

var (
	ErrNegativeBalance = errors.New("economy: balance would go negative")
	ErrBalanceCap      = errors.New("economy: balance would exceed cap")
)

// ─── Package-level singleton ──────────────────────────────────────────────────

// Economy is the default package-level EconomyValidator.
var Economy = NewEconomyValidator()
