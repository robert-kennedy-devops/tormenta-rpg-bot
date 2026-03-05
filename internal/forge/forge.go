package forge

import "fmt"

const MaxUpgradeLevel = 10

// Success rates for target level (next level).
var SuccessRateByTargetLevel = map[int]float64{
	1:  1.00,
	2:  0.90,
	3:  0.80,
	4:  0.70,
	5:  0.60,
	6:  0.50,
	7:  0.40,
	8:  0.30,
	9:  0.20,
	10: 0.10,
}

// BreakRateOnFailByTargetLevel controls break chance when a fail can break.
var BreakRateOnFailByTargetLevel = map[int]float64{
	5:  0.10,
	6:  0.15,
	7:  0.20,
	8:  0.25,
	9:  0.30,
	10: 0.35,
}

type OutcomeStatus string

const (
	OutcomeSuccess  OutcomeStatus = "success"
	OutcomeFailSafe OutcomeStatus = "fail_safe"
	OutcomeBroken   OutcomeStatus = "broken"
)

type Outcome struct {
	Status      OutcomeStatus
	OldLevel    int
	TargetLevel int
	NewLevel    int
	Broken      bool
}

func CanAttempt(currentLevel int) bool {
	return currentLevel >= 0 && currentLevel < MaxUpgradeLevel
}

func ValidateLevel(currentLevel int) error {
	if currentLevel < 0 || currentLevel > MaxUpgradeLevel {
		return fmt.Errorf("invalid current level: %d", currentLevel)
	}
	return nil
}

func SuccessChance(currentLevel int) (float64, error) {
	if err := ValidateLevel(currentLevel); err != nil {
		return 0, err
	}
	target := currentLevel + 1
	ch, ok := SuccessRateByTargetLevel[target]
	if !ok {
		return 0, fmt.Errorf("missing success chance for target +%d", target)
	}
	return ch, nil
}

func Attempt(currentLevel int, successRoll, breakRoll float64) (Outcome, error) {
	if err := ValidateLevel(currentLevel); err != nil {
		return Outcome{}, err
	}
	if !CanAttempt(currentLevel) {
		return Outcome{}, fmt.Errorf("item already at max level +%d", MaxUpgradeLevel)
	}

	target := currentLevel + 1
	successChance, err := SuccessChance(currentLevel)
	if err != nil {
		return Outcome{}, err
	}

	if successRoll < successChance {
		return Outcome{
			Status:      OutcomeSuccess,
			OldLevel:    currentLevel,
			TargetLevel: target,
			NewLevel:    target,
		}, nil
	}

	// Fails up to +4 never break.
	if target <= 4 {
		return Outcome{
			Status:      OutcomeFailSafe,
			OldLevel:    currentLevel,
			TargetLevel: target,
			NewLevel:    currentLevel,
		}, nil
	}

	breakChance := BreakRateOnFailByTargetLevel[target]
	if breakRoll < breakChance {
		return Outcome{
			Status:      OutcomeBroken,
			OldLevel:    currentLevel,
			TargetLevel: target,
			NewLevel:    currentLevel,
			Broken:      true,
		}, nil
	}

	return Outcome{
		Status:      OutcomeFailSafe,
		OldLevel:    currentLevel,
		TargetLevel: target,
		NewLevel:    currentLevel,
	}, nil
}
