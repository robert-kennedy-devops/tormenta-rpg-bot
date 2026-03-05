package menu

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

type DungeonMenuOptions struct {
	HasActive      bool
	ActiveContinue string
	EntryRows      [][]tgbotapi.InlineKeyboardButton
}

func DungeonMenu(opts DungeonMenuOptions) tgbotapi.InlineKeyboardMarkup {
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, 32)

	if opts.HasActive {
		if opts.ActiveContinue != "" {
			rows = append(rows, Row(Btn(opts.ActiveContinue, "dungeon_continue")))
		}
		rows = append(rows, Row(Btn("🚪 Abandonar Masmorra", "dungeon_abandon")))
	} else {
		rows = append(rows, opts.EntryRows...)
	}

	rows = append(rows, Row(Btn("🏰 Menu", "menu_main")))
	return Keyboard(rows...)
}
