package handlers

import (
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/tormenta-bot/internal/assets"
	"github.com/tormenta-bot/internal/database"
	"github.com/tormenta-bot/internal/drops"
	"github.com/tormenta-bot/internal/game"
	menukit "github.com/tormenta-bot/internal/menu"
	"github.com/tormenta-bot/internal/models"
)

// =============================================
// DUNGEON HANDLERS
// =============================================

func showDungeonMenu(chatID int64, msgID int, userID int64) {
	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}
	game.TickEnergy(char)
	database.SaveCharacter(char)

	dungeons := game.GetAvailableDungeons(char.Level)
	eBar := game.EnergyBar(char.Energy, char.EnergyMax)

	caption := fmt.Sprintf(
		"🏚️ *Masmorras*\n\n%s ⚡ *%d*/%d Energia\n\n"+
			"_Masmorras são desafios em andares sequenciais. Complete todos os andares para a recompensa máxima!_\n\n"+
			"⚡ Custo por andar: *%d* energia\n\n",
		eBar, char.Energy, char.EnergyMax, game.EnergyDungeonEnter,
	)

	var entryRows [][]tgbotapi.InlineKeyboardButton
	activeLabel := ""

	// Show active run first
	activeRun, _ := database.GetActiveDungeonRun(char.ID)
	if activeRun != nil {
		d := game.Dungeons[activeRun.DungeonID]
		caption += fmt.Sprintf("⚔️ *Masmorra Ativa:* %s %s — Andar *%d*/%d\n\n", d.Emoji, d.Name, activeRun.Floor, d.Floors)
		activeLabel = fmt.Sprintf("▶️ Continuar %s %s (Andar %d)", d.Emoji, d.Name, activeRun.Floor)
	} else {
		for _, d := range dungeons {
			bestFloor, completions := database.GetDungeonBest(char.ID, d.ID)
			diffEmoji := game.DifficultyEmoji(d.Difficulty)
			completionStr := ""
			if completions > 0 {
				completionStr = fmt.Sprintf(" ✅×%d", completions)
			}
			bestStr := ""
			if bestFloor > 0 {
				bestStr = fmt.Sprintf(" (melhor: andar %d)", bestFloor)
			}
			locked := ""
			if char.Energy < game.EnergyDungeonEnter {
				locked = " ⚡insuf."
			}

			caption += fmt.Sprintf(
				"%s %s *%s*%s\n%s_%s_\nNv.%d-%d | %d andares | 🪙+%d | 💎+%d%s\n\n",
				diffEmoji, d.Emoji, d.Name, completionStr,
				bestStr+"\n", d.Description,
				d.MinLevel, d.MaxLevel, d.Floors, d.RewardGold, d.RewardDiamonds, locked,
			)

			canEnter := char.Energy >= game.EnergyDungeonEnter
			btnLabel := fmt.Sprintf("%s %s Entrar (-%d⚡)", d.Emoji, d.Name, game.EnergyDungeonEnter)
			if !canEnter {
				btnLabel = fmt.Sprintf("🔒 %s %s (sem energia)", d.Emoji, d.Name)
			}
			entryRows = append(entryRows, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(btnLabel, "dungeon_enter_"+d.ID),
			))
		}
	}

	kb := menukit.DungeonMenu(menukit.DungeonMenuOptions{
		HasActive:      activeRun != nil,
		ActiveContinue: activeLabel,
		EntryRows:      entryRows,
	})
	editPhoto(chatID, msgID, "travel", caption, &kb)
}

