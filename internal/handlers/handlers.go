package handlers

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/tormenta-bot/internal/assets"
	"github.com/tormenta-bot/internal/database"
	"github.com/tormenta-bot/internal/game"
	"github.com/tormenta-bot/internal/models"
)

var Bot *tgbotapi.BotAPI

// =============================================
// IN-MEMORY STATE (per user)
// =============================================

var (
	creationState = map[int64]map[string]string{}    // char creation flow
	shopQtyState  = map[int64]*models.ShopQtyState{} // pending buy qty selection (legado)
	shopCart      = map[int64]*models.ShopCart{}     // carrinho multi-item
	sellCart      = map[int64]*models.SellCart{}     // seleção multi-venda
	navCurrent    = map[int64]string{}               // tela atual por usuário (navegação)
	navBack       = map[int64]map[string]string{}    // destino de voltar por menu dinâmico
)

func isDynamicBackMenu(dest string) bool {
	switch dest {
	case "menu_inventory", "menu_shop", "menu_skills", "menu_sell":
		return true
	}
	return false
}

func rememberMenuBack(userID int64, dest string) {
	if !isDynamicBackMenu(dest) {
		return
	}
	from := navCurrent[userID]
	if from == "" {
		from = "menu_main"
	}
	if from == dest {
		return
	}
	if _, ok := navBack[userID]; !ok {
		navBack[userID] = map[string]string{}
	}
	navBack[userID][dest] = from
}

func menuBackDest(userID int64, menuDest, fallback string) string {
	if byMenu, ok := navBack[userID]; ok {
		if dest, ok2 := byMenu[menuDest]; ok2 && dest != "" {
			return dest
		}
	}
	return fallback
}

func menuBackButton(userID int64, menuDest, fallback string) tgbotapi.InlineKeyboardButton {
	dest := menuBackDest(userID, menuDest, fallback)
	return tgbotapi.NewInlineKeyboardButtonData("⬅️ Voltar", dest)
}

func screenFromCallback(data string) string {
	switch {
	case data == "menu_main",
		data == "menu_status",
		data == "menu_inventory",
		data == "menu_equip",
		data == "menu_skills",
		data == "menu_shop",
		data == "menu_sell",
		data == "menu_travel",
		data == "menu_explore",
		data == "menu_energy",
		data == "menu_diamonds",
		data == "menu_dungeon",
		data == "menu_vip",
		data == "menu_pvp",
		data == "menu_rank":
		return data
	case strings.HasPrefix(data, "inv_tab_"),
		strings.HasPrefix(data, "inv_item_"),
		strings.HasPrefix(data, "inv_back_"),
		strings.HasPrefix(data, "inv_unequip_"),
		strings.HasPrefix(data, "inv_use_"),
		strings.HasPrefix(data, "inv_equip_"):
		return "menu_inventory"
	case strings.HasPrefix(data, "shop_page_"),
		strings.HasPrefix(data, "shop_add_"),
		strings.HasPrefix(data, "shop_inc_"),
		strings.HasPrefix(data, "shop_dec_"),
		strings.HasPrefix(data, "shop_rem_"),
		data == "shop_checkout",
		data == "shop_confirm_buy",
		data == "shop_cancel_buy",
		strings.HasPrefix(data, "shop_item_"),
		strings.HasPrefix(data, "shop_qty_"):
		return "menu_shop"
	case strings.HasPrefix(data, "skill_branch_"),
		strings.HasPrefix(data, "skill_learn_"):
		return "menu_skills"
	case strings.HasPrefix(data, "sell_"):
		// inclui sell_tab_, sell_item_, sell_back_, sell_add_, checkout...
		return "menu_sell"
	}
	return ""
}

// =============================================
// MESSAGE DISPATCHER
// =============================================

func HandleMessage(msg *tgbotapi.Message) {
	userID := msg.From.ID
	database.UpsertPlayer(userID, msg.From.UserName)

	// Check if banned
	if banned := database.IsPlayerBanned(userID); banned {
		sendText(msg.Chat.ID, "🚫 Sua conta está suspensa. Entre em contato com o suporte.")
		return
	}

	// Aplica regen passiva em qualquer interação de mensagem.
	tickEnergyForPlayer(userID)

	// Normalize: "/gm@BotName painel" → "/gm painel"
	if msg.Text == "" && msg.Command() != "" {
		args := msg.CommandArguments()
		if args != "" {
			msg.Text = "/" + msg.Command() + " " + args
		} else {
			msg.Text = "/" + msg.Command()
		}
	}

	// GM commands take priority
	if HandleGMCommand(msg) {
		return
	}

	switch msg.Text {
	case "/start":
		handleStart(msg)
	case "/menu":
		handleMainMenuNew(msg.Chat.ID, userID)
	default:
		handleTextInput(msg)
	}
}

// =============================================
// CALLBACK DISPATCHER
// =============================================

func HandleCallback(cb *tgbotapi.CallbackQuery) {
	userID := cb.From.ID
	chatID := cb.Message.Chat.ID
	msgID := cb.Message.MessageID
	data := cb.Data
	rememberMenuBack(userID, data)

	Bot.Request(tgbotapi.NewCallback(cb.ID, ""))

	// Check if banned
	if database.IsPlayerBanned(userID) {
		return
	}

	// Aplica regen passiva em qualquer interação de callback.
	tickEnergyForPlayer(userID)

	// GM callbacks
	if HandleGMCallback(cb) {
		return
	}

	if len(cb.Message.Photo) > 0 {
		markAsPhoto(chatID, msgID)
	}

	switch {
	// ── NAVIGATION ──────────────────────────────────────
	case data == "menu_main":
		editMainMenuPhoto(chatID, msgID, userID)
	case data == "combat_resume":
		handleCombatResume(chatID, msgID, userID)
	case data == "menu_status":
		showStatus(chatID, msgID, userID)
	case data == "menu_inventory":
		showInventoryHome(chatID, msgID, userID)
	case data == "menu_equip":
		showEquipScreen(chatID, msgID, userID)
	case strings.HasPrefix(data, "equip_slot_"):
		showSlotItems(chatID, msgID, userID, strings.TrimPrefix(data, "equip_slot_"))
	case strings.HasPrefix(data, "equip_item_"):
		// format: equip_item_<itemID>_<slot>
		rest := strings.TrimPrefix(data, "equip_item_")
		// slot is after last underscore
		parts := strings.Split(rest, "_")
		if len(parts) >= 2 {
			slot := parts[len(parts)-1]
			itemID := strings.Join(parts[:len(parts)-1], "_")
			handleEquipItem(chatID, msgID, userID, itemID, slot)
		}
	case strings.HasPrefix(data, "equip_remove_"):
		handleUnequipSlot(chatID, msgID, userID, strings.TrimPrefix(data, "equip_remove_"))
	case data == "inv_tab_consumable":
		showInventory(chatID, msgID, userID, "consumable")
	case data == "inv_tab_weapon":
		showInventory(chatID, msgID, userID, "weapon")
	case data == "inv_tab_armor":
		showInventory(chatID, msgID, userID, "armor")
	case data == "inv_tab_all":
		showInventory(chatID, msgID, userID, "all")
	case data == "inv_tab_accessory":
		showInventory(chatID, msgID, userID, "accessory")
	case strings.HasPrefix(data, "inv_item_"):
		rest := strings.TrimPrefix(data, "inv_item_")
		parts := strings.SplitN(rest, "_", 2)
		if len(parts) == 2 {
			showInventoryItem(chatID, msgID, userID, parts[0], parts[1])
		}
	case strings.HasPrefix(data, "inv_back_"):
		showInventory(chatID, msgID, userID, strings.TrimPrefix(data, "inv_back_"))
	case strings.HasPrefix(data, "inv_unequip_"):
		rest := strings.TrimPrefix(data, "inv_unequip_")
		parts := strings.SplitN(rest, "_", 2)
		if len(parts) == 2 {
			handleInventoryUnequip(chatID, msgID, userID, parts[0], parts[1])
		}
	case data == "menu_skills":
		showSkillTree(chatID, msgID, userID)
	case data == "menu_shop":
		showShopMenu(chatID, msgID, userID)
	case data == "menu_travel":
		showTravelMenu(chatID, msgID, userID)
	case data == "menu_explore":
		showExploreMenu(chatID, msgID, userID)

	// ── CHARACTER CREATION ───────────────────────────────
	case data == "create_character":
		startCharacterCreation(chatID, msgID, userID)
	case strings.HasPrefix(data, "race_"):
		handleRaceSelect(chatID, msgID, userID, strings.TrimPrefix(data, "race_"))
	case strings.HasPrefix(data, "class_"):
		handleClassSelect(chatID, msgID, userID, strings.TrimPrefix(data, "class_"))
	case data == "delete_character":
		confirmDeleteCharacter(chatID, msgID, userID)
	case data == "delete_confirm":
		handleDeleteCharacter(chatID, msgID, userID)

	// ── ENERGY ──────────────────────────────────────────
	case data == "menu_energy":
		showEnergyMenu(chatID, msgID, userID)
	case data == "energy_heal_hp":
		handleEnergyHealHP(chatID, msgID, userID)
	case data == "energy_heal_mp":
		handleEnergyHealMP(chatID, msgID, userID)
	case data == "energy_heal_both":
		handleEnergyHealBoth(chatID, msgID, userID)

	// ── DIAMONDS ────────────────────────────────────────
	case data == "menu_diamonds":
		showDiamondMenu(chatID, msgID, userID)
	case data == "diamond_shop":
		showDiamondShop(chatID, msgID, userID)
	case strings.HasPrefix(data, "diamond_buy_"):
		handleDiamondBuy(chatID, msgID, userID, strings.TrimPrefix(data, "diamond_buy_"))
	case strings.HasPrefix(data, "diamond_item_"):
		handleDiamondItemBuy(chatID, msgID, userID, strings.TrimPrefix(data, "diamond_item_"))
	case data == "diamond_packages":
		showDiamondMenu(chatID, msgID, userID)
	case data == "daily_bonus":
		handleDailyBonus(chatID, msgID, userID)

	// ── SHOP ────────────────────────────────────────────
	case strings.HasPrefix(data, "shop_page_"):
		showShopPage(chatID, msgID, userID, strings.TrimPrefix(data, "shop_page_"))
	case strings.HasPrefix(data, "shop_add_"):
		handleShopAddItem(chatID, msgID, userID, strings.TrimPrefix(data, "shop_add_"))
	case strings.HasPrefix(data, "shop_inc_"):
		handleShopChangeQty(chatID, msgID, userID, strings.TrimPrefix(data, "shop_inc_"), +1)
	case strings.HasPrefix(data, "shop_dec_"):
		handleShopChangeQty(chatID, msgID, userID, strings.TrimPrefix(data, "shop_dec_"), -1)
	case strings.HasPrefix(data, "shop_rem_"):
		handleShopRemoveItem(chatID, msgID, userID, strings.TrimPrefix(data, "shop_rem_"))
	case data == "shop_checkout":
		handleShopCheckout(chatID, msgID, userID)
	case data == "shop_confirm_buy":
		handleShopConfirmBuy(chatID, msgID, userID)
	case data == "shop_cancel_buy":
		showShopMenu(chatID, msgID, userID)
	// legado — mantido por sessões em aberto
	case strings.HasPrefix(data, "shop_item_"):
		handleShopAddItem(chatID, msgID, userID, strings.TrimPrefix(data, "shop_item_"))
	case strings.HasPrefix(data, "shop_qty_"):
		handleShopQty(chatID, msgID, userID, strings.TrimPrefix(data, "shop_qty_"))

	// ── SELL ────────────────────────────────────────────
	case data == "menu_sell":
		showSellHome(chatID, msgID, userID)
	case data == "sell_tab_consumable":
		showSellPage(chatID, msgID, userID, "consumable")
	case data == "sell_tab_weapon":
		showSellPage(chatID, msgID, userID, "weapon")
	case data == "sell_tab_armor":
		showSellPage(chatID, msgID, userID, "armor")
	case data == "sell_tab_accessory":
		showSellPage(chatID, msgID, userID, "accessory")
	case data == "sell_tab_all":
		showSellPage(chatID, msgID, userID, "all")
	case strings.HasPrefix(data, "sell_item_"):
		rest := strings.TrimPrefix(data, "sell_item_")
		parts := strings.SplitN(rest, "_", 2)
		if len(parts) == 2 {
			showSellItem(chatID, msgID, userID, parts[0], parts[1])
		}
	case strings.HasPrefix(data, "sell_back_"):
		showSellPage(chatID, msgID, userID, strings.TrimPrefix(data, "sell_back_"))
	case strings.HasPrefix(data, "sell_add_"):
		rest := strings.TrimPrefix(data, "sell_add_")
		parts := strings.SplitN(rest, "_", 2)
		if len(parts) == 2 {
			char, _ := database.GetCharacter(userID)
			if char != nil {
				_ = addSellItemToCart(userID, char.ID, parts[1])
			}
			showSellItem(chatID, msgID, userID, parts[0], parts[1])
		} else {
			// legado: sell_add_<itemID>
			handleSellAddItem(chatID, msgID, userID, rest)
		}
	case strings.HasPrefix(data, "sell_inc_"):
		handleSellChangeQty(chatID, msgID, userID, strings.TrimPrefix(data, "sell_inc_"), +1)
	case strings.HasPrefix(data, "sell_dec_"):
		handleSellChangeQty(chatID, msgID, userID, strings.TrimPrefix(data, "sell_dec_"), -1)
	case strings.HasPrefix(data, "sell_rem_"):
		handleSellRemoveItem(chatID, msgID, userID, strings.TrimPrefix(data, "sell_rem_"))
	case data == "sell_checkout":
		handleSellCheckout(chatID, msgID, userID)
	case data == "sell_confirm_all":
		handleSellConfirmAll(chatID, msgID, userID)
	case data == "sell_cancel":
		delete(sellCart, userID)
		showSellMenu(chatID, msgID, userID)

	// ── TRAVEL ──────────────────────────────────────────
	case strings.HasPrefix(data, "travel_"):
		handleTravel(chatID, msgID, userID, strings.TrimPrefix(data, "travel_"))

	// ── EXPLORE / COMBAT ────────────────────────────────
	case strings.HasPrefix(data, "fight_"):
		handleFightStart(chatID, msgID, userID, strings.TrimPrefix(data, "fight_"))
	case data == "combat_attack":
		handleCombatAttack(chatID, msgID, userID)
	case strings.HasPrefix(data, "combat_skill_"):
		handleCombatSkill(chatID, msgID, userID, strings.TrimPrefix(data, "combat_skill_"))
	case data == "combat_flee":
		handleCombatFlee(chatID, msgID, userID)
	case data == "combat_use_item":
		showCombatItemMenu(chatID, msgID, userID)
	case strings.HasPrefix(data, "combat_item_"):
		handleCombatItem(chatID, msgID, userID, strings.TrimPrefix(data, "combat_item_"))

	// ── INVENTORY ───────────────────────────────────────
	case strings.HasPrefix(data, "inv_use_"):
		handleInventoryUse(chatID, msgID, userID, strings.TrimPrefix(data, "inv_use_"))
	case strings.HasPrefix(data, "inv_equip_"):
		handleInventoryEquip(chatID, msgID, userID, strings.TrimPrefix(data, "inv_equip_"))

	// ── SKILLS ──────────────────────────────────────────
	case data == "menu_skills":
		showSkillTree(chatID, msgID, userID)
	case strings.HasPrefix(data, "skill_branch_"):
		showSkillBranch(chatID, msgID, userID, strings.TrimPrefix(data, "skill_branch_"))
	case strings.HasPrefix(data, "skill_learn_"):
		handleLearnSkill(chatID, msgID, userID, strings.TrimPrefix(data, "skill_learn_"))

	// ── DUNGEON ─────────────────────────────────────────
	case data == "menu_vip":
		showVIPPanel(chatID, msgID, userID)
	case data == "vip_buy":
		showVIPBuyOptions(chatID, msgID)
	case data == "vip_buy_30":
		handleVIPPurchase(chatID, msgID, userID, "30")
	case data == "vip_buy_90":
		handleVIPPurchase(chatID, msgID, userID, "90")
	case data == "vip_buy_perm":
		handleVIPPurchase(chatID, msgID, userID, "perm")
	case data == "vip_hunt_stop":
		handleAutoHuntStop(chatID, msgID, userID)
	case data == "vip_hunt_report":
		handleAutoHuntReport(chatID, msgID, userID)

	// ── Callbacks compactos VIP (prefixo vh_) ─────────────────────────
	// vh_cfg_<mcode>  — abre tela de configuração da área
	case strings.HasPrefix(data, "vh_cfg_"):
		mapID := cm(strings.TrimPrefix(data, "vh_cfg_"))
		showAutoHuntConfig(chatID, msgID, userID, mapID)

	// vh_tog_<mcode>_<selectedRaw>  — toggle de habilidade
	case strings.HasPrefix(data, "vh_tog_"):
		rest := strings.TrimPrefix(data, "vh_tog_")
		idx := strings.Index(rest, "_")
		if idx >= 0 {
			mapID := cm(rest[:idx])
			showAutoHuntConfigWithSkills(chatID, msgID, userID, mapID, rest[idx+1:])
		}

	// vh_sm_<mcode>_<mode>  — modo rápido (attack | smart | skill_all)
	case strings.HasPrefix(data, "vh_sm_"):
		handleAutoHuntSetMode(chatID, msgID, userID, strings.TrimPrefix(data, "vh_sm_"))

	// vh_ok_<mcode>_<skills+>  — iniciar com habilidades selecionadas
	case strings.HasPrefix(data, "vh_ok_"):
		handleAutoHuntConfirm(chatID, msgID, userID, strings.TrimPrefix(data, "vh_ok_"))

	// vh_pot_<mcode>_<raw>  — abrir potion picker
	case strings.HasPrefix(data, "vh_pot_"):
		rest := strings.TrimPrefix(data, "vh_pot_")
		idx := strings.Index(rest, "_")
		if idx >= 0 {
			showPotionPicker(chatID, msgID, userID, cm(rest[:idx]), rest[idx+1:])
		}

	// vh_pt_<mcode>_<raw>  — toggle de poção / threshold
	case strings.HasPrefix(data, "vh_pt_"):
		rest := strings.TrimPrefix(data, "vh_pt_")
		idx := strings.Index(rest, "_")
		if idx >= 0 {
			showPotionPicker(chatID, msgID, userID, cm(rest[:idx]), rest[idx+1:])
		}

	// vh_pk_<mcode>_<raw>  — confirmar poções → tela de escolha de modo
	case strings.HasPrefix(data, "vh_pk_"):
		rest := strings.TrimPrefix(data, "vh_pk_")
		idx := strings.Index(rest, "_")
		if idx >= 0 {
			handlePotionsDone(chatID, msgID, userID, cm(rest[:idx]), rest[idx+1:])
		}

	// vh_go_<mcode>_<mode>_<potionRaw>  — iniciar com modo + poções
	case strings.HasPrefix(data, "vh_go_"):
		handleAutoHuntStartFull(chatID, msgID, userID, strings.TrimPrefix(data, "vh_go_"))

	// Separador visual — ignora clique
	case data == "vip_noop":
		// nada

	// ── Rotas legadas (mantidas por compatibilidade) ───────────────────
	case strings.HasPrefix(data, "vip_hunt_config_"):
		handleAutoHuntStart_legacy(chatID, msgID, userID, strings.TrimPrefix(data, "vip_hunt_config_"))
	case strings.HasPrefix(data, "vip_hunt_setmode_"):
		handleAutoHuntSetMode(chatID, msgID, userID, strings.TrimPrefix(data, "vip_hunt_setmode_"))
	case strings.HasPrefix(data, "vip_hunt_pickskills_"):
		rest := strings.TrimPrefix(data, "vip_hunt_pickskills_")
		parts := strings.SplitN(rest, "_", 2)
		mapID := parts[0]
		selectedRaw := ""
		if len(parts) > 1 {
			selectedRaw = parts[1]
		}
		showSkillPicker(chatID, msgID, userID, mapID, selectedRaw)
	case strings.HasPrefix(data, "vip_hunt_confirm_"):
		handleAutoHuntConfirm(chatID, msgID, userID, strings.TrimPrefix(data, "vip_hunt_confirm_"))
	case strings.HasPrefix(data, "vip_hunt_start_"):
		handleAutoHuntStart_legacy(chatID, msgID, userID, strings.TrimPrefix(data, "vip_hunt_start_"))
	case strings.HasPrefix(data, "vip_hunt_startfull_"):
		handleAutoHuntStartFull(chatID, msgID, userID, strings.TrimPrefix(data, "vip_hunt_startfull_"))
	case strings.HasPrefix(data, "vip_hunt_potions_"):
		rest := strings.TrimPrefix(data, "vip_hunt_potions_")
		idx := strings.Index(rest, "_")
		if idx >= 0 {
			showPotionPicker(chatID, msgID, userID, rest[:idx], rest[idx+1:])
		}
	case strings.HasPrefix(data, "vip_hunt_potionsdone_"):
		rest := strings.TrimPrefix(data, "vip_hunt_potionsdone_")
		idx := strings.Index(rest, "_")
		if idx >= 0 {
			handlePotionsDone(chatID, msgID, userID, rest[:idx], rest[idx+1:])
		}
	case strings.HasPrefix(data, "vip_hunt_cfg_toggle_"):
		handleHuntSkillToggle(chatID, msgID, userID, strings.TrimPrefix(data, "vip_hunt_cfg_toggle_"))
	case data == "menu_dungeon":
		showDungeonMenu(chatID, msgID, userID)
	case strings.HasPrefix(data, "dungeon_enter_"):
		handleDungeonEnter(chatID, msgID, userID, strings.TrimPrefix(data, "dungeon_enter_"))
	case data == "dungeon_continue":
		handleDungeonContinue(chatID, msgID, userID)
	case data == "dungeon_fight":
		handleDungeonFight(chatID, msgID, userID)
	case data == "dungeon_abandon":
		handleDungeonAbandon(chatID, msgID, userID)
	case data == "dungeon_floor_item":
		handleDungeonFloorItem(chatID, msgID, userID)
	case strings.HasPrefix(data, "dungeon_use_item_"):
		handleDungeonUseItem(chatID, msgID, userID, strings.TrimPrefix(data, "dungeon_use_item_"))
	case data == "dungeon_back_to_combat":
		handleDungeonBackToCombat(chatID, msgID, userID)

	// ── PVP ─────────────────────────────────────────────
	case data == "menu_pvp":
		showPVPMenu(chatID, msgID, userID)
	case data == "pvp_player_list":
		showPVPPlayerList(chatID, msgID, userID)
	case strings.HasPrefix(data, "pvp_select_"):
		if n, err := strconv.Atoi(strings.TrimPrefix(data, "pvp_select_")); err == nil {
			handlePVPSelectPlayer(chatID, msgID, userID, n)
		}
	case strings.HasPrefix(data, "pvp_stake_"):
		if n, err := strconv.Atoi(strings.TrimPrefix(data, "pvp_stake_")); err == nil {
			handlePVPSendChallenge(chatID, msgID, userID, n)
		}
	case strings.HasPrefix(data, "pvp_accept_"):
		if n, err := strconv.Atoi(strings.TrimPrefix(data, "pvp_accept_")); err == nil {
			handlePVPAccept(chatID, msgID, userID, n)
		}
	case strings.HasPrefix(data, "pvp_decline_"):
		if n, err := strconv.Atoi(strings.TrimPrefix(data, "pvp_decline_")); err == nil {
			handlePVPDecline(chatID, msgID, userID, n)
		}
	case data == "pvp_attack":
		handlePVPAttackTurn(chatID, msgID, userID)
	case strings.HasPrefix(data, "pvp_skill_"):
		handlePVPSkillTurn(chatID, msgID, userID, strings.TrimPrefix(data, "pvp_skill_"))
	case data == "pvp_item_menu":
		showPVPItemMenu(chatID, msgID, userID)
	case strings.HasPrefix(data, "pvp_item_"):
		handlePVPUseItem(chatID, msgID, userID, strings.TrimPrefix(data, "pvp_item_"))
	case data == "pvp_continue":
		handlePVPContinue(chatID, msgID, userID)

	// ── RANK ────────────────────────────────────────────
	case data == "menu_rank":
		showRankMenu(chatID, msgID, userID)
	case data == "rank_personal":
		showPersonalStats(chatID, msgID, userID)

	// ── PIX ─────────────────────────────────────────────
	case data == "menu_pix":
		showPixShop(chatID, msgID, userID)
	case strings.HasPrefix(data, "pix_buy_"):
		handlePixBuy(chatID, msgID, userID, strings.TrimPrefix(data, "pix_buy_"))
	case data == "pix_check":
		handlePixCheck(chatID, msgID, userID)
	case strings.HasPrefix(data, "pix_devconfirm_"):
		handlePixDevConfirm(chatID, msgID, strings.TrimPrefix(data, "pix_devconfirm_"))
	}

	if screen := screenFromCallback(data); screen != "" {
		navCurrent[userID] = screen
	}
}

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
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		log.Println("[Energy] Background regen worker started (interval: 1m)")
		for range ticker.C {
			chars, err := database.GetCharactersNeedingEnergyTick(5000)
			if err != nil {
				log.Printf("[Energy] failed to load characters: %v", err)
				continue
			}
			updated := 0
			for i := range chars {
				if game.TickEnergy(&chars[i]) <= 0 {
					continue
				}
				if err := database.SaveCharacterEnergy(chars[i].ID, chars[i].Energy, chars[i].EnergyMax, chars[i].EnergyRegenAt); err != nil {
					log.Printf("[Energy] failed to save char %d: %v", chars[i].ID, err)
					continue
				}
				updated++
			}
			if updated > 0 {
				log.Printf("[Energy] regenerated energy for %d character(s)", updated)
			}
		}
	}()
}

