package season

import "fmt"

// ─── Season reward tiers ──────────────────────────────────────────────────────

// RewardTier is the bracket a player falls into at season end.
type RewardTier string

const (
	TierImmortal  RewardTier = "Imortal"
	TierLegendary RewardTier = "Lendário"
	TierMaster    RewardTier = "Mestre"
	TierPlatinum  RewardTier = "Platina"
	TierGold      RewardTier = "Ouro"
	TierSilver    RewardTier = "Prata"
	TierBronze    RewardTier = "Bronze"
)

// SeasonReward defines what a player in each tier receives.
type SeasonReward struct {
	Tier          RewardTier
	MinRating     int
	Diamonds      int
	ExclusiveItem string // cosmetic / seasonal item ID
	TitleUnlock   string // title displayed next to name
	GoldBonus     int    // one-time gold bonus
}

// RewardTable is the season-end reward table.
var RewardTable = []SeasonReward{
	{TierImmortal, 2400, 500, "immortal_crown", "🏆 Imortal", 50000},
	{TierLegendary, 2000, 300, "legendary_cloak", "💎 Lendário", 30000},
	{TierMaster, 1600, 200, "master_badge", "🥇 Mestre", 15000},
	{TierPlatinum, 1300, 100, "platinum_ring", "🥈 Platina", 8000},
	{TierGold, 1100, 50, "gold_coin_charm", "🥉 Ouro", 4000},
	{TierSilver, 900, 25, "silver_sword_skin", "⚔️ Prata", 2000},
	{TierBronze, 0, 10, "bronze_shield_skin", "🗡️ Bronze", 500},
}

// RewardForRating returns the appropriate reward for a given final season rating.
func RewardForRating(rating int) SeasonReward {
	for _, r := range RewardTable {
		if rating >= r.MinRating {
			return r
		}
	}
	return RewardTable[len(RewardTable)-1]
}

// SeasonEndSummary generates the end-of-season narrative for a player.
func SeasonEndSummary(season *Season, playerName string, finalRating, wins, losses int) string {
	reward := RewardForRating(finalRating)
	return fmt.Sprintf(
		"🏁 *Temporada %d — %s %s encerrada!*\n\n"+
			"Jogador: *%s*\n"+
			"Rating Final: *%d* (%s)\n"+
			"Resultado: %dV / %dD\n\n"+
			"🎁 *Recompensas:*\n"+
			"💎 %d Diamantes\n"+
			"💰 %d Ouro\n"+
			"🏅 Título: %s\n"+
			"🎖️ Item Exclusivo: %s\n\n"+
			"A nova temporada começa agora. Boa sorte, aventureiro!",
		season.ID, season.Emoji, season.Name,
		playerName,
		finalRating, string(reward.Tier),
		wins, losses,
		reward.Diamonds,
		reward.GoldBonus,
		reward.TitleUnlock,
		reward.ExclusiveItem,
	)
}