func handleDungeonEnter(chatID int64, msgID int, userID int64, dungeonID string) {
	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}
	game.TickEnergy(char)

	d, ok := game.Dungeons[dungeonID]
	if !ok {
		return
	}

	// Check active run
	existing, _ := database.GetActiveDungeonRun(char.ID)
	if existing != nil {
		editPhoto(chatID, msgID, "travel", "❌ Você já tem uma masmorra ativa! Continue ou abandone primeiro.", bkp("menu_dungeon"))
		return
	}

	if char.Level < d.MinLevel {
		editPhoto(chatID, msgID, "travel",
			fmt.Sprintf("❌ *%s* requer nível mínimo *%d*!\nSeu nível: *%d*", d.Name, d.MinLevel, char.Level),
			bkp("menu_dungeon"))
		return
	}

	if !game.ConsumeDungeonEnergy(char) {
		editPhoto(chatID, msgID, "travel",
			fmt.Sprintf("❌ Energia insuficiente!\nPrecisa: *%d* ⚡ | Você tem: *%d* ⚡\n\nAguarde a recarga ou use a loja de diamantes!",
				game.EnergyDungeonEnter, char.Energy),
			bkp("menu_dungeon"))
		return
	}

	run, err := database.CreateDungeonRun(char.ID, dungeonID)
	if err != nil {
		return
	}
	database.SaveCharacter(char)

	renderDungeonFloor(chatID, msgID, char, run, &d, "⚔️ *Você entrou na masmorra!*\n")
}

func handleDungeonContinue(chatID int64, msgID int, userID int64) {
	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}
	run, _ := database.GetActiveDungeonRun(char.ID)
	if run == nil {
		showDungeonMenu(chatID, msgID, userID)
		return
	}
	d := game.Dungeons[run.DungeonID]
	renderDungeonFloor(chatID, msgID, char, run, &d, "")
}

func handleDungeonFight(chatID int64, msgID int, userID int64) {
	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}
	game.TickEnergy(char)

	run, _ := database.GetActiveDungeonRun(char.ID)
	if run == nil {
		showDungeonMenu(chatID, msgID, userID)
		return
	}
	d := game.Dungeons[run.DungeonID]

	// Energy cost to enter the floor combat
	if !game.ConsumeAttackEnergy(char) {
		renderDungeonFloor(chatID, msgID, char, run, &d,
			fmt.Sprintf("❌ *Sem energia para lutar!*\nPrecisa *%d*⚡\n\nUse um item de energia ou aguarde a recarga.\n", game.EnergyPerAttack))
		return
	}

	monster := game.RollDungeonMonster(run.DungeonID, run.Floor)
	if monster == nil {
		return
	}

	// Start turn-based combat — same engine as Explore
	char.State = "dungeon_combat"
	char.CombatMonsterID = monster.ID
	char.CombatMonsterHP = monster.HP
	database.SaveCharacter(char)

	renderDungeonCombat(chatID, msgID, char, monster, run, &d,
		fmt.Sprintf("⚔️ *Andar %d — %s %s apareceu!*\n\n", run.Floor, monster.Emoji, monster.Name))
}

func handleDungeonAbandon(chatID int64, msgID int, userID int64) {
	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}
	run, _ := database.GetActiveDungeonRun(char.ID)
	if run == nil {
		showDungeonMenu(chatID, msgID, userID)
		return
	}
	d := game.Dungeons[run.DungeonID]

	// Partial rewards
	rewardGold, rewardDiamonds, _ := game.DungeonCompleteRewards(run.DungeonID, run.Floor-1, char)
	char.Gold += rewardGold
	if rewardDiamonds > 0 {
		char.Diamonds += rewardDiamonds
		database.LogDiamond(char.ID, rewardDiamonds, "dungeon_abandon_partial")
	}
	char.CombatMonsterPoisonTurns = 0
	char.CombatMonsterPoisonDmg = 0
	char.PoisonTurns = 0
	char.PoisonDmg = 0
	resetPVEEffects(char.ID)
	database.FinishDungeonRun(run.ID, "abandoned")
	database.UpdateDungeonBest(char.ID, run.DungeonID, run.Floor-1, false)
	database.SaveCharacter(char)

	caption := fmt.Sprintf(
		"🚪 *Masmorra abandonada*\n\n%s %s — Andar %d/%d\n\n🎁 Recompensa parcial: +%d 🪙",
		d.Emoji, d.Name, run.Floor-1, d.Floors, rewardGold,
	)
	if rewardDiamonds > 0 {
		caption += fmt.Sprintf(" | +%d 💎", rewardDiamonds)
	}
	editPhoto(chatID, msgID, "travel", caption, bkp("menu_dungeon"))
}

