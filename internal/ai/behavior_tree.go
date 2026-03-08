package ai

import "github.com/tormenta-bot/internal/engine"

// ─── Behaviour Tree ────────────────────────────────────────────────────────────
//
// A lightweight behaviour tree implementation that drives complex monster AI.
// Nodes evaluate their children left-to-right and return Success/Failure/Running.

// NodeStatus is the outcome of evaluating a BT node.
type NodeStatus int

const (
	NodeSuccess NodeStatus = 0
	NodeFailure NodeStatus = 1
	NodeRunning NodeStatus = 2
)

// Node is a behaviour tree node that can be evaluated.
type Node interface {
	Eval(ctx *BTContext) NodeStatus
}

// BTContext provides combat state to behaviour tree nodes.
type BTContext struct {
	MonsterHPPct    float64
	PlayerHPPct     float64
	TurnNumber      int
	Adaptation      engine.MonsterAdaptation
	ChosenAction    engine.AIActionKind // populated by action leaves
	LastPlayerAction string
}

// ─── Composite nodes ──────────────────────────────────────────────────────────

// Sequence succeeds only if ALL children succeed (short-circuits on first failure).
type Sequence struct{ Children []Node }

func (s *Sequence) Eval(ctx *BTContext) NodeStatus {
	for _, c := range s.Children {
		if status := c.Eval(ctx); status != NodeSuccess {
			return status
		}
	}
	return NodeSuccess
}

// Selector succeeds when ANY child succeeds (short-circuits on first success).
type Selector struct{ Children []Node }

func (s *Selector) Eval(ctx *BTContext) NodeStatus {
	for _, c := range s.Children {
		if status := c.Eval(ctx); status == NodeSuccess {
			return NodeSuccess
		}
	}
	return NodeFailure
}

// ─── Condition nodes ──────────────────────────────────────────────────────────

// HPBelowCondition succeeds when monster HP% < threshold.
type HPBelowCondition struct{ Threshold float64 }

func (c *HPBelowCondition) Eval(ctx *BTContext) NodeStatus {
	if ctx.MonsterHPPct < c.Threshold {
		return NodeSuccess
	}
	return NodeFailure
}

// TurnModCondition succeeds when turn number is divisible by N.
type TurnModCondition struct{ Mod int }

func (c *TurnModCondition) Eval(ctx *BTContext) NodeStatus {
	if ctx.TurnNumber > 0 && ctx.TurnNumber%c.Mod == 0 {
		return NodeSuccess
	}
	return NodeFailure
}

// AdaptedCondition succeeds when the monster has reached a given AI tier.
type AdaptedCondition struct{ MinTier engine.AITier }

func (c *AdaptedCondition) Eval(ctx *BTContext) NodeStatus {
	if ctx.Adaptation.Tier >= c.MinTier {
		return NodeSuccess
	}
	return NodeFailure
}

// ─── Action leaves ────────────────────────────────────────────────────────────

// SetAction is a leaf that assigns the chosen action.
type SetAction struct{ Action engine.AIActionKind }

func (a *SetAction) Eval(ctx *BTContext) NodeStatus {
	ctx.ChosenAction = a.Action
	return NodeSuccess
}

// ─── Pre-built behaviour trees ────────────────────────────────────────────────

// BasicMonsterTree builds a simple but adaptive behaviour tree.
// Priority: Enrage when low HP → Heal when wounded → Skill on cooldown mod → Basic attack.
func BasicMonsterTree() Node {
	return &Selector{
		Children: []Node{
			// Enrage when < 15% HP and veteran+ tier
			&Sequence{
				Children: []Node{
					&HPBelowCondition{0.15},
					&AdaptedCondition{engine.AITierVeteran},
					&SetAction{engine.AIActionEnrage},
				},
			},
			// Skill attack every 3 turns for adapted tier+
			&Sequence{
				Children: []Node{
					&AdaptedCondition{engine.AITierAdapted},
					&TurnModCondition{3},
					&SetAction{engine.AIActionSkillAttack},
				},
			},
			// Default: basic attack
			&SetAction{engine.AIActionBasicAttack},
		},
	}
}

// EliteBossTree builds a more complex tree for elite and boss monsters.
func EliteBossTree() Node {
	return &Selector{
		Children: []Node{
			// Enrage at < 20% for all elite bosses
			&Sequence{
				Children: []Node{
					&HPBelowCondition{0.20},
					&SetAction{engine.AIActionEnrage},
				},
			},
			// Self-heal at < 40% HP
			&Sequence{
				Children: []Node{
					&HPBelowCondition{0.40},
					&TurnModCondition{2},
					&SetAction{engine.AIActionHeal},
				},
			},
			// Skill every 2 turns
			&Sequence{
				Children: []Node{
					&TurnModCondition{2},
					&SetAction{engine.AIActionSkillAttack},
				},
			},
			&SetAction{engine.AIActionBasicAttack},
		},
	}
}

// EvalTree evaluates a tree and returns the chosen action (defaults to basic attack).
func EvalTree(tree Node, ctx *BTContext) engine.AIActionKind {
	ctx.ChosenAction = engine.AIActionBasicAttack // default
	tree.Eval(ctx)
	return ctx.ChosenAction
}
