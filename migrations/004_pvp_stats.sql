CREATE TABLE IF NOT EXISTS pvp_stats (
    id SERIAL PRIMARY KEY,
    character_id INT NOT NULL,
    kills INT NOT NULL DEFAULT 0,
    deaths INT NOT NULL DEFAULT 0
);