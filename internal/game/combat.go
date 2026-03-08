package game

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/tormenta-bot/internal/models"
)

// =============================================
// COMBAT RESULT
// =============================================

type CombatResult struct {
	PlayerDamage   int
	MonsterDamage  int
	PlayerMessage  string
	MonsterMessage string
	IsCritical     bool
	IsPlayerMiss   bool
	IsMonsterMiss  bool
	IsMonsterCrit  bool
	PlayerRoll     int
	MonsterRoll    int
	EnergyUsed     int
	StatusEffect   string
	BuffValue      int
	// Veneno aplicado ao monstro neste turno
	AppliesPoison bool
	PoisonDmg     int
	PoisonTurns   int
	// mantido por compatibilidade
	IsEvaded bool
}

type LevelUpResult struct {
	NewLevel    int
	HPGained    int
	MPGained    int
	SkillPoints int
}

// rollD20 rola um dado de 20 faces (1-20).
func rollD20() int { return rand.Intn(20) + 1 }

// rollDice rola N dados de S faces.
func rollDice(n, s int) int {
	total := 0
	for i := 0; i < n; i++ {
		total += rand.Intn(s) + 1
	}
	return total
}

// equippedBonuses retorna (hitBonus, caBonus) dos itens equipados.
func equippedBonuses(char *models.Character) (int, int) {
	return char.EquipHitBonus, char.EquipCABonus
}

// calcPlayerCA calcula a CA total do personagem.
func calcPlayerCA(char *models.Character, equipCA int) int {
	defAttr := DefensiveAttr(char.Class, char.Constitution, char.Dexterity, char.Intelligence)
	return CharacterCA(char.Class, defAttr, equipCA)
}

// calcPlayerAttackBonus calcula o bônus de ataque d20 do personagem.
func calcPlayerAttackBonus(char *models.Character, equipHit int) int {
	return CharacterAttackBonus(char.Class, char.Level, char.Strength, char.Dexterity, char.Intelligence) + equipHit
}

// =============================================
// ATAQUE BÁSICO DO JOGADOR
// =============================================

func PlayerAttack(char *models.Character, monster *models.Monster) CombatResult {
	result := CombatResult{EnergyUsed: EnergyPerAttack}

	equipHit, equipCA := equippedBonuses(char)
	atkBonus := calcPlayerAttackBonus(char, equipHit)

	d20 := rollD20()
	result.PlayerRoll = d20
	total := d20 + atkBonus

	isCrit := d20 == 20
	isFumble := d20 == 1

	if isFumble {
		result.IsPlayerMiss = true
		result.PlayerMessage = fmt.Sprintf(
			"🎲 [1] 💨 *Fumble!* Você escorregou e errou feio!",
		)
	} else if !isCrit && total < monster.CA {
		result.IsPlayerMiss = true
		result.PlayerMessage = fmt.Sprintf(
			"🎲 [%d]+%d=*%d* vs CA %d — ❌ *Errou!* %s desviou.",
			d20, atkBonus, total, monster.CA, monster.Name,
		)
	} else {
		dice, sides, attrMod := ClassDamageRoll(char.Class, char.Strength, char.Dexterity, char.Intelligence)
		dmg := rollDice(dice, sides) + attrMod
		if dmg < 1 {
			dmg = 1
		}
		if char.Race == "halforc" {
			dmg = int(float64(dmg) * 1.25)
		}

		if isCrit {
			dmg *= 2
			result.IsCritical = true
			result.PlayerMessage = fmt.Sprintf(
				"🎲 [20] ⭐ *CRÍTICO!* %dd%d+%d → *%d* de dano! (dobrado)",
				dice, sides, attrMod, dmg,
			)
		} else {
			result.PlayerMessage = fmt.Sprintf(
				"🎲 [%d]+%d=*%d* vs CA %d ✅ — %dd%d+%d → *%d* de dano.",
				d20, atkBonus, total, monster.CA, dice, sides, attrMod, dmg,
			)
		}
		result.PlayerDamage = dmg
	}

	mr := monsterAttack(monster, char, equipCA)
	result.MonsterDamage = mr.MonsterDamage
	result.MonsterMessage = mr.MonsterMessage
	result.IsMonsterMiss = mr.IsMonsterMiss
	result.IsMonsterCrit = mr.IsMonsterCrit
	result.MonsterRoll = mr.MonsterRoll
	return result
}

// =============================================
// ATAQUE DE HABILIDADE
// =============================================

