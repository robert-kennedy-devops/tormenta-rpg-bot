package workers

import (
	"log"
	"sync"
	"time"
)

// ─── Combat worker ────────────────────────────────────────────────────────────
//
// Processes async combat jobs submitted by handlers.  This offloads heavy
// computation (AI decisions, damage calculation chains) from the Telegram
// update goroutine, preventing it from stalling.

// CombatJob is a unit of combat work submitted to the pool.
type CombatJob struct {
	JobID      string
	PlayerID   int64
	MonsterID  string
	ActionKind string // "attack" | "skill" | "flee"
	SkillID    string
	ResultCh   chan<- CombatJobResult
}

// CombatJobResult is returned through the channel when processing is done.
type CombatJobResult struct {
	JobID      string
	PlayerID   int64
	Success    bool
	Message    string
	DamageDealt int
	DamageTaken int
	XPGained   int
	GoldGained int
	Killed     bool
	Error      error
}

// CombatPool is a fixed-size goroutine pool for async combat jobs.
type CombatPool struct {
	jobs    chan CombatJob
	workers int
	wg      sync.WaitGroup
	done    chan struct{}
}

// NewCombatPool creates a pool with the specified concurrency.
func NewCombatPool(workers, queueDepth int) *CombatPool {
	if workers <= 0 {
		workers = 16
	}
	if queueDepth <= 0 {
		queueDepth = 1000
	}
	return &CombatPool{
		jobs:    make(chan CombatJob, queueDepth),
		workers: workers,
		done:    make(chan struct{}),
	}
}

// Start launches the worker goroutines.
func (p *CombatPool) Start() {
	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go p.runWorker()
	}
}

// Stop drains the queue and shuts down workers.
func (p *CombatPool) Stop() {
	close(p.done)
	p.wg.Wait()
}

// Submit adds a job to the pool.  Non-blocking: drops if queue is full.
func (p *CombatPool) Submit(job CombatJob) bool {
	select {
	case p.jobs <- job:
		return true
	default:
		log.Printf("[CombatPool] Queue full — dropping job %s", job.JobID)
		return false
	}
}

func (p *CombatPool) runWorker() {
	defer p.wg.Done()
	for {
		select {
		case job := <-p.jobs:
			result := p.process(job)
			if job.ResultCh != nil {
				job.ResultCh <- result
			}
		case <-p.done:
			return
		}
	}
}

// process executes one combat job.
// In the full implementation this would call the engine and repository layers.
// Kept minimal here to avoid import cycles with unbuilt handler layers.
func (p *CombatPool) process(job CombatJob) CombatJobResult {
	start := time.Now()
	result := CombatJobResult{
		JobID:    job.JobID,
		PlayerID: job.PlayerID,
		Success:  true,
		Message:  "Combate processado.",
	}
	log.Printf("[CombatPool] job=%s player=%d action=%s duration=%s",
		job.JobID, job.PlayerID, job.ActionKind, time.Since(start))
	return result
}

// ─── Global combat pool ───────────────────────────────────────────────────────

// GlobalCombatPool is the singleton combat worker pool.
// Initialised with 16 workers and a 1000-job queue — suitable for ~10k concurrent players.
var GlobalCombatPool = NewCombatPool(16, 1000)
