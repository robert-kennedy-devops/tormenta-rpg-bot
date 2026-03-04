package handlers

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/tormenta-bot/internal/database"
	"github.com/tormenta-bot/internal/game"
)

// =============================================
// LOJA DE DIAMANTES (tela inicial)
// =============================================

func showPixShop(chatID int64, msgID int, userID int64) {
	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}

	pending, _ := database.GetPendingPixPayments(char.ID)
	pendingSection := ""
	if len(pending) > 0 {
		pendingSection = fmt.Sprintf("\u23f3 *%d pagamento(s) pendente(s)*\n", len(pending))
		for _, p := range pending {
			if pkg := game.GetDiamondPackage(p.PackageID); pkg != nil {
				pendingSection += fmt.Sprintf("  \u2022 %s %s \u2014 R$ %.2f\n", pkg.Emoji, pkg.Name, p.AmountBRL)
			}
		}
		pendingSection += "\n"
	}

	caption := fmt.Sprintf(
		"\U0001f48e *Comprar Diamantes via Pix*\n\n\U0001f48e Seus diamantes: *%d*\n\n%s*Selecione o pacote:*\n\n",
		char.Diamonds, pendingSection,
	)
	for _, pkg := range game.DiamondPackages {
		bonus := ""
		if pkg.Bonus > 0 {
			bonus = fmt.Sprintf(" _(+%d b\u00f4nus)_", pkg.Bonus)
		}
		caption += fmt.Sprintf(
			"%s *%s* \u2014 *%d* \U0001f48e%s\n   \U0001f4b3 *%s*\n\n",
			pkg.Emoji, pkg.Name, pkg.Amount, bonus, pkg.Price,
		)
	}
	caption += "_Diamantes liberados em segundos ap\u00f3s confirma\u00e7\u00e3o do Pix._"

	var rows [][]tgbotapi.InlineKeyboardButton
	for _, pkg := range game.DiamondPackages {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("%s %s \u2014 %d\U0001f48e por %s", pkg.Emoji, pkg.Name, pkg.Amount+pkg.Bonus, pkg.Price),
				"pix_buy_"+pkg.ID,
			),
		))
	}
	if len(pending) > 0 {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("\U0001f504 Verificar Pagamento", "pix_check"),
		))
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("\U0001f4e6 Voltar", "menu_diamonds"),
	))
	kb := tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}
	editPhoto(chatID, msgID, "shop", caption, &kb)
}

// =============================================
// GERAR PAGAMENTO PIX VIA MERCADO PAGO
// =============================================

