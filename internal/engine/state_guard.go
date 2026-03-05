package engine

// StateEngine centralizes state transition validation for gameplay flows.
// It is intentionally permissive for unknown states to preserve backward compatibility.
type StateEngine struct {
	allowed map[string]map[string]bool
}

func NewStateEngine() *StateEngine {
	return &StateEngine{
		allowed: map[string]map[string]bool{
			"idle":           {"combat": true, "dungeon": true, "dungeon_combat": true, "auto_hunt": true},
			"combat":         {"idle": true, "dungeon_combat": true},
			"dungeon":        {"idle": true, "dungeon_combat": true},
			"dungeon_combat": {"idle": true, "dungeon": true},
			"auto_hunt":      {"idle": true},
		},
	}
}

func (e *StateEngine) CanTransition(from, to string) bool {
	if from == to {
		return true
	}
	next, ok := e.allowed[from]
	if !ok {
		return true
	}
	return next[to]
}
