package handlers

import (
	"fmt"
	"log"
	"math/rand"
	"sort"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/tormenta-bot/internal/database"
	"github.com/tormenta-bot/internal/drops"
	"github.com/tormenta-bot/internal/game"
	menukit "github.com/tormenta-bot/internal/menu"
	"github.com/tormenta-bot/internal/models"
	"github.com/tormenta-bot/internal/services"
	"github.com/tormenta-bot/internal/timers"
)

// =============================================
// MAP CODE — callbacks curtos para caber nos 64 bytes do Telegram
// =============================================

var mapToCode = map[string]string{
	"village":           "vil",
	"village_outskirts": "vout",
	"dark_forest":       "dfor",
	"crystal_cave":      "ccav",
	"ancient_dungeon":   "adun",
	"dragon_peak":       "dpk",
}
var codeToMap = map[string]string{
	"vil":  "village",
	"vout": "village_outskirts",
	"dfor": "dark_forest",
	"ccav": "crystal_cave",
	"adun": "ancient_dungeon",
	"dpk":  "dragon_peak",
}

func mc(mapID string) string {
	if code, ok := mapToCode[mapID]; ok {
		return code
	}
	return mapID
}
func cm(code string) string {
	if id, ok := codeToMap[code]; ok {
		return id
	}
	return code
}

// skillIndexMap encoda IDs de habilidades em números curtos para callbacks.
// Cada classe tem 6 habilidades max → 1 dígito por habilidade.
// Formato: "0" = primeira skill da lista ordenada da classe, "1" = segunda, etc.
// decodeSkills("warrior", "02") → ["warrior_slash","warrior_cleave"]

func decodeSkills(char *models.Character, encoded string) []string {
	allSkills := sortedClassSkills(char)
	var result []string
	for _, ch := range encoded {
		idx := int(ch - '0')
		if idx >= 0 && idx < len(allSkills) {
			result = append(result, allSkills[idx].ID)
		}
	}
	return result
}

func sortedClassSkills(char *models.Character) []models.Skill {
	learned, _ := database.GetLearnedSkills(char.ID)
	learnedIDs := map[string]bool{}
	for _, ls := range learned {
		learnedIDs[ls.SkillID] = true
	}
	var skills []models.Skill
	for _, sk := range game.Skills {
		if sk.Class == char.Class && learnedIDs[sk.ID] {
			skills = append(skills, sk)
		}
	}
	sort.Slice(skills, func(i, j int) bool {
		return skills[i].MPCost < skills[j].MPCost
	})
	return skills
}

// =============================================
// VIP PANEL
// =============================================

func showVIPPanel(chatID int64, msgID int, userID int64) {
	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}
	_, _ = processAutoHuntOffline(char.ID, autoHuntRefreshCycles)
	char, _ = database.GetCharacter(userID)
	if char == nil {
		return
	}
	isVIP := database.IsVIP(userID)

	vipStatus := "❌ Não ativo"
	if isVIP {
		player, _ := database.GetPlayer(userID)
		if player != nil && player.VIPExpiresAt != nil {
			remaining := time.Until(*player.VIPExpiresAt)
			days := int(remaining.Hours() / 24)
			hours := int(remaining.Hours()) % 24
			vipStatus = fmt.Sprintf("✅ Ativo — %dd %dh restantes", days, hours)
		} else if player != nil && player.IsVIP {
			vipStatus = "✅ Ativo — Permanente"
		}
	}

	session, _ := database.GetAutoHuntSession(char.ID)
	huntStatus := "⏹️ Inativo"
	huntDetails := ""
	if session != nil && session.Status == "running" {
		elapsed := time.Since(session.StartedAt)
		huntStatus = fmt.Sprintf("🏹 Caçando em *%s*", game.Maps[session.MapID].Name)
		huntDetails = fmt.Sprintf(
			"\n⏱️ Há: *%s*\n⚔️ Kills: *%d* | ✨ +*%d* XP | 🪙 +*%d*\n⭐ Nível: *%d* | 💎 Diamantes: *%d*\n🎯 Modo: *%s*",
			formatDuration(elapsed), session.TotalKills, session.TotalXP, session.TotalGold,
			char.Level, char.Diamonds,
			skillModeLabel(session.SkillConfig.Mode),
		)
	}

	normalMax := game.MaxEnergy(char.Level)
	vipMax := game.MaxEnergyVIP(char.Level)

	caption := fmt.Sprintf(
		"👑 *Status VIP*\n\n"+
			"Status: %s\n\n"+
			"*Benefícios VIP:*\n"+
			"⚡ Energia máx: *%d* (normal: %d)\n"+
			"⏱️ Recarga: *5 min* (normal: 10 min)\n"+
			"🏹 Caça automática com habilidades e poções\n\n"+
			"*Caça Automática:*\n"+
			"%s%s",
		vipStatus, vipMax, normalMax, huntStatus, huntDetails,
	)

	var huntRows [][]tgbotapi.InlineKeyboardButton
	if isVIP {
		if session != nil && session.Status == "running" {
		} else {
			maps := game.GetAvailableMapsForHunt(char.Level)
			for _, m := range maps {
				huntRows = append(huntRows, tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData(
						fmt.Sprintf("🏹 Caçar em %s %s", m.Emoji, m.Name),
						"vh_cfg_"+mc(m.ID),
					),
				))
			}
		}
	} else {
		caption += "\n\n_Obtenha VIP para desbloquear esses benefícios!_"
	}
	kb := menukit.VIPPanel(menukit.VIPPanelOptions{
		IsVIP:      isVIP,
		HasSession: session != nil && session.Status == "running",
		HuntRows:   huntRows,
	})
	editPhoto(chatID, msgID, "menu", caption, &kb)
}

// =============================================
// TELA DE CONFIGURAÇÃO DE HABILIDADES E POÇÕES
// =============================================

func showAutoHuntConfig(chatID int64, msgID int, userID int64, mapID string) {
	showAutoHuntConfigWithSkills(chatID, msgID, userID, mapID, "")
}

