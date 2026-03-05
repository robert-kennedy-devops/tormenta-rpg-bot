package menu

import "fmt"

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

func GMDashboard() tgbotapi.InlineKeyboardMarkup {
	return Keyboard(
		Row(Btn("👥 Jogadores", "gm_players"), Btn("🚫 Banidos", "gm_banned")),
		Row(Btn("🏆 Ranking", "gm_rank"), Btn("🔄 Atualizar", "gm_menu")),
	)
}

func GMPlayerPanel(charID int, banLabel, banAction string) tgbotapi.InlineKeyboardMarkup {
	return Keyboard(
		Row(Btn("💎 Diamantes", fmt.Sprintf("gm_dpanel_%d", charID)), Btn("🪙 Ouro", fmt.Sprintf("gm_gpanel_%d", charID))),
		Row(Btn(banLabel, banAction), Btn("📋 Info completo", fmt.Sprintf("gm_panel_%d", charID))),
		Row(Btn("🔙 Painel GM", "gm_menu")),
	)
}

func GMDiamondPanel(charID int) tgbotapi.InlineKeyboardMarkup {
	return Keyboard(
		Row(
			Btn("+1", fmt.Sprintf("gm_dplus_%d_1", charID)),
			Btn("+5", fmt.Sprintf("gm_dplus_%d_5", charID)),
			Btn("+10", fmt.Sprintf("gm_dplus_%d_10", charID)),
			Btn("+50", fmt.Sprintf("gm_dplus_%d_50", charID)),
		),
		Row(
			Btn("+100", fmt.Sprintf("gm_dplus_%d_100", charID)),
			Btn("+200", fmt.Sprintf("gm_dplus_%d_200", charID)),
			Btn("+500", fmt.Sprintf("gm_dplus_%d_500", charID)),
			Btn("+1000", fmt.Sprintf("gm_dplus_%d_1000", charID)),
		),
		Row(
			Btn("-1", fmt.Sprintf("gm_dminus_%d_1", charID)),
			Btn("-5", fmt.Sprintf("gm_dminus_%d_5", charID)),
			Btn("-10", fmt.Sprintf("gm_dminus_%d_10", charID)),
			Btn("-50", fmt.Sprintf("gm_dminus_%d_50", charID)),
		),
		Row(
			Btn("❌ Zerar tudo", fmt.Sprintf("gm_dminus_%d_99999", charID)),
			Btn("🔙 Voltar", fmt.Sprintf("gm_panel_%d", charID)),
		),
	)
}

func GMGoldPanel(charID int) tgbotapi.InlineKeyboardMarkup {
	return Keyboard(
		Row(
			Btn("+100", fmt.Sprintf("gm_gplus_%d_100", charID)),
			Btn("+500", fmt.Sprintf("gm_gplus_%d_500", charID)),
			Btn("+1000", fmt.Sprintf("gm_gplus_%d_1000", charID)),
			Btn("+5000", fmt.Sprintf("gm_gplus_%d_5000", charID)),
		),
		Row(
			Btn("-100", fmt.Sprintf("gm_gminus_%d_100", charID)),
			Btn("-500", fmt.Sprintf("gm_gminus_%d_500", charID)),
			Btn("-1000", fmt.Sprintf("gm_gminus_%d_1000", charID)),
			Btn("-5000", fmt.Sprintf("gm_gminus_%d_5000", charID)),
		),
		Row(
			Btn("❌ Zerar", fmt.Sprintf("gm_gminus_%d_9999999", charID)),
			Btn("🔙 Voltar", fmt.Sprintf("gm_panel_%d", charID)),
		),
	)
}

func GMConfirm(yesAction, yesLabel, noAction, noLabel string) tgbotapi.InlineKeyboardMarkup {
	return Keyboard(Row(Btn(yesLabel, yesAction), Btn(noLabel, noAction)))
}

func GMPlayerResult(charID int) tgbotapi.InlineKeyboardMarkup {
	return Keyboard(
		Row(Btn("📋 Ver jogador", fmt.Sprintf("gm_panel_%d", charID))),
		Row(Btn("🔙 Painel GM", "gm_menu")),
	)
}

func GMBackToDashboard() tgbotapi.InlineKeyboardMarkup {
	return Keyboard(Row(Btn("🔙 Painel GM", "gm_menu")))
}

func GMPixInline() tgbotapi.InlineKeyboardMarkup {
	return Keyboard(
		Row(Btn("🔄 Atualizar", "gm_pix")),
		Row(Btn("🔙 Painel GM", "gm_menu")),
	)
}
