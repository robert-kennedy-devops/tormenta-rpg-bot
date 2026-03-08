-- Migration 007: VIP system + auto hunt

-- VIP status on players table
ALTER TABLE players
    ADD COLUMN IF NOT EXISTS is_vip        BOOLEAN     NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS vip_expires_at TIMESTAMPTZ DEFAULT NULL;

-- Auto hunt sessions
CREATE TABLE IF NOT EXISTS auto_hunt_sessions (
    id              SERIAL PRIMARY KEY,
    character_id    INT REFERENCES characters(id) ON DELETE CASCADE,
    map_id          VARCHAR(50) NOT NULL,
    started_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_tick_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ticks_done      INT NOT NULL DEFAULT 0,
    total_xp        INT NOT NULL DEFAULT 0,
    total_gold      INT NOT NULL DEFAULT 0,
    total_kills     INT NOT NULL DEFAULT 0,
    status          VARCHAR(20) NOT NULL DEFAULT 'running', -- running | stopped | out_of_energy
    UNIQUE(character_id)
);

CREATE INDEX IF NOT EXISTS idx_auto_hunt_status ON auto_hunt_sessions(status)
    WHERE status = 'running';
