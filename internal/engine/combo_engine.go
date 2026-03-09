package engine

import (
	"sync"
	"time"
)

// ─── Combo definitions ────────────────────────────────────────────────────────

// ComboStep is one skill in a sequential combo chain.
type ComboStep struct {
	SkillID    string
	TimeWindow int // seconds allowed between this step and the next (0 = unlimited)
}

// ComboDef describes a full combo chain with its bonus effect on completion.
type ComboDef struct {
	ID                   string
	Name                 string
	Emoji                string
	Description          string
	ClassID              string  // "" = any class
	Steps                []ComboStep
	BonusEffect          Effect  // bonus effect triggered on completion
	BonusDamageMultiplier float64 // multiplier applied to the final step's damage (e.g. 1.5 = +50%)
}

// ─── Combo tracker ────────────────────────────────────────────────────────────

// ComboProgress tracks a character's current position in an active combo chain.
type ComboProgress struct {
	ComboID     string
	CurrentStep int       // next step index to match
	LastSkillAt time.Time
}

// ComboEngine tracks combo states per character and manages all combo definitions.
type ComboEngine struct {
	mu     sync.RWMutex
	states map[int]*ComboProgress // charID → progress
	combos []*ComboDef
	// index maps first-step skillID → list of combo IDs that start with it.
	index map[string][]string
}

// GlobalCombos is the package-level combo engine singleton.
var GlobalCombos = &ComboEngine{
	states: make(map[int]*ComboProgress),
	index:  make(map[string][]string),
}

func init() {
	for _, c := range defaultCombos() {
		GlobalCombos.Register(c)
	}
}

// Register adds a ComboDef to the engine and updates the first-step index.
func (ce *ComboEngine) Register(c *ComboDef) {
	ce.mu.Lock()
	defer ce.mu.Unlock()
	ce.combos = append(ce.combos, c)
	if len(c.Steps) > 0 {
		first := c.Steps[0].SkillID
		ce.index[first] = append(ce.index[first], c.ID)
	}
}

// RecordSkill records a skill use for charID and returns the triggered ComboDef if
// any combo was completed. Returns nil when no combo finished.
func (ce *ComboEngine) RecordSkill(charID int, skillID string) *ComboDef {
	ce.mu.Lock()
	defer ce.mu.Unlock()

	now := time.Now()
	prog := ce.states[charID]

	// Try to advance existing combo progress.
	if prog != nil {
		combo := ce.findCombo(prog.ComboID)
		if combo != nil {
			step := combo.Steps[prog.CurrentStep]
			// Check time window
			if step.TimeWindow > 0 && now.Sub(prog.LastSkillAt) > time.Duration(step.TimeWindow)*time.Second {
				// Timed out — reset progress
				delete(ce.states, charID)
				prog = nil
			} else if step.SkillID == skillID {
				prog.CurrentStep++
				prog.LastSkillAt = now
				// Combo complete?
				if prog.CurrentStep >= len(combo.Steps) {
					delete(ce.states, charID)
					return combo
				}
				ce.states[charID] = prog
				return nil
			} else {
				// Wrong skill — reset and fall through to check new chains
				delete(ce.states, charID)
				prog = nil
			}
		}
	}

	// Check if skillID starts any new combo chain.
	ids, ok := ce.index[skillID]
	if !ok {
		return nil
	}
	// Pick the first matching combo that starts with this skill.
	for _, id := range ids {
		combo := ce.findCombo(id)
		if combo == nil || len(combo.Steps) == 0 {
			continue
		}
		if combo.Steps[0].SkillID != skillID {
			continue
		}
		// Single-step combo? Complete immediately.
		if len(combo.Steps) == 1 {
			return combo
		}
		ce.states[charID] = &ComboProgress{
			ComboID:     id,
			CurrentStep: 1, // we've matched step 0
			LastSkillAt: now,
		}
		return nil
	}
	return nil
}

// GetProgress returns the current combo progress for a character (may be nil).
func (ce *ComboEngine) GetProgress(charID int) *ComboProgress {
	ce.mu.RLock()
	defer ce.mu.RUnlock()
	return ce.states[charID]
}

// Reset cancels any active combo chain for a character.
func (ce *ComboEngine) Reset(charID int) {
	ce.mu.Lock()
	defer ce.mu.Unlock()
	delete(ce.states, charID)
}

// AllCombos returns a copy of all registered combos.
func (ce *ComboEngine) AllCombos() []*ComboDef {
	ce.mu.RLock()
	defer ce.mu.RUnlock()
	out := make([]*ComboDef, len(ce.combos))
	copy(out, ce.combos)
	return out
}

