package skills

import "github.com/tormenta-bot/internal/models"

// BarbarianSkills returns 15 new barbarian skills (class had ZERO before).
// IDs use "barb2_" prefix (rpg/skill_trees_extended.go uses "barb_").
//
// Branches: "frenesi" (rage/offense), "resistencia" (defense/sustain), "tribal" (utility/support)
// Synergy: Rage → berserk state → Rampage scales with rage stacks
// Builds: Berserker (pure offense), Juggernaut (tanky+regen), Warchief (support+AoE)
func BarbarianSkills() []models.Skill {
	return []models.Skill{
		// ── FRENESI (ofensivo — rage stacks) ──────────────────────────────
		{
			ID: "barb2_rage", Class: "barbarian", Branch: "frenesi", Tier: 1,
			Name: "Fúria", Emoji: "💢", PointCost: 1, RequiredLevel: 1,
			MPCost: 0, EnergyCost: 5,
			Role: RoleBuff,
			Description: "Entra em estado de frenesi: +60% ATK e -30% DEF por 3 turnos. Base de todo o ramo ofensivo.",
		},
		{
			ID: "barb2_blood_thirst", Class: "barbarian", Branch: "frenesi", Tier: 2,
			Name: "Sede de Sangue", Emoji: "🩸", PointCost: 1, RequiredLevel: 8,
			Requires: "barb2_rage", MPCost: 10, Damage: 30, DamageType: "physical",
			Role: RoleDirect, Scaling: 0.6,
			Description: "Ataque voraz: 30 + 60% STR. Cura 25% do dano causado. Em frenesi: cura duplicada.",
		},
		{
			ID: "barb2_war_cry", Class: "barbarian", Branch: "frenesi", Tier: 3,
			Name: "Grito de Guerra", Emoji: "📢", PointCost: 2, RequiredLevel: 25,
			Requires: "barb2_blood_thirst", MPCost: 20,
			Role: RoleAoE,
			Description: "Grito que debilita todos os inimigos: -25% ATK e DEF por 3 turnos. Em frenesi: também aplica medo (stun 1 turno).",
			AppliesStatus: "stun", AppliesStatusTurns: 1,
		},
		{
			ID: "barb2_rampage", Class: "barbarian", Branch: "frenesi", Tier: 4,
			Name: "Devastação", Emoji: "🌪️", PointCost: 3, RequiredLevel: 45,
			Requires: "barb2_war_cry", MPCost: 30, Damage: 50, DamageType: "physical",
			Role: RoleAoE, Scaling: 0.7,
			Description: "Ataca todos os inimigos 3 vezes (50 + 70% STR cada). Em frenesi: ignora 25% armadura.",
		},
		{
			ID: "barb2_lord_of_war", Class: "barbarian", Branch: "frenesi", Tier: 5,
			Name: "Senhor da Guerra", Emoji: "⚔️", PointCost: 4, RequiredLevel: 70,
			Requires: "barb2_rampage", MPCost: 60, EnergyCost: 15,
			Damage: 100, DamageType: "physical",
			Role: RoleUlt, Scaling: 1.2, Cooldown: 7,
			Description: "Ultimate: encarna o deus da guerra — 100 + 120% STR. Entra em superfrenesi (5 turnos): +100% ATK, imune a stun e controle.",
		},

		// ── RESISTÊNCIA (defesa + sustain) ────────────────────────────────
		{
			ID: "barb2_toughness", Class: "barbarian", Branch: "resistencia", Tier: 1,
			Name: "Pele Grossa", Emoji: "🪨", PointCost: 1, RequiredLevel: 1,
			Passive: true, Role: RolePassive,
			Description: "+20 HP máximo permanente. Passivo.",
		},
		{
			ID: "barb2_endure", Class: "barbarian", Branch: "resistencia", Tier: 2,
			Name: "Resistir", Emoji: "💪", PointCost: 1, RequiredLevel: 10,
			Requires: "barb2_toughness", MPCost: 15,
			Role: RoleBuff, Cooldown: 3,
			Description: "Postura resistente: absorve o próximo ataque que causaria mais de 30% HP. Regenera 15 HP/turno por 2 turnos.",
		},
		{
			ID: "barb2_stone_skin", Class: "barbarian", Branch: "resistencia", Tier: 3,
			Name: "Pele de Pedra", Emoji: "🗿", PointCost: 2, RequiredLevel: 28,
			Requires: "barb2_endure", MPCost: 0, Passive: true,
			Role: RolePassive,
			Description: "Passivo: reduz 15% de todo dano físico recebido permanentemente.",
		},
		{
			ID: "barb2_immortal_will", Class: "barbarian", Branch: "resistencia", Tier: 4,
			Name: "Vontade Imortal", Emoji: "💛", PointCost: 3, RequiredLevel: 50,
			Requires: "barb2_stone_skin", MPCost: 40,
			Role: RoleBuff, Cooldown: 5,
			Description: "Ativa vontade imortal: se HP cair abaixo de 10% neste turno, volta para 30% HP. Uma vez por combate.",
		},
		{
			ID: "barb2_mountain_stance", Class: "barbarian", Branch: "resistencia", Tier: 5,
			Name: "Postura da Montanha", Emoji: "🏔️", PointCost: 4, RequiredLevel: 72,
			Requires: "barb2_immortal_will", MPCost: 50, EnergyCost: 10,
			Role: RoleUlt, Cooldown: 6,
			Description: "Ultimate: transforma em rocha por 2 turnos — imune a TODO dano e efeitos. Ao sair: devolve dano acumulado em explosão.",
		},

		// ── TRIBAL (suporte + AoE) ─────────────────────────────────────────
		{
			ID: "barb2_war_paint", Class: "barbarian", Branch: "tribal", Tier: 1,
			Name: "Pintura de Guerra", Emoji: "🎨", PointCost: 1, RequiredLevel: 5,
			MPCost: 15, Role: RoleBuff,
			Description: "Pinta o corpo com pigmentos tribais: +20% ATK e imunidade a veneno por 4 turnos.",
		},
		{
			ID: "barb2_totemic_roar", Class: "barbarian", Branch: "tribal", Tier: 2,
			Name: "Rugido Totêmico", Emoji: "🦁", PointCost: 1, RequiredLevel: 12,
			Requires: "barb2_war_paint", MPCost: 20,
			Role: RoleDebuff, Cooldown: 2,
			Description: "Rugido ancestral: aterroriza o inimigo (-30% ATK por 3 turnos) e o stuna por 1 turno.",
			AppliesStatus: "stun", AppliesStatusTurns: 1,
		},
		{
			ID: "barb2_battle_shout", Class: "barbarian", Branch: "tribal", Tier: 3,
			Name: "Brado de Batalha", Emoji: "📯", PointCost: 2, RequiredLevel: 30,
			Requires: "barb2_totemic_roar", MPCost: 25, Damage: 35, DamageType: "physical",
			Role: RoleAoE, Scaling: 0.5,
			Description: "Golpe com grito: 35 + 50% STR em área. Todos aliados recebem +20% ATK por 3 turnos.",
		},
		{
			ID: "barb2_titan_smash", Class: "barbarian", Branch: "tribal", Tier: 4,
			Name: "Impacto do Titã", Emoji: "💥", PointCost: 3, RequiredLevel: 55,
			Requires: "barb2_battle_shout", MPCost: 45, Damage: 70, DamageType: "physical",
			Role: RoleAoE, Scaling: 0.9, Cooldown: 3,
			Description: "Golpeia o chão: 70 + 90% STR em área. Atordoa todos por 1 turno e causa sangramento.",
			AppliesStatus: "stun", AppliesStatusTurns: 1,
		},
		{
			ID: "barb2_ancestral_spirit", Class: "barbarian", Branch: "tribal", Tier: 5,
			Name: "Espírito Ancestral", Emoji: "👁️", PointCost: 4, RequiredLevel: 75,
			Requires: "barb2_titan_smash", MPCost: 70, EnergyCost: 20,
			Role: RoleUlt, Cooldown: 8,
			Description: "Ultimate: convoca espírito ancestral por 3 turnos — duplica todos os seus stats, cura 20% HP/turno e faz seus ataques AoE.",
		},
	}
}
