package game

import (
	"time"

	"github.com/tormenta-bot/internal/models"
)

// =============================================
// ENERGY CONSTANTS
// =============================================

const (
	EnergyRegenInterval    = 10 * time.Minute // 1 energy every 10 min (normal)
	EnergyRegenIntervalVIP = 5 * time.Minute  // 1 energy every 5 min (VIP)
	EnergyBaseMax          = 100              // fixed cap (normal)
	EnergyBaseMaxVIP       = 200              // fixed cap (VIP)
	EnergyPerLevel         = 0                // sem crescimento de energia por nível

	// Costs (energy → vitals)
	EnergyPerHP = 5 // 1 energy = 5 HP recovered
	EnergyPerMP = 3 // 1 energy = 3 MP recovered

	// Diamond costs
	DiamondFullEnergyRefill = 30 // diamonds to fully refill energy
	DiamondRevive           = 20 // diamonds to revive without gold loss
	DiamondFullHeal         = 15 // diamonds to fully heal HP+MP

	// Diamond earn rates
	DiamondBossKillChance = 25 // % chance boss drops 1 diamond
	DiamondDailyBonus     = 3  // diamonds per daily login
)

// =============================================
// ENERGY CALCULATIONS
// =============================================

// MaxEnergy returns the energy cap — fixed at EnergyBaseMax regardless of level.
func MaxEnergy(level int) int {
	return EnergyBaseMax
}

// MaxEnergyVIP returns the VIP energy cap — fixed at EnergyBaseMaxVIP regardless of level.
func MaxEnergyVIP(level int) int {
	return EnergyBaseMaxVIP
}

// RegenInterval returns the regen interval based on VIP status.
func RegenInterval(isVIP bool) time.Duration {
	if isVIP {
		return EnergyRegenIntervalVIP
	}
	return EnergyRegenInterval
}

// TickEnergy applies accumulated regeneration since last tick.
// Returns how many energy points were restored.
func TickEnergy(char *models.Character) int {
	// Heurística simples e estável: no modelo atual, VIP possui EnergyMax maior
	// que o cap base normal.
	isVIP := char.EnergyMax > EnergyBaseMax
	return TickEnergyVIP(char, isVIP)
}

// TickEnergyVIP applies accumulated regeneration with VIP-aware interval.
func TickEnergyVIP(char *models.Character, isVIP bool) int {
	if char == nil {
		return 0
	}
	if isVIP {
		char.EnergyMax = EnergyBaseMaxVIP
	} else {
		char.EnergyMax = EnergyBaseMax
	}

	interval := RegenInterval(isVIP)
	now := time.Now()
	nowUnix := now.Unix()
	intervalSec := int64(interval.Seconds())
	if intervalSec <= 0 {
		return 0
	}

	// Inicializa timestamp se necessário.
	if char.LastEnergyUpdate <= 0 {
		char.LastEnergyUpdate = nowUnix
		char.EnergyRegenAt = now
		return 0
	}

	// Já no máximo — apenas avança o clock para não acumular "ticks fantasmas"
	if char.Energy >= char.EnergyMax {
		char.Energy = char.EnergyMax // garante que nunca ultrapasse
		char.LastEnergyUpdate = nowUnix
		char.EnergyRegenAt = now
		return 0
	}

	elapsedSec := nowUnix - char.LastEnergyUpdate
	if elapsedSec < intervalSec {
		return 0
	}
	ticks := int(elapsedSec / intervalSec)
	if ticks == 0 {
		return 0
	}

	// Calcula quantos ticks realmente serão usados (limitado pelo espaço disponível)
	space := char.EnergyMax - char.Energy
	gain := ticks
	if gain > space {
		gain = space
	}
	char.Energy += gain
	if char.Energy > char.EnergyMax {
		char.Energy = char.EnergyMax
	}

	// Avança o clock apenas pelos ticks efetivamente consumidos,
	// mas descarta ticks excedentes se já chegou ao máximo.
	if char.Energy >= char.EnergyMax {
		// Evita acúmulo de ganho quando já está no teto.
		char.LastEnergyUpdate = nowUnix
		char.EnergyRegenAt = now
	} else {
		char.LastEnergyUpdate += int64(ticks) * intervalSec
		char.EnergyRegenAt = time.Unix(char.LastEnergyUpdate, 0)
	}
	return gain
}

// NextRegenIn returns duration until next energy point is gained
func NextRegenIn(char *models.Character) time.Duration {
	isVIP := char.EnergyMax > EnergyBaseMax
	return NextRegenInVIP(char, isVIP)
}

// NextRegenInVIP returns duration until next energy point, considering VIP.
func NextRegenInVIP(char *models.Character, isVIP bool) time.Duration {
	if char == nil {
		return 0
	}
	if char.Energy >= char.EnergyMax {
		return 0
	}
	interval := RegenInterval(isVIP)
	if char.LastEnergyUpdate <= 0 {
		return interval
	}
	elapsed := time.Now().Unix() - char.LastEnergyUpdate
	remaining := interval - time.Duration(elapsed)*time.Second
	if remaining < 0 {
		return 0
	}
	return remaining
}

// =============================================
// ENERGY ↔ HP RECOVERY
// =============================================