// =============================================
// START
// =============================================

func handleStart(msg *tgbotapi.Message) {
	char, _ := database.GetCharacter(msg.From.ID)
	chatID := msg.Chat.ID
	if char == nil {
		caption := "🏰 *Bem-vindo ao Mundo de Tormenta!*\n\nVocê chega à fronteira de um mundo repleto de magia, perigo e aventura. Terras vastas, masmorras profundas e dragões ancestrais aguardam os corajosos.\n\n⚔️ *RPG baseado em Tormenta 20!*\n\nCrie seu personagem para começar."
		kb := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("⚔️ Criar Personagem", "create_character"),
			),
		)
		sendPhoto(chatID, "welcome", caption, &kb)
	} else {
		// Tick energy regen on every entry
		game.TickEnergy(char)
		database.SaveCharacter(char)
		handleMainMenuNew(chatID, msg.From.ID)
	}
}

// =============================================
// MAIN MENU
// =============================================

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

	rows := [][]tgbotapi.InlineKeyboardButton{}

	// Atalho contextual: reduz confusão e "vai e vem" durante combate.
	if char.State == "combat" || char.State == "dungeon_combat" {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⚔️ Voltar ao Combate", "combat_resume"),
		))
	}

	rows = append(rows,
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📊 Status", "menu_status"),
			tgbotapi.NewInlineKeyboardButtonData("🎒 Inventário", "menu_inventory"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🌟 Habilidades", "menu_skills"),
			tgbotapi.NewInlineKeyboardButtonData("⚡ Energia", "menu_energy"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⚔️ Explorar", "menu_explore"),
			tgbotapi.NewInlineKeyboardButtonData("🗺️ Viajar", "menu_travel"),
		),
	)

	// Sempre mostra "Vender" no menu principal; o handler valida se há loja no local.
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("🏪 Loja", "menu_shop"),
		tgbotapi.NewInlineKeyboardButtonData("💰 Vender", "menu_sell"),
	))
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("💎 Diamantes", "menu_diamonds"),
	))

	rows = append(rows,
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏚️ Masmorras", "menu_dungeon"),
			tgbotapi.NewInlineKeyboardButtonData("⚔️ Arena PVP", "menu_pvp"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏆 Ranking", "menu_rank"),
			tgbotapi.NewInlineKeyboardButtonData("👑 VIP & Caça Auto", "menu_vip"),
		),
	)
	return caption, tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}
}

func handleCombatResume(chatID int64, msgID int, userID int64) {
	char, _ := database.GetCharacter(userID)
	if char == nil || (char.State != "combat" && char.State != "dungeon_combat") {
		editMainMenuPhoto(chatID, msgID, userID)
		return
	}
	monster := game.Monsters[char.CombatMonsterID]
	if char.State == "dungeon_combat" {
		run, _ := database.GetActiveDungeonRun(char.ID)
		if run != nil {
			d := game.Dungeons[run.DungeonID]
			renderDungeonCombat(chatID, msgID, char, &monster, run, &d, "↩️ Continuando combate...\n")
			return
		}
	}
	renderCombat(chatID, msgID, char, &monster, "↩️ Continuando combate...\n")
}

// =============================================
// CHARACTER CREATION
// =============================================

func startCharacterCreation(chatID int64, msgID int, userID int64) {
	creationState[userID] = map[string]string{}

	human := game.Races["human"]
	elf := game.Races["elf"]
	dwarf := game.Races["dwarf"]
	halforc := game.Races["halforc"]

	caption := "🧬 *Escolha sua Raça*\n\n" +
		"Cada raça define seus atributos base e habilidade racial única.\n\n" +
		fmt.Sprintf("%s *Humano* — _%s_\n  ✨ _%s_\n\n", human.Emoji, human.Description, human.Trait) +
		fmt.Sprintf("%s *Elfo* — _%s_\n  ✨ _%s_\n\n", elf.Emoji, elf.Description, elf.Trait) +
		fmt.Sprintf("%s *Anão* — _%s_\n  ✨ _%s_\n\n", dwarf.Emoji, dwarf.Description, dwarf.Trait) +
		fmt.Sprintf("%s *Meio-Orc* — _%s_\n  ✨ _%s_", halforc.Emoji, halforc.Description, halforc.Trait)

	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("👤 Humano", "race_human"),
			tgbotapi.NewInlineKeyboardButtonData("🧝 Elfo", "race_elf"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⛏️ Anão", "race_dwarf"),
			tgbotapi.NewInlineKeyboardButtonData("👹 Meio-Orc", "race_halforc"),
		),
	)
	editPhoto(chatID, msgID, "welcome", caption, &kb)
}

func handleRaceSelect(chatID int64, msgID int, userID int64, race string) {
	if creationState[userID] == nil {
		creationState[userID] = map[string]string{}
	}
	creationState[userID]["race"] = race
	r := game.Races[race]
	warrior := game.Classes["warrior"]
	mage := game.Classes["mage"]
	rogue := game.Classes["rogue"]
	archer := game.Classes["archer"]

	caption := fmt.Sprintf(
		"%s *%s* selecionado!\n\n📖 _%s_\n✨ Traço: *%s*\n\n"+
			"💪 FOR:+%d 🤸 DEX:+%d 🏋️ CON:+%d\n🧠 INT:+%d 🦉 SAB:+%d 😎 CAR:+%d\n❤️ HP:+%d 💙 MP:+%d\n\n"+
			"⚔️ *Escolha sua Classe:*\n\n"+
			"%s *Guerreiro* — _%s_\n_%s_\n\n"+
			"%s *Mago* — _%s_\n_%s_\n\n"+
			"%s *Ladino* — _%s_\n_%s_\n\n"+
			"%s *Arqueiro* — _%s_\n_%s_",
		r.Emoji, r.Name, r.Description, r.Trait,
		r.BonusStr, r.BonusDex, r.BonusCon, r.BonusInt, r.BonusWis, r.BonusCha,
		r.BonusHP, r.BonusMP,
		warrior.Emoji, warrior.Description, fmt.Sprintf("❤️%d HP | 🎯 Papel: %s", warrior.BaseHP, warrior.Role),
		mage.Emoji, mage.Description, fmt.Sprintf("💙%d MP | 🎯 Papel: %s", mage.BaseMP, mage.Role),
		rogue.Emoji, rogue.Description, fmt.Sprintf("❤️%d HP | 🎯 Papel: %s", rogue.BaseHP, rogue.Role),
		archer.Emoji, archer.Description, fmt.Sprintf("❤️%d HP | 🎯 Papel: %s", archer.BaseHP, archer.Role),
	)
	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⚔️ Guerreiro", "class_warrior"),
			tgbotapi.NewInlineKeyboardButtonData("🧙 Mago", "class_mage"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🗡️ Ladino", "class_rogue"),
			tgbotapi.NewInlineKeyboardButtonData("🏹 Arqueiro", "class_archer"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⬅️ Voltar", "create_character"),
		),
	)
	editPhoto(chatID, msgID, assets.RaceImageKey(race), caption, &kb)
}

func handleClassSelect(chatID int64, msgID int, userID int64, class string) {
	state := creationState[userID]
	if state == nil || state["race"] == "" {
		editMsg(chatID, msgID, "❌ Selecione uma raça primeiro.", nil)
		return
	}
	state["class"] = class
	state["awaiting_name"] = "true"
	c := game.Classes[class]
	caption := fmt.Sprintf(
		"%s *%s*\n\n📖 _%s_\n🎯 Papel: *%s* | ❤️ %d HP (+%d/nível) | 💙 %d MP (+%d/nível)\n\n✏️ *Escreva o nome do seu personagem* (3-20 caracteres):",
		c.Emoji, c.Name, c.Description, c.Role,
		c.BaseHP, c.HPPerLevel, c.BaseMP, c.MPPerLevel,
	)
	editPhoto(chatID, msgID, assets.ClassImageKey(class), caption, nil)
}

func handleTextInput(msg *tgbotapi.Message) {
	// GM commands and pending GM sessions take priority
	if HandleGMCommand(msg) {
		return
	}
	if handleGMTextInput(msg) {
		return
	}

	userID := msg.From.ID

	state := creationState[userID]
	if state == nil || state["awaiting_name"] != "true" {
		return
	}
	name := strings.TrimSpace(msg.Text)
	if len(name) < 3 || len(name) > 20 {
		sendText(msg.Chat.ID, "❌ Nome deve ter entre 3 e 20 caracteres. Tente novamente:")
		return
	}
	race, class := state["race"], state["class"]
	hp, mp, atk, def, matk, mdef, spd := game.CalculateBaseStats(race, class)
	energyMax := game.MaxEnergy(1)

	char := &models.Character{
		PlayerID: userID, Name: name, Race: race, Class: class,
		Level: 1, Experience: 0, ExperienceNext: game.ExperienceForLevel(2),
		HP: hp, HPMax: hp, MP: mp, MPMax: mp,
		Energy: energyMax, EnergyMax: energyMax, EnergyRegenAt: time.Now(),
		Diamonds: 5, // welcome bonus
		Strength: 10 + game.Races[race].BonusStr, Dexterity: 10 + game.Races[race].BonusDex,
		Constitution: 10 + game.Races[race].BonusCon, Intelligence: 10 + game.Races[race].BonusInt,
		Wisdom: 10 + game.Races[race].BonusWis, Charisma: 10 + game.Races[race].BonusCha,
		Attack: atk, Defense: def, MagicAttack: matk, MagicDefense: mdef, Speed: spd,
		Gold: 50, CurrentMap: "village", State: "idle",
	}

	if err := database.CreateCharacter(char); err != nil {
		log.Printf("Error creating character: %v", err)
		sendText(msg.Chat.ID, "❌ Erro ao criar personagem. Tente novamente.")
		return
	}

	starterWeapons := map[string]string{"warrior": "sword_iron", "mage": "staff_oak", "rogue": "dagger_iron", "archer": "bow_short"}
	starterWeapon := starterWeapons[class]
	database.AddItem(char.ID, starterWeapon, "weapon", 1)
	database.AddItem(char.ID, "leather_armor", "armor", 1)
	database.AddItem(char.ID, "potion_small", "consumable", 3)
	database.AddItem(char.ID, "energy_drink", "consumable", 2)
	database.EquipItem(char.ID, starterWeapon, "weapon")
	database.EquipItem(char.ID, "leather_armor", "armor")
	delete(creationState, userID)

	r, c := game.Races[race], game.Classes[class]
	sw := game.Items[starterWeapon]
	caption := fmt.Sprintf(
		"🎉 *%s criado com sucesso!*\n\n%s %s | %s %s | Nível *1*\n\n❤️ HP: *%d* | 💙 MP: *%d* | ⚡ Energia: *%d*\n⚔️ Atq: *%d* | 🛡️ Def: *%d*\n💎 Diamantes de boas-vindas: *5*\n\n🎒 *Itens iniciais:*\n• %s %s\n• 🥋 Armadura de Couro\n• 🧪 3x Poção Pequena\n• ⚡ 2x Bebida Energética\n\nBoa sorte, *%s*! ⚔️",
		name, r.Emoji, r.Name, c.Emoji, c.Name,
		hp, mp, energyMax, atk, def,
		sw.Emoji, sw.Name, name,
	)
	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏰 Menu Principal", "menu_main"),
		),
	)
	sendPhoto(msg.Chat.ID, assets.ClassImageKey(class), caption, &kb)
}

// =============================================
// STATUS
// =============================================