// showAutoHuntConfigWithSkills é a tela principal de configuração com toggles de habilidades inline.
// selectedEncoded = índices das habilidades selecionadas (ex: "02" = 1ª e 3ª skills)
func showAutoHuntConfigWithSkills(chatID int64, msgID int, userID int64, mapID, selectedEncoded string) {
	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}
	if !database.IsVIP(userID) {
		editPhoto(chatID, msgID, "menu", "👑 *Caça automática é exclusiva VIP!*", bkp("menu_vip"))
		return
	}
	m, ok := game.Maps[mapID]
	if !ok {
		editPhoto(chatID, msgID, "menu", "❌ Área inválida.", bkp("menu_vip"))
		return
	}

	classSkills := sortedClassSkills(char)

	// Parse índices selecionados
	selectedIdx := map[int]bool{}
	for _, ch := range selectedEncoded {
		idx := int(ch - '0')
		if idx >= 0 && idx < len(classSkills) {
			selectedIdx[idx] = true
		}
	}

	// Poções disponíveis no inventário
	inv, _ := database.GetInventory(char.ID)
	var hpPotions, mpPotions []models.Item
	for _, slot := range inv {
		if slot.ItemType != "consumable" || slot.Quantity <= 0 {
			continue
		}
		item, ok2 := game.Items[slot.ItemID]
		if !ok2 {
			continue
		}
		if item.HealHP > 0 && item.HealMP == 0 {
			hpPotions = append(hpPotions, item)
		} else if item.HealMP > 0 && item.HealHP == 0 {
			mpPotions = append(mpPotions, item)
		}
	}

	// ── Caption ──────────────────────────────────────────
	caption := fmt.Sprintf(
		"⚙️ *Configurar Caça — %s %s*\n\n"+
			"❤️ *%d*/%d  |  🔵 *%d*/%d  |  ⚡ *%d*/%d\n\n",
		m.Emoji, m.Name,
		char.HP, char.HPMax, char.MP, char.MPMax, char.Energy, char.EnergyMax,
	)

	if len(classSkills) == 0 {
		caption += "_Nenhuma habilidade aprendida — apenas ataque normal disponível._\n"
	} else {
		caption += fmt.Sprintf("*Habilidades disponíveis (%d):*\n", len(classSkills))
		for i, sk := range classSkills {
			mark := "☐"
			if selectedIdx[i] {
				mark = "✅"
			}
			caption += fmt.Sprintf("%s %s *%s*  _(MP: %d)_\n", mark, sk.Emoji, sk.Name, sk.MPCost)
		}
	}

	if len(hpPotions)+len(mpPotions) > 0 {
		caption += "\n*Poções no inventário:*\n"
		for _, p := range hpPotions {
			qty := database.GetItemCount(char.ID, p.ID)
			caption += fmt.Sprintf("%s *%s* ×%d\n", p.Emoji, p.Name, qty)
		}
		for _, p := range mpPotions {
			qty := database.GetItemCount(char.ID, p.ID)
			caption += fmt.Sprintf("%s *%s* ×%d\n", p.Emoji, p.Name, qty)
		}
	}

	// ── Botões ────────────────────────────────────────────
	var rows [][]tgbotapi.InlineKeyboardButton
	mcode := mc(mapID)

	// Toggles de habilidades: clique remove/adiciona o índice do encoded
	for i, sk := range classSkills {
		check := "☐"
		newEncoded := ""
		if selectedIdx[i] {
			check = "✅"
			// Remove este índice
			for j := range selectedIdx {
				if j != i {
					newEncoded += fmt.Sprintf("%d", j)
				}
			}
		} else {
			// Adiciona este índice
			newEncoded = selectedEncoded + fmt.Sprintf("%d", i)
		}
		label := fmt.Sprintf("%s %s %s  (MP: %d)", check, sk.Emoji, sk.Name, sk.MPCost)
		// vh_tog_<mcode>_<encoded>  — máx: "vh_tog_vout_012345" = 18 bytes ✅
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(label, "vh_tog_"+mcode+"_"+newEncoded),
		))
	}

	// Separador visual antes dos modos
	if len(classSkills) > 0 {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("─────────────────────", "vip_noop"),
		))
	}

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("⚔️ Só Ataque Normal", "vh_sm_"+mcode+"_attack"),
	))
	if len(classSkills) > 0 {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🎯 Todas as Habilidades (rodízio)", "vh_sm_"+mcode+"_skill_all"),
		))
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🧠 Modo Inteligente (habilidade se tiver MP)", "vh_sm_"+mcode+"_smart"),
		))
		// Botão de iniciar com seleção atual
		count := len(selectedIdx)
		startLabel := fmt.Sprintf("▶️ Iniciar com %d habilidade(s) selecionada(s)", count)
		if count == 0 {
			startLabel = "▶️ Iniciar (marque habilidades acima)"
		}
		// vh_ok_<mcode>_<encoded>  — máx: "vh_ok_vout_012345" = 17 bytes ✅
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(startLabel, "vh_ok_"+mcode+"_"+selectedEncoded),
		))
	}

	if len(hpPotions)+len(mpPotions) > 0 {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🧪 Configurar Uso de Poções", "vh_pot_"+mcode+"_"),
		))
	}

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("⬅️ Voltar", "menu_vip"),
	))
	kb := tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}
	editPhoto(chatID, msgID, "menu", caption, &kb)
}

// showSkillPicker — mantido para compatibilidade com callbacks legados.
// Redireciona para a nova tela integrada showAutoHuntConfigWithSkills.
func showSkillPicker(chatID int64, msgID int, userID int64, mapID, selectedRaw string) {
	showAutoHuntConfigWithSkills(chatID, msgID, userID, mapID, selectedRaw)
}

