package handlers

import (
	"fmt"
	"math/rand"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/tormenta-bot/internal/database"
	"github.com/tormenta-bot/internal/game"
	menukit "github.com/tormenta-bot/internal/menu"
	"github.com/tormenta-bot/internal/models"
)

// pvpChallengeTarget stores the target charID selected via button (userID -> charID)
var pvpChallengeTarget = map[int64]int{}

// =============================================
// MENU PVP
// =============================================

func showPVPMenu(chatID int64, msgID int, userID int64) {
	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}
	game.TickEnergy(char)
	database.SaveCharacter(char)

	stats, _ := database.GetOrCreatePVPStats(char.ID)
	rank := game.PVPRankTitle(stats.Rating)

	// Desafio pendente
	pending, _ := database.GetPendingChallenge(char.ID)
	pendingSection := ""
	hasPending := false
	pendingMatchID := 0
	if pending != nil {
		challenger, _ := database.GetCharacterByID(pending.ChallengerID)
		if challenger != nil {
			stakeStr := "sem aposta"
			if pending.StakeGold > 0 {
				stakeStr = fmt.Sprintf("aposta: *%d* 🪙", pending.StakeGold)
			}
			pendingSection = fmt.Sprintf("⚔️ *Desafio Recebido!*\n%s Nv.*%d* desafia você (%s)\n\n",
				challenger.Name, challenger.Level, stakeStr)
			hasPending = true
			pendingMatchID = pending.ID
		}
	}

	// Partida ativa
	active, _ := database.GetActivePVPMatch(char.ID)

	caption := fmt.Sprintf(
		"⚔️ *Arena PVP*\n\n"+
			"🏅 *%s*\n"+
			"📊 Rating: *%d* | W: *%d* L: *%d* D: *%d*\n"+
			"🔥 Sequência: *%d* | Melhor: *%d*\n\n"+
			"%s"+
			"*Como funciona:*\n"+
			"• Escolha um jogador da lista e desafie\n"+
			"• Rolagem d20 + bônus vs CA do oponente\n"+
			"• Use habilidades e itens no combate\n"+
			"• Vencedor leva a aposta em ouro",
		rank, stats.Rating, stats.Wins, stats.Losses, stats.Draws,
		stats.Streak, stats.BestStreak,
		pendingSection,
	)

	kb := menukit.PVPMenu(menukit.PVPMenuOptions{
		HasPending:     hasPending,
		PendingMatchID: pendingMatchID,
		HasActive:      active != nil,
	})
	editPhoto(chatID, msgID, "combat", caption, &kb)
}

// =============================================
// LISTA DE JOGADORES
// =============================================

func showPVPPlayerList(chatID int64, msgID int, userID int64) {
	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}

	players, err := database.GetRecentPlayers(char.ID, 12)
	if err != nil || len(players) == 0 {
		editPhoto(chatID, msgID, "combat",
			"⚔️ *Desafiar Jogador*\n\n_Nenhum outro jogador encontrado ainda._\n\nConvide amigos para jogar!",
			bkp("menu_pvp"))
		return
	}

	caption := "⚔️ *Escolha o jogador para desafiar:*\n\n"
	var rows [][]tgbotapi.InlineKeyboardButton

	for i := 0; i < len(players); i++ {
		p := players[i]
		pvpStats, _ := database.GetOrCreatePVPStats(p.ID)
		re := pvpRankEmoji(pvpStats.Rating)
		defCA := game.CharacterCA(p.Class,
			game.DefensiveAttr(p.Class, p.Constitution, p.Dexterity, p.Intelligence),
			p.EquipCABonus)
		caption += fmt.Sprintf("%s%s *%s* — Nv.*%d* | %s | CA:%d | R:%d\n",
			re, game.Races[p.Race].Emoji, p.Name, p.Level,
			game.Classes[p.Class].Name, defCA, pvpStats.Rating)

		btn := tgbotapi.NewInlineKeyboardButtonData(
			fmt.Sprintf("%s%s %s Nv.%d", re, game.Races[p.Race].Emoji, p.Name, p.Level),
			fmt.Sprintf("pvp_select_%d", p.ID),
		)

		if i+1 < len(players) {
			p2 := players[i+1]
			pvpStats2, _ := database.GetOrCreatePVPStats(p2.ID)
			re2 := pvpRankEmoji(pvpStats2.Rating)
			defCA2 := game.CharacterCA(p2.Class,
				game.DefensiveAttr(p2.Class, p2.Constitution, p2.Dexterity, p2.Intelligence),
				p2.EquipCABonus)
			caption += fmt.Sprintf("%s%s *%s* — Nv.*%d* | %s | CA:%d | R:%d\n",
				re2, game.Races[p2.Race].Emoji, p2.Name, p2.Level,
				game.Classes[p2.Class].Name, defCA2, pvpStats2.Rating)
			btn2 := tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("%s%s %s Nv.%d", re2, game.Races[p2.Race].Emoji, p2.Name, p2.Level),
				fmt.Sprintf("pvp_select_%d", p2.ID),
			)
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(btn, btn2))
			i++
		} else {
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(btn))
		}
	}

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("⬅️ Voltar", "menu_pvp"),
	))
	kb := tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}
	editPhoto(chatID, msgID, "combat", caption, &kb)
}

