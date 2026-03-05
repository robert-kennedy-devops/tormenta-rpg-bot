package handlers

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/tormenta-bot/internal/database"
	"github.com/tormenta-bot/internal/game"
	menukit "github.com/tormenta-bot/internal/menu"
)

// =============================================
// GM / ADMIN PANEL
// =============================================
// Variáveis de ambiente:
//   GM_IDS=6035979086                — IDs Telegram dos GMs separados por vírgula
//   GM_LOG_CHAT=<chatID>        — (opcional) ID de chat onde ações GM são logadas

// ── Auth ──────────────────────────────────────────────────

var (
	gmIDsMap     map[int64]bool
	gmIDsRawLast string
	gmIDsMu      sync.Mutex
)

func isGM(userID int64) bool {
	raw := os.Getenv("GM_IDS")
	gmIDsMu.Lock()
	defer gmIDsMu.Unlock()
	if gmIDsMap == nil || raw != gmIDsRawLast {
		gmIDsMap = map[int64]bool{}
		gmIDsRawLast = raw
		for _, part := range strings.Split(raw, ",") {
			part = strings.TrimSpace(part)
			if id, err := strconv.ParseInt(part, 10, 64); err == nil {
				gmIDsMap[id] = true
			}
		}
		log.Printf("[GM] IDs carregados: %v (GM_IDS=%q)", gmIDsMap, raw)
	}
	return gmIDsMap[userID]
}

// gmLog sends an action log to GM_LOG_CHAT if configured.
func gmLog(gmUserID int64, action string) {
	logChatStr := os.Getenv("GM_LOG_CHAT")
	if logChatStr == "" {
		return
	}
	logChatID, err := strconv.ParseInt(logChatStr, 10, 64)
	if err != nil || logChatID == 0 {
		return
	}
	ts := time.Now().Format("02/01 15:04")
	text := fmt.Sprintf("📋 *[GM LOG]* `%s`\n🕐 %s\n👤 GM ID: `%d`\n📝 %s", ts, ts, gmUserID, action)
	msg := tgbotapi.NewMessage(logChatID, text)
	msg.ParseMode = "Markdown"
	Bot.Send(msg)
	log.Printf("[GM] %d → %s", gmUserID, action)
}

// ── Sessão GM (armazena estado entre callbacks) ───────────

type gmSession struct {
	TargetCharID int
	PendingCmd   string // "ban_confirm" | "diamond_set" | "gold_set"
	PendingVal   string
}

var (
	gmSessions   = map[int64]*gmSession{}
	gmSessionsMu sync.Mutex
)

func getGMSession(gmID int64) *gmSession {
	gmSessionsMu.Lock()
	defer gmSessionsMu.Unlock()
	if gmSessions[gmID] == nil {
		gmSessions[gmID] = &gmSession{}
	}
	return gmSessions[gmID]
}

// ── Entry points ──────────────────────────────────────────

