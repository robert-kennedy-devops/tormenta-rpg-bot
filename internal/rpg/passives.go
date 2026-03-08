package rpg

import "github.com/tormenta-bot/internal/engine"

// ─── Passive ability registry ─────────────────────────────────────────────────
//
// Each passive registered here will be picked up by engine.AccumulatePassives().
// Key = traitKey from RaceDef.TraitKey or ClassDef.TraitKey (or any skill ID).

func init() {
	// ── Race passives ──────────────────────────────────────────────────────
	engine.PassiveRegistry["human_versatile"] = engine.PassiveBonus{} // XP handled in game layer
	engine.PassiveRegistry["elf_arcane_sight"] = engine.PassiveBonus{} // magic% handled in damage calc
	engine.PassiveRegistry["dwarf_stonekin"] = engine.PassiveBonus{DefenseBonus: 3}
	engine.PassiveRegistry["halforc_blood_fury"] = engine.PassiveBonus{AttackBonus: 4}
	engine.PassiveRegistry["goblin_sneaky"] = engine.PassiveBonus{CritChance: 15}
	engine.PassiveRegistry["qareen_flame_soul"] = engine.PassiveBonus{} // element handled in damage calc
	engine.PassiveRegistry["minotaur_rampage"] = engine.PassiveBonus{AttackBonus: 5}

	// ── Class passives ─────────────────────────────────────────────────────
	engine.PassiveRegistry["warrior_weapon_mastery"] = engine.PassiveBonus{AttackBonus: 3}
	engine.PassiveRegistry["mage_arcane_surge"] = engine.PassiveBonus{MPBonus: 15}
	engine.PassiveRegistry["rogue_backstab"] = engine.PassiveBonus{CritChance: 5}
	engine.PassiveRegistry["archer_eagle_eye"] = engine.PassiveBonus{CritChance: 10}
	engine.PassiveRegistry["paladin_holy_aura"] = engine.PassiveBonus{HPBonus: 20, DefenseBonus: 2}
	engine.PassiveRegistry["cleric_divine_grace"] = engine.PassiveBonus{MPBonus: 20}
	engine.PassiveRegistry["barbarian_rage"] = engine.PassiveBonus{HPBonus: 30, AttackBonus: 2}
	engine.PassiveRegistry["bard_inspiration"] = engine.PassiveBonus{MPBonus: 10, SpeedBonus: 1}
}

// ─── Passive helpers (for UI display) ─────────────────────────────────────────

// PassiveDescription returns a human-readable description of a passive key.
func PassiveDescription(traitKey string) string {
	descs := map[string]string{
		"human_versatile":    "Versátil: +10% XP ganho.",
		"elf_arcane_sight":   "Visão Élfica: +20% dano mágico.",
		"dwarf_stonekin":     "Pele de Pedra: -15% dano recebido.",
		"halforc_blood_fury": "Fúria do Sangue: +25% dano físico.",
		"goblin_sneaky":      "Furtivo: +15% chance de crítico.",
		"qareen_flame_soul":  "Alma de Chama: +30% dano de fogo; imune a Burn.",
		"minotaur_rampage":   "Devastação: ignora 10% da armadura do alvo.",
		"warrior_weapon_mastery": "Maestria em Armas: +5% dano por especialização.",
		"mage_arcane_surge":  "Surto Arcano: a cada 4 feitiços, 1 é gratuito.",
		"rogue_backstab":     "Facada pelas Costas: +40% dano em flanqueamento.",
		"archer_eagle_eye":   "Olho de Águia: +10% crit a distância.",
		"paladin_holy_aura":  "Aura Sagrada: companheiros curam +10%.",
		"cleric_divine_grace": "Graça Divina: 10% de curas duplas; imune a Curse.",
		"barbarian_rage":     "Fúria Bárbara: +60% ataque ao entrar em Berserk.",
		"bard_inspiration":   "Inspiração: party ganha +15% XP pós-batalha.",
	}
	if d, ok := descs[traitKey]; ok {
		return d
	}
	return "Habilidade passiva."
}