func handlePVPSelectPlayer(chatID int64, msgID int, userID int64, targetCharID int) {
	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}
	target, _ := database.GetCharacterByID(targetCharID)
	if target == nil {
		editPhoto(chatID, msgID, "combat", "❌ Jogador não encontrado!", bkp("pvp_player_list"))
		return
	}
	if target.ID == char.ID {
		editPhoto(chatID, msgID, "combat", "❌ Você não pode se desafiar!", bkp("pvp_player_list"))
		return
	}

	pvpChallengeTarget[userID] = target.ID
	stakeOptions := game.PVPStakeOptions(char)
	stats, _ := database.GetOrCreatePVPStats(target.ID)
	defCA := game.CharacterCA(target.Class,
		game.DefensiveAttr(target.Class, target.Constitution, target.Dexterity, target.Intelligence),
		target.EquipCABonus)
	myAtkBonus := game.CharacterAttackBonus(char.Class, char.Level,
		char.Strength, char.Dexterity, char.Intelligence) + char.EquipHitBonus

	caption := fmt.Sprintf(
		"⚔️ *Desafiar %s?*\n\n"+
			"%s Nv.*%d* | %s %s\n"+
			"🏅 %s (Rating: *%d*)\n\n"+
			"🛡️ CA do oponente: *%d* | Seu ataque: *+%d*\n\n"+
			"*Escolha a aposta:*",
		target.Name,
		game.Races[target.Race].Emoji, target.Level,
		game.Classes[target.Class].Emoji, game.Classes[target.Class].Name,
		game.PVPRankTitle(stats.Rating), stats.Rating,
		defCA, myAtkBonus,
	)

	var rows [][]tgbotapi.InlineKeyboardButton
	for _, stake := range stakeOptions {
		label := "Sem aposta"
		if stake > 0 {
			label = fmt.Sprintf("Apostar %d 🪙", stake)
		}
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(label, fmt.Sprintf("pvp_stake_%d", stake)),
		))
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("⬅️ Voltar à lista", "pvp_player_list"),
	))
	kb := tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}
	editPhoto(chatID, msgID, "combat", caption, &kb)
}

func handlePVPSendChallenge(chatID int64, msgID int, userID int64, stakeGold int) {
	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}
	targetID := pvpChallengeTarget[userID]
	if targetID == 0 {
		showPVPMenu(chatID, msgID, userID)
		return
	}
	delete(pvpChallengeTarget, userID)

	if stakeGold > 0 && char.Gold < stakeGold {
		editPhoto(chatID, msgID, "combat",
			fmt.Sprintf("❌ Ouro insuficiente para apostar *%d* 🪙!", stakeGold), bkp("menu_pvp"))
		return
	}
	target, _ := database.GetCharacterByID(targetID)
	if target == nil {
		editPhoto(chatID, msgID, "combat", "❌ Oponente não encontrado!", bkp("menu_pvp"))
		return
	}

	match, err := database.CreatePVPChallenge(char.ID, targetID, stakeGold, char.HP, target.HP)
	if err != nil {
		return
	}

	stakeStr := "sem aposta"
	if stakeGold > 0 {
		stakeStr = fmt.Sprintf("aposta de *%d* 🪙", stakeGold)
	}
	editPhoto(chatID, msgID, "combat",
		fmt.Sprintf("⚔️ *Desafio enviado para %s!*\n\n%s\n\n_Expira em 5 minutos._", target.Name, stakeStr),
		bkp("menu_pvp"))
	notifyPVPChallenge(target.PlayerID, char, stakeGold, match.ID)
}