func showStatus(chatID int64, msgID int, userID int64) {
	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}
	game.TickEnergy(char)
	database.SaveCharacter(char)

	skills, _ := database.GetLearnedSkills(char.ID)
	equipped, _ := database.GetEquippedItems(char.ID)

	equippedStr := "Nenhum"
	if len(equipped) > 0 {
		var parts []string
		for _, e := range equipped {
			if item, ok := game.Items[e.ItemID]; ok {
				parts = append(parts, item.Emoji+" "+item.Name)
			}
		}
		equippedStr = strings.Join(parts, "\n  • ")
	}
	skillStr := "Nenhuma aprendida"
	if len(skills) > 0 {
		var parts []string
		for _, s := range skills {
			if sk, ok := game.Skills[s.SkillID]; ok {
				p := ""
				if sk.Passive {
					p = " *(passiva)*"
				}
				parts = append(parts, sk.Emoji+" "+sk.Name+p)
			}
		}
		skillStr = strings.Join(parts, "\n  • ")
	}

	r, c := game.Races[char.Race], game.Classes[char.Class]
	xpPct := int(float64(char.Experience) / float64(char.ExperienceNext) * 10)
	xpBar := strings.Repeat("▓", xpPct) + strings.Repeat("░", 10-xpPct)
	eBar := game.EnergyBar(char.Energy, char.EnergyMax)

	regenStr := ""
	if char.Energy < char.EnergyMax {
		d := game.NextRegenIn(char)
		regenStr = fmt.Sprintf(" _(+1 em %dm%02ds)_", int(d.Minutes()), int(d.Seconds())%60)
	}

	recalculateStats(char) // garante EquipCABonus/HitBonus
	defAttr := game.DefensiveAttr(char.Class, char.Constitution, char.Dexterity, char.Intelligence)
	playerCA := game.CharacterCA(char.Class, defAttr, char.EquipCABonus)
	atkBonus := game.CharacterAttackBonus(char.Class, char.Level, char.Strength, char.Dexterity, char.Intelligence) + char.EquipHitBonus

	caption := fmt.Sprintf(
		"📊 *Status de %s*\n\n%s %s | %s %s | Nível *%d*\n✨ `[%s]` %d/%d XP | 🌟 *%d* pts habilidade\n\n"+
			"💖 *Vitals*\n❤️ HP: *%d*/%d | 💙 MP: *%d*/%d\n⚡ %s *%d*/%d%s\n\n"+
			"💰 *Moedas*\n🪙 Ouro: *%d* | 💎 Diamantes: *%d*\n\n"+
			"⚔️ *Atributos*\n💪 FOR:*%d* 🤸 DEX:*%d* 🏋️ CON:*%d*\n🧠 INT:*%d* 🦉 SAB:*%d* 😎 CAR:*%d*\n\n"+
			"🎯 *Combate (Tormenta 20)*\n🛡️ CA:*%d* | 🎲 Bônus Ataque:*+%d*\n⚔️ Atq:*%d* 🔮 MAtq:*%d* 💨 Vel:*%d*\n\n"+
			"🗡️ *Equipado:*\n  • %s\n\n📚 *Habilidades:*\n  • %s%s",
		char.Name, r.Emoji, r.Name, c.Emoji, c.Name, char.Level,
		xpBar, char.Experience, char.ExperienceNext, char.SkillPoints,
		char.HP, char.HPMax, char.MP, char.MPMax,
		eBar, char.Energy, char.EnergyMax, regenStr,
		char.Gold, char.Diamonds,
		char.Strength, char.Dexterity, char.Constitution,
		char.Intelligence, char.Wisdom, char.Charisma,
		playerCA, atkBonus, char.Attack, char.MagicAttack, char.Speed,
		equippedStr, skillStr,
		func() string {
			if !char.XPBoostExpiry.IsZero() && time.Now().Before(char.XPBoostExpiry) {
				remaining := time.Until(char.XPBoostExpiry).Round(time.Minute)
				return fmt.Sprintf("\n\n📖 *Bênção do Sábio ativa!* +50%% XP por *%s*", remaining)
			}
			return ""
		}(),
	)
	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⚡ Energia", "menu_energy"),
			tgbotapi.NewInlineKeyboardButtonData("💎 Diamantes", "menu_diamonds"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🗑️ Apagar Personagem", "delete_character"),
			tgbotapi.NewInlineKeyboardButtonData("🏰 Menu", "menu_main"),
		),
	)
	editPhoto(chatID, msgID, "status", caption, &kb)
}

// =============================================
// ⚡ ENERGY MENU
// =============================================

func showEnergyMenu(chatID int64, msgID int, userID int64) {
	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}
	game.TickEnergy(char)
	database.SaveCharacter(char)

	eBar := game.EnergyBar(char.Energy, char.EnergyMax)
	hpMissing := char.HPMax - char.HP
	mpMissing := char.MPMax - char.MP
	hpCost := (hpMissing + game.EnergyPerHP - 1) / game.EnergyPerHP
	mpCost := (mpMissing + game.EnergyPerMP - 1) / game.EnergyPerMP
	bothCost := game.InnEnergyCost(char)

	regenStr := ""
	if char.Energy < char.EnergyMax {
		d := game.NextRegenIn(char)
		regenStr = fmt.Sprintf("⏱️ Próxima recarga: *%dm %02ds*\n", int(d.Minutes()), int(d.Seconds())%60)
	} else {
		regenStr = "✅ Energia *CHEIA*!\n"
	}
	regenEvery := game.RegenInterval(char.EnergyMax > game.EnergyBaseMax)
	regenMinutes := int(regenEvery / time.Minute)

	caption := fmt.Sprintf(
		"⚡ *Sistema de Energia*\n\n"+
			"%s *%d*/%d Energia\n%s\n\n"+
			"*Como funciona:*\n"+
			"• 1 ⚡ = *%d HP* recuperado\n"+
			"• 1 ⚡ = *%d MP* recuperado\n"+
			"• Regenera 1 ⚡ a cada *%d minutos*\n\n"+
			"*Recuperação disponível:*\n"+
			"❤️ HP faltando: *%d* → custa *%d* ⚡\n"+
			"💙 MP faltando: *%d* → custa *%d* ⚡\n"+
			"💖 Ambos → custa *%d* ⚡\n\n"+
			"💎 *30 diamantes* = recarga total instantânea",
		eBar, char.Energy, char.EnergyMax, regenStr,
		game.EnergyPerHP, game.EnergyPerMP, regenMinutes,
		hpMissing, hpCost,
		mpMissing, mpCost,
		bothCost,
	)

	rows := [][]tgbotapi.InlineKeyboardButton{}
	if hpMissing > 0 && char.Energy > 0 {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("❤️ Recuperar HP (%d⚡)", hpCost),
				"energy_heal_hp",
			),
		))
	}
	if mpMissing > 0 && char.Energy > 0 {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("💙 Recuperar MP (%d⚡)", mpCost),
				"energy_heal_mp",
			),
		))
	}
	if (hpMissing > 0 || mpMissing > 0) && char.Energy > 0 {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("💖 Recuperar Tudo (%d⚡)", bothCost),
				"energy_heal_both",
			),
		))
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("💎 Recarregar com Diamantes", "diamond_buy_energy_full"),
	))
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("🏰 Menu", "menu_main"),
	))
	kb := tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}
	editPhoto(chatID, msgID, "rest", caption, &kb)
}

func handleEnergyHealHP(chatID int64, msgID int, userID int64) {
	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}
	game.TickEnergy(char)
	hpGained, energySpent := game.RecoverHPWithEnergy(char, char.Energy)
	database.SaveCharacter(char)
	if hpGained == 0 {
		editPhoto(chatID, msgID, "rest", "❌ HP já está cheio ou sem energia!", bkp("menu_energy"))
		return
	}
	caption := fmt.Sprintf("❤️ *HP recuperado!*\n\n+%d HP (gastou %d⚡)\n\n❤️ HP: *%d*/%d\n⚡ Energia: *%d*/%d",
		hpGained, energySpent, char.HP, char.HPMax, char.Energy, char.EnergyMax)
	editPhoto(chatID, msgID, "rest", caption, bkp("menu_energy"))
}

func handleEnergyHealMP(chatID int64, msgID int, userID int64) {
	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}
	game.TickEnergy(char)
	mpGained, energySpent := game.RecoverMPWithEnergy(char, char.Energy)
	database.SaveCharacter(char)
	if mpGained == 0 {
		editPhoto(chatID, msgID, "rest", "❌ MP já está cheio ou sem energia!", bkp("menu_energy"))
		return
	}
	caption := fmt.Sprintf("💙 *MP recuperado!*\n\n+%d MP (gastou %d⚡)\n\n💙 MP: *%d*/%d\n⚡ Energia: *%d*/%d",
		mpGained, energySpent, char.MP, char.MPMax, char.Energy, char.EnergyMax)
	editPhoto(chatID, msgID, "rest", caption, bkp("menu_energy"))
}

func handleEnergyHealBoth(chatID int64, msgID int, userID int64) {
	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}
	game.TickEnergy(char)
	hpGained, eHP := game.RecoverHPWithEnergy(char, char.Energy)
	mpGained, eMP := game.RecoverMPWithEnergy(char, char.Energy)
	database.SaveCharacter(char)
	if hpGained+mpGained == 0 {
		editPhoto(chatID, msgID, "rest", "❌ Já está cheio ou sem energia!", bkp("menu_energy"))
		return
	}
	caption := fmt.Sprintf(
		"💖 *Recuperação completa!*\n\n+%d HP | +%d MP\nTotal gasto: %d⚡\n\n❤️ HP: *%d*/%d | 💙 MP: *%d*/%d\n⚡ Energia: *%d*/%d",
		hpGained, mpGained, eHP+eMP,
		char.HP, char.HPMax, char.MP, char.MPMax,
		char.Energy, char.EnergyMax,
	)
	editPhoto(chatID, msgID, "rest", caption, bkp("menu_energy"))
}

// =============================================
// 💎 DIAMOND MENU & SHOP
// =============================================

func showDiamondMenu(chatID int64, msgID int, userID int64) {
	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}
	caption := fmt.Sprintf(
		"💎 *Diamantes*\n\nVocê tem: *%d* 💎\n\n"+
			"*Diamantes servem para:*\n"+
			"⚡ Recarregar Energia instantaneamente\n"+
			"💖 Cura divina (HP+MP completos)\n"+
			"🔮 Token de reviver (sem perder ouro)\n"+
			"📖 Bênção do Sábio (+50%% XP por 30min)\n\n"+
			"*Ganhe diamantes:*\n"+
			"• 🎁 Bônus diário: *3 💎* por dia\n"+
			"• 💀 Derrotar bosses (chance de drop)\n"+
			"• 💰 Comprar pacotes abaixo",
		char.Diamonds,
	)
	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🎁 Bônus Diário", "daily_bonus"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🛒 Loja de Diamantes", "diamond_shop"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("💳 Comprar via Pix", "menu_pix"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏰 Menu", "menu_main"),
		),
	)
	editPhoto(chatID, msgID, "shop", caption, &kb)
}

func showDiamondShop(chatID int64, msgID int, userID int64) {
	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}
	caption := fmt.Sprintf("💎 *Loja de Diamantes*\n💎 Diamantes: *%d*\n\n", char.Diamonds)

	var rows [][]tgbotapi.InlineKeyboardButton

	// Seção 1: Serviços instantâneos (efeito imediato, sem inventário)
	caption += "⚡ *Serviços Instantâneos:*\n"
	// Ordenar por custo para exibição consistente
	type dItem struct {
		id   string
		item game.DiamondItem
	}
	var dItems []dItem
	for id, it := range game.DiamondItems {
		dItems = append(dItems, dItem{id, it})
	}
	sort.Slice(dItems, func(i, j int) bool { return dItems[i].item.Cost < dItems[j].item.Cost })
	for _, di := range dItems {
		item := di.item
		canBuy := "✅"
		if char.Diamonds < item.Cost {
			canBuy = "❌"
		}
		caption += fmt.Sprintf("%s %s *%s* — *%d* 💎\n", canBuy, item.Emoji, item.Name, item.Cost)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("%s %s (%d💎)", item.Emoji, item.Name, item.Cost),
				"diamond_buy_"+item.ID,
			),
		))
	}

	// Seção 2: Itens especiais (vão para o inventário)
	var diamondConsumables []models.Item
	for _, item := range game.Items {
		if item.DiamondPrice > 0 && item.MinLevel <= char.Level {
			diamondConsumables = append(diamondConsumables, item)
		}
	}
	sort.Slice(diamondConsumables, func(i, j int) bool {
		return diamondConsumables[i].DiamondPrice < diamondConsumables[j].DiamondPrice
	})

	if len(diamondConsumables) > 0 {
		caption += "🧪 *Itens Especiais:*\n"
		for _, item := range diamondConsumables {
			canBuy := "✅"
			if char.Diamonds < item.DiamondPrice {
				canBuy = "❌"
			}
			caption += fmt.Sprintf("%s %s *%s* — *%d* 💎\n", canBuy, item.Emoji, item.Name, item.DiamondPrice)
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					fmt.Sprintf("%s %s (%d💎)", item.Emoji, item.Name, item.DiamondPrice),
					"diamond_item_"+item.ID,
				),
			))
		}
	}

	rows = append(rows,
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("💳 Comprar Diamantes via Pix", "menu_pix"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⬅️ Voltar aos Diamantes", "menu_diamonds"),
		),
	)
	kb := tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}
	editPhoto(chatID, msgID, "shop", caption, &kb)
}

func handleDiamondBuy(chatID int64, msgID int, userID int64, itemID string) {
	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}
	game.TickEnergy(char)

	item, ok := game.DiamondItems[itemID]
	if !ok {
		return
	}
	if char.Diamonds < item.Cost {
		editPhoto(chatID, msgID, "shop", fmt.Sprintf("❌ Diamantes insuficientes!\nPrecisa de *%d* 💎 mas tem *%d* 💎", item.Cost, char.Diamonds), bkp("menu_diamonds"))
		return
	}

	char.Diamonds -= item.Cost
	var result string

	switch itemID {
	case "energy_full":
		gained := char.EnergyMax - char.Energy
		char.Energy = char.EnergyMax
		result = fmt.Sprintf("⚡ Energia totalmente recarregada!\n+%d⚡ → *%d*/%d", gained, char.Energy, char.EnergyMax)
	case "hp_full":
		hpG := char.HPMax - char.HP
		mpG := char.MPMax - char.MP
		char.HP = char.HPMax
		char.MP = char.MPMax
		result = fmt.Sprintf("💖 Cura Divina!\n+%d HP, +%d MP\n❤️ *%d*/%d | 💙 *%d*/%d", hpG, mpG, char.HP, char.HPMax, char.MP, char.MPMax)
	default:
		// Unknown service — should not happen
		result = "✅ Serviço aplicado!"
	}

	database.LogDiamond(char.ID, -item.Cost, "buy_"+itemID)
	database.SaveCharacter(char)
	caption := fmt.Sprintf("%s *%s*\n\n%s\n\n💎 Diamantes restantes: *%d*", item.Emoji, item.Name, result, char.Diamonds)
	editPhoto(chatID, msgID, "shop", caption, bkp("menu_diamonds"))
}

// handleDiamondItemBuy compra um item do Items[] usando DiamondPrice.
func handleDiamondItemBuy(chatID int64, msgID int, userID int64, itemID string) {
	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}
	item, ok := game.Items[itemID]
	if !ok || item.DiamondPrice <= 0 {
		showDiamondShop(chatID, msgID, userID)
		return
	}
	if char.Diamonds < item.DiamondPrice {
		editPhoto(chatID, msgID, "shop",
			fmt.Sprintf("❌ Diamantes insuficientes!\nPrecisa de *%d* 💎 mas tem *%d* 💎", item.DiamondPrice, char.Diamonds),
			bkp("diamond_shop"))
		return
	}
	char.Diamonds -= item.DiamondPrice
	database.AddItem(char.ID, item.ID, item.Type, 1)
	database.LogDiamond(char.ID, -item.DiamondPrice, "buy_item_"+itemID)
	database.SaveCharacter(char)
	caption := fmt.Sprintf(
		"✅ *Comprado com Diamantes!*\n\n%s *%s*\nPago: *%d* 💎\n\n💎 Diamantes restantes: *%d*",
		item.Emoji, item.Name, item.DiamondPrice, char.Diamonds,
	)
	editPhoto(chatID, msgID, "shop", caption, bkp("diamond_shop"))
}

func handleDailyBonus(chatID int64, msgID int, userID int64) {
	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}
	diamonds, ok := database.ClaimDailyBonus(char.ID)
	if !ok {
		editPhoto(chatID, msgID, "shop", "🎁 *Bônus diário já coletado!*\n\nVolte amanhã para mais *3 💎*!", bkp("menu_diamonds"))
		return
	}
	char.Diamonds += diamonds
	database.SaveCharacter(char)
	database.LogDiamond(char.ID, diamonds, "daily_bonus")
	caption := fmt.Sprintf("🎁 *Bônus Diário Coletado!*\n\n+%d 💎 adicionados!\n\nDiamantes totais: *%d* 💎\n\nVolte amanhã para mais diamantes!", diamonds, char.Diamonds)
	editPhoto(chatID, msgID, "shop", caption, bkp("menu_diamonds"))
}

// =============================================
// 🏪 SHOP (with quantity selection)
// =============================================

func showShopMenu(chatID int64, msgID int, userID int64) {
	showShopPage(chatID, msgID, userID, "consumable")
}

