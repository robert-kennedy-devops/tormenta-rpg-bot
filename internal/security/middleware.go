package security

// middleware.go — security gate that every Telegram callback and message must
// pass before reaching game logic.
//
// Responsibilities (applied in order):
//   1. User ID validation (positive int64)
//   2. Raw callback data validation (safe charset, length ≤ 64)
//   3. Callback deduplication / multi-click guard
//   4. Per-user rate limiting (per action key)
//   5. Speed-hack detection
//   6. Behaviour anomaly observation
//   7. Flagged-user block (users under active investigation)
//   8. Permission checks (GM-only commands)
//
// Integration:
//   Wrap your existing callback handler:
//
//       gate := security.NewGate(gmIDs)
//       bot.HandleCallback(func(cb *tgbotapi.CallbackQuery) {
//           if err := gate.CheckCallback(cb); err != nil {
//               secLog.Warn(cb.From.ID, "callback blocked", err.Error())
//               return
//           }
//           existingHandler(cb)
//       })

import (
	"errors"
	"fmt"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// ─── Gate ─────────────────────────────────────────────────────────────────────

// Gate is the main security middleware.  It combines the Validator, RateLimiter,
// ExploitDetector and BehaviorDetector into a single check point.
type Gate struct {
	v       *Validator
	rl      *UserRateLimiter
	exploit *ExploitDetector
	behav   *BehaviorDetector
	eco     *EconomyValidator
	log     *SecLogger

	gmIDs map[int64]bool

	// callbackTTL is the deduplication window for callback IDs.
	callbackTTL time.Duration
}

// GateOptions allows tuning the Gate at construction time.
type GateOptions struct {
	// RateLimits overrides the default per-action limits.
	RateLimits map[ActionKey]Limit
	// BehaviorCfg overrides the default behaviour-detection thresholds.
	BehaviorCfg BehaviorConfig
	// CallbackTTL is the multi-click deduplication window (default 1s).
	CallbackTTL time.Duration
	// Logger is an optional pre-configured SecLogger. If nil a default is used.
	Logger *SecLogger
}

// NewGate creates a Gate using the provided set of GM Telegram user IDs.
// Pass an empty map if you have no GM commands to protect.
func NewGate(gmIDs map[int64]bool, opts ...GateOptions) *Gate {
	var o GateOptions
	if len(opts) > 0 {
		o = opts[0]
	}
	if o.CallbackTTL == 0 {
		o.CallbackTTL = time.Second
	}
	logger := o.Logger
	if logger == nil {
		logger = Log
	}

	var rl *UserRateLimiter
	if o.RateLimits != nil {
		rl = NewUserRateLimiter(o.RateLimits)
	} else {
		rl = RL
	}

	var behav *BehaviorDetector
	if o.BehaviorCfg != (BehaviorConfig{}) {
		behav = NewBehaviorDetector(o.BehaviorCfg)
	} else {
		behav = Behavior
	}

	return &Gate{
		v:           V,
		rl:          rl,
		exploit:     Detector,
		behav:       behav,
		eco:         Economy,
		log:         logger,
		gmIDs:       gmIDs,
		callbackTTL: o.CallbackTTL,
	}
}

// ─── CheckCallback ────────────────────────────────────────────────────────────

// CheckCallback is the main entry point.  Call it at the top of every callback
// handler.  Returns a non-nil error (and logs it) when the request should be
// rejected.
func (g *Gate) CheckCallback(cb *tgbotapi.CallbackQuery) error {
	if cb == nil || cb.From == nil {
		return errors.New("nil callback or user")
	}
	userID := cb.From.ID

	// 1. Validate user ID.
	if err := g.v.UserID(userID); err != nil {
		g.log.SecurityEvent(userID, EventInvalidInput, "user_id", "invalid userID", "")
		return err
	}

	// 2. Validate callback data.
	data := cb.Data
	if _, err := g.v.SafeCallback(data); err != nil {
		g.log.SecurityEvent(userID, EventInvalidInput, "callback_data",
			fmt.Sprintf("data=%q err=%v", data, err), "")
		return err
	}

	// 3. Speed-hack / duplicate action detection.
	if !g.exploit.RecordAction(userID) {
		g.log.SecurityEvent(userID, EventExploit, string(ExploitSpeedHack),
			"sub-human interval", "")
		return errors.New("action rate too high")
	}

	// 4. Callback deduplication (prevents multi-click exploits).
	if !g.exploit.CallbackAllowed(userID, data, g.callbackTTL) {
		g.log.SecurityEvent(userID, EventExploit, string(ExploitCallbackDup),
			"duplicate callback within window", "")
		return errors.New("duplicate callback")
	}

	// 5. Per-user rate limiting (use ActionCallback as the broadest bucket).
	action := callbackToAction(data)
	if !g.rl.Allow(userID, action) {
		g.log.SecurityEvent(userID, EventRateLimit, string(action),
			"rate limit exceeded", "")
		return fmt.Errorf("rate limit: %s", action)
	}

	// 6. Behaviour anomaly observation.
	if ev := g.behav.ObserveCallback(userID); ev != nil {
		g.log.SecurityEvent(userID, EventAnomaly, string(ev.Kind), ev.Detail, "")
	}

	// 7. Flagged-user block.
	if g.behav.IsFlagged(userID) {
		g.log.SecurityEvent(userID, EventBlocked, "flagged_user",
			"user is under active review", "")
		return errors.New("account under review — contact support")
	}

	return nil
}

// ─── CheckMessage ─────────────────────────────────────────────────────────────

// CheckMessage validates an incoming text message (command or free text).
func (g *Gate) CheckMessage(msg *tgbotapi.Message) error {
	if msg == nil || msg.From == nil {
		return errors.New("nil message or user")
	}
	userID := msg.From.ID

	if err := g.v.UserID(userID); err != nil {
		return err
	}

	if !g.rl.Allow(userID, ActionGeneral) {
		g.log.SecurityEvent(userID, EventRateLimit, "message", "general rate exceeded", "")
		return errors.New("too many messages")
	}

	if ev := g.behav.ObserveAction(userID); ev != nil {
		g.log.SecurityEvent(userID, EventAnomaly, string(ev.Kind), ev.Detail, "")
	}

	if g.behav.IsFlagged(userID) {
		return errors.New("account under review — contact support")
	}

	return nil
}

// ─── CheckGMCommand ───────────────────────────────────────────────────────────

// CheckGMCommand returns an error when the caller is not in the GM list.
func (g *Gate) CheckGMCommand(userID int64) error {
	if !g.gmIDs[userID] {
		g.log.SecurityEvent(userID, EventPermissionDenied, "gm_command",
			"non-GM attempted GM action", "")
		return errors.New("permission denied: not a GM")
	}
	return nil
}

// ─── CheckShopBuy ─────────────────────────────────────────────────────────────

// CheckShopBuy validates a shop purchase through the full economy pipeline.
// Callers still need to apply the player lock themselves before calling this.
func (g *Gate) CheckShopBuy(userID int64, charID int, cart []CartItem, gold, diamonds int) error {
	if !g.rl.Allow(userID, ActionShopBuy) {
		g.log.SecurityEvent(userID, EventRateLimit, "shop_buy", "buy rate exceeded", "")
		return errors.New("too many purchase attempts")
	}
	return g.eco.ValidatePurchase(userID, charID, cart, gold, diamonds)
}

// CheckShopSell validates a sell operation.
func (g *Gate) CheckShopSell(userID int64, charID int, items []SellItem, currentGold int) error {
	if !g.rl.Allow(userID, ActionShopSell) {
		g.log.SecurityEvent(userID, EventRateLimit, "shop_sell", "sell rate exceeded", "")
		return errors.New("too many sell attempts")
	}
	return g.eco.ValidateSale(userID, charID, items, currentGold)
}

// CheckDrop validates a loot drop before adding items to the inventory.
func (g *Gate) CheckDrop(userID int64, charID, inventorySize int, drops []DropEntry) error {
	if ev := g.behav.ObserveDungeonKill(userID); ev != nil {
		g.log.SecurityEvent(userID, EventAnomaly, string(ev.Kind), ev.Detail, "")
	}
	return g.eco.ValidateDrop(userID, charID, drops, inventorySize)
}

// CheckPaymentCredit validates a diamond credit from payment and ensures the
// transaction ID has not been processed before.
func (g *Gate) CheckPaymentCredit(userID int64, charID int, txID string, diamonds int) error {
	if err := g.eco.ValidatePaymentCredit(userID, charID, txID, diamonds); err != nil {
		g.log.SecurityEvent(userID, EventEconomyViolation, "payment_credit", err.Error(), "")
		return err
	}
	if !g.exploit.TxAllowed(userID, txID) {
		g.log.SecurityEvent(userID, EventExploit, string(ExploitTxDup),
			"txID="+txID, "duplicate payment delivery")
		return fmt.Errorf("transaction %s already processed", txID)
	}
	return nil
}

// ─── Helper: map callback prefix → ActionKey ──────────────────────────────────

// callbackToAction maps the beginning of a callback data string to the most
// specific ActionKey for rate-limiting purposes.
func callbackToAction(data string) ActionKey {
	switch {
	case strings.HasPrefix(data, "shop_confirm_buy"),
		strings.HasPrefix(data, "shop_buy"):
		return ActionShopBuy
	case strings.HasPrefix(data, "shop_sell"),
		strings.HasPrefix(data, "sell_confirm"):
		return ActionShopSell
	case strings.HasPrefix(data, "dungeon"),
		strings.HasPrefix(data, "enter_dungeon"):
		return ActionDungeon
	case strings.HasPrefix(data, "combat"),
		strings.HasPrefix(data, "use_skill"),
		strings.HasPrefix(data, "attack"):
		return ActionCombat
	case strings.HasPrefix(data, "pvp"),
		strings.HasPrefix(data, "arena"):
		return ActionPVP
	case strings.HasPrefix(data, "market"),
		strings.HasPrefix(data, "listing"):
		return ActionMarket
	case strings.HasPrefix(data, "auction"),
		strings.HasPrefix(data, "bid"):
		return ActionAuction
	case strings.HasPrefix(data, "guild_bank"),
		strings.HasPrefix(data, "deposit"),
		strings.HasPrefix(data, "withdraw"):
		return ActionGuildBank
	case strings.HasPrefix(data, "forge"):
		return ActionForge
	default:
		return ActionCallback
	}
}