func notifyPVPChallenge(defenderTelegramID int64, challenger *models.Character, stakeGold int, matchID int) {
	stakeStr := "sem aposta"
	if stakeGold > 0 {
		stakeStr = fmt.Sprintf("aposta: *%d* 🪙", stakeGold)
	}
	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ Aceitar", fmt.Sprintf("pvp_accept_%d", matchID)),
			tgbotapi.NewInlineKeyboardButtonData("❌ Recusar", fmt.Sprintf("pvp_decline_%d", matchID)),
		),
	)
	sendMsg(defenderTelegramID, fmt.Sprintf(
		"⚔️ *DESAFIO PVP!*\n\n*%s* (Nv.%d) desafia você!\n%s\n\nExpira em 5 minutos!",
		challenger.Name, challenger.Level, stakeStr), &kb)
}

func handlePVPAccept(chatID int64, msgID int, userID int64, matchID int) {
	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}
	database.AcceptPVPChallenge(matchID)
	match, _ := database.GetActivePVPMatch(char.ID)
	if match == nil {
		editPhoto(chatID, msgID, "combat", "❌ Partida não encontrada ou expirada!", bkp("menu_pvp"))
		return
	}
	renderPVPCombat(chatID, msgID, char, match, "⚔️ *Duelo iniciado!* Que vença o melhor!\n")
}

func handlePVPDecline(chatID int64, msgID int, _ int64, matchID int) {
	database.DeclinePVPChallenge(matchID)
	editPhoto(chatID, msgID, "combat", "❌ *Desafio recusado.*\n\nSeu oponente foi notificado.", bkp("menu_pvp"))
}

func handlePVPContinue(chatID int64, msgID int, userID int64) {
	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}
	match, _ := database.GetActivePVPMatch(char.ID)
	if match == nil {
		showPVPMenu(chatID, msgID, userID)
		return
	}
	renderPVPCombat(chatID, msgID, char, match, "")
}

// =============================================
// COMBATE — ATAQUE BÁSICO
// =============================================

func handlePVPAttackTurn(chatID int64, msgID int, userID int64) {
	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}
	game.TickEnergy(char)
	database.SaveCharacter(char)

	match, _ := database.GetActivePVPMatch(char.ID)
	if match == nil {
		showPVPMenu(chatID, msgID, userID)
		return
	}

	isChallenger := match.ChallengerID == char.ID
	if !isMyTurn(match, isChallenger) {
		renderPVPCombat(chatID, msgID, char, match, "⏳ *Aguardando o oponente atacar...*\n")
		return
	}
	selfFx := getPVPEffects(char.ID)
	if selfFx.ConsumeSkipOwnTurn() {
		// Perde o turno e passa a vez.
		nextTurn := 1
		if match.Turn == 1 {
			nextTurn = 0
		}
		database.UpdatePVPMatch(match.ID, nextTurn, match.ChallengerHP, match.DefenderHP)
		match.Turn = nextTurn
		renderPVPCombat(chatID, msgID, char, match, "❄️ *Você perdeu este turno por efeito de controle!*\n")
		return
	}

	opponent := pvpOpponent(match, isChallenger)
	if opponent == nil {
		return
	}

	// Aplica DoT de veneno no player antes do ataque
	dotLog := ""
	if char.PoisonTurns > 0 {
		dotDmg, msg := game.ApplyPlayerPoisonDoT(char)
		if isChallenger {
			match.ChallengerHP -= dotDmg
		} else {
			match.DefenderHP -= dotDmg
		}
		dotLog = msg
		database.SaveCharacter(char)
	}
	if burnDmg, burnMsg := selfFx.ApplyEnemyDot(); burnDmg > 0 {
		if isChallenger {
			match.ChallengerHP -= burnDmg
		} else {
			match.DefenderHP -= burnDmg
		}
		dotLog += burnMsg
	}

	attFx := getPVPEffects(char.ID)
	defFx := getPVPEffects(opponent.ID)
	attCalc := *char
	defCalc := *opponent
	attCalc.EquipHitBonus -= attFx.EffectiveAtkPenalty()
	defCalc.EquipCABonus = defCalc.EquipCABonus + defFx.EffectiveCABonus() - defFx.EffectiveCAPenalty()

	result := game.PVPAttack(&attCalc, &defCalc)
	if result.Damage > 0 && !result.IsMiss && !result.IsCritical && attFx.ConsumeForceCrit() {
		result.Damage *= 2
		result.IsCritical = true
		result.Message += "\n👤 *Crítico garantido ativado!*"
	}
	damage := attFx.ApplyOutgoingDamage(result.Damage, result.Damage > 0)
	damage = defFx.ApplyIncomingDamage(damage)
	attFx.AdvanceTurn()
	defFx.AdvanceTurn()
	processPVPResult(chatID, msgID, char, opponent, match, isChallenger, damage, dotLog+result.Message+"\n")
}

