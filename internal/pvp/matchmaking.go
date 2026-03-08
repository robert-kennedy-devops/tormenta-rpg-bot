package pvp

import (
	"errors"
	"sync"
	"time"
)

var (
	ErrAlreadyQueued = errors.New("você já está na fila de PVP")
	ErrNotQueued     = errors.New("você não está na fila de PVP")
	ErrNoOpponent    = errors.New("nenhum oponente disponível no momento")
)

// QueueEntry is one player waiting for a match.
type QueueEntry struct {
	PlayerID   int64
	PlayerName string
	Rating     int
	CharLevel  int
	QueuedAt   time.Time
}

// MatchResult is a found match between two players.
type MatchResult struct {
	PlayerA QueueEntry
	PlayerB QueueEntry
}

// Matchmaker manages the PVP queue.
type Matchmaker struct {
	mu    sync.Mutex
	queue []*QueueEntry
}

// GlobalMatchmaker is the singleton queue.
var GlobalMatchmaker = &Matchmaker{}

const (
	initialRatingWindow = 100  // ±100 rating initially
	ratingExpansionRate = 25   // expand window by 25 per 30s of waiting
	maxRatingWindow     = 400  // max spread
)

// Enqueue adds a player to the PVP queue.
func (m *Matchmaker) Enqueue(entry QueueEntry) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, e := range m.queue {
		if e.PlayerID == entry.PlayerID {
			return ErrAlreadyQueued
		}
	}
	m.queue = append(m.queue, &entry)
	return nil
}

// Dequeue removes a player from the queue.
func (m *Matchmaker) Dequeue(playerID int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, e := range m.queue {
		if e.PlayerID == playerID {
			m.queue = append(m.queue[:i], m.queue[i+1:]...)
			return nil
		}
	}
	return ErrNotQueued
}

// FindMatch attempts to match a player with a suitable opponent.
// Returns ErrNoOpponent if no suitable match exists yet.
func (m *Matchmaker) FindMatch(playerID int64) (*MatchResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var seeker *QueueEntry
	var seekerIdx int
	for i, e := range m.queue {
		if e.PlayerID == playerID {
			seeker = e
			seekerIdx = i
			break
		}
	}
	if seeker == nil {
		return nil, ErrNotQueued
	}

	// Compute expanded window based on wait time
	waited := time.Since(seeker.QueuedAt)
	expansions := int(waited.Seconds() / 30)
	window := initialRatingWindow + expansions*ratingExpansionRate
	if window > maxRatingWindow {
		window = maxRatingWindow
	}

	// Find best opponent within window
	var best *QueueEntry
	var bestIdx int
	bestDiff := window + 1
	for i, e := range m.queue {
		if e.PlayerID == seeker.PlayerID {
			continue
		}
		diff := seeker.Rating - e.Rating
		if diff < 0 {
			diff = -diff
		}
		if diff <= window && diff < bestDiff {
			best = e
			bestIdx = i
			bestDiff = diff
		}
	}
	if best == nil {
		return nil, ErrNoOpponent
	}

	// Remove both from queue
	result := &MatchResult{PlayerA: *seeker, PlayerB: *best}
	toRemove := []int{seekerIdx, bestIdx}
	if seekerIdx > bestIdx {
		toRemove[0], toRemove[1] = toRemove[1], toRemove[0]
	}
	m.queue = append(m.queue[:toRemove[1]], m.queue[toRemove[1]+1:]...)
	m.queue = append(m.queue[:toRemove[0]], m.queue[toRemove[0]+1:]...)
	return result, nil
}

// QueueSize returns the current number of queued players.
func (m *Matchmaker) QueueSize() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.queue)
}
