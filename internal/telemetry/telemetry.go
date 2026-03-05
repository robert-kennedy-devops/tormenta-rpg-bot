package telemetry

import (
	"encoding/json"
	"os"
	"strings"
	"time"

	"github.com/tormenta-bot/internal/database"
)

const (
	EventPlayerLogin  = "player_login"
	EventDungeonClear = "dungeon_clear"
	EventItemUpgrade  = "item_upgrade"
	EventPixPurchase  = "pix_purchase"
)

func Enabled() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv("TELEMETRY_ENABLED")))
	return v == "1" || v == "true" || v == "yes" || v == "on"
}

func Track(playerID int64, characterID int, event string, payload map[string]interface{}) {
	if !Enabled() || event == "" || database.DB == nil {
		return
	}
	if payload == nil {
		payload = map[string]interface{}{}
	}
	payload["tracked_at"] = time.Now().UTC().Format(time.RFC3339Nano)
	raw, _ := json.Marshal(payload)
	_, _ = database.DB.Exec(`
		INSERT INTO analytics_events (player_id, character_id, event, payload, created_at)
		VALUES ($1,$2,$3,$4::jsonb,NOW())
	`, playerID, characterID, event, string(raw))
}
