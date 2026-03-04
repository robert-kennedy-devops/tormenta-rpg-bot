-- Migration 003: Energy System & Diamonds
-- Adds stamina/energy, diamonds (premium), sell price

-- Energy & Diamonds on characters
ALTER TABLE characters
  ADD COLUMN IF NOT EXISTS energy        INT NOT NULL DEFAULT 100,
  ADD COLUMN IF NOT EXISTS energy_max    INT NOT NULL DEFAULT 100,
  ADD COLUMN IF NOT EXISTS energy_regen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  ADD COLUMN IF NOT EXISTS diamonds      INT NOT NULL DEFAULT 0;

-- Diamond transaction log
CREATE TABLE IF NOT EXISTS diamond_log (
    id           SERIAL PRIMARY KEY,
    character_id INT REFERENCES characters(id) ON DELETE CASCADE,
    amount       INT NOT NULL,          -- positive = gain, negative = spend
    reason       VARCHAR(100) NOT NULL,
    created_at   TIMESTAMPTZ DEFAULT NOW()
);

-- Daily login bonus control
CREATE TABLE IF NOT EXISTS daily_bonus (
    character_id INT PRIMARY KEY REFERENCES characters(id) ON DELETE CASCADE,
    last_claim   TIMESTAMPTZ NOT NULL DEFAULT '2000-01-01'
);

-- Shop quantity state (tracks pending buy quantity per user)
-- Stored in-memory in handlers; no table needed.

CREATE INDEX IF NOT EXISTS idx_diamond_log_char ON diamond_log(character_id);
