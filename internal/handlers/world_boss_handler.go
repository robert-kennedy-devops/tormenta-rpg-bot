package handlers

import (
	"fmt"
	"sort"
	"time"

	"github.com/tormenta-bot/internal/database"
	menukit "github.com/tormenta-bot/internal/menu"
	"github.com/tormenta-bot/internal/world"
)

func showWorldBossMenu(chatID int64, msgID int, userID int64) {
	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}
	active := world.Global.Active()
	if active == nil {
		caption := "🌍 *Boss Mundial*\n\n_Nenhum boss ativo no momento._\n\nOs bosses mundiais spawnam a cada 12 horas e ficam disponíveis por 30 minutos. Todos os jogadores podem participar!"
		kb := menukit.WorldBossNoActive()
		editPhoto(chatID, msgID, "menu", caption, &kb)
		return
	}
	def := active.Def
	hpPct := 0
	if active.MaxHP > 0 {
		hpPct = active.CurrentHP * 100 / active.MaxHP
	}
	timeLeft := time.Until(active.ExpiresAt)
	mins := int(timeLeft.Minutes())
	caption := fmt.Sprintf(
		"🌍 *Boss Mundial: %s %s*\n\n"+
			"❤️ HP: *%d*/%d (%d%%)\n"+
			"⚔️ ATK: *%d* | 🛡️ DEF: *%d* | Lv.*%d*\n"+
			"☄️ Elemento: *%s* | 🎯 Fraqueza: *%s*\n\n"+
			"👥 Participantes: *%d*\n"+
			"⏰ Tempo restante: *%d min*\n\n"+
			"_%s_",
		def.Emoji, def.Name,
		active.CurrentHP, active.MaxHP, hpPct,
		def.BaseAttack, def.BaseDefense, def.Level,
		def.Element, def.Weakness,
		len(active.Participants),
		mins,
		def.Description,
	)
	kb := menukit.WorldBossMenu(def.Name)
	editPhoto(chatID, msgID, "menu", caption, &kb)
}

func handleWorldBossAttack(chatID int64, msgID int, userID int64) {
	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}
	active := world.Global.Active()
	if active == nil {
		editMsg(chatID, msgID, "❌ Nenhum boss ativo no momento.", &backKeyboard)
		return
	}
	// Base damage from character attack
	dmg := char.Attack*2 + char.Level*3
	killed, dmgDealt, err := world.Global.AttackBoss(userID, char.Name, dmg)
	if err != nil {
		editMsg(chatID, msgID, "❌ "+err.Error(), &backKeyboard)
		return
	}
	def := active.Def
	var caption string
	if killed {
		gold, xp, _ := world.Global.Rewards(userID)
		char.Gold += gold
		char.Experience += xp
		database.SaveCharacter(char)
		caption = fmt.Sprintf(
			"💀 *%s %s foi derrotado!*\n\n"+
				"Você causou *%d* de dano!\n\n"+
				"🏆 *Suas recompensas:*\n"+
				"🪙 Ouro: *+%d*\n"+
				"✨ XP: *+%d*",
			def.Emoji, def.Name, dmgDealt, gold, xp,
		)
	} else {
		hpPct := 0
		if active.MaxHP > 0 {
			hpPct = active.CurrentHP * 100 / active.MaxHP
		}
		caption = fmt.Sprintf(
			"⚔️ *Você atacou %s %s!*\n\n"+
				"Dano causado: *%d*\n"+
				"HP do Boss: *%d*/%d (%d%%)\n\n"+
				"_Continue atacando para derrotá-lo!_",
			def.Emoji, def.Name, dmgDealt,
			active.CurrentHP, active.MaxHP, hpPct,
		)
	}
	kb := menukit.WorldBossMenu(def.Name)
	editMsg(chatID, msgID, caption, &kb)
}

func showWorldBossStatus(chatID int64, msgID int, userID int64) {
	active := world.Global.Active()
	if active == nil {
		showWorldBossMenu(chatID, msgID, userID)
		return
	}
	def := active.Def
	hpPct := 0
	if active.MaxHP > 0 {
		hpPct = active.CurrentHP * 100 / active.MaxHP
	}
	timeLeft := time.Until(active.ExpiresAt)
	caption := fmt.Sprintf(
		"📊 *Status: %s %s*\n\n"+
			"❤️ HP: *%d*/%d (%d%%)\n"+
			"👥 Participantes: *%d*\n"+
			"⏰ Expira em: *%.0f min*",
		def.Emoji, def.Name,
		active.CurrentHP, active.MaxHP, hpPct,
		len(active.Participants),
		timeLeft.Minutes(),
	)
	kb := menukit.WorldBossMenu(def.Name)
	editMsg(chatID, msgID, caption, &kb)
}

func showWorldBossLeaderboard(chatID int64, msgID int, userID int64) {
	active := world.Global.Active()
	if active == nil {
		showWorldBossMenu(chatID, msgID, userID)
		return
	}
	// Sort participants by damage
	type entry struct {
		name string
		dmg  int
	}
	entries := make([]entry, 0, len(active.Participants))
	for _, p := range active.Participants {
		entries = append(entries, entry{name: p.PlayerName, dmg: p.Damage})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].dmg > entries[j].dmg })

	caption := fmt.Sprintf("🏆 *Placar — %s %s*\n\n", active.Def.Emoji, active.Def.Name)
	medals := []string{"🥇", "🥈", "🥉"}
	for i, e := range entries {
		medal := "▪️"
		if i < len(medals) {
			medal = medals[i]
		}
		caption += fmt.Sprintf("%s #%d *%s* — %d dano\n", medal, i+1, e.name, e.dmg)
		if i >= 9 {
			break
		}
	}
	if len(entries) == 0 {
		caption += "_Nenhum participante ainda._"
	}
	kb := menukit.WorldBossMenu(active.Def.Name)
	editMsg(chatID, msgID, caption, &kb)
}

func handleRaidJoin(chatID int64, msgID int, userID int64, raidID string) {
	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}
	raidDef, ok := world.Raids[raidID]
	if !ok {
		editMsg(chatID, msgID, "❌ Raid inválido.", &backKeyboard)
		return
	}
	session, err := world.GlobalRaids.CreateSession(raidID, world.RaidNormal, userID, char.Name)
	if err != nil {
		editMsg(chatID, msgID, "❌ Erro ao criar sessão de raid: "+err.Error(), &backKeyboard)
		return
	}
	caption := fmt.Sprintf(
		"⚔️ *%s %s*\n\n_%s_\n\n"+
			"👥 Participantes: *1*/%d (mín. %d)\n"+
			"Sessão ID: *%d*\n\n"+
			"_Aguardando mais jogadores para começar..._",
		raidDef.Emoji, raidDef.Name, raidDef.Description,
		raidDef.MaxPlayers, raidDef.MinPlayers,
		session.ID,
	)
	kb := menukit.MenuOnly()
	editMsg(chatID, msgID, caption, &kb)
}
