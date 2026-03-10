package rpg

// ─── Skill role ───────────────────────────────────────────────────────────────

// SkillRole classifies the mechanical purpose of a skill.
// Every skill must belong to exactly one role so the validator and UI can
// detect mechanically-identical skills and categorise them for the player.
type SkillRole string

const (
	RoleDirect  SkillRole = "DIRECT_DAMAGE" // single-target burst
	RoleAoE     SkillRole = "AOE"           // area-of-effect damage
	RoleDoT     SkillRole = "DOT"           // damage-over-time
	RoleBuff    SkillRole = "BUFF"          // self or ally stat increase
	RoleDebuff  SkillRole = "DEBUFF"        // enemy stat reduction
	RoleControl SkillRole = "CONTROL"       // stun / freeze / silence
	RoleHeal    SkillRole = "HEAL"          // restore HP / remove status
	RoleUtility SkillRole = "UTILITY"       // mobility, escape, resource
	RoleSummon  SkillRole = "SUMMON"        // spawn minions / totems
	RolePassive SkillRole = "PASSIVE"       // permanent stat bonus
)

// ─── Skill tree node ──────────────────────────────────────────────────────────

// SkillNode represents one ability in the class skill tree.
type SkillNode struct {
	ID          string
	Name        string
	Emoji       string
	Description string
	Class       string
	Branch      string    // specialization branch within the class
	Tier        int       // 1 = basic, 4 = ultimate
	PointCost   int       // skill points required to unlock
	Role        SkillRole // mechanical classification (required for validator)

	// Combat properties.
	MPCost     int
	EnergyCost int
	Effects    []engine_Effect // resolved at engine layer

	RequiredLevel int
	Requires      []string // prerequisite skill IDs
	IsPassive     bool
	IsUltimate    bool // true for tier-4 capstone skills
}

// engine_Effect is a local alias (avoids import cycles at this layer).
// The actual engine.Effect is created when the service layer resolves the skill.
type engine_Effect struct {
	TypeName    string // "damage" | "heal" | "apply_status" | "stat_buff" | "stat_debuff" | etc.
	Element     string
	BasePower   int
	StatusKind  string
	StatusTurns int
	StatusDmgPT int
	StatName    string // for stat_buff / stat_debuff: "attack", "defense", "speed", etc.
	StatDelta   int    // amount to change the stat (negative = debuff)
}

// ─── Skill tree ───────────────────────────────────────────────────────────────

// SkillTree holds the full skill tree for one class, organised by branch.
type SkillTree struct {
	ClassID  string
	Branches []Branch
}

// Branch is a specialization path within a class.
type Branch struct {
	ID    string
	Name  string
	Emoji string
	Nodes []SkillNode
}

// Get returns a node by ID.
func (st *SkillTree) Get(skillID string) (SkillNode, bool) {
	for _, b := range st.Branches {
		for _, n := range b.Nodes {
			if n.ID == skillID {
				return n, true
			}
		}
	}
	return SkillNode{}, false
}

// Available returns nodes the character can unlock given their learned skills and level.
func (st *SkillTree) Available(learnedIDs []string, characterLevel int) []SkillNode {
	learned := make(map[string]bool, len(learnedIDs))
	for _, id := range learnedIDs {
		learned[id] = true
	}
	var out []SkillNode
	for _, b := range st.Branches {
		for _, n := range b.Nodes {
			if learned[n.ID] {
				continue
			}
			if characterLevel < n.RequiredLevel {
				continue
			}
			prereqsMet := true
			for _, req := range n.Requires {
				if !learned[req] {
					prereqsMet = false
					break
				}
			}
			if prereqsMet {
				out = append(out, n)
			}
		}
	}
	return out
}

// ─── Tree registry ────────────────────────────────────────────────────────────

// Trees holds all class skill trees.
var Trees = map[string]*SkillTree{}

func init() {
	Trees["barbarian"] = barbarianTree()
	Trees["paladin"] = paladinTree()
	Trees["cleric"] = clericTree()
	Trees["bard"] = bardTree()
}

func barbarianTree() *SkillTree {
	return &SkillTree{
		ClassID: "barbarian",
		Branches: []Branch{
			{
				ID: "furia", Name: "Fúria", Emoji: "💢",
				Nodes: []SkillNode{
					{ID: "barb_rage", Class: "barbarian", Branch: "furia", Tier: 1, PointCost: 1,
						Name: "Fúria", Emoji: "💢", RequiredLevel: 1,
						MPCost: 0, EnergyCost: 5,
						Description: "Entra em Berserk por 3 turnos: +60% ATK, -30% DEF.",
					},
					{ID: "barb_blood_thirst", Class: "barbarian", Branch: "furia", Tier: 2, PointCost: 1,
						Name: "Sede de Sangue", Emoji: "🩸", RequiredLevel: 5, Requires: []string{"barb_rage"},
						MPCost: 0, EnergyCost: 8,
						Description: "Cura 20% do dano causado neste turno.",
					},
					{ID: "barb_death_cry", Class: "barbarian", Branch: "furia", Tier: 3, PointCost: 2,
						Name: "Grito Mortal", Emoji: "💀", RequiredLevel: 15, Requires: []string{"barb_blood_thirst"},
						MPCost: 10, EnergyCost: 10,
						Description: "Atordoa o inimigo por 1 turno; causa 2d8+FOR de dano.",
					},
					{ID: "barb_rampage", Class: "barbarian", Branch: "furia", Tier: 4, PointCost: 3,
						Name: "Devastação", Emoji: "🌪️", RequiredLevel: 30, Requires: []string{"barb_death_cry"},
						MPCost: 20, EnergyCost: 15, IsUltimate: true,
						Description: "Ataca 3 vezes. Cada acerto ignora 25% da armadura do alvo.",
					},
				},
			},
			{
				ID: "resistencia", Name: "Resistência", Emoji: "🛡️",
				Nodes: []SkillNode{
					{ID: "barb_toughness", Class: "barbarian", Branch: "resistencia", Tier: 1, PointCost: 1,
						Name: "Resistência", Emoji: "🛡️", RequiredLevel: 1, IsPassive: true,
						Description: "+15 HP máximo permanente.",
					},
					{ID: "barb_stone_skin", Class: "barbarian", Branch: "resistencia", Tier: 2, PointCost: 1,
						Name: "Pele de Pedra", Emoji: "🪨", RequiredLevel: 8, Requires: []string{"barb_toughness"},
						MPCost: 0, IsPassive: true,
						Description: "-10% dano físico recebido permanente.",
					},
				},
			},
		},
	}
}

