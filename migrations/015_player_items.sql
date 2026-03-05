-- Player item instances for upgrade/rarity progression

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS player_items (
    instance_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    character_id INT NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    template_id VARCHAR(64) NOT NULL,
    upgrade_level INT NOT NULL DEFAULT 0,
    broken BOOLEAN NOT NULL DEFAULT FALSE,
    equipped BOOLEAN NOT NULL DEFAULT FALSE,
    equipped_slot VARCHAR(20) DEFAULT NULL,
    rarity_override VARCHAR(20) DEFAULT NULL,
    stat_overrides JSONB DEFAULT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_player_items_character ON player_items(character_id);
CREATE INDEX IF NOT EXISTS idx_player_items_template ON player_items(character_id, template_id);
CREATE INDEX IF NOT EXISTS idx_player_items_equipped ON player_items(character_id, equipped, equipped_slot);

