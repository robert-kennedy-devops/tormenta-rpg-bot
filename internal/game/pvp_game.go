package game

import (
	"fmt"
	"math/rand"

	"github.com/tormenta-bot/internal/models"
)

// =============================================
// PVP RESULT
// =============================================

type PVPResult struct {
	Damage        int
	Message       string // mensagem para quem atacou
	IsCritical    bool
	IsFumble      bool
	IsMiss        bool
	D20Roll       int
	AtkBonus      int
	TotalRoll     int
	TargetCA      int
	AppliesPoison bool
	PoisonDmg     int
	PoisonTurns   int
}

// =============================================
// HELPERS DE CA/ATAQUE
// =============================================

// pvpCA calcula a CA do personagem para PVP (inclui equipamento).
func pvpCA(char *models.Character) int {
	defAttr := DefensiveAttr(char.Class, char.Constitution, char.Dexterity, char.Intelligence)
	return CharacterCA(char.Class, defAttr, char.EquipCABonus)
}

// pvpAtkBonus calcula o bônus de ataque d20 do personagem para PVP.
func pvpAtkBonus(char *models.Character) int {
	return CharacterAttackBonus(char.Class, char.Level,
		char.Strength, char.Dexterity, char.Intelligence) + char.EquipHitBonus
}

// =============================================
// ATAQUE BÁSICO PVP (d20 + CA)
// =============================================

func PVPAttack(attacker *models.Character, defender *models.Character) PVPResult {
	result := PVPResult{}

	atkBonus := pvpAtkBonus(attacker)
	defCA := pvpCA(defender)
	d20 := rollD20()
	total := d20 + atkBonus
	isCrit := d20 == 20
	isFumble := d20 == 1

	result.D20Roll = d20
	result.AtkBonus = atkBonus
	result.TotalRoll = total
	result.TargetCA = defCA

	if isFumble {
		result.IsFumble = true
		result.IsMiss = true
		result.Message = fmt.Sprintf(
			"🎲 [1] 💨 *Fumble!* %s escorregou e errou feio!",
			attacker.Name,
		)
		return result
	}

	if !isCrit && total < defCA {
		result.IsMiss = true
		result.Message = fmt.Sprintf(
			"🎲 [%d]+%d=*%d* vs CA *%d* — ❌ *Errou!* %s desviou.",
			d20, atkBonus, total, defCA, defender.Name,
		)
		return result
	}

	// Dano
	dice, sides, attrMod := ClassDamageRoll(attacker.Class, attacker.Strength, attacker.Dexterity, attacker.Intelligence)
	dmg := rollDice(dice, sides) + attrMod
	if dmg < 1 {
		dmg = 1
	}

	// Bônus raciais
	if attacker.Race == "halforc" {
		dmg = int(float64(dmg) * 1.25)
	}
	if defender.Race == "dwarf" {
		dmg = int(float64(dmg) * 0.85)
	}

	if isCrit {
		dmg *= 2
		result.IsCritical = true
		result.Message = fmt.Sprintf(
			"🎲 [20] ⭐ *CRÍTICO!* %dd%d+%d → *%d* de dano em %s! (dobrado)",
			dice, sides, attrMod, dmg, defender.Name,
		)
	} else {
		result.Message = fmt.Sprintf(
			"🎲 [%d]+%d=*%d* vs CA *%d* ✅ — %dd%d+%d → *%d* de dano em %s.",
			d20, atkBonus, total, defCA, dice, sides, attrMod, dmg, defender.Name,
		)
	}

	result.Damage = dmg
	return result
}

// =============================================
// ATAQUE DE HABILIDADE PVP (d20 + CA)
// =============================================