// showPotionPicker exibe seleção de poções para uso automático na caça.
// selectedRaw = IDs separados por "+" | "healAt=N" | "manaAt=N"
// Formato callback: "vip_hunt_potions_<mapID>_<potions+>|h<healAt>|m<manaAt>"
func showPotionPicker(chatID int64, msgID int, userID int64, mapID, selectedRaw string) {
	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}
	inv, _ := database.GetInventory(char.ID)
	mcode := mc(mapID)

	// Parse estado: "potion_small+mana_potion_small|h50|m30"
	selectedPotions := map[string]bool{}
	healAt := 50
	manaAt := 30
	parts := strings.SplitN(selectedRaw, "|", 3)
	if parts[0] != "" {
		for _, id := range strings.Split(parts[0], "+") {
			if id != "" {
				selectedPotions[id] = true
			}
		}
	}
	for _, part := range parts[1:] {
		if strings.HasPrefix(part, "h") {
			fmt.Sscanf(part[1:], "%d", &healAt)
		} else if strings.HasPrefix(part, "m") {
			fmt.Sscanf(part[1:], "%d", &manaAt)
		}
	}

	m := game.Maps[mapID]
	caption := fmt.Sprintf(
		"🧪 *Configurar Poções — %s %s*\n\n"+
			"❤️ HP: *%d*/%d | 🔵 MP: *%d*/%d\n\n"+
			"Selecione as poções para uso automático.\n"+
			"HP abaixo de *%d%%* | MP abaixo de *%d%%*\n\n",
		m.Emoji, m.Name, char.HP, char.HPMax, char.MP, char.MPMax, healAt, manaAt,
	)

	var rows [][]tgbotapi.InlineKeyboardButton
	hasAny := false
	for _, slot := range inv {
		if slot.ItemType != "consumable" || slot.Quantity <= 0 {
			continue
		}
		item, ok := game.Items[slot.ItemID]
		if !ok || (item.HealHP == 0 && item.HealMP == 0) {
			continue
		}
		hasAny = true
		check := "☐"
		var newPotions []string
		if selectedPotions[item.ID] {
			check = "✅"
			for id := range selectedPotions {
				if id != item.ID {
					newPotions = append(newPotions, id)
				}
			}
		} else {
			for id := range selectedPotions {
				newPotions = append(newPotions, id)
			}
			newPotions = append(newPotions, item.ID)
		}
		healInfo := fmt.Sprintf("+%d HP", item.HealHP)
		if item.HealHP == 0 {
			healInfo = fmt.Sprintf("+%d MP", item.HealMP)
		}
		newRaw := buildPotionRaw(newPotions, healAt, manaAt)
		qty := database.GetItemCount(char.ID, item.ID)
		label := fmt.Sprintf("%s %s %s ×%d (%s)", check, item.Emoji, item.Name, qty, healInfo)
		// vh_pt_ = potion toggle
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(label, "vh_pt_"+mcode+"_"+newRaw),
		))
	}
	if !hasAny {
		caption += "_Nenhuma poção de HP/MP no inventário._\n"
	}

	// Threshold HP
	caption += "\n*Usar HP quando abaixo de:*"
	thHP := tgbotapi.NewInlineKeyboardRow()
	for _, pct := range []int{25, 50, 75} {
		mark := ""
		if healAt == pct {
			mark = "✅ "
		}
		newRaw := buildPotionRaw(potionMapToSlice(selectedPotions), pct, manaAt)
		thHP = append(thHP, tgbotapi.NewInlineKeyboardButtonData(
			fmt.Sprintf("%s%d%%", mark, pct), "vh_pt_"+mcode+"_"+newRaw,
		))
	}
	rows = append(rows, thHP)

	// Threshold MP
	caption += " | *MP:*"
	thMP := tgbotapi.NewInlineKeyboardRow()
	for _, pct := range []int{20, 30, 50} {
		mark := ""
		if manaAt == pct {
			mark = "✅ "
		}
		newRaw := buildPotionRaw(potionMapToSlice(selectedPotions), healAt, pct)
		thMP = append(thMP, tgbotapi.NewInlineKeyboardButtonData(
			fmt.Sprintf("%s%d%%", mark, pct), "vh_pt_"+mcode+"_"+newRaw,
		))
	}
	rows = append(rows, thMP)

	confirmRaw := buildPotionRaw(potionMapToSlice(selectedPotions), healAt, manaAt)
	// vh_pk_ = potion confirm (done)
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(
			fmt.Sprintf("✅ Confirmar (%d poção(ões))", len(selectedPotions)),
			"vh_pk_"+mcode+"_"+confirmRaw,
		),
	))
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("⬅️ Voltar", "vh_cfg_"+mcode),
	))
	kb := tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}
	editPhoto(chatID, msgID, "menu", caption, &kb)
}

func buildPotionRaw(potions []string, healAt, manaAt int) string {
	potionStr := strings.Join(potions, "+")
	return fmt.Sprintf("%s|h%d|m%d", potionStr, healAt, manaAt)
}

func potionMapToSlice(m map[string]bool) []string {
	var result []string
	for id := range m {
		result = append(result, id)
	}
	return result
}

func handleAutoHuntStart(chatID int64, msgID int, userID int64, mapID string, cfg database.AutoHuntSkillConfig) {
	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}
	if !database.IsVIP(userID) {
		editPhoto(chatID, msgID, "menu", "👑 *Caça automática é exclusiva VIP!*", bkp("menu_vip"))
		return
	}
	if char.State != "idle" {
		editPhoto(chatID, msgID, "menu", "❌ Termine o combate atual antes de iniciar a caça automática.", bkp("menu_vip"))
		return
	}
	if char.Energy < game.EnergyCombatEnter {
		editPhoto(chatID, msgID, "menu",
			fmt.Sprintf("❌ *Energia insuficiente!*\n\n⚡ *%d*/%d — precisa *%d* ⚡", char.Energy, char.EnergyMax, game.EnergyCombatEnter),
			bkp("menu_vip"))
		return
	}
	m, ok := game.Maps[mapID]
	if !ok || len(m.Monsters) == 0 {
		editPhoto(chatID, msgID, "menu", "❌ Área inválida para caça.", bkp("menu_vip"))
		return
	}
	if cfg.Mode == "skill" && len(cfg.Skills) == 0 {
		cfg.Mode = "smart"
	}

	char.CurrentMap = mapID
	char.State = "auto_hunt"
	database.SaveCharacter(char)

	session, err := database.StartAutoHunt(char.ID, mapID, cfg)
	if err != nil {
		log.Printf("[AutoHunt] Error starting session: %v", err)
		return
	}

	// Lança o loop contínuo de combate para esta sessão
	_ = timers.SetAfter(userID, autoHuntTimerKey, autoHuntCombatInterval)

	modeLabel := skillModeLabel(cfg.Mode)
	skillSummary := ""
	if cfg.Mode == "skill" && len(cfg.Skills) > 0 {
		skillSummary = fmt.Sprintf("\n🎯 Habilidades: *%d* selecionada(s)", len(cfg.Skills))
	}
	potionSummary := ""
	if len(cfg.Potions) > 0 {
		potionSummary = fmt.Sprintf("\n🧪 Poções: *%d* tipo(s) ativado(s) (HP<%d%% | MP<%d%%)",
			len(cfg.Potions), cfg.HealAt, cfg.ManaAt)
	}

	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📊 Ver Status", "menu_vip"),
			tgbotapi.NewInlineKeyboardButtonData("⏹️ Parar", "vip_hunt_stop"),
		),
	)
	editPhoto(chatID, msgID, "menu",
		fmt.Sprintf("🏹 *Caça automática iniciada!*\n\n"+
			"📍 Área: *%s %s*\n"+
			"⚔️ Modo: *%s*%s%s\n"+
			"⚡ Energia: *%d*/%d\n"+
			"❤️ HP: *%d*/%d | 🔵 MP: *%d*/%d\n"+
			"⭐ Nível: *%d* | 💎 Diamantes: *%d*\n\n"+
			"_Seu personagem está caçando continuamente. "+
			"O progresso é processado por ciclos quando você abrir o relatório._",
			m.Emoji, m.Name, modeLabel, skillSummary, potionSummary,
			char.Energy, char.EnergyMax, char.HP, char.HPMax, char.MP, char.MPMax,
			char.Level, char.Diamonds),
		&kb)

	log.Printf("[AutoHunt] Started charID=%d map=%s mode=%s skills=%v sessionID=%d",
		char.ID, mapID, cfg.Mode, cfg.Skills, session.ID)
}

