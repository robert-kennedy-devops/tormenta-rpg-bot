package menu

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

type VIPPanelOptions struct {
	IsVIP      bool
	HasSession bool
	HuntRows   [][]tgbotapi.InlineKeyboardButton
}

func VIPPanel(opts VIPPanelOptions) tgbotapi.InlineKeyboardMarkup {
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, 8)

	if opts.IsVIP {
		if opts.HasSession {
			rows = append(rows, Row(
				Btn("📊 Ver Relatório", "vip_hunt_report"),
				Btn("⏹️ Parar Caça", "vip_hunt_stop"),
			))
		} else {
			rows = append(rows, opts.HuntRows...)
		}
	} else {
		rows = append(rows, Row(Btn("💎 Comprar VIP", "vip_buy")))
	}

	rows = append(rows, Row(Btn("🏰 Menu", "menu_main")))
	return Keyboard(rows...)
}