func showShopPage(chatID int64, msgID int, userID int64, pageType string) {
	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}

	cart := shopCart[userID]
	if cart == nil {
		cart = &models.ShopCart{TabType: pageType}
		shopCart[userID] = cart
	}
	cart.TabType = pageType

	tabNames := map[string]string{"consumable": "Consumíveis", "weapon": "Armas", "armor": "Armaduras", "accessory": "Acessórios"}
	caption := fmt.Sprintf("🏪 *Loja — %s*\n🪙 *%d* ouro | 💎 *%d* diamantes\n\n", tabNames[pageType], char.Gold, char.Diamonds)

	// Carrinho resumido no topo
	if len(cart.Items) > 0 {
		caption += "🛒 *Carrinho:*\n"
		total := 0
		for _, ci := range cart.Items {
			it := game.Items[ci.ItemID]
			subtotal := it.Price * ci.Qty
			total += subtotal
			caption += fmt.Sprintf("  %s %s ×%d = *%d*🪙\n", it.Emoji, it.Name, ci.Qty, subtotal)
		}
		caption += fmt.Sprintf("*Total: %d*🪙\n\n", total)
	}

	// Lista de itens (somente ouro — sem DiamondPrice exclusivo)
	var rows [][]tgbotapi.InlineKeyboardButton
	var itemsSorted []models.Item
	for _, item := range game.Items {
		if item.Type != pageType || item.MinLevel > char.Level {
			continue
		}
		if item.ClassReq != "" && item.ClassReq != char.Class {
			continue
		}
		// Excluir itens que só se vendem com diamante (Price == 0)
		if item.Price <= 0 {
			continue
		}
		itemsSorted = append(itemsSorted, item)
	}
	// Ordenar por nível mínimo depois nome
	sort.Slice(itemsSorted, func(i, j int) bool {
		if itemsSorted[i].MinLevel != itemsSorted[j].MinLevel {
			return itemsSorted[i].MinLevel < itemsSorted[j].MinLevel
		}
		return itemsSorted[i].Name < itemsSorted[j].Name
	})

	for _, item := range itemsSorted {
		afford := char.Gold >= item.Price
		emoji := "🛒"
		if !afford {
			emoji = "💸"
		}
		// Checar se já está no carrinho
		inCart := 0
		for _, ci := range cart.Items {
			if ci.ItemID == item.ID {
				inCart = ci.Qty
				break
			}
		}
		cartTag := ""
		if inCart > 0 {
			cartTag = fmt.Sprintf(" [%d no carrinho]", inCart)
		}
		caption += fmt.Sprintf("%s *%s* — *%d*🪙%s\n", item.Emoji, item.Name, item.Price, cartTag)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("%s %s %s (+1)", emoji, item.Emoji, item.Name),
				"shop_add_"+item.ID,
			),
		))
	}

	// Botão de checkout se há itens no carrinho
	if len(cart.Items) > 0 {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("✅ Ver Carrinho (%d item(ns))", len(cart.Items)),
				"shop_checkout",
			),
		))
	}

	// Tabs de navegação
	rows = append(rows, []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("🧪 Consumíveis", "shop_page_consumable"),
		tgbotapi.NewInlineKeyboardButtonData("⚔️ Armas", "shop_page_weapon"),
	})
	rows = append(rows, []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("🛡️ Armaduras", "shop_page_armor"),
		tgbotapi.NewInlineKeyboardButtonData("💍 Acessórios", "shop_page_accessory"),
	})
	rows = append(rows, []tgbotapi.InlineKeyboardButton{
		menuBackButton(userID, "menu_shop", "menu_main"),
	})

	imageKey := assets.ItemTypeImageKey(pageType)
	kb := tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}
	editPhoto(chatID, msgID, imageKey, caption, &kb)
}

// handleShopAddItem adiciona +1 unidade de um item ao carrinho e reexibe a página.
func handleShopAddItem(chatID int64, msgID int, userID int64, itemID string) {
	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}
	item, ok := game.Items[itemID]
	if !ok || item.Price <= 0 {
		return
	}
	cart := shopCart[userID]
	if cart == nil {
		cart = &models.ShopCart{TabType: "consumable"}
		shopCart[userID] = cart
	}
	// Adiciona ou incrementa
	found := false
	for i, ci := range cart.Items {
		if ci.ItemID == itemID {
			cart.Items[i].Qty++
			found = true
			break
		}
	}
	if !found {
		cart.Items = append(cart.Items, models.ShopCartItem{ItemID: itemID, Qty: 1})
	}
	showShopPage(chatID, msgID, userID, cart.TabType)
}

// handleShopChangeQty incrementa ou decrementa qty de item no carrinho.
func handleShopChangeQty(chatID int64, msgID int, userID int64, itemID string, delta int) {
	cart := shopCart[userID]
	if cart == nil {
		showShopMenu(chatID, msgID, userID)
		return
	}
	for i, ci := range cart.Items {
		if ci.ItemID == itemID {
			cart.Items[i].Qty += delta
			if cart.Items[i].Qty <= 0 {
				cart.Items = append(cart.Items[:i], cart.Items[i+1:]...)
			}
			break
		}
	}
	handleShopCheckout(chatID, msgID, userID)
}

// handleShopRemoveItem remove um item do carrinho.
func handleShopRemoveItem(chatID int64, msgID int, userID int64, itemID string) {
	cart := shopCart[userID]
	if cart == nil {
		showShopMenu(chatID, msgID, userID)
		return
	}
	for i, ci := range cart.Items {
		if ci.ItemID == itemID {
			cart.Items = append(cart.Items[:i], cart.Items[i+1:]...)
			break
		}
	}
	if len(cart.Items) == 0 {
		showShopPage(chatID, msgID, userID, cart.TabType)
		return
	}
	handleShopCheckout(chatID, msgID, userID)
}

// handleShopCheckout exibe o carrinho completo com controles de quantidade.
func handleShopCheckout(chatID int64, msgID int, userID int64) {
	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}
	cart := shopCart[userID]
	if cart == nil || len(cart.Items) == 0 {
		showShopMenu(chatID, msgID, userID)
		return
	}

	totalGold := 0
	caption := fmt.Sprintf("🛒 *Carrinho de Compras*\n🪙 Ouro disponível: *%d*\n\n", char.Gold)
	var rows [][]tgbotapi.InlineKeyboardButton

	for _, ci := range cart.Items {
		item := game.Items[ci.ItemID]
		subtotal := item.Price * ci.Qty
		totalGold += subtotal
		caption += fmt.Sprintf("%s *%s* — %d × *%d*🪙 = *%d*🪙\n_%s_\n", item.Emoji, item.Name, ci.Qty, item.Price, subtotal, item.Description)
		rows = append(rows, []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("➖ %s", item.Name), "shop_dec_"+item.ID),
			tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("%d", ci.Qty), "shop_checkout"),
			tgbotapi.NewInlineKeyboardButtonData("➕", "shop_inc_"+item.ID),
			tgbotapi.NewInlineKeyboardButtonData("🗑️", "shop_rem_"+item.ID),
		})
	}

	caption += fmt.Sprintf("\n💰 *Total: %d*🪙", totalGold)
	canAfford := char.Gold >= totalGold
	if !canAfford {
		caption += fmt.Sprintf("\n❌ Faltam *%d*🪙", totalGold-char.Gold)
	}

	if canAfford {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("✅ Comprar tudo por %d🪙", totalGold),
				"shop_confirm_buy",
			),
		))
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("🔙 Continuar comprando", "shop_page_"+cart.TabType),
		tgbotapi.NewInlineKeyboardButtonData("❌ Cancelar", "shop_cancel_buy"),
	))

	kb := tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}
	editPhoto(chatID, msgID, "shop", caption, &kb)
}

func handleShopQty(chatID int64, msgID int, userID int64, qtyStr string) {
	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}
	state := shopQtyState[userID]
	if state == nil {
		showShopMenu(chatID, msgID, userID)
		return
	}
	item := game.Items[state.ItemID]

	if qtyStr == "diamond" {
		state.Quantity = 1
		state.PayWith = "diamonds"
	} else {
		qty := 1
		fmt.Sscanf(qtyStr, "%d", &qty)
		if qty < 1 {
			qty = 1
		}
		state.Quantity = qty
		state.PayWith = "gold"
	}

	totalGold := item.Price * state.Quantity
	totalDiamond := item.DiamondPrice * state.Quantity

	canAfford := false
	payLabel := ""
	if state.PayWith == "diamonds" {
		canAfford = char.Diamonds >= totalDiamond
		payLabel = fmt.Sprintf("*%d* 💎", totalDiamond)
	} else {
		canAfford = char.Gold >= totalGold
		payLabel = fmt.Sprintf("*%d* 🪙", totalGold)
	}

	caption := fmt.Sprintf(
		"🛒 *Confirmar Compra*\n\n%s *%s* × *%d*\nTotal: %s\n\n🪙 Ouro: *%d* | 💎 *%d*",
		item.Emoji, item.Name, state.Quantity,
		payLabel, char.Gold, char.Diamonds,
	)

	rows := [][]tgbotapi.InlineKeyboardButton{
		{
			tgbotapi.NewInlineKeyboardButtonData("1x", "shop_qty_1"),
			tgbotapi.NewInlineKeyboardButtonData("3x", "shop_qty_3"),
			tgbotapi.NewInlineKeyboardButtonData("5x", "shop_qty_5"),
			tgbotapi.NewInlineKeyboardButtonData("10x", "shop_qty_10"),
		},
	}
	if canAfford {
		confirmLabel := fmt.Sprintf("✅ Confirmar %dx por %s", state.Quantity, payLabel)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(confirmLabel, "shop_confirm_buy"),
		))
	} else {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❌ Sem recursos suficientes", "shop_cancel_buy"),
		))
	}
	if item.DiamondPrice > 0 {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("💎 Pagar com Diamantes (%d💎/un)", item.DiamondPrice),
				"shop_qty_diamond",
			),
		))
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("❌ Cancelar", "shop_cancel_buy"),
	))

	kb := tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}
	editPhoto(chatID, msgID, assets.ItemTypeImageKey(item.Type), caption, &kb)
}

func handleShopConfirmBuy(chatID int64, msgID int, userID int64) {
	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}
	cart := shopCart[userID]
	if cart == nil || len(cart.Items) == 0 {
		showShopMenu(chatID, msgID, userID)
		return
	}

	// Calcular total
	total := 0
	for _, ci := range cart.Items {
		item := game.Items[ci.ItemID]
		total += item.Price * ci.Qty
	}
	if char.Gold < total {
		editPhoto(chatID, msgID, "shop", fmt.Sprintf("❌ Ouro insuficiente! Precisa *%d* 🪙", total), bkp("menu_shop"))
		return
	}

	char.Gold -= total
	summary := ""
	for _, ci := range cart.Items {
		item := game.Items[ci.ItemID]
		database.AddItem(char.ID, item.ID, item.Type, ci.Qty)
		summary += fmt.Sprintf("%s *%s* × %d\n", item.Emoji, item.Name, ci.Qty)
	}

	database.SaveCharacter(char)
	delete(shopCart, userID)

	caption := fmt.Sprintf(
		"✅ *Compra realizada!*\n\n%s\n💰 Pago: *%d* 🪙\n🪙 Ouro restante: *%d*",
		summary, total, char.Gold,
	)
	editPhoto(chatID, msgID, "shop", caption, bkp("menu_shop"))
}

// =============================================
// 💰 SELL MENU
// =============================================

func showSellMenu(chatID int64, msgID int, userID int64) {
	// Compat legado.
	showSellHome(chatID, msgID, userID)
}

func showSellHome(chatID int64, msgID int, userID int64) {
	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}

	cart := sellCart[userID]
	if cart == nil {
		cart = &models.SellCart{}
		sellCart[userID] = cart
	}

	caption := fmt.Sprintf("💰 *Vender Itens*\n🪙 Ouro atual: *%d*\n\nSelecione uma categoria:", char.Gold)
	if len(cart.Items) > 0 {
		caption += "\n\n🛒 *Carrinho:*\n"
		total := 0
		for _, sc := range cart.Items {
			it := game.Items[sc.ItemID]
			sub := it.SellPrice * sc.Qty
			total += sub
			caption += fmt.Sprintf("  %s %s ×%d = *%d*🪙\n", it.Emoji, it.Name, sc.Qty, sub)
		}
		caption += fmt.Sprintf("*Total: +%d*🪙\n\n", total)
	}

	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🧪 Consumíveis", "sell_tab_consumable"),
			tgbotapi.NewInlineKeyboardButtonData("⚔️ Armas", "sell_tab_weapon"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🛡️ Armaduras", "sell_tab_armor"),
			tgbotapi.NewInlineKeyboardButtonData("💍 Acessórios", "sell_tab_accessory"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📋 Todos", "sell_tab_all"),
		),
	)

	if len(cart.Items) > 0 {
		kb.InlineKeyboard = append(kb.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("✅ Ver Carrinho (%d item(ns))", len(cart.Items)),
				"sell_checkout",
			),
		))
	}
	kb.InlineKeyboard = append(kb.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(
		menuBackButton(userID, "menu_sell", "menu_main"),
	))
	editPhoto(chatID, msgID, "shop", caption, &kb)
}

func showSellPage(chatID int64, msgID int, userID int64, filterType string) {
	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}
	inv, _ := database.GetInventory(char.ID)

	tabNames := map[string]string{
		"all":        "Todos",
		"consumable": "Consumíveis",
		"weapon":     "Armas",
		"armor":      "Armaduras",
		"accessory":  "Acessórios",
	}
	tabName := tabNames[filterType]
	if tabName == "" {
		tabName = "Todos"
		filterType = "all"
	}

	caption := fmt.Sprintf("💰 *Vendas — %s*\n🪙 Ouro atual: *%d*\n\nEscolha um item:", tabName, char.Gold)
	var rows [][]tgbotapi.InlineKeyboardButton
	hasItems := false

	for _, invItem := range inv {
		item, ok := game.Items[invItem.ItemID]
		if !ok || item.SellPrice <= 0 {
			continue
		}
		if filterType != "all" && item.Type != filterType {
			continue
		}
		hasItems = true
		label := fmt.Sprintf("%s %s", item.Emoji, item.Name)
		if invItem.Equipped {
			label += " ✅"
		}
		label += fmt.Sprintf(" x%d", invItem.Quantity)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				label,
				"sell_item_"+filterType+"_"+item.ID,
			),
		))
	}

	if !hasItems {
		caption += "\n_Nenhum item para vender nesta categoria._"
	}

	rows = append(rows, []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("🧪 Consumíveis", "sell_tab_consumable"),
		tgbotapi.NewInlineKeyboardButtonData("⚔️ Armas", "sell_tab_weapon"),
	})
	rows = append(rows, []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("🛡️ Armaduras", "sell_tab_armor"),
		tgbotapi.NewInlineKeyboardButtonData("💍 Acessórios", "sell_tab_accessory"),
	})
	rows = append(rows, []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("📋 Todos", "sell_tab_all"),
		tgbotapi.NewInlineKeyboardButtonData("⬅️ Voltar", "menu_sell"),
	})

	cart := sellCart[userID]
	if cart != nil && len(cart.Items) > 0 {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("✅ Ver Carrinho (%d item(ns))", len(cart.Items)),
				"sell_checkout",
			),
		))
	}

	kb := tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}
	editPhoto(chatID, msgID, "shop", caption, &kb)
}

func showSellItem(chatID int64, msgID int, userID int64, tab string, itemID string) {
	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}
	item, ok := game.Items[itemID]
	if !ok || item.SellPrice <= 0 {
		showSellPage(chatID, msgID, userID, tab)
		return
	}
	count := database.GetItemCount(char.ID, itemID)
	if count <= 0 {
		showSellPage(chatID, msgID, userID, tab)
		return
	}

	cart := sellCart[userID]
	inCart := 0
	if cart != nil {
		for _, sc := range cart.Items {
			if sc.ItemID == itemID {
				inCart = sc.Qty
				break
			}
		}
	}

	equippedTag := ""
	inv, _ := database.GetInventory(char.ID)
	for _, ii := range inv {
		if ii.ItemID == itemID && ii.Equipped {
			equippedTag = "\n✅ *Equipado no momento*"
			break
		}
	}

	caption := fmt.Sprintf(
		"💰 *Item para Venda*\n\n%s *%s*\nPreço de venda: *%d*🪙/un\nQuantidade no inventário: *%d*\nNo carrinho: *%d*\n\n_%s_%s",
		item.Emoji, item.Name, item.SellPrice, count, inCart, item.Description, equippedTag,
	)
	rows := [][]tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("➕ Adicionar ao carrinho", "sell_add_"+tab+"_"+item.ID),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⬅️ Voltar", "sell_back_"+tab),
		),
	}
	if cart != nil && len(cart.Items) > 0 {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("✅ Ver Carrinho (%d item(ns))", len(cart.Items)),
				"sell_checkout",
			),
		))
	}
	kb := tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}
	editPhoto(chatID, msgID, "shop", caption, &kb)
}

func addSellItemToCart(userID int64, charID int, itemID string) bool {
	item, ok := game.Items[itemID]
	if !ok || item.SellPrice <= 0 {
		return false
	}
	maxQty := database.GetItemCount(charID, itemID)
	if maxQty <= 0 {
		return false
	}
	cart := sellCart[userID]
	if cart == nil {
		cart = &models.SellCart{}
		sellCart[userID] = cart
	}
	for i, sc := range cart.Items {
		if sc.ItemID == itemID {
			if cart.Items[i].Qty < maxQty {
				cart.Items[i].Qty++
			}
			return true
		}
	}
	cart.Items = append(cart.Items, models.SellCartItem{ItemID: itemID, Qty: 1})
	return true
}

// handleSellAddItem adiciona +1 de um item ao carrinho de venda.
func handleSellAddItem(chatID int64, msgID int, userID int64, itemID string) {
	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}
	if !addSellItemToCart(userID, char.ID, itemID) {
		return
	}
	showSellHome(chatID, msgID, userID)
}

// handleSellChangeQty incrementa ou decrementa a qty de um item no carrinho de venda.
func handleSellChangeQty(chatID int64, msgID int, userID int64, itemID string, delta int) {
	cart := sellCart[userID]
	if cart == nil {
		showSellMenu(chatID, msgID, userID)
		return
	}
	for i, sc := range cart.Items {
		if sc.ItemID == itemID {
			cart.Items[i].Qty += delta
			if cart.Items[i].Qty <= 0 {
				cart.Items = append(cart.Items[:i], cart.Items[i+1:]...)
			}
			break
		}
	}
	handleSellCheckout(chatID, msgID, userID)
}

// handleSellRemoveItem remove um item do carrinho de venda.
func handleSellRemoveItem(chatID int64, msgID int, userID int64, itemID string) {
	cart := sellCart[userID]
	if cart == nil {
		showSellMenu(chatID, msgID, userID)
		return
	}
	for i, sc := range cart.Items {
		if sc.ItemID == itemID {
			cart.Items = append(cart.Items[:i], cart.Items[i+1:]...)
			break
		}
	}
	if len(cart.Items) == 0 {
		showSellMenu(chatID, msgID, userID)
		return
	}
	handleSellCheckout(chatID, msgID, userID)
}