func handleAutoHuntStop(chatID int64, msgID int, userID int64) {
	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}
	_, _ = processAutoHuntOffline(char.ID, autoHuntStopCatchupCycles)
	database.StopAutoHunt(char.ID, "stopped")
	_ = timers.Clear(userID, autoHuntTimerKey)
	char.State = "idle"
	database.SaveCharacter(char)
	editMainMenuPhoto(chatID, msgID, userID)
}

func handleAutoHuntReport(chatID int64, msgID int, userID int64) {
	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}
	_, _ = processAutoHuntOffline(char.ID, autoHuntReportCycles)
	char, _ = database.GetCharacter(userID)
	if char == nil {
		return
	}
	session, _ := database.GetAutoHuntSession(char.ID)
	if session == nil {
		showVIPPanel(chatID, msgID, userID)
		return
	}
	recalculateStats(char)
	elapsed := time.Since(session.StartedAt)
	m := game.Maps[session.MapID]
	statusEmoji := "🏹"
	if session.Status != "running" {
		statusEmoji = "⏹️"
	}
	// Resumo de poções configuradas
	potionInfo := ""
	if len(session.SkillConfig.Potions) > 0 {
		potionInfo = fmt.Sprintf("\n🧪 Poções: *%d tipo(s)* (HP<%d%% | MP<%d%%)",
			len(session.SkillConfig.Potions), session.SkillConfig.HealAt, session.SkillConfig.ManaAt)
	}

	// CA e bônus de ataque atuais
	defAttr := game.DefensiveAttr(char.Class, char.Constitution, char.Dexterity, char.Intelligence)
	playerCA := game.CharacterCA(char.Class, defAttr, char.EquipCABonus)
	atkBonus := game.CharacterAttackBonus(char.Class, char.Level, char.Strength, char.Dexterity, char.Intelligence) + char.EquipHitBonus

	editPhoto(chatID, msgID, "menu",
		fmt.Sprintf("%s *Relatório de Caça*\n\n"+
			"📍 %s *%s*\n"+
			"⚔️ Modo: *%s*%s\n"+
			"⏱️ Duração: *%s*\n\n"+
			"⚔️ Monstros mortos: *%d*\n"+
			"✨ XP total: *+%d*\n"+
			"🪙 Ouro total: *+%d*\n\n"+
			"❤️ HP: *%d*/%d | 🔵 MP: *%d*/%d\n"+
			"⚡ Energia: *%d*/%d\n"+
			"⭐ Nível: *%d* | 💎 Diamantes: *%d*\n\n"+
			"🎲 *Combate (d20)*\n🛡️ CA: *%d* | 🎯 Bônus ataque: *+%d*",
			statusEmoji, m.Emoji, m.Name,
			skillModeLabel(session.SkillConfig.Mode), potionInfo,
			formatDuration(elapsed),
			session.TotalKills, session.TotalXP, session.TotalGold,
			char.HP, char.HPMax, char.MP, char.MPMax,
			char.Energy, char.EnergyMax,
			char.Level, char.Diamonds,
			playerCA, atkBonus),
		&tgbotapi.InlineKeyboardMarkup{InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
			{tgbotapi.NewInlineKeyboardButtonData("🔄 Atualizar", "vip_hunt_report"),
				tgbotapi.NewInlineKeyboardButtonData("⏹️ Parar", "vip_hunt_stop")},
			{tgbotapi.NewInlineKeyboardButtonData("🏰 Menu", "menu_main")},
		}})
}

// =============================================
// AUTO HUNT BACKGROUND WORKER
// =============================================

const (
	autoHuntCombatInterval    = 60 * time.Second
	autoHuntRefreshCycles     = 30
	autoHuntReportCycles      = 180
	autoHuntStopCatchupCycles = 180
	autoHuntTimerKey          = "auto_hunt_next_cycle"
)

var autoHuntService = services.NewAutoHuntService(autoHuntCombatInterval)

func StartAutoHuntWorker() {
	log.Println("🏹 Auto hunt offline mode enabled (worker disabled)")
}

// processAutoHuntOffline aplica ciclos pendentes com base no tempo desde o último tick.
func processAutoHuntOffline(charID int, maxCycles int) (int, error) {
	return autoHuntService.ProcessOffline(
		charID,
		maxCycles,
		database.GetAutoHuntSession,
		doAutoHuntCombat,
	)
}

