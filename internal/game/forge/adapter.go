package forge

import core "github.com/tormenta-bot/internal/forge"

const MaxUpgradeLevel = core.MaxUpgradeLevel

type Outcome = core.Outcome
type OutcomeStatus = core.OutcomeStatus

const (
	OutcomeSuccess  = core.OutcomeSuccess
	OutcomeFailSafe = core.OutcomeFailSafe
	OutcomeBroken   = core.OutcomeBroken
)

func CanAttempt(currentLevel int) bool                { return core.CanAttempt(currentLevel) }
func SuccessChance(currentLevel int) (float64, error) { return core.SuccessChance(currentLevel) }
func Attempt(currentLevel int, successRoll, breakRoll float64) (Outcome, error) {
	return core.Attempt(currentLevel, successRoll, breakRoll)
}
