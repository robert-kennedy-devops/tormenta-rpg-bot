package database

// =============================================
// IMAGE CACHE (implements assets.FileIDStore)
// =============================================

// ImageCacheStore implements assets.FileIDStore using PostgreSQL
type ImageCacheStore struct{}

func (s *ImageCacheStore) SaveFileID(key, fileID string) error {
	_, err := DB.Exec(`
		INSERT INTO image_cache (key, file_id)
		VALUES ($1, $2)
		ON CONFLICT (key) DO UPDATE SET file_id = $2, updated_at = NOW()
	`, key, fileID)
	return err
}

func (s *ImageCacheStore) LoadFileID(key string) (string, bool) {
	var fileID string
	err := DB.QueryRow(`SELECT file_id FROM image_cache WHERE key = $1`, key).Scan(&fileID)
	if err != nil {
		return "", false
	}
	return fileID, true
}

func (s *ImageCacheStore) LoadAll() map[string]string {
	rows, err := DB.Query(`SELECT key, file_id FROM image_cache`)
	if err != nil {
		return nil
	}
	defer rows.Close()

	result := make(map[string]string)
	for rows.Next() {
		var key, fileID string
		if err := rows.Scan(&key, &fileID); err == nil {
			result[key] = fileID
		}
	}
	_ = rows.Err()
	return result
}
