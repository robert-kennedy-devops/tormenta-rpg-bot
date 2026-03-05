package menu

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

type MainMenuOptions struct {
	InCombat bool
}

// MainMenu builds the main navigation keyboard.
func MainMenu(opts MainMenuOptions) tgbotapi.InlineKeyboardMarkup {
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, 8)

	if opts.InCombat {
		rows = append(rows, Row(
			Btn("⚔️ Voltar ao Combate", "combat_resume"),
		))
	}

	rows = append(rows,
		Row(
			Btn("📊 Status", "menu_status"),
			Btn("🎒 Inventário", "menu_inventory"),
		),
		Row(
			Btn("🌟 Habilidades", "menu_skills"),
			Btn("⚡ Energia", "menu_energy"),
		),
		Row(
			Btn("⚔️ Explorar", "menu_explore"),
			Btn("🗺️ Viajar", "menu_travel"),
		),
		Row(
			Btn("🏪 Loja", "menu_shop"),
			Btn("💰 Vender", "menu_sell"),
		),
		Row(
			Btn("💎 Diamantes", "menu_diamonds"),
		),
		Row(
			Btn("🏚️ Masmorras", "menu_dungeon"),
			Btn("⚔️ Arena PVP", "menu_pvp"),
		),
		Row(
			Btn("🏆 Ranking", "menu_rank"),
			Btn("👑 VIP & Caça Auto", "menu_vip"),
		),
	)

	return Keyboard(rows...)
}
