-- Professional energy timestamp model
-- Adds unix-based last_energy_update while keeping legacy energy_regen_at compatibility.

ALTER TABLE characters
  ADD COLUMN IF NOT EXISTS last_energy_update BIGINT;

-- Backfill from legacy energy_regen_at when available.
UPDATE characters
SET last_energy_update = EXTRACT(EPOCH FROM COALESCE(energy_regen_at, NOW()))::BIGINT
WHERE last_energy_update IS NULL OR last_energy_update <= 0;

ALTER TABLE characters
  ALTER COLUMN last_energy_update SET DEFAULT EXTRACT(EPOCH FROM NOW())::BIGINT;