// doAutoHuntCombat executa um único combate da caça automática.
// Retorna true se deve continuar a caça, false se deve parar.
func doAutoHuntCombat(charID int, session *database.AutoHuntSession) bool {
	char, err := database.GetCharacterByID(charID)
	if err != nil || char == nil {
		database.StopAutoHunt(charID, "stopped")
		return false
	}
	playerID := char.PlayerID

	if !database.IsVIP(playerID) {
		database.StopAutoHunt(charID, "stopped")
		_ = timers.Clear(playerID, autoHuntTimerKey)
		char.State = "idle"
		database.SaveCharacter(char)
		notifyUser(playerID, "⚠️ *Caça automática pausada*\n\nSeu VIP expirou. Renove para continuar.")
		return false
	}

	game.TickEnergyVIP(char, true)

	if !game.ConsumeAttackEnergy(char) {
		database.StopAutoHunt(charID, "out_of_energy")
		_ = timers.Clear(playerID, autoHuntTimerKey)
		char.State = "idle"
		database.SaveCharacter(char)
		// Busca totais atualizados para o relatório
		s, _ := database.GetAutoHuntSession(charID)
		elapsed := time.Since(session.StartedAt)
		m := game.Maps[session.MapID]
		var kills, xp, gold int
		if s != nil {
			kills, xp, gold = s.TotalKills, s.TotalXP, s.TotalGold
		}
		notifyUser(playerID, fmt.Sprintf(
			"⚡ *Energia esgotada — caça automática encerrada!*\n\n"+
				"📍 %s *%s*\n⏱️ Duração: *%s*\n\n"+
				"⚔️ Kills: *%d*\n✨ XP: *+%d*\n🪙 Ouro: *+%d*\n\n"+
				"❤️ HP: *%d*/%d | 🔵 MP: *%d*/%d\n"+
				"_Recarregue a energia para reiniciar a caça._",
			m.Emoji, m.Name, formatDuration(elapsed), kills, xp, gold,
			char.HP, char.HPMax, char.MP, char.MPMax))
		return false
	}

	monsters := game.GetMonstersForMap(session.MapID)
	if len(monsters) == 0 {
		database.StopAutoHunt(charID, "stopped")
		_ = timers.Clear(playerID, autoHuntTimerKey)
		return false
	}
	monster := monsters[rand.Intn(len(monsters))]

	// Garante que EquipCABonus e EquipHitBonus estão atualizados antes do combate
	recalculateStats(char)

	// Simula combate usando habilidades conforme config
	monsterHP := monster.HP
	rounds := 0
	skillIdx := 0
	hits, misses, crits := 0, 0, 0
	// Estado de veneno local para esta luta (monstro envenenado pelo player)
	monsterPoisonTurns := 0
	monsterPoisonDmg := 0
	fx := TempEffects{}

	for monsterHP > 0 && char.HP > 0 && rounds < 30 {
		// Aplica DoT de veneno no player
		if dotDmg, _ := game.ApplyPlayerPoisonDoT(char); dotDmg > 0 {
			char.HP -= dotDmg
			if char.HP < 0 {
				char.HP = 0
			}
		}
		if char.HP <= 0 {
			break
		}

		// Aplica DoT de veneno no monstro
		if monsterPoisonTurns > 0 {
			monsterHP -= monsterPoisonDmg
			monsterPoisonTurns--
			if monsterPoisonTurns == 0 {
				monsterPoisonDmg = 0
			}
		}
		if burnDmg, _ := fx.ApplyEnemyDot(); burnDmg > 0 {
			monsterHP -= burnDmg
		}
		if monsterHP <= 0 {
			break
		}

		var r game.CombatResult
		sk := pickSkillForTick(char, session.SkillConfig, &skillIdx)
		if sk != nil && skillRequiresShield(sk.ID) && !hasEquippedShield(char.ID) {
			sk = nil
		}
		monsterAdj := monster
		if pen := fx.EffectiveCAPenalty(); pen > 0 {
			monsterAdj.CA -= pen
			if monsterAdj.CA < 1 {
				monsterAdj.CA = 1
			}
		}
		if ap := fx.EffectiveAtkPenalty(); ap > 0 {
			monsterAdj.Attack -= ap
			if monsterAdj.Attack < 1 {
				monsterAdj.Attack = 1
			}
		}
		origCA := char.EquipCABonus
		char.EquipCABonus = origCA + fx.EffectiveCABonus()
		if sk != nil {
			r = game.PlayerSkillAttack(char, sk, &monsterAdj)
			char.MP -= sk.MPCost
			if char.MP < 0 {
				char.MP = 0
			}
			if !r.IsPlayerMiss {
				if !(sk.ID == "w_blood_rage" && char.HPMax > 0 && (char.HP*100/char.HPMax) >= 30) {
					_ = applySkillEffectsPVE(sk.ID, &fx)
				}
				critMin := 21
				switch sk.ID {
				case "w_rampage", "m_meteor", "a_headshot":
					critMin = 18
				case "r_vital_strike":
					critMin = 17
				case "w_cleave":
					critMin = 19
				}
				if critMin <= 20 && r.PlayerRoll >= critMin && r.PlayerDamage > 0 && !r.IsCritical {
					r.PlayerDamage *= 2
					r.IsCritical = true
				}
				if sk.ID == "r_death_blow" && monsterHP <= monster.HP/4 && r.PlayerDamage > 0 {
					r.PlayerDamage *= 3
				}
				switch sk.ID {
				case "a_multishot":
					r.PlayerDamage *= 3
				case "a_volley":
					r.PlayerDamage *= 5
				case "a_quick_shot":
					if rand.Intn(100) < 30 {
						r.PlayerDamage *= 2
					}
				}
				if sk.ID == "m_arcane_burst" && fx.EffectiveCABonus() > 0 && r.PlayerDamage > 0 {
					r.PlayerDamage *= 2
				}
			}
			// Aplica veneno ao monstro se a skill envenenou
			if r.AppliesPoison && !r.IsPlayerMiss {
				if r.PoisonTurns > monsterPoisonTurns || r.PoisonDmg > monsterPoisonDmg {
					monsterPoisonTurns = r.PoisonTurns
					monsterPoisonDmg = r.PoisonDmg
				}
			}
		} else {
			r = game.PlayerAttack(char, &monsterAdj)
		}
		char.EquipCABonus = origCA
		if r.PlayerDamage > 0 && !r.IsPlayerMiss && !r.IsCritical && fx.ConsumeForceCrit() {
			r.PlayerDamage *= 2
			r.IsCritical = true
		}
		if r.MonsterDamage > 0 && fx.ConsumeSkipEnemyAttack() {
			r.MonsterDamage = 0
		}
		r.PlayerDamage = fx.ApplyOutgoingDamage(r.PlayerDamage, r.PlayerDamage > 0)
		r.MonsterDamage = fx.ApplyIncomingDamage(r.MonsterDamage)
		monsterHP -= r.PlayerDamage
		char.HP -= r.MonsterDamage
		if char.HP < 0 {
			char.HP = 0
		}
		fx.AdvanceTurn()
		// Contabilizar acertos/erros para o log
		if r.IsCritical {
			crits++
		} else if r.IsPlayerMiss {
			misses++
		} else if r.PlayerDamage > 0 {
			hits++
		}
		rounds++
	}

	// Usa poções após o combate se necessário
	applyAutoPotions(char, session.SkillConfig)

	if char.HP <= 0 {
		// Token de reviver também protege na auto-caça.
		if database.GetItemCount(char.ID, "revive_token") > 0 {
			_ = database.RemoveItem(char.ID, "revive_token", 1)
			char.HP = char.HPMax
			char.MP = char.MPMax
			char.PoisonTurns = 0
			char.PoisonDmg = 0
			database.SaveCharacter(char)
			notifyUser(playerID, fmt.Sprintf(
				"🔮 *Token de Reviver ativado na caça automática!*\n\n"+
					"Você foi salvo contra *%s %s*.\n\n"+
					"❤️ HP: *%d/%d* | 💙 MP: *%d/%d*\n"+
					"_A caça automática continuará._",
				monster.Emoji, monster.Name, char.HP, char.HPMax, char.MP, char.MPMax))
			return true
		}

		xpLost, goldLost := game.ApplyDeathPenalty(char)
		database.SaveCharacter(char)
		database.StopAutoHunt(charID, "stopped")
		_ = timers.Clear(playerID, autoHuntTimerKey)
		char.State = "idle"
		database.SaveCharacter(char)
		s, _ := database.GetAutoHuntSession(charID)
		var kills, xp, gold int
		if s != nil {
			kills, xp, gold = s.TotalKills, s.TotalXP, s.TotalGold
		}
		notifyUser(playerID, fmt.Sprintf(
			"💀 *Seu personagem morreu durante a caça automática!*\n\n"+
				"Derrotado por *%s %s*\n\n"+
				"📊 *Resultado:*\n"+
				"⚔️ Kills: *%d* | ✨ XP: *+%d* | 🪙 Ouro: *+%d*\n\n"+
				"Perdeu: *%d* XP e *%d* 🪙",
			monster.Emoji, monster.Name, kills, xp, gold, xpLost, goldLost))
		return false
	}

	// Recompensas
	xpGain := game.CalculateXPGain(char, &monster)
	goldGain := monster.GoldReward + rand.Intn(monster.GoldReward/2+1)
	char.Experience += xpGain
	char.Gold += goldGain

	// Regen MP passiva por combate (10% do max, mínimo 5)
	mpRegen := char.MPMax / 10
	if mpRegen < 5 {
		mpRegen = 5
	}
	char.MP += mpRegen
	if char.MP > char.MPMax {
		char.MP = char.MPMax
	}

	if monster.DiamondChance > 0 && rand.Intn(100) < monster.DiamondChance {
		char.Diamonds++
		database.LogDiamond(char.ID, 1, "autohunt_drop_"+monster.ID)
	}

	if lvlUp := game.CheckLevelUp(char); lvlUp != nil {
		game.ApplyLevelUp(char, lvlUp)
		char.EnergyMax = game.MaxEnergyVIP(char.Level)
	}
	_ = applyMaterialDrops(char, &monster, drops.ModeAutoHunt)

	database.SaveCharacter(char)
	database.UpdateAutoHuntTick(session.ID, xpGain, goldGain)
	database.LogCombat(char.ID, monster.ID, "win", xpGain, goldGain)
	_ = timers.SetAfter(playerID, autoHuntTimerKey, autoHuntCombatInterval)

	log.Printf("[AutoHunt] combat charID=%d monster=%s(CA:%d) rounds=%d hits=%d misses=%d crits=%d xp=%d gold=%d hp=%d/%d mp=%d/%d",
		char.ID, monster.ID, monster.CA, rounds, hits, misses, crits, xpGain, goldGain, char.HP, char.HPMax, char.MP, char.MPMax)
	return true
}

