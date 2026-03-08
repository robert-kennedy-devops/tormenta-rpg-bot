// Package pvp provides the player-versus-player combat arena, matchmaking
// ladder and seasonal ranking for Tormenta.
package pvp

import (
	"sort"
	"sync"
	"time"
)

// ─── Rating system (Elo-inspired) ─────────────────────────────────────────────

const (
	DefaultRating = 1000
	KFactor       = 32   // Elo K-factor
)

// RatingEntry is one player's current PVP rating.
type RatingEntry struct {
	PlayerID   int64
	PlayerName string
	Rating     int
	Wins       int
	Losses     int
	Streak     int // current win streak
	LastFight  time.Time
	SeasonID   int
}

// ExpectedScore computes the expected win probability for player A vs B.
func ExpectedScore(ratingA, ratingB int) float64 {
	return 1.0 / (1.0 + Pow10((float64(ratingB-ratingA))/400.0))
}

// Pow10 computes 10^x using a simple approach (avoids math import cycle concerns).
func Pow10(x float64) float64 {
	result := 1.0
	for i := 0; i < 10; i++ {
		result *= 1.0 + x/10.0
	}
	return result
}

// NewRating computes the updated Elo rating after a match.
// winner = true if player A won.
func NewRating(ratingA, ratingB int, winner bool) (newA, newB int) {
	ea := ExpectedScore(ratingA, ratingB)
	sa := 0.0
	if winner {
		sa = 1.0
	}
	newA = ratingA + int(float64(KFactor)*(sa-ea))
	newB = ratingB + int(float64(KFactor)*((1-sa)-(1-ea)))
	if newA < 0 {
		newA = 0
	}
	if newB < 0 {
		newB = 0
	}
	return
}

// Division returns the rank division name for a given rating.
func Division(rating int) string {
	switch {
	case rating >= 2400:
		return "🏆 Imortal"
	case rating >= 2000:
		return "💎 Lendário"
	case rating >= 1600:
		return "🥇 Mestre"
	case rating >= 1300:
		return "🥈 Platina"
	case rating >= 1100:
		return "🥉 Ouro"
	case rating >= 900:
		return "⚔️ Prata"
	default:
		return "🗡️ Bronze"
	}
}

// ─── Ranking store ────────────────────────────────────────────────────────────

// RankStore manages PVP rating records.
type RankStore struct {
	mu      sync.RWMutex
	entries map[int64]*RatingEntry
}

// GlobalRankStore is the singleton store.
var GlobalRankStore = &RankStore{entries: make(map[int64]*RatingEntry)}

// Get returns (or lazily creates) a rating entry for a player.
func (s *RankStore) Get(playerID int64, playerName string) *RatingEntry {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.entries[playerID]
	if !ok {
		e = &RatingEntry{
			PlayerID:   playerID,
			PlayerName: playerName,
			Rating:     DefaultRating,
		}
		s.entries[playerID] = e
	}
	return e
}

// Update saves a modified entry.
func (s *RankStore) Update(e *RatingEntry) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[e.PlayerID] = e
}

// Leaderboard returns the top N players sorted by rating.
func (s *RankStore) Leaderboard(n int) []*RatingEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	all := make([]*RatingEntry, 0, len(s.entries))
	for _, e := range s.entries {
		all = append(all, e)
	}
	sort.Slice(all, func(i, j int) bool { return all[i].Rating > all[j].Rating })
	if n > 0 && len(all) > n {
		all = all[:n]
	}
	return all
}

// ApplyResult applies a match result to both players' ratings.
func (s *RankStore) ApplyResult(winnerID, loserID int64, winnerName, loserName string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	winner := s.getOrCreate(winnerID, winnerName)
	loser := s.getOrCreate(loserID, loserName)

	newW, newL := NewRating(winner.Rating, loser.Rating, true)
	winner.Rating = newW
	winner.Wins++
	winner.Streak++
	winner.LastFight = time.Now()

	loser.Rating = newL
	loser.Losses++
	loser.Streak = 0
	loser.LastFight = time.Now()
}

func (s *RankStore) getOrCreate(playerID int64, playerName string) *RatingEntry {
	e, ok := s.entries[playerID]
	if !ok {
		e = &RatingEntry{
			PlayerID:   playerID,
			PlayerName: playerName,
			Rating:     DefaultRating,
		}
		s.entries[playerID] = e
	}
	return e
}

// SeasonReset resets all ratings to a soft reset (half the distance to default).
func (s *RankStore) SeasonReset(newSeasonID int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, e := range s.entries {
		// Soft reset: move halfway toward DefaultRating
		e.Rating = DefaultRating + (e.Rating-DefaultRating)/2
		e.Wins = 0
		e.Losses = 0
		e.Streak = 0
		e.SeasonID = newSeasonID
	}
}
