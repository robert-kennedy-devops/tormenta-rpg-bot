package handlers

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/tormenta-bot/internal/crafting"
	"github.com/tormenta-bot/internal/database"
	"github.com/tormenta-bot/internal/forge"
	"github.com/tormenta-bot/internal/game"
	"github.com/tormenta-bot/internal/items"
)

func showForgeMenu(chatID int64, msgID int, userID int64) {
	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}
	inv, _ := database.GetInventory(char.ID)

	caption := "🔨 *Forja de Equipamentos*\n\nSelecione um equipamento para aprimorar (+1 a +10):"
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, 12)

	for _, it := range inv {
		base, ok := game.Items[it.ItemID]
		if !ok {
			continue
		}
		if base.Type != "weapon" && base.Type != "armor" && base.Type != "accessory" {
			continue
		}

		progress, _ := database.GetBestPlayerItemForTemplate(char.ID, base.ID)
		levelTag := "+0"
		if progress != nil {
			levelTag = fmt.Sprintf("+%d", progress.UpgradeLevel)
		}
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("%s %s (%s)", base.Emoji, base.Name, levelTag),
				"forge_pick_"+base.ID,
			),
		))
	}

	if len(rows) == 0 {
		caption += "\n\n_Nenhum equipamento encontrado no inventário._"
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("🏰 Menu", "menu_main"),
	))
	kb := tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}
	editPhoto(chatID, msgID, "shop", caption, &kb)
}

func showForgeItem(chatID int64, msgID int, userID int64, itemID string) {
	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}
	it, ok := game.Items[itemID]
	if !ok {
		showForgeMenu(chatID, msgID, userID)
		return
	}
	if database.GetItemCount(char.ID, itemID) <= 0 {
		showForgeMenu(chatID, msgID, userID)
		return
	}

	progress, _ := database.GetBestPlayerItemForTemplate(char.ID, itemID)
	if progress == nil {
		_ = database.CreatePlayerItemInstances(char.ID, itemID, 1)
		progress, _ = database.GetBestPlayerItemForTemplate(char.ID, itemID)
	}

	currentLevel := 0
	if progress != nil {
		currentLevel = progress.UpgradeLevel
	}
	canAttempt := forge.CanAttempt(currentLevel)
	target := currentLevel + 1
	chance, _ := forge.SuccessChance(currentLevel)

	materialID, materialQty := forgeCostForTarget(target)
	materialName := materialID
	materialEmoji := "🧱"
	if m, ok := game.Items[materialID]; ok {
		materialName = m.Name
		materialEmoji = m.Emoji
	}
	haveMat := database.GetItemCount(char.ID, materialID)
	hasMat := haveMat >= materialQty

	status := "✅ Pronto"
	if !canAttempt {
		status = "🏁 Item no nível máximo"
	} else if !hasMat {
		status = "❌ Materiais insuficientes"
	}

	caption := fmt.Sprintf(
		"🔨 *Forjar Item*\n\n%s *%s*\nAtual: *+%d*  →  Alvo: *+%d*\nChance de sucesso: *%.0f%%*\n\nCusto:\n%s *%s* x%d (você tem: %d)\n\nStatus: %s\n\n_Até +4 não quebra. A partir de +5 pode quebrar ao falhar._",
		it.Emoji, it.Name, currentLevel, target, chance*100,
		materialEmoji, materialName, materialQty, haveMat,
		status,
	)

	rows := make([][]tgbotapi.InlineKeyboardButton, 0, 3)
	if canAttempt && hasMat {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⚒️ Tentar Forja", "forge_try_"+itemID),
		))
	}
	rows = append(rows,
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("⬅️ Voltar", "menu_forge")),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("🏰 Menu", "menu_main")),
	)
	kb := tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}
	editPhoto(chatID, msgID, "shop", caption, &kb)
}

func handleForgeTry(chatID int64, msgID int, userID int64, itemID string) {
	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}
	progress, _ := database.GetBestPlayerItemForTemplate(char.ID, itemID)
	if progress == nil {
		showForgeItem(chatID, msgID, userID, itemID)
		return
	}

	if !forge.CanAttempt(progress.UpgradeLevel) {
		showForgeItem(chatID, msgID, userID, itemID)
		return
	}

	target := progress.UpgradeLevel + 1
	materialID, materialQty := forgeCostForTarget(target)
	if database.GetItemCount(char.ID, materialID) < materialQty {
		showForgeItem(chatID, msgID, userID, itemID)
		return
	}
	_ = database.RemoveItem(char.ID, materialID, materialQty)

	out, err := forge.Attempt(progress.UpgradeLevel, rand.Float64(), rand.Float64())
	if err != nil {
		editPhoto(chatID, msgID, "shop", "❌ Erro na forja.", bkp("menu_forge"))
		return
	}

	item := game.Items[itemID]
	switch out.Status {
	case forge.OutcomeSuccess:
		_ = database.UpdatePlayerItemForge(progress.InstanceID, out.NewLevel, false)
		editPhoto(chatID, msgID, "shop",
			fmt.Sprintf("✅ Forja bem-sucedida!\n\n%s *%s* agora está em *+%d*.", item.Emoji, item.Name, out.NewLevel),
			bkp("menu_forge"))
	case forge.OutcomeBroken:
		_ = database.UpdatePlayerItemForge(progress.InstanceID, out.NewLevel, true)
		_ = database.RemoveItem(char.ID, itemID, 1)
		editPhoto(chatID, msgID, "shop",
			fmt.Sprintf("💥 Forja falhou e o item quebrou!\n\n%s *%s* foi destruído.", item.Emoji, item.Name),
			bkp("menu_forge"))
	default:
		editPhoto(chatID, msgID, "shop",
			fmt.Sprintf("⚠️ Forja falhou.\n\n%s *%s* permanece em *+%d*.", item.Emoji, item.Name, out.NewLevel),
			bkp("menu_forge"))
	}
}

