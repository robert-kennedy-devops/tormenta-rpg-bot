package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/tormenta-bot/internal/models"
)

var jsonMarshal = json.Marshal
var jsonUnmarshal = json.Unmarshal

var DB *sql.DB

func Connect() error {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://tormenta:tormenta123@localhost:5432/tormenta_rpg?sslmode=disable"
	}
	var err error
	for i := 0; i < 10; i++ {
		DB, err = sql.Open("postgres", dsn)
		if err != nil {
			log.Printf("DB open error: %v, retry %d/10", err, i+1)
			time.Sleep(2 * time.Second)
			continue
		}
		if err = DB.Ping(); err != nil {
			log.Printf("DB ping error: %v, retry %d/10", err, i+1)
			_ = DB.Close()
			time.Sleep(2 * time.Second)
			continue
		}
		break
	}
	if err != nil {
		if DB != nil {
			_ = DB.Close()
		}
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(10)
	DB.SetConnMaxLifetime(5 * time.Minute)
	log.Println("✅ Database connected!")
	return nil
}

func Migrate() {
	if DB == nil {
		log.Println("⚠️  Migrate() chamado sem DB inicializado")
		return
	}
	stmts := []string{
		`ALTER TABLE characters ADD COLUMN IF NOT EXISTS poison_turns int NOT NULL DEFAULT 0`,
		`ALTER TABLE characters ADD COLUMN IF NOT EXISTS poison_dmg int NOT NULL DEFAULT 0`,
		`ALTER TABLE characters ADD COLUMN IF NOT EXISTS combat_monster_poison_turns int NOT NULL DEFAULT 0`,
		`ALTER TABLE characters ADD COLUMN IF NOT EXISTS combat_monster_poison_dmg int NOT NULL DEFAULT 0`,
		`ALTER TABLE inventory ADD COLUMN IF NOT EXISTS slot VARCHAR(20) DEFAULT NULL`,
		`ALTER TABLE characters ADD COLUMN IF NOT EXISTS last_energy_update BIGINT`,
		`UPDATE characters
		  SET last_energy_update = EXTRACT(EPOCH FROM COALESCE(energy_regen_at, NOW()))::BIGINT
		  WHERE last_energy_update IS NULL OR last_energy_update <= 0`,
		`UPDATE inventory
		  SET slot='weapon'
		  WHERE equipped=true AND COALESCE(slot,'')='' AND item_type='weapon'`,
		`UPDATE inventory
		  SET slot='chest'
		  WHERE equipped=true AND COALESCE(slot,'')='' AND item_type='armor'`,
		`UPDATE inventory
		  SET slot=CASE
		    WHEN item_id LIKE 'necklace_%'
		      OR item_id LIKE 'amulet_%'
		      OR item_id LIKE 'pendant_%' THEN 'accessory2'
		    ELSE 'accessory1'
		  END
		  WHERE equipped=true AND COALESCE(slot,'')='' AND item_type='accessory'`,
		`UPDATE inventory
		  SET slot='offhand'
		  WHERE equipped=true AND (item_id='iron_shield' OR item_id LIKE 'shield_%')`,
		`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`,
		`CREATE TABLE IF NOT EXISTS player_items (
			instance_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			character_id INT NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
			template_id VARCHAR(64) NOT NULL,
			upgrade_level INT NOT NULL DEFAULT 0,
			broken BOOLEAN NOT NULL DEFAULT FALSE,
			equipped BOOLEAN NOT NULL DEFAULT FALSE,
			equipped_slot VARCHAR(20) DEFAULT NULL,
			rarity_override VARCHAR(20) DEFAULT NULL,
			stat_overrides JSONB DEFAULT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_player_items_character ON player_items(character_id)`,
		`CREATE INDEX IF NOT EXISTS idx_player_items_template ON player_items(character_id, template_id)`,
		`CREATE TABLE IF NOT EXISTS player_timers (
			player_id BIGINT NOT NULL,
			key VARCHAR(64) NOT NULL,
			end_time BIGINT NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
			PRIMARY KEY (player_id, key)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_player_timers_end_time ON player_timers(end_time)`,
		`CREATE TABLE IF NOT EXISTS item_usage_stats (
			item_id TEXT PRIMARY KEY,
			purchase_count BIGINT NOT NULL DEFAULT 0,
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_item_usage_updated_at ON item_usage_stats(updated_at)`,
		`CREATE INDEX IF NOT EXISTS idx_characters_updated_at ON characters(updated_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_characters_state ON characters(state)`,
		`CREATE INDEX IF NOT EXISTS idx_characters_level ON characters(level DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_inventory_char_item ON inventory(character_id, item_id)`,
		`CREATE INDEX IF NOT EXISTS idx_inventory_char_type ON inventory(character_id, item_type)`,
		`CREATE INDEX IF NOT EXISTS idx_player_items_char_slot ON player_items(character_id, equipped_slot, equipped)`,
		`CREATE INDEX IF NOT EXISTS idx_pix_status_created ON pix_payments(status, created_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_pix_expires_status ON pix_payments(status, expires_at)`,
		`CREATE INDEX IF NOT EXISTS idx_auto_hunt_char_status ON auto_hunt_sessions(character_id, status)`,
		`CREATE INDEX IF NOT EXISTS idx_auto_hunt_last_tick ON auto_hunt_sessions(last_tick_at)`,
		`CREATE INDEX IF NOT EXISTS idx_dungeon_runs_state_started ON dungeon_runs(state, started_at)`,
		`CREATE INDEX IF NOT EXISTS idx_characters_energy_tick ON characters(last_energy_update) WHERE energy < energy_max`,
		`CREATE TABLE IF NOT EXISTS analytics_events (
			id BIGSERIAL PRIMARY KEY,
			player_id BIGINT NOT NULL,
			character_id INT NOT NULL DEFAULT 0,
			event VARCHAR(64) NOT NULL,
			payload JSONB NOT NULL DEFAULT '{}'::jsonb,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_analytics_event_created ON analytics_events(event, created_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_analytics_player_created ON analytics_events(player_id, created_at DESC)`,
		`CREATE TABLE IF NOT EXISTS gm_action_logs (
			id BIGSERIAL PRIMARY KEY,
			gm_player_id BIGINT NOT NULL,
			target_character_id INT NOT NULL DEFAULT 0,
			action VARCHAR(64) NOT NULL,
			details TEXT NOT NULL DEFAULT '',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_gm_action_logs_gm_created ON gm_action_logs(gm_player_id, created_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_gm_action_logs_target_created ON gm_action_logs(target_character_id, created_at DESC)`,
	}
	for _, stmt := range stmts {
		if _, err := DB.Exec(stmt); err != nil {
			log.Printf("migration warning: %v", err)
		}
	}
	log.Println("✅ Migrations applied!")
}