// HandleGMCommand processes /gm <subcommand> text messages.
func HandleGMCommand(msg *tgbotapi.Message) bool {
	// Support both "/gm cmd" and "/gm@BotName cmd"
	text := strings.TrimSpace(msg.Text)
	// Strip @BotName suffix from command: "/gm@TormentaBot painel" -> "/gm painel"
	if idx := strings.Index(text, "@"); idx > 0 && strings.HasPrefix(text, "/gm") {
		space := strings.Index(text[idx:], " ")
		if space < 0 {
			text = text[:idx]
		} else {
			text = text[:idx] + text[idx+space:]
		}
	}
	if !strings.HasPrefix(text, "/gm") {
		return false
	}
	// Log every /gm attempt so we can diagnose auth issues
	log.Printf("[GM] /gm recebido de userID=%d | isGM=%v | GM_IDS=%q | text=%q",
		msg.From.ID, isGM(msg.From.ID), os.Getenv("GM_IDS"), text)
	if !isGM(msg.From.ID) {
		sendText(msg.Chat.ID, "❌ Acesso negado. Seu ID não está na lista de GMs.")
		return true
	}
	msg.Text = text // use the normalized text for the rest of the function

	parts := strings.Fields(text)
	if len(parts) < 2 {
		showGMHelp(msg.Chat.ID)
		return true
	}

	cmd := strings.ToLower(parts[1])
	switch cmd {
	case "painel", "panel", "menu":
		showGMDashboard(msg.Chat.ID)

	case "buscar", "find", "search":
		if len(parts) < 3 {
			sendText(msg.Chat.ID, "❌ Uso: `/gm buscar <nome>`")
			return true
		}
		name := strings.Join(parts[2:], " ")
		gmFindAndShow(msg.Chat.ID, msg.From.ID, name)

	case "ban":
		if len(parts) < 3 {
			sendText(msg.Chat.ID, "❌ Uso: `/gm ban <nome> [razão]`")
			return true
		}
		// Try to split name and reason — last word that starts with # is reason
		args := parts[2:]
		name, reason := parseNameReason(args)
		if reason == "" {
			reason = "Violação dos termos de uso"
		}
		gmBanByName(msg.Chat.ID, msg.From.ID, name, reason)

	case "unban":
		if len(parts) < 3 {
			sendText(msg.Chat.ID, "❌ Uso: `/gm unban <nome>`")
			return true
		}
		name := strings.Join(parts[2:], " ")
		gmUnbanByName(msg.Chat.ID, msg.From.ID, name)

	case "diamond", "diamante", "gem":
		// /gm diamond <nome> <+/-amount>
		if len(parts) < 4 {
			sendText(msg.Chat.ID, "❌ Uso: `/gm diamond <nome> <+50 ou -10>`")
			return true
		}
		name := strings.Join(parts[2:len(parts)-1], " ")
		gmAdjustDiamond(msg.Chat.ID, msg.From.ID, name, parts[len(parts)-1])

	case "gold", "ouro":
		// /gm gold <nome> <+/-amount>
		if len(parts) < 4 {
			sendText(msg.Chat.ID, "❌ Uso: `/gm gold <nome> <+100 ou -50>`")
			return true
		}
		name := strings.Join(parts[2:len(parts)-1], " ")
		gmAdjustGold(msg.Chat.ID, msg.From.ID, name, parts[len(parts)-1])

	case "info":
		if len(parts) < 3 {
			sendText(msg.Chat.ID, "❌ Uso: `/gm info <nome>`")
			return true
		}
		name := strings.Join(parts[2:], " ")
		gmInfo(msg.Chat.ID, name)

	case "id":
		// /gm id <telegramID>  — lookup player by Telegram user ID
		if len(parts) < 3 {
			sendText(msg.Chat.ID, "❌ Uso: `/gm id <telegramID>`")
			return true
		}
		tgID, err := strconv.ParseInt(parts[2], 10, 64)
		if err != nil {
			sendText(msg.Chat.ID, "❌ ID inválido.")
			return true
		}
		gmInfoByTGID(msg.Chat.ID, tgID)

	case "banidos", "banned", "bans":
		gmShowBanned(msg.Chat.ID)

	case "jogadores", "players", "list":
		gmListPlayers(msg.Chat.ID)

	case "rank", "ranking":
		gmShowRanking(msg.Chat.ID)

	case "pix", "pagamentos":
		gmShowPixPayments(msg.Chat.ID)

	case "vip":
		// /gm vip <nome> <dias> — 0=permanente, -1=revogar
		if len(parts) < 4 {
			sendText(msg.Chat.ID, "❌ Uso: `/gm vip <nome> <dias>`\n`/gm vip <nome> 0` = permanente\n`/gm vip <nome> -1` = revogar")
			return true
		}
		name := strings.Join(parts[2:len(parts)-1], " ")
		days := 0
		fmt.Sscan(parts[len(parts)-1], &days)
		chars, _ := database.SearchCharacters(name, 1)
		if len(chars) == 0 {
			sendText(msg.Chat.ID, "❌ Personagem não encontrado: "+escMd(name))
			return true
		}
		c := chars[0]
		if days == -1 {
			SetVIPFromGM(c.ID, false, 0)
			gmLog(msg.From.ID, fmt.Sprintf("revoked VIP charID=%d", c.ID))
			sendMsg(msg.Chat.ID, fmt.Sprintf("✅ VIP revogado de *%s*", escMd(c.Name)), nil)
		} else {
			SetVIPFromGM(c.ID, true, days)
			gmLog(msg.From.ID, fmt.Sprintf("granted VIP %dd charID=%d", days, c.ID))
			dayStr := "permanente"
			if days > 0 {
				dayStr = fmt.Sprintf("%d dias", days)
			}
			sendMsg(msg.Chat.ID, fmt.Sprintf("👑 VIP ativado para *%s* (%s)", escMd(c.Name), dayStr), nil)
			notifyUser(int64(c.PlayerID), fmt.Sprintf("👑 *VIP Ativado pelo GM!*\n\nBenefícios por %s:\n⚡ Energia dobrada\n⏱️ Recarga 2x mais rápida\n🏹 Caça automática desbloqueada!", dayStr))
		}

	default:
		showGMHelp(msg.Chat.ID)
	}
	return true
}

// HandleGMCallback handles gm_ prefixed inline callbacks.
func HandleGMCallback(cb *tgbotapi.CallbackQuery) bool {
	if !isGM(cb.From.ID) || !strings.HasPrefix(cb.Data, "gm_") {
		return false
	}
	Bot.Request(tgbotapi.NewCallback(cb.ID, ""))

	chatID := cb.Message.Chat.ID
	msgID := cb.Message.MessageID
	gmID := cb.From.ID
	data := cb.Data

	switch {
	case data == "gm_menu":
		showGMDashboard(chatID)
	case data == "gm_players":
		gmListPlayersInline(chatID, msgID)
	case data == "gm_banned":
		gmShowBannedInline(chatID, msgID)
	case data == "gm_rank":
		gmShowRankingInline(chatID, msgID)
	case data == "gm_pix":
		gmShowPixInline(chatID, msgID)

	// Player panel by charID
	case strings.HasPrefix(data, "gm_panel_"):
		charID, _ := strconv.Atoi(strings.TrimPrefix(data, "gm_panel_"))
		gmShowPlayerPanel(chatID, msgID, gmID, charID)

	// Ban flow
	case strings.HasPrefix(data, "gm_ban_"):
		charID, _ := strconv.Atoi(strings.TrimPrefix(data, "gm_ban_"))
		gmInitBan(chatID, msgID, charID)
	case strings.HasPrefix(data, "gm_banconfirm_"):
		// format: gm_banconfirm_<charID>_<reasonIdx>
		parts := strings.SplitN(strings.TrimPrefix(data, "gm_banconfirm_"), "_", 2)
		charID, _ := strconv.Atoi(parts[0])
		reason := ""
		if len(parts) > 1 {
			reason = decodeBanReason(parts[1])
		}
		gmExecuteBan(chatID, msgID, gmID, charID, reason)
	case strings.HasPrefix(data, "gm_bancustom_"):
		charID, _ := strconv.Atoi(strings.TrimPrefix(data, "gm_bancustom_"))
		gmAskCustomBanReason(chatID, msgID, gmID, charID)

	// Unban
	case strings.HasPrefix(data, "gm_unban_"):
		charID, _ := strconv.Atoi(strings.TrimPrefix(data, "gm_unban_"))
		gmExecuteUnban(chatID, msgID, gmID, charID)

	// Diamond adjust
	case strings.HasPrefix(data, "gm_dplus_"):
		// gm_dplus_<charID>_<amount>
		parts := strings.SplitN(strings.TrimPrefix(data, "gm_dplus_"), "_", 2)
		charID, _ := strconv.Atoi(parts[0])
		amount, _ := strconv.Atoi(parts[1])
		gmDoAdjustDiamond(chatID, msgID, gmID, charID, amount)
	case strings.HasPrefix(data, "gm_dminus_"):
		parts := strings.SplitN(strings.TrimPrefix(data, "gm_dminus_"), "_", 2)
		charID, _ := strconv.Atoi(parts[0])
		amount, _ := strconv.Atoi(parts[1])
		gmDoAdjustDiamond(chatID, msgID, gmID, charID, -amount)
	case strings.HasPrefix(data, "gm_dpanel_"):
		charID, _ := strconv.Atoi(strings.TrimPrefix(data, "gm_dpanel_"))
		gmShowDiamondPanel(chatID, msgID, charID)

	// Gold adjust
	case strings.HasPrefix(data, "gm_gplus_"):
		parts := strings.SplitN(strings.TrimPrefix(data, "gm_gplus_"), "_", 2)
		charID, _ := strconv.Atoi(parts[0])
		amount, _ := strconv.Atoi(parts[1])
		gmDoAdjustGold(chatID, msgID, gmID, charID, amount)
	case strings.HasPrefix(data, "gm_gminus_"):
		parts := strings.SplitN(strings.TrimPrefix(data, "gm_gminus_"), "_", 2)
		charID, _ := strconv.Atoi(parts[0])
		amount, _ := strconv.Atoi(parts[1])
		gmDoAdjustGold(chatID, msgID, gmID, charID, -amount)
	case strings.HasPrefix(data, "gm_gpanel_"):
		charID, _ := strconv.Atoi(strings.TrimPrefix(data, "gm_gpanel_"))
		gmShowGoldPanel(chatID, msgID, charID)
	}
	return true
}

