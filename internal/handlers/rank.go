package handlers

import (
	"fmt"

	"github.com/tormenta-bot/internal/database"
	menukit "github.com/tormenta-bot/internal/menu"
)

func showRankMenu(chatID int64, msgID int, userID int64) {
	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}
	caption := "🏆 *Ranking*\n\n_Escolha uma opção:_"
	kb := menukit.RankMenu()
	editPhoto(chatID, msgID, "menu", caption, &kb)
}

func showPersonalStats(chatID int64, msgID int, userID int64) {
	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}
	caption := fmt.Sprintf(
		"👤 *%s* — Nv. *%d* %s %s\n\n"+
			"❤️ HP: *%d*/%d\n"+
			"💙 MP: *%d*/%d\n"+
			"✨ XP: *%d*/%d\n"+
			"🪙 Ouro: *%d*\n"+
			"💎 Diamantes: *%d*\n\n"+
			"⚔️ ATK: *%d* | 🛡️ DEF: *%d*\n"+
			"🔮 MATK: *%d* | 🔰 MDEF: *%d*\n"+
			"💨 SPD: *%d*",
		char.Name, char.Level, char.Race, char.Class,
		char.HP, char.HPMax,
		char.MP, char.MPMax,
		char.Experience, char.ExperienceNext,
		char.Gold,
		char.Diamonds,
		char.Attack, char.Defense,
		char.MagicAttack, char.MagicDefense,
		char.Speed,
	)
	editMsg(chatID, msgID, caption, &backKeyboard)
}
