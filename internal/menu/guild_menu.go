package menu

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

// GuildMainMenu is the guild home screen for a member.
func GuildMainMenu() tgbotapi.InlineKeyboardMarkup {
	return Keyboard(
		Row(Btn("👥 Membros", "guild_members"), Btn("🏦 Banco", "guild_bank")),
		Row(Btn("⚔️ Guerra", "guild_war"), Btn("🌟 Bônus", "guild_buffs")),
		Row(Btn("📋 Info", "guild_info"), Btn("🚪 Sair", "guild_leave")),
		Row(Btn("🏰 Menu", "menu_main")),
	)
}

// GuildNoGuild is shown when the player has no guild.
func GuildNoGuild() tgbotapi.InlineKeyboardMarkup {
	return Keyboard(
		Row(Btn("⚔️ Criar Guilda", "guild_create")),
		Row(Btn("🔍 Buscar Guilda", "guild_search")),
		Row(Btn("🏰 Menu", "menu_main")),
	)
}

// GuildConfirmLeave asks for confirmation before leaving.
func GuildConfirmLeave() tgbotapi.InlineKeyboardMarkup {
	return Keyboard(
		Row(Btn("✅ Sim, sair", "guild_leave_confirm"), Btn("❌ Cancelar", "menu_guild")),
	)
}

// GuildWarMenu shows territory war options.
func GuildWarMenu() tgbotapi.InlineKeyboardMarkup {
	return Keyboard(
		Row(Btn("🌲 Floresta Sombria", "gwar_attack_darkwood")),
		Row(Btn("🏰 Fortaleza de Ferro", "gwar_attack_iron_keep")),
		Row(Btn("🌊 Templo Submerso", "gwar_attack_sunken_temple")),
		Row(Btn("🐉 Pico do Dragão", "gwar_attack_dragons_peak")),
		Row(Btn("⬅️ Voltar", "menu_guild")),
	)
}

// GuildBankMenu shows bank deposit/withdraw options.
func GuildBankMenu(balance int) tgbotapi.InlineKeyboardMarkup {
	return Keyboard(
		Row(Btn("💰 Depositar 100 ouro", "guild_bank_dep_100"), Btn("💰 Depositar 500 ouro", "guild_bank_dep_500")),
		Row(Btn("💰 Depositar 1000 ouro", "guild_bank_dep_1000")),
		Row(Btn("⬅️ Voltar", "menu_guild")),
	)
}
