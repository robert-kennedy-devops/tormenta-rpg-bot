-- Migration 010: adiciona xp_boost_expiry para Bênção do Sábio
-- O campo guarda até quando o bônus de +50% XP está ativo.
-- Valor NULL ou no passado = sem boost ativo.

ALTER TABLE characters
    ADD COLUMN IF NOT EXISTS xp_boost_expiry TIMESTAMPTZ NOT NULL DEFAULT '1970-01-01 00:00:00+00';
