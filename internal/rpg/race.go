// Package rpg defines the full RPG ruleset for Tormenta: races, classes,
// attributes, skill trees, talents and passive abilities.
// It extends (but never replaces) the existing game/data.go definitions —
// new races and classes live here; old ones are kept in game/data.go for
// backward compatibility.
package rpg

// ─── Race definition ─────────────────────────────────────────────────────────

// RaceDef is the canonical race template used by the RPG module.
// The older models.Race struct in models/models.go is still used for DB storage;
// RaceDef adds the extended mechanics (elemental affinities, trait functions).
type RaceDef struct {
	ID          string
	Name        string
	Emoji       string
	Description string

	// Base stat bonuses applied at character creation.
	BonusHP  int
	BonusMP  int
	BonusStr int
	BonusDex int
	BonusCon int
	BonusInt int
	BonusWis int
	BonusCha int

	// Passive trait key (matched in passives.go).
	TraitKey string
	TraitDesc string

	// Elemental affinities for the engine.
	Resistances []string // element IDs
	Weaknesses  []string
}

// ─── Registry ─────────────────────────────────────────────────────────────────

// Races is the master map of all playable races including new Tormenta ones.
var Races = map[string]RaceDef{

	// ── Legacy races (keep IDs consistent with game/data.go) ──────────────
	"human": {
		ID: "human", Name: "Humano", Emoji: "👤",
		Description: "Adaptáveis e versáteis. Bônus em todas as atributos.",
		BonusHP: 10, BonusMP: 5,
		BonusStr: 1, BonusDex: 1, BonusCon: 1, BonusInt: 1, BonusWis: 1, BonusCha: 2,
		TraitKey: "human_versatile", TraitDesc: "+10% XP ganho em todas as atividades.",
	},
	"elf": {
		ID: "elf", Name: "Elfo", Emoji: "🧝",
		Description: "Ágeis e conectados à magia élfica.",
		BonusHP: 5, BonusMP: 20,
		BonusDex: 3, BonusInt: 3, BonusWis: 2, BonusCha: 1,
		TraitKey: "elf_arcane_sight", TraitDesc: "+20% dano mágico; visão no escuro.",
		Resistances: []string{"magic"},
	},
	"dwarf": {
		ID: "dwarf", Name: "Anão", Emoji: "⛏️",
		Description: "Resistentes como as pedras que esculpem.",
		BonusHP: 25,
		BonusStr: 2, BonusCon: 4, BonusWis: 1,
		TraitKey: "dwarf_stonekin", TraitDesc: "-15% dano recebido; imune a Freeze.",
		Resistances: []string{"physical", "ice"},
	},
	"halforc": {
		ID: "halforc", Name: "Meio-Orc", Emoji: "👹",
		Description: "Brutal em combate, com força descomunal.",
		BonusHP: 20,
		BonusStr: 4, BonusCon: 3, BonusInt: -1, BonusCha: -1,
		TraitKey: "halforc_blood_fury", TraitDesc: "+25% dano físico; ativa Berserk abaixo de 20% HP.",
	},

	// ── New Tormenta races ─────────────────────────────────────────────────
	"goblin": {
		ID: "goblin", Name: "Goblin", Emoji: "👺",
		Description: "Pequenos e rápidos — compensam a fragilidade com esperteza e veneno.",
		BonusHP: -10, BonusMP: 10,
		BonusDex: 4, BonusInt: 2, BonusCha: -1,
		TraitKey: "goblin_sneaky", TraitDesc: "+15% chance de crítico; habilidades de veneno têm +1 turno de duração.",
		Weaknesses: []string{"holy"},
	},
	"qareen": {
		ID: "qareen", Name: "Qareen", Emoji: "🧞",
		Description: "Seres de chama e desejo — mestres da magia elemental e da persuasão.",
		BonusHP: 0, BonusMP: 30,
		BonusInt: 4, BonusCha: 4, BonusDex: 1,
		TraitKey: "qareen_flame_soul", TraitDesc: "+30% dano de Fogo; imune a Burn; habilidades custam -1 MP.",
		Resistances: []string{"fire"},
		Weaknesses:  []string{"ice"},
	},
	"minotaur": {
		ID: "minotaur", Name: "Minotauro", Emoji: "🐂",
		Description: "Guerreiros indomáveis do labirinto — força e resistência sem igual.",
		BonusHP: 40, BonusMP: -10,
		BonusStr: 5, BonusCon: 4, BonusDex: -2,
		TraitKey: "minotaur_rampage", TraitDesc: "Ataques físicos ignoram 10% da armadura do alvo; +1 dado de dano em carga.",
		Resistances: []string{"physical"},
		Weaknesses:  []string{"magic"},
	},
}

// Get returns a race by ID, plus a found flag.
func Get(id string) (RaceDef, bool) {
	r, ok := Races[id]
	return r, ok
}

// AllIDs returns all available race IDs.
func AllIDs() []string {
	ids := make([]string, 0, len(Races))
	for id := range Races {
		ids = append(ids, id)
	}
	return ids
}
