-- Migration 006: Mercado Pago PIX integration
-- Adds Mercado Pago payment ID, QR code data, and expiry to pix_payments

ALTER TABLE pix_payments
    ADD COLUMN IF NOT EXISTS mp_payment_id BIGINT        DEFAULT NULL,
    ADD COLUMN IF NOT EXISTS qr_code       TEXT          DEFAULT NULL,
    ADD COLUMN IF NOT EXISTS qr_code_b64   TEXT          DEFAULT NULL,
    ADD COLUMN IF NOT EXISTS expires_at    TIMESTAMPTZ   DEFAULT (NOW() + INTERVAL '30 minutes');

-- Fast lookup by MP payment ID (used by webhook + polling)
CREATE UNIQUE INDEX IF NOT EXISTS idx_pix_mp_id
    ON pix_payments(mp_payment_id)
    WHERE mp_payment_id IS NOT NULL;

-- Back-fill expires_at for any existing rows that have NULL
UPDATE pix_payments
    SET expires_at = created_at + INTERVAL '30 minutes'
    WHERE expires_at IS NULL;