// applyAutoPotions usa poções configuradas na caça automática quando HP/MP estão baixos.
func applyAutoPotions(char *models.Character, cfg database.AutoHuntSkillConfig) {
	if len(cfg.Potions) == 0 {
		return
	}
	healAt := cfg.HealAt
	if healAt == 0 {
		healAt = 50
	}
	manaAt := cfg.ManaAt
	if manaAt == 0 {
		manaAt = 30
	}

	hpPct := char.HP * 100 / char.HPMax
	mpPct := 100
	if char.MPMax > 0 {
		mpPct = char.MP * 100 / char.MPMax
	}

	for _, potionID := range cfg.Potions {
		item, ok := game.Items[potionID]
		if !ok {
			continue
		}
		qty := database.GetItemCount(char.ID, potionID)
		if qty <= 0 {
			continue
		}
		used := false
		// Poção de HP
		if item.HealHP > 0 && item.HealMP == 0 && hpPct < healAt && char.HP < char.HPMax {
			char.HP += item.HealHP
			if char.HP > char.HPMax {
				char.HP = char.HPMax
			}
			hpPct = char.HP * 100 / char.HPMax
			used = true
		}
		// Poção de MP
		if item.HealMP > 0 && item.HealHP == 0 && mpPct < manaAt && char.MP < char.MPMax {
			char.MP += item.HealMP
			if char.MP > char.MPMax {
				char.MP = char.MPMax
			}
			if char.MPMax > 0 {
				mpPct = char.MP * 100 / char.MPMax
			}
			used = true
		}
		// Antídoto: usa automaticamente se envenenado
		if item.CurePoison && char.PoisonTurns > 0 {
			char.PoisonTurns = 0
			char.PoisonDmg = 0
			used = true
		}
		if used {
			database.RemoveItem(char.ID, potionID, 1)
		}
	}
}

// pickSkillForTick decide qual habilidade usar no tick baseado na configuração.
// Retorna nil para usar ataque normal.
func pickSkillForTick(char *models.Character, cfg database.AutoHuntSkillConfig, idx *int) *models.Skill {
	switch cfg.Mode {
	case "attack":
		return nil

	case "skill":
		// Rodízio pelas habilidades selecionadas na config
		if len(cfg.Skills) == 0 {
			return nil
		}
		for range cfg.Skills {
			skillID := cfg.Skills[*idx%len(cfg.Skills)]
			*idx++
			sk, ok := game.Skills[skillID]
			if ok && char.MP >= sk.MPCost {
				return &sk
			}
		}
		return nil // sem MP para qualquer habilidade da lista

	case "smart":
		// Usa a habilidade de maior dano que caiba no MP atual
		// Prioriza as selecionadas; se não tiver, tenta qualquer uma da classe
		candidates := cfg.Skills
		if len(candidates) == 0 {
			for id, sk := range game.Skills {
				if sk.Class == char.Class {
					candidates = append(candidates, id)
				}
			}
		}
		var best *models.Skill
		for _, skillID := range candidates {
			sk, ok := game.Skills[skillID]
			if !ok || char.MP < sk.MPCost {
				continue
			}
			if best == nil || sk.Damage > best.Damage {
				skCopy := sk
				best = &skCopy
			}
		}
		return best
	}
	return nil
}

// =============================================
// CALLBACK ROUTERS (chamados de handlers.go)
// =============================================

