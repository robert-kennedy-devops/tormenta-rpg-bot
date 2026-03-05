package timers

import (
	"database/sql"
	"time"

	"github.com/tormenta-bot/internal/database"
)

func Set(playerID int64, key string, endTime int64) error {
	_, err := database.DB.Exec(`
		INSERT INTO player_timers (player_id, key, end_time)
		VALUES ($1, $2, $3)
		ON CONFLICT (player_id, key)
		DO UPDATE SET end_time=EXCLUDED.end_time, updated_at=NOW()
	`, playerID, key, endTime)
	return err
}

func SetAfter(playerID int64, key string, d time.Duration) error {
	return Set(playerID, key, time.Now().Add(d).Unix())
}

func Get(playerID int64, key string) (int64, bool, error) {
	var end int64
	err := database.DB.QueryRow(`
		SELECT end_time
		FROM player_timers
		WHERE player_id=$1 AND key=$2
	`, playerID, key).Scan(&end)
	if err == sql.ErrNoRows {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	return end, true, nil
}

func Clear(playerID int64, key string) error {
	_, err := database.DB.Exec(`
		DELETE FROM player_timers
		WHERE player_id=$1 AND key=$2
	`, playerID, key)
	return err
}

func Remaining(playerID int64, key string, now int64) (time.Duration, bool, error) {
	end, ok, err := Get(playerID, key)
	if err != nil || !ok {
		return 0, false, err
	}
	if now <= 0 {
		now = time.Now().Unix()
	}
	sec := end - now
	if sec <= 0 {
		return 0, false, nil
	}
	return time.Duration(sec) * time.Second, true, nil
}
