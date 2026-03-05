-- Generic timer storage for cooldowns and asynchronous systems

CREATE TABLE IF NOT EXISTS player_timers (
    player_id BIGINT NOT NULL,
    key VARCHAR(64) NOT NULL,
    end_time BIGINT NOT NULL, -- unix timestamp (seconds)
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (player_id, key)
);

CREATE INDEX IF NOT EXISTS idx_player_timers_end_time ON player_timers(end_time);

