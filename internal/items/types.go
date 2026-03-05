package items

import "github.com/tormenta-bot/internal/models"

// StatBlock stores item combat modifiers in a structured form.
type StatBlock struct {
	AttackBonus   int
	MagicAtkBonus int
	DefenseBonus  int
	MagicDefBonus int
	SpeedBonus    int
	CABonus       int
	HitBonus      int
	HPBonus       int
	MPBonus       int
}

// ItemTemplate is the static definition used as source-of-truth for item rules.
type ItemTemplate struct {
	ID          string
	BaseItemID  string
	Name        string
	Description string
	Emoji       string
	Type        string
	Rarity      models.Rarity
	Price       int
	SellPrice   int
	MinLevel    int
	ClassReq    string
	Slot        string
	Stats       StatBlock
}

// PlayerItem is an owned item instance with progression metadata.
type PlayerItem struct {
	InstanceID   string
	CharacterID  int
	TemplateID   string
	UpgradeLevel int
	Broken       bool
	Quantity     int
	Equipped     bool
	EquippedSlot string
	CustomRarity *models.Rarity
	CustomStats  *StatBlock
}

// EffectiveStats returns the runtime stats from template + upgrades + overrides.
func (pi PlayerItem) EffectiveStats(t ItemTemplate) StatBlock {
	base := t.Stats
	if pi.CustomStats != nil {
		base = *pi.CustomStats
	}
	if pi.Broken {
		return StatBlock{}
	}

	// Linear 5% per upgrade level, rounded down.
	m := 100 + (pi.UpgradeLevel * 5)
	return StatBlock{
		AttackBonus:   base.AttackBonus * m / 100,
		MagicAtkBonus: base.MagicAtkBonus * m / 100,
		DefenseBonus:  base.DefenseBonus * m / 100,
		MagicDefBonus: base.MagicDefBonus * m / 100,
		SpeedBonus:    base.SpeedBonus * m / 100,
		CABonus:       base.CABonus * m / 100,
		HitBonus:      base.HitBonus * m / 100,
		HPBonus:       base.HPBonus * m / 100,
		MPBonus:       base.MPBonus * m / 100,
	}
}
