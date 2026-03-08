package rpgdata

// SkillNode is a single node in a class skill tree.
// It holds the ID of the skill (matches AllSkills keys) plus the IDs of
// the skills that must be unlocked before this one can be learned.
type SkillNode struct {
	SkillID  string
	Requires []string // IDs of prerequisite skills (empty = root node)
}

// Branch is one thematic branch within a class skill tree.
type Branch struct {
	ID    string // matches the branch field on individual skills
	Name  string
	Emoji string
	Nodes []SkillNode // ordered T1 → T4
}

// SkillTree represents the full skill tree for one class.
type SkillTree struct {
	ClassID  string
	Branches []Branch
}

// AllSkillTrees holds the skill tree definition for every class.
// The Requires chains are derived automatically from the branch order:
// each tier-N skill requires the tier-(N-1) skill in the same branch.
var AllSkillTrees = buildSkillTrees()

// ─── Tree builder ─────────────────────────────────────────────────────────────

// classBranchDefs maps classID → branch definitions (id, name, emoji).
// Skills within each branch are retrieved from AllSkills by the naming
// convention "<classID>_<branch>_t<tier>" set in skills.go.
var classBranchDefs = map[string][3][3]string{
	// classID: {{branchID, name, emoji}, ...}
	"warrior": {
		{"protetor", "Protetor", "🛡️"},
		{"berserker", "Berserker", "🪓"},
		{"duelista", "Duelista", "⚔️"},
	},
	"paladin": {
		{"sagrado", "Sagrado", "✨"},
		{"protecao", "Proteção", "🛡️"},
		{"redencao", "Redenção", "🕊️"},
	},
	"barbarian": {
		{"furia", "Fúria", "💢"},
		{"selvagem", "Selvagem", "🐗"},
		{"resistencia", "Resistência", "🪨"},
	},
	"rogue": {
		{"assassino", "Assassino", "🗡️"},
		{"esquiva", "Esquiva", "💨"},
		{"veneno", "Veneno", "🧪"},
	},
	"mage": {
		{"elementalista", "Elementalista", "🔥"},
		{"arcanista", "Arcanista", "🔮"},
		{"ilusionista", "Ilusionista", "🌀"},
	},
	"cleric": {
		{"curandeiro", "Curandeiro", "💚"},
		{"sagrado", "Sagrado", "☀️"},
		{"protetor", "Protetor", "🛡️"},
	},
	"bard": {
		{"inspiracao", "Inspiração", "🎵"},
		{"encantamento", "Encantamento", "💫"},
		{"performance", "Performance", "🎭"},
	},
	"archer": {
		{"atirador", "Atirador", "🎯"},
		{"rastreador", "Rastreador", "🐾"},
		{"armadilha", "Armadilha", "🪤"},
	},
	"druid": {
		{"natureza", "Natureza", "🌿"},
		{"metamorfose", "Metamorfose", "🦋"},
		{"cura_natural", "Cura Natural", "🌸"},
	},
	"necromancer": {
		{"morte", "Morte", "💀"},
		{"dreno", "Dreno", "🩸"},
		{"sombra", "Sombra", "👤"},
	},
}

func buildSkillTrees() map[string]SkillTree {
	trees := make(map[string]SkillTree, len(classBranchDefs))
	for classID, branchDefs := range classBranchDefs {
		branches := make([]Branch, 0, 3)
		for _, bd := range branchDefs {
			branchID, branchName, branchEmoji := bd[0], bd[1], bd[2]
			nodes := make([]SkillNode, 0, 4)
			var prevID string
			for tier := 1; tier <= 4; tier++ {
				skillID := FormatSkillID(classID, branchID, tier)
				req := []string{}
				if prevID != "" {
					req = []string{prevID}
				}
				nodes = append(nodes, SkillNode{
					SkillID:  skillID,
					Requires: req,
				})
				prevID = skillID
			}
			branches = append(branches, Branch{
				ID:    branchID,
				Name:  branchName,
				Emoji: branchEmoji,
				Nodes: nodes,
			})
		}
		trees[classID] = SkillTree{
			ClassID:  classID,
			Branches: branches,
		}
	}
	return trees
}

// ─── Query helpers ────────────────────────────────────────────────────────────

// TreeForClass returns the SkillTree for the given classID, or a zero value if
// the class has no registered tree.
func TreeForClass(classID string) (SkillTree, bool) {
	t, ok := AllSkillTrees[classID]
	return t, ok
}

// BranchNodes returns the ordered slice of SkillNodes for a specific branch
// within a class tree.  Returns nil if the branch does not exist.
func BranchNodes(classID, branchID string) []SkillNode {
	tree, ok := AllSkillTrees[classID]
	if !ok {
		return nil
	}
	for _, b := range tree.Branches {
		if b.ID == branchID {
			return b.Nodes
		}
	}
	return nil
}

// CanUnlock reports whether a player may unlock the given skill given the set
// of already-unlocked skill IDs.  A skill is unlockable if all its Requires
// entries are present in unlocked.
func CanUnlock(skillID string, unlocked map[string]bool) bool {
	// Find the node across all trees
	for _, tree := range AllSkillTrees {
		for _, branch := range tree.Branches {
			for _, node := range branch.Nodes {
				if node.SkillID == skillID {
					for _, req := range node.Requires {
						if !unlocked[req] {
							return false
						}
					}
					return true
				}
			}
		}
	}
	// Not found in any tree — allow (fallback for hand-placed skills)
	return true
}
