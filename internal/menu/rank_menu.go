package menu

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

func RankMenu() tgbotapi.InlineKeyboardMarkup {
	return Keyboard(
		Row(Btn("👤 Minhas Estatísticas", "rank_personal")),
		Row(Btn("🏰 Menu", "menu_main")),
	)
}
