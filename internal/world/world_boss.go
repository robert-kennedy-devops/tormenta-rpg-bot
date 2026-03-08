// Package world manages global game events: world bosses, seasonal raids and
// cross-player world events that affect all online players simultaneously.
package world

import (
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"
)

// ─── Spawn configuration ──────────────────────────────────────────────────────

const (
	BossSpawnInterval = 12 * time.Hour // bosses respawn every 12h
	BossWindowDuration = 30 * time.Minute // players have 30m to defeat it
)

// ─── World boss definition ────────────────────────────────────────────────────

// BossDef is the static template for a world boss.
type BossDef struct {
	ID          string
	Name        string
	Emoji       string
	Description string
	BaseHP      int     // HP at 1 participant; scales with participant count
	BaseAttack  int
	BaseDefense int
	Level       int
	Element     string  // primary element ("fire","ice", etc.)
	Weakness    string
	Rewards     BossRewardTable
}

// BossRewardTable defines what drops when the boss is killed.
type BossRewardTable struct {
	GoldPerRank    []int    // gold for ranks 1,2,3,4,5+ (index 0 = rank 1)
	XPPerRank      []int
	LegendaryItems []string // item IDs; random draw for top 3 participants
	GuaranteedDrop string   // item dropped for all participants
}

// Bosses is the registry of all world bosses.
var Bosses = map[string]BossDef{
	"tormenta": {
		ID: "tormenta", Name: "Tormenta", Emoji: "🌀",
		Description: "A tempestade primordial que corrói o mundo — o maior perigo de Arton.",
		BaseHP: 100_000, BaseAttack: 80, BaseDefense: 40, Level: 50,
		Element: "dark", Weakness: "holy",
		Rewards: BossRewardTable{
			GoldPerRank: []int{5000, 3000, 2000, 1000, 500},
			XPPerRank:   []int{10000, 7000, 5000, 3000, 1000},
			LegendaryItems: []string{"tormenta_blade", "tormenta_shield", "tormenta_crown"},
			GuaranteedDrop: "tormenta_shard",
		},
	},
	"scorpion_king": {
		ID: "scorpion_king", Name: "Rei Escorpião", Emoji: "🦂",
		Description: "Senhor do deserto de veneno — sua cauda traz morte instantânea.",
		BaseHP: 50_000, BaseAttack: 60, BaseDefense: 30, Level: 30,
		Element: "poison", Weakness: "fire",
		Rewards: BossRewardTable{
			GoldPerRank: []int{2000, 1200, 800, 400, 200},
			XPPerRank:   []int{4000, 2500, 1500, 800, 300},
			LegendaryItems: []string{"venom_fang", "scorpion_armor"},
			GuaranteedDrop: "poison_gland",
		},
	},
	"frost_dragon": {
		ID: "frost_dragon", Name: "Dragão de Gelo", Emoji: "🐲",
		Description: "Dragão ancião das montanhas geladas. Seu sopro congela o tempo.",
		BaseHP: 80_000, BaseAttack: 75, BaseDefense: 50, Level: 45,
		Element: "ice", Weakness: "fire",
		Rewards: BossRewardTable{
			GoldPerRank: []int{4000, 2500, 1500, 800, 400},
			XPPerRank:   []int{8000, 5000, 3000, 1500, 500},
			LegendaryItems: []string{"frostbite_sword", "ice_dragon_scale"},
			GuaranteedDrop: "dragon_scale",
		},
	},
}

// ─── Active boss instance ─────────────────────────────────────────────────────

// BossStatus tracks an active boss spawn.
type BossStatus string

const (
	BossAlive    BossStatus = "alive"
	BossKilled   BossStatus = "killed"
	BossExpired  BossStatus = "expired" // window closed before defeat
)

// Participant tracks one player's contribution to a boss fight.
type Participant struct {
	PlayerID   int64
	PlayerName string
	Damage     int
	Rank       int // determined after boss dies
}

// ActiveBoss is a live boss encounter.
type ActiveBoss struct {
	BossID       string
	Def          BossDef
	CurrentHP    int
	MaxHP        int
	Status       BossStatus
	SpawnedAt    time.Time
	ExpiresAt    time.Time
	KilledAt     *time.Time
	Participants map[int64]*Participant // playerID → contribution
}

// ScaleHP scales boss HP based on participant count.
// More players = harder boss.
func ScaleHP(baseHP, participants int) int {
	if participants <= 1 {
		return baseHP
	}
	// Scale: +60% HP per additional participant, soft cap at 10x
	scaled := float64(baseHP) * (1.0 + 0.6*float64(participants-1))
	max := float64(baseHP) * 10
	if scaled > max {
		scaled = max
	}
	return int(scaled)
}

// ─── Boss manager ────────────────────────────────────────────────────────────

// BossManager controls the lifecycle of world boss spawns.
type BossManager struct {
	mu         sync.RWMutex
	activeBoss *ActiveBoss
	nextSpawn  time.Time
}

