package skills

import "github.com/tormenta-bot/internal/models"

// WarriorSkills returns 15 new warrior skills across 3 branches.
// Legacy warrior skills (w_iron_skin, w_shield_slam, etc.) remain in data.go.
// IDs use "wa_" prefix to avoid collisions.
//
// Branches: "protetor" (tank), "berserker" (offense), "campiao" (duelist)
// Viable builds:
//   - Tank: protetor branch (mitigation, taunt, counter)
//   - Berserker: berserker branch (burst, rage, lifesteal)
//   - Duelist: campiao branch (precision, parry, execute)
//   - Hybrid: mix protetor+campiao (resilient fighter)
func WarriorSkills() []models.Skill {
	return []models.Skill{
		// ── PROTETOR (T3-T5 expansion) ────────────────────────────────────
		{
			ID: "wa_bulwark", Class: "warrior", Branch: "protetor", Tier: 3,
			Name: "Bastião", Emoji: "🏰", PointCost: 2, RequiredLevel: 30,
			Requires: "w_shield_wall", MPCost: 25,
			Role: RoleBuff, Passive: false, Scaling: 0,
			Description: "Reduz 40% do próximo dano recebido por 2 turnos. Gera ameaça tripla.",
			AppliesStatus: "shield", AppliesStatusTurns: 2,
		},
		{
			ID: "wa_taunt", Class: "warrior", Branch: "protetor", Tier: 3,
			Name: "Provocação", Emoji: "😤", PointCost: 2, RequiredLevel: 35,
			Requires: "wa_bulwark", MPCost: 15,
			Role: RoleDebuff, Passive: false, Scaling: 0,
			Description: "Força o inimigo a atacar apenas você por 2 turnos. Reduz DEF inimigo em 15%.",
			AppliesStatus: "curse", AppliesStatusTurns: 2,
		},
		{
			ID: "wa_counter_strike", Class: "warrior", Branch: "protetor", Tier: 4,
			Name: "Contra-Ataque", Emoji: "⚡", PointCost: 3, RequiredLevel: 50,
			Requires: "wa_taunt", MPCost: 30, Damage: 35, DamageType: "physical",
			Role: RoleDirect, Scaling: 0.6,
			Description: "Retalia instantaneamente quando bloqueado: 35 + 60% ATK. Ignora 30% armadura.",
			RequiresStatus: "shield", SynergyMult: 0.5,
		},
		{
			ID: "wa_fortress", Class: "warrior", Branch: "protetor", Tier: 5,
			Name: "Fortaleza Viva", Emoji: "🏯", PointCost: 4, RequiredLevel: 70,
			Requires: "wa_counter_strike", MPCost: 50, EnergyCost: 10,
			Role: RoleUlt, Passive: false, Cooldown: 5,
			Description: "Ultimate: por 3 turnos torna-se imune a stun/blind, recupera 5% HP/turno e reflete 20% do dano recebido.",
		},

		// ── BERSERKER (T3-T5 expansion) ───────────────────────────────────
		{
			// Renomeado: "Sede de Sangue" conflitava com barb2_blood_thirst (barbarian).
			// Warrior recebe versão mais técnica/marcial: "Golpe Voraz".
			ID: "wa_bloodlust", Class: "warrior", Branch: "berserker", Tier: 3,
			Name: "Golpe Voraz", Emoji: "🩸", PointCost: 2, RequiredLevel: 30,
			Requires: "w_warcry", MPCost: 20, Damage: 40, DamageType: "physical",
			Role: RoleDirect, Scaling: 0.7,
			Description: "Ataque marcial brutal que drena a vitalidade do inimigo: 40 + 70% ATK. Cura 30% do dano causado. Cura dobrada se inimigo abaixo de 50% HP.",
			AppliesStatus: "bleed", AppliesStatusTurns: 3,
		},
		{
			ID: "wa_execute", Class: "warrior", Branch: "berserker", Tier: 4,
			Name: "Execução", Emoji: "💀", PointCost: 3, RequiredLevel: 50,
			Requires: "wa_bloodlust", MPCost: 35, Damage: 60, DamageType: "physical",
			Role: RoleDirect, Scaling: 0.9,
			Description: "Executa inimigo: dano dobrado se alvo com <30% HP. Aplica sangramento por 4 turnos.",
			RequiresStatus: "bleed", SynergyMult: 1.0,
			AppliesStatus: "bleed", AppliesStatusTurns: 4,
		},
		{
			ID: "wa_avatar_war", Class: "warrior", Branch: "berserker", Tier: 5,
			Name: "Avatar da Guerra", Emoji: "⚔️", PointCost: 4, RequiredLevel: 75,
			Requires: "wa_execute", MPCost: 60, EnergyCost: 15,
			Role: RoleUlt, Cooldown: 8,
			Description: "Ultimate: entra em estado de frenesi por 4 turnos — +80% ATK, +50% velocidade, imune a stun. Cura 10% HP/turno.",
		},

		// ── CAMPIÃO (T3-T5 expansion) ──────────────────────────────────────
		{
			ID: "wa_precision", Class: "warrior", Branch: "campiao", Tier: 3,
			Name: "Golpe de Precisão", Emoji: "🎯", PointCost: 2, RequiredLevel: 30,
			Requires: "w_armor_break", MPCost: 25, Damage: 45, DamageType: "physical",
			Role: RoleDirect, Scaling: 0.8,
			Description: "Ataque preciso: +30% chance de crítico. Dano de crítico aumentado para ×3.",
		},
		{
			ID: "wa_parry", Class: "warrior", Branch: "campiao", Tier: 3,
			Name: "Aparar", Emoji: "🤺", PointCost: 2, RequiredLevel: 35,
			Requires: "wa_precision", MPCost: 15,
			Role: RoleBuff, Cooldown: 3,
			Description: "Postura defensiva: próximo ataque físico recebido é bloqueado e contra-atacado com 80% do dano.",
		},
		{
			ID: "wa_blade_storm", Class: "warrior", Branch: "campiao", Tier: 4,
			Name: "Tempestade de Lâminas", Emoji: "🌪️", PointCost: 3, RequiredLevel: 55,
			Requires: "wa_parry", MPCost: 40, Damage: 30, DamageType: "physical",
			Role: RoleAoE, Scaling: 0.5,
			Description: "Redemoinho cortante: atinge todos os inimigos por 30 + 50% ATK. Chance de sangramento em cada alvo.",
			AppliesStatus: "bleed", AppliesStatusTurns: 2,
		},
		{
			ID: "wa_legendary_strike", Class: "warrior", Branch: "campiao", Tier: 5,
			Name: "Golpe Lendário", Emoji: "🌟", PointCost: 4, RequiredLevel: 80,
			Requires: "wa_blade_storm", MPCost: 70, EnergyCost: 20,
			Damage: 120, DamageType: "physical",
			Role: RoleUlt, Scaling: 1.2, Cooldown: 6,
			Description: "Ultimate: golpe devastador 120 + 120% ATK. Ignora toda armadura. Inimigos em sangramento recebem +100% dano.",
			RequiresStatus: "bleed", SynergyMult: 1.0,
		},
	}
}