// =============================================
// COMBATE — HABILIDADE
// =============================================

func handlePVPSkillTurn(chatID int64, msgID int, userID int64, skillID string) {
	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}
	game.TickEnergy(char)

	match, _ := database.GetActivePVPMatch(char.ID)
	if match == nil {
		showPVPMenu(chatID, msgID, userID)
		return
	}

	isChallenger := match.ChallengerID == char.ID
	if !isMyTurn(match, isChallenger) {
		renderPVPCombat(chatID, msgID, char, match, "⏳ *Aguardando o oponente atacar...*\n")
		return
	}
	selfFx := getPVPEffects(char.ID)
	if selfFx.ConsumeSkipOwnTurn() {
		nextTurn := 1
		if match.Turn == 1 {
			nextTurn = 0
		}
		database.UpdatePVPMatch(match.ID, nextTurn, match.ChallengerHP, match.DefenderHP)
		match.Turn = nextTurn
		renderPVPCombat(chatID, msgID, char, match, "❄️ *Você perdeu este turno por efeito de controle!*\n")
		return
	}

	sk, ok := game.Skills[skillID]
	if !ok || sk.Passive {
		renderPVPCombat(chatID, msgID, char, match, "❌ Habilidade inválida!\n")
		return
	}
	if skillRequiresShield(skillID) && !hasEquippedShield(char.ID) {
		renderPVPCombat(chatID, msgID, char, match, "🛡️ *Esta habilidade exige escudo equipado no slot Escudo.*\n")
		return
	}
	if char.MP < sk.MPCost {
		renderPVPCombat(chatID, msgID, char, match,
			fmt.Sprintf("❌ *MP insuficiente!* Precisa *%d* MP | você tem *%d*.\n", sk.MPCost, char.MP))
		return
	}

	char.MP -= sk.MPCost

	// Aplica DoT de veneno no player antes do ataque
	dotLog := ""
	if char.PoisonTurns > 0 {
		dotDmg, msg := game.ApplyPlayerPoisonDoT(char)
		if isChallenger {
			match.ChallengerHP -= dotDmg
		} else {
			match.DefenderHP -= dotDmg
		}
		dotLog = msg
	}
	if burnDmg, burnMsg := selfFx.ApplyEnemyDot(); burnDmg > 0 {
		if isChallenger {
			match.ChallengerHP -= burnDmg
		} else {
			match.DefenderHP -= burnDmg
		}
		dotLog += burnMsg
	}
	database.SaveCharacter(char)

	opponent := pvpOpponent(match, isChallenger)
	if opponent == nil {
		return
	}

	attFx := getPVPEffects(char.ID)
	defFx := getPVPEffects(opponent.ID)
	attCalc := *char
	defCalc := *opponent
	attCalc.EquipHitBonus -= attFx.EffectiveAtkPenalty()
	defCalc.EquipCABonus = defCalc.EquipCABonus + defFx.EffectiveCABonus() - defFx.EffectiveCAPenalty()

	result := game.PVPSkillAttack(&attCalc, &sk, &defCalc)
	critMin := 21
	switch sk.ID {
	case "w_rampage", "m_meteor", "a_headshot":
		critMin = 18
	case "r_vital_strike":
		critMin = 17
	case "w_cleave":
		critMin = 19
	}
	if critMin <= 20 && result.D20Roll >= critMin && result.Damage > 0 && !result.IsMiss && !result.IsCritical {
		result.Damage *= 2
		result.IsCritical = true
		result.Message += fmt.Sprintf("\n⭐ Crítico especial (%d-20) ativado.", critMin)
	}
	if sk.ID == "r_phantom" && result.IsMiss {
		base := sk.Damage + int(float64(char.Dexterity)*0.3)
		if base < 1 {
			base = 1
		}
		result.IsMiss = false
		result.Damage = base
		result.Message += "\n👻 *Forma Fantasma:* ataque atravessa defesa e acerta."
	}
	// Execução em alvo com HP baixo.
	targetHP := match.DefenderHP
	targetMaxHP := 1
	if isChallenger {
		targetHP = match.DefenderHP
		if opponent.HPMax > 0 {
			targetMaxHP = opponent.HPMax
		}
	} else {
		targetHP = match.ChallengerHP
		if opponent.HPMax > 0 {
			targetMaxHP = opponent.HPMax
		}
	}
	if sk.ID == "r_death_blow" && targetHP <= targetMaxHP/4 && result.Damage > 0 && !result.IsMiss {
		result.Damage *= 3
		result.Message += "\n💀 *Execução!* Alvo abaixo de 25% HP: dano x3."
	}
	switch sk.ID {
	case "a_multishot":
		if result.Damage > 0 && !result.IsMiss {
			result.Damage *= 3
			result.Message += "\n🌧️ *Chuva de Flechas:* 3 acertos."
		}
	case "a_volley":
		if result.Damage > 0 && !result.IsMiss {
			result.Damage *= 5
			result.Message += "\n⛈️ *Saraivada:* 5 acertos."
		}
	case "a_quick_shot":
		if result.Damage > 0 && !result.IsMiss && rand.Intn(100) < 30 {
			result.Damage *= 2
			result.Message += "\n🏹 *Tiro Rápido:* segundo disparo ativado!"
		}
	}
	if sk.ID == "m_arcane_burst" && attFx.EffectiveCABonus() > 0 && result.Damage > 0 {
		result.Damage *= 2
		result.Message += "\n💫 *Escudo Arcano ativo:* dano dobrado!"
	}
	if result.Damage > 0 && !result.IsMiss && !result.IsCritical && attFx.ConsumeForceCrit() {
		result.Damage *= 2
		result.IsCritical = true
		result.Message += "\n👤 *Crítico garantido ativado!*"
	}
	// Se a skill envenena, aplica ao oponente
	if result.AppliesPoison && !result.IsMiss {
		opponent.PoisonTurns = result.PoisonTurns
		opponent.PoisonDmg = result.PoisonDmg
		database.SaveCharacter(opponent)
	}
	effectMsg := ""
	if !result.IsMiss {
		if sk.ID == "w_blood_rage" && char.HPMax > 0 && (char.HP*100/char.HPMax) >= 30 {
			effectMsg = "🩸 *Fúria Sangrenta* requer HP abaixo de 30%.\n"
		} else {
			effectMsg = formatEffectMsg(applySkillEffectsPVP(sk.ID, attFx, defFx))
		}
	}
	damage := attFx.ApplyOutgoingDamage(result.Damage, result.Damage > 0)
	damage = defFx.ApplyIncomingDamage(damage)
	attFx.AdvanceTurn()
	defFx.AdvanceTurn()
	processPVPResult(chatID, msgID, char, opponent, match, isChallenger, damage, dotLog+result.Message+"\n"+effectMsg)
}