// handleSellCheckout exibe o carrinho de venda com controles ➖/➕/🗑️, idêntico ao checkout de compra.
func handleSellCheckout(chatID int64, msgID int, userID int64) {
	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}
	cart := sellCart[userID]
	if cart == nil || len(cart.Items) == 0 {
		showSellMenu(chatID, msgID, userID)
		return
	}

	totalGold := 0
	caption := fmt.Sprintf("🛒 *Carrinho de Venda*\n🪙 Ouro atual: *%d*\n\n", char.Gold)
	var rows [][]tgbotapi.InlineKeyboardButton

	for _, sc := range cart.Items {
		item := game.Items[sc.ItemID]
		subtotal := item.SellPrice * sc.Qty
		totalGold += subtotal
		maxQty := database.GetItemCount(char.ID, sc.ItemID)
		caption += fmt.Sprintf("%s *%s* — %d × *%d*🪙/un = *%d*🪙 (tens %d)\n_%s_\n",
			item.Emoji, item.Name, sc.Qty, item.SellPrice, subtotal, maxQty, item.Description)
		rows = append(rows, []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("➖ %s", item.Name), "sell_dec_"+item.ID),
			tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("%d", sc.Qty), "sell_checkout"),
			tgbotapi.NewInlineKeyboardButtonData("➕", "sell_inc_"+item.ID),
			tgbotapi.NewInlineKeyboardButtonData("🗑️", "sell_rem_"+item.ID),
		})
	}

	caption += fmt.Sprintf("\n💰 *Total: +%d*🪙", totalGold)

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(
			fmt.Sprintf("✅ Vender tudo por +%d🪙", totalGold),
			"sell_confirm_all",
		),
	))
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("🔙 Continuar selecionando", "menu_sell"),
		tgbotapi.NewInlineKeyboardButtonData("❌ Cancelar", "sell_cancel"),
	))

	kb := tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}
	editPhoto(chatID, msgID, "shop", caption, &kb)
}

// handleSellConfirmAll confirma a venda de todos os itens no carrinho.
func handleSellConfirmAll(chatID int64, msgID int, userID int64) {
	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}
	cart := sellCart[userID]
	if cart == nil || len(cart.Items) == 0 {
		showSellMenu(chatID, msgID, userID)
		return
	}

	totalGained := 0
	summary := ""
	for _, sc := range cart.Items {
		item, ok := game.Items[sc.ItemID]
		if !ok {
			continue
		}
		count := database.GetItemCount(char.ID, sc.ItemID)
		qty := sc.Qty
		if qty > count {
			qty = count
		}
		if qty <= 0 {
			continue
		}
		database.UnequipItem(char.ID, sc.ItemID)
		database.RemoveItem(char.ID, sc.ItemID, qty)
		gained := item.SellPrice * qty
		totalGained += gained
		summary += fmt.Sprintf("%s *%s* ×%d → *+%d*🪙\n", item.Emoji, item.Name, qty, gained)
	}

	char.Gold += totalGained
	database.SaveCharacter(char)
	recalculateStats(char)
	database.SaveCharacter(char)
	delete(sellCart, userID)

	caption := fmt.Sprintf(
		"✅ *Venda realizada!*\n\n%s\n💰 Total recebido: *+%d* 🪙\n🪙 Ouro total: *%d*",
		summary, totalGained, char.Gold,
	)
	editPhoto(chatID, msgID, "shop", caption, bkp("menu_sell"))
}

// =============================================
// TRAVEL
// =============================================

func showTravelMenu(chatID int64, msgID int, userID int64) {
	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}
	current := game.Maps[char.CurrentMap]
	game.TickEnergy(char)
	database.SaveCharacter(char)
	eBar := game.EnergyBar(char.Energy, char.EnergyMax)

	caption := fmt.Sprintf("🗺️ *Viagem*\n\n📍 *%s %s*\n_%s_\nNível: %d-%d\n\n"+
		"⚡ %s *%d*/%d Energia\n*Custo por viagem:* %d ⚡\n\n*Para onde viajar?*",
		current.Emoji, current.Name, current.Description, current.MinLevel, current.MaxLevel,
		eBar, char.Energy, char.EnergyMax, game.EnergyTravelCost)

	var rows [][]tgbotapi.InlineKeyboardButton
	for _, destID := range current.ConnectsTo {
		dest := game.Maps[destID]
		warn := ""
		if char.Level < dest.MinLevel {
			warn = " ⚠️"
		}
		energyWarn := ""
		if char.Energy < game.EnergyTravelCost {
			energyWarn = " ⚡insuf."
		}
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("%s %s (Nv.%d-%d)%s%s", dest.Emoji, dest.Name, dest.MinLevel, dest.MaxLevel, warn, energyWarn),
				"travel_"+destID,
			),
		))
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		menuBackButton(userID, "menu_travel", "menu_main"),
	))
	kb := tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}
	editPhoto(chatID, msgID, "travel", caption, &kb)
}

func handleTravel(chatID int64, msgID int, userID int64, destID string) {
	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}
	dest, ok := game.Maps[destID]
	if !ok {
		return
	}
	current := game.Maps[char.CurrentMap]
	connected := false
	for _, c := range current.ConnectsTo {
		if c == destID {
			connected = true
			break
		}
	}
	if !connected {
		editPhoto(chatID, msgID, "travel", "❌ Destino não acessível daqui!", bkp("menu_travel"))
		return
	}

	game.TickEnergy(char)

	// Verificar energia para viajar
	if !game.ConsumeTravelEnergy(char) {
		eBar := game.EnergyBar(char.Energy, char.EnergyMax)
		editPhoto(chatID, msgID, "travel",
			fmt.Sprintf("❌ *Sem energia para viajar!*\n\n%s *%d*/%d ⚡\nPrecisa: *%d* ⚡\n\nAguarde a recarga ou use um item de energia.",
				eBar, char.Energy, char.EnergyMax, game.EnergyTravelCost),
			&tgbotapi.InlineKeyboardMarkup{InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
				{tgbotapi.NewInlineKeyboardButtonData("⚡ Energia", "menu_energy"),
					tgbotapi.NewInlineKeyboardButtonData("🏰 Menu", "menu_main")},
			}})
		database.SaveCharacter(char)
		return
	}

	warn := ""
	if char.Level < dest.MinLevel {
		warn = fmt.Sprintf("\n\n⚠️ _Nível recomendado: %d. Perigo elevado!_", dest.MinLevel)
	}
	char.CurrentMap = destID
	database.SaveCharacter(char)
	eBar := game.EnergyBar(char.Energy, char.EnergyMax)
	caption := fmt.Sprintf("🗺️ *Chegou em %s %s!*\n\n_%s_%s\n\n⚡ %s *%d*/%d",
		dest.Emoji, dest.Name, dest.Description, warn, eBar, char.Energy, char.EnergyMax)
	editPhoto(chatID, msgID, assets.MapImageKey(destID), caption, bkp("menu_travel"))
}

// =============================================
// EXPLORE
// =============================================

func showExploreMenu(chatID int64, msgID int, userID int64) {
	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}
	game.TickEnergyVIP(char, database.IsVIP(userID))
	database.SaveCharacter(char)

	current := game.Maps[char.CurrentMap]
	monsters := game.GetMonstersForMap(char.CurrentMap)
	if len(monsters) == 0 {
		editPhoto(chatID, msgID, assets.MapImageKey(char.CurrentMap),
			fmt.Sprintf("🌿 *%s*\n\nSem monstros aqui. Viaje para explorar!", current.Name),
			bkp("menu_travel"))
		return
	}

	// Energia necessária para explorar
	canExplore := char.Energy >= game.EnergyCombatEnter
	eBar := game.EnergyBar(char.Energy, char.EnergyMax)

	caption := fmt.Sprintf(
		"⚔️ *Explorar: %s %s*\n\n"+
			"❤️ *%d*/%d HP | 💙 *%d*/%d MP\n"+
			"⚡ %s *%d*/%d Energia\n\n"+
			"_A cada batalha você encontra um inimigo aleatório desta área._\n\n"+
			"*Monstros desta região:* %d tipos\n"+
			"*Custo por combate:* %d ⚡",
		current.Emoji, current.Name,
		char.HP, char.HPMax, char.MP, char.MPMax,
		eBar, char.Energy, char.EnergyMax,
		len(monsters), game.EnergyCombatEnter,
	)

	exploreLabel := fmt.Sprintf("⚔️ Explorar! (-%d⚡)", game.EnergyCombatEnter)
	if !canExplore {
		exploreLabel = "⚡ Sem energia para explorar"
		caption += "\n\n❌ *Energia insuficiente!* Aguarde a recarga ou use um item."
	}

	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(exploreLabel, "fight_random"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⚡ Energia", "menu_energy"),
			tgbotapi.NewInlineKeyboardButtonData("🏰 Menu", "menu_main"),
		),
	)
	editPhoto(chatID, msgID, assets.MapImageKey(char.CurrentMap), caption, &kb)
}

// =============================================
// COMBAT
// =============================================

func handleFightStart(chatID int64, msgID int, userID int64, _ string) {
	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}
	game.TickEnergy(char)

	// Random encounter — sorteia 1 dos 5 monstros da área
	monsters := game.GetMonstersForMap(char.CurrentMap)
	if len(monsters) == 0 {
		editPhoto(chatID, msgID, "menu", "Nenhum monstro aqui.", bkp("menu_explore"))
		return
	}
	// Sempre sorteia aleatoriamente, ignora monsterID passado
	monster := monsters[rand.Intn(len(monsters))]

	// Verificar energia antes de entrar no combate (custa 1 energia por encontro)
	if !game.ConsumeAttackEnergy(char) {
		eBar := game.EnergyBar(char.Energy, char.EnergyMax)
		editPhoto(chatID, msgID, assets.MapImageKey(char.CurrentMap),
			fmt.Sprintf("❌ *Sem energia para explorar!*\n\n%s *%d*/%d ⚡\nPrecisa: *%d* ⚡\n\nAguarde a recarga ou use um item de energia.",
				eBar, char.Energy, char.EnergyMax, game.EnergyCombatEnter),
			&tgbotapi.InlineKeyboardMarkup{InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
				{tgbotapi.NewInlineKeyboardButtonData("⚡ Energia", "menu_energy"),
					tgbotapi.NewInlineKeyboardButtonData("🏰 Menu", "menu_main")},
			}})
		database.SaveCharacter(char)
		return
	}

	char.State = "combat"
	char.CombatMonsterID = monster.ID
	char.CombatMonsterHP = monster.HP
	database.SaveCharacter(char)

	eBar := game.EnergyBar(char.Energy, char.EnergyMax)
	renderCombat(chatID, msgID, char, &monster,
		fmt.Sprintf("🎲 *Encontro aleatório!*\n%s *%s* apareceu!\n\n⚡ %s *%d*/%d\n\n",
			monster.Emoji, monster.Name, eBar, char.Energy, char.EnergyMax))
}

func handleCombatAttack(chatID int64, msgID int, userID int64) {
	char, _ := database.GetCharacter(userID)
	if char == nil || (char.State != "combat" && char.State != "dungeon_combat") {
		return
	}
	game.TickEnergy(char)
	recalculateStats(char) // garante EquipCABonus e EquipHitBonus atualizados
	monster := game.Monsters[char.CombatMonsterID]
	fx := getPVEEffects(char.ID)

	// DoT no monstro
	dotMonsterDmg, dotMonsterMsg := game.ApplyMonsterPoisonDoT(char, &monster)
	char.CombatMonsterHP -= dotMonsterDmg
	burnDmg, burnMsg := fx.ApplyEnemyDot()
	char.CombatMonsterHP -= burnDmg
	// DoT no player
	dotPlayerDmg, dotPlayerMsg := game.ApplyPlayerPoisonDoT(char)
	char.HP -= dotPlayerDmg
	if char.HP < 0 {
		char.HP = 0
	}

	monsterAdj := monster
	if pen := fx.EffectiveCAPenalty(); pen > 0 {
		monsterAdj.CA -= pen
		if monsterAdj.CA < 1 {
			monsterAdj.CA = 1
		}
	}
	if ap := fx.EffectiveAtkPenalty(); ap > 0 {
		monsterAdj.Attack -= ap
		if monsterAdj.Attack < 1 {
			monsterAdj.Attack = 1
		}
	}
	origCA := char.EquipCABonus
	char.EquipCABonus = origCA + fx.EffectiveCABonus()
	result := game.PlayerAttack(char, &monsterAdj)
	char.EquipCABonus = origCA
	// Passo das Sombras: próximo ataque crítico garantido.
	if result.PlayerDamage > 0 && !result.IsPlayerMiss && !result.IsCritical && fx.ConsumeForceCrit() {
		result.PlayerDamage *= 2
		result.IsCritical = true
		result.PlayerMessage += "\n👤 *Crítico garantido ativado!*"
	}
	// Nevasca / gelo: inimigo perde próximo ataque.
	if result.MonsterDamage > 0 && fx.ConsumeSkipEnemyAttack() {
		result.MonsterDamage = 0
		result.MonsterMessage += "\n🌪️ *Inimigo perdeu o ataque!*"
	}
	result.PlayerDamage = fx.ApplyOutgoingDamage(result.PlayerDamage, result.PlayerDamage > 0)
	result.MonsterDamage = fx.ApplyIncomingDamage(result.MonsterDamage)
	char.CombatMonsterHP -= result.PlayerDamage
	char.HP -= result.MonsterDamage
	if char.HP < 0 {
		char.HP = 0
	}
	fx.AdvanceTurn()
	combatLog := dotMonsterMsg + burnMsg + dotPlayerMsg + result.PlayerMessage + "\n" + result.MonsterMessage + "\n"
	processCombatResult(chatID, msgID, char, &monster, result, combatLog)
}

func handleCombatSkill(chatID int64, msgID int, userID int64, skillID string) {
	char, _ := database.GetCharacter(userID)
	if char == nil || (char.State != "combat" && char.State != "dungeon_combat") {
		return
	}
	game.TickEnergy(char)
	recalculateStats(char) // garante EquipCABonus e EquipHitBonus atualizados
	sk, ok := game.Skills[skillID]
	if !ok {
		return
	}
	if skillRequiresShield(skillID) && !hasEquippedShield(char.ID) {
		monster := game.Monsters[char.CombatMonsterID]
		renderCurrentCombatView(chatID, msgID, char, &monster, "🛡️ *Esta habilidade exige escudo equipado no slot Escudo.*\n")
		return
	}
	monster := game.Monsters[char.CombatMonsterID]
	fx := getPVEEffects(char.ID)

	// Verificar MP
	if char.MP < sk.MPCost {
		renderCurrentCombatView(chatID, msgID, char, &monster, "❌ *MP insuficiente!*\n")
		return
	}

	// DoT no monstro
	dotMonsterDmg, dotMonsterMsg := game.ApplyMonsterPoisonDoT(char, &monster)
	char.CombatMonsterHP -= dotMonsterDmg
	burnDmg, burnMsg := fx.ApplyEnemyDot()
	char.CombatMonsterHP -= burnDmg
	// DoT no player
	dotPlayerDmg, dotPlayerMsg := game.ApplyPlayerPoisonDoT(char)
	char.HP -= dotPlayerDmg
	if char.HP < 0 {
		char.HP = 0
	}

	monsterAdj := monster
	if pen := fx.EffectiveCAPenalty(); pen > 0 {
		monsterAdj.CA -= pen
		if monsterAdj.CA < 1 {
			monsterAdj.CA = 1
		}
	}
	if ap := fx.EffectiveAtkPenalty(); ap > 0 {
		monsterAdj.Attack -= ap
		if monsterAdj.Attack < 1 {
			monsterAdj.Attack = 1
		}
	}
	origCA := char.EquipCABonus
	char.EquipCABonus = origCA + fx.EffectiveCABonus()
	result := game.PlayerSkillAttack(char, &sk, &monsterAdj)
	char.EquipCABonus = origCA
	char.MP -= sk.MPCost
	// Faixas de crítico por skill (descrição).
	critMin := 21
	switch sk.ID {
	case "w_rampage", "m_meteor", "a_headshot":
		critMin = 18
	case "r_vital_strike":
		critMin = 17
	case "w_cleave":
		critMin = 19
	}
	if critMin <= 20 && result.PlayerRoll >= critMin && result.PlayerDamage > 0 && !result.IsPlayerMiss && !result.IsCritical {
		result.PlayerDamage *= 2
		result.IsCritical = true
		result.PlayerMessage += fmt.Sprintf("\n⭐ Crítico especial (%d-20) ativado.", critMin)
	}
	// Forma Fantasma: ignora CA (auto-acerto).
	if sk.ID == "r_phantom" && result.IsPlayerMiss {
		base := sk.Damage + int(float64(char.Dexterity)*0.3)
		if base < 1 {
			base = 1
		}
		result.IsPlayerMiss = false
		result.PlayerDamage = base
		result.PlayerMessage += "\n👻 *Forma Fantasma:* ataque atravessa defesa e acerta."
	}
	// Golpe Mortal: execução em alvo fraco.
	if sk.ID == "r_death_blow" && char.CombatMonsterHP <= monster.HP/4 && result.PlayerDamage > 0 && !result.IsPlayerMiss {
		result.PlayerDamage *= 3
		result.PlayerMessage += "\n💀 *Execução!* Inimigo abaixo de 25% HP: dano x3."
	}
	// Multi-hit simplificado por descrição.
	switch sk.ID {
	case "a_multishot":
		if result.PlayerDamage > 0 && !result.IsPlayerMiss {
			result.PlayerDamage *= 3
			result.PlayerMessage += "\n🌧️ *Chuva de Flechas:* 3 acertos."
		}
	case "a_volley":
		if result.PlayerDamage > 0 && !result.IsPlayerMiss {
			result.PlayerDamage *= 5
			result.PlayerMessage += "\n⛈️ *Saraivada:* 5 acertos."
		}
	case "a_quick_shot":
		if result.PlayerDamage > 0 && !result.IsPlayerMiss && rand.Intn(100) < 30 {
			result.PlayerDamage *= 2
			result.PlayerMessage += "\n🏹 *Tiro Rápido:* segundo disparo ativado!"
		}
	}
	if sk.ID == "m_arcane_burst" && fx.EffectiveCABonus() > 0 && result.PlayerDamage > 0 {
		result.PlayerDamage *= 2
		result.PlayerMessage += "\n💫 *Escudo Arcano ativo:* dano dobrado!"
	}
	// Passo das Sombras: crítico garantido para este ataque.
	if result.PlayerDamage > 0 && !result.IsPlayerMiss && !result.IsCritical && fx.ConsumeForceCrit() {
		result.PlayerDamage *= 2
		result.IsCritical = true
		result.PlayerMessage += "\n👤 *Crítico garantido ativado!*"
	}
	// Habilidade que remove ataque inimigo.
	if sk.ID == "m_blizzard" && result.MonsterDamage > 0 {
		result.MonsterDamage = 0
		result.MonsterMessage += "\n🌪️ *Nevasca:* inimigo perdeu o ataque!"
		fx.ConsumeSkipEnemyAttack()
	}
	if result.MonsterDamage > 0 && fx.ConsumeSkipEnemyAttack() {
		result.MonsterDamage = 0
		result.MonsterMessage += "\n🌪️ *Inimigo perdeu o ataque!*"
	}
	result.PlayerDamage = fx.ApplyOutgoingDamage(result.PlayerDamage, result.PlayerDamage > 0)
	result.MonsterDamage = fx.ApplyIncomingDamage(result.MonsterDamage)
	char.CombatMonsterHP -= result.PlayerDamage
	char.HP -= result.MonsterDamage
	if char.HP < 0 {
		char.HP = 0
	}
	// Aplica veneno ao monstro se a skill envenenou
	if result.AppliesPoison && !result.IsPlayerMiss {
		if result.PoisonTurns > char.CombatMonsterPoisonTurns || result.PoisonDmg > char.CombatMonsterPoisonDmg {
			char.CombatMonsterPoisonTurns = result.PoisonTurns
			char.CombatMonsterPoisonDmg = result.PoisonDmg
		}
	}
	effectMsg := ""
	if !result.IsPlayerMiss {
		if sk.ID == "w_blood_rage" && char.HPMax > 0 && (char.HP*100/char.HPMax) >= 30 {
			effectMsg = "🩸 *Fúria Sangrenta* requer HP abaixo de 30%.\n"
		} else {
			effectMsg = formatEffectMsg(applySkillEffectsPVE(sk.ID, fx))
		}
	}
	fx.AdvanceTurn()
	combatLog := dotMonsterMsg + burnMsg + dotPlayerMsg + result.PlayerMessage + "\n" + result.MonsterMessage + "\n" + effectMsg
	processCombatResult(chatID, msgID, char, &monster, result, combatLog)
}

