-- Migration 002: Image File ID Cache
-- Stores Telegram file_ids to avoid re-uploading images

CREATE TABLE IF NOT EXISTS image_cache (
    key VARCHAR(100) PRIMARY KEY,
    file_id VARCHAR(200) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TRIGGER update_image_cache_updated_at BEFORE UPDATE ON image_cache
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
