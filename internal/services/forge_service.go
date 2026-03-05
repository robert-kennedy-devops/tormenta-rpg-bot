package services

import (
	"fmt"
	"math/rand"
	"os"
	"strings"

	"github.com/tormenta-bot/internal/forge"
	"github.com/tormenta-bot/internal/items"
)

type ForgeProfile string

const (
	ForgeProfileLegacy10 ForgeProfile = "legacy10"
	ForgeProfileClassic5 ForgeProfile = "classic5"
)

type ForgeService struct {
	profile ForgeProfile
}

func NewForgeService() *ForgeService {
	return &ForgeService{profile: resolveForgeProfile()}
}

func resolveForgeProfile() ForgeProfile {
	v := strings.ToLower(strings.TrimSpace(os.Getenv("FORGE_PROFILE")))
	switch v {
	case "classic", "classic5", "mmorpg":
		return ForgeProfileClassic5
	default:
		return ForgeProfileLegacy10
	}
}

func (s *ForgeService) Profile() ForgeProfile {
	return s.profile
}

func (s *ForgeService) MaxLevel() int {
	if s.profile == ForgeProfileClassic5 {
		return 5
	}
	return forge.MaxUpgradeLevel
}

func (s *ForgeService) CanAttempt(currentLevel int) bool {
	return currentLevel >= 0 && currentLevel < s.MaxLevel()
}

func (s *ForgeService) SuccessChance(currentLevel int) (float64, error) {
	if !s.CanAttempt(currentLevel) && currentLevel != s.MaxLevel() {
		return 0, fmt.Errorf("invalid current level: %d", currentLevel)
	}
	target := currentLevel + 1
	switch s.profile {
	case ForgeProfileClassic5:
		switch target {
		case 1:
			return 1.00, nil
		case 2:
			return 0.90, nil
		case 3:
			return 0.70, nil
		case 4:
			return 0.50, nil
		case 5:
			return 0.30, nil
		default:
			return 0, fmt.Errorf("missing success chance for target +%d", target)
		}
	default:
		return forge.SuccessChance(currentLevel)
	}
}

func (s *ForgeService) MaterialCostForTarget(target int) (string, int) {
	if target < 1 {
		target = 1
	}
	switch s.profile {
	case ForgeProfileClassic5:
		switch {
		case target <= 2:
			return items.MaterialIronOre, 1
		case target <= 3:
			return items.MaterialSilverOre, 1
		case target <= 4:
			return items.MaterialGoldOre, 1
		default:
			return items.MaterialMagicEssence, 1
		}
	default:
		switch {
		case target <= 5:
			return items.MaterialForgeStone, 1
		case target <= 8:
			return items.MaterialRefinedStone, 1
		default:
			return items.MaterialArcaneEssence, 1
		}
	}
}

func (s *ForgeService) Attempt(currentLevel int, successRoll, breakRoll float64) (forge.Outcome, error) {
	if currentLevel < 0 || currentLevel > s.MaxLevel() {
		return forge.Outcome{}, fmt.Errorf("invalid current level: %d", currentLevel)
	}
	if !s.CanAttempt(currentLevel) {
		return forge.Outcome{}, fmt.Errorf("item already at max level +%d", s.MaxLevel())
	}

	switch s.profile {
	case ForgeProfileClassic5:
		target := currentLevel + 1
		successChance, err := s.SuccessChance(currentLevel)
		if err != nil {
			return forge.Outcome{}, err
		}
		if successRoll < successChance {
			return forge.Outcome{
				Status:      forge.OutcomeSuccess,
				OldLevel:    currentLevel,
				TargetLevel: target,
				NewLevel:    target,
			}, nil
		}
		// Classic mode: can break on fail.
		// Keep low-risk behavior for early levels by requiring +4/+5 to break.
		if target >= 4 && breakRoll < 0.25 {
			return forge.Outcome{
				Status:      forge.OutcomeBroken,
				OldLevel:    currentLevel,
				TargetLevel: target,
				NewLevel:    currentLevel,
				Broken:      true,
			}, nil
		}
		return forge.Outcome{
			Status:      forge.OutcomeFailSafe,
			OldLevel:    currentLevel,
			TargetLevel: target,
			NewLevel:    currentLevel,
		}, nil
	default:
		return forge.Attempt(currentLevel, successRoll, breakRoll)
	}
}

func (s *ForgeService) AttemptRandom(currentLevel int) (forge.Outcome, error) {
	return s.Attempt(currentLevel, rand.Float64(), rand.Float64())
}
