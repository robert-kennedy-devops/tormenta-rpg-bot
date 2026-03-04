package assets

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
)

// ImageKey maps logical names to file paths
var ImageKey = map[string]string{
	// UI
	"welcome":   "ui/welcome.png",
	"menu":      "ui/menu.png",
	"status":    "ui/status.png",
	"inventory": "ui/inventory.png",
	"skills":    "ui/skills.png",
	"shop":      "ui/shop.png",
	"travel":    "ui/travel.png",
	"combat":    "ui/combat.png",
	"victory":   "ui/victory.png",
	"defeat":    "ui/defeat.png",
	"rest":      "ui/rest.png",

	// Races
	"race_human":   "races/human.png",
	"race_elf":     "races/elf.png",
	"race_dwarf":   "races/dwarf.png",
	"race_halforc": "races/halforc.png",

	// Classes
	"class_warrior": "classes/warrior.png",
	"class_mage":    "classes/mage.png",
	"class_rogue":   "classes/rogue.png",
	"class_archer":  "classes/archer.png",

	// Monsters
	"monster_rat":           "monsters/rat.png",
	"monster_goblin":        "monsters/goblin.png",
	"monster_slime":         "monsters/slime.png",
	"monster_wolf":          "monsters/wolf.png",
	"monster_orc":           "monsters/orc.png",
	"monster_troll":         "monsters/troll.png",
	"monster_bandit_leader": "monsters/bandit_leader.png",
	"monster_bat":           "monsters/bat.png",
	"monster_spider":        "monsters/spider.png",
	"monster_golem":         "monsters/golem.png",
	"monster_undead_knight": "monsters/undead_knight.png",
	"monster_demon":         "monsters/demon.png",
	"monster_necromancer":   "monsters/necromancer.png",
	"monster_vampire_lord":  "monsters/vampire_lord.png",
	"monster_dragon_young":  "monsters/dragon_young.png",
	"monster_dragon_elder":  "monsters/dragon_elder.png",

	// New monsters
	"monster_mushroom":          "monsters/mushroom.png",
	"monster_crow":              "monsters/crow.png",
	"monster_harpy":             "monsters/harpy.png",
	"monster_werewolf":          "monsters/werewolf.png",
	"monster_stone_golem_shard": "monsters/stone_golem_shard.png",
	"monster_crystal_wraith":    "monsters/crystal_wraith.png",
	"monster_shadow_assassin":   "monsters/shadow_assassin.png",
	"monster_lich":              "monsters/lich.png",
	"monster_wyvern":            "monsters/wyvern.png",
	"monster_phoenix":           "monsters/phoenix.png",

	// Maps
	"map_village":          "maps/village.png",
	"map_village_outskirts":"maps/village_outskirts.png",
	"map_dark_forest":      "maps/dark_forest.png",
	"map_forest_camp":      "maps/forest_camp.png",
	"map_crystal_cave":     "maps/crystal_cave.png",
	"map_ancient_dungeon":  "maps/ancient_dungeon.png",
	"map_dungeon_outpost":  "maps/dungeon_outpost.png",
	"map_dragon_peak":      "maps/dragon_peak.png",

	// Items
	"item_weapon":     "items/weapons.png",
	"item_armor":      "items/armors.png",
	"item_consumable": "items/consumables.png",
}

// Manager handles image paths and file_id caching
type Manager struct {
	baseDir string
	cache   map[string]string // key -> telegram file_id
	mu      sync.RWMutex
	db      FileIDStore
}

// FileIDStore is the interface for persisting file_ids
type FileIDStore interface {
	SaveFileID(key, fileID string) error
	LoadFileID(key string) (string, bool)
	LoadAll() map[string]string
}

var Default *Manager

// Init initializes the global asset manager
func Init(baseDir string, store FileIDStore) error {
	log.Printf("🖼️  Generating images in %s...", baseDir)
	if err := GenerateAllImages(baseDir); err != nil {
		return fmt.Errorf("image generation failed: %w", err)
	}

	Default = &Manager{
		baseDir: baseDir,
		cache:   make(map[string]string),
		db:      store,
	}

	// Load cached file_ids from DB
	if store != nil {
		cached := store.LoadAll()
		for k, v := range cached {
			Default.cache[k] = v
		}
		log.Printf("🖼️  Loaded %d cached file_ids", len(cached))
	}

	log.Printf("✅ Image manager ready (%d images)", len(ImageKey))
	return nil
}

// GetPath returns the absolute path for an image key
func (m *Manager) GetPath(key string) string {
	rel, ok := ImageKey[key]
	if !ok {
		return ""
	}
	return filepath.Join(m.baseDir, rel)
}

// GetFileID returns a cached Telegram file_id for an image key
func (m *Manager) GetFileID(key string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.cache[key]
}

// SetFileID saves a Telegram file_id for an image key
func (m *Manager) SetFileID(key, fileID string) {
	m.mu.Lock()
	m.cache[key] = fileID
	m.mu.Unlock()

	if m.db != nil {
		if err := m.db.SaveFileID(key, fileID); err != nil {
			log.Printf("Warning: could not save file_id for %s: %v", key, err)
		}
	}
}

// FileExists checks if the image file exists on disk
func (m *Manager) FileExists(key string) bool {
	path := m.GetPath(key)
	if path == "" {
		return false
	}
	_, err := os.Stat(path)
	return err == nil
}

// MonsterImageKey returns the image key for a monster ID
func MonsterImageKey(monsterID string) string {
	return "monster_" + monsterID
}

// MapImageKey returns the image key for a map ID
func MapImageKey(mapID string) string {
	return "map_" + mapID
}

// RaceImageKey returns the image key for a race ID
func RaceImageKey(raceID string) string {
	return "race_" + raceID
}

// ClassImageKey returns the image key for a class ID
func ClassImageKey(classID string) string {
	return "class_" + classID
}

// ItemTypeImageKey returns the image key for an item type
func ItemTypeImageKey(itemType string) string {
	return "item_" + itemType
}
