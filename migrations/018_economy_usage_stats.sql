-- Migration 018: usage stats for dynamic economy pricing

CREATE TABLE IF NOT EXISTS item_usage_stats (
    item_id      TEXT PRIMARY KEY,
    purchase_count BIGINT NOT NULL DEFAULT 0,
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_item_usage_updated_at ON item_usage_stats(updated_at);