// =============================================
// COMBATE — ITEM (não gasta turno)
// =============================================

func showPVPItemMenu(chatID int64, msgID int, userID int64) {
	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}
	match, _ := database.GetActivePVPMatch(char.ID)
	if match == nil {
		showPVPMenu(chatID, msgID, userID)
		return
	}

	invItems, _ := database.GetInventory(char.ID)
	caption := "🎒 *Usar Item no PVP*\n\n_Itens não gastam turno — você ainda ataca depois._\n\n"
	var rows [][]tgbotapi.InlineKeyboardButton
	count := 0

	for _, inv := range invItems {
		if inv.ItemType != "consumable" {
			continue
		}
		item, ok := game.Items[inv.ItemID]
		if !ok {
			continue
		}
		var effects []string
		if item.HealHP > 0 {
			effects = append(effects, fmt.Sprintf("+%dHP", item.HealHP))
		}
		if item.HealMP > 0 {
			effects = append(effects, fmt.Sprintf("+%dMP", item.HealMP))
		}
		if item.RestoreEnergy > 0 {
			effects = append(effects, fmt.Sprintf("+%d⚡", item.RestoreEnergy))
		}
		caption += fmt.Sprintf("%s *%s* ×%d — %s\n", item.Emoji, item.Name, inv.Quantity, strings.Join(effects, " "))
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("Usar %s %s (×%d)", item.Emoji, item.Name, inv.Quantity),
				"pvp_item_"+item.ID,
			),
		))
		count++
	}
	if count == 0 {
		caption += "_Sem consumíveis no inventário._"
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("⬅️ Voltar ao Combate", "pvp_continue"),
	))
	kb := tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}
	editPhoto(chatID, msgID, "combat", caption, &kb)
}