func renderDungeonFloor(chatID int64, msgID int, char *models.Character, run *database.DungeonRun, d *game.Dungeon, log string) {
	floorData := game.GetDungeonFloor(d.ID, run.Floor)
	if floorData == nil {
		return
	}

	hpPct := int(float64(char.HP) / float64(char.HPMax) * 8)
	if hpPct < 0 {
		hpPct = 0
	}
	hpBar := strings.Repeat("❤️", hpPct) + strings.Repeat("🖤", 8-hpPct)
	eBar := game.EnergyBar(char.Energy, char.EnergyMax)
	diffEmoji := game.DifficultyEmoji(d.Difficulty)

	bossTag := ""
	if floorData.IsBoss {
		bossTag = " 🔥 *BOSS*"
	}

	caption := fmt.Sprintf(
		"%s %s *%s* — Andar *%d*/%d%s\n\n"+
			"📍 _%s_\n\n"+
			"%s %d/%d HP | 💙 %d/%d MP\n"+
			"⚡ %s *%d*/%d Energia\n\n"+
			"%s",
		diffEmoji, d.Emoji, d.Name, run.Floor, d.Floors, bossTag,
		floorData.Description,
		hpBar, char.HP, char.HPMax, char.MP, char.MPMax,
		eBar, char.Energy, char.EnergyMax,
		log,
	)

	canFight := char.Energy >= game.EnergyPerAttack
	fightLabel := fmt.Sprintf("⚔️ Entrar em Combate (-%d⚡)", game.EnergyPerAttack)
	if !canFight {
		fightLabel = "⚡ Sem energia para lutar"
	}

	rows := [][]tgbotapi.InlineKeyboardButton{
		{tgbotapi.NewInlineKeyboardButtonData(fightLabel, "dungeon_fight")},
		{
			tgbotapi.NewInlineKeyboardButtonData("🎒 Itens", "dungeon_floor_item"),
			tgbotapi.NewInlineKeyboardButtonData("🚪 Abandonar", "dungeon_abandon"),
		},
		{tgbotapi.NewInlineKeyboardButtonData("🏰 Menu", "menu_main")},
	}
	kb := tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}
	editPhoto(chatID, msgID, assets.MapImageKey("crystal_cave"), caption, &kb)
}

// renderDungeonCombat shows the turn-based combat screen inside a dungeon floor.
// Identical layout to renderCombat but buttons return to dungeon flow on victory.
func renderDungeonCombat(chatID int64, msgID int, char *models.Character, monster *models.Monster, run *database.DungeonRun, d *game.Dungeon, combatLog string) {

	// Proteção contra divisão por zero
	maxPlayerHP := char.HPMax
	if maxPlayerHP <= 0 {
		maxPlayerHP = 1
	}

	maxMonsterHP := monster.HP
	if maxMonsterHP <= 0 {
		maxMonsterHP = 1
	}

	// Calcula proporção da barra (0 a 8)
	pHP := int(float64(char.HP) / float64(maxPlayerHP) * 8)
	mHP := int(float64(char.CombatMonsterHP) / float64(maxMonsterHP) * 8)

	// Clamp (impede valores inválidos)
	if pHP < 0 {
		pHP = 0
	}
	if pHP > 8 {
		pHP = 8
	}

	if mHP < 0 {
		mHP = 0
	}
	if mHP > 8 {
		mHP = 8
	}

	pBar := strings.Repeat("❤️", pHP) + strings.Repeat("🖤", 8-pHP)
	mBar := strings.Repeat("💚", mHP) + strings.Repeat("🖤", 8-mHP)

	eBar := game.EnergyBar(char.Energy, char.EnergyMax)
	diffEmoji := game.DifficultyEmoji(d.Difficulty)

	bossTag := ""
	floorData := game.GetDungeonFloor(d.ID, run.Floor)
	if floorData != nil && floorData.IsBoss {
		bossTag = " 🔥 BOSS"
	}

	caption := fmt.Sprintf(
		"%s %s *%s* — Andar *%d*/%d%s\n\n"+
			"%s *%s* Nv.*%d*\n%s %d/%d HP\n\n"+
			"%s *%s* Nv.*%d*\n%s %d/%d HP | 💙 %d/%d MP | ⚡ %d\n%s\n\n"+
			"━━━━━━━━━━━━\n%s",
		diffEmoji, d.Emoji, d.Name, run.Floor, d.Floors, bossTag,
		monster.Emoji, monster.Name, monster.Level, mBar, char.CombatMonsterHP, monster.HP,
		game.Races[char.Race].Emoji, char.Name, char.Level, pBar, char.HP, char.HPMax, char.MP, char.MPMax, char.Energy,
		eBar, truncateCombatLog(combatLog, 4),
	)

	// Skill buttons
	learnedSkills, _ := database.GetLearnedSkills(char.ID)
	var skillBtns []tgbotapi.InlineKeyboardButton
	for _, ls := range learnedSkills {
		sk := game.Skills[ls.SkillID]
		if sk.Passive {
			continue
		}
		label := fmt.Sprintf("%s%s", sk.Emoji, sk.Name)
		if sk.MPCost > 0 {
			label += fmt.Sprintf("(%dMP)", sk.MPCost)
		}
		skillBtns = append(skillBtns, tgbotapi.NewInlineKeyboardButtonData(label, "combat_skill_"+sk.ID))
	}

	rows := [][]tgbotapi.InlineKeyboardButton{{
		tgbotapi.NewInlineKeyboardButtonData("⚔️ Atacar", "combat_attack"),
		tgbotapi.NewInlineKeyboardButtonData("🎒 Item", "combat_use_item"),
		tgbotapi.NewInlineKeyboardButtonData("🚪 Abandonar", "dungeon_abandon"),
	}}
	for i := 0; i < len(skillBtns); i += 2 {
		if i+1 < len(skillBtns) {
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(skillBtns[i], skillBtns[i+1]))
		} else {
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(skillBtns[i]))
		}
	}
	kb := tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}
	editPhoto(chatID, msgID, assets.MonsterImageKey(monster.ID), caption, &kb)
}

