package energy

import (
	"os"
	"strings"
	"time"

	"github.com/tormenta-bot/internal/database"
)

type Service struct{}

func NewService() *Service { return &Service{} }

func (s *Service) Tick(limit int) (int, error) {
	if limit <= 0 {
		limit = 500
	}
	rows, err := database.GetEnergyTickCandidates(limit)
	if err != nil {
		return 0, err
	}
	updated := 0
	now := time.Now().Unix()
	fixedCap := energyFixedCapEnabled()

	for _, r := range rows {
		maxEnergy := r.EnergyMax
		interval := int64(600)
		if r.IsVIP {
			interval = 300
			if fixedCap {
				maxEnergy = 200
			}
		} else if fixedCap {
			maxEnergy = 100
		}
		if maxEnergy < 1 {
			maxEnergy = 100
		}
		if r.Energy >= maxEnergy {
			continue
		}
		last := r.LastEnergyUpdate
		if last <= 0 {
			last = now
		}
		elapsed := now - last
		if elapsed < interval {
			continue
		}
		regained := int(elapsed / interval)
		newEnergy := r.Energy + regained
		if newEnergy > maxEnergy {
			newEnergy = maxEnergy
		}
		newLast := last + int64(regained)*interval
		if err := database.SaveCharacterEnergy(r.CharacterID, newEnergy, maxEnergy, time.Unix(newLast, 0)); err != nil {
			continue
		}
		updated++
	}
	return updated, nil
}

func energyFixedCapEnabled() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv("ENERGY_FIXED_CAP")))
	return v == "1" || v == "true" || v == "yes" || v == "on"
}
