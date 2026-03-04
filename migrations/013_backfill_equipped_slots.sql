-- ============================================================
-- MIGRATION 013: Backfill equipped inventory slots
-- Normaliza `inventory.slot` para itens equipados legados.
-- ============================================================

-- Garante existência da coluna em bases antigas.
ALTER TABLE inventory ADD COLUMN IF NOT EXISTS slot VARCHAR(20) DEFAULT NULL;

-- Equipamentos antigos sem slot explícito.
UPDATE inventory
SET slot = 'weapon'
WHERE equipped = true
  AND COALESCE(slot, '') = ''
  AND item_type = 'weapon';

-- Armaduras antigas (sistema pré-slots) caem no slot de peito por padrão.
UPDATE inventory
SET slot = 'chest'
WHERE equipped = true
  AND COALESCE(slot, '') = ''
  AND item_type = 'armor';

-- Acessórios legados: prioriza colar/amuleto/pingente como accessory2,
-- restante como accessory1.
UPDATE inventory
SET slot = CASE
    WHEN item_id LIKE 'necklace_%'
      OR item_id LIKE 'amulet_%'
      OR item_id LIKE 'pendant_%' THEN 'accessory2'
    ELSE 'accessory1'
END
WHERE equipped = true
  AND COALESCE(slot, '') = ''
  AND item_type = 'accessory';
