package menu

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/tormenta-bot/internal/models"
)

func PixShop(packages []models.DiamondPackage, hasPending bool) tgbotapi.InlineKeyboardMarkup {
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, len(packages)+2)
	for _, pkg := range packages {
		rows = append(rows, Row(Btn(
			fmt.Sprintf("%s %s — %d💎 por %s", pkg.Emoji, pkg.Name, pkg.Amount+pkg.Bonus, pkg.Price),
			"pix_buy_"+pkg.ID,
		)))
	}
	if hasPending {
		rows = append(rows, Row(Btn("🔄 Verificar Pagamento", "pix_check")))
	}
	rows = append(rows, Row(Btn("📦 Voltar", "menu_diamonds")))
	return Keyboard(rows...)
}

func PixBuyActions() tgbotapi.InlineKeyboardMarkup {
	return Keyboard(
		Row(Btn("🔄 Verificar Pagamento", "pix_check")),
		Row(Btn("💳 Outro Pacote", "menu_pix"), Btn("🏰 Menu", "menu_main")),
	)
}

func PixNoPending() tgbotapi.InlineKeyboardMarkup {
	return Keyboard(
		Row(Btn("💎 Comprar Diamantes", "menu_pix")),
		Row(Btn("🏰 Menu", "menu_main")),
	)
}

func PixNoPendingPtr() *tgbotapi.InlineKeyboardMarkup {
	kb := PixNoPending()
	return &kb
}

func PixConfirmed() tgbotapi.InlineKeyboardMarkup {
	return Keyboard(Row(Btn("🏰 Menu Principal", "menu_main")))
}

func PixConfirmedPtr() *tgbotapi.InlineKeyboardMarkup {
	kb := PixConfirmed()
	return &kb
}

func PixPendingStatusRows(hasDevButton bool, devMPID int64) [][]tgbotapi.InlineKeyboardButton {
	rows := [][]tgbotapi.InlineKeyboardButton{
		Row(Btn("🔄 Verificar Novamente", "pix_check")),
		Row(Btn("💳 Outro Pacote", "menu_pix"), Btn("🏰 Menu", "menu_main")),
	}
	if hasDevButton && devMPID > 0 {
		rows = append([][]tgbotapi.InlineKeyboardButton{
			Row(Btn(fmt.Sprintf("🧪 [DEV] Confirmar ID %d", devMPID), fmt.Sprintf("pix_devconfirm_%d", devMPID))),
		}, rows...)
	}
	return rows
}
