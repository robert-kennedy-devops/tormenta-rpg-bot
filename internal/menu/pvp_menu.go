package menu

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

type PVPMenuOptions struct {
	HasPending     bool
	PendingMatchID int
	HasActive      bool
}

func PVPMenu(opts PVPMenuOptions) tgbotapi.InlineKeyboardMarkup {
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, 6)

	if opts.HasPending && opts.PendingMatchID > 0 {
		rows = append(rows, Row(
			Btn("✅ Aceitar Desafio", "pvp_accept_"+itoa(opts.PendingMatchID)),
			Btn("❌ Recusar", "pvp_decline_"+itoa(opts.PendingMatchID)),
		))
	}
	if opts.HasActive {
		rows = append(rows, Row(
			Btn("⚔️ Continuar Combate PVP", "pvp_continue"),
		))
	}

	rows = append(rows,
		Row(Btn("⚔️ Desafiar Jogador", "pvp_player_list")),
		Row(Btn("🏆 Ranking PVP", "menu_rank")),
		Row(Btn("🏰 Menu", "menu_main")),
	)
	return Keyboard(rows...)
}
