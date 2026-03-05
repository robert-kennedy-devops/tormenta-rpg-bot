package menu

import "fmt"

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

func StartWelcome() tgbotapi.InlineKeyboardMarkup {
	return Keyboard(Row(Btn("⚔️ Criar Personagem", "create_character")))
}

func RaceSelect() tgbotapi.InlineKeyboardMarkup {
	return Keyboard(
		Row(Btn("👤 Humano", "race_human"), Btn("🧝 Elfo", "race_elf")),
		Row(Btn("⛏️ Anão", "race_dwarf"), Btn("👹 Meio-Orc", "race_halforc")),
	)
}

func ClassSelect() tgbotapi.InlineKeyboardMarkup {
	return Keyboard(
		Row(Btn("⚔️ Guerreiro", "class_warrior"), Btn("🧙 Mago", "class_mage")),
		Row(Btn("🗡️ Ladino", "class_rogue"), Btn("🏹 Arqueiro", "class_archer")),
		Row(Btn("⬅️ Voltar", "create_character")),
	)
}

func StatusMenu() tgbotapi.InlineKeyboardMarkup {
	return Keyboard(
		Row(Btn("⚡ Energia", "menu_energy"), Btn("💎 Diamantes", "menu_diamonds")),
		Row(Btn("🗑️ Apagar Personagem", "delete_character"), Btn("🏰 Menu", "menu_main")),
	)
}

func EnergyMenuRows(hpMissing, mpMissing, hpCost, mpCost, bothCost int, hasEnergy bool) [][]tgbotapi.InlineKeyboardButton {
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, 6)
	if hpMissing > 0 && hasEnergy {
		rows = append(rows, Row(Btn(fmt.Sprintf("❤️ Recuperar HP (%d⚡)", hpCost), "energy_heal_hp")))
	}
	if mpMissing > 0 && hasEnergy {
		rows = append(rows, Row(Btn(fmt.Sprintf("💙 Recuperar MP (%d⚡)", mpCost), "energy_heal_mp")))
	}
	if (hpMissing > 0 || mpMissing > 0) && hasEnergy {
		rows = append(rows, Row(Btn(fmt.Sprintf("💖 Recuperar Tudo (%d⚡)", bothCost), "energy_heal_both")))
	}
	rows = append(rows,
		Row(Btn("💎 Recarregar com Diamantes", "diamond_buy_energy_full")),
		Row(Btn("🏰 Menu", "menu_main")),
	)
	return rows
}

func DiamondMenu() tgbotapi.InlineKeyboardMarkup {
	return Keyboard(
		Row(Btn("🎁 Bônus Diário", "daily_bonus")),
		Row(Btn("🛒 Loja de Diamantes", "diamond_shop")),
		Row(Btn("💳 Comprar via Pix", "menu_pix")),
		Row(Btn("🏰 Menu", "menu_main")),
	)
}

func ExploreMenu(exploreLabel string) tgbotapi.InlineKeyboardMarkup {
	return Keyboard(
		Row(Btn(exploreLabel, "fight_random")),
		Row(Btn("⚡ Energia", "menu_energy"), Btn("🏰 Menu", "menu_main")),
	)
}

func EnergyAndMenu() tgbotapi.InlineKeyboardMarkup {
	return Keyboard(Row(Btn("⚡ Energia", "menu_energy"), Btn("🏰 Menu", "menu_main")))
}

func VictoryExploreAndMenu() tgbotapi.InlineKeyboardMarkup {
	return Keyboard(Row(Btn("⚔️ Explorar Mais", "menu_explore"), Btn("🏰 Menu", "menu_main")))
}

func MenuOnly() tgbotapi.InlineKeyboardMarkup {
	return Keyboard(Row(Btn("🏰 Menu", "menu_main")))
}

func DungeonContinueAndMenu() tgbotapi.InlineKeyboardMarkup {
	return Keyboard(Row(Btn("▶️ Continuar Masmorra", "dungeon_continue"), Btn("🏰 Menu", "menu_main")))
}

func DungeonAndMenu() tgbotapi.InlineKeyboardMarkup {
	return Keyboard(Row(Btn("🏚️ Masmorras", "menu_dungeon"), Btn("🏰 Menu", "menu_main")))
}

func EquipHome() tgbotapi.InlineKeyboardMarkup {
	return Keyboard(
		Row(Btn("⚔️ Arma", "equip_slot_weapon"), Btn("⛑️ Cabeça", "equip_slot_head")),
		Row(Btn("🛡️ Peito", "equip_slot_chest"), Btn("🧤 Mãos", "equip_slot_hands")),
		Row(Btn("🦵 Pernas", "equip_slot_legs"), Btn("👢 Pés", "equip_slot_feet")),
		Row(Btn("🛡️ Escudo", "equip_slot_offhand")),
		Row(Btn("💍 Anel", "equip_slot_accessory1"), Btn("📿 Colar", "equip_slot_accessory2")),
		Row(Btn("🎒 Inventário", "menu_inventory")),
		Row(Btn("🏰 Menu", "menu_main")),
	)
}

func DeleteConfirm() tgbotapi.InlineKeyboardMarkup {
	return Keyboard(Row(Btn("✅ Sim, apagar", "delete_confirm"), Btn("❌ Cancelar", "menu_status")))
}

func DeleteDone() tgbotapi.InlineKeyboardMarkup {
	return Keyboard(Row(Btn("⚔️ Criar Novo", "create_character")))
}