// ── Help / Dashboard ──────────────────────────────────────

func showGMHelp(chatID int64) {
	sendMsg(chatID,
		"🛡️ *Comandos GM*\n\n"+
			"`/gm painel` — painel interativo\n"+
			"`/gm buscar <nome>` — buscar jogador\n"+
			"`/gm info <nome>` — ficha completa\n"+
			"`/gm ban <nome> [razão]` — banir jogador\n"+
			"`/gm unban <nome>` — desbanir\n"+
			"`/gm diamond <nome> +50` — adicionar diamantes\n"+
			"`/gm diamond <nome> -10` — remover diamantes\n"+
			"`/gm gold <nome> +1000` — adicionar ouro\n"+
			"`/gm gold <nome> -500` — remover ouro\n"+
			"`/gm banidos` — ver banidos\n"+
			"`/gm jogadores` — listar jogadores\n"+
			"`/gm rank` — ranking global\n"+
			"`/gm pix` — pagamentos Pix recentes\n",
		nil,
	)
}

func showGMDashboard(chatID int64) {
	players, _ := database.GetAllPlayers(500)
	total := len(players)
	banned := 0
	for _, p := range players {
		if p.Banned {
			banned++
		}
	}
	text := fmt.Sprintf(
		"🛡️ *Painel GM*\n\n"+
			"👥 Jogadores registrados: *%d*\n"+
			"✅ Ativos: *%d* | 🚫 Banidos: *%d*\n\n"+
			"Use os botões abaixo ou `/gm buscar <nome>` para gerenciar jogadores.",
		total, total-banned, banned,
	)
	kb := menukit.GMDashboard()
	sendMsg(chatID, text, &kb)
}

// ── Player search & panel ─────────────────────────────────

func gmFindAndShow(chatID int64, gmID int64, name string) {
	char, _ := database.SearchCharacterFullByName(name)
	if char == nil {
		sendText(chatID, fmt.Sprintf("❌ Personagem *%s* não encontrado.\n\nVerifique o nome e tente novamente.", escMd(name)))
		return
	}
	getGMSession(gmID).TargetCharID = char.ID
	sendMsg(chatID, buildCharPanelText(char), buildCharPanelKB(char))
}

func gmShowPlayerPanel(chatID int64, msgID int, gmID int64, charID int) {
	char, err := database.GetCharacterFullByID(charID)
	if char == nil || err != nil {
		editMsg(chatID, msgID, "❌ Personagem não encontrado.", &backKeyboard)
		return
	}
	getGMSession(gmID).TargetCharID = charID
	editMsg(chatID, msgID, buildCharPanelText(char), buildCharPanelKB(char))
}

func buildCharPanelText(char *database.CharFull) string {
	r := game.Races[char.Race]
	c := game.Classes[char.Class]
	status := "✅ Ativo"
	if char.Banned {
		status = "🚫 *BANIDO*"
	}
	return fmt.Sprintf(
		"🛡️ *GM — %s*\n\n"+
			"Status: %s\n"+
			"TG ID: `%d` | Char ID: `%d`\n"+
			"%s %s | %s %s | Nv.*%d*\n\n"+
			"❤️ %d/%d HP | 💙 %d/%d MP\n"+
			"⚡ %d/%d Energia\n"+
			"🪙 *%d* ouro | 💎 *%d* diamantes\n"+
			"✨ XP: %d/%d | 💀 Mortes: %d",
		escMd(char.Name),
		status,
		char.PlayerID, char.ID,
		r.Emoji, char.Race, c.Emoji, char.Class, char.Level,
		char.HP, char.HPMax, char.MP, char.MPMax,
		char.Energy, char.EnergyMax,
		char.Gold, char.Diamonds,
		char.Experience, char.ExperienceNext, char.Deaths,
	)
}

