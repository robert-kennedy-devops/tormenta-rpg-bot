package skills

import "github.com/tormenta-bot/internal/models"

// RogueSkills returns 15 new rogue/ladino skills.
// Legacy: r_backstab, r_poison_blade, r_evasion, etc. (data.go)
// IDs use "rg_" prefix.
//
// Branches: "assassino" (burst/stealth), "envenenador" (DoT/debuff), "sombra" (utility/evasion)
// Synergy: Expose → bleed → Death Mark ×2
// Builds: Assassin (burst+execute), Poisoner (stacking DoT), Shadow (evasion+control), Hybrid
func RogueSkills() []models.Skill {
	return []models.Skill{
		// ── ASSASSINO (burst + execute) ────────────────────────────────────
		{
			ID: "rg_shadow_strike", Class: "rogue", Branch: "assassino", Tier: 3,
			Name: "Golpe das Sombras", Emoji: "🗡️", PointCost: 2, RequiredLevel: 30,
			Requires: "r_backstab", MPCost: 25, Damage: 55, DamageType: "physical",
			Role: RoleDirect, Scaling: 0.8,
			Description: "Ataque furtivo: 55 + 80% DEX. Se em furtividade: dano ×2 e aplica cegueira 2 turnos.",
			AppliesStatus: "blind", AppliesStatusTurns: 2,
		},
		{
			ID: "rg_death_mark", Class: "rogue", Branch: "assassino", Tier: 4,
			Name: "Marca da Morte", Emoji: "☠️", PointCost: 3, RequiredLevel: 50,
			Requires: "rg_shadow_strike", MPCost: 30, Damage: 70, DamageType: "physical",
			Role: RoleDirect, Scaling: 1.0,
			Description: "Marca o inimigo: 70 + 100% DEX. Se em sangramento: dano dobrado. Remove sangramento ao usar.",
			RequiresStatus: "bleed", SynergyMult: 1.0,
		},
		{
			ID: "rg_assassinate", Class: "rogue", Branch: "assassino", Tier: 5,
			Name: "Assassinar", Emoji: "💀", PointCost: 4, RequiredLevel: 75,
			Requires: "rg_death_mark", MPCost: 60, EnergyCost: 15,
			Damage: 90, DamageType: "physical",
			Role: RoleUlt, Scaling: 1.3, Cooldown: 6,
			Description: "Ultimate: 90 + 130% DEX. Se alvo com <40% HP: execução instantânea (dano ×3). Sempre crítico em furtividade.",
		},

		// ── ENVENENADOR (DoT + stacking debuffs) ──────────────────────────
		{
			ID: "rg_toxic_cloud", Class: "rogue", Branch: "envenenador", Tier: 3,
			Name: "Nuvem Tóxica", Emoji: "☁️", PointCost: 2, RequiredLevel: 30,
			Requires: "r_poison_blade", MPCost: 30, Damage: 15, DamageType: "poison",
			Role: RoleAoE, Scaling: 0.4,
			Description: "Espalha nuvem venenosa: 15 + 40% DEX. Aplica veneno de 3 turnos em TODOS os inimigos.",
			PoisonDmgPerTurn: 20, PoisonTurnsCount: 3,
			AppliesStatus: "poison", AppliesStatusTurns: 3,
		},
		{
			ID: "rg_virulent_poison", Class: "rogue", Branch: "envenenador", Tier: 4,
			Name: "Veneno Virulento", Emoji: "🧪", PointCost: 3, RequiredLevel: 50,
			Requires: "rg_toxic_cloud", MPCost: 35, DamageType: "poison",
			Role: RoleDoT, Scaling: 0.5,
			Description: "Aplica veneno que escala: 25 dano/turno, +10 a cada turno, por 5 turnos. Alvo envenenado fica cego no 3° turno.",
			PoisonDmgPerTurn: 25, PoisonTurnsCount: 5,
			AppliesStatus: "poison", AppliesStatusTurns: 5,
		},
		{
			ID: "rg_lethal_dose", Class: "rogue", Branch: "envenenador", Tier: 5,
			Name: "Dose Letal", Emoji: "💉", PointCost: 4, RequiredLevel: 78,
			Requires: "rg_virulent_poison", MPCost: 50, EnergyCost: 10,
			Damage: 40, DamageType: "poison",
			Role: RoleUlt, Scaling: 0.8, Cooldown: 5,
			Description: "Ultimate: 40 + 80% DEX imediato + veneno mortal (50 dano/turno por 6 turnos). Alvo envenenado recebe +50% de todo dano.",
			PoisonDmgPerTurn: 50, PoisonTurnsCount: 6,
			AppliesStatus: "poison", AppliesStatusTurns: 6,
		},

		// ── SOMBRA (evasion + controle + utility) ─────────────────────────
		{
			ID: "rg_shadow_step", Class: "rogue", Branch: "sombra", Tier: 3,
			Name: "Passo das Sombras", Emoji: "👻", PointCost: 2, RequiredLevel: 30,
			Requires: "r_evasion", MPCost: 20,
			Role: RoleUtility, Cooldown: 3,
			Description: "Teleporta atrás do inimigo e entra em furtividade por 1 turno. Próximo ataque é garantidamente crítico.",
		},
		{
			ID: "rg_expose", Class: "rogue", Branch: "sombra", Tier: 3,
			Name: "Expor", Emoji: "🩸", PointCost: 2, RequiredLevel: 35,
			Requires: "rg_shadow_step", MPCost: 25, Damage: 35, DamageType: "physical",
			Role: RoleDebuff, Scaling: 0.5,
			Description: "Ataque calculado: 35 + 50% DEX. Reduz 25% armadura do alvo e aplica sangramento 3 turnos.",
			AppliesStatus: "bleed", AppliesStatusTurns: 3,
		},
		{
			ID: "rg_smoke_bomb", Class: "rogue", Branch: "sombra", Tier: 4,
			Name: "Bomba de Fumaça", Emoji: "💨", PointCost: 3, RequiredLevel: 55,
			Requires: "rg_expose", MPCost: 35,
			Role: RoleControl, Cooldown: 4,
			Description: "Lança bomba de fumaça: cega TODOS os inimigos por 2 turnos (-60% precisão) e você entra em furtividade completa.",
			AppliesStatus: "blind", AppliesStatusTurns: 2,
		},
		{
			ID: "rg_killing_spree", Class: "rogue", Branch: "sombra", Tier: 5,
			Name: "Espiral da Morte", Emoji: "🌀", PointCost: 4, RequiredLevel: 80,
			Requires: "rg_smoke_bomb", MPCost: 80, EnergyCost: 20,
			Damage: 50, DamageType: "physical",
			Role: RoleUlt, Scaling: 0.9, Cooldown: 8,
			Description: "Ultimate: ataca 5 vezes rapidamente (50 + 90% DEX cada). Cada inimigo diferente atacado regenera 10 MP. Se qualquer alvo morrer, resets o cooldown.",
		},
	}
}
