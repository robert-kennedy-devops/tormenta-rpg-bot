package menu

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

// WorldBossMenu is shown when a boss is active.
func WorldBossMenu(bossName string) tgbotapi.InlineKeyboardMarkup {
	return Keyboard(
		Row(Btn("⚔️ Atacar Boss!", "worldboss_attack")),
		Row(Btn("🏆 Placar de Dano", "worldboss_leaderboard")),
		Row(Btn("📊 Status do Boss", "worldboss_status")),
		Row(Btn("🏰 Menu", "menu_main")),
	)
}

// WorldBossNoActive is shown when no boss is currently active.
func WorldBossNoActive() tgbotapi.InlineKeyboardMarkup {
	return Keyboard(
		Row(Btn("🔔 Notificar quando spawnar", "worldboss_notify")),
		Row(Btn("📜 Histórico de Bosses", "worldboss_history")),
		Row(Btn("🏰 Menu", "menu_main")),
	)
}

// RaidMenu shows available raids.
func RaidMenu() tgbotapi.InlineKeyboardMarkup {
	return Keyboard(
		Row(Btn("🏰 Cidadela da Tormenta", "raid_join_tormenta_citadel")),
		Row(Btn("🌋 Covil de Sombras", "raid_join_shadow_lair")),
		Row(Btn("🐉 Fortaleza do Dragão", "raid_join_dragon_fortress")),
		Row(Btn("🏰 Menu", "menu_main")),
	)
}