func buildCharPanelKB(char *database.CharFull) *tgbotapi.InlineKeyboardMarkup {
	banLabel, banAction := "🚫 Banir", fmt.Sprintf("gm_ban_%d", char.ID)
	if char.Banned {
		banLabel, banAction = "✅ Desbanir", fmt.Sprintf("gm_unban_%d", char.ID)
	}
	kb := menukit.GMPlayerPanel(char.ID, banLabel, banAction)
	return &kb
}

// ── Ban flow ──────────────────────────────────────────────

var banReasonPresets = []string{
	"Trapaça / hacking",
	"Comportamento tóxico",
	"Spam / automação",
	"Fraude / abuso do sistema de pagamento",
	"Conteúdo impróprio",
	"Violação dos termos de uso",
}

func encodeBanReason(idx int) string { return strconv.Itoa(idx) }
func decodeBanReason(s string) string {
	idx, err := strconv.Atoi(s)
	if err != nil || idx < 0 || idx >= len(banReasonPresets) {
		return s // raw string fallback
	}
	return banReasonPresets[idx]
}

func gmInitBan(chatID int64, msgID int, charID int) {
	char, _ := database.GetCharacterFullByID(charID)
	if char == nil {
		editMsg(chatID, msgID, "❌ Personagem não encontrado.", &backKeyboard)
		return
	}
	if char.Banned {
		editMsg(chatID, msgID,
			fmt.Sprintf("⚠️ *%s* já está banido.\n\nDeseja *desbanir* este jogador?", escMd(char.Name)),
			buildConfirmKB(
				fmt.Sprintf("gm_unban_%d", charID), "✅ Sim, desbanir",
				fmt.Sprintf("gm_panel_%d", charID), "❌ Cancelar",
			),
		)
		return
	}

	text := fmt.Sprintf("🚫 *Banir: %s*\n\nSelecione o motivo do banimento:", escMd(char.Name))
	var rows [][]tgbotapi.InlineKeyboardButton
	for i, reason := range banReasonPresets {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"• "+reason,
				fmt.Sprintf("gm_banconfirm_%d_%s", charID, encodeBanReason(i)),
			),
		))
	}
	rows = append(rows,
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✏️ Motivo personalizado", fmt.Sprintf("gm_bancustom_%d", charID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❌ Cancelar", fmt.Sprintf("gm_panel_%d", charID)),
		),
	)
	kb := tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}
	editMsg(chatID, msgID, text, &kb)
}

func gmAskCustomBanReason(chatID int64, msgID int, gmID int64, charID int) {
	getGMSession(gmID).PendingCmd = fmt.Sprintf("ban_%d", charID)
	editMsg(chatID, msgID,
		fmt.Sprintf("✏️ *Motivo personalizado*\n\nDigite o motivo do banimento do personagem ID `%d`:\n\n_Responda a esta mensagem com o motivo._", charID),
		&tgbotapi.InlineKeyboardMarkup{InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
			{tgbotapi.NewInlineKeyboardButtonData("❌ Cancelar", fmt.Sprintf("gm_ban_%d", charID))},
		}},
	)
}

func gmExecuteBan(chatID int64, msgID int, gmID int64, charID int, reason string) {
	char, _ := database.GetCharacterFullByID(charID)
	if char == nil {
		editMsg(chatID, msgID, "❌ Personagem não encontrado.", &backKeyboard)
		return
	}
	if char.Banned {
		editMsg(chatID, msgID, "⚠️ Jogador já está banido.", &backKeyboard)
		return
	}
	if err := database.BanPlayer(char.PlayerID); err != nil {
		editMsg(chatID, msgID, "❌ Erro ao banir: "+err.Error(), &backKeyboard)
		return
	}
	if reason == "" {
		reason = "Violação dos termos de uso"
	}

	result := fmt.Sprintf(
		"🚫 *Banimento executado!*\n\n"+
			"👤 Personagem: *%s*\n"+
			"🆔 TG ID: `%d` | Char ID: `%d`\n"+
			"📝 Motivo: _%s_",
		char.Name, char.PlayerID, char.ID, reason,
	)
	editMsg(chatID, msgID, result,
		func() *tgbotapi.InlineKeyboardMarkup {
			kb := menukit.GMPlayerResult(charID)
			return &kb
		}(),
	)

	gmLog(gmID, fmt.Sprintf("BAN char=%d tg=%d name=%s reason=%s", charID, char.PlayerID, char.Name, reason))

	// Notify the banned player
	notification := fmt.Sprintf(
		"🚫 *Conta suspensa*\n\n"+
			"Sua conta foi banida.\n"+
			"📝 Motivo: _%s_\n\n"+
			"Se acredita que foi um engano, entre em contato com o suporte.",
		reason,
	)
	notifyPlayer(char.PlayerID, notification)
}

func gmExecuteUnban(chatID int64, msgID int, gmID int64, charID int) {
	char, _ := database.GetCharacterFullByID(charID)
	if char == nil {
		editMsg(chatID, msgID, "❌ Personagem não encontrado.", &backKeyboard)
		return
	}
	if !char.Banned {
		editMsg(chatID, msgID, "⚠️ Jogador não está banido.", &backKeyboard)
		return
	}
	if err := database.UnbanPlayer(char.PlayerID); err != nil {
		editMsg(chatID, msgID, "❌ Erro ao desbanir: "+err.Error(), &backKeyboard)
		return
	}

	result := fmt.Sprintf(
		"✅ *Desbanimento executado!*\n\n"+
			"👤 Personagem: *%s*\n"+
			"🆔 TG ID: `%d` | Char ID: `%d`",
		char.Name, char.PlayerID, char.ID,
	)
	editMsg(chatID, msgID, result,
		func() *tgbotapi.InlineKeyboardMarkup {
			kb := menukit.GMPlayerResult(charID)
			return &kb
		}(),
	)
	gmLog(gmID, fmt.Sprintf("UNBAN char=%d tg=%d name=%s", charID, char.PlayerID, char.Name))
	notifyPlayer(char.PlayerID, "✅ *Sua conta foi reativada!*\n\nBoas aventuras em Tormenta RPG! 🗡️")
}

