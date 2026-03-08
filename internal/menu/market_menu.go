package menu

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

// MarketMainMenu is the player marketplace home.
func MarketMainMenu() tgbotapi.InlineKeyboardMarkup {
	return Keyboard(
		Row(Btn("🔍 Navegar Listagens", "market_browse")),
		Row(Btn("📦 Minhas Listagens", "market_mine")),
		Row(Btn("🏷️ Listar Item", "market_list")),
		Row(Btn("🔨 Leilões", "market_auctions")),
		Row(Btn("🏰 Menu", "menu_main")),
	)
}

// MarketBrowseMenu shows item category filters.
func MarketBrowseMenu() tgbotapi.InlineKeyboardMarkup {
	return Keyboard(
		Row(Btn("⚔️ Armas", "market_cat_weapon"), Btn("🛡️ Armaduras", "market_cat_armor")),
		Row(Btn("🧪 Consumíveis", "market_cat_consumable"), Btn("💍 Acessórios", "market_cat_accessory")),
		Row(Btn("📦 Todos", "market_cat_all")),
		Row(Btn("⬅️ Voltar", "menu_market")),
	)
}

// MarketAuctionMenu shows auction options.
func MarketAuctionMenu() tgbotapi.InlineKeyboardMarkup {
	return Keyboard(
		Row(Btn("🔨 Leilões Ativos", "market_auctions_list")),
		Row(Btn("📤 Criar Leilão", "market_auction_create")),
		Row(Btn("⬅️ Voltar", "menu_market")),
	)
}
