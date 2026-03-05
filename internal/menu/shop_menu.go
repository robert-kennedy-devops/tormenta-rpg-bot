package menu

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

func ShopCategoryRows() [][]tgbotapi.InlineKeyboardButton {
	return [][]tgbotapi.InlineKeyboardButton{
		Row(
			Btn("🧪 Consumíveis", "shop_tab_consumable"),
			Btn("⚔️ Armas", "shop_tab_weapon"),
		),
		Row(
			Btn("🛡️ Armaduras", "shop_tab_armor"),
			Btn("💍 Acessórios", "shop_tab_accessory"),
		),
	}
}

func ShopCheckoutRow(itemCount int) []tgbotapi.InlineKeyboardButton {
	return Row(
		Btn("✅ Ver Carrinho ("+itoa(itemCount)+" item(ns))", "shop_checkout"),
	)
}

func ShopBackRow() []tgbotapi.InlineKeyboardButton {
	return Row(Btn("⬅️ Voltar", "menu_shop"))
}

func ShopHomeKeyboard(itemCount int, backButton tgbotapi.InlineKeyboardButton) tgbotapi.InlineKeyboardMarkup {
	rows := ShopCategoryRows()
	if itemCount > 0 {
		rows = append(rows, ShopCheckoutRow(itemCount))
	}
	rows = append(rows, Row(backButton))
	return Keyboard(rows...)
}

func itoa(v int) string {
	if v == 0 {
		return "0"
	}
	sign := ""
	if v < 0 {
		sign = "-"
		v = -v
	}
	buf := [20]byte{}
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = byte('0' + (v % 10))
		v /= 10
	}
	return sign + string(buf[i:])
}