// handleDungeonMonsterDeath is called when a monster dies during dungeon turn-based combat.
// It applies rewards then either advances the floor or completes the dungeon.
func handleDungeonMonsterDeath(chatID int64, msgID int, char *models.Character, monster *models.Monster, combatLog string) {
	run, _ := database.GetActiveDungeonRun(char.ID)
	if run == nil {
		// Run was lost somehow — clean up and go to menu
		char.State = "idle"
		char.CombatMonsterID = ""
		char.CombatMonsterHP = 0
		char.CombatMonsterPoisonTurns = 0
		char.CombatMonsterPoisonDmg = 0
		database.SaveCharacter(char)
		showDungeonMenu(chatID, msgID, char.PlayerID)
		return
	}
	d := game.Dungeons[run.DungeonID]

	// Apply rewards
	xp := game.CalculateXPGain(char, monster)
	goldGain := monster.GoldReward
	char.Experience += xp
	char.Gold += goldGain
	char.State = "idle"
	char.CombatMonsterID = ""
	char.CombatMonsterHP = 0
	char.CombatMonsterPoisonTurns = 0
	char.CombatMonsterPoisonDmg = 0
	resetPVEEffects(char.ID)

	// Diamond drop
	diamondGained := 0
	if monster.DiamondChance > 0 && game.RollDiamondDrop(monster.DiamondChance) {
		diamondGained = 1
		char.Diamonds++
		database.LogDiamond(char.ID, 1, "dungeon_drop")
	}

	// Item drops
	lootDrops := game.RollDrops(monster, char.Level)
	var dropLines []string
	for _, drop := range lootDrops {
		item, ok := game.Items[drop.ItemID]
		if !ok {
			continue
		}
		if item.Type == "chest" {
			chestGold, chestItemID := game.OpenChest(drop.ItemID, char.Level)
			char.Gold += chestGold
			if chestGold > 0 {
				dropLines = append(dropLines, fmt.Sprintf("📦 %s → *+%d* 🪙", item.Name, chestGold))
			}
			if chestItemID != "" {
				ci, ciOK := game.Items[chestItemID]
				if ciOK {
					database.AddItem(char.ID, chestItemID, ci.Type, 1)
					dropLines = append(dropLines, fmt.Sprintf("  ✨ %s *%s* (%s)", ci.Emoji, ci.Name, ci.Rarity.Name()))
				}
			}
		} else {
			database.AddItem(char.ID, drop.ItemID, item.Type, drop.Quantity)
			qtyStr := ""
			if drop.Quantity > 1 {
				qtyStr = fmt.Sprintf("%dx ", drop.Quantity)
			}
			dropLines = append(dropLines, fmt.Sprintf("🎁 %s%s *%s* (%s)", item.Emoji, qtyStr, item.Name, item.Rarity.Name()))
		}
	}
	dropLines = append(dropLines, applyMaterialDrops(char, monster, drops.ModeDungeon)...)

	// Level up
	lvlUp := game.CheckLevelUp(char)
	if lvlUp != nil {
		game.ApplyLevelUp(char, lvlUp)
		char.EnergyMax = game.MaxEnergy(char.Level)
	}

	// Build result log
	resultLog := combatLog + fmt.Sprintf("\n✅ *%s %s derrotado!*\n+%d XP | +%d 🪙", monster.Emoji, monster.Name, xp, goldGain)
	if diamondGained > 0 {
		resultLog += " | +1 💎"
	}
	if len(dropLines) > 0 {
		shown := dropLines
		if len(shown) > 3 {
			shown = shown[:3]
		}
		for _, l := range shown {
			resultLog += "\n" + l
		}
	}
	if lvlUp != nil {
		resultLog += fmt.Sprintf("\n🎉 *NÍVEL UP! Nv.%d*", lvlUp.NewLevel)
	}

	nextFloor := run.Floor + 1
	isCompleted := nextFloor > d.Floors

	if isCompleted {
		// Dungeon complete!
		rewardGold, rewardDiamonds, rewardItem := game.DungeonCompleteRewards(run.DungeonID, d.Floors, char)
		char.Gold += rewardGold
		char.Diamonds += rewardDiamonds
		if rewardItem != "" {
			item := game.Items[rewardItem]
			database.AddItem(char.ID, rewardItem, item.Type, 1)
		}
		database.LogDiamond(char.ID, rewardDiamonds, "dungeon_complete_"+run.DungeonID)
		database.FinishDungeonRun(run.ID, "completed")
		database.UpdateDungeonBest(char.ID, run.DungeonID, d.Floors, true)
		database.SaveCharacter(char)

		itemStr := ""
		if rewardItem != "" {
			item := game.Items[rewardItem]
			itemStr = fmt.Sprintf("\n%s %s", item.Emoji, item.Name)
		}
		caption := truncateCombatLog(resultLog, 6) + fmt.Sprintf(
			"\n\n🏆 *MASMORRA COMPLETA!*\n%s %s terminada!\n\n🎁 *Recompensas Finais:*\n+%d 🪙 | +%d 💎%s",
			d.Emoji, d.Name, rewardGold, rewardDiamonds, itemStr,
		)
		kb := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🏚️ Masmorras", "menu_dungeon"),
				tgbotapi.NewInlineKeyboardButtonData("🏰 Menu", "menu_main"),
			),
		)
		editPhoto(chatID, msgID, "victory", caption, &kb)
		return
	}

	// Advance floor
	database.AdvanceDungeonFloor(run.ID, nextFloor, 1)
	run.Floor = nextFloor

	// Check energy for next floor
	if !game.ConsumeDungeonEnergy(char) {
		database.SaveCharacter(char)
		caption := truncateCombatLog(resultLog, 5) + fmt.Sprintf(
			"\n\n⚡ *Energia insuficiente para o próximo andar!*\nPrecisa: *%d*⚡ | Você tem: *%d*⚡\n\n_A masmorra fica salva. Recupere energia e volte!_",
			game.EnergyDungeonEnter, char.Energy,
		)
		kb := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("⚡ Energia", "menu_energy"),
				tgbotapi.NewInlineKeyboardButtonData("▶️ Continuar Masmorra", "dungeon_continue"),
			),
		)
		editPhoto(chatID, msgID, "rest", caption, &kb)
		return
	}

	database.SaveCharacter(char)
	renderDungeonFloor(chatID, msgID, char, run, &d,
		truncateCombatLog(resultLog, 5)+fmt.Sprintf("\n\n➡️ *Avançando para o Andar %d!*\n", nextFloor))
}