func PlayerSkillAttack(char *models.Character, skill *models.Skill, monster *models.Monster) CombatResult {
	result := CombatResult{EnergyUsed: EnergyPerSkill}

	if char.MP < skill.MPCost {
		result.PlayerMessage = "❌ MP insuficiente para usar esta habilidade!"
		return result
	}

	if skill.Damage == 0 {
		result.PlayerMessage = skillBuffEffect(skill)
		result.StatusEffect = skill.DamageType
		result.BuffValue = 30
		mr := monsterAttack(monster, char, char.EquipCABonus)
		result.MonsterDamage = mr.MonsterDamage
		result.MonsterMessage = mr.MonsterMessage
		result.IsMonsterMiss = mr.IsMonsterMiss
		result.IsMonsterCrit = mr.IsMonsterCrit
		return result
	}

	equipHit, equipCA := equippedBonuses(char)
	atkBonus := calcPlayerAttackBonus(char, equipHit)
	d20 := rollD20()
	result.PlayerRoll = d20
	total := d20 + atkBonus
	isCrit := d20 == 20
	isFumble := d20 == 1

	if isFumble {
		result.IsPlayerMiss = true
		result.PlayerMessage = fmt.Sprintf(
			"🎲 [1] 💨 *Fumble!* %s *%s* falhou — conjuração desperdiçada.",
			skill.Emoji, skill.Name,
		)
	} else if !isCrit && total < monster.CA {
		result.IsPlayerMiss = true
		result.PlayerMessage = fmt.Sprintf(
			"🎲 [%d]+%d=*%d* vs CA %d ❌ — %s *%s* falhou! %s resistiu.",
			d20, atkBonus, total, monster.CA, skill.Emoji, skill.Name, monster.Name,
		)
	} else {
		var dmg int
		var dmgDesc string
		switch skill.DamageType {
		case "magic":
			base := char.MagicAttack + skill.Damage - int(float64(monster.MagicDef)*0.3)
			if base < 1 {
				base = 1
			}
			dmg = base
			dmgDesc = fmt.Sprintf("🔮 %d mágico", dmg)
		case "poison":
			base := skill.Damage + int(float64(char.Dexterity)*0.3)
			if base < 1 {
				base = 1
			}
			dmg = base
			dmgDesc = fmt.Sprintf("☠️ %d veneno", dmg)
			if skill.PoisonTurnsCount > 0 {
				result.AppliesPoison = true
				result.PoisonDmg = skill.PoisonDmgPerTurn
				result.PoisonTurns = skill.PoisonTurnsCount
			}
		default:
			_, sides, attrMod := ClassDamageRoll(char.Class, char.Strength, char.Dexterity, char.Intelligence)
			dmg = rollDice(1, sides) + attrMod + skill.Damage
			if dmg < 1 {
				dmg = 1
			}
			dmgDesc = fmt.Sprintf("⚔️ %d físico", dmg)
		}
		if char.Race == "elf" && skill.DamageType == "magic" {
			dmg = int(float64(dmg) * 1.2)
		}

		if isCrit {
			dmg *= 2
			result.IsCritical = true
			result.PlayerMessage = fmt.Sprintf(
				"🎲 [20] ⭐ *CRÍTICO!* %s *%s* — %s → *%d* de dano! (dobrado)",
				skill.Emoji, skill.Name, dmgDesc, dmg,
			)
		} else {
			result.PlayerMessage = fmt.Sprintf(
				"🎲 [%d]+%d=*%d* vs CA %d ✅ — %s *%s* → *%d* de dano.",
				d20, atkBonus, total, monster.CA, skill.Emoji, skill.Name, dmg,
			)
		}
		result.PlayerDamage = dmg
	}

	mr := monsterAttack(monster, char, equipCA)
	result.MonsterDamage = mr.MonsterDamage
	result.MonsterMessage = mr.MonsterMessage
	result.IsMonsterMiss = mr.IsMonsterMiss
	result.IsMonsterCrit = mr.IsMonsterCrit
	result.MonsterRoll = mr.MonsterRoll
	return result
}

// =============================================
// ATAQUE DO MONSTRO
// =============================================

