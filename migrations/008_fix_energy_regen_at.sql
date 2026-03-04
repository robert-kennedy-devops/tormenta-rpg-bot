-- Migration 008: Fix energy_regen_at for existing characters
-- Reseta EnergyRegenAt para NOW() em personagens com valor NULL ou muito antigo
UPDATE characters
SET energy_regen_at = NOW()
WHERE energy_regen_at IS NULL
   OR energy_regen_at < NOW() - INTERVAL '30 days';