// CombosByClass returns combos valid for a given classID (including universal "").
func (ce *ComboEngine) CombosByClass(classID string) []*ComboDef {
	ce.mu.RLock()
	defer ce.mu.RUnlock()
	var out []*ComboDef
	for _, c := range ce.combos {
		if c.ClassID == "" || c.ClassID == classID {
			out = append(out, c)
		}
	}
	return out
}

func (ce *ComboEngine) findCombo(id string) *ComboDef {
	for _, c := range ce.combos {
		if c.ID == id {
			return c
		}
	}
	return nil
}

// ─── Default combo definitions ────────────────────────────────────────────────

func defaultCombos() []*ComboDef {
	return []*ComboDef{
		// ── Warrior combos ──
		{
			ID: "warrior_iron_trinity", Name: "Trindade de Ferro", Emoji: "⚔️",
			Description: "Golpe Firme → Corte Poderoso → Executar: dano triplo com destruição de armadura.",
			ClassID:     "warrior",
			Steps: []ComboStep{
				{SkillID: "warr_firm_strike", TimeWindow: 30},
				{SkillID: "warr_power_slash", TimeWindow: 30},
				{SkillID: "warr_execute", TimeWindow: 30},
			},
			BonusEffect:           Effect{Type: EffectDamage, Element: ElementPhysical, BasePower: 80},
			BonusDamageMultiplier: 1.5,
		},
		{
			ID: "warrior_shield_wall", Name: "Muralha de Escudo", Emoji: "🛡️",
			Description: "Pancada de Escudo → Baluarte → Contra-Golpe: perfeita defesa que pune o atacante.",
			ClassID:     "warrior",
			Steps: []ComboStep{
				{SkillID: "warr_shield_bash", TimeWindow: 30},
				{SkillID: "warr_bulwark", TimeWindow: 30},
				{SkillID: "warr_counter_strike", TimeWindow: 30},
			},
			BonusEffect:           Effect{Type: EffectApplyStatus, StatusKind: StatusStun, StatusTurns: 2},
			BonusDamageMultiplier: 1.3,
		},
		{
			ID: "warrior_blade_storm_combo", Name: "Combo Tempestade", Emoji: "🌪️",
			Description: "Ataque Rápido → Redemoinho → Tempestade de Lâminas: destruição total em AoE.",
			ClassID:     "warrior",
			Steps: []ComboStep{
				{SkillID: "warr_quick_strike", TimeWindow: 30},
				{SkillID: "warr_whirlwind", TimeWindow: 30},
				{SkillID: "warr_bladestorm", TimeWindow: 30},
			},
			BonusEffect:           Effect{Type: EffectAoE, Element: ElementPhysical, BasePower: 60},
			BonusDamageMultiplier: 1.4,
		},

		// ── Mage combos ──
		{
			ID: "mage_tri_elemental", Name: "Tri-Elemental", Emoji: "🌈",
			Description: "Bola de Fogo → Raio → Lança de Gelo: cobertura total de elementos com dano massivo.",
			ClassID:     "mage",
			Steps: []ComboStep{
				{SkillID: "mage_fireball", TimeWindow: 30},
				{SkillID: "mage_lightning_bolt", TimeWindow: 30},
				{SkillID: "mage_ice_lance", TimeWindow: 30},
			},
			BonusEffect:           Effect{Type: EffectAoE, Element: ElementMagic, BasePower: 70},
			BonusDamageMultiplier: 1.6,
		},
		{
			ID: "mage_arcane_cascade", Name: "Cascata Arcana", Emoji: "✨",
			Description: "Míssil Mágico → Torrente Arcana → Singularidade: três camadas de dano arcano.",
			ClassID:     "mage",
			Steps: []ComboStep{
				{SkillID: "mage_magic_missile", TimeWindow: 30},
				{SkillID: "mage_arcane_torrent", TimeWindow: 30},
				{SkillID: "mage_singularity", TimeWindow: 30},
			},
			BonusEffect:           Effect{Type: EffectDamage, Element: ElementMagic, BasePower: 90},
			BonusDamageMultiplier: 1.5,
		},
		{
			ID: "mage_time_assault", Name: "Assalto Temporal", Emoji: "⏰",
			Description: "Distorção Temporal → Explosão de Força → Aniquilação Arcana: ataque duplo devastador.",
			ClassID:     "mage",
			Steps: []ComboStep{
				{SkillID: "mage_time_warp", TimeWindow: 30},
				{SkillID: "mage_force_blast", TimeWindow: 30},
				{SkillID: "mage_arcane_annihilation", TimeWindow: 30},
			},
			BonusEffect:           Effect{Type: EffectApplyStatus, StatusKind: StatusSilence, StatusTurns: 3},
			BonusDamageMultiplier: 1.7,
		},

		// ── Rogue combos ──
		{
			ID: "rogue_assassin_art", Name: "Arte do Assassino", Emoji: "🗡️",
			Description: "Ataque Pelas Costas → Marca da Morte → Assassinar: arte letal perfeita.",
			ClassID:     "rogue",
			Steps: []ComboStep{
				{SkillID: "rog_backstab", TimeWindow: 30},
				{SkillID: "rog_death_mark", TimeWindow: 30},
				{SkillID: "rog_assassinate", TimeWindow: 30},
			},
			BonusEffect:           Effect{Type: EffectDamage, Element: ElementPhysical, BasePower: 100},
			BonusDamageMultiplier: 2.0,
		},
		{
			ID: "rogue_poison_mastery", Name: "Mestre dos Venenos", Emoji: "☠️",
			Description: "Lâmina Envenenada → Toxina → Toxina Viral: venenos que se acumulam ao extremo.",
			ClassID:     "rogue",
			Steps: []ComboStep{
				{SkillID: "rog_poison_blade", TimeWindow: 30},
				{SkillID: "rog_toxin", TimeWindow: 30},
				{SkillID: "rog_viral_toxin", TimeWindow: 30},
			},
			BonusEffect:           Effect{Type: EffectApplyStatus, StatusKind: StatusPoison, StatusTurns: 8, StatusDmgPT: 20},
			BonusDamageMultiplier: 1.5,
		},
		{
			ID: "rogue_shadow_dance", Name: "Dança das Sombras", Emoji: "🌑",
			Description: "Passo nas Sombras → Clone de Sombra → Golpe das Sombras: dança mortal.",
			ClassID:     "rogue",
			Steps: []ComboStep{
				{SkillID: "rog_shadow_step", TimeWindow: 30},
				{SkillID: "rog_shadow_clone", TimeWindow: 30},
				{SkillID: "rog_shadow_strike", TimeWindow: 30},
			},
			BonusEffect:           Effect{Type: EffectDamage, Element: ElementPhysical, BasePower: 75},
			BonusDamageMultiplier: 1.8,
		},

		// ── Archer combos ──
		{
			ID: "archer_rain_of_arrows", Name: "Chuva de Flechas", Emoji: "🌧️",
			Description: "Tiro Mirado → Tiro Múltiplo → Bombardeio: dilúvio de flechas imparável.",
			ClassID:     "archer",
			Steps: []ComboStep{
				{SkillID: "arch_aimed_shot", TimeWindow: 30},
				{SkillID: "arch_multi_shot", TimeWindow: 30},
				{SkillID: "arch_barrage", TimeWindow: 30},
			},
			BonusEffect:           Effect{Type: EffectAoE, Element: ElementPhysical, BasePower: 60},
			BonusDamageMultiplier: 1.6,
		},
		{
			ID: "archer_nature_wrath_combo", Name: "Ira da Natureza", Emoji: "🌿",
			Description: "Flecha da Natureza → Enredar → Ira da Natureza: poder natural desencadeado.",
			ClassID:     "archer",
			Steps: []ComboStep{
				{SkillID: "arch_nature_arrow", TimeWindow: 30},
				{SkillID: "arch_entangle", TimeWindow: 30},
				{SkillID: "arch_nature_wrath", TimeWindow: 30},
			},
			BonusEffect:           Effect{Type: EffectApplyStatus, StatusKind: StatusPoison, StatusTurns: 5, StatusDmgPT: 10},
			BonusDamageMultiplier: 1.5,
		},
		{
			ID: "archer_perfect_kill", Name: "Abate Perfeito", Emoji: "🎯",
			Description: "Marca do Caçador → Tiro na Cabeça → Tiro Fatal: sequência de abate perfeita.",
			ClassID:     "archer",
			Steps: []ComboStep{
				{SkillID: "arch_hunter_mark", TimeWindow: 30},
				{SkillID: "arch_headshot", TimeWindow: 30},
				{SkillID: "arch_kill_shot", TimeWindow: 30},
			},
			BonusEffect:           Effect{Type: EffectDamage, Element: ElementPhysical, BasePower: 95},
			BonusDamageMultiplier: 2.0,
		},

		// ── Paladin combos ──
		{
			ID: "paladin_divine_retribution", Name: "Retribuição Divina", Emoji: "⚖️",
			Description: "Marca da Justiça → Ira Sagrada → Julgamento Divino: punição divina absoluta.",
			ClassID:     "paladin",
			Steps: []ComboStep{
				{SkillID: "pal_mark_of_justice", TimeWindow: 30},
				{SkillID: "pal_holy_wrath", TimeWindow: 30},
				{SkillID: "pal_divine_judgment", TimeWindow: 30},
			},
			BonusEffect:           Effect{Type: EffectDamage, Element: ElementHoly, BasePower: 100},
			BonusDamageMultiplier: 1.8,
		},
		{
			ID: "paladin_holy_fortress", Name: "Fortaleza Sagrada", Emoji: "🏰",
			Description: "Abençoar → Égide Divina → Escudo Divino: defesa sagrada impenetrável.",
			ClassID:     "paladin",
			Steps: []ComboStep{
				{SkillID: "pal_bless", TimeWindow: 30},
				{SkillID: "pal_divine_aegis", TimeWindow: 30},
				{SkillID: "pal_divine_shield", TimeWindow: 30},
			},
			BonusEffect:           Effect{Type: EffectApplyStatus, StatusKind: StatusProtect, StatusTurns: 5},
			BonusDamageMultiplier: 1.0,
		},

		// ── Cleric combos ──
		{
			ID: "cleric_light_of_salvation", Name: "Luz da Salvação", Emoji: "🌟",
			Description: "Cura → Bênção → Milagre: restauração completa guiada pela luz.",
			ClassID:     "cleric",
			Steps: []ComboStep{
				{SkillID: "cler_heal", TimeWindow: 30},
				{SkillID: "cler_blessing", TimeWindow: 30},
				{SkillID: "cler_miracle", TimeWindow: 30},
			},
			BonusEffect:           Effect{Type: EffectHeal, BasePower: 80},
			BonusDamageMultiplier: 1.0,
		},
		{
			ID: "cleric_divine_wrath", Name: "Ira Divina", Emoji: "☀️",
			Description: "Punição Sagrada → Fogo Divino → Avatar da Luz: poder divino em sua forma mais pura.",
			ClassID:     "cleric",
			Steps: []ComboStep{
				{SkillID: "cler_holy_smite", TimeWindow: 30},
				{SkillID: "cler_divine_fire", TimeWindow: 30},
				{SkillID: "cler_avatar_of_light", TimeWindow: 30},
			},
			BonusEffect:           Effect{Type: EffectAoE, Element: ElementHoly, BasePower: 90},
			BonusDamageMultiplier: 1.6,
		},

		// ── Barbarian combos ──
		{
			ID: "barbarian_blood_frenzy_combo", Name: "Frenesi de Sangue", Emoji: "💢",
			Description: "Fúria → Sede de Sangue → Devastação: espiral de fúria e cura.",
			ClassID:     "barbarian",
			Steps: []ComboStep{
				{SkillID: "barb_rage", TimeWindow: 30},
				{SkillID: "barb_blood_thirst", TimeWindow: 30},
				{SkillID: "barb_rampage", TimeWindow: 30},
			},
			BonusEffect:           Effect{Type: EffectApplyStatus, StatusKind: StatusBerserk, StatusTurns: 5},
			BonusDamageMultiplier: 1.7,
		},
		{
			ID: "barbarian_tribal_fury", Name: "Fúria Tribal", Emoji: "🥁",
			Description: "Tambores Tribais → Ritual de Sangue → Impacto do Titã: chamado ancestral devastador.",
			ClassID:     "barbarian",
			Steps: []ComboStep{
				{SkillID: "barb_tribal_drums", TimeWindow: 30},
				{SkillID: "barb_blood_ritual", TimeWindow: 30},
				{SkillID: "barb_titan_slam", TimeWindow: 30},
			},
			BonusEffect:           Effect{Type: EffectAoE, Element: ElementPhysical, BasePower: 120},
			BonusDamageMultiplier: 1.8,
		},

		// ── Bard combos ──
		{
			ID: "bard_symphony_of_destruction", Name: "Sinfonia da Destruição", Emoji: "🎼",
			Description: "Hino de Batalha → Discordância → Final da Ópera: música que destrói.",
			ClassID:     "bard",
			Steps: []ComboStep{
				{SkillID: "bard_battle_hymn", TimeWindow: 30},
				{SkillID: "bard_discordance", TimeWindow: 30},
				{SkillID: "bard_finale", TimeWindow: 30},
			},
			BonusEffect:           Effect{Type: EffectAoE, Element: ElementDark, BasePower: 80},
			BonusDamageMultiplier: 1.7,
		},
		{
			ID: "bard_grand_performance", Name: "Grande Apresentação", Emoji: "🎭",
			Description: "Cantiga do Sono → Pesadelo → Grande Ilusão: domínio total da mente.",
			ClassID:     "bard",
			Steps: []ComboStep{
				{SkillID: "bard_lullaby", TimeWindow: 30},
				{SkillID: "bard_nightmare", TimeWindow: 30},
				{SkillID: "bard_grand_illusion", TimeWindow: 30},
			},
			BonusEffect:           Effect{Type: EffectApplyStatus, StatusKind: StatusStun, StatusTurns: 4},
			BonusDamageMultiplier: 1.6,
		},

	}
}