// handlePotionsDone salva a config de poções e volta para a tela de modo de combate
// para que o jogador possa iniciar a caça com modo + poções configurados.
func handlePotionsDone(chatID int64, msgID int, userID int64, mapID, raw string) {
	// Parse o raw para extrair poções e thresholds
	potions, healAt, manaAt := parsePotionRaw(raw)

	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}
	m := game.Maps[mapID]
	mcode := mc(mapID)

	potionNames := ""
	for _, pid := range potions {
		if item, ok := game.Items[pid]; ok {
			potionNames += fmt.Sprintf("• %s %s\n", item.Emoji, item.Name)
		}
	}
	if potionNames == "" {
		potionNames = "_Nenhuma_\n"
	}

	caption := fmt.Sprintf(
		"✅ *Poções Configuradas — %s %s*\n\n"+
			"🧪 *Poções ativas:*\n%s\n"+
			"❤️ HP abaixo de *%d%%* | 🔵 MP abaixo de *%d%%*\n\n"+
			"*Escolha o modo de combate para iniciar:*",
		m.Emoji, m.Name, potionNames, healAt, manaAt,
	)

	potionSuffix := buildPotionRaw(potions, healAt, manaAt)

	learned, _ := database.GetLearnedSkills(char.ID)
	learnedIDs := map[string]bool{}
	for _, ls := range learned {
		learnedIDs[ls.SkillID] = true
	}
	hasSkills := false
	for _, sk := range game.Skills {
		if sk.Class == char.Class && learnedIDs[sk.ID] {
			hasSkills = true
			break
		}
	}

	var rows [][]tgbotapi.InlineKeyboardButton
	// vh_go_ = go/start com mode + potions
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("⚔️ Ataque Normal", "vh_go_"+mcode+"_attack_"+potionSuffix),
	))
	if hasSkills {
		rows = append(rows,
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🎯 Todas as Habilidades", "vh_go_"+mcode+"_skill_all_"+potionSuffix),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🧠 Modo Inteligente", "vh_go_"+mcode+"_smart_"+potionSuffix),
			),
		)
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("⬅️ Voltar", "vh_pot_"+mcode+"_"+raw),
	))
	kb := tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}
	editPhoto(chatID, msgID, "menu", caption, &kb)
}

func parsePotionRaw(raw string) (potions []string, healAt, manaAt int) {
	healAt = 50
	manaAt = 30
	parts := strings.SplitN(raw, "|", 3)
	if parts[0] != "" {
		for _, id := range strings.Split(parts[0], "+") {
			if id != "" {
				potions = append(potions, id)
			}
		}
	}
	for _, part := range parts[1:] {
		if strings.HasPrefix(part, "h") {
			fmt.Sscanf(part[1:], "%d", &healAt)
		} else if strings.HasPrefix(part, "m") {
			fmt.Sscanf(part[1:], "%d", &manaAt)
		}
	}
	return
}

// handleHuntSkillToggle processa o clique em toggle de habilidade na tela de configuração.
// rest = "<mcode>_<encodedSelected>"
func handleHuntSkillToggle(chatID int64, msgID int, userID int64, rest string) {
	idx := strings.Index(rest, "_")
	if idx < 0 {
		showVIPPanel(chatID, msgID, userID)
		return
	}
	mapID := cm(rest[:idx])
	selectedEncoded := rest[idx+1:]
	showAutoHuntConfigWithSkills(chatID, msgID, userID, mapID, selectedEncoded)
}

// splitMapIDFromRest separa o mapID do restante do callback, testando contra maps conhecidos.
func splitMapIDFromRest(rest string) (mapID, remainder string) {
	var candidates []string
	for id := range game.Maps {
		candidates = append(candidates, id)
	}
	// Ordena do maior para o menor para evitar match parcial
	sort.Slice(candidates, func(i, j int) bool {
		return len(candidates[i]) > len(candidates[j])
	})
	for _, id := range candidates {
		prefix := id + "_"
		if strings.HasPrefix(rest, prefix) {
			return id, rest[len(prefix):]
		}
		if rest == id {
			return id, ""
		}
	}
	// Fallback
	parts := strings.SplitN(rest, "_", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return rest, ""
}

// Chamado pelo callback "vip_hunt_config_<mapID>" e "vip_hunt_start_<mapID>" (legado).
func handleAutoHuntStart_legacy(chatID int64, msgID int, userID int64, mapID string) {
	showAutoHuntConfig(chatID, msgID, userID, mapID)
}

// handleAutoHuntStartFull processa início com modo + poções: vh_go_<mcode>_<mode>_<potionRaw>
// ex: "vh_go_dfor_smart_potion_small|h50|m30" → mapID="dark_forest", mode="smart"
func handleAutoHuntStartFull(chatID int64, msgID int, userID int64, rest string) {
	// rest = "<mcode>_<mode>_<potionRaw>"
	// mcode é o primeiro segmento (curto, sem "_" exceto separador)
	// mode pode ser "attack", "smart", "skill_all"
	var mcode, modeStr, potionRaw string
	for _, suffix := range []string{"_skill_all_", "_smart_", "_attack_"} {
		if idx := strings.Index(rest, suffix); idx >= 0 {
			mcode = rest[:idx]
			modeStr = strings.Trim(suffix, "_")
			potionRaw = rest[idx+len(suffix):]
			break
		}
	}
	if mcode == "" {
		showVIPPanel(chatID, msgID, userID)
		return
	}
	mapID := cm(mcode)
	potions, healAt, manaAt := parsePotionRaw(potionRaw)
	cfg := database.AutoHuntSkillConfig{
		Mode: modeStr, Skills: []string{},
		Potions: potions, HealAt: healAt, ManaAt: manaAt,
	}
	if modeStr == "skill_all" {
		cfg.Mode = "skill"
		char, _ := database.GetCharacter(userID)
		if char != nil {
			learned, _ := database.GetLearnedSkills(char.ID)
			learnedIDs := map[string]bool{}
			for _, ls := range learned {
				learnedIDs[ls.SkillID] = true
			}
			for _, sk := range game.Skills {
				if sk.Class == char.Class && learnedIDs[sk.ID] {
					cfg.Skills = append(cfg.Skills, sk.ID)
				}
			}
		}
	}
	handleAutoHuntStart(chatID, msgID, userID, mapID, cfg)
}

// handleAutoHuntSetMode processa a escolha rápida de modo.
// rest = "<mcode>_attack" | "<mcode>_skill_all" | "<mcode>_smart"
// (recebe código curto do mapa, ex: "vout" para "village_outskirts")
func handleAutoHuntSetMode(chatID int64, msgID int, userID int64, rest string) {
	var mcode, mode string
	for _, suffix := range []string{"_skill_all", "_smart", "_attack"} {
		if strings.HasSuffix(rest, suffix) {
			mcode = strings.TrimSuffix(rest, suffix)
			mode = strings.TrimPrefix(suffix, "_")
			break
		}
	}
	if mcode == "" {
		showVIPPanel(chatID, msgID, userID)
		return
	}
	mapID := cm(mcode)
	cfg := database.AutoHuntSkillConfig{Mode: mode, Skills: []string{}, Potions: []string{}}

	if mode == "skill_all" {
		cfg.Mode = "skill"
		char, _ := database.GetCharacter(userID)
		if char != nil {
			learned, _ := database.GetLearnedSkills(char.ID)
			learnedIDs := map[string]bool{}
			for _, ls := range learned {
				learnedIDs[ls.SkillID] = true
			}
			for _, sk := range game.Skills {
				if sk.Class == char.Class && learnedIDs[sk.ID] {
					cfg.Skills = append(cfg.Skills, sk.ID)
				}
			}
		}
	}
	handleAutoHuntStart(chatID, msgID, userID, mapID, cfg)
}

// handleAutoHuntConfirm processa o botão "▶️ Iniciar com N habilidade(s)".
// rest = "<mcode>_<encodedIndices>"  ex: "vout_025"
func handleAutoHuntConfirm(chatID int64, msgID int, userID int64, rest string) {
	idx := strings.Index(rest, "_")
	if idx < 0 {
		showVIPPanel(chatID, msgID, userID)
		return
	}
	mcode := rest[:idx]
	encoded := rest[idx+1:]
	mapID := cm(mcode)

	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}
	skills := decodeSkills(char, encoded)
	cfg := database.AutoHuntSkillConfig{Mode: "skill", Skills: skills, Potions: []string{}}
	if len(skills) == 0 {
		cfg.Mode = "smart"
	}
	handleAutoHuntStart(chatID, msgID, userID, mapID, cfg)
}