func PVPSkillAttack(attacker *models.Character, skill *models.Skill, defender *models.Character) PVPResult {
	result := PVPResult{}

	// Habilidade passiva ou sem dano: sem rolagem
	if skill.Passive || skill.Damage == 0 {
		result.Message = fmt.Sprintf(
			"%s *%s* usada! _(habilidades de suporte não causam dano direto em PVP)_",
			skill.Emoji, skill.Name,
		)
		return result
	}

	atkBonus := pvpAtkBonus(attacker)
	defCA := pvpCA(defender)
	d20 := rollD20()
	total := d20 + atkBonus
	isCrit := d20 == 20
	isFumble := d20 == 1

	result.D20Roll = d20
	result.AtkBonus = atkBonus
	result.TotalRoll = total
	result.TargetCA = defCA

	if isFumble {
		result.IsFumble = true
		result.IsMiss = true
		result.Message = fmt.Sprintf(
			"🎲 [1] %s *%s* — 💨 *Fumble!* Conjuração desperdiçada!",
			skill.Emoji, skill.Name,
		)
		return result
	}

	if !isCrit && total < defCA {
		result.IsMiss = true
		result.Message = fmt.Sprintf(
			"🎲 [%d]+%d=*%d* vs CA *%d* — %s *%s* ❌ *Falhou!* %s resistiu.",
			d20, atkBonus, total, defCA, skill.Emoji, skill.Name, defender.Name,
		)
		return result
	}

	// Dano por tipo de habilidade
	var dmg int
	var dmgDesc string
	switch skill.DamageType {
	case "magic":
		dice, sides, mod := 1, 6, attrMod(attacker.Intelligence)
		rolled := rollDice(dice, sides) + mod + skill.Damage
		if rolled < 1 {
			rolled = 1
		}
		dmg = rolled
		if attacker.Race == "elf" {
			dmg = int(float64(dmg) * 1.2)
		}
		dmgDesc = fmt.Sprintf("1d6+%d+%d mágico", mod, skill.Damage)
	case "poison":
		base := skill.Damage + int(float64(attacker.Dexterity)*0.3)
		if base < 1 {
			base = 1
		}
		dmg = base
		dmgDesc = fmt.Sprintf("%d veneno", dmg)
		if skill.PoisonTurnsCount > 0 {
			result.AppliesPoison = true
			result.PoisonDmg = skill.PoisonDmgPerTurn
			result.PoisonTurns = skill.PoisonTurnsCount
		}
	default: // physical
		dice, sides, attrM := ClassDamageRoll(attacker.Class, attacker.Strength, attacker.Dexterity, attacker.Intelligence)
		rolled := rollDice(dice, sides) + attrM + skill.Damage
		if rolled < 1 {
			rolled = 1
		}
		dmg = rolled
		if attacker.Race == "halforc" {
			dmg = int(float64(dmg) * 1.15)
		}
		if defender.Race == "dwarf" {
			dmg = int(float64(dmg) * 0.85)
		}
		dmgDesc = fmt.Sprintf("%dd%d+%d+%d físico", dice, sides, attrM, skill.Damage)
	}

	if isCrit {
		dmg *= 2
		result.IsCritical = true
		result.Message = fmt.Sprintf(
			"🎲 [20] ⭐ *CRÍTICO!* %s *%s* — %s → *%d* de dano em %s! (dobrado)",
			skill.Emoji, skill.Name, dmgDesc, dmg, defender.Name,
		)
	} else {
		result.Message = fmt.Sprintf(
			"🎲 [%d]+%d=*%d* vs CA *%d* ✅ — %s *%s* — %s → *%d* de dano em %s.",
			d20, atkBonus, total, defCA, skill.Emoji, skill.Name, dmgDesc, dmg, defender.Name,
		)
	}

	result.Damage = dmg
	return result
}

// =============================================
// ELO
// =============================================

func CalculateELO(winnerRating, loserRating int) (int, int) {
	diff := float64(loserRating-winnerRating) / 400.0
	pow := eloExp(diff * 2.302585093)
	expected := 1.0 / (1.0 + pow)
	change := int(float64(32) * (1.0 - expected))
	if change < 1 {
		change = 1
	}
	if change > 32 {
		change = 32
	}
	return winnerRating + change, loserRating - change
}

func eloExp(x float64) float64 {
	if x > 20 {
		return 1e9
	}
	if x < -20 {
		return 0
	}
	result, term := 1.0, 1.0
	for i := 1; i <= 30; i++ {
		term *= x / float64(i)
		result += term
		if term > -1e-9 && term < 1e-9 {
			break
		}
	}
	if result < 0 {
		return 0
	}
	return result
}

// =============================================
// HELPERS PÚBLICOS
// =============================================

func PVPStakeOptions(char *models.Character) []int {
	opts := []int{0, 50, 100, 200, 500}
	var valid []int
	for _, o := range opts {
		if char.Gold >= o {
			valid = append(valid, o)
		}
	}
	return valid
}

func PVPRankTitle(rating int) string {
	switch {
	case rating >= 2000:
		return "👑 Lendário"
	case rating >= 1600:
		return "💎 Grão-Mestre"
	case rating >= 1400:
		return "🥇 Mestre"
	case rating >= 1200:
		return "🥈 Platina"
	case rating >= 1100:
		return "🥉 Ouro"
	case rating >= 1000:
		return "⚔️ Prata"
	default:
		return "🗡️ Bronze"
	}
}

// Ensure rand is seeded (Go 1.20+ auto-seeds, kept for older versions)
var _ = func() int { rand.Seed(42); return 0 }()