func monsterAttack(monster *models.Monster, char *models.Character, equipCA int) CombatResult {
	result := CombatResult{}

	playerCA := calcPlayerCA(char, equipCA)
	atkBonus := MonsterAttackBonus(monster)

	d20 := rollD20()
	result.MonsterRoll = d20
	total := d20 + atkBonus
	isCrit := d20 == 20
	isFumble := d20 == 1

	if isFumble {
		result.IsMonsterMiss = true
		result.MonsterMessage = fmt.Sprintf(
			"🎲 [1] %s *%s* tropeçou — *Fumble!* Ataque desperdiçado.",
			monster.Emoji, monster.Name,
		)
		return result
	}
	if !isCrit && total < playerCA {
		result.IsMonsterMiss = true
		result.MonsterMessage = fmt.Sprintf(
			"🎲 [%d]+%d=*%d* vs CA %d — %s *%s* ❌ *errou!* Você desviou.",
			d20, atkBonus, total, playerCA, monster.Emoji, monster.Name,
		)
		return result
	}

	dice, sides := MonsterDamageDice(monster)
	dmgBonus := monster.Attack / 5
	dmg := rollDice(dice, sides) + dmgBonus
	if dmg < 1 {
		dmg = 1
	}
	if char.Race == "dwarf" {
		dmg = int(float64(dmg) * 0.85)
	}

	if isCrit {
		dmg *= 2
		result.IsMonsterCrit = true
		result.MonsterMessage = fmt.Sprintf(
			"🎲 [20] 💥 *CRÍTICO!* %s *%s* → *%d* de dano! (dobrado)",
			monster.Emoji, monster.Name, dmg,
		)
	} else {
		result.MonsterMessage = fmt.Sprintf(
			"🎲 [%d]+%d=*%d* vs CA %d ✅ — %s *%s* → *%d* de dano.",
			d20, atkBonus, total, playerCA, monster.Emoji, monster.Name, dmg,
		)
	}
	result.MonsterDamage = dmg
	if monster.PoisonChance > 0 && monster.PoisonDmg > 0 && monster.PoisonTurns > 0 && rand.Intn(100) < monster.PoisonChance {
		// Monstros aplicam/renovam veneno quando mais forte que o DoT atual do jogador.
		if char.PoisonTurns <= 0 || monster.PoisonTurns > char.PoisonTurns || monster.PoisonDmg > char.PoisonDmg {
			char.PoisonTurns = monster.PoisonTurns
			char.PoisonDmg = monster.PoisonDmg
			result.MonsterMessage += fmt.Sprintf("\n☠️ *%s envenenou você!* %d dano/turno por %d turnos.",
				monster.Name, monster.PoisonDmg, monster.PoisonTurns)
		}
	}
	return result
}

func skillBuffEffect(skill *models.Skill) string {
	switch skill.ID {
	case "warrior_shield":
		return "🛡️ **Escudo de Ferro** ativado! Dano reduzido por 2 turnos."
	case "warrior_battlecry":
		return "📣 **Grito de Guerra!** Ataque aumentado em 30% por 3 turnos."
	case "warrior_berserker":
		return "💢 **Modo Berserk** ativado! +50% ataque, -20% defesa."
	case "mage_arcane_shield":
		return "🔮 **Escudo Arcano** conjurado! Absorvendo dano mágico."
	case "rogue_shadow_step":
		return "👤 **Passo das Sombras!** Próximo ataque é crítico garantido."
	case "rogue_smoke_bomb":
		return "💨 **Bomba de Fumaça!** Acurácia inimiga reduzida em 50%."
	case "archer_eagle_eye":
		return "🦅 **Olho de Águia!** Dano crítico dobrado por 2 turnos."
	}
	return "✨ Habilidade ativada!"
}

// =============================================
// FUGA / LEVEL UP / REST / DEATH / XP
// =============================================

func TryFlee(char *models.Character, monster *models.Monster) bool {
	speedDiff := char.Speed - monster.Speed
	chance := 40 + speedDiff*5
	if chance < 10 {
		chance = 10
	}
	if chance > 85 {
		chance = 85
	}
	return rand.Intn(100) < chance
}

func CheckLevelUp(char *models.Character) *LevelUpResult {
	if char.Experience < char.ExperienceNext {
		return nil
	}
	if char.Level >= 100 {
		return nil
	}
	class := Classes[char.Class]
	newLevel := char.Level + 1
	return &LevelUpResult{
		NewLevel:    newLevel,
		HPGained:    class.HPPerLevel + char.Constitution/3,
		MPGained:    class.MPPerLevel + char.Intelligence/5,
		SkillPoints: 1, // +1 ponto por nível (total 99 ao longo de lv1→100)
	}
}