func handlePVPUseItem(chatID int64, msgID int, userID int64, itemID string) {
	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}
	match, _ := database.GetActivePVPMatch(char.ID)
	if match == nil {
		showPVPMenu(chatID, msgID, userID)
		return
	}

	item, ok := game.Items[itemID]
	if !ok || database.GetItemCount(char.ID, itemID) <= 0 {
		renderPVPCombat(chatID, msgID, char, match, "❌ Item não encontrado no inventário!\n")
		return
	}

	var effects []string
	if item.HealHP > 0 {
		old := char.HP
		char.HP += item.HealHP
		if char.HP > char.HPMax {
			char.HP = char.HPMax
		}
		gained := char.HP - old
		if gained > 0 {
			effects = append(effects, fmt.Sprintf("+%d❤️", gained))
		}
	}
	if item.HealMP > 0 {
		old := char.MP
		char.MP += item.HealMP
		if char.MP > char.MPMax {
			char.MP = char.MPMax
		}
		gained := char.MP - old
		if gained > 0 {
			effects = append(effects, fmt.Sprintf("+%d💙", gained))
		}
	}
	if item.RestoreEnergy > 0 {
		old := char.Energy
		char.Energy += item.RestoreEnergy
		if char.Energy > char.EnergyMax {
			char.Energy = char.EnergyMax
		}
		gained := char.Energy - old
		if gained > 0 {
			effects = append(effects, fmt.Sprintf("+%d⚡", gained))
		}
	}

	database.RemoveItem(char.ID, itemID, 1)
	database.SaveCharacter(char)

	// Atualiza HP na partida se curou HP
	if item.HealHP > 0 {
		isChallenger := match.ChallengerID == char.ID
		newCHP, newDHP := match.ChallengerHP, match.DefenderHP
		if isChallenger {
			newCHP = char.HP
		} else {
			newDHP = char.HP
		}
		database.UpdatePVPMatch(match.ID, match.Turn, newCHP, newDHP)
		match.ChallengerHP = newCHP
		match.DefenderHP = newDHP
	}

	effectStr := strings.Join(effects, " ")
	renderPVPCombat(chatID, msgID, char, match,
		fmt.Sprintf("%s *%s* usada! %s _(não gasta turno)_\n", item.Emoji, item.Name, effectStr))
}

// =============================================
// LÓGICA CENTRAL DO TURNO
// =============================================

func processPVPResult(chatID int64, msgID int, char *models.Character, opponent *models.Character,
	match *database.PVPChallenge, isChallenger bool, damage int, log string) {

	cHP := match.ChallengerHP
	dHP := match.DefenderHP
	if isChallenger {
		dHP -= damage
	} else {
		cHP -= damage
	}
	if cHP < 0 {
		cHP = 0
	}
	if dHP < 0 {
		dHP = 0
	}

	if cHP <= 0 || dHP <= 0 {
		winnerID := match.ChallengerID
		loserID := match.DefenderID
		if cHP <= 0 {
			winnerID, loserID = match.DefenderID, match.ChallengerID
		}
		database.FinishPVPMatch(match.ID, winnerID)
		var winner, loser *models.Character
		if winnerID == char.ID {
			winner, loser = char, opponent
		} else {
			winner, loser = opponent, char
		}
		handlePVPEnd(chatID, msgID, char, winner, loser, match, winnerID, loserID, log)
		return
	}

	nextTurn := 1
	if match.Turn == 1 {
		nextTurn = 0
	}
	database.UpdatePVPMatch(match.ID, nextTurn, cHP, dHP)
	match.Turn = nextTurn
	match.ChallengerHP = cHP
	match.DefenderHP = dHP

	renderPVPCombat(chatID, msgID, char, match, log)
	notifyPVPTurn(opponent.PlayerID, char.Name)
}

// =============================================
// FIM DO COMBATE
// =============================================