// ── Diamond panel ─────────────────────────────────────────

func gmShowDiamondPanel(chatID int64, msgID int, charID int) {
	char, _ := database.GetCharacterFullByID(charID)
	if char == nil {
		editMsg(chatID, msgID, "❌ Personagem não encontrado.", &backKeyboard)
		return
	}
	text := fmt.Sprintf(
		"💎 *Ajustar Diamantes — %s*\n\n"+
			"💎 Saldo atual: *%d* diamantes\n\n"+
			"Selecione o ajuste:",
		char.Name, char.Diamonds,
	)
	kb := menukit.GMDiamondPanel(charID)
	editMsg(chatID, msgID, text, &kb)
}

func gmDoAdjustDiamond(chatID int64, msgID int, gmID int64, charID int, delta int) {
	char, _ := database.GetCharacterFullByID(charID)
	if char == nil {
		editMsg(chatID, msgID, "❌ Personagem não encontrado.", &backKeyboard)
		return
	}
	newVal := char.Diamonds + delta
	if newVal < 0 {
		newVal = 0
	}
	if err := database.GMSetDiamonds(char.ID, newVal); err != nil {
		editMsg(chatID, msgID, "❌ Erro: "+err.Error(), &backKeyboard)
		return
	}
	if delta > 0 {
		database.LogDiamond(char.ID, delta, "gm_grant")
	}

	sign := "+"
	if delta < 0 {
		sign = ""
	}
	gmLog(gmID, fmt.Sprintf("DIAMOND char=%d name=%s %s%d → %d", charID, char.Name, sign, delta, newVal))

	// Reload and re-show panel
	char.Diamonds = newVal
	text := fmt.Sprintf(
		"💎 *Ajustar Diamantes — %s*\n\n"+
			"💎 Saldo atual: *%d* diamantes\n"+
			"✅ Ajuste: *%s%d* (era %d)\n\n"+
			"Selecione o ajuste:",
		char.Name, newVal, sign, delta, char.Diamonds,
	)
	kb := menukit.GMDiamondPanel(charID)
	editMsg(chatID, msgID, text, &kb)

	// Notify the player silently
	if delta > 0 {
		notifyPlayer(char.PlayerID, fmt.Sprintf("🎁 *+%d diamantes* foram adicionados à sua conta pelo suporte!\n\n💎 Saldo atual: *%d*", delta, newVal))
	}
}

// ── Gold panel ────────────────────────────────────────────

func gmShowGoldPanel(chatID int64, msgID int, charID int) {
	char, _ := database.GetCharacterFullByID(charID)
	if char == nil {
		editMsg(chatID, msgID, "❌ Personagem não encontrado.", &backKeyboard)
		return
	}
	text := fmt.Sprintf(
		"🪙 *Ajustar Ouro — %s*\n\n"+
			"🪙 Saldo atual: *%d* ouro\n\n"+
			"Selecione o ajuste:",
		char.Name, char.Gold,
	)
	kb := buildGoldPanelKB(charID)
	editMsg(chatID, msgID, text, &kb)
}

func buildGoldPanelKB(charID int) tgbotapi.InlineKeyboardMarkup {
	return menukit.GMGoldPanel(charID)
}

func gmDoAdjustGold(chatID int64, msgID int, gmID int64, charID int, delta int) {
	char, _ := database.GetCharacterFullByID(charID)
	if char == nil {
		editMsg(chatID, msgID, "❌ Personagem não encontrado.", &backKeyboard)
		return
	}
	newVal := char.Gold + delta
	if newVal < 0 {
		newVal = 0
	}
	if err := database.GMSetGold(char.ID, newVal); err != nil {
		editMsg(chatID, msgID, "❌ Erro: "+err.Error(), &backKeyboard)
		return
	}

	sign := "+"
	if delta < 0 {
		sign = ""
	}
	gmLog(gmID, fmt.Sprintf("GOLD char=%d name=%s %s%d → %d", charID, char.Name, sign, delta, newVal))

	text := fmt.Sprintf(
		"🪙 *Ajustar Ouro — %s*\n\n"+
			"🪙 Saldo atual: *%d* ouro\n"+
			"✅ Ajuste: *%s%d* (era %d)\n\n"+
			"Selecione o ajuste:",
		char.Name, newVal, sign, delta, char.Gold,
	)
	kb := buildGoldPanelKB(charID)
	editMsg(chatID, msgID, text, &kb)
}

// ── List views ────────────────────────────────────────────

func gmListPlayersInline(chatID int64, msgID int) {
	players, _ := database.GetAllPlayers(50)
	text := fmt.Sprintf("👥 *Jogadores (%d)*\n\n", len(players))
	for _, p := range players {
		status := "✅"
		if p.Banned {
			status = "🚫"
		}
		uname := p.Username
		if uname == "" {
			uname = "sem usuário"
		}
		text += fmt.Sprintf("%s `%d` @%s\n", status, p.ID, escMd(uname))
	}
	editMsg(chatID, msgID, text, &tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: menukit.GMBackToDashboard().InlineKeyboard,
	})
}

