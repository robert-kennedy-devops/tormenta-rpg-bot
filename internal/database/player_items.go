package database

import (
	"database/sql"
	"time"
)

type PlayerItemRow struct {
	InstanceID   string
	CharacterID  int
	TemplateID   string
	UpgradeLevel int
	Broken       bool
	Equipped     bool
	EquippedSlot string
	CreatedAt    time.Time
}

func CreatePlayerItemInstances(charID int, templateID string, qty int) error {
	if qty < 1 {
		return nil
	}
	for i := 0; i < qty; i++ {
		if _, err := DB.Exec(`
			INSERT INTO player_items (character_id, template_id)
			VALUES ($1, $2)
		`, charID, templateID); err != nil {
			return err
		}
	}
	return nil
}

func DeleteOnePlayerItemInstance(charID int, templateID string, preferUnequipped bool) (bool, error) {
	where := "character_id=$1 AND template_id=$2"
	if preferUnequipped {
		where += " AND equipped=false"
	}
	res, err := DB.Exec(`
		DELETE FROM player_items
		WHERE instance_id IN (
			SELECT instance_id
			FROM player_items
			WHERE `+where+`
			ORDER BY equipped ASC, created_at ASC
			LIMIT 1
		)
	`, charID, templateID)
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

func GetBestPlayerItemForTemplate(charID int, templateID string) (*PlayerItemRow, error) {
	row := &PlayerItemRow{}
	err := DB.QueryRow(`
		SELECT instance_id, character_id, template_id, upgrade_level, broken,
		       equipped, COALESCE(equipped_slot,''), created_at
		FROM player_items
		WHERE character_id=$1 AND template_id=$2
		ORDER BY broken ASC, upgrade_level DESC, created_at ASC
		LIMIT 1
	`, charID, templateID).Scan(
		&row.InstanceID, &row.CharacterID, &row.TemplateID, &row.UpgradeLevel, &row.Broken,
		&row.Equipped, &row.EquippedSlot, &row.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return row, err
}

func UpdatePlayerItemForge(instanceID string, newLevel int, broken bool) error {
	_, err := DB.Exec(`
		UPDATE player_items
		SET upgrade_level=$2, broken=$3, updated_at=NOW()
		WHERE instance_id=$1
	`, instanceID, newLevel, broken)
	return err
}

func EquipPlayerItemByTemplateAndSlot(charID int, templateID string, slot string) error {
	if slot != "" {
		if _, err := DB.Exec(`
			UPDATE player_items
			SET equipped=false, equipped_slot=NULL, updated_at=NOW()
			WHERE character_id=$1 AND equipped_slot=$2
		`, charID, slot); err != nil {
			return err
		}
	}

	_, err := DB.Exec(`
		UPDATE player_items
		SET equipped=true, equipped_slot=$3, updated_at=NOW()
		WHERE instance_id IN (
			SELECT instance_id
			FROM player_items
			WHERE character_id=$1 AND template_id=$2 AND broken=false
			ORDER BY upgrade_level DESC, created_at ASC
			LIMIT 1
		)
	`, charID, templateID, slot)
	return err
}

func UnequipPlayerItemsBySlot(charID int, slot string) error {
	_, err := DB.Exec(`
		UPDATE player_items
		SET equipped=false, equipped_slot=NULL, updated_at=NOW()
		WHERE character_id=$1 AND equipped_slot=$2
	`, charID, slot)
	return err
}

func UnequipPlayerItemsByTemplate(charID int, templateID string) error {
	_, err := DB.Exec(`
		UPDATE player_items
		SET equipped=false, equipped_slot=NULL, updated_at=NOW()
		WHERE character_id=$1 AND template_id=$2
	`, charID, templateID)
	return err
}

func SyncInventoryEquippedFromInstances(charID int, templateID string) error {
	var equippedCount int
	if err := DB.QueryRow(`
		SELECT COUNT(1)
		FROM player_items
		WHERE character_id=$1 AND template_id=$2 AND equipped=true
	`, charID, templateID).Scan(&equippedCount); err != nil {
		return err
	}
	_, err := DB.Exec(`
		UPDATE inventory
		SET equipped=$3
		WHERE character_id=$1 AND item_id=$2
	`, charID, templateID, equippedCount > 0)
	return err
}
