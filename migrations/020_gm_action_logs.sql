-- Migration 020: GM action audit log

CREATE TABLE IF NOT EXISTS gm_action_logs (
    id BIGSERIAL PRIMARY KEY,
    gm_player_id BIGINT NOT NULL,
    target_character_id INT NOT NULL DEFAULT 0,
    action VARCHAR(64) NOT NULL,
    details TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_gm_action_logs_gm_created ON gm_action_logs(gm_player_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_gm_action_logs_target_created ON gm_action_logs(target_character_id, created_at DESC);
