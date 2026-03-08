package handlers

import (
	"fmt"
	"strings"

	"github.com/tormenta-bot/internal/database"
	"github.com/tormenta-bot/internal/market"
	menukit "github.com/tormenta-bot/internal/menu"
)

func showMarketMenu(chatID int64, msgID int, userID int64) {
	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}
	listings, _ := market.GlobalStore.ListActive("", 5, 0)
	caption := fmt.Sprintf(
		"🏪 *Mercado de Jogadores*\n\n"+
			"Compre e venda itens diretamente com outros aventureiros!\n\n"+
			"📦 Listagens ativas: *%d*\n"+
			"💰 Seu ouro: *%d*\n\n"+
			"_Escolha uma opção:_",
		len(listings), char.Gold,
	)
	kb := menukit.MarketMainMenu()
	editPhoto(chatID, msgID, "menu", caption, &kb)
}

func showMarketBrowse(chatID int64, msgID int, userID int64) {
	listings, _ := market.GlobalStore.ListActive("", 20, 0)
	caption := "🔍 *Navegar Mercado*\n\n"
	if len(listings) == 0 {
		caption += "_Nenhuma listagem disponível no momento._"
	} else {
		for i, l := range listings {
			if i >= 10 {
				caption += fmt.Sprintf("\n_...e mais %d itens_", len(listings)-10)
				break
			}
			caption += fmt.Sprintf("• *%s* — %d ouro (x%d)\n", l.ItemName, l.UnitPrice, l.Quantity)
		}
	}
	kb := menukit.MarketBrowseMenu()
	editMsg(chatID, msgID, caption, &kb)
}

func showMyListings(chatID int64, msgID int, userID int64) {
	char, _ := database.GetCharacter(userID)
	if char == nil {
		return
	}
	listings, _ := market.GlobalStore.ListBySeller(int64(char.ID))
	caption := "📦 *Minhas Listagens*\n\n"
	if len(listings) == 0 {
		caption += "_Você não possui listagens ativas._"
	} else {
		for _, l := range listings {
			statusEmoji := "✅"
			if l.Status != market.ListingActive {
				statusEmoji = "❌"
			}
			caption += fmt.Sprintf("%s *%s* x%d — %d ouro cada\n", statusEmoji, l.ItemName, l.Quantity, l.UnitPrice)
		}
	}
	kb := menukit.MenuOnly()
	editMsg(chatID, msgID, caption, &kb)
}

func showAuctionsMenu(chatID int64, msgID int, userID int64) {
	caption := "🔨 *Leilões*\n\n_Participe de leilões de itens raros e exclusivos!_"
	kb := menukit.MarketAuctionMenu()
	editMsg(chatID, msgID, caption, &kb)
}

func showActiveAuctions(chatID int64, msgID int, userID int64) {
	auctions, _ := market.GlobalAuctionStore.ListOpen(20, 0)
	caption := "🔨 *Leilões Ativos*\n\n"
	if len(auctions) == 0 {
		caption += "_Nenhum leilão ativo no momento._"
	} else {
		for _, a := range auctions {
			caption += fmt.Sprintf("• *%s* — Lance atual: *%d* ouro | Término: %s\n",
				a.ItemName, a.CurrentBid, a.EndsAt.Format("15:04"))
		}
	}
	kb := menukit.MenuOnly()
	editMsg(chatID, msgID, caption, &kb)
}

func showMarketCategory(chatID int64, msgID int, userID int64, category string) {
	filterID := ""
	if category != "all" {
		filterID = category
	}
	listings, _ := market.GlobalStore.ListActive(filterID, 20, 0)
	catLabel := "Todos"
	if category != "all" && len(category) > 0 {
		catLabel = strings.ToUpper(category[:1]) + category[1:]
	}
	caption := fmt.Sprintf("🔍 *Mercado — %s*\n\n", catLabel)
	if len(listings) == 0 {
		caption += "_Nenhum item desta categoria disponível._"
	} else {
		for i, l := range listings {
			if i >= 10 {
				caption += fmt.Sprintf("\n_...e mais %d itens_", len(listings)-10)
				break
			}
			caption += fmt.Sprintf("• *%s* — %d ouro (x%d)\n", l.ItemName, l.UnitPrice, l.Quantity)
		}
	}
	kb := menukit.MarketBrowseMenu()
	editMsg(chatID, msgID, caption, &kb)
}
