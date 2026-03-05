-- Migration 021: optimize centralized energy regeneration scans
CREATE INDEX IF NOT EXISTS idx_characters_energy_tick
ON characters(last_energy_update)
WHERE energy < energy_max;

