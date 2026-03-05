package menu

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

// Btn creates an inline keyboard button.
func Btn(text, callbackData string) tgbotapi.InlineKeyboardButton {
	return tgbotapi.NewInlineKeyboardButtonData(text, callbackData)
}

// Row creates a keyboard row from buttons.
func Row(buttons ...tgbotapi.InlineKeyboardButton) []tgbotapi.InlineKeyboardButton {
	return buttons
}

// Keyboard creates a keyboard markup from rows.
func Keyboard(rows ...[]tgbotapi.InlineKeyboardButton) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}
}
