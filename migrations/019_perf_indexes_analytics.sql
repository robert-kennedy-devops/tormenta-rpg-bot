-- Migration 019: performance indexes + optional analytics table

-- characters
CREATE INDEX IF NOT EXISTS idx_characters_updated_at ON characters(updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_characters_state ON characters(state);
CREATE INDEX IF NOT EXISTS idx_characters_level ON characters(level DESC);

-- inventory
CREATE INDEX IF NOT EXISTS idx_inventory_char_item ON inventory(character_id, item_id);
CREATE INDEX IF NOT EXISTS idx_inventory_char_type ON inventory(character_id, item_type);

-- player_items
CREATE INDEX IF NOT EXISTS idx_player_items_char_slot ON player_items(character_id, equipped_slot, equipped);

-- pix_payments
CREATE INDEX IF NOT EXISTS idx_pix_status_created ON pix_payments(status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_pix_expires_status ON pix_payments(status, expires_at);

-- auto_hunt_sessions
CREATE INDEX IF NOT EXISTS idx_auto_hunt_char_status ON auto_hunt_sessions(character_id, status);
CREATE INDEX IF NOT EXISTS idx_auto_hunt_last_tick ON auto_hunt_sessions(last_tick_at);

-- dungeon_runs
CREATE INDEX IF NOT EXISTS idx_dungeon_runs_state_started ON dungeon_runs(state, started_at);

-- optional analytics
CREATE TABLE IF NOT EXISTS analytics_events (
    id BIGSERIAL PRIMARY KEY,
    player_id BIGINT NOT NULL,
    character_id INT NOT NULL DEFAULT 0,
    event VARCHAR(64) NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_analytics_event_created ON analytics_events(event, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_analytics_player_created ON analytics_events(player_id, created_at DESC);
