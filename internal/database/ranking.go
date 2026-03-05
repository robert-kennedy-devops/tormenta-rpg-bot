package database

import "database/sql"

// GetRanking returns the global leaderboard (top N)
func GetRanking(limit int) ([]RankEntry, error) {
	rows, err := DB.Query(`
		SELECT c.id, c.name, c.race, c.class, c.level, c.experience,
		       COALESCE(p.rating, 1000) as pvp_rating,
		       COALESCE(p.wins, 0) as pvp_wins,
		       COALESCE((SELECT SUM(completions) FROM dungeon_best WHERE character_id=c.id), 0) as dungeon_completions,
		       (c.level * 1000 + c.experience + COALESCE(p.rating,1000) + COALESCE(p.wins,0)*50) AS score
		FROM characters c
		LEFT JOIN pvp_stats p ON p.character_id = c.id
		ORDER BY score DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var entries []RankEntry
	pos := 1
	for rows.Next() {
		var e RankEntry
		if err := rows.Scan(&e.CharID, &e.Name, &e.Race, &e.Class, &e.Level, &e.Experience,
			&e.PVPRating, &e.PVPWins, &e.DungeonCompletions, &e.Score); err != nil {
			return nil, err
		}
		e.Position = pos
		pos++
		entries = append(entries, e)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return entries, nil
}

type RankEntry struct {
	Position           int
	CharID             int
	Name               string
	Race               string
	Class              string
	Level              int
	Experience         int
	PVPRating          int
	PVPWins            int
	DungeonCompletions int
	Score              int
}

func GetPlayerRank(charID int) (int, int, error) {
	var position, score int
	err := DB.QueryRow(`
		SELECT position, score FROM (
			SELECT c.id,
			       (c.level * 1000 + c.experience + COALESCE(p.rating,1000) + COALESCE(p.wins,0)*50) AS score,
			       ROW_NUMBER() OVER (ORDER BY (c.level * 1000 + c.experience + COALESCE(p.rating,1000) + COALESCE(p.wins,0)*50) DESC) as position
			FROM characters c
			LEFT JOIN pvp_stats p ON p.character_id = c.id
		) ranked WHERE id=$1
	`, charID).Scan(&position, &score)
	if err == sql.ErrNoRows {
		return 0, 0, nil
	}
	return position, score, err
}