func handlePVPEnd(chatID int64, msgID int, viewingChar *models.Character,
	winner, loser *models.Character, match *database.PVPChallenge, winnerID, loserID int, log string) {

	winnerStats, _ := database.GetOrCreatePVPStats(winnerID)
	loserStats, _ := database.GetOrCreatePVPStats(loserID)
	oldWR := winnerStats.Rating
	oldLR := loserStats.Rating
	newWR, newLR := game.CalculateELO(winnerStats.Rating, loserStats.Rating)

	winnerStats.Wins++
	winnerStats.Streak++
	if winnerStats.Streak > winnerStats.BestStreak {
		winnerStats.BestStreak = winnerStats.Streak
	}
	loserStats.Losses++
	loserStats.Streak = 0

	database.UpdatePVPStats(winnerID, winnerStats.Wins, winnerStats.Losses, winnerStats.Draws,
		newWR, winnerStats.Streak, winnerStats.BestStreak)
	database.UpdatePVPStats(loserID, loserStats.Wins, loserStats.Losses, loserStats.Draws,
		newLR, loserStats.Streak, loserStats.BestStreak)

	// Limpa veneno de ambos ao fim do PvP
	winner.PoisonTurns = 0
	winner.PoisonDmg = 0
	loser.PoisonTurns = 0
	loser.PoisonDmg = 0
	resetPVPEffects(winner.ID)
	resetPVPEffects(loser.ID)
	database.SaveCharacter(winner)
	database.SaveCharacter(loser)

	goldMsg := ""
	if match.StakeGold > 0 {
		w, _ := database.GetCharacterByID(winnerID)
		l, _ := database.GetCharacterByID(loserID)
		if w != nil && l != nil {
			w.Gold += match.StakeGold
			l.Gold -= match.StakeGold
			if l.Gold < 0 {
				l.Gold = 0
			}
			database.SaveCharacter(w)
			database.SaveCharacter(l)
			goldMsg = fmt.Sprintf("\n🪙 *+%d ouro* transferido ao vencedor!", match.StakeGold)
		}
	}

	isWinner := viewingChar.ID == winnerID
	oldRating, newRating := oldLR, newLR
	if isWinner {
		oldRating, newRating = oldWR, newWR
	}
	ratingChange := newRating - oldRating
	ratingStr := fmt.Sprintf("+%d", ratingChange)
	if ratingChange < 0 {
		ratingStr = fmt.Sprintf("%d", ratingChange)
	}

	resultEmoji, resultText := "💀", "*DERROTA!*"
	if isWinner {
		resultEmoji, resultText = "🏆", "*VITÓRIA!*"
	}

	caption := log + fmt.Sprintf(
		"\n%s %s\n\n⚔️ *%s* derrotou *%s*!\n\n📊 Rating: *%d* → *%d* (%s)%s",
		resultEmoji, resultText, winner.Name, loser.Name,
		oldRating, newRating, ratingStr, goldMsg,
	)

	imgKey := "defeat"
	if isWinner {
		imgKey = "victory"
	}

	editPhoto(chatID, msgID, imgKey, caption, &tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
			{tgbotapi.NewInlineKeyboardButtonData("⚔️ Arena PVP", "menu_pvp")},
			{tgbotapi.NewInlineKeyboardButtonData("🏆 Ranking", "menu_rank"),
				tgbotapi.NewInlineKeyboardButtonData("🏰 Menu", "menu_main")},
		},
	})

	if loser.PlayerID != viewingChar.PlayerID {
		notifyPVPResult(loser.PlayerID, winner.Name, false, match.StakeGold, newLR)
	}
	if winner.PlayerID != viewingChar.PlayerID {
		notifyPVPResult(winner.PlayerID, loser.Name, true, match.StakeGold, newWR)
	}
}

func notifyPVPTurn(telegramID int64, opponentName string) {
	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⚔️ Ver Combate", "pvp_continue"),
		),
	)
	sendMsg(telegramID, fmt.Sprintf("⚔️ *É sua vez!*\n\n*%s* atacou. Responda!", opponentName), &kb)
}

func notifyPVPResult(telegramID int64, opponentName string, won bool, stakeGold int, newRating int) {
	if won {
		sendMsg(telegramID, fmt.Sprintf("🏆 *Você venceu %s!*\nNovo rating: *%d*", opponentName, newRating), nil)
	} else {
		msg := fmt.Sprintf("💀 *%s te derrotou!*\nNovo rating: *%d*", opponentName, newRating)
		if stakeGold > 0 {
			msg += fmt.Sprintf("\n🪙 -%d ouro", stakeGold)
		}
		sendMsg(telegramID, msg, nil)
	}
}

// =============================================
// RENDER DO COMBATE PVP
// =============================================