func handlePixBuy(chatID int64, msgID int, userID int64, packageID string) {
	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}
	pkg := game.GetDiamondPackage(packageID)
	if pkg == nil {
		editPhoto(chatID, msgID, "shop", "\u274c Pacote n\u00e3o encontrado!", &backKeyboard)
		return
	}

	// Show loading state immediately
	editPhoto(chatID, msgID, "shop",
		fmt.Sprintf("\u23f3 *Gerando Pix...*\n\n%s *%s*\n\U0001f4b5 *%s*\n\n_Aguarde um instante..._", pkg.Emoji, pkg.Name, pkg.Price),
		nil)

	// Call Mercado Pago API
	result, err := game.CreateMPPixPayment(pkg, char.ID)
	if err != nil {
		log.Printf("[PIX] CreateMPPixPayment error for user %d: %v", userID, err)
		editPhoto(chatID, msgID, "shop",
			"\u274c *Erro ao gerar Pix.*\n\nO servi\u00e7o de pagamento est\u00e1 indispon\u00edvel no momento. Tente novamente em instantes.",
			&backKeyboard)
		return
	}

	total := pkg.Amount + pkg.Bonus
	if err := database.CreatePixPayment(char.ID, packageID, total, pkg.PriceBRL,
		result.TxID, result.PaymentID, result.QRCode, result.QRCodeB64); err != nil {
		log.Printf("[PIX] CreatePixPayment DB error: %v", err)
		editPhoto(chatID, msgID, "shop", "\u274c Erro interno. Tente novamente.", &backKeyboard)
		return
	}

	bonus := ""
	if pkg.Bonus > 0 {
		bonus = fmt.Sprintf(" (+%d b\u00f4nus)", pkg.Bonus)
	}
	expireStr := result.ExpiresAt.In(time.FixedZone("BRT", -3*3600)).Format("15:04")

	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("\U0001f504 Verificar Pagamento", "pix_check"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("\U0001f4b3 Outro Pacote", "menu_pix"),
			tgbotapi.NewInlineKeyboardButtonData("\U0001f3f0 Menu", "menu_main"),
		),
	)

	// Try to send QR code as image (best UX) 
	if result.QRCodeB64 != "" {
		if imgData, err2 := base64.StdEncoding.DecodeString(result.QRCodeB64); err2 == nil {
			caption := fmt.Sprintf(
				"📸 *QR Code Pix*\n\n%s *%s*\n💎 *%d* diamantes%s\n💵 *%s*\n\n"+
					"*1.* Abra seu banco\n*2.* Escaneie o QR Code *ou* use Pix Copia e Cola\n\n"+
					"⏰ Válido até *%s* (horário de Brasília)",
				pkg.Emoji, pkg.Name, total, bonus, pkg.Price, expireStr,
			)
			sendPixQRPhoto(chatID, msgID, imgData, caption, result.QRCode, &kb)
			return
		}
	}

	// Fallback: se tiver URL da imagem, tenta baixar diretamente
	if result.QRCodeURL != "" {
		if imgData, err2 := game.FetchQRCodeImagePublic(result.QRCodeURL); err2 == nil {
			caption := fmt.Sprintf(
				"📸 *QR Code Pix*\n\n%s *%s*\n💎 *%d* diamantes%s\n💵 *%s*\n\n"+
					"*1.* Abra seu banco\n*2.* Escaneie o QR Code *ou* use Pix Copia e Cola\n\n"+
					"⏰ Válido até *%s* (horário de Brasília)",
				pkg.Emoji, pkg.Name, total, bonus, pkg.Price, expireStr,
			)
			sendPixQRPhoto(chatID, msgID, imgData, caption, result.QRCode, &kb)
			return
		}
	}

	// Fallback: text-only with copia e cola
	caption := fmt.Sprintf(
		"\U0001f4b3 *Pix Gerado!*\n\n%s *%s*\n\U0001f48e *%d* diamantes%s\n\U0001f4b5 *%s*\n\n"+
			"*Pix Copia e Cola:*\n```\n%s\n```\n\n"+
			"\u23f0 V\u00e1lido at\u00e9 *%s* (hor\u00e1rio de Bras\u00edlia)\n"+
			"\U0001f511 ID: `%d`",
		pkg.Emoji, pkg.Name, total, bonus, pkg.Price,
		game.FormatPixCode(result.QRCode), expireStr, result.PaymentID,
	)
	editPhoto(chatID, msgID, "shop", caption, &kb)
}

// sendPixQRPhoto sends the QR code PNG as a Telegram photo,
// then follows with a separate message containing the copia-e-cola code.
func sendPixQRPhoto(chatID int64, oldMsgID int, imgData []byte, caption string, pixCode string, kb *tgbotapi.InlineKeyboardMarkup) {
	// Delete old "loading" message
	Bot.Request(tgbotapi.NewDeleteMessage(chatID, oldMsgID))

	// Send QR code as photo
	photoFile := tgbotapi.FileBytes{Name: "qrcode_pix.png", Bytes: imgData}
	photo := tgbotapi.NewPhoto(chatID, photoFile)
	photo.Caption = caption
	photo.ParseMode = "Markdown"
	photo.ReplyMarkup = kb
	if msg, err := Bot.Send(photo); err == nil {
		markAsPhoto(chatID, msg.MessageID)
	}

	// Send code in a separate message so user can tap & copy easily
	codeMsg := tgbotapi.NewMessage(chatID,
		"\U0001f4cb *Pix Copia e Cola:*\n```\n"+pixCode+"\n```\n_Toque para copiar, depois cole no seu banco._")
	codeMsg.ParseMode = "Markdown"
	Bot.Send(codeMsg)
}

// =============================================
// VERIFICAR PAGAMENTOS PENDENTES
// =============================================