func gmShowBannedInline(chatID int64, msgID int) {
	players, _ := database.GetAllPlayers(500)
	text := "🚫 *Jogadores Banidos*\n\n"
	count := 0
	var rows [][]tgbotapi.InlineKeyboardButton
	for _, p := range players {
		if !p.Banned {
			continue
		}
		count++
		uname := p.Username
		if uname == "" {
			uname = fmt.Sprintf("ID_%d", p.ID)
		}
		text += fmt.Sprintf("• `%d` @%s\n", p.ID, escMd(uname))
	}
	if count == 0 {
		text += "_Nenhum jogador banido._"
	} else {
		text += fmt.Sprintf("\nTotal: *%d*", count)
	}
	rows = append(rows, menukit.GMBackToDashboard().InlineKeyboard...)
	kb := tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}
	editMsg(chatID, msgID, text, &kb)
}

func gmShowRankingInline(chatID int64, msgID int) {
	entries, _ := rankingSvc.Top(15)
	text := "🏆 *Top 15 Jogadores*\n\n"
	for _, e := range entries {
		text += fmt.Sprintf("#%d %s Nv.%d | XP:%d | PVP:%d\n",
			e.Position, escMd(e.Name), e.Level, e.Score, e.PVPRating)
	}
	editMsg(chatID, msgID, text, &tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: menukit.GMBackToDashboard().InlineKeyboard,
	})
}

// ── PIX payments panel ────────────────────────────────────

func gmShowPixInline(chatID int64, msgID int) {
	payments, _ := database.GetAllPendingPixPayments()
	recent, _ := database.GetRecentPixPayments(20)

	text := "💳 *Pagamentos Pix*\n\n"

	if len(payments) > 0 {
		text += fmt.Sprintf("⏳ *Pendentes: %d*\n", len(payments))
		for _, p := range payments {
			if pkg := game.GetDiamondPackage(p.PackageID); pkg != nil {
				text += fmt.Sprintf("  • `%s` — %s R$ %.2f\n", p.TxID[:12], pkg.Name, p.AmountBRL)
			}
		}
		text += "\n"
	} else {
		text += "✅ Nenhum pagamento pendente\n\n"
	}

	if len(recent) > 0 {
		text += "*Últimos pagamentos:*\n"
		for _, p := range recent {
			status := "⏳"
			switch p.Status {
			case "paid":
				status = "✅"
			case "expired", "cancelled":
				status = "❌"
			}
			if pkg := game.GetDiamondPackage(p.PackageID); pkg != nil {
				text += fmt.Sprintf("%s `%d` — %s R$ %.2f\n",
					status, p.CharacterID, pkg.Name, p.AmountBRL)
			}
		}
	}

	editMsg(chatID, msgID, text, func() *tgbotapi.InlineKeyboardMarkup {
		kb := menukit.GMPixInline()
		return &kb
	}())
}

func gmShowPixPayments(chatID int64) {
	payments, _ := database.GetAllPendingPixPayments()
	recent, _ := database.GetRecentPixPayments(20)

	text := "💳 *Pagamentos Pix*\n\n"

	if len(payments) > 0 {
		text += fmt.Sprintf("⏳ *Pendentes: %d*\n", len(payments))
		for _, p := range payments {
			if pkg := game.GetDiamondPackage(p.PackageID); pkg != nil {
				text += fmt.Sprintf("  • `%s` — %s R$ %.2f\n", p.TxID[:12], pkg.Name, p.AmountBRL)
			}
		}
		text += "\n"
	} else {
		text += "✅ Nenhum pagamento pendente\n\n"
	}

	if len(recent) > 0 {
		text += "*Últimos 20 pagamentos:*\n"
		for _, p := range recent {
			status := "⏳"
			switch p.Status {
			case "paid":
				status = "✅"
			case "expired", "cancelled":
				status = "❌"
			}
			if pkg := game.GetDiamondPackage(p.PackageID); pkg != nil {
				text += fmt.Sprintf("%s CharID `%d` — %s R$ %.2f\n",
					status, p.CharacterID, pkg.Name, p.AmountBRL)
			}
		}
	}

	sendMsg(chatID, text, nil)
}

// NotifyGMPixPaid envia notificação ao GM quando um Pix é confirmado.
func NotifyGMPixPaid(charID, diamonds int, packageID string, amountBRL float64) {
	gmIDs := os.Getenv("GM_IDS")
	if gmIDs == "" {
		return
	}
	pkgName := packageID
	if pkg := game.GetDiamondPackage(packageID); pkg != nil {
		pkgName = pkg.Name
	}
	text := fmt.Sprintf(
		"💳 *Pix Confirmado!*\n\n"+
			"CharID: `%d`\n"+
			"Pacote: %s\n"+
			"💎 Diamantes: *+%d*\n"+
			"💵 Valor: R$ %.2f",
		charID, pkgName, diamonds, amountBRL,
	)
	for _, part := range strings.Split(gmIDs, ",") {
		part = strings.TrimSpace(part)
		if gmID, err := strconv.ParseInt(part, 10, 64); err == nil {
			msg := tgbotapi.NewMessage(gmID, text)
			msg.ParseMode = "Markdown"
			Bot.Send(msg)
		}
	}
}

// ── Text command variants (for /gm <cmd>) ─────────────────