func paladinTree() *SkillTree {
	return &SkillTree{
		ClassID: "paladin",
		Branches: []Branch{
			{
				ID: "sagrado", Name: "Sagrado", Emoji: "⚜️",
				Nodes: []SkillNode{
					{ID: "pal_smite", Class: "paladin", Branch: "sagrado", Tier: 1, PointCost: 1,
						Name: "Golpe Divino", Emoji: "✨", RequiredLevel: 1,
						MPCost: 10, EnergyCost: 5,
						Description: "Adiciona 1d8 de dano sagrado ao próximo ataque.",
					},
					{ID: "pal_lay_on_hands", Class: "paladin", Branch: "sagrado", Tier: 2, PointCost: 1,
						Name: "Imposição de Mãos", Emoji: "🙏", RequiredLevel: 5, Requires: []string{"pal_smite"},
						MPCost: 20, EnergyCost: 5,
						Description: "Cura 3d8+SAB de HP.",
					},
					{ID: "pal_holy_shield", Class: "paladin", Branch: "sagrado", Tier: 3, PointCost: 2,
						Name: "Escudo Sagrado", Emoji: "🛡️", RequiredLevel: 15, Requires: []string{"pal_lay_on_hands"},
						MPCost: 30, EnergyCost: 5,
						Description: "Cria escudo que absorve 50% do dano por 2 turnos.",
					},
					{ID: "pal_divine_judgment", Class: "paladin", Branch: "sagrado", Tier: 4, PointCost: 3,
						Name: "Julgamento Divino", Emoji: "☀️", RequiredLevel: 30, Requires: []string{"pal_holy_shield"},
						MPCost: 50, EnergyCost: 10, IsUltimate: true,
						Description: "Dano sagrado massivo: 5d10+SAB. Dobrado contra mortos-vivos e demônios.",
					},
				},
			},
		},
	}
}

func clericTree() *SkillTree {
	return &SkillTree{
		ClassID: "cleric",
		Branches: []Branch{
			{
				ID: "cura", Name: "Cura", Emoji: "💚",
				Nodes: []SkillNode{
					{ID: "cler_heal", Class: "cleric", Branch: "cura", Tier: 1, PointCost: 1,
						Name: "Cura", Emoji: "💚", RequiredLevel: 1,
						MPCost: 15, EnergyCost: 5,
						Description: "Cura 2d8+SAB HP.",
					},
					{ID: "cler_mass_heal", Class: "cleric", Branch: "cura", Tier: 2, PointCost: 1,
						Name: "Cura em Área", Emoji: "💖", RequiredLevel: 8, Requires: []string{"cler_heal"},
						MPCost: 30, EnergyCost: 8,
						Description: "Cura todos os aliados por 2d6+SAB.",
					},
					{ID: "cler_resurr", Class: "cleric", Branch: "cura", Tier: 4, PointCost: 3,
						Name: "Ressurreição", Emoji: "🌟", RequiredLevel: 40, Requires: []string{"cler_mass_heal"},
						MPCost: 80, EnergyCost: 20, IsUltimate: true,
						Description: "Ressuscita um aliado caído com 50% HP. Cooldown: 1 hora.",
					},
				},
			},
		},
	}
}

func bardTree() *SkillTree {
	return &SkillTree{
		ClassID: "bard",
		Branches: []Branch{
			{
				ID: "musica", Name: "Música", Emoji: "🎵",
				Nodes: []SkillNode{
					{ID: "bard_inspire", Class: "bard", Branch: "musica", Tier: 1, PointCost: 1,
						Name: "Inspiração", Emoji: "🎵", RequiredLevel: 1,
						MPCost: 10, EnergyCost: 5,
						Description: "+20% ATK para o próximo ataque de um aliado.",
					},
					{ID: "bard_lullaby", Class: "bard", Branch: "musica", Tier: 2, PointCost: 1,
						Name: "Cantiga do Sono", Emoji: "😴", RequiredLevel: 6, Requires: []string{"bard_inspire"},
						MPCost: 20, EnergyCost: 8,
						Description: "Aplica Stun no inimigo por 1 turno (resistência: SAB).",
					},
					{ID: "bard_symphony", Class: "bard", Branch: "musica", Tier: 4, PointCost: 3,
						Name: "Sinfonia da Vitória", Emoji: "🎼", RequiredLevel: 35, Requires: []string{"bard_lullaby"},
						MPCost: 60, EnergyCost: 15, IsUltimate: true,
						Description: "Toda a party ganha +30% ATK e +20% DEF por 5 turnos.",
					},
				},
			},
		},
	}
}