func handlePixCheck(chatID int64, msgID int, userID int64) {
	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}
	pending, _ := database.GetPendingPixPayments(char.ID)
	if len(pending) == 0 {
		editPhoto(chatID, msgID, "shop",
			"\u2705 *Nenhum pagamento pendente!*\n\nSeus diamantes foram creditados ou o pagamento expirou.",
			&tgbotapi.InlineKeyboardMarkup{InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
				{tgbotapi.NewInlineKeyboardButtonData("\U0001f48e Comprar Diamantes", "menu_pix")},
				{tgbotapi.NewInlineKeyboardButtonData("\U0001f3f0 Menu", "menu_main")},
			}})
		return
	}

	// Verifica cada pagamento pendente no AbacatePay
	credited := 0
	for _, p := range pending {
		if p.TxID == "" {
			continue
		}
		status, err := game.CheckAbacatePayStatus(p.TxID)
		if err != nil {
			log.Printf("[PIX] CheckAbacatePayStatus %s: %v", p.TxID, err)
			continue
		}
		if status == "approved" {
			charID, diamonds, err2 := database.ConfirmPixPaymentByTxID(p.TxID)
			if err2 == nil && diamonds > 0 {
				credited += diamonds
				database.LogDiamond(charID, diamonds, "pix_abacatepay")
				log.Printf("[PIX] Credited %d diamonds to charID %d via check (txID=%s)", diamonds, charID, p.TxID)
			}
		}
	}

	if credited > 0 {
		char, _ = database.GetCharacter(userID) // reload with updated diamonds
		editPhoto(chatID, msgID, "shop",
			fmt.Sprintf("\u2705 *Pagamento Confirmado!*\n\n\U0001f48e *+%d diamantes* adicionados!\n\nTotal: *%d* \U0001f48e\n\n_Boas aventuras!_",
				credited, char.Diamonds),
			&tgbotapi.InlineKeyboardMarkup{InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
				{tgbotapi.NewInlineKeyboardButtonData("\U0001f3f0 Menu Principal", "menu_main")},
			}})
		return
	}

	// Still pending — show status with refresh button
	caption := fmt.Sprintf("\u23f3 *Aguardando pagamento* (%d)\n\n", len(pending))
	for _, p := range pending {
		if pkg := game.GetDiamondPackage(p.PackageID); pkg != nil {
			timeLeft := ""
			caption += fmt.Sprintf("\u2022 %s *%s* \u2014 R$ %.2f\n  \U0001f48e %d diamantes%s\n  ID: `%d`\n\n",
				pkg.Emoji, pkg.Name, p.AmountBRL, p.Diamonds, timeLeft, p.MPPaymentID)
		}
	}
	caption += "_Confirmado automaticamente ap\u00f3s o Pix. Toque em verificar novamente se j\u00e1 pagou._"

	rows := [][]tgbotapi.InlineKeyboardButton{
		{tgbotapi.NewInlineKeyboardButtonData("\U0001f504 Verificar Novamente", "pix_check")},
		{tgbotapi.NewInlineKeyboardButtonData("\U0001f4b3 Outro Pacote", "menu_pix"),
			tgbotapi.NewInlineKeyboardButtonData("\U0001f3f0 Menu", "menu_main")},
	}
	// DEV mode: manual confirm button
	if os.Getenv("PIX_MANUAL_CONFIRM") == "true" && len(pending) > 0 {
		rows = append([][]tgbotapi.InlineKeyboardButton{
			{tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("\U0001f9ea [DEV] Confirmar ID %d", pending[0].MPPaymentID),
				fmt.Sprintf("pix_devconfirm_%d", pending[0].MPPaymentID),
			)},
		}, rows...)
	}
	kb := tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}
	editPhoto(chatID, msgID, "shop", caption, &kb)
}

// =============================================
// DEV: confirmar manualmente (apenas quando PIX_MANUAL_CONFIRM=true)
// =============================================

