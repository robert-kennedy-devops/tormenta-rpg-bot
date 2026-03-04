-- ═══════════════════════════════════════════════════════════
-- Migration 005: Ban system, drop system, pvp fixes
-- Run after 004_big_update.sql
-- ═══════════════════════════════════════════════════════════

-- ── Ban column on players ─────────────────────────────────
ALTER TABLE players ADD COLUMN IF NOT EXISTS banned BOOLEAN NOT NULL DEFAULT FALSE;

-- ── Deaths column on characters (if not already added) ───
ALTER TABLE characters ADD COLUMN IF NOT EXISTS deaths INT NOT NULL DEFAULT 0;

-- ── PVP draws column (some installs may be missing it) ───
ALTER TABLE pvp_stats ADD COLUMN IF NOT EXISTS draws INT NOT NULL DEFAULT 0;

-- ── Ensure pvp_challenges table exists (some installs use pvp_matches) ─
CREATE TABLE IF NOT EXISTS pvp_challenges (
    id             SERIAL PRIMARY KEY,
    challenger_id  INT REFERENCES characters(id) ON DELETE CASCADE,
    defender_id    INT REFERENCES characters(id) ON DELETE CASCADE,
    stake_gold     INT NOT NULL DEFAULT 0,
    state          VARCHAR(20) NOT NULL DEFAULT 'pending',
    turn           INT NOT NULL DEFAULT 0,
    challenger_hp  INT NOT NULL DEFAULT 0,
    defender_hp    INT NOT NULL DEFAULT 0,
    winner_id      INT REFERENCES characters(id),
    created_at     TIMESTAMPTZ DEFAULT NOW(),
    expires_at     TIMESTAMPTZ DEFAULT NOW() + INTERVAL '5 minutes'
);

-- ── Ensure all base tables exist ──────────────────────────
CREATE TABLE IF NOT EXISTS dungeon_runs (
    id              SERIAL PRIMARY KEY,
    character_id    INT REFERENCES characters(id) ON DELETE CASCADE,
    dungeon_id      VARCHAR(50) NOT NULL,
    floor           INT NOT NULL DEFAULT 1,
    state           VARCHAR(20) NOT NULL DEFAULT 'active',
    monsters_killed INT NOT NULL DEFAULT 0,
    started_at      TIMESTAMPTZ DEFAULT NOW(),
    finished_at     TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS dungeon_best (
    character_id    INT REFERENCES characters(id) ON DELETE CASCADE,
    dungeon_id      VARCHAR(50) NOT NULL,
    best_floor      INT NOT NULL DEFAULT 0,
    completions     INT NOT NULL DEFAULT 0,
    updated_at      TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (character_id, dungeon_id)
);

CREATE TABLE IF NOT EXISTS pvp_stats (
    character_id    INT PRIMARY KEY REFERENCES characters(id) ON DELETE CASCADE,
    wins            INT NOT NULL DEFAULT 0,
    losses          INT NOT NULL DEFAULT 0,
    draws           INT NOT NULL DEFAULT 0,
    rating          INT NOT NULL DEFAULT 1000,
    streak          INT NOT NULL DEFAULT 0,
    best_streak     INT NOT NULL DEFAULT 0,
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS pix_payments (
    id              SERIAL PRIMARY KEY,
    character_id    INT REFERENCES characters(id) ON DELETE CASCADE,
    package_id      VARCHAR(20) NOT NULL,
    diamonds        INT NOT NULL,
    amount_brl      DECIMAL(10,2) NOT NULL,
    txid            VARCHAR(40) UNIQUE NOT NULL,
    status          VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    paid_at         TIMESTAMPTZ,
    expires_at      TIMESTAMPTZ DEFAULT NOW() + INTERVAL '30 minutes'
);

CREATE TABLE IF NOT EXISTS daily_bonus (
    character_id    INT PRIMARY KEY REFERENCES characters(id) ON DELETE CASCADE,
    last_claim      TIMESTAMPTZ NOT NULL DEFAULT '2000-01-01'
);

CREATE TABLE IF NOT EXISTS image_cache (
    key        VARCHAR(100) PRIMARY KEY,
    file_id    VARCHAR(200) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS diamond_log (
    id              SERIAL PRIMARY KEY,
    character_id    INT REFERENCES characters(id) ON DELETE CASCADE,
    amount          INT NOT NULL,
    reason          VARCHAR(100) NOT NULL,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS combat_log (
    id              SERIAL PRIMARY KEY,
    character_id    INT REFERENCES characters(id) ON DELETE CASCADE,
    monster_id      VARCHAR(50),
    result          VARCHAR(20),
    exp_gained      INT DEFAULT 0,
    gold_gained     INT DEFAULT 0,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

-- ── Indexes ───────────────────────────────────────────────
CREATE INDEX IF NOT EXISTS idx_dungeon_char_state   ON dungeon_runs(character_id, state);
CREATE INDEX IF NOT EXISTS idx_pvp_challenger       ON pvp_challenges(challenger_id, state);
CREATE INDEX IF NOT EXISTS idx_pvp_defender         ON pvp_challenges(defender_id, state);
CREATE INDEX IF NOT EXISTS idx_pix_txid             ON pix_payments(txid);
CREATE INDEX IF NOT EXISTS idx_pix_char             ON pix_payments(character_id, status);
CREATE INDEX IF NOT EXISTS idx_diamond_log_char     ON diamond_log(character_id);
CREATE INDEX IF NOT EXISTS idx_combat_log_char      ON combat_log(character_id);
CREATE INDEX IF NOT EXISTS idx_players_banned       ON players(banned) WHERE banned = TRUE;
