package handlers

import (
	"fmt"
	"log"
	"math"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/tormenta-bot/internal/database"
	"github.com/tormenta-bot/internal/game"
	menukit "github.com/tormenta-bot/internal/menu"
	"github.com/tormenta-bot/internal/models"
)

func tickEnergyForPlayer(userID int64) {
	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}
	game.TickEnergy(char)
	database.SaveCharacter(char)
}

// StartEnergyRegenWorker ticks energy for offline players in background.
func StartEnergyRegenWorker() {
	// Compat: regen agora é totalmente baseada em timestamp (sem worker).
	log.Println("[Energy] timestamp regen mode enabled (worker disabled)")
}

func handleStart(msg *tgbotapi.Message) {
	char, _ := database.GetCharacter(msg.From.ID)
	chatID := msg.Chat.ID
	if char == nil {
		caption := "🏰 *Bem-vindo ao Mundo de Tormenta!*\n\nVocê chega à fronteira de um mundo repleto de magia, perigo e aventura. Terras vastas, masmorras profundas e dragões ancestrais aguardam os corajosos.\n\n⚔️ *RPG baseado em Tormenta 20!*\n\nCrie seu personagem para começar."
		kb := menukit.StartWelcome()
		sendPhoto(chatID, "welcome", caption, &kb)
	} else {
		// Tick energy regen on every entry
		game.TickEnergy(char)
		database.SaveCharacter(char)
		handleMainMenuNew(chatID, msg.From.ID)
	}
}

func handleMainMenuNew(chatID int64, userID int64) {
	char, _ := database.GetCharacter(userID)
	if char == nil {
		sendText(chatID, "❌ Sem personagem! Use /start.")
		return
	}
	game.TickEnergy(char)
	database.SaveCharacter(char)
	caption, kb := buildMainMenuContent(char)
	sendPhoto(chatID, "menu", caption, &kb)
}

func editMainMenuPhoto(chatID int64, msgID int, userID int64) {
	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}
	game.TickEnergy(char)
	database.SaveCharacter(char)
	caption, kb := buildMainMenuContent(char)
	editPhoto(chatID, msgID, "menu", caption, &kb)
}

func buildMainMenuContent(char *models.Character) (string, tgbotapi.InlineKeyboardMarkup) {
	currentMap := game.Maps[char.CurrentMap]
	r := game.Races[char.Race]
	c := game.Classes[char.Class]

	xpPct := 0
	if char.ExperienceNext > 0 {
		ratio := float64(char.Experience) / float64(char.ExperienceNext)
		ratio = math.Max(0, math.Min(1, ratio))
		xpPct = int(math.Round(ratio * 10))
	}
	if xpPct < 0 {
		xpPct = 0
	}
	if xpPct > 10 {
		xpPct = 10
	}
	xpBar := strings.Repeat("▓", xpPct) + strings.Repeat("░", 10-xpPct)

	hpPct := 0
	if char.HPMax > 0 {
		ratio := float64(char.HP) / float64(char.HPMax)
		ratio = math.Max(0, math.Min(1, ratio))
		hpPct = int(math.Round(ratio * 8))
	}
	if hpPct < 0 {
		hpPct = 0
	}
	if hpPct > 8 {
		hpPct = 8
	}
	hpBar := strings.Repeat("❤️", hpPct) + strings.Repeat("🖤", 8-hpPct)
	energyBar := game.EnergyBar(char.Energy, char.EnergyMax)

	nextRegen := ""
	if char.Energy < char.EnergyMax {
		d := game.NextRegenIn(char)
		mins := int(d.Minutes())
		secs := int(d.Seconds()) % 60
		nextRegen = fmt.Sprintf(" _(+1 em %dm%02ds)_", mins, secs)
	}

	caption := fmt.Sprintf(
		"🏰 *Menu Principal*\n\n"+
			"%s *%s* | %s %s | Nv.*%d*\n"+
			"%s %d/%d HP\n"+
			"💙 %d/%d MP\n"+
			"⚡ %s %d/%d Energia%s\n"+
			"💎 *%d* diamantes | 🪙 *%d* ouro\n"+
			"✨ XP `[%s]` %d/%d\n\n"+
			"📍 *%s %s*",
		r.Emoji, char.Name, c.Emoji, c.Name, char.Level,
		hpBar, char.HP, char.HPMax,
		char.MP, char.MPMax,
		energyBar, char.Energy, char.EnergyMax, nextRegen,
		char.Diamonds, char.Gold,
		xpBar, char.Experience, char.ExperienceNext,
		currentMap.Emoji, currentMap.Name,
	)

	kb := menukit.MainMenu(menukit.MainMenuOptions{
		InCombat: char.State == "combat" || char.State == "dungeon_combat",
	})
	return caption, kb
}
