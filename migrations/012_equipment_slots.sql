-- ============================================================
-- MIGRATION 012: Equipment Slot System
-- Adiciona coluna `slot` na tabela inventory para suportar
-- múltiplos slots de equipamento por personagem.
-- ============================================================

-- Adiciona coluna slot (nullable para itens não equipáveis)
ALTER TABLE inventory ADD COLUMN IF NOT EXISTS slot VARCHAR(20) DEFAULT NULL;

-- Índice para busca rápida por personagem + slot
CREATE INDEX IF NOT EXISTS idx_inventory_slot ON inventory(character_id, slot);

-- Comentário dos slots válidos:
-- weapon, head, chest, hands, legs, feet, accessory1, accessory2