func gmBanByName(chatID int64, gmID int64, name string, reason string) {
	char, _ := database.SearchCharacterFullByName(name)
	if char == nil {
		sendText(chatID, fmt.Sprintf("❌ Personagem *%s* não encontrado.", escMd(name)))
		return
	}
	if char.Banned {
		sendText(chatID, fmt.Sprintf("⚠️ *%s* já está banido.", escMd(char.Name)))
		return
	}
	if err := database.BanPlayer(char.PlayerID); err != nil {
		sendText(chatID, "❌ Erro ao banir: "+err.Error())
		return
	}
	gmLog(gmID, fmt.Sprintf("BAN char=%d tg=%d name=%s reason=%s", char.ID, char.PlayerID, char.Name, reason))
	sendText(chatID, fmt.Sprintf("🚫 *%s* banido!\n📝 Motivo: _%s_", escMd(char.Name), escMd(reason)))
	notifyPlayer(char.PlayerID, fmt.Sprintf("🚫 *Conta suspensa*\n📝 Motivo: _%s_\n\nEntre em contato com o suporte se acredita que foi um engano.", reason))
}

func gmUnbanByName(chatID int64, gmID int64, name string) {
	char, _ := database.SearchCharacterFullByName(name)
	if char == nil {
		sendText(chatID, fmt.Sprintf("❌ Personagem *%s* não encontrado.", escMd(name)))
		return
	}
	if !char.Banned {
		sendText(chatID, fmt.Sprintf("⚠️ *%s* não está banido.", escMd(char.Name)))
		return
	}
	if err := database.UnbanPlayer(char.PlayerID); err != nil {
		sendText(chatID, "❌ Erro ao desbanir: "+err.Error())
		return
	}
	gmLog(gmID, fmt.Sprintf("UNBAN char=%d tg=%d name=%s", char.ID, char.PlayerID, char.Name))
	sendText(chatID, fmt.Sprintf("✅ *%s* desbanido!", escMd(char.Name)))
	notifyPlayer(char.PlayerID, "✅ *Sua conta foi reativada!*\n\nBoas aventuras! 🗡️")
}

func gmAdjustDiamond(chatID int64, gmID int64, name string, amtStr string) {
	char, _ := database.SearchCharacterFullByName(name)
	if char == nil {
		sendText(chatID, fmt.Sprintf("❌ Personagem *%s* não encontrado.", escMd(name)))
		return
	}
	// Strip + sign for Atoi
	amtStr = strings.TrimPrefix(amtStr, "+")
	amount, err := strconv.Atoi(amtStr)
	if err != nil {
		sendText(chatID, "❌ Valor inválido. Exemplos: `+50` ou `-10`")
		return
	}
	newVal := char.Diamonds + amount
	if newVal < 0 {
		newVal = 0
	}
	if err := database.GMSetDiamonds(char.ID, newVal); err != nil {
		sendText(chatID, "❌ Erro: "+err.Error())
		return
	}
	if amount > 0 {
		database.LogDiamond(char.ID, amount, "gm_grant")
	}
	sign := "+"
	if amount < 0 {
		sign = ""
	}
	gmLog(gmID, fmt.Sprintf("DIAMOND char=%d name=%s %s%d → %d", char.ID, char.Name, sign, amount, newVal))
	sendText(chatID, fmt.Sprintf("✅ *%s* | 💎 %s%d diamantes → Saldo: *%d* 💎", escMd(char.Name), sign, amount, newVal))
	if amount > 0 {
		notifyPlayer(char.PlayerID, fmt.Sprintf("🎁 *+%d diamantes* adicionados pelo suporte!\n💎 Saldo: *%d*", amount, newVal))
	}
}

func gmAdjustGold(chatID int64, gmID int64, name string, amtStr string) {
	char, _ := database.SearchCharacterFullByName(name)
	if char == nil {
		sendText(chatID, fmt.Sprintf("❌ Personagem *%s* não encontrado.", escMd(name)))
		return
	}
	amtStr = strings.TrimPrefix(amtStr, "+")
	amount, err := strconv.Atoi(amtStr)
	if err != nil {
		sendText(chatID, "❌ Valor inválido. Exemplos: `+1000` ou `-500`")
		return
	}
	newVal := char.Gold + amount
	if newVal < 0 {
		newVal = 0
	}
	if err := database.GMSetGold(char.ID, newVal); err != nil {
		sendText(chatID, "❌ Erro: "+err.Error())
		return
	}
	sign := "+"
	if amount < 0 {
		sign = ""
	}
	gmLog(gmID, fmt.Sprintf("GOLD char=%d name=%s %s%d → %d", char.ID, char.Name, sign, amount, newVal))
	sendText(chatID, fmt.Sprintf("✅ *%s* | 🪙 %s%d ouro → Saldo: *%d* 🪙", escMd(char.Name), sign, amount, newVal))
}

func gmInfo(chatID int64, name string) {
	char, _ := database.SearchCharacterFullByName(name)
	if char == nil {
		sendText(chatID, fmt.Sprintf("❌ Personagem *%s* não encontrado.", escMd(name)))
		return
	}
	pvpStats, _ := database.GetOrCreatePVPStats(char.ID)
	r := game.Races[char.Race]
	c := game.Classes[char.Class]
	status := "✅ Ativo"
	if char.Banned {
		status = "🚫 BANIDO"
	}
	text := fmt.Sprintf(
		"📋 *Info GM — %s*\n\n"+
			"Status: %s | TG ID: `%d` | Char ID: `%d`\n"+
			"%s %s Nv.*%d* | %s %s\n"+
			"❤️ %d/%d HP | 💙 %d/%d MP\n"+
			"⚡ %d/%d Energia\n"+
			"🪙 %d ouro | 💎 %d diamantes\n"+
			"✨ XP: %d/%d | 💀 Mortes: %d\n"+
			"⚔️ PVP: Rating *%d* | W:%d L:%d D:%d\n"+
			"📍 Mapa: %s",
		char.Name, status,
		char.PlayerID, char.ID,
		r.Emoji, char.Race, char.Level, c.Emoji, char.Class,
		char.HP, char.HPMax, char.MP, char.MPMax,
		char.Energy, char.EnergyMax,
		char.Gold, char.Diamonds,
		char.Experience, char.ExperienceNext, char.Deaths,
		pvpStats.Rating, pvpStats.Wins, pvpStats.Losses, pvpStats.Draws,
		char.CurrentMap,
	)
	sendMsg(chatID, text, buildCharPanelKB(char))
}