// =============================================
// PLAYER
// =============================================

func UpsertPlayer(id int64, username string) error {
	_, err := DB.Exec(`
		INSERT INTO players (id, username) VALUES ($1, $2)
		ON CONFLICT (id) DO UPDATE SET username=$2, updated_at=NOW()
	`, id, username)
	return err
}

// GetPlayer returns a player with VIP info.
func GetPlayer(playerID int64) (*models.Player, error) {
	p := &models.Player{}
	err := DB.QueryRow(`
		SELECT id, username, created_at, is_vip, vip_expires_at
		FROM players WHERE id=$1
	`, playerID).Scan(&p.ID, &p.Username, &p.CreatedAt, &p.IsVIP, &p.VIPExpiresAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return p, err
}

// IsVIP returns true if the player currently has active VIP.
func IsVIP(playerID int64) bool {
	p, err := GetPlayer(playerID)
	if err != nil || p == nil {
		return false
	}
	return p.IsVIPActive()
}

// SetVIP sets VIP status for a player.
// duration=0 means permanent VIP; duration>0 adds that duration from now.
func SetVIP(playerID int64, active bool, duration time.Duration) error {
	var expiresAt *time.Time
	if active && duration > 0 {
		t := time.Now().Add(duration)
		expiresAt = &t
	}
	_, err := DB.Exec(`
		UPDATE players SET is_vip=$1, vip_expires_at=$2 WHERE id=$3
	`, active, expiresAt, playerID)
	return err
}

// ── AUTO HUNT ─────────────────────────────────────────────

// AutoHuntSkillConfig define como o personagem usa habilidades durante a caça automática.
// Mode: "attack" = só ataque normal | "skill" = rodízio pelas Skills | "smart" = habilidade se tiver MP, senão ataque
type AutoHuntSkillConfig struct {
	Mode    string   `json:"mode"`    // "attack" | "skill" | "smart"
	Skills  []string `json:"skills"`  // IDs das habilidades habilitadas
	Potions []string `json:"potions"` // IDs das poções de HP/MP habilitadas
	HealAt  int      `json:"heal_at"` // % de HP para usar poção de HP (ex: 50 = abaixo de 50%)
	ManaAt  int      `json:"mana_at"` // % de MP para usar poção de MP (ex: 30 = abaixo de 30%)
}

type AutoHuntSession struct {
	ID          int
	CharacterID int
	MapID       string
	StartedAt   time.Time
	LastTickAt  time.Time
	TicksDone   int
	TotalXP     int
	TotalGold   int
	TotalKills  int
	Status      string
	SkillConfig AutoHuntSkillConfig
}

func GetAutoHuntSession(charID int) (*AutoHuntSession, error) {
	s := &AutoHuntSession{}
	var skillConfigJSON string
	err := DB.QueryRow(`
		SELECT id, character_id, map_id, started_at, last_tick_at,
		       ticks_done, total_xp, total_gold, total_kills, status,
		       skill_config::text
		FROM auto_hunt_sessions WHERE character_id=$1
	`, charID).Scan(&s.ID, &s.CharacterID, &s.MapID, &s.StartedAt, &s.LastTickAt,
		&s.TicksDone, &s.TotalXP, &s.TotalGold, &s.TotalKills, &s.Status,
		&skillConfigJSON)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err == nil {
		parseSkillConfig(skillConfigJSON, &s.SkillConfig)
	}
	return s, err
}

func StartAutoHunt(charID int, mapID string, cfg AutoHuntSkillConfig) (*AutoHuntSession, error) {
	cfgJSON, _ := marshalSkillConfig(cfg)
	s := &AutoHuntSession{}
	var skillConfigJSON string
	err := DB.QueryRow(`
		INSERT INTO auto_hunt_sessions (character_id, map_id, skill_config)
		VALUES ($1, $2, $3::jsonb)
		ON CONFLICT (character_id) DO UPDATE
			SET map_id=$2, started_at=NOW(), last_tick_at=NOW(),
			    ticks_done=0, total_xp=0, total_gold=0, total_kills=0,
			    status='running', skill_config=$3::jsonb
		RETURNING id, character_id, map_id, started_at, last_tick_at,
		          ticks_done, total_xp, total_gold, total_kills, status,
		          skill_config::text
	`, charID, mapID, cfgJSON).Scan(&s.ID, &s.CharacterID, &s.MapID, &s.StartedAt, &s.LastTickAt,
		&s.TicksDone, &s.TotalXP, &s.TotalGold, &s.TotalKills, &s.Status,
		&skillConfigJSON)
	if err == nil {
		parseSkillConfig(skillConfigJSON, &s.SkillConfig)
	}
	return s, err
}

func UpdateAutoHuntTick(sessionID, xpGain, goldGain int) error {
	_, err := DB.Exec(`
		UPDATE auto_hunt_sessions
		SET last_tick_at=NOW(), ticks_done=ticks_done+1,
		    total_xp=total_xp+$1, total_gold=total_gold+$2, total_kills=total_kills+1
		WHERE id=$3
	`, xpGain, goldGain, sessionID)
	return err
}

func StopAutoHunt(charID int, reason string) error {
	_, err := DB.Exec(`
		UPDATE auto_hunt_sessions SET status=$1 WHERE character_id=$2
	`, reason, charID)
	return err
}

func GetRunningAutoHunts() ([]AutoHuntSession, error) {
	rows, err := DB.Query(`
		SELECT id, character_id, map_id, started_at, last_tick_at,
		       ticks_done, total_xp, total_gold, total_kills, status,
		       skill_config::text
		FROM auto_hunt_sessions WHERE status='running'
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var sessions []AutoHuntSession
	for rows.Next() {
		var s AutoHuntSession
		var skillConfigJSON string
		if err := rows.Scan(&s.ID, &s.CharacterID, &s.MapID, &s.StartedAt, &s.LastTickAt,
			&s.TicksDone, &s.TotalXP, &s.TotalGold, &s.TotalKills, &s.Status,
			&skillConfigJSON); err != nil {
			return nil, err
		}
		parseSkillConfig(skillConfigJSON, &s.SkillConfig)
		sessions = append(sessions, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return sessions, nil
}

// ── Skill config JSON helpers ─────────────────────────────

func parseSkillConfig(raw string, cfg *AutoHuntSkillConfig) {
	if raw == "" {
		cfg.Mode = "attack"
		cfg.Skills = []string{}
		return
	}
	// Simple manual JSON parse to avoid importing encoding/json at package level
	// We use the existing import via marshalSkillConfig
	if err := jsonUnmarshal([]byte(raw), cfg); err != nil {
		cfg.Mode = "attack"
		cfg.Skills = []string{}
	}
}

func marshalSkillConfig(cfg AutoHuntSkillConfig) (string, error) {
	b, err := jsonMarshal(cfg)
	return string(b), err
}

// =============================================
// CHARACTER — full read / write
// =============================================

func GetCharacter(playerID int64) (*models.Character, error) {
	char := &models.Character{}
	err := DB.QueryRow(`
		SELECT id, player_id, name, race, class, level, experience, experience_next,
			hp, hp_max, mp, mp_max,
			COALESCE(energy, 100), COALESCE(energy_max, 100),
			COALESCE(energy_regen_at, NOW()),
			COALESCE(last_energy_update, EXTRACT(EPOCH FROM COALESCE(energy_regen_at, NOW()))::BIGINT),
			COALESCE(diamonds, 0),
			strength, dexterity, constitution, intelligence,
			wisdom, charisma, attack, defense, magic_attack, magic_defense, speed,
			gold, current_map, state,
			COALESCE(combat_monster_id,''), combat_monster_hp, skill_points,
			COALESCE(deaths, 0),
			COALESCE(xp_boost_expiry, '1970-01-01 00:00:00+00'),
			COALESCE(poison_turns, 0), COALESCE(poison_dmg, 0),
			COALESCE(combat_monster_poison_turns, 0), COALESCE(combat_monster_poison_dmg, 0)
		FROM characters WHERE player_id=$1
	`, playerID).Scan(
		&char.ID, &char.PlayerID, &char.Name, &char.Race, &char.Class,
		&char.Level, &char.Experience, &char.ExperienceNext,
		&char.HP, &char.HPMax, &char.MP, &char.MPMax,
		&char.Energy, &char.EnergyMax, &char.EnergyRegenAt, &char.LastEnergyUpdate, &char.Diamonds,
		&char.Strength, &char.Dexterity, &char.Constitution, &char.Intelligence,
		&char.Wisdom, &char.Charisma,
		&char.Attack, &char.Defense, &char.MagicAttack, &char.MagicDefense, &char.Speed,
		&char.Gold, &char.CurrentMap, &char.State,
		&char.CombatMonsterID, &char.CombatMonsterHP, &char.SkillPoints,
		&char.Deaths, &char.XPBoostExpiry,
		&char.PoisonTurns, &char.PoisonDmg,
		&char.CombatMonsterPoisonTurns, &char.CombatMonsterPoisonDmg,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return char, err
}

func CreateCharacter(char *models.Character) error {
	if char.LastEnergyUpdate <= 0 {
		char.LastEnergyUpdate = time.Now().Unix()
	}
	return DB.QueryRow(`
		INSERT INTO characters (
			player_id, name, race, class, level, experience, experience_next,
			hp, hp_max, mp, mp_max,
			energy, energy_max, energy_regen_at, last_energy_update, diamonds,
			strength, dexterity, constitution, intelligence,
			wisdom, charisma, attack, defense, magic_attack, magic_defense, speed,
			gold, current_map, state
		) VALUES (
			$1,$2,$3,$4,$5,$6,$7,
			$8,$9,$10,$11,
			$12,$13,$14,$15,$16,
			$17,$18,$19,$20,
			$21,$22,$23,$24,$25,$26,$27,
			$28,$29,$30
		) RETURNING id
	`,
		char.PlayerID, char.Name, char.Race, char.Class,
		char.Level, char.Experience, char.ExperienceNext,
		char.HP, char.HPMax, char.MP, char.MPMax,
		char.Energy, char.EnergyMax, char.EnergyRegenAt, char.LastEnergyUpdate, char.Diamonds,
		char.Strength, char.Dexterity, char.Constitution, char.Intelligence,
		char.Wisdom, char.Charisma,
		char.Attack, char.Defense, char.MagicAttack, char.MagicDefense, char.Speed,
		char.Gold, char.CurrentMap, char.State,
	).Scan(&char.ID)
}

func SaveCharacter(char *models.Character) error {
	// Garante que energy nunca ultrapasse energy_max antes de salvar
	if char.Energy > char.EnergyMax {
		char.Energy = char.EnergyMax
	}
	if char.Energy < 0 {
		char.Energy = 0
	}
	if char.LastEnergyUpdate <= 0 {
		char.LastEnergyUpdate = time.Now().Unix()
	}
	_, err := DB.Exec(`
		UPDATE characters SET
			level=$1, experience=$2, experience_next=$3,
			hp=$4, hp_max=$5, mp=$6, mp_max=$7,
			energy=$8, energy_max=$9, energy_regen_at=$10, last_energy_update=$11, diamonds=$12,
			strength=$13, dexterity=$14, constitution=$15, intelligence=$16,
			wisdom=$17, charisma=$18, attack=$19, defense=$20,
			magic_attack=$21, magic_defense=$22, speed=$23,
			gold=$24, current_map=$25, state=$26,
			combat_monster_id=$27, combat_monster_hp=$28, skill_points=$29,
			deaths=$30, xp_boost_expiry=$31,
			poison_turns=$32, poison_dmg=$33,
			combat_monster_poison_turns=$34, combat_monster_poison_dmg=$35,
			updated_at=NOW()
		WHERE id=$36
	`,
		char.Level, char.Experience, char.ExperienceNext,
		char.HP, char.HPMax, char.MP, char.MPMax,
		char.Energy, char.EnergyMax, char.EnergyRegenAt, char.LastEnergyUpdate, char.Diamonds,
		char.Strength, char.Dexterity, char.Constitution, char.Intelligence,
		char.Wisdom, char.Charisma,
		char.Attack, char.Defense, char.MagicAttack, char.MagicDefense, char.Speed,
		char.Gold, char.CurrentMap, char.State,
		char.CombatMonsterID, char.CombatMonsterHP, char.SkillPoints,
		char.Deaths, char.XPBoostExpiry,
		char.PoisonTurns, char.PoisonDmg,
		char.CombatMonsterPoisonTurns, char.CombatMonsterPoisonDmg,
		char.ID,
	)
	return err
}

// SaveCharacterEnergy updates only energy-related fields.
func SaveCharacterEnergy(charID int, energy int, energyMax int, regenAt time.Time) error {
	if energy > energyMax {
		energy = energyMax
	}
	if energy < 0 {
		energy = 0
	}
	_, err := DB.Exec(`
		UPDATE characters
		SET energy=$1, energy_max=$2, energy_regen_at=$3,
		    last_energy_update=EXTRACT(EPOCH FROM $3)::BIGINT,
		    updated_at=NOW()
		WHERE id=$4
	`, energy, energyMax, regenAt, charID)
	return err
}

type EnergyTickCandidate struct {
	CharacterID      int
	Energy           int
	EnergyMax        int
	LastEnergyUpdate int64
	IsVIP            bool
}

// GetEnergyTickCandidates returns characters that may recover energy.
// VIP state is resolved from players so interval/cap can be applied without loading full character rows.
func GetEnergyTickCandidates(limit int) ([]EnergyTickCandidate, error) {
	if limit <= 0 {
		limit = 500
	}
	rows, err := DB.Query(`
		SELECT c.id, COALESCE(c.energy, 0), COALESCE(c.energy_max, 100),
		       COALESCE(c.last_energy_update, EXTRACT(EPOCH FROM NOW())::BIGINT),
		       COALESCE(p.is_vip, FALSE),
		       p.vip_expires_at
		FROM characters c
		LEFT JOIN players p ON p.id = c.player_id
		WHERE COALESCE(c.energy, 0) < COALESCE(c.energy_max, 100)
		ORDER BY c.last_energy_update ASC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	now := time.Now()
	out := make([]EnergyTickCandidate, 0, limit)
	for rows.Next() {
		var c EnergyTickCandidate
		var vipExpiresAt sql.NullTime
		if err := rows.Scan(&c.CharacterID, &c.Energy, &c.EnergyMax, &c.LastEnergyUpdate, &c.IsVIP, &vipExpiresAt); err != nil {
			return nil, err
		}
		// Keep compatibility with existing VIP semantics: active if flag=true and not expired.
		if c.IsVIP && vipExpiresAt.Valid && vipExpiresAt.Time.Before(now) {
			c.IsVIP = false
		}
		out = append(out, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// GetCharactersNeedingEnergyTick returns characters that are not at full energy.
func GetCharactersNeedingEnergyTick(limit int) ([]models.Character, error) {
	rows, err := DB.Query(`
		SELECT id, energy, energy_max, COALESCE(energy_regen_at, NOW())
		FROM characters
		WHERE energy < energy_max
		ORDER BY energy_regen_at ASC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chars []models.Character
	for rows.Next() {
		var c models.Character
		if err := rows.Scan(&c.ID, &c.Energy, &c.EnergyMax, &c.EnergyRegenAt); err != nil {
			return nil, err
		}
		chars = append(chars, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return chars, nil
}

func DeleteCharacter(playerID int64) error {
	_, err := DB.Exec("DELETE FROM characters WHERE player_id=$1", playerID)
	return err
}

// =============================================
// SKILLS
// =============================================

func GetLearnedSkills(charID int) ([]models.CharacterSkill, error) {
	rows, err := DB.Query(`SELECT id, character_id, skill_id, level FROM character_skills WHERE character_id=$1`, charID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var skills []models.CharacterSkill
	for rows.Next() {
		var s models.CharacterSkill
		if err := rows.Scan(&s.ID, &s.CharacterID, &s.SkillID, &s.Level); err != nil {
			return nil, err
		}
		skills = append(skills, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return skills, nil
}

func HasSkill(charID int, skillID string) bool {
	var count int
	if err := DB.QueryRow(`SELECT COUNT(*) FROM character_skills WHERE character_id=$1 AND skill_id=$2`, charID, skillID).Scan(&count); err != nil {
		return false
	}
	return count > 0
}

func LearnSkill(charID int, skillID string) error {
	_, err := DB.Exec(`
		INSERT INTO character_skills (character_id, skill_id) VALUES ($1,$2)
		ON CONFLICT (character_id, skill_id) DO NOTHING
	`, charID, skillID)
	return err
}

// ResetSkills apaga todas as habilidades aprendidas e retorna os IDs
// das skills deletadas para o caller calcular os pontos exatos via PointCost.
func ResetSkills(charID int) (skillIDs []string, err error) {
	rows, queryErr := DB.Query(`SELECT skill_id FROM character_skills WHERE character_id=$1`, charID)
	if queryErr != nil {
		return nil, queryErr
	}
	defer rows.Close()
	for rows.Next() {
		var sid string
		if err := rows.Scan(&sid); err != nil {
			return nil, err
		}
		skillIDs = append(skillIDs, sid)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	_, err = DB.Exec(`DELETE FROM character_skills WHERE character_id=$1`, charID)
	return skillIDs, err
}

// =============================================
// INVENTORY
// =============================================

func isEquipableItemType(itemType string) bool {
	switch itemType {
	case "weapon", "armor", "accessory":
		return true
	default:
		return false
	}
}

func GetInventory(charID int) ([]models.InventoryItem, error) {
	rows, err := DB.Query(`
		SELECT id, character_id, item_id, item_type, quantity, equipped, COALESCE(slot,'')
		FROM inventory WHERE character_id=$1 ORDER BY equipped DESC, item_type, item_id
	`, charID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []models.InventoryItem
	for rows.Next() {
		var item models.InventoryItem
		if err := rows.Scan(&item.ID, &item.CharacterID, &item.ItemID, &item.ItemType, &item.Quantity, &item.Equipped, &item.Slot); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func AddItem(charID int, itemID, itemType string, qty int) error {
	_, err := DB.Exec(`
		INSERT INTO inventory (character_id, item_id, item_type, quantity)
		VALUES ($1,$2,$3,$4)
		ON CONFLICT (character_id, item_id) DO UPDATE SET quantity = inventory.quantity + $4
	`, charID, itemID, itemType, qty)
	if err != nil {
		return err
	}
	if isEquipableItemType(itemType) {
		if err := CreatePlayerItemInstances(charID, itemID, qty); err != nil {
			return err
		}
	}
	return nil
}

func RemoveItem(charID int, itemID string, qty int) error {
	var current int
	if err := DB.QueryRow(`SELECT quantity FROM inventory WHERE character_id=$1 AND item_id=$2`, charID, itemID).Scan(&current); err != nil {
		if err == sql.ErrNoRows {
			current = 0
		} else {
			return err
		}
	}
	if current <= qty {
		_, err := DB.Exec(`DELETE FROM inventory WHERE character_id=$1 AND item_id=$2`, charID, itemID)
		if err != nil {
			return err
		}
		for i := 0; i < current; i++ {
			ok, _ := DeleteOnePlayerItemInstance(charID, itemID, true)
			if !ok {
				_, _ = DeleteOnePlayerItemInstance(charID, itemID, false)
			}
		}
		return nil
	}
	_, err := DB.Exec(`UPDATE inventory SET quantity=quantity-$3 WHERE character_id=$1 AND item_id=$2`, charID, itemID, qty)
	if err != nil {
		return err
	}
	for i := 0; i < qty; i++ {
		ok, delErr := DeleteOnePlayerItemInstance(charID, itemID, true)
		if delErr != nil {
			return delErr
		}
		if !ok {
			_, _ = DeleteOnePlayerItemInstance(charID, itemID, false)
		}
	}
	_ = SyncInventoryEquippedFromInstances(charID, itemID)
	return nil
}

func GetItemCount(charID int, itemID string) int {
	var count int
	if err := DB.QueryRow(`SELECT COALESCE(quantity,0) FROM inventory WHERE character_id=$1 AND item_id=$2`, charID, itemID).Scan(&count); err != nil {
		return 0
	}
	return count
}

func EquipItem(charID int, itemID string, itemType string) error {
	return EquipItemSlot(charID, itemID, itemType, "")
}

// EquipItemSlot equips an item into a specific slot, unequipping any previous item in that slot.
func EquipItemSlot(charID int, itemID string, itemType string, slot string) error {
	if slot != "" {
		// Unequip whatever is in this slot
		if _, err := DB.Exec(`UPDATE inventory SET equipped=false, slot=NULL WHERE character_id=$1 AND slot=$2`, charID, slot); err != nil {
			return err
		}
		_ = UnequipPlayerItemsBySlot(charID, slot)
	} else {
		// Legacy: unequip by type (for weapon only fallback)
		if _, err := DB.Exec(`UPDATE inventory SET equipped=false WHERE character_id=$1 AND item_type=$2`, charID, itemType); err != nil {
			return err
		}
		_ = UnequipPlayerItemsByTemplate(charID, itemID)
	}
	var err error
	if slot != "" {
		_, err = DB.Exec(`UPDATE inventory SET equipped=true, slot=$3 WHERE character_id=$1 AND item_id=$2`, charID, itemID, slot)
	} else {
		_, err = DB.Exec(`UPDATE inventory SET equipped=true WHERE character_id=$1 AND item_id=$2`, charID, itemID)
	}
	if err != nil {
		return err
	}
	if isEquipableItemType(itemType) {
		if eErr := EquipPlayerItemByTemplateAndSlot(charID, itemID, slot); eErr != nil {
			return eErr
		}
	}
	return nil
}

// UnequipSlot removes the equipped item from a specific slot.
func UnequipSlot(charID int, slot string) error {
	_, err := DB.Exec(`UPDATE inventory SET equipped=false, slot=NULL WHERE character_id=$1 AND slot=$2`, charID, slot)
	if err != nil {
		return err
	}
	return UnequipPlayerItemsBySlot(charID, slot)
}

func UnequipItem(charID int, itemID string) error {
	_, err := DB.Exec(`UPDATE inventory SET equipped=false WHERE character_id=$1 AND item_id=$2`, charID, itemID)
	if err != nil {
		return err
	}
	return UnequipPlayerItemsByTemplate(charID, itemID)
}

func GetEquippedItems(charID int) ([]models.InventoryItem, error) {
	rows, err := DB.Query(`
		SELECT id, character_id, item_id, item_type, quantity, equipped, COALESCE(slot,'')
		FROM inventory WHERE character_id=$1 AND equipped=true
	`, charID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []models.InventoryItem
	for rows.Next() {
		var item models.InventoryItem
		if err := rows.Scan(&item.ID, &item.CharacterID, &item.ItemID, &item.ItemType, &item.Quantity, &item.Equipped, &item.Slot); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

// =============================================
// COMBAT LOG
// =============================================

func LogCombat(charID int, monsterID, result string, exp, gold int) error {
	_, err := DB.Exec(`
		INSERT INTO combat_log (character_id, monster_id, result, exp_gained, gold_gained)
		VALUES ($1,$2,$3,$4,$5)
	`, charID, monsterID, result, exp, gold)
	return err
}

// =============================================
// DIAMOND LOG
// =============================================

func LogDiamond(charID int, amount int, reason string) error {
	_, err := DB.Exec(`
		INSERT INTO diamond_log (character_id, amount, reason) VALUES ($1,$2,$3)
	`, charID, amount, reason)
	return err
}

// =============================================
// DAILY BONUS
// =============================================

// ClaimDailyBonus returns (diamonds, ok). ok=false if already claimed today.
func ClaimDailyBonus(charID int) (int, bool) {
	var lastClaim time.Time
	err := DB.QueryRow(`SELECT last_claim FROM daily_bonus WHERE character_id=$1`, charID).Scan(&lastClaim)
	if err == sql.ErrNoRows {
		lastClaim = time.Time{} // never claimed
	} else if err != nil {
		// Falha de leitura no banco: não concede bônus para evitar inconsistências.
		return 0, false
	}

	now := time.Now()
	y1, m1, d1 := lastClaim.Date()
	y2, m2, d2 := now.Date()
	if y1 == y2 && m1 == m2 && d1 == d2 {
		return 0, false // already claimed
	}

	diamonds := 3
	if _, err := DB.Exec(`
		INSERT INTO daily_bonus (character_id, last_claim) VALUES ($1, $2)
		ON CONFLICT (character_id) DO UPDATE SET last_claim=$2
	`, charID, now); err != nil {
		return 0, false
	}
	return diamonds, true
}

// =============================================
// IMAGE CACHE
// =============================================

type ImageCache struct{}

func (c ImageCache) SaveFileID(key, fileID string) error {
	_, err := DB.Exec(`
		INSERT INTO image_cache (key, file_id) VALUES ($1,$2)
		ON CONFLICT (key) DO UPDATE SET file_id=$2, updated_at=NOW()
	`, key, fileID)
	return err
}

func (c ImageCache) LoadFileID(key string) (string, bool) {
	var fileID string
	err := DB.QueryRow(`SELECT file_id FROM image_cache WHERE key=$1`, key).Scan(&fileID)
	if err != nil {
		return "", false
	}
	return fileID, true
}

func (c ImageCache) LoadAll() map[string]string {
	rows, err := DB.Query(`SELECT key, file_id FROM image_cache`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	result := map[string]string{}
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			continue
		}
		result[k] = v
	}
	_ = rows.Err()
	return result
}

// =============================================
// DUNGEON OPERATIONS
// =============================================

type DungeonRun struct {
	ID             int
	CharacterID    int
	DungeonID      string
	Floor          int
	State          string
	MonstersKilled int
}

func GetActiveDungeonRun(charID int) (*DungeonRun, error) {
	run := &DungeonRun{}
	err := DB.QueryRow(`
		SELECT id, character_id, dungeon_id, floor, state, monsters_killed
		FROM dungeon_runs WHERE character_id=$1 AND state='active'
		ORDER BY started_at DESC LIMIT 1
	`, charID).Scan(&run.ID, &run.CharacterID, &run.DungeonID, &run.Floor, &run.State, &run.MonstersKilled)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return run, err
}

func CreateDungeonRun(charID int, dungeonID string) (*DungeonRun, error) {
	run := &DungeonRun{CharacterID: charID, DungeonID: dungeonID, Floor: 1, State: "active"}
	err := DB.QueryRow(`
		INSERT INTO dungeon_runs (character_id, dungeon_id, floor, state)
		VALUES ($1,$2,$3,'active') RETURNING id
	`, charID, dungeonID, 1).Scan(&run.ID)
	return run, err
}

func AdvanceDungeonFloor(runID, newFloor, monstersKilled int) error {
	_, err := DB.Exec(`
		UPDATE dungeon_runs SET floor=$1, monsters_killed=monsters_killed+$2 WHERE id=$3
	`, newFloor, monstersKilled, runID)
	return err
}

func FinishDungeonRun(runID int, state string) error {
	_, err := DB.Exec(`
		UPDATE dungeon_runs SET state=$1, finished_at=NOW() WHERE id=$2
	`, state, runID)
	return err
}

func UpdateDungeonBest(charID int, dungeonID string, floor int, completed bool) error {
	completions := 0
	if completed {
		completions = 1
	}
	_, err := DB.Exec(`
		INSERT INTO dungeon_best (character_id, dungeon_id, best_floor, completions, updated_at)
		VALUES ($1,$2,$3,$4,NOW())
		ON CONFLICT (character_id, dungeon_id) DO UPDATE
		  SET best_floor=GREATEST(dungeon_best.best_floor,$3),
		      completions=dungeon_best.completions+$4,
		      updated_at=NOW()
	`, charID, dungeonID, floor, completions)
	return err
}

func GetDungeonBest(charID int, dungeonID string) (bestFloor, completions int) {
	DB.QueryRow(`SELECT best_floor, completions FROM dungeon_best WHERE character_id=$1 AND dungeon_id=$2`,
		charID, dungeonID).Scan(&bestFloor, &completions)
	return
}

// =============================================
// PVP OPERATIONS
// =============================================

type PVPChallenge struct {
	ID           int
	ChallengerID int
	DefenderID   int
	StakeGold    int
	State        string
	Turn         int
	ChallengerHP int
	DefenderHP   int
	WinnerID     int
}

type PVPStats struct {
	CharacterID int
	Wins        int
	Losses      int
	Draws       int
	Rating      int
	Streak      int
	BestStreak  int
}

func CreatePVPChallenge(challengerID, defenderID, stakeGold, challengerHP, defenderHP int) (*PVPChallenge, error) {
	ch := &PVPChallenge{
		ChallengerID: challengerID, DefenderID: defenderID,
		StakeGold: stakeGold, State: "pending",
		ChallengerHP: challengerHP, DefenderHP: defenderHP,
	}
	err := DB.QueryRow(`
		INSERT INTO pvp_challenges (challenger_id, defender_id, stake_gold, state, challenger_hp, defender_hp)
		VALUES ($1,$2,$3,'pending',$4,$5) RETURNING id
	`, challengerID, defenderID, stakeGold, challengerHP, defenderHP).Scan(&ch.ID)
	return ch, err
}

func GetPendingChallenge(defenderID int) (*PVPChallenge, error) {
	ch := &PVPChallenge{}
	err := DB.QueryRow(`
		SELECT id, challenger_id, defender_id, stake_gold, state, turn, challenger_hp, defender_hp, COALESCE(winner_id,0)
		FROM pvp_challenges
		WHERE defender_id=$1 AND state='pending' AND expires_at > NOW()
		ORDER BY created_at DESC LIMIT 1
	`, defenderID).Scan(&ch.ID, &ch.ChallengerID, &ch.DefenderID, &ch.StakeGold, &ch.State, &ch.Turn, &ch.ChallengerHP, &ch.DefenderHP, &ch.WinnerID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return ch, err
}

func GetActivePVPMatch(charID int) (*PVPChallenge, error) {
	ch := &PVPChallenge{}
	err := DB.QueryRow(`
		SELECT id, challenger_id, defender_id, stake_gold, state, turn, challenger_hp, defender_hp, COALESCE(winner_id,0)
		FROM pvp_challenges
		WHERE (challenger_id=$1 OR defender_id=$1) AND state='active'
		ORDER BY created_at DESC LIMIT 1
	`, charID).Scan(&ch.ID, &ch.ChallengerID, &ch.DefenderID, &ch.StakeGold, &ch.State, &ch.Turn, &ch.ChallengerHP, &ch.DefenderHP, &ch.WinnerID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return ch, err
}

func AcceptPVPChallenge(matchID int) error {
	_, err := DB.Exec(`UPDATE pvp_challenges SET state='active' WHERE id=$1`, matchID)
	return err
}

func DeclinePVPChallenge(matchID int) error {
	_, err := DB.Exec(`UPDATE pvp_challenges SET state='declined' WHERE id=$1`, matchID)
	return err
}

func UpdatePVPMatch(matchID, turn, challengerHP, defenderHP int) error {
	_, err := DB.Exec(`
		UPDATE pvp_challenges SET turn=$1, challenger_hp=$2, defender_hp=$3 WHERE id=$4
	`, turn, challengerHP, defenderHP, matchID)
	return err
}

func FinishPVPMatch(matchID, winnerID int) error {
	_, err := DB.Exec(`
		UPDATE pvp_challenges SET state='finished', winner_id=$1 WHERE id=$2
	`, winnerID, matchID)
	return err
}

func GetOrCreatePVPStats(charID int) (*PVPStats, error) {
	stats := &PVPStats{CharacterID: charID}
	err := DB.QueryRow(`
		SELECT wins, losses, draws, rating, streak, best_streak
		FROM pvp_stats WHERE character_id=$1
	`, charID).Scan(&stats.Wins, &stats.Losses, &stats.Draws, &stats.Rating, &stats.Streak, &stats.BestStreak)
	if err == sql.ErrNoRows {
		// Create default stats
		if _, execErr := DB.Exec(`INSERT INTO pvp_stats (character_id) VALUES ($1) ON CONFLICT DO NOTHING`, charID); execErr != nil {
			return nil, execErr
		}
		stats.Rating = 1000
		return stats, nil
	}
	return stats, err
}

func UpdatePVPStats(charID, wins, losses, draws, rating, streak, bestStreak int) error {
	_, err := DB.Exec(`
		INSERT INTO pvp_stats (character_id, wins, losses, draws, rating, streak, best_streak, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,NOW())
		ON CONFLICT (character_id) DO UPDATE
		  SET wins=$2, losses=$3, draws=$4, rating=$5, streak=$6, best_streak=$7, updated_at=NOW()
	`, charID, wins, losses, draws, rating, streak, bestStreak)
	return err
}

// ClampCharacterEnergy enforces persisted energy bounds in bulk.
func ClampCharacterEnergy() error {
	_, err := DB.Exec(`
		UPDATE characters
		SET energy = LEAST(GREATEST(COALESCE(energy,0), 0), COALESCE(energy_max,100))
		WHERE COALESCE(energy,0) < 0 OR COALESCE(energy,0) > COALESCE(energy_max,100)
	`)
	return err
}

// CleanupExpiredDungeonRuns marks active runs older than ttlHours as expired.
func CleanupExpiredDungeonRuns(ttlHours int) (int64, error) {
	if ttlHours <= 0 {
		ttlHours = 1
	}
	res, err := DB.Exec(`
		UPDATE dungeon_runs
		SET state='expired', finished_at=NOW()
		WHERE state='active' AND started_at < NOW() - ($1::text || ' hours')::interval
	`, ttlHours)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func IncrementItemUsage(itemID string, qty int) error {
	if itemID == "" || qty <= 0 {
		return nil
	}
	_, err := DB.Exec(`
		INSERT INTO item_usage_stats (item_id, purchase_count, updated_at)
		VALUES ($1,$2,NOW())
		ON CONFLICT (item_id) DO UPDATE
		SET purchase_count=item_usage_stats.purchase_count + EXCLUDED.purchase_count,
		    updated_at=NOW()
	`, itemID, qty)
	return err
}

func GetItemUsage(itemID string) (int64, error) {
	var c int64
	err := DB.QueryRow(`
		SELECT COALESCE(purchase_count,0) FROM item_usage_stats WHERE item_id=$1
	`, itemID).Scan(&c)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return c, err
}

// GetCharacterByPlayerID is an alias for GetCharacter (used by GM lookup).
func GetCharacterByPlayerID(playerID int64) (*models.Character, error) {
	return GetCharacter(playerID)
}

// GetCharacterByID for PVP notifications and Pix webhook
func GetCharacterByID(charID int) (*models.Character, error) {
	char := &models.Character{}
	err := DB.QueryRow(`
		SELECT id, player_id, name, race, class, level, experience, experience_next,
			hp, hp_max, mp, mp_max,
			COALESCE(energy, 100), COALESCE(energy_max, 100),
			COALESCE(energy_regen_at, NOW()),
			COALESCE(last_energy_update, EXTRACT(EPOCH FROM COALESCE(energy_regen_at, NOW()))::BIGINT),
			COALESCE(diamonds, 0),
			strength, dexterity, constitution, intelligence,
			wisdom, charisma, attack, defense, magic_attack, magic_defense, speed,
			gold, current_map, state,
			COALESCE(combat_monster_id,''), combat_monster_hp, skill_points,
			COALESCE(deaths,0),
			COALESCE(xp_boost_expiry, '1970-01-01 00:00:00+00'),
			COALESCE(poison_turns, 0), COALESCE(poison_dmg, 0),
			COALESCE(combat_monster_poison_turns, 0), COALESCE(combat_monster_poison_dmg, 0)
		FROM characters WHERE id=$1
	`, charID).Scan(
		&char.ID, &char.PlayerID, &char.Name, &char.Race, &char.Class,
		&char.Level, &char.Experience, &char.ExperienceNext,
		&char.HP, &char.HPMax, &char.MP, &char.MPMax,
		&char.Energy, &char.EnergyMax, &char.EnergyRegenAt, &char.LastEnergyUpdate, &char.Diamonds,
		&char.Strength, &char.Dexterity, &char.Constitution, &char.Intelligence,
		&char.Wisdom, &char.Charisma,
		&char.Attack, &char.Defense, &char.MagicAttack, &char.MagicDefense, &char.Speed,
		&char.Gold, &char.CurrentMap, &char.State,
		&char.CombatMonsterID, &char.CombatMonsterHP, &char.SkillPoints,
		&char.Deaths, &char.XPBoostExpiry,
		&char.PoisonTurns, &char.PoisonDmg,
		&char.CombatMonsterPoisonTurns, &char.CombatMonsterPoisonDmg,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return char, err
}

// SearchCharacterByName finds a character by exact name (for PVP challenges)
// SearchCharacters returns up to `limit` characters matching the name (partial, case-insensitive).
func SearchCharacters(name string, limit int) ([]models.Character, error) {
	rows, err := DB.Query(`
		SELECT id, player_id, name, race, class, level, experience, experience_next,
		       hp, hp_max, mp, mp_max, energy, energy_max, energy_regen_at,
		       COALESCE(last_energy_update, EXTRACT(EPOCH FROM COALESCE(energy_regen_at, NOW()))::BIGINT),
		       diamonds,
		       strength, dexterity, constitution, intelligence, wisdom, charisma,
		       attack, defense, magic_attack, magic_defense, speed,
		       gold, current_map, state, combat_monster_id, combat_monster_hp, skill_points, deaths
		FROM characters WHERE LOWER(name) LIKE LOWER($1) LIMIT $2
	`, "%"+name+"%", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var chars []models.Character
	for rows.Next() {
		var c models.Character
		if err := rows.Scan(&c.ID, &c.PlayerID, &c.Name, &c.Race, &c.Class, &c.Level,
			&c.Experience, &c.ExperienceNext, &c.HP, &c.HPMax, &c.MP, &c.MPMax,
			&c.Energy, &c.EnergyMax, &c.EnergyRegenAt, &c.LastEnergyUpdate, &c.Diamonds,
			&c.Strength, &c.Dexterity, &c.Constitution, &c.Intelligence,
			&c.Wisdom, &c.Charisma, &c.Attack, &c.Defense,
			&c.MagicAttack, &c.MagicDefense, &c.Speed,
			&c.Gold, &c.CurrentMap, &c.State,
			&c.CombatMonsterID, &c.CombatMonsterHP, &c.SkillPoints, &c.Deaths); err != nil {
			return nil, err
		}
		chars = append(chars, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return chars, nil
}

// GetRecentPlayers returns up to `limit` players who were active recently,
// excluding the given characterID (the challenger). Used for the PVP player list.
func GetRecentPlayers(excludeCharID int, limit int) ([]models.Character, error) {
	rows, err := DB.Query(`
		SELECT id, player_id, name, race, class, level, experience, experience_next,
		       hp, hp_max, mp, mp_max, energy, energy_max, energy_regen_at,
		       COALESCE(last_energy_update, EXTRACT(EPOCH FROM COALESCE(energy_regen_at, NOW()))::BIGINT),
		       diamonds,
		       strength, dexterity, constitution, intelligence, wisdom, charisma,
		       attack, defense, magic_attack, magic_defense, speed,
		       gold, current_map, state, combat_monster_id, combat_monster_hp, skill_points, deaths
		FROM characters
		WHERE id != $1 AND is_banned = FALSE
		ORDER BY updated_at DESC
		LIMIT $2
	`, excludeCharID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var chars []models.Character
	for rows.Next() {
		var c models.Character
		if err := rows.Scan(&c.ID, &c.PlayerID, &c.Name, &c.Race, &c.Class, &c.Level,
			&c.Experience, &c.ExperienceNext, &c.HP, &c.HPMax, &c.MP, &c.MPMax,
			&c.Energy, &c.EnergyMax, &c.EnergyRegenAt, &c.LastEnergyUpdate, &c.Diamonds,
			&c.Strength, &c.Dexterity, &c.Constitution, &c.Intelligence,
			&c.Wisdom, &c.Charisma, &c.Attack, &c.Defense,
			&c.MagicAttack, &c.MagicDefense, &c.Speed,
			&c.Gold, &c.CurrentMap, &c.State,
			&c.CombatMonsterID, &c.CombatMonsterHP, &c.SkillPoints, &c.Deaths); err != nil {
			return nil, err
		}
		chars = append(chars, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return chars, nil
}

func SearchCharacterByName(name string) (*models.Character, error) {
	char := &models.Character{}
	err := DB.QueryRow(`
		SELECT id, player_id, name, race, class, level, experience, experience_next,
			hp, hp_max, mp, mp_max,
			COALESCE(energy, 100), COALESCE(energy_max, 100),
			COALESCE(energy_regen_at, NOW()),
			COALESCE(last_energy_update, EXTRACT(EPOCH FROM COALESCE(energy_regen_at, NOW()))::BIGINT),
			COALESCE(diamonds, 0),
			strength, dexterity, constitution, intelligence,
			wisdom, charisma, attack, defense, magic_attack, magic_defense, speed,
			gold, current_map, state,
			COALESCE(combat_monster_id,''), combat_monster_hp, skill_points,
			COALESCE(deaths,0)
		FROM characters WHERE LOWER(name)=LOWER($1) LIMIT 1
	`, name).Scan(
		&char.ID, &char.PlayerID, &char.Name, &char.Race, &char.Class,
		&char.Level, &char.Experience, &char.ExperienceNext,
		&char.HP, &char.HPMax, &char.MP, &char.MPMax,
		&char.Energy, &char.EnergyMax, &char.EnergyRegenAt, &char.LastEnergyUpdate, &char.Diamonds,
		&char.Strength, &char.Dexterity, &char.Constitution, &char.Intelligence,
		&char.Wisdom, &char.Charisma,
		&char.Attack, &char.Defense, &char.MagicAttack, &char.MagicDefense, &char.Speed,
		&char.Gold, &char.CurrentMap, &char.State,
		&char.CombatMonsterID, &char.CombatMonsterHP, &char.SkillPoints,
		&char.Deaths,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return char, err
}

// =============================================
// PLAYER EXTENDED — BAN SYSTEM + GM TOOLS
// =============================================

// CharFull is a combined character+player struct for GM display
type CharFull struct {
	*models.Character
	Banned bool
}

// PlayerRecord is a player row with ban status
type PlayerRecord struct {
	ID       int64
	Username string
	Banned   bool
}

func LogGMAction(gmPlayerID int64, targetCharacterID int, action, details string) error {
	if action == "" {
		return nil
	}
	_, err := DB.Exec(`
		INSERT INTO gm_action_logs (gm_player_id, target_character_id, action, details)
		VALUES ($1,$2,$3,$4)
	`, gmPlayerID, targetCharacterID, action, details)
	return err
}

type EconomicHistoryEntry struct {
	Source    string
	Amount    int
	Reason    string
	CreatedAt time.Time
}

// GetEconomicHistory returns a concise history from diamond_log (and future sources).
func GetEconomicHistory(charID int, limit int) ([]EconomicHistoryEntry, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := DB.Query(`
		SELECT amount, reason, created_at
		FROM diamond_log
		WHERE character_id=$1
		ORDER BY created_at DESC
		LIMIT $2
	`, charID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []EconomicHistoryEntry
	for rows.Next() {
		var e EconomicHistoryEntry
		e.Source = "diamond_log"
		if err := rows.Scan(&e.Amount, &e.Reason, &e.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func GetAllPlayers(limit int) ([]PlayerRecord, error) {
	rows, err := DB.Query(`
		SELECT id, COALESCE(username,''), COALESCE(banned,false)
		FROM players ORDER BY id DESC LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var players []PlayerRecord
	for rows.Next() {
		var p PlayerRecord
		if err := rows.Scan(&p.ID, &p.Username, &p.Banned); err != nil {
			return nil, err
		}
		players = append(players, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return players, nil
}

func BanPlayer(playerID int64) error {
	_, err := DB.Exec(`UPDATE players SET banned=true WHERE id=$1`, playerID)
	return err
}

func UnbanPlayer(playerID int64) error {
	_, err := DB.Exec(`UPDATE players SET banned=false WHERE id=$1`, playerID)
	return err
}

func IsPlayerBanned(playerID int64) bool {
	var banned bool
	if err := DB.QueryRow(`SELECT COALESCE(banned,false) FROM players WHERE id=$1`, playerID).Scan(&banned); err != nil {
		return false
	}
	return banned
}

func GMSetGold(charID int, gold int) error {
	_, err := DB.Exec(`UPDATE characters SET gold=$1 WHERE id=$2`, gold, charID)
	return err
}

func GMSetDiamonds(charID int, diamonds int) error {
	_, err := DB.Exec(`UPDATE characters SET diamonds=$1 WHERE id=$2`, diamonds, charID)
	return err
}

// SearchCharacterFullByName returns CharFull with ban status
func SearchCharacterFullByName(name string) (*CharFull, error) {
	char, err := SearchCharacterByName(name)
	if char == nil || err != nil {
		return nil, err
	}
	banned := IsPlayerBanned(char.PlayerID)
	return &CharFull{Character: char, Banned: banned}, nil
}

// GetCharacterFullByID returns CharFull with ban status
func GetCharacterFullByID(charID int) (*CharFull, error) {
	char, err := GetCharacterByID(charID)
	if char == nil || err != nil {
		return nil, err
	}
	banned := IsPlayerBanned(char.PlayerID)
	return &CharFull{Character: char, Banned: banned}, nil
}

// GetCharacterByID returns a character by ID — but also marks Banned
// For GM gm.go: it uses GetCharacterByID and needs Banned field —
// We handle it via CharFull above.
