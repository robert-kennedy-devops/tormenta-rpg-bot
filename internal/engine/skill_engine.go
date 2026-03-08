package engine

// ─── Skill resolution ──────────────────────────────────────────────────────────

// SkillUseRequest bundles all context needed to resolve a skill.
type SkillUseRequest struct {
	SkillID     string
	SkillEffects []Effect  // pre-built from skill definition
	MPCost      int
	EnergyCost  int

	AttackerMP    int
	AttackerEnergy int
}

// SkillUseResult is the outcome of attempting to use a skill.
type SkillUseResult struct {
	OK           bool
	FailReason   string
	MPConsumed   int
	EnergyConsumed int
	Effects      ProcessResult
}

// ResolveSkillUse validates resources, then delegates to ProcessEffects.
func ResolveSkillUse(req SkillUseRequest, attacker, target Combatant) SkillUseResult {
	if req.AttackerMP < req.MPCost {
		return SkillUseResult{OK: false, FailReason: "MP insuficiente."}
	}
	if req.AttackerEnergy < req.EnergyCost {
		return SkillUseResult{OK: false, FailReason: "Energia insuficiente."}
	}
	if attacker.GetStatusSet().Has(StatusSilence) {
		return SkillUseResult{OK: false, FailReason: "Você está silenciado e não pode usar habilidades!"}
	}

	result := ProcessEffects(req.SkillEffects, attacker, target)
	return SkillUseResult{
		OK:             true,
		MPConsumed:     req.MPCost,
		EnergyConsumed: req.EnergyCost,
		Effects:        result,
	}
}

// ─── Passive skill helpers ─────────────────────────────────────────────────────

// PassiveBonus accumulates stat bonuses from a list of passive skill IDs.
// The registry maps skillID → stat name → delta.
type PassiveBonus struct {
	AttackBonus  int
	DefenseBonus int
	HPBonus      int
	MPBonus      int
	SpeedBonus   int
	CritChance   int // extra % crit chance
}

// PassiveRegistry maps skillID to its passive bonus contribution.
// Populated at startup from the RPG module (passives.go).
var PassiveRegistry = map[string]PassiveBonus{}

// AccumulatePassives sums all passive bonuses for the given skill IDs.
func AccumulatePassives(learnedSkillIDs []string) PassiveBonus {
	total := PassiveBonus{}
	for _, id := range learnedSkillIDs {
		b, ok := PassiveRegistry[id]
		if !ok {
			continue
		}
		total.AttackBonus += b.AttackBonus
		total.DefenseBonus += b.DefenseBonus
		total.HPBonus += b.HPBonus
		total.MPBonus += b.MPBonus
		total.SpeedBonus += b.SpeedBonus
		total.CritChance += b.CritChance
	}
	return total
}
