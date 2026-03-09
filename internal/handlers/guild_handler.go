package handlers

import (
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/tormenta-bot/internal/database"
	"github.com/tormenta-bot/internal/guild"
	menukit "github.com/tormenta-bot/internal/menu"
)

// guildOf returns the guild for the given Telegram userID (used as playerID).
func guildOf(userID int64) (*guild.Guild, *guild.Member) {
	m, err := guild.GlobalService.GetMember(userID)
	if err != nil {
		return nil, nil
	}
	g, err := guild.GlobalService.GetGuild(m.GuildID)
	if err != nil {
		return nil, nil
	}
	return g, m
}

func showGuildMenu(chatID int64, msgID int, userID int64) {
	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}
	g, _ := guildOf(userID)
	if g == nil {
		caption := "⚔️ *Guildas*\n\n_Você não pertence a nenhuma guilda._\n\nUna-se a uma guilda para participar de guerras territoriais, acessar o banco coletivo e obter bônus de grupo!"
		kb := menukit.GuildNoGuild()
		editPhoto(chatID, msgID, "menu", caption, &kb)
		return
	}
	perks := guild.PerkForLevel(g.Level)
	members, _ := guild.GlobalService.ListMembers(g.ID)
	caption := fmt.Sprintf(
		"⚔️ *Guilda: %s* %s\n\n"+
			"🏆 Nível: *%d* | 👥 Membros: *%d*/%d\n"+
			"✨ XP: *%d*/%d\n"+
			"🏦 Banco: *%d* ouro\n"+
			"🌟 Bônus: XP+%d%% | Ouro+%d%%\n\n"+
			"_Escolha uma opção:_",
		g.Name, g.Emoji,
		g.Level, len(members), guild.MaxMembersForLevel(g.Level),
		g.XP, g.XPNext,
		g.BankGold,
		perks.XPBonusPct, perks.GoldBonusPct,
	)
	kb := menukit.GuildMainMenu()
	editPhoto(chatID, msgID, "menu", caption, &kb)
}

func showGuildMembers(chatID int64, msgID int, userID int64) {
	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}
	g, _ := guildOf(userID)
	if g == nil {
		showGuildMenu(chatID, msgID, userID)
		return
	}
	members, _ := guild.GlobalService.ListMembers(g.ID)
	caption := fmt.Sprintf("👥 *Membros de %s* (Nv.%d)\n\n", g.Name, g.Level)
	for _, m := range members {
		rankEmoji := "👤"
		switch m.Rank {
		case guild.RankLeader:
			rankEmoji = "👑"
		case guild.RankOfficer:
			rankEmoji = "⭐"
		}
		caption += fmt.Sprintf("%s ID:%d — _%s_\n", rankEmoji, m.PlayerID, string(m.Rank))
	}
	if len(members) == 0 {
		caption += "_Nenhum membro encontrado._"
	}
	kb := menukit.MenuOnly()
	editMsg(chatID, msgID, caption, &kb)
}

func showGuildBank(chatID int64, msgID int, userID int64) {
	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}
	g, _ := guildOf(userID)
	if g == nil {
		showGuildMenu(chatID, msgID, userID)
		return
	}
	caption := fmt.Sprintf(
		"🏦 *Banco da Guilda %s*\n\n"+
			"Saldo atual: *%d* ouro\n"+
			"Seu ouro: *%d*\n\n"+
			"_Selecione um valor para depositar:_",
		g.Name, g.BankGold, char.Gold,
	)
	kb := menukit.GuildBankMenu(g.BankGold)
	editMsg(chatID, msgID, caption, &kb)
}

func showGuildWar(chatID int64, msgID int, userID int64) {
	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}
	g, _ := guildOf(userID)
	if g == nil {
		showGuildMenu(chatID, msgID, userID)
		return
	}
	territoryName := "Nenhum"
	if g.TerritoryID != "" {
		if t, ok := guild.Territories[g.TerritoryID]; ok {
			territoryName = t.Emoji + " " + t.Name
		}
	}
	caption := fmt.Sprintf(
		"⚔️ *Guerras Territoriais*\n\n"+
			"Guilda: *%s* | Território: *%s*\n\n"+
			"Conquiste territórios para obter bônus de recursos e XP!\n\n"+
			"_Escolha um território para declarar guerra:_",
		g.Name, territoryName,
	)
	kb := menukit.GuildWarMenu()
	editMsg(chatID, msgID, caption, &kb)
}