func handleCombatFlee(chatID int64, msgID int, userID int64) {
	char, _ := database.GetCharacter(userID)
	if char == nil || (char.State != "combat" && char.State != "dungeon_combat") {
		return
	}
	monster := game.Monsters[char.CombatMonsterID]

	// In dungeon, flee is not allowed — player must abandon the whole run
	if char.State == "dungeon_combat" {
		renderCombat(chatID, msgID, char, &monster, "❌ *Não é possível fugir dentro de uma masmorra!*\nUse o botão *Abandonar* para sair da masmorra.\n")
		return
	}

	if game.TryFlee(char, &monster) {
		char.State = "idle"
		char.CombatMonsterID = ""
		char.CombatMonsterHP = 0
		char.CombatMonsterPoisonTurns = 0
		char.CombatMonsterPoisonDmg = 0
		resetPVEEffects(char.ID)
		database.SaveCharacter(char)
		database.LogCombat(char.ID, monster.ID, "flee", 0, 0)
		editPhoto(chatID, msgID, assets.MapImageKey(char.CurrentMap),
			fmt.Sprintf("💨 Fugiu de *%s %s*!", monster.Emoji, monster.Name), bkp("menu_explore"))
	} else {
		dmg := rand.Intn(monster.Attack/2+1) + monster.Attack/2
		char.HP -= dmg
		if char.HP < 0 {
			char.HP = 0
		}
		if char.HP == 0 {
			handlePlayerDeath(chatID, msgID, char, &monster)
			return
		}
		database.SaveCharacter(char)
		renderCombat(chatID, msgID, char, &monster, fmt.Sprintf("❌ *Fuga falhou!* -%d HP\n", dmg))
	}
}

func showCombatItemMenu(chatID int64, msgID int, userID int64) {
	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}
	monster := game.Monsters[char.CombatMonsterID]
	items, _ := database.GetInventory(char.ID)

	caption := "🎒 *Usar Item no Combate*\n\n"
	var rows [][]tgbotapi.InlineKeyboardButton
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
				"combat_item_"+item.ID,
			),
		))
	}
	if len(rows) == 0 {
		caption += "_Sem consumíveis no inventário._"
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("⬅️ Voltar ao Combate", "fight_"+monster.ID),
	))
	if char.State == "dungeon_combat" {
		rows[len(rows)-1] = tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⬅️ Voltar ao Combate", "dungeon_back_to_combat"),
		)
	}
	kb := tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}
	editPhoto(chatID, msgID, assets.MonsterImageKey(monster.ID), caption, &kb)
}

func handleCombatItem(chatID int64, msgID int, userID int64, itemID string) {
	char, _ := database.GetCharacter(userID)
	if char == nil || (char.State != "combat" && char.State != "dungeon_combat") {
		return
	}
	item, ok := game.Items[itemID]
	if !ok {
		return
	}
	count := database.GetItemCount(char.ID, itemID)
	if count <= 0 {
		monster := game.Monsters[char.CombatMonsterID]
		renderCombat(chatID, msgID, char, &monster, "❌ Item não encontrado!\n")
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
	if item.CurePoison {
		if char.PoisonTurns > 0 {
			char.PoisonTurns = 0
			char.PoisonDmg = 0
			effects += "☠️ Envenenamento curado! "
		} else {
			effects += "❓ Não estava envenenado. "
		}
	}

	monster := game.Monsters[char.CombatMonsterID]
	if err := database.RemoveItem(char.ID, itemID, 1); err != nil {
		renderCurrentCombatView(chatID, msgID, char, &monster, "❌ Erro ao consumir item. Tente novamente.\n")
		return
	}
	if err := database.SaveCharacter(char); err != nil {
		renderCurrentCombatView(chatID, msgID, char, &monster, "❌ Erro ao salvar personagem. Tente novamente.\n")
		return
	}

	renderCurrentCombatView(chatID, msgID, char, &monster, fmt.Sprintf("%s *%s* usada! %s\n", item.Emoji, item.Name, effects))
}

func processCombatResult(chatID int64, msgID int, char *models.Character, monster *models.Monster, _ game.CombatResult, combatLog string) {
	if char.CombatMonsterHP <= 0 {
		if char.State == "dungeon_combat" {
			handleDungeonMonsterDeath(chatID, msgID, char, monster, combatLog)
		} else {
			handleMonsterDeath(chatID, msgID, char, monster, combatLog)
		}
		return
	}
	if char.HP <= 0 {
		if char.State == "dungeon_combat" {
			handleDungeonPlayerDeath(chatID, msgID, char, monster)
		} else {
			handlePlayerDeath(chatID, msgID, char, monster)
		}
		return
	}
	database.SaveCharacter(char)
	if char.State == "dungeon_combat" {
		run, _ := database.GetActiveDungeonRun(char.ID)
		if run != nil {
			d := game.Dungeons[run.DungeonID]
			renderDungeonCombat(chatID, msgID, char, monster, run, &d, combatLog)
			return
		}
	}
	renderCombat(chatID, msgID, char, monster, combatLog)
}

func handleMonsterDeath(chatID int64, msgID int, char *models.Character, monster *models.Monster, combatLog string) {
	xp := game.CalculateXPGain(char, monster)
	goldGain := monster.GoldReward + rand.Intn(monster.GoldReward/2+1)
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
	if monster.DiamondChance > 0 && rand.Intn(100) < monster.DiamondChance {
		diamondGained = 1
		char.Diamonds++
		database.LogDiamond(char.ID, 1, "monster_drop_"+monster.ID)
	}

	// ── Item drop system ────────────────────────────────────
	drops := game.RollDrops(monster, char.Level)
	var dropLines []string
	for _, drop := range drops {
		item, ok := game.Items[drop.ItemID]
		if !ok {
			continue
		}
		if item.Type == "chest" {
			chestGold, chestItemID := game.OpenChest(drop.ItemID, char.Level)
			char.Gold += chestGold
			if chestGold > 0 {
				dropLines = append(dropLines, fmt.Sprintf("%s %s aberto! *+%d* 🪙", item.Emoji, item.Name, chestGold))
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
			dropLines = append(dropLines, fmt.Sprintf("🎁 %s%s *%s* (%s)", qtyStr, item.Emoji, item.Name, item.Rarity.Name()))
		}
	}
	// ────────────────────────────────────────────────────────

	lvlUp := game.CheckLevelUp(char)
	if lvlUp != nil {
		game.ApplyLevelUp(char, lvlUp)
		// Update energy max on level up
		char.EnergyMax = game.MaxEnergy(char.Level)
	}
	database.SaveCharacter(char)
	database.LogCombat(char.ID, monster.ID, "win", xp, goldGain)

	caption := truncateCombatLog(combatLog, 3) + fmt.Sprintf("\n🏆 *%s %s derrotado!*\n\n✨ +%d XP | 🪙 +%d | ❤️ %d/%d | 💙 %d/%d",
		monster.Emoji, monster.Name, xp, goldGain, char.HP, char.HPMax, char.MP, char.MPMax)
	if diamondGained > 0 {
		caption += "\n💎 *+1 Diamante! (drop raro!)*"
	}
	if len(dropLines) > 0 {
		shownDrops := dropLines
		extra := 0
		if len(shownDrops) > 4 {
			extra = len(shownDrops) - 4
			shownDrops = shownDrops[:4]
		}
		caption += "\n\n🎁 *Drops:*\n" + strings.Join(shownDrops, "\n")
		if extra > 0 {
			caption += fmt.Sprintf("\n_...+%d item(s)_", extra)
		}
	}
	if lvlUp != nil {
		caption += fmt.Sprintf("\n\n🎉 *NÍVEL UP! Nv.%d!*\n+%d HP | +%d MP | ⚡ Energia máx.: %d | 🌟 +1 ponto",
			lvlUp.NewLevel, lvlUp.HPGained, lvlUp.MPGained, char.EnergyMax)
	}

	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⚔️ Explorar Mais", "menu_explore"),
			tgbotapi.NewInlineKeyboardButtonData("🏰 Menu", "menu_main"),
		),
	)
	editPhoto(chatID, msgID, "victory", caption, &kb)
}

func handlePlayerDeath(chatID int64, msgID int, char *models.Character, monster *models.Monster) {
	// Check revive token
	if database.GetItemCount(char.ID, "revive_token") > 0 {
		database.RemoveItem(char.ID, "revive_token", 1)
		char.HP = char.HPMax
		char.MP = char.MPMax
		char.State = "idle"
		char.CombatMonsterID = ""
		char.CombatMonsterHP = 0
		char.CombatMonsterPoisonTurns = 0
		char.CombatMonsterPoisonDmg = 0
		char.PoisonTurns = 0
		char.PoisonDmg = 0
		resetPVEEffects(char.ID)
		database.SaveCharacter(char)
		caption := fmt.Sprintf("🔮 *Token de Reviver ativado!*\n\nVocê quase morreu para *%s %s*, mas o token te salvou!\n\n❤️ HP restaurado: *%d*/%d\n💙 MP restaurado: *%d*/%d",
			monster.Emoji, monster.Name, char.HP, char.HPMax, char.MP, char.MPMax)
		kb := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏰 Menu", "menu_main"),
		))
		editPhoto(chatID, msgID, "victory", caption, &kb)
		return
	}

	xpLost, goldLost := game.ApplyDeathPenalty(char)
	resetPVEEffects(char.ID)
	database.SaveCharacter(char)
	database.LogCombat(char.ID, monster.ID, "lose", 0, 0)

	caption := fmt.Sprintf("💀 *Derrotado por %s %s!*\n\nAcordou na Vila de Trifort...\n\n❤️ %d/%d HP | 💙 %d/%d MP\n🪙 -%d ouro | ✨ -%d XP\n\n💡 _Dica: Use ⚡ Energia para se curar, ou compre um Token de Reviver na Loja de Diamantes!_",
		monster.Emoji, monster.Name, char.HP, char.HPMax, char.MP, char.MPMax, goldLost, xpLost)
	kb := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("🏰 Menu", "menu_main"),
	))
	editPhoto(chatID, msgID, "defeat", caption, &kb)
}

// handleDungeonPlayerDeath handles death inside a dungeon combat.
// The run is terminated and the player suffers a reduced death penalty.
func handleDungeonPlayerDeath(chatID int64, msgID int, char *models.Character, monster *models.Monster) {
	// Check revive token first
	if database.GetItemCount(char.ID, "revive_token") > 0 {
		database.RemoveItem(char.ID, "revive_token", 1)
		char.HP = char.HPMax
		char.MP = char.MPMax
		char.State = "idle"
		char.CombatMonsterID = ""
		char.CombatMonsterHP = 0
		char.CombatMonsterPoisonTurns = 0
		char.CombatMonsterPoisonDmg = 0
		char.PoisonTurns = 0
		char.PoisonDmg = 0
		resetPVEEffects(char.ID)
		database.SaveCharacter(char)
		caption := fmt.Sprintf("🔮 *Token de Reviver ativado!*\n\nVocê quase morreu para *%s %s* na masmorra, mas o token te salvou!\n\n❤️ HP restaurado: *%d*/%d\n💙 MP restaurado: *%d*/%d\n\n_Você permanece no andar atual._",
			monster.Emoji, monster.Name, char.HP, char.HPMax, char.MP, char.MPMax)
		kb := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("▶️ Continuar Masmorra", "dungeon_continue"),
			tgbotapi.NewInlineKeyboardButtonData("🏰 Menu", "menu_main"),
		))
		editPhoto(chatID, msgID, "victory", caption, &kb)
		return
	}

	// Terminate the dungeon run
	run, _ := database.GetActiveDungeonRun(char.ID)
	floorReached := 0
	dungeonName := "Masmorra"
	if run != nil {
		floorReached = run.Floor - 1
		d := game.Dungeons[run.DungeonID]
		dungeonName = d.Emoji + " " + d.Name
		database.FinishDungeonRun(run.ID, "abandoned")
		database.UpdateDungeonBest(char.ID, run.DungeonID, floorReached, false)
	}

	// Reduced gold penalty in dungeon (10% instead of normal)
	goldLost := char.Gold / 10
	char.Gold -= goldLost
	char.HP = char.HPMax / 4
	char.State = "idle"
	char.CombatMonsterID = ""
	char.CombatMonsterHP = 0
	char.CombatMonsterPoisonTurns = 0
	char.CombatMonsterPoisonDmg = 0
	char.PoisonTurns = 0
	char.PoisonDmg = 0
	resetPVEEffects(char.ID)
	database.SaveCharacter(char)
	database.LogCombat(char.ID, monster.ID, "lose", 0, 0)

	caption := fmt.Sprintf(
		"💀 *Derrotado por %s %s na masmorra!*\n\n%s abandonada no andar *%d*.\n\n❤️ %d/%d HP | 🪙 -%d ouro\n\n💡 _Compre um Token de Reviver para não perder o progresso!_",
		monster.Emoji, monster.Name, dungeonName, floorReached, char.HP, char.HPMax, goldLost)
	kb := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("🏚️ Masmorras", "menu_dungeon"),
		tgbotapi.NewInlineKeyboardButtonData("🏰 Menu", "menu_main"),
	))
	editPhoto(chatID, msgID, "defeat", caption, &kb)
}

// truncateCombatLog keeps only the last maxLines lines of combat log
// to prevent Telegram caption overflow (limit: 1024 chars for photos).
func truncateCombatLog(log string, maxLines int) string {
	lines := strings.Split(strings.TrimRight(log, "\n"), "\n")
	if len(lines) <= maxLines {
		return log
	}
	return strings.Join(lines[len(lines)-maxLines:], "\n") + "\n"
}

func renderCombat(chatID int64, msgID int, char *models.Character, monster *models.Monster, combatLog string) {
	pHP := 0
	if char.HPMax > 0 {
		ratio := float64(char.HP) / float64(char.HPMax)
		ratio = math.Max(0, math.Min(1, ratio))
		pHP = int(math.Round(ratio * 8))
	}
	mHP := 0
	if monster.HP > 0 {
		ratio := float64(char.CombatMonsterHP) / float64(monster.HP)
		ratio = math.Max(0, math.Min(1, ratio))
		mHP = int(math.Round(ratio * 8))
	}
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

	caption := fmt.Sprintf(
		"⚔️ *COMBATE!*\n\n%s *%s* Nv.*%d*\n%s %d/%d HP\n\n%s *%s* Nv.*%d*\n%s %d/%d HP | 💙 %d/%d MP | ⚡ %d\n%s\n\n━━━━━━━━━━━━\n%s",
		monster.Emoji, monster.Name, monster.Level, mBar, char.CombatMonsterHP, monster.HP,
		game.Races[char.Race].Emoji, char.Name, char.Level, pBar, char.HP, char.HPMax, char.MP, char.MPMax, char.Energy,
		eBar, truncateCombatLog(combatLog, 4),
	)

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
		tgbotapi.NewInlineKeyboardButtonData("💨 Fugir", "combat_flee"),
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

func renderCurrentCombatView(chatID int64, msgID int, char *models.Character, monster *models.Monster, combatLog string) {
	if char != nil && char.State == "dungeon_combat" {
		run, _ := database.GetActiveDungeonRun(char.ID)
		if run != nil {
			d := game.Dungeons[run.DungeonID]
			renderDungeonCombat(chatID, msgID, char, monster, run, &d, combatLog)
			return
		}
	}
	renderCombat(chatID, msgID, char, monster, combatLog)
}

// =============================================
// INVENTORY
// =============================================

func showEquipScreen(chatID int64, msgID int, userID int64) {
	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}
	sanitizeEquippedConflicts(char.ID)
	recalculateStats(char)

	equipped, _ := database.GetEquippedItems(char.ID)
	// Build slot map: slot -> itemID
	slotMap := map[string]string{}
	for _, e := range equipped {
		if e.Slot != "" {
			slotMap[e.Slot] = e.ItemID
		} else {
			// legacy: infer slot from item definition
			item := game.Items[e.ItemID]
			if item.Slot != "" {
				slotMap[item.Slot] = e.ItemID
			}
		}
	}

	slotLabel := func(slot, emoji, name string) string {
		if id, ok := slotMap[slot]; ok {
			it := game.Items[id]
			return fmt.Sprintf("%s %-14s %s%s", emoji, name+":", it.Emoji, it.Name)
		}
		return fmt.Sprintf("%s %-14s _vazio_", emoji, name+":")
	}

	ca := game.CharacterCA(char.Class,
		game.DefensiveAttr(char.Class, char.Constitution, char.Dexterity, char.Intelligence),
		char.EquipCABonus)

	caption := fmt.Sprintf(
		"⚔️ *Equipamentos de %s*\n"+
			"%s %s\n\n"+
			"```\n"+
			"%s\n"+
			"%s\n"+
			"%s\n"+
			"%s\n"+
			"%s\n"+
			"%s\n"+
			"──────────\n"+
			"%s\n"+
			"%s\n"+
			"%s\n"+
			"```\n\n"+
			"❤️ HP:*%d/%d* ⚔️ Atq:*%d* 🔮 MAtq:*%d*\n"+
			"🛡️ CA:*%d* Def:*%d* 💙 MP:*%d/%d* 💨 Spd:*%d*",
		char.Name,
		game.Races[char.Race].Emoji, game.Classes[char.Class].Emoji+game.Classes[char.Class].Name,
		slotLabel("weapon", "⚔️", "Arma"),
		slotLabel("head", "⛑️", "Cabeça"),
		slotLabel("chest", "🛡️", "Peito"),
		slotLabel("hands", "🧤", "Mãos"),
		slotLabel("legs", "🦵", "Pernas"),
		slotLabel("feet", "👢", "Pés"),
		slotLabel("offhand", "🛡️", "Escudo"),
		slotLabel("accessory1", "💍", "Anel"),
		slotLabel("accessory2", "📿", "Colar"),
		char.HP, char.HPMax, char.Attack, char.MagicAttack,
		ca, char.Defense, char.MP, char.MPMax, char.Speed,
	)

	rows := [][]tgbotapi.InlineKeyboardButton{
		{
			tgbotapi.NewInlineKeyboardButtonData("⚔️ Arma", "equip_slot_weapon"),
			tgbotapi.NewInlineKeyboardButtonData("⛑️ Cabeça", "equip_slot_head"),
		},
		{
			tgbotapi.NewInlineKeyboardButtonData("🛡️ Peito", "equip_slot_chest"),
			tgbotapi.NewInlineKeyboardButtonData("🧤 Mãos", "equip_slot_hands"),
		},
		{
			tgbotapi.NewInlineKeyboardButtonData("🦵 Pernas", "equip_slot_legs"),
			tgbotapi.NewInlineKeyboardButtonData("👢 Pés", "equip_slot_feet"),
		},
		{
			tgbotapi.NewInlineKeyboardButtonData("🛡️ Escudo", "equip_slot_offhand"),
		},
		{
			tgbotapi.NewInlineKeyboardButtonData("💍 Anel", "equip_slot_accessory1"),
			tgbotapi.NewInlineKeyboardButtonData("📿 Colar", "equip_slot_accessory2"),
		},
		{tgbotapi.NewInlineKeyboardButtonData("🎒 Inventário", "menu_inventory")},
		{tgbotapi.NewInlineKeyboardButtonData("🏰 Menu", "menu_main")},
	}
	kb := tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}
	editPhoto(chatID, msgID, "inventory", caption, &kb)
}