func handlePixDevConfirm(chatID int64, msgID int, mpIDStr string) {
	if os.Getenv("PIX_MANUAL_CONFIRM") != "true" {
		return
	}
	var mpID int64
	fmt.Sscanf(mpIDStr, "%d", &mpID)
	if mpID == 0 {
		return
	}
	charID, diamonds, err := database.ConfirmPixPaymentByMPID(mpID)
	if err != nil || diamonds == 0 {
		editPhoto(chatID, msgID, "shop", "\u274c Pagamento j\u00e1 processado ou n\u00e3o encontrado.", &backKeyboard)
		return
	}
	database.LogDiamond(charID, diamonds, "pix_dev_confirm")
	char, _ := database.GetCharacterByID(charID)
	total := 0
	if char != nil {
		total = char.Diamonds
	}
	editPhoto(chatID, msgID, "shop",
		fmt.Sprintf("\u2705 *[DEV] Pagamento Confirmado!*\n\n\U0001f48e *+%d diamantes* creditados!\n\nTotal: *%d* \U0001f48e", diamonds, total),
		&backKeyboard)
}

// =============================================
// WEBHOOK DO MERCADO PAGO (chamado por HandleMPWebhook em main.go)
// =============================================

// HandleAbacateWebhookNotification processa webhook do AbacatePay (pix.qrcode.paid).
func HandleAbacateWebhookNotification(abacateID string) (playerID int64, diamonds int, err error) {
	charID, diamonds, err := database.ConfirmPixPaymentByTxID(abacateID)
	if err != nil || diamonds == 0 {
		return 0, diamonds, err
	}
	database.LogDiamond(charID, diamonds, "pix_abacatepay_webhook")

	// Notify GM about the payment
	go func() {
		if rec, err := database.GetPixPayment(abacateID); err == nil && rec != nil {
			NotifyGMPixPaid(charID, diamonds, rec.PackageID, rec.AmountBRL)
		}
	}()

	var pid int64
	database.DB.QueryRow(`
		SELECT p.id FROM players p
		JOIN characters c ON c.player_id = p.id
		WHERE c.id=$1
	`, charID).Scan(&pid)

	return pid, diamonds, nil
}

// HandleMPWebhookNotification mantido por compatibilidade.
func HandleMPWebhookNotification(mpPaymentID int64) (playerID int64, diamonds int, err error) {
	return 0, 0, fmt.Errorf("use HandleAbacateWebhookNotification")
}

// =============================================
// POLLING GOROUTINE (chamado de main.go)
// =============================================

// StartPixPolling starts a background goroutine that polls Mercado Pago every
// 15 seconds for pending payments and automatically credits diamonds.
func StartPixPolling() {
	go func() {
		log.Println("[PIX] Polling goroutine started (interval: 15s)")
		for {
			time.Sleep(15 * time.Second)
			pollPendingPixPayments()
		}
	}()
}

func pollPendingPixPayments() {
	pending, err := database.GetAllPendingPixPayments()
	if err != nil || len(pending) == 0 {
		return
	}
	for _, p := range pending {
		if p.TxID == "" {
			continue
		}
		status, err := game.CheckAbacatePayStatus(p.TxID)
		if err != nil {
			continue
		}
		if status != "approved" {
			continue
		}
		charID, diamonds, err := database.ConfirmPixPaymentByTxID(p.TxID)
		if err != nil || diamonds == 0 {
			continue
		}
		database.LogDiamond(charID, diamonds, "pix_abacatepay_poll")
		log.Printf("[PIX] Poll confirmed %d diamonds for charID %d (txID=%s)", diamonds, charID, p.TxID)

		// Notify player via Telegram
		notifyPixConfirmed(charID, diamonds, p.PackageID)
		// Notify GM
		NotifyGMPixPaid(charID, diamonds, p.PackageID, p.AmountBRL)
	}
}

func notifyPixConfirmed(charID, diamonds int, packageID string) {
	var playerID int64
	database.DB.QueryRow(`
		SELECT p.id FROM players p
		JOIN characters c ON c.player_id = p.id
		WHERE c.id=$1
	`, charID).Scan(&playerID)
	if playerID == 0 {
		return
	}

	pkg := game.GetDiamondPackage(packageID)
	pkgName := "Diamantes"
	if pkg != nil {
		pkgName = pkg.Name
	}

	msg := tgbotapi.NewMessage(playerID,
		fmt.Sprintf("\u2705 *Pix Confirmado!*\n\n"+
			"\U0001f48e *+%d diamantes* do pacote *%s* foram creditados na sua conta!\n\n"+
			"_Use /menu para acessar o jogo._",
			diamonds, pkgName))
	msg.ParseMode = "Markdown"
	Bot.Send(msg)
}