// =============================================
// VIP PURCHASE
// =============================================

func showVIPBuyOptions(chatID int64, msgID int) {
	caption := "👑 *Comprar VIP*\n\n" +
		"*Benefícios:*\n" +
		"⚡ Dobro de energia máxima\n" +
		"⏱️ Recarga 2x mais rápida (8min)\n" +
		"🏹 Caça automática com habilidades e poções\n\n" +
		"*Planos disponíveis:*\n\n" +
		"💎 *30 dias* — 500 💎\n" +
		"💎 *90 dias* — 1200 💎\n" +
		"💎 *Permanente* — 3000 💎\n\n" +
		"_Pagamento com diamantes do jogo_"

	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("30 dias (500💎)", "vip_buy_30"),
			tgbotapi.NewInlineKeyboardButtonData("90 dias (1200💎)", "vip_buy_90"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("♾️ Permanente (3000💎)", "vip_buy_perm"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⬅️ Voltar", "menu_vip"),
		),
	)
	editPhoto(chatID, msgID, "menu", caption, &kb)
}

func handleVIPPurchase(chatID int64, msgID int, userID int64, plan string) {
	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}
	costs := map[string]int{"30": 500, "90": 1200, "perm": 3000}
	durations := map[string]time.Duration{
		"30": 30 * 24 * time.Hour, "90": 90 * 24 * time.Hour, "perm": 0,
	}
	labels := map[string]string{"30": "30 dias", "90": "90 dias", "perm": "Permanente"}

	cost, ok := costs[plan]
	if !ok {
		return
	}
	if char.Diamonds < cost {
		editPhoto(chatID, msgID, "menu",
			fmt.Sprintf("❌ *Diamantes insuficientes!*\n\nPrecisa: *%d* 💎\nVocê tem: *%d* 💎", cost, char.Diamonds),
			&tgbotapi.InlineKeyboardMarkup{InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
				{tgbotapi.NewInlineKeyboardButtonData("⬅️ Voltar", "menu_vip")},
			}})
		return
	}
	char.Diamonds -= cost
	database.SaveCharacter(char)
	database.LogDiamond(char.ID, -cost, "vip_purchase_"+plan)
	database.SetVIP(userID, true, durations[plan])
	newMax := game.MaxEnergyVIP(char.Level)
	char.EnergyMax = newMax
	if char.Energy > newMax {
		char.Energy = newMax
	}
	database.SaveCharacter(char)
	editPhoto(chatID, msgID, "menu",
		fmt.Sprintf("👑 *VIP Ativado!*\n\n*Plano:* %s\n*Custo:* %d 💎\n\n"+
			"⚡ Nova energia máxima: *%d*\n⏱️ Recarga: *5 minutos*\n"+
			"🏹 Caça automática com habilidades e poções desbloqueada!",
			labels[plan], cost, newMax),
		&tgbotapi.InlineKeyboardMarkup{InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
			{tgbotapi.NewInlineKeyboardButtonData("👑 Painel VIP", "menu_vip")},
		}})
}

// =============================================
// GM VIP MANAGEMENT
// =============================================

func SetVIPFromGM(charID int, active bool, days int) error {
	char, err := database.GetCharacterByID(charID)
	if err != nil || char == nil {
		return fmt.Errorf("character not found")
	}
	var duration time.Duration
	if days > 0 {
		duration = time.Duration(days) * 24 * time.Hour
	}
	if err := database.SetVIP(char.PlayerID, active, duration); err != nil {
		return err
	}
	if active {
		char.EnergyMax = game.MaxEnergyVIP(char.Level)
		if char.Energy > char.EnergyMax {
			char.Energy = char.EnergyMax
		}
	} else {
		char.EnergyMax = game.MaxEnergy(char.Level)
		if char.Energy > char.EnergyMax {
			char.Energy = char.EnergyMax
		}
		database.StopAutoHunt(charID, "stopped")
		if char.State == "auto_hunt" {
			char.State = "idle"
		}
	}
	database.SaveCharacter(char)
	return nil
}

// =============================================
// HELPERS
// =============================================

func skillModeLabel(mode string) string {
	switch mode {
	case "skill":
		return "🎯 Habilidades (rodízio)"
	case "smart":
		return "🧠 Inteligente"
	default:
		return "⚔️ Ataque Normal"
	}
}

func notifyUser(playerID int64, text string) {
	if Bot == nil {
		return
	}
	msg := tgbotapi.NewMessage(playerID, text)
	msg.ParseMode = "Markdown"
	Bot.Send(msg)
}

func formatDuration(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	if h > 0 {
		return fmt.Sprintf("%dh %dm", h, m)
	}
	if m > 0 {
		return fmt.Sprintf("%dm %ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}
