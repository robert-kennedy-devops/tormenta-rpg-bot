-- Tormenta RPG Bot - Database Schema
-- Migration 001 - Initial Setup

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- =============================================
-- PLAYERS
-- =============================================
CREATE TABLE IF NOT EXISTS players (
    id BIGINT PRIMARY KEY,  -- Telegram user ID
    username VARCHAR(100),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- =============================================
-- CHARACTERS
-- =============================================
CREATE TABLE IF NOT EXISTS characters (
    id SERIAL PRIMARY KEY,
    player_id BIGINT REFERENCES players(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    race VARCHAR(20) NOT NULL,       -- human, elf, dwarf, halforc
    class VARCHAR(20) NOT NULL,      -- warrior, mage, rogue, archer

    -- Core Stats
    level INT DEFAULT 1,
    experience INT DEFAULT 0,
    experience_next INT DEFAULT 100,

    -- Health & Energy
    hp INT DEFAULT 0,
    hp_max INT DEFAULT 0,
    mp INT DEFAULT 0,
    mp_max INT DEFAULT 0,

    -- Attributes
    strength INT DEFAULT 10,
    dexterity INT DEFAULT 10,
    constitution INT DEFAULT 10,
    intelligence INT DEFAULT 10,
    wisdom INT DEFAULT 10,
    charisma INT DEFAULT 10,

    -- Combat Stats
    attack INT DEFAULT 5,
    defense INT DEFAULT 5,
    magic_attack INT DEFAULT 5,
    magic_defense INT DEFAULT 5,
    speed INT DEFAULT 5,

    -- Currency
    gold INT DEFAULT 50,

    -- Location
    current_map VARCHAR(50) DEFAULT 'village',

    -- State
    state VARCHAR(50) DEFAULT 'idle', -- idle, combat, traveling, shop
    combat_monster_id VARCHAR(50),
    combat_monster_hp INT DEFAULT 0,

    -- Skill Points
    skill_points INT DEFAULT 0,

    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- =============================================
-- LEARNED SKILLS
-- =============================================
CREATE TABLE IF NOT EXISTS character_skills (
    id SERIAL PRIMARY KEY,
    character_id INT REFERENCES characters(id) ON DELETE CASCADE,
    skill_id VARCHAR(50) NOT NULL,
    level INT DEFAULT 1,
    UNIQUE(character_id, skill_id)
);

-- =============================================
-- INVENTORY
-- =============================================
CREATE TABLE IF NOT EXISTS inventory (
    id SERIAL PRIMARY KEY,
    character_id INT REFERENCES characters(id) ON DELETE CASCADE,
    item_id VARCHAR(50) NOT NULL,
    item_type VARCHAR(20) NOT NULL,  -- weapon, armor, consumable
    quantity INT DEFAULT 1,
    equipped BOOLEAN DEFAULT FALSE,
    UNIQUE(character_id, item_id)
);

-- =============================================
-- COMBAT LOG
-- =============================================
CREATE TABLE IF NOT EXISTS combat_log (
    id SERIAL PRIMARY KEY,
    character_id INT REFERENCES characters(id) ON DELETE CASCADE,
    monster_id VARCHAR(50),
    result VARCHAR(20),  -- win, lose, flee
    exp_gained INT DEFAULT 0,
    gold_gained INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW()
);

-- =============================================
-- TRAVEL LOG
-- =============================================
CREATE TABLE IF NOT EXISTS travel_log (
    id SERIAL PRIMARY KEY,
    character_id INT REFERENCES characters(id) ON DELETE CASCADE,
    from_map VARCHAR(50),
    to_map VARCHAR(50),
    created_at TIMESTAMP DEFAULT NOW()
);

-- =============================================
-- INDEXES
-- =============================================
CREATE INDEX IF NOT EXISTS idx_characters_player_id ON characters(player_id);
CREATE INDEX IF NOT EXISTS idx_inventory_character_id ON inventory(character_id);
CREATE INDEX IF NOT EXISTS idx_character_skills_character_id ON character_skills(character_id);

-- =============================================
-- TRIGGER: updated_at
-- =============================================
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_characters_updated_at BEFORE UPDATE ON characters
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_players_updated_at BEFORE UPDATE ON players
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