// handleDungeonFloorItem shows the item menu during a dungeon floor (before combat starts)
func handleDungeonFloorItem(chatID int64, msgID int, userID int64) {
	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}
	run, _ := database.GetActiveDungeonRun(char.ID)
	if run == nil {
		showDungeonMenu(chatID, msgID, userID)
		return
	}
	d := game.Dungeons[run.DungeonID]

	items, _ := database.GetInventory(char.ID)
	caption := fmt.Sprintf("🎒 *Itens — Andar %d/%d*\n\n", run.Floor, d.Floors)
	var rows [][]tgbotapi.InlineKeyboardButton

	count := 0
	for _, inv := range items {
		if inv.ItemType != "consumable" {
			continue
		}
		item, ok := game.Items[inv.ItemID]
		if !ok {
			continue
		}
		effects := []string{}
		if item.HealHP > 0 {
			effects = append(effects, fmt.Sprintf("+%dHP", item.HealHP))
		}
		if item.HealMP > 0 {
			effects = append(effects, fmt.Sprintf("+%dMP", item.HealMP))
		}
		if item.RestoreEnergy > 0 {
			effects = append(effects, fmt.Sprintf("+%d⚡", item.RestoreEnergy))
		}
		caption += fmt.Sprintf("%s *%s* ×%d — %s\n", item.Emoji, item.Name, inv.Quantity, strings.Join(effects, " "))
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("Usar %s %s", item.Emoji, item.Name),
				"dungeon_use_item_"+item.ID,
			),
		))
		count++
	}
	if count == 0 {
		caption += "_Sem consumíveis no inventário._"
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("⬅️ Voltar ao Andar", "dungeon_continue"),
	))
	kb := tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}
	editPhoto(chatID, msgID, assets.MapImageKey("crystal_cave"), caption, &kb)
}

