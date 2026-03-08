-- Migration 011: Fix pvp_stats table
-- Adds missing columns if the table was created without them
-- Safe to run multiple times (uses IF NOT EXISTS / DO NOTHING patterns)

-- Garante que a tabela existe com estrutura mínima
CREATE TABLE IF NOT EXISTS pvp_stats (
    character_id INT PRIMARY KEY REFERENCES characters(id) ON DELETE CASCADE,
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Adiciona colunas faltantes uma a uma (cada ADD COLUMN é idempotente via IF NOT EXISTS)
ALTER TABLE pvp_stats ADD COLUMN IF NOT EXISTS wins        INT NOT NULL DEFAULT 0;
ALTER TABLE pvp_stats ADD COLUMN IF NOT EXISTS losses      INT NOT NULL DEFAULT 0;
ALTER TABLE pvp_stats ADD COLUMN IF NOT EXISTS draws       INT NOT NULL DEFAULT 0;
ALTER TABLE pvp_stats ADD COLUMN IF NOT EXISTS rating      INT NOT NULL DEFAULT 1000;
ALTER TABLE pvp_stats ADD COLUMN IF NOT EXISTS streak      INT NOT NULL DEFAULT 0;
ALTER TABLE pvp_stats ADD COLUMN IF NOT EXISTS best_streak INT NOT NULL DEFAULT 0;