func renderPVPCombat(chatID int64, msgID int, char *models.Character, match *database.PVPChallenge, log string) {
	isChallenger := match.ChallengerID == char.ID
	myHP, opponentHP := match.ChallengerHP, match.DefenderHP
	opponentID := match.DefenderID
	if !isChallenger {
		myHP, opponentHP = match.DefenderHP, match.ChallengerHP
		opponentID = match.ChallengerID
	}

	opponent, _ := database.GetCharacterByID(opponentID)
	opponentName := "Oponente"
	opponentMaxHP := opponentHP + 1
	var opponentClass, opponentRace string
	if opponent != nil {
		opponentName = opponent.Name
		opponentMaxHP = opponent.HPMax
		opponentClass = opponent.Class
		opponentRace = opponent.Race
	}

	myTurn := isMyTurn(match, isChallenger)

	myHPPct := int(float64(myHP) / float64(char.HPMax) * 8)
	opHPPct := int(float64(opponentHP) / float64(opponentMaxHP) * 8)
	if myHPPct < 0 {
		myHPPct = 0
	}
	if opHPPct < 0 {
		opHPPct = 0
	}

	myBar := strings.Repeat("❤️", myHPPct) + strings.Repeat("🖤", 8-myHPPct)
	opBar := strings.Repeat("💜", opHPPct) + strings.Repeat("🖤", 8-opHPPct)

	// Informações de combate: CA e bônus de ataque
	myAtkBonus := game.CharacterAttackBonus(char.Class, char.Level,
		char.Strength, char.Dexterity, char.Intelligence) + char.EquipHitBonus
	opCA := 10
	myCA := game.CharacterCA(char.Class,
		game.DefensiveAttr(char.Class, char.Constitution, char.Dexterity, char.Intelligence),
		char.EquipCABonus)
	if opponent != nil {
		opCA = game.CharacterCA(opponentClass,
			game.DefensiveAttr(opponentClass, opponent.Constitution, opponent.Dexterity, opponent.Intelligence),
			opponent.EquipCABonus)
	}
	_ = opponentRace

	turnMsg := "⏳ *Aguardando oponente...*"
	if myTurn {
		turnMsg = "⚔️ *SUA VEZ DE ATACAR!*"
	}

	stakeStr := ""
	if match.StakeGold > 0 {
		stakeStr = fmt.Sprintf(" | 🪙 Aposta: *%d*", match.StakeGold)
	}

	caption := fmt.Sprintf(
		"⚔️ *Arena PVP*%s\n\n"+
			"👤 *%s* | CA:*%d* | Atq:*+%d*\n"+
			"%s %d/%d HP | 💙 %d/%d MP\n\n"+
			"👤 *%s* | CA:*%d*\n"+
			"%s %d/%d HP\n\n"+
			"%s\n\n%s",
		stakeStr,
		char.Name, myCA, myAtkBonus,
		myBar, myHP, char.HPMax, char.MP, char.MPMax,
		opponentName, opCA,
		opBar, opponentHP, opponentMaxHP,
		turnMsg,
		truncateCombatLog(log, 5),
	)

	rows := [][]tgbotapi.InlineKeyboardButton{}
	if myTurn {
		// Ataque básico
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⚔️ Atacar", "pvp_attack"),
		))
		// Habilidades aprendidas (não passivas)
		learnedSkills, _ := database.GetLearnedSkills(char.ID)
		var skillBtns []tgbotapi.InlineKeyboardButton
		for _, ls := range learnedSkills {
			sk := game.Skills[ls.SkillID]
			if sk.Passive {
				continue
			}
			label := fmt.Sprintf("%s%s", sk.Emoji, sk.Name)
			if sk.MPCost > 0 {
				label += fmt.Sprintf("(%dMP)", sk.MPCost)
			}
			skillBtns = append(skillBtns, tgbotapi.NewInlineKeyboardButtonData(label, "pvp_skill_"+sk.ID))
		}
		for i := 0; i < len(skillBtns); i += 2 {
			if i+1 < len(skillBtns) {
				rows = append(rows, tgbotapi.NewInlineKeyboardRow(skillBtns[i], skillBtns[i+1]))
			} else {
				rows = append(rows, tgbotapi.NewInlineKeyboardRow(skillBtns[i]))
			}
		}
	}

	// Itens sempre disponíveis
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("🎒 Usar Item", "pvp_item_menu"),
	))
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("🏰 Menu", "menu_main"),
	))

	kb := tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}
	editPhoto(chatID, msgID, "combat", caption, &kb)
}

// =============================================
// HELPERS INTERNOS
// =============================================

func isMyTurn(match *database.PVPChallenge, isChallenger bool) bool {
	return (match.Turn == 0 && isChallenger) || (match.Turn == 1 && !isChallenger)
}

func pvpOpponent(match *database.PVPChallenge, isChallenger bool) *models.Character {
	opponentID := match.DefenderID
	if !isChallenger {
		opponentID = match.ChallengerID
	}
	opponent, _ := database.GetCharacterByID(opponentID)
	return opponent
}

func pvpRankEmoji(rating int) string {
	switch {
	case rating >= 2000:
		return "👑"
	case rating >= 1600:
		return "💎"
	case rating >= 1400:
		return "🥇"
	case rating >= 1200:
		return "🥈"
	case rating >= 1100:
		return "🥉"
	case rating >= 1000:
		return "⚔️"
	default:
		return "🗡️"
	}
}