func forgeCostForTarget(target int) (string, int) {
	switch {
	case target <= 5:
		return items.MaterialForgeStone, 1
	case target <= 8:
		return items.MaterialRefinedStone, 1
	default:
		return items.MaterialArcaneEssence, 1
	}
}

func showCraftingMenu(chatID int64, msgID int, userID int64) {
	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}

	caption := "🧰 *Crafting*\n\nSelecione uma receita:"
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, 10)

	keys := make([]string, 0, len(crafting.DefaultRecipes))
	for k := range crafting.DefaultRecipes {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, id := range keys {
		r := crafting.DefaultRecipes[id]
		can := canCraftRecipe(char.ID, r)
		label := "❌"
		if can {
			label = "✅"
		}
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("%s %s", label, r.Name), "craft_preview_"+r.ID),
		))
	}

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("🏰 Menu", "menu_main"),
	))
	kb := tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}
	editPhoto(chatID, msgID, "shop", caption, &kb)
}

func showCraftRecipe(chatID int64, msgID int, userID int64, recipeID string) {
	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}
	r, ok := crafting.DefaultRecipes[recipeID]
	if !ok {
		showCraftingMenu(chatID, msgID, userID)
		return
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("🧰 *%s*\n\nMateriais:\n", r.Name))
	for matID, need := range r.Materials {
		have := database.GetItemCount(char.ID, matID)
		item, ok := game.Items[matID]
		name, emoji := matID, "🧱"
		if ok {
			name, emoji = item.Name, item.Emoji
		}
		status := "❌"
		if have >= need {
			status = "✅"
		}
		b.WriteString(fmt.Sprintf("%s %s *%s* x%d (você: %d)\n", status, emoji, name, need, have))
	}

	can := canCraftRecipe(char.ID, r)
	result := r.ResultItemID
	if it, ok := game.Items[r.ResultItemID]; ok {
		result = it.Emoji + " " + it.Name
	}
	b.WriteString(fmt.Sprintf("\nResultado: *%s* x%d", result, r.ResultQty))
	if !can {
		b.WriteString("\n\n_Recursos insuficientes para fabricar agora._")
	}

	rows := make([][]tgbotapi.InlineKeyboardButton, 0, 3)
	if can {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🛠️ Fabricar", "craft_make_"+r.ID),
		))
	}
	rows = append(rows,
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("⬅️ Voltar", "menu_crafting")),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("🏰 Menu", "menu_main")),
	)
	kb := tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}
	editPhoto(chatID, msgID, "shop", b.String(), &kb)
}

func handleCraftMake(chatID int64, msgID int, userID int64, recipeID string) {
	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}
	r, ok := crafting.DefaultRecipes[recipeID]
	if !ok {
		showCraftingMenu(chatID, msgID, userID)
		return
	}
	if !canCraftRecipe(char.ID, r) {
		showCraftRecipe(chatID, msgID, userID, recipeID)
		return
	}

	for matID, need := range r.Materials {
		_ = database.RemoveItem(char.ID, matID, need)
	}

	resultType := "material"
	if it, ok := game.Items[r.ResultItemID]; ok {
		resultType = it.Type
	}
	_ = database.AddItem(char.ID, r.ResultItemID, resultType, r.ResultQty)

	resultName := r.ResultItemID
	resultEmoji := "🎁"
	if it, ok := game.Items[r.ResultItemID]; ok {
		resultName = it.Name
		resultEmoji = it.Emoji
	}
	editPhoto(chatID, msgID, "shop",
		fmt.Sprintf("✅ Fabricação concluída!\n\n%s *%s* x%d adicionado ao inventário.", resultEmoji, resultName, r.ResultQty),
		bkp("menu_crafting"))
}

func canCraftRecipe(charID int, recipe crafting.Recipe) bool {
	inv := make(map[string]int, len(recipe.Materials))
	for matID := range recipe.Materials {
		inv[matID] = database.GetItemCount(charID, matID)
	}
	return crafting.CanCraft(inv, recipe)
}