func showGuildSearch(chatID int64, msgID int, _ int64) {
	caption := "🔍 *Buscar Guilda*\n\n" +
		"As guildas funcionam por convite.\n\n" +
		"Para entrar em uma guilda, peça ao líder ou a um oficial que te convide.\n\n" +
		"_Use ⚔️ Criar Guilda para fundar a sua própria!_"
	kb := menukit.GuildNoGuild()
	editMsg(chatID, msgID, caption, &kb)
}

// handleGuildNameInput processes text input for guild name during creation.
// Returns true if the input was consumed.
func handleGuildNameInput(msg *tgbotapi.Message) bool {
	userID := msg.From.ID
	state := creationState[userID]
	if state == nil || state["awaiting_guild_name"] != "true" {
		return false
	}
	name := strings.TrimSpace(msg.Text)
	if len(name) < 3 || len(name) > 20 {
		sendText(msg.Chat.ID, "❌ Nome deve ter entre 3 e 20 caracteres. Tente novamente:")
		return true
	}
	char, _ := database.GetCharacter(userID)
	if char == nil {
		return true
	}
	const createCost = 1000
	if char.Gold < createCost {
		sendText(msg.Chat.ID, fmt.Sprintf("❌ Você precisa de *%d* ouro para criar uma guilda.\n\nSeu ouro: *%d*", createCost, char.Gold))
		delete(creationState[userID], "awaiting_guild_name")
		return true
	}
	g, err := guild.GlobalService.CreateGuild(userID, char.Name, name, "", "")
	if err != nil {
		sendText(msg.Chat.ID, "❌ Erro ao criar guilda: "+err.Error())
		delete(creationState[userID], "awaiting_guild_name")
		return true
	}
	char.Gold -= createCost
	database.SaveCharacter(char)
	delete(creationState[userID], "awaiting_guild_name")
	kb := menukit.MenuOnly()
	sendMsg(msg.Chat.ID, fmt.Sprintf("⚔️ *Guilda %s criada com sucesso!*\n\nVocê é agora o líder!\n🪙 Ouro restante: *%d*", g.Name, char.Gold), &kb)
	return true
}

func showGuildBuffs(chatID int64, msgID int, userID int64) {
	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}
	g, _ := guildOf(userID)
	if g == nil {
		showGuildMenu(chatID, msgID, userID)
		return
	}
	perks := guild.PerkForLevel(g.Level)
	caption := fmt.Sprintf(
		"🌟 *Bônus da Guilda %s* (Nv.%d)\n\n"+
			"💰 Bônus de ouro: *+%d%%*\n"+
			"✨ Bônus de XP: *+%d%%*\n"+
			"📦 Drop+: *+%d%%*\n"+
			"🏦 Banco máximo: *%d* ouro\n\n"+
			"_Aumente o nível da guilda para desbloquear mais bônus!_",
		g.Name, g.Level,
		perks.GoldBonusPct, perks.XPBonusPct, perks.DropRateBonus, perks.MaxBankGold,
	)
	kb := menukit.MenuOnly()
	editMsg(chatID, msgID, caption, &kb)
}

func showGuildInfo(chatID int64, msgID int, userID int64) {
	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}
	g, _ := guildOf(userID)
	if g == nil {
		showGuildMenu(chatID, msgID, userID)
		return
	}
	members, _ := guild.GlobalService.ListMembers(g.ID)
	caption := fmt.Sprintf(
		"📋 *Informações: %s*\n\n"+
			"🏆 Nível: *%d*\n"+
			"👥 Membros: *%d*/%d\n"+
			"✨ XP: *%d*/%d\n"+
			"🏦 Banco: *%d* ouro",
		g.Name, g.Level, len(members), guild.MaxMembersForLevel(g.Level),
		g.XP, g.XPNext, g.BankGold,
	)
	kb := menukit.MenuOnly()
	editMsg(chatID, msgID, caption, &kb)
}