func ApplyLevelUp(char *models.Character, result *LevelUpResult) {
	char.Level = result.NewLevel
	char.HPMax += result.HPGained
	char.HP = char.HPMax
	char.MPMax += result.MPGained
	char.MP = char.MPMax
	char.SkillPoints += result.SkillPoints
	char.ExperienceNext = ExperienceForLevel(char.Level + 1)
	char.Attack += 2
	char.Defense += 1
	char.MagicAttack += 1
	char.MagicDefense += 1
	if char.Level%5 == 0 {
		switch char.Class {
		case "warrior":
			char.Strength += 2
			char.Constitution += 1
		case "mage":
			char.Intelligence += 2
			char.Wisdom += 1
		case "rogue":
			char.Dexterity += 2
			char.Charisma += 1
		case "archer":
			char.Dexterity += 2
			char.Wisdom += 1
		case "paladin":
			char.Strength += 1
			char.Constitution += 1
			char.Charisma += 1
		case "cleric":
			char.Wisdom += 2
			char.Constitution += 1
		case "barbarian":
			char.Strength += 3
			char.Constitution += 1
		case "bard":
			char.Charisma += 2
			char.Dexterity += 1
		}
	}
}

func RestAtInn(char *models.Character) (cost int, canAfford bool) {
	cost = 10 + char.Level*2
	canAfford = char.Gold >= cost
	return
}

func ApplyDeathPenalty(char *models.Character) (xpLost, goldLost int) {
	xpLost = char.Experience / 10
	goldLost = char.Gold / 10
	char.Experience -= xpLost
	if char.Experience < 0 {
		char.Experience = 0
	}
	char.Gold -= goldLost
	if char.Gold < 0 {
		char.Gold = 0
	}
	char.HP = char.HPMax * 30 / 100
	if char.HP < 1 {
		char.HP = 1
	}
	char.MP = char.MPMax / 2
	char.State = "idle"
	char.CombatMonsterID = ""
	char.CombatMonsterHP = 0
	char.CurrentMap = "village"
	return
}

func CalculateXPGain(char *models.Character, monster *models.Monster) int {
	xp := monster.ExpReward
	levelDiff := char.Level - monster.Level
	if levelDiff > 5 {
		xp = int(float64(xp) * math.Pow(0.8, float64(levelDiff-5)))
	} else if levelDiff < -3 {
		xp = int(float64(xp) * 1.3)
	}
	if char.Race == "human" {
		xp = int(float64(xp) * 1.1)
	}
	// Bênção do Sábio: +50% XP se boost ativo
	if !char.XPBoostExpiry.IsZero() && time.Now().Before(char.XPBoostExpiry) {
		xp = int(float64(xp) * 1.5)
	}
	if xp < 1 {
		xp = 1
	}
	return xp
}

func RollDiamondDrop(chance int) bool {
	if chance <= 0 {
		return false
	}
	return rand.Intn(100) < chance
}

// ApplyMonsterPoisonDoT aplica o DoT de veneno no monstro.
// Reduz o turno restante e retorna o dano causado + mensagem de log.
func ApplyMonsterPoisonDoT(char *models.Character, monster *models.Monster) (int, string) {
	if char.CombatMonsterPoisonTurns <= 0 {
		return 0, ""
	}
	dmg := char.CombatMonsterPoisonDmg
	char.CombatMonsterPoisonTurns--
	if char.CombatMonsterPoisonTurns == 0 {
		char.CombatMonsterPoisonDmg = 0
	}
	remaining := char.CombatMonsterPoisonTurns
	msg := fmt.Sprintf("☠️ *Veneno!* %s sofre *%d* de dano (%d turn(s) restante(s)).\n", monster.Name, dmg, remaining)
	return dmg, msg
}

// ApplyPlayerPoisonDoT aplica o DoT de veneno no player.
// Retorna o dano causado + mensagem de log.
func ApplyPlayerPoisonDoT(char *models.Character) (int, string) {
	if char.PoisonTurns <= 0 {
		return 0, ""
	}
	dmg := char.PoisonDmg
	char.PoisonTurns--
	if char.PoisonTurns == 0 {
		char.PoisonDmg = 0
	}
	remaining := char.PoisonTurns
	msg := fmt.Sprintf("☠️ *Você está envenenado!* -%d HP (%d turn(s) restante(s)).\n", dmg, remaining)
	return dmg, msg
}