// Global is the singleton BossManager.
var Global = &BossManager{nextSpawn: time.Now().Add(BossSpawnInterval)}

// SpawnIfReady spawns a random boss if the interval has elapsed and no boss is alive.
func (m *BossManager) SpawnIfReady() *ActiveBoss {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	if m.activeBoss != nil && m.activeBoss.Status == BossAlive {
		if now.After(m.activeBoss.ExpiresAt) {
			m.activeBoss.Status = BossExpired
		}
		return nil
	}
	if now.Before(m.nextSpawn) {
		return nil
	}

	// Choose a random boss
	keys := make([]string, 0, len(Bosses))
	for k := range Bosses {
		keys = append(keys, k)
	}
	bossID := keys[rand.Intn(len(keys))]
	def := Bosses[bossID]

	m.activeBoss = &ActiveBoss{
		BossID:       bossID,
		Def:          def,
		MaxHP:        def.BaseHP,
		CurrentHP:    def.BaseHP,
		Status:       BossAlive,
		SpawnedAt:    now,
		ExpiresAt:    now.Add(BossWindowDuration),
		Participants: make(map[int64]*Participant),
	}
	m.nextSpawn = now.Add(BossSpawnInterval)
	return m.activeBoss
}

// Active returns the current alive boss or nil.
func (m *BossManager) Active() *ActiveBoss {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.activeBoss == nil || m.activeBoss.Status != BossAlive {
		return nil
	}
	return m.activeBoss
}

// AttackBoss deals damage from a player and returns (isDead, dmgDealt, error).
func (m *BossManager) AttackBoss(playerID int64, playerName string, damage int) (killed bool, dmgDealt int, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.activeBoss == nil || m.activeBoss.Status != BossAlive {
		err = fmt.Errorf("não há nenhum boss ativo no momento")
		return
	}
	if time.Now().After(m.activeBoss.ExpiresAt) {
		m.activeBoss.Status = BossExpired
		err = fmt.Errorf("o boss escapou — tente na próxima vez!")
		return
	}

	// Rescale HP if this is a new participant
	p, known := m.activeBoss.Participants[playerID]
	if !known {
		newParticipants := len(m.activeBoss.Participants) + 1
		m.activeBoss.MaxHP = ScaleHP(m.activeBoss.Def.BaseHP, newParticipants)
		// Keep current HP ratio
		ratio := float64(m.activeBoss.CurrentHP) / float64(m.activeBoss.MaxHP)
		m.activeBoss.CurrentHP = int(ratio * float64(m.activeBoss.MaxHP))
		p = &Participant{PlayerID: playerID, PlayerName: playerName}
		m.activeBoss.Participants[playerID] = p
	}

	dmgDealt = damage
	if dmgDealt > m.activeBoss.CurrentHP {
		dmgDealt = m.activeBoss.CurrentHP
	}
	p.Damage += dmgDealt
	m.activeBoss.CurrentHP -= dmgDealt

	if m.activeBoss.CurrentHP <= 0 {
		now := time.Now()
		m.activeBoss.Status = BossKilled
		m.activeBoss.KilledAt = &now
		killed = true
		m.assignRanks()
	}
	return
}

// assignRanks sorts participants by damage and assigns ranks (called while locked).
func (m *BossManager) assignRanks() {
	type ranked struct {
		p    *Participant
		rank int
	}
	// Build sorted list (simple insertion sort — participant count is small)
	sorted := make([]*Participant, 0, len(m.activeBoss.Participants))
	for _, p := range m.activeBoss.Participants {
		sorted = append(sorted, p)
	}
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].Damage > sorted[i].Damage {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	for rank, p := range sorted {
		p.Rank = rank + 1
	}
}

// Rewards returns the reward for a participant after the boss is killed.
func (m *BossManager) Rewards(playerID int64) (gold, xp int, itemDrops []string) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.activeBoss == nil || m.activeBoss.Status != BossKilled {
		return
	}
	p, ok := m.activeBoss.Participants[playerID]
	if !ok {
		return
	}
	def := m.activeBoss.Def
	rankIdx := int(math.Min(float64(p.Rank-1), float64(len(def.Rewards.GoldPerRank)-1)))
	gold = def.Rewards.GoldPerRank[rankIdx]
	xp = def.Rewards.XPPerRank[rankIdx]

	// Guaranteed drop
	if def.Rewards.GuaranteedDrop != "" {
		itemDrops = append(itemDrops, def.Rewards.GuaranteedDrop)
	}
	// Legendary draw for top 3
	if p.Rank <= 3 && len(def.Rewards.LegendaryItems) > 0 {
		item := def.Rewards.LegendaryItems[rand.Intn(len(def.Rewards.LegendaryItems))]
		itemDrops = append(itemDrops, item)
	}
	return
}