func showSlotItems(chatID int64, msgID int, userID int64, slot string) {
	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}
	sanitizeEquippedConflicts(char.ID)

	slotEmojis := map[string]string{
		"weapon": "⚔️", "head": "⛑️", "chest": "🛡️",
		"hands": "🧤", "legs": "🦵", "feet": "👢", "offhand": "🛡️",
		"accessory1": "💍", "accessory2": "📿",
	}

	inv, _ := database.GetInventory(char.ID)
	// Current equipped in this slot
	var currentName string
	for _, e := range inv {
		item, ok := game.Items[e.ItemID]
		if !ok {
			continue
		}
		if isItemEquippedInSlot(e, item, slot) {
			currentName = item.Name
		}
	}

	caption := fmt.Sprintf("%s *Slot: %s*\n", slotEmojis[slot], slotDisplayName(slot))
	if currentName != "" {
		caption += fmt.Sprintf("✅ Equipado: *%s*\n", currentName)
	} else {
		caption += "_Slot vazio_\n"
	}
	caption += "\n*Itens disponíveis:*\n"

	var rows [][]tgbotapi.InlineKeyboardButton
	count := 0

	for _, inv := range inv {
		item, ok := game.Items[inv.ItemID]
		if !ok {
			continue
		}
		if !itemFitsSlot(item, slot) {
			continue
		}
		if item.ClassReq != "" && item.ClassReq != char.Class {
			continue
		}
		if item.MinLevel > char.Level {
			continue
		}

		isEquipped := isItemEquippedInSlot(inv, item, slot)
		equipped := ""
		if isEquipped {
			equipped = " ✅"
		}

		stats := itemStatSummary(item)
		caption += fmt.Sprintf("%s *%s*%s\n%s %s\n", item.Emoji, item.Name, equipped, item.Rarity.Emoji(), stats)
		count++

		label := fmt.Sprintf("%s %s %s", item.Rarity.Emoji(), item.Emoji, item.Name)
		if isEquipped {
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("❌ Desequipar", "equip_remove_"+slot),
				tgbotapi.NewInlineKeyboardButtonData("🔄 "+item.Name, "equip_item_"+item.ID+"_"+slot),
			))
		} else {
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(label, "equip_item_"+item.ID+"_"+slot),
			))
		}
	}

	if count == 0 {
		caption += fmt.Sprintf("_Nenhum item para este slot. Encontre %s em combate ou na loja!_", slotDisplayName(slot))
	}

	rows = append(rows,
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("⬅️ Equipamentos", "menu_equip")),
	)
	kb := tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}
	editPhoto(chatID, msgID, "inventory", caption, &kb)
}

func itemFitsSlot(item models.Item, slot string) bool {
	if item.Slot != "" {
		return item.Slot == slot
	}
	if item.Type == "weapon" {
		return slot == "weapon"
	}
	if item.Type == "accessory" {
		return slot == "accessory1" || slot == "accessory2"
	}
	return false
}

func isItemEquippedInSlot(inv models.InventoryItem, item models.Item, slot string) bool {
	if !inv.Equipped {
		return false
	}
	return normalizeEquippedSlot(inv, item) == slot
}

// itemStatSummary formats a compact stat line for an item
func itemStatSummary(item models.Item) string {
	var parts []string
	if item.AttackBonus > 0 {
		parts = append(parts, fmt.Sprintf("⚔️+%d", item.AttackBonus))
	}
	if item.MagicAtkBonus > 0 {
		parts = append(parts, fmt.Sprintf("🔮+%d", item.MagicAtkBonus))
	}
	if item.DefenseBonus > 0 {
		parts = append(parts, fmt.Sprintf("🛡️+%d", item.DefenseBonus))
	}
	if item.MagicDefBonus > 0 {
		parts = append(parts, fmt.Sprintf("💜+%d", item.MagicDefBonus))
	}
	if item.CABonus > 0 {
		parts = append(parts, fmt.Sprintf("CA+%d", item.CABonus))
	}
	if item.HitBonus > 0 {
		parts = append(parts, fmt.Sprintf("🎯+%d", item.HitBonus))
	}
	if item.SpeedBonus > 0 {
		parts = append(parts, fmt.Sprintf("💨+%d", item.SpeedBonus))
	}
	if item.HPBonus > 0 {
		parts = append(parts, fmt.Sprintf("❤️+%d", item.HPBonus))
	}
	if item.MPBonus > 0 {
		parts = append(parts, fmt.Sprintf("💙+%d", item.MPBonus))
	}
	if len(parts) == 0 {
		return "_sem bônus_"
	}
	return strings.Join(parts, " ")
}

func handleEquipItem(chatID int64, msgID int, userID int64, itemID, slot string) {
	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}

	item, ok := game.Items[itemID]
	if !ok {
		return
	}

	// Class restriction
	if item.ClassReq != "" && item.ClassReq != char.Class {
		editPhoto(chatID, msgID, "inventory",
			fmt.Sprintf("❌ Somente *%s* pode equipar *%s*!", game.Classes[item.ClassReq].Name, item.Name),
			bkp("menu_equip"))
		return
	}
	// Level restriction
	if item.MinLevel > char.Level {
		editPhoto(chatID, msgID, "inventory",
			fmt.Sprintf("❌ Precisa de nível *%d* para equipar *%s*! (você é nível *%d*)", item.MinLevel, item.Name, char.Level),
			bkp("menu_equip"))
		return
	}

	sanitizeEquippedConflicts(char.ID)
	database.EquipItemSlot(char.ID, itemID, item.Type, slot)
	recalculateStats(char)
	database.SaveCharacter(char)

	ca := game.CharacterCA(char.Class,
		game.DefensiveAttr(char.Class, char.Constitution, char.Dexterity, char.Intelligence),
		char.EquipCABonus)

	stats := itemStatSummary(item)
	caption := fmt.Sprintf(
		"✅ *%s %s* equipado no slot *%s*!\n\n%s\n\n"+
			"📊 *Stats atuais:*\n"+
			"⚔️ Ataque: *%d*  🔮 Mágico: *%d*\n"+
			"🛡️ Defesa: *%d*  CA: *%d*\n"+
			"❤️ HP: *%d*/%d  💙 MP: *%d*/%d",
		item.Emoji, item.Name, slotDisplayName(slot),
		stats,
		char.Attack, char.MagicAttack,
		char.Defense, ca,
		char.HP, char.HPMax, char.MP, char.MPMax,
	)
	editPhoto(chatID, msgID, assets.ItemTypeImageKey(item.Type), caption,
		&tgbotapi.InlineKeyboardMarkup{InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
			{tgbotapi.NewInlineKeyboardButtonData("⚔️ Ver Equipamentos", "menu_equip")},
			{tgbotapi.NewInlineKeyboardButtonData("🎒 Inventário", "menu_inventory")},
		}})
}

func handleUnequipSlot(chatID int64, msgID int, userID int64, slot string) {
	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}
	if slot == "" {
		showSlotItems(chatID, msgID, userID, slot)
		return
	}
	unequipByResolvedSlot(char.ID, slot)
	recalculateStats(char)
	database.SaveCharacter(char)
	showSlotItems(chatID, msgID, userID, slot)
}

// unequipByResolvedSlot remove itens equipados considerando slot persistido e
// fallback legado (slot deduzido por tipo/item).
func unequipByResolvedSlot(charID int, slot string) {
	// Primeiro limpa por slot persistido.
	_ = database.UnequipSlot(charID, slot)

	// Depois limpa estados legados sem slot salvo.
	inv, _ := database.GetInventory(charID)
	for _, invItem := range inv {
		if !invItem.Equipped {
			continue
		}
		item, ok := game.Items[invItem.ItemID]
		if !ok {
			continue
		}
		if normalizeEquippedSlot(invItem, item) == slot {
			_ = database.UnequipItem(charID, invItem.ItemID)
		}
	}
}

func slotDisplayName(slot string) string {
	names := map[string]string{
		"weapon": "Arma", "head": "Cabeça", "chest": "Peito",
		"hands": "Mãos", "legs": "Pernas", "feet": "Pés", "offhand": "Escudo",
		"accessory1": "Anel", "accessory2": "Colar",
	}
	if n, ok := names[slot]; ok {
		return n
	}
	return slot
}

func showInventoryHome(chatID int64, msgID int, userID int64) {
	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}

	caption := fmt.Sprintf(
		"🎒 *Inventário de %s*\n🪙 *%d* | 💎 *%d*\n\nSelecione uma categoria:",
		char.Name, char.Gold, char.Diamonds,
	)

	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🧪 Consumíveis", "inv_tab_consumable"),
			tgbotapi.NewInlineKeyboardButtonData("⚔️ Armas", "inv_tab_weapon"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🛡️ Armaduras", "inv_tab_armor"),
			tgbotapi.NewInlineKeyboardButtonData("💍 Acessórios", "inv_tab_accessory"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📋 Todos", "inv_tab_all"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⚔️ Equipamentos", "menu_equip"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏰 Menu", "menu_main"),
		),
	)
	editPhoto(chatID, msgID, "inventory", caption, &kb)
}

func showInventory(chatID int64, msgID int, userID int64, filterType string) {
	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}
	items, _ := database.GetInventory(char.ID)

	tabNames := map[string]string{
		"all":        "Todos",
		"consumable": "Consumíveis",
		"weapon":     "Armas",
		"armor":      "Armaduras",
		"accessory":  "Acessórios",
	}
	tabName := tabNames[filterType]
	if tabName == "" {
		tabName = "Todos"
		filterType = "all"
	}
	caption := fmt.Sprintf(
		"🎒 *Inventário — %s*\n👤 %s\n🪙 *%d* | 💎 *%d*\n\nEscolha um item:",
		tabName, char.Name, char.Gold, char.Diamonds,
	)
	var rows [][]tgbotapi.InlineKeyboardButton
	count := 0

	for _, inv := range items {
		item, ok := game.Items[inv.ItemID]
		if !ok {
			continue
		}
		if filterType != "all" && item.Type != filterType {
			continue
		}

		label := fmt.Sprintf("%s %s", item.Emoji, item.Name)
		if inv.Equipped {
			label += " ✅"
		}
		if inv.Quantity > 1 {
			label += fmt.Sprintf(" x%d", inv.Quantity)
		}
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(label, "inv_item_"+filterType+"_"+item.ID),
		))
		count++
	}
	if count == 0 {
		caption += "\n_Nenhum item nesta categoria._"
	}

	// Tabs
	rows = append(rows, []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("🧪 Consumíveis", "inv_tab_consumable"),
		tgbotapi.NewInlineKeyboardButtonData("⚔️ Armas", "inv_tab_weapon"),
	})
	rows = append(rows, []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("🛡️ Armaduras", "inv_tab_armor"),
		tgbotapi.NewInlineKeyboardButtonData("💍 Acessórios", "inv_tab_accessory"),
	})
	rows = append(rows, []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("📋 Todos", "inv_tab_all"),
		tgbotapi.NewInlineKeyboardButtonData("⬅️ Voltar", "menu_inventory"),
	})
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("⚔️ Equipamentos", "menu_equip"),
	))
	kb := tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}
	editPhoto(chatID, msgID, "inventory", caption, &kb)
}

func showInventoryItem(chatID int64, msgID int, userID int64, tab string, itemID string) {
	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}
	item, ok := game.Items[itemID]
	if !ok {
		showInventory(chatID, msgID, userID, tab)
		return
	}

	inv, _ := database.GetInventory(char.ID)
	var invItem *models.InventoryItem
	for i := range inv {
		if inv[i].ItemID == itemID {
			invItem = &inv[i]
			break
		}
	}
	if invItem == nil || invItem.Quantity <= 0 {
		showInventory(chatID, msgID, userID, tab)
		return
	}

	caption := fmt.Sprintf(
		"🎒 *Item do Inventário*\n\n%s *%s*\nTipo: *%s*\nQuantidade: *%d*\n\n_%s_",
		item.Emoji, item.Name, item.Type, invItem.Quantity, item.Description,
	)
	if invItem.Equipped {
		caption += "\n\n✅ Equipado"
	}
	if item.Type != "consumable" && item.Type != "chest" {
		stats := itemStatSummary(item)
		if stats != "" {
			caption += fmt.Sprintf("\n📊 %s", stats)
		}
	}

	var rows [][]tgbotapi.InlineKeyboardButton
	if item.Type == "consumable" {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🧪 Usar", "inv_use_"+item.ID),
		))
	} else if item.Type != "chest" {
		if invItem.Equipped {
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("📤 Desequipar", "inv_unequip_"+tab+"_"+item.ID),
			))
		} else {
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("✅ Equipar", "inv_equip_"+item.ID),
			))
		}
	}

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("⬅️ Voltar", "inv_back_"+tab),
	))
	kb := tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}
	editPhoto(chatID, msgID, assets.ItemTypeImageKey(item.Type), caption, &kb)
}

func handleInventoryUnequip(chatID int64, msgID int, userID int64, tab string, itemID string) {
	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}
	_ = database.UnequipItem(char.ID, itemID)
	recalculateStats(char)
	_ = database.SaveCharacter(char)
	showInventoryItem(chatID, msgID, userID, tab, itemID)
}

func handleInventoryUse(chatID int64, msgID int, userID int64, itemID string) {
	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}
	item, ok := game.Items[itemID]
	if !ok {
		return
	}
	if database.GetItemCount(char.ID, itemID) <= 0 {
		editPhoto(chatID, msgID, "inventory", "❌ Item não encontrado!", bkp("menu_inventory"))
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
	// Tomo do Esquecimento: reseta todas as habilidades e devolve os pontos
	if itemID == "skill_reset" {
		learnedSkills, _ := database.GetLearnedSkills(char.ID)
		refunded := 0
		for _, ls := range learnedSkills {
			if sk, ok := game.Skills[ls.SkillID]; ok {
				pc := sk.PointCost
				if pc == 0 {
					pc = 1
				}
				refunded += pc
			}
		}
		_, dbErr := database.ResetSkills(char.ID)
		if dbErr != nil {
			editPhoto(chatID, msgID, assets.ItemTypeImageKey(item.Type),
				"❌ Erro ao resetar habilidades. Tente novamente.", bkp("menu_inventory"))
			return
		}
		database.RemoveItem(char.ID, itemID, 1)
		char.SkillPoints += refunded
		database.SaveCharacter(char)
		caption := fmt.Sprintf(
			"📜 *Tomo do Esquecimento usado!*\n\nTodas as habilidades foram apagadas.\n🌟 *+%d pontos* devolvidos!\nTotal disponível: *%d pontos*\n\n_Acesse Habilidades para redistribuir._",
			refunded, char.SkillPoints,
		)
		editPhoto(chatID, msgID, "skills", caption, bkp("menu_skills"))
		return
	}
	// Antídoto: cura envenenamento
	if item.CurePoison {
		if char.PoisonTurns > 0 {
			char.PoisonTurns = 0
			char.PoisonDmg = 0
			effects += "☠️ Envenenamento curado! "
		} else {
			effects += "❓ Não estava envenenado. "
		}
	}
	// Bênção do Sábio: ativa bônus de XP
	if item.XPBoostMinutes > 0 {
		char.XPBoostExpiry = time.Now().Add(time.Duration(item.XPBoostMinutes) * time.Minute)
		effects += fmt.Sprintf("📖 +50%% XP por %d min! ", item.XPBoostMinutes)
	}
	// Token de Reviver: ativa automaticamente
	if itemID == "revive_token" {
		editPhoto(chatID, msgID, assets.ItemTypeImageKey(item.Type),
			"🔮 *Token de Reviver*\n\n_Este item é ativado *automaticamente* ao morrer em combate._\n\nGuarde-o para emergências!",
			bkp("menu_inventory"))
		return
	}

	if err := database.RemoveItem(char.ID, itemID, 1); err != nil {
		editPhoto(chatID, msgID, assets.ItemTypeImageKey(item.Type),
			"❌ Erro ao consumir item. Tente novamente.", bkp("menu_inventory"))
		return
	}
	if err := database.SaveCharacter(char); err != nil {
		editPhoto(chatID, msgID, assets.ItemTypeImageKey(item.Type),
			"❌ Erro ao salvar personagem. Tente novamente.", bkp("menu_inventory"))
		return
	}
	caption := fmt.Sprintf("%s *%s* usada!\n\n%s\n❤️ %d/%d | 💙 %d/%d | ⚡ %d/%d",
		item.Emoji, item.Name, effects, char.HP, char.HPMax, char.MP, char.MPMax, char.Energy, char.EnergyMax)
	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⬅️ Inventário", "menu_inventory"),
		),
	)
	editPhoto(chatID, msgID, assets.ItemTypeImageKey(item.Type), caption, &kb)
}

