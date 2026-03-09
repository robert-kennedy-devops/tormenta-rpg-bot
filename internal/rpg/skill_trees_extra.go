package rpg

import engine "github.com/tormenta-bot/internal/engine"

// init appends the final 6 skills to reach 210 total.
// Runs after skill_trees_extended.go (alphabetical order: "extra" > "extended").
func init() {
	appendExtraSkills()
}

func appendExtraSkills() {
	// ── Archer: Natureza — 2 extra skills ────────────────────────────────────
	// (Rogue extra skills were added directly in skill_trees_rogue.go)
	if t, ok := Trees["archer"]; ok {
		for i, b := range t.Branches {
			if b.ID == "natureza" {
				t.Branches[i].Nodes = append(t.Branches[i].Nodes,
					SkillNode{
						ID: "arch_poison_rain", Class: "archer", Branch: "natureza",
						Name: "Chuva Venenosa", Emoji: "☠️", Tier: 3, PointCost: 2, RequiredLevel: 28,
						MPCost: 45, EnergyCost: 10,
						Description: "Dispara chuva de flechas envenenadas em AoE: 2d6+DEX de dano + veneno em AoE por 5 turnos (6/turno).",
						Requires: []string{"arch_entangle"},
						Effects: []engine_Effect{
							{TypeName: "aoe", Element: "poison", BasePower: 28},
							{TypeName: "apply_status", StatusKind: "poison", StatusTurns: 5, StatusDmgPT: 6},
						},
					},
					SkillNode{
						ID: "arch_gale_force", Class: "archer", Branch: "natureza",
						Name: "Força do Vendaval", Emoji: "💨", Tier: 3, PointCost: 2, RequiredLevel: 32,
						MPCost: 48, EnergyCost: 11,
						Description: "Flecha carregada de vento tempestuoso: 4d10+DEX de dano e aplica Stun + Haste no arqueiro por 2 turnos.",
						Requires: []string{"arch_storm_arrow"},
						Effects: []engine_Effect{
							{TypeName: "damage", Element: "lightning", BasePower: 52},
							{TypeName: "apply_status", StatusKind: "stun", StatusTurns: 1},
							{TypeName: "apply_status", StatusKind: "haste", StatusTurns: 2},
						},
					},
				)
				break
			}
		}
	}

	// ── Barbarian: Resistência — 2 extra skills ───────────────────────────────
	if t, ok := Trees["barbarian"]; ok {
		for i, b := range t.Branches {
			if b.ID == "resistencia" {
				t.Branches[i].Nodes = append(t.Branches[i].Nodes,
					SkillNode{
						ID: "barb_war_cry_resist", Class: "barbarian", Branch: "resistencia",
						Name: "Grito de Resistência", Emoji: "📣", Tier: 3, PointCost: 2, RequiredLevel: 35,
						MPCost: 30, EnergyCost: 10,
						Description: "Grito que fortalece a resistência: Proteção por 3 turnos e remove Stun/Freeze instantaneamente.",
						Requires: []string{"barb_battle_hardened"},
						Effects: []engine_Effect{
							{TypeName: "apply_status", StatusKind: "protect", StatusTurns: 3},
							{TypeName: "remove_status", StatusKind: "stun"},
							{TypeName: "remove_status", StatusKind: "freeze"},
						},
					},
					SkillNode{
						ID: "barb_indestructible", Class: "barbarian", Branch: "resistencia",
						Name: "Indestrutível", Emoji: "💎", Tier: 5, PointCost: 4, RequiredLevel: 75,
						IsPassive: true,
						Description: "Passivo lendário: +120 HP máximo, -20% de todo dano recebido e 5% de chance de negar completamente qualquer golpe.",
						Requires: []string{"barb_mountain_stance"},
					},
				)
				break
			}
		}
	}

	// Register new passives (rogue passives registered in skill_trees_extended.go init)
	engine.PassiveRegistry["rog_shadow_arts"] = engine.PassiveBonus{CritChance: 12, AttackBonus: 5, SpeedBonus: 2}
	engine.PassiveRegistry["rog_pickpocket"] = engine.PassiveBonus{AttackBonus: 2}
	engine.PassiveRegistry["barb_indestructible"] = engine.PassiveBonus{HPBonus: 120, DefenseBonus: 15}
}
