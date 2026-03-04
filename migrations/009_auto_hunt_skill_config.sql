-- Migration 009: Adiciona skill_config à auto_hunt_sessions
-- Necessário porque a migration 007 não incluía essa coluna.

ALTER TABLE auto_hunt_sessions
    ADD COLUMN IF NOT EXISTS skill_config JSONB NOT NULL DEFAULT '{"mode":"attack","skills":[],"potions":[],"heal_at":50,"mana_at":30}'::jsonb;