func handleInventoryEquip(chatID int64, msgID int, userID int64, itemID string) {
	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}
	item, ok := game.Items[itemID]
	if !ok {
		return
	}
	// Equipar a partir do inventário não deve abrir tela secundária.
	// Resolve slot automaticamente (incluindo acessórios legados sem Slot explícito).
	slot := resolveInventoryEquipSlot(char, item)

	// Class restriction
	if item.ClassReq != "" && item.ClassReq != char.Class {
		editPhoto(chatID, msgID, "inventory",
			fmt.Sprintf("❌ Somente *%s* pode equipar *%s*!", game.Classes[item.ClassReq].Name, item.Name),
			bkp("menu_inventory"))
		return
	}
	// Level restriction
	if item.MinLevel > char.Level {
		editPhoto(chatID, msgID, "inventory",
			fmt.Sprintf("❌ Precisa de nível *%d* para equipar *%s*! (você é nível *%d*)", item.MinLevel, item.Name, char.Level),
			bkp("menu_inventory"))
		return
	}

	sanitizeEquippedConflicts(char.ID)
	database.EquipItemSlot(char.ID, itemID, item.Type, slot)
	recalculateStats(char)
	database.SaveCharacter(char)
	showInventoryHome(chatID, msgID, userID)
}

func normalizeEquippedSlot(inv models.InventoryItem, item models.Item) string {
	if inv.Slot != "" {
		return inv.Slot
	}
	if item.Slot != "" {
		return item.Slot
	}
	lower := strings.ToLower(inv.ItemID + " " + item.Name)
	if strings.Contains(lower, "shield") || strings.Contains(lower, "escudo") {
		return "offhand"
	}
	switch item.Type {
	case "weapon":
		return "weapon"
	case "armor":
		// Fallback para registros muito antigos sem slot persistido.
		return "chest"
	case "accessory":
		return "accessory1"
	}
	return ""
}

// sanitizeEquippedConflicts corrige estados legados onde mais de um item ficou
// marcado como equipado para o mesmo slot.
func sanitizeEquippedConflicts(charID int) {
	inv, _ := database.GetInventory(charID)
	occupied := map[string]string{} // slot -> itemID
	for _, it := range inv {
		if !it.Equipped {
			continue
		}
		item, ok := game.Items[it.ItemID]
		if !ok {
			continue
		}
		slot := normalizeEquippedSlot(it, item)
		if slot == "" {
			continue
		}
		if existingID, exists := occupied[slot]; exists && existingID != it.ItemID {
			_ = database.UnequipItem(charID, it.ItemID)
			continue
		}
		occupied[slot] = it.ItemID
	}
}

func resolveInventoryEquipSlot(char *models.Character, item models.Item) string {
	if item.Slot != "" {
		return item.Slot
	}
	if item.Type == "weapon" {
		return "weapon"
	}
	if strings.Contains(strings.ToLower(item.ID), "shield") || strings.Contains(strings.ToLower(item.Name), "escudo") {
		return "offhand"
	}
	if item.Type != "accessory" {
		return "chest"
	}

	occupied := map[string]bool{}
	equipped, _ := database.GetEquippedItems(char.ID)
	for _, e := range equipped {
		s := e.Slot
		if s == "" {
			if it, ok := game.Items[e.ItemID]; ok {
				s = it.Slot
			}
		}
		if s != "" {
			occupied[s] = true
		}
	}

	if !occupied["accessory1"] {
		return "accessory1"
	}
	if !occupied["accessory2"] {
		return "accessory2"
	}

	lower := strings.ToLower(item.Name)
	if strings.Contains(lower, "colar") || strings.Contains(lower, "amuleto") || strings.Contains(lower, "pingente") {
		return "accessory2"
	}
	return "accessory1"
}

func skillRequiresShield(skillID string) bool {
	switch skillID {
	case "w_shield_bash":
		return true
	}
	return false
}

func hasEquippedShield(charID int) bool {
	equipped, _ := database.GetEquippedItems(charID)
	for _, e := range equipped {
		if e.Slot == "offhand" {
			return true
		}
		id := strings.ToLower(e.ItemID)
		if strings.Contains(id, "shield") {
			return true
		}
		if it, ok := game.Items[e.ItemID]; ok {
			if strings.Contains(strings.ToLower(it.Name), "escudo") {
				return true
			}
		}
	}
	return false
}

func recalculateStats(char *models.Character) {
	baseHP, baseMP, baseAtk, baseDef, baseMatk, baseMdef, _ := game.CalculateBaseStats(char.Race, char.Class)
	lv := char.Level - 1
	char.Attack = baseAtk + lv*2
	char.Defense = baseDef + lv
	char.MagicAttack = baseMatk + lv
	char.MagicDefense = baseMdef + lv

	// Resetar bônus de CA e Hit antes de recalcular
	char.EquipCABonus = 0
	char.EquipHitBonus = 0

	// Recalcular HP/MP max base (level-up grants are tracked separately)
	newHPMax := baseHP + lv*8
	newMPMax := baseMP + lv*4

	equipped, _ := database.GetEquippedItems(char.ID)
	for _, e := range equipped {
		item := game.Items[e.ItemID]
		char.Attack += item.AttackBonus
		char.Defense += item.DefenseBonus
		char.MagicAttack += item.MagicAtkBonus
		char.MagicDefense += item.MagicDefBonus
		char.EquipCABonus += item.CABonus
		char.EquipHitBonus += item.HitBonus
		newHPMax += item.HPBonus
		newMPMax += item.MPBonus
	}
	// Apply HP/MP max from equipment (preserva proporção quando possível).
	oldHPMax := char.HPMax
	if oldHPMax > 0 {
		ratio := float64(char.HP) / float64(oldHPMax)
		char.HPMax = newHPMax
		char.HP = int(ratio * float64(newHPMax))
	} else {
		char.HPMax = newHPMax
		if char.HP <= 0 || char.HP > char.HPMax {
			char.HP = char.HPMax
		}
	}
	if char.HP < 1 {
		char.HP = 1
	}
	if char.HP > char.HPMax {
		char.HP = char.HPMax
	}

	oldMPMax := char.MPMax
	char.MPMax = newMPMax
	if oldMPMax <= 0 && char.MP <= 0 {
		char.MP = char.MPMax
	}
	if char.MP > char.MPMax {
		char.MP = char.MPMax
	}
	if char.MP < 0 {
		char.MP = 0
	}
}

// =============================================
// SKILL TREE
// =============================================

// branchEmoji retorna emoji do ramo de build para exibição
func branchEmoji(branch string) string {
	switch branch {
	case "protetor":
		return "🛡️"
	case "berserker":
		return "💢"
	case "campiao":
		return "👑"
	case "piromante":
		return "🔥"
	case "crionita":
		return "❄️"
	case "arcanista":
		return "⚡"
	case "assassino":
		return "🗡️"
	case "envenenador":
		return "☠️"
	case "sombra":
		return "👤"
	case "atirador":
		return "🎯"
	case "cacador":
		return "🏹"
	case "arcano":
		return "✨"
	}
	return "🌿"
}

// branchLabel retorna nome do ramo para exibição
func branchLabel(branch string) string {
	switch branch {
	case "protetor":
		return "Protetor"
	case "berserker":
		return "Berserker"
	case "campiao":
		return "Campeão"
	case "piromante":
		return "Piromante"
	case "crionita":
		return "Crionita"
	case "arcanista":
		return "Arcanista"
	case "assassino":
		return "Assassino"
	case "envenenador":
		return "Envenenador"
	case "sombra":
		return "Sombra"
	case "atirador":
		return "Atirador"
	case "cacador":
		return "Caçador"
	case "arcano":
		return "Arcano"
	}
	return branch
}

func branchOrderIndex(class, branch string) int {
	switch class {
	case "warrior":
		switch branch {
		case "protetor":
			return 0
		case "berserker":
			return 1
		case "campiao":
			return 2
		}
	case "mage":
		switch branch {
		case "piromante":
			return 0
		case "crionita":
			return 1
		case "arcanista":
			return 2
		}
	case "rogue":
		switch branch {
		case "assassino":
			return 0
		case "envenenador":
			return 1
		case "sombra":
			return 2
		}
	case "archer":
		switch branch {
		case "atirador":
			return 0
		case "cacador":
			return 1
		case "arcano":
			return 2
		}
	}
	return 100
}

// showSkillTree exibe a árvore de habilidades organizada por ramos de build.
// Cada ramo tem 4 tiers. Custo: T1=1pt T2=1pt T3=2pts T4=3pts
// O jogador ganha 1 ponto por nível (total 19 pontos ao atingir nível 20).
func showSkillTree(chatID int64, msgID int, userID int64) {
	showSkillBranch(chatID, msgID, userID, "")
}

func showSkillBranch(chatID int64, msgID int, userID int64, activeBranch string) {
	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}
	learned, _ := database.GetLearnedSkills(char.ID)
	learnedMap := map[string]bool{}
	for _, l := range learned {
		learnedMap[l.SkillID] = true
	}

	// Agrupar habilidades da classe por ramo
	allSkills := game.GetSkillsForClass(char.Class)
	branches := map[string][]models.Skill{}
	branchOrder := []string{}
	for _, sk := range allSkills {
		if _, exists := branches[sk.Branch]; !exists {
			branchOrder = append(branchOrder, sk.Branch)
		}
		branches[sk.Branch] = append(branches[sk.Branch], sk)
	}
	sort.SliceStable(branchOrder, func(i, j int) bool {
		ii := branchOrderIndex(char.Class, branchOrder[i])
		jj := branchOrderIndex(char.Class, branchOrder[j])
		if ii != jj {
			return ii < jj
		}
		return branchLabel(branchOrder[i]) < branchLabel(branchOrder[j])
	})

	// Se nenhum ramo selecionado, usa o primeiro
	if activeBranch == "" {
		activeBranch = branchOrder[0]
	}

	c := game.Classes[char.Class]

	// Calcular pontos totais gastos
	spent := 0
	for _, l := range learned {
		if sk, ok := game.Skills[l.SkillID]; ok {
			pc := sk.PointCost
			if pc == 0 {
				pc = 1
			}
			spent += pc
		}
	}

	caption := fmt.Sprintf(
		"🌟 *Árvore de Habilidades*\n%s %s | Nível *%d*\n🔵 Pontos disponíveis: *%d* | Gastos: *%d*\n\n"+
			"_Ganho: +1 ponto por nível (máx 19 ao lv20)_\n"+
			"_Custo: T1=1pt · T2=1pt · T3=2pts · T4=3pts_\n\n",
		c.Emoji, c.Name, char.Level, char.SkillPoints, spent,
	)

	// Exibe habilidades do ramo ativo
	skillsInBranch := branches[activeBranch]
	// Ordenação estável para evitar botões "pulando" entre renderizações.
	sort.SliceStable(skillsInBranch, func(i, j int) bool {
		if skillsInBranch[i].Tier != skillsInBranch[j].Tier {
			return skillsInBranch[i].Tier < skillsInBranch[j].Tier
		}
		if skillsInBranch[i].RequiredLevel != skillsInBranch[j].RequiredLevel {
			return skillsInBranch[i].RequiredLevel < skillsInBranch[j].RequiredLevel
		}
		return skillsInBranch[i].Name < skillsInBranch[j].Name
	})

	caption += fmt.Sprintf("%s *Ramo: %s*\n", branchEmoji(activeBranch), branchLabel(activeBranch))

	var rows [][]tgbotapi.InlineKeyboardButton

	for _, sk := range skillsInBranch {
		pc := sk.PointCost
		if pc == 0 {
			pc = 1
		}

		var status string
		canLearn := false

		switch {
		case learnedMap[sk.ID]:
			status = "✅"
		case char.Level < sk.RequiredLevel:
			status = fmt.Sprintf("🔒 Nv.%d", sk.RequiredLevel)
		case sk.Requires != "" && !learnedMap[sk.Requires]:
			if req, ok := game.Skills[sk.Requires]; ok {
				status = fmt.Sprintf("🔗 requer %s", req.Name)
			} else {
				status = "🔗"
			}
		case char.SkillPoints < pc:
			status = fmt.Sprintf("💤 sem pts (%d)", pc)
		default:
			status = "📖 disponível"
			canLearn = true
		}

		passiveTag := ""
		if sk.Passive {
			passiveTag = " *(passiva)*"
		}
		mpTag := ""
		if sk.MPCost > 0 {
			mpTag = fmt.Sprintf(" %dMP", sk.MPCost)
		}

		caption += fmt.Sprintf("\nT%d %s *%s*%s%s — *%dpt(s)*\n%s\n_%s_\n",
			sk.Tier, sk.Emoji, sk.Name, passiveTag, mpTag, pc, status, sk.Description)

		if canLearn {
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					fmt.Sprintf("📖 %s %s (%dpt)", sk.Emoji, sk.Name, pc),
					"skill_learn_"+sk.ID,
				),
			))
		}
	}

	// Botões de navegação entre ramos
	var branchBtns []tgbotapi.InlineKeyboardButton
	for _, b := range branchOrder {
		mark := ""
		if b == activeBranch {
			mark = "› "
		}
		branchBtns = append(branchBtns, tgbotapi.NewInlineKeyboardButtonData(
			fmt.Sprintf("%s%s%s", mark, branchEmoji(b), branchLabel(b)),
			"skill_branch_"+b,
		))
	}
	// 3 botões por linha
	for i := 0; i < len(branchBtns); i += 3 {
		end := i + 3
		if end > len(branchBtns) {
			end = len(branchBtns)
		}
		rows = append(rows, branchBtns[i:end])
	}

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("🏰 Menu", "menu_main"),
	))

	kb := tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}
	editPhoto(chatID, msgID, "skills", caption, &kb)
}

func handleLearnSkill(chatID int64, msgID int, userID int64, skillID string) {
	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}
	sk, ok := game.Skills[skillID]
	if !ok {
		return
	}

	pc := sk.PointCost
	if pc == 0 {
		pc = 1
	}

	// Validações
	if sk.Class != char.Class {
		editPhoto(chatID, msgID, "skills", "❌ Habilidade não é da sua classe!", bkp("menu_skills"))
		return
	}
	if database.HasSkill(char.ID, skillID) {
		editPhoto(chatID, msgID, "skills", "❌ Já aprendida!", bkp("menu_skills"))
		return
	}
	if char.SkillPoints < pc {
		editPhoto(chatID, msgID, "skills",
			fmt.Sprintf("❌ Pontos insuficientes! Precisa: *%d* | Disponível: *%d*", pc, char.SkillPoints),
			bkp("menu_skills"))
		return
	}
	if char.Level < sk.RequiredLevel {
		editPhoto(chatID, msgID, "skills",
			fmt.Sprintf("❌ Requer nível *%d*! (você tem %d)", sk.RequiredLevel, char.Level),
			bkp("menu_skills"))
		return
	}
	if sk.Requires != "" && !database.HasSkill(char.ID, sk.Requires) {
		r := game.Skills[sk.Requires]
		editPhoto(chatID, msgID, "skills",
			fmt.Sprintf("❌ Aprenda *%s* primeiro!", r.Name),
			bkp("menu_skills"))
		return
	}

	database.LearnSkill(char.ID, skillID)
	char.SkillPoints -= pc
	database.SaveCharacter(char)

	passiveNote := ""
	if sk.Passive {
		passiveNote = "\n_Efeito passivo aplicado automaticamente em combate._"
	}
	caption := fmt.Sprintf(
		"🌟 *Habilidade Aprendida!*\n\n%s *%s*\nRamo: %s%s\n\n_%s_\n\n🔵 Pontos restantes: *%d*",
		sk.Emoji, sk.Name, branchLabel(sk.Branch), passiveNote, sk.Description, char.SkillPoints,
	)
	editPhoto(chatID, msgID, "skills", caption, bkp("menu_skills"))
}

// =============================================
// DELETE CHARACTER
// =============================================

func confirmDeleteCharacter(chatID int64, msgID int, _ int64) {
	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ Sim, apagar", "delete_confirm"),
			tgbotapi.NewInlineKeyboardButtonData("❌ Cancelar", "menu_status"),
		),
	)
	editPhoto(chatID, msgID, "defeat", "⚠️ *Apagar Personagem?*\n\nIsso é *PERMANENTE*. Todo progresso será perdido!", &kb)
}

func handleDeleteCharacter(chatID int64, msgID int, userID int64) {
	database.DeleteCharacter(userID)
	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⚔️ Criar Novo", "create_character"),
		),
	)
	editPhoto(chatID, msgID, "defeat", "✅ Personagem apagado. Até a próxima aventura!", &kb)
}

// =============================================
// HELPERS
// =============================================

// backKb retorna teclado com botao Voltar apontando para a secao correta.
// bkp e o atalho que retorna ponteiro (para usar com editPhoto).
func bkp(dest string) *tgbotapi.InlineKeyboardMarkup {
	kb := backKb(dest)
	return &kb
}

func backKb(dest string) tgbotapi.InlineKeyboardMarkup {
	labels := map[string]string{
		"menu_main":      "🏰 Menu Principal",
		"menu_status":    "⬅️ Voltar ao Status",
		"menu_inventory": "⬅️ Voltar ao Inventário",
		"menu_equip":     "⬅️ Equipamentos",
		"menu_skills":    "⬅️ Voltar as Habilidades",
		"menu_shop":      "⬅️ Voltar a Loja",
		"menu_sell":      "⬅️ Voltar as Vendas",
		"menu_travel":    "⬅️ Voltar ao Mapa",
		"menu_explore":   "⬅️ Voltar a Exploracao",
		"menu_energy":    "⬅️ Voltar a Energia",
		"menu_diamonds":  "⬅️ Voltar aos Diamantes",
		"menu_dungeon":   "⬅️ Voltar ao Dungeon",
		"menu_vip":       "⬅️ Voltar ao VIP",
		"menu_pvp":       "⬅️ Voltar ao PVP",
		"menu_rank":      "⬅️ Voltar ao Ranking",
		"diamond_shop":   "⬅️ Voltar a Loja de Diamantes",
	}
	label, ok := labels[dest]
	if !ok {
		label = "⬅️ Voltar"
	}
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(label, dest),
		),
	)
}

var backKeyboard = backKb("menu_main")

func sendMsg(chatID int64, text string, kb *tgbotapi.InlineKeyboardMarkup) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	if kb != nil {
		msg.ReplyMarkup = kb
	}
	if _, err := Bot.Send(msg); err != nil {
		log.Printf("sendMsg: %v", err)
	}
}

func sendText(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	Bot.Send(msg)
}

func editMsg(chatID int64, msgID int, text string, kb *tgbotapi.InlineKeyboardMarkup) {
	edit := tgbotapi.NewEditMessageText(chatID, msgID, text)
	edit.ParseMode = "Markdown"
	if kb != nil {
		edit.ReplyMarkup = kb
	}
	if _, err := Bot.Send(edit); err != nil {
		if isNotModified(err) {
			return // conteúdo já está correto, nada a fazer
		}
		sendMsg(chatID, text, kb)
	}
}