func gmShowRanking(chatID int64) {
	entries, _ := rankingSvc.Top(15)
	text := "🏆 *Top 15 — GM View*\n\n"
	for _, e := range entries {
		text += fmt.Sprintf("#%d %s Nv.%d | XP:%d | PVP:%d\n",
			e.Position, escMd(e.Name), e.Level, e.Score, e.PVPRating)
	}
	sendMsg(chatID, text, nil)
}

func gmShowBanned(chatID int64) {
	players, _ := database.GetAllPlayers(500)
	text := "🚫 *Jogadores Banidos*\n\n"
	count := 0
	for _, p := range players {
		if p.Banned {
			uname := p.Username
			if uname == "" {
				uname = fmt.Sprintf("ID_%d", p.ID)
			}
			text += fmt.Sprintf("• `%d` @%s\n", p.ID, uname)
			count++
		}
	}
	if count == 0 {
		text += "_Nenhum jogador banido._"
	} else {
		text += fmt.Sprintf("\nTotal: *%d*", count)
	}
	sendMsg(chatID, text, nil)
}

func gmListPlayers(chatID int64) {
	players, _ := database.GetAllPlayers(50)
	text := fmt.Sprintf("👥 *Jogadores (%d mais recentes)*\n\n", len(players))
	for _, p := range players {
		s := "✅"
		if p.Banned {
			s = "🚫"
		}
		uname := p.Username
		if uname == "" {
			uname = "—"
		}
		text += fmt.Sprintf("%s `%d` @%s\n", s, p.ID, escMd(uname))
	}
	sendMsg(chatID, text, nil)
}

// ── Utilities ─────────────────────────────────────────────

// parseNameReason splits ["Aragorn", "de", "Aragor", "#trapaça"] → name="Aragorn de Aragor", reason="trapaça"
func parseNameReason(args []string) (name, reason string) {
	for i, a := range args {
		if strings.HasPrefix(a, "#") {
			name = strings.Join(args[:i], " ")
			reason = strings.Join(args[i:], " ")
			reason = strings.TrimPrefix(reason, "#")
			return
		}
	}
	return strings.Join(args, " "), ""
}

func buildConfirmKB(yesAction, yesLabel, noAction, noLabel string) *tgbotapi.InlineKeyboardMarkup {
	kb := menukit.GMConfirm(yesAction, yesLabel, noAction, noLabel)
	return &kb
}

func notifyPlayer(playerID int64, text string) {
	if playerID == 0 {
		return
	}
	msg := tgbotapi.NewMessage(playerID, text)
	msg.ParseMode = "Markdown"
	if _, err := Bot.Send(msg); err != nil {
		log.Printf("[GM] notifyPlayer %d: %v", playerID, err)
	}
}

// escMd escapes Markdown special characters in user-provided strings.
func escMd(s string) string {
	return strings.NewReplacer("*", "\\*", "_", "\\_", "`", "\\`", "[", "\\[").Replace(s)
}

// ── GM Text Input (custom ban reason) ────────────────────

// handleGMTextInput is called before regular text handling.
// Returns true if the text was consumed by a pending GM command.
func handleGMTextInput(msg *tgbotapi.Message) bool {
	gmID := msg.From.ID
	sess := getGMSession(gmID)
	if sess.PendingCmd == "" {
		return false
	}

	// Handle "ban_<charID>" pending session
	if strings.HasPrefix(sess.PendingCmd, "ban_") {
		charIDStr := strings.TrimPrefix(sess.PendingCmd, "ban_")
		charID, err := strconv.Atoi(charIDStr)
		if err != nil {
			sess.PendingCmd = ""
			return false
		}
		reason := strings.TrimSpace(msg.Text)
		if reason == "" {
			sendText(msg.Chat.ID, "❌ Motivo não pode ser vazio. Tente novamente:")
			return true
		}
		sess.PendingCmd = ""
		// Execute ban with the custom reason
		char, _ := database.GetCharacterFullByID(charID)
		if char == nil {
			sendText(msg.Chat.ID, "❌ Personagem não encontrado.")
			return true
		}
		if err := database.BanPlayer(char.PlayerID); err != nil {
			sendText(msg.Chat.ID, "❌ Erro ao banir: "+err.Error())
			return true
		}
		gmLog(gmID, fmt.Sprintf("BAN char=%d tg=%d name=%s reason=%s", charID, char.PlayerID, char.Name, reason))
		sendText(msg.Chat.ID, fmt.Sprintf(
			"🚫 *Banimento executado!*\n\n👤 *%s*\n📝 Motivo: _%s_", char.Name, reason))
		notifyPlayer(char.PlayerID, fmt.Sprintf(
			"🚫 *Conta suspensa*\n📝 Motivo: _%s_\n\nEntre em contato com o suporte se acredita que foi um engano.", reason))
		return true
	}

	return false
}

// ── Lookup by Telegram ID ─────────────────────────────────

func gmInfoByTGID(chatID int64, tgID int64) {
	char, err := database.GetCharacter(tgID)
	if char == nil || err != nil {
		sendText(chatID, fmt.Sprintf("❌ Nenhum personagem encontrado para TG ID `%d`.", tgID))
		return
	}
	banned := database.IsPlayerBanned(tgID)
	charFull := &database.CharFull{Character: char, Banned: banned}
	sendMsg(chatID, buildCharPanelText(charFull), buildCharPanelKB(charFull))
}
