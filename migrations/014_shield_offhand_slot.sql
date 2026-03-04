-- ============================================================
-- MIGRATION 014: Dedicated offhand slot for shields
-- Move shield equipment to slot `offhand`.
-- ============================================================

ALTER TABLE inventory ADD COLUMN IF NOT EXISTS slot VARCHAR(20) DEFAULT NULL;

-- Qualquer escudo equipado deve ocupar slot de escudo.
UPDATE inventory
SET slot = 'offhand'
WHERE equipped = true
  AND (
    item_id = 'iron_shield'
    OR item_id LIKE 'shield_%'
  );