// handleDungeonUseItem uses a consumable item during a dungeon floor (before combat)
func handleDungeonUseItem(chatID int64, msgID int, userID int64, itemID string) {
	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}
	run, _ := database.GetActiveDungeonRun(char.ID)
	if run == nil {
		showDungeonMenu(chatID, msgID, userID)
		return
	}
	d := game.Dungeons[run.DungeonID]

	item, ok := game.Items[itemID]
	if !ok {
		return
	}
	count := database.GetItemCount(char.ID, itemID)
	if count <= 0 {
		renderDungeonFloor(chatID, msgID, char, run, &d, "❌ Item não encontrado!\n")
		return
	}

	effects := ""
	if item.HealHP > 0 {
		old := char.HP
		char.HP += item.HealHP
		if char.HP > char.HPMax {
			char.HP = char.HPMax
		}
		effects += fmt.Sprintf("+%d❤️ ", char.HP-old)
	}
	if item.HealMP > 0 {
		old := char.MP
		char.MP += item.HealMP
		if char.MP > char.MPMax {
			char.MP = char.MPMax
		}
		effects += fmt.Sprintf("+%d💙 ", char.MP-old)
	}
	if item.RestoreEnergy > 0 {
		old := char.Energy
		char.Energy += item.RestoreEnergy
		if char.Energy > char.EnergyMax {
			char.Energy = char.EnergyMax
		}
		effects += fmt.Sprintf("+%d⚡ ", char.Energy-old)
	}

	database.RemoveItem(char.ID, itemID, 1)
	database.SaveCharacter(char)
	renderDungeonFloor(chatID, msgID, char, run, &d,
		fmt.Sprintf("%s *%s* usada! %s\n", item.Emoji, item.Name, effects))
}

// handleDungeonBackToCombat re-renders the active dungeon combat screen (from item menu)
func handleDungeonBackToCombat(chatID int64, msgID int, userID int64) {
	char, _ := database.GetCharacter(userID)
	if char == nil || char.State != "dungeon_combat" {
		return
	}
	run, _ := database.GetActiveDungeonRun(char.ID)
	if run == nil {
		showDungeonMenu(chatID, msgID, userID)
		return
	}
	d := game.Dungeons[run.DungeonID]
	monster := game.Monsters[char.CombatMonsterID]
	renderDungeonCombat(chatID, msgID, char, &monster, run, &d, "")
}