func handleGuildLeave(chatID int64, msgID int, userID int64) {
	caption := "🚪 *Sair da Guilda*\n\nTem certeza que deseja sair da guilda?\n_Esta ação não pode ser desfeita._"
	kb := menukit.GuildConfirmLeave()
	editMsg(chatID, msgID, caption, &kb)
}

func handleGuildLeaveConfirm(chatID int64, msgID int, userID int64) {
	err := guild.GlobalService.Leave(userID)
	if err != nil {
		editMsg(chatID, msgID, "❌ Erro ao sair da guilda: "+err.Error(), &backKeyboard)
		return
	}
	caption := "✅ Você saiu da guilda com sucesso."
	kb := menukit.MenuOnly()
	editMsg(chatID, msgID, caption, &kb)
}

func handleGuildCreate(chatID int64, msgID int, userID int64) {
	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}
	const createCost = 1000
	if char.Gold < createCost {
		editMsg(chatID, msgID, fmt.Sprintf("❌ Você precisa de *%d* ouro para criar uma guilda.\n\nSeu ouro: *%d*", createCost, char.Gold), &backKeyboard)
		return
	}
	caption := fmt.Sprintf(
		"⚔️ *Criar Guilda*\n\n"+
			"Custo: *%d* ouro\n"+
			"Seu ouro: *%d*\n\n"+
			"✏️ Envie o nome da sua guilda (3-20 caracteres):",
		createCost, char.Gold,
	)
	if creationState[userID] == nil {
		creationState[userID] = map[string]string{}
	}
	creationState[userID]["awaiting_guild_name"] = "true"
	kb := menukit.MenuOnly()
	editMsg(chatID, msgID, caption, &kb)
}

func handleGuildBankDeposit(chatID int64, msgID int, userID int64, amountStr string) {
	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}
	amount, err := strconv.Atoi(amountStr)
	if err != nil || amount <= 0 {
		showGuildBank(chatID, msgID, userID)
		return
	}
	if char.Gold < amount {
		editMsg(chatID, msgID, fmt.Sprintf("❌ Ouro insuficiente.\n\nSeu ouro: *%d* | Depósito: *%d*", char.Gold, amount), &backKeyboard)
		return
	}
	result, err := guild.GlobalService.Deposit(userID, amount)
	if err != nil {
		editMsg(chatID, msgID, "❌ Erro ao depositar: "+err.Error(), &backKeyboard)
		return
	}
	char.Gold -= amount
	database.SaveCharacter(char)
	caption := fmt.Sprintf("✅ *%d* ouro depositado!\nBanco da guilda: *%d* ouro.", result.TxAmount, result.NewBalance)
	kb := menukit.MenuOnly()
	editMsg(chatID, msgID, caption, &kb)
}

func handleGuildWarAttack(chatID int64, msgID int, userID int64, territoryID string) {
	g, _ := guildOf(userID)
	if g == nil {
		showGuildMenu(chatID, msgID, userID)
		return
	}
	t, ok := guild.Territories[territoryID]
	if !ok {
		editMsg(chatID, msgID, "❌ Território inválido.", &backKeyboard)
		return
	}
	war, err := guild.GlobalWarService.DeclareWar(g.ID, territoryID)
	var caption string
	if err == guild.ErrWarAlreadyActive {
		caption = fmt.Sprintf(
			"⚔️ *%s %s*\n\nJá existe uma guerra ativa neste território.\nDono atual: *%s*\n\nBônus: %d ouro/h | XP+%d%%",
			t.Emoji, t.Name, t.OwnerGuildName,
			t.Bonus.GoldIncomePerHour, t.Bonus.XPBonusPct,
		)
	} else if err != nil {
		editMsg(chatID, msgID, "❌ Erro: "+err.Error(), &backKeyboard)
		return
	} else {
		caption = fmt.Sprintf(
			"⚔️ *Guerra Declarada!*\n\nTerritório: *%s %s*\n*%s* vs *%s*\n\nGuerra começa em 30 min | Duração: 1h\nBônus: %d ouro/h | XP+%d%%",
			t.Emoji, t.Name, war.AttackerName, war.DefenderName,
			t.Bonus.GoldIncomePerHour, t.Bonus.XPBonusPct,
		)
	}
	kb := menukit.GuildWarMenu()
	editMsg(chatID, msgID, caption, &kb)
}