// RecoverHPWithEnergy spends energy to restore HP.
// Returns actual hp gained and energy spent.
func RecoverHPWithEnergy(char *models.Character, energyToSpend int) (hpGained, energySpent int) {
	if char.HP >= char.HPMax {
		return 0, 0
	}
	if energyToSpend > char.Energy {
		energyToSpend = char.Energy
	}
	if energyToSpend <= 0 {
		return 0, 0
	}
	maxRecoverable := char.HPMax - char.HP
	potentialHP := energyToSpend * EnergyPerHP
	if potentialHP > maxRecoverable {
		// Only spend what's needed
		energyToSpend = (maxRecoverable + EnergyPerHP - 1) / EnergyPerHP
		potentialHP = maxRecoverable
	}
	char.HP += potentialHP
	if char.HP > char.HPMax {
		char.HP = char.HPMax
	}
	char.Energy -= energyToSpend
	return potentialHP, energyToSpend
}

// RecoverMPWithEnergy spends energy to restore MP.
// Returns actual mp gained and energy spent.
func RecoverMPWithEnergy(char *models.Character, energyToSpend int) (mpGained, energySpent int) {
	if char.MP >= char.MPMax {
		return 0, 0
	}
	if energyToSpend > char.Energy {
		energyToSpend = char.Energy
	}
	if energyToSpend <= 0 {
		return 0, 0
	}
	maxRecoverable := char.MPMax - char.MP
	potentialMP := energyToSpend * EnergyPerMP
	if potentialMP > maxRecoverable {
		energyToSpend = (maxRecoverable + EnergyPerMP - 1) / EnergyPerMP
		potentialMP = maxRecoverable
	}
	char.MP += potentialMP
	if char.MP > char.MPMax {
		char.MP = char.MPMax
	}
	char.Energy -= energyToSpend
	return potentialMP, energyToSpend
}

// =============================================
// ENERGY BARS
// =============================================

// EnergyBar renders a visual energy bar
func EnergyBar(current, max int) string {
	if max == 0 {
		return ""
	}
	pct := current * 8 / max
	if pct < 0 {
		pct = 0
	}
	if pct > 8 {
		pct = 8
	}
	bar := ""
	for i := 0; i < 8; i++ {
		if i < pct {
			bar += "⚡"
		} else {
			bar += "🔲"
		}
	}
	return bar
}

// =============================================
// DIAMOND OPERATIONS
// =============================================

// DiamondPackages lists available diamond purchase bundles
// DiamondPackages are defined in game/pix.go

// DiamondShopItems lists items purchasable exclusively with diamonds
// DiamondItem describes a diamond-only purchasable service (instant effect, no inventory)
type DiamondItem struct {
	ID          string
	Name        string
	Emoji       string
	Description string
	Cost        int // diamonds
}

// DiamondItems contains only instant-effect services (not items that go to inventory).
// Items like revive_token, xp_boost, energy_elixir and skill_reset are in game.Items
// with DiamondPrice set, and appear in the shop via handleDiamondItemBuy.
var DiamondItems = map[string]DiamondItem{
	"energy_full": {
		ID: "energy_full", Name: "Recarga Total de Energia", Emoji: "⚡",
		Description: "Recupera toda a sua Energia instantaneamente.",
		Cost:        DiamondFullEnergyRefill,
	},
	"hp_full": {
		ID: "hp_full", Name: "Cura Divina", Emoji: "💖",
		Description: "Restaura todo o HP e MP instantaneamente.",
		Cost:        DiamondFullHeal,
	},
}

// =============================================
// ENERGY COST FOR INN
// =============================================

// InnEnergyCost returns energy cost to rest at an inn
// (replaces gold cost for HP/MP recovery)
func InnEnergyCost(char *models.Character) int {
	missingHP := char.HPMax - char.HP
	missingMP := char.MPMax - char.MP
	hpEnergy := (missingHP + EnergyPerHP - 1) / EnergyPerHP
	mpEnergy := (missingMP + EnergyPerMP - 1) / EnergyPerMP
	total := hpEnergy + mpEnergy
	if total < 1 && (missingHP > 0 || missingMP > 0) {
		total = 1
	}
	return total
}

// =============================================
// ENERGY COST PER COMBAT ACTION
// =============================================

const (
	EnergyPerAttack    = 2  // basic attack costs 2 energy (kept for compatibility)
	EnergyPerSkill     = 4  // skill attack (kept for compatibility)
	EnergyCombatEnter  = 1  // entering any combat costs 1 energy
	EnergyTravelCost   = 1  // traveling between maps costs 1 energy
	EnergyDungeonEnter = 10 // entering a dungeon floor costs 10 energy
	EnergyPVPMatch     = 5  // starting a PVP match costs 5 energy
)

// ConsumeAttackEnergy deducts 1 energy when entering combat (not per attack).
func ConsumeAttackEnergy(char *models.Character) bool {
	if char.Energy < EnergyCombatEnter {
		return false
	}
	char.Energy -= EnergyCombatEnter
	return true
}

// ConsumeSkillEnergy — skills no longer cost extra energy, same as entering combat.
func ConsumeSkillEnergy(char *models.Character) bool {
	return true // skills don't cost extra energy anymore
}

// ConsumeTravelEnergy deducts 1 energy when traveling between maps.
func ConsumeTravelEnergy(char *models.Character) bool {
	if char.Energy < EnergyTravelCost {
		return false
	}
	char.Energy -= EnergyTravelCost
	return true
}

// ConsumeDungeonEnergy deducts energy for entering a dungeon floor.
// Returns false if not enough energy.
func ConsumeDungeonEnergy(char *models.Character) bool {
	if char.Energy < EnergyDungeonEnter {
		return false
	}
	char.Energy -= EnergyDungeonEnter
	return true
}

// ConsumePVPEnergy deducts energy for a PVP match.
func ConsumePVPEnergy(char *models.Character) bool {
	if char.Energy < EnergyPVPMatch {
		return false
	}
	char.Energy -= EnergyPVPMatch
	return true
}
