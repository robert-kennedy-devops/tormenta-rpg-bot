package menu

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

func InventoryFilterRows() [][]tgbotapi.InlineKeyboardButton {
	return [][]tgbotapi.InlineKeyboardButton{
		Row(
			Btn("🧪 Consumíveis", "inv_tab_consumable"),
			Btn("⚔️ Armas", "inv_tab_weapon"),
		),
		Row(
			Btn("🛡️ Armaduras", "inv_tab_armor"),
			Btn("💍 Acessórios", "inv_tab_accessory"),
		),
		Row(
			Btn("🧱 Materiais", "inv_tab_material"),
			Btn("📋 Todos", "inv_tab_all"),
		),
	}
}

func InventoryHomeKeyboard() tgbotapi.InlineKeyboardMarkup {
	rows := InventoryFilterRows()
	rows = append(rows,
		Row(Btn("⚔️ Equipamentos", "menu_equip")),
		Row(Btn("🏰 Menu", "menu_main")),
	)
	return Keyboard(rows...)
}

func InventoryPageFooterRows() [][]tgbotapi.InlineKeyboardButton {
	return [][]tgbotapi.InlineKeyboardButton{
		Row(
			Btn("🧪 Consumíveis", "inv_tab_consumable"),
			Btn("⚔️ Armas", "inv_tab_weapon"),
		),
		Row(
			Btn("🛡️ Armaduras", "inv_tab_armor"),
			Btn("💍 Acessórios", "inv_tab_accessory"),
		),
		Row(
			Btn("🧱 Materiais", "inv_tab_material"),
		),
		Row(
			Btn("📋 Todos", "inv_tab_all"),
			Btn("⬅️ Voltar", "menu_inventory"),
		),
		Row(
			Btn("⚔️ Equipamentos", "menu_equip"),
		),
	}
}
