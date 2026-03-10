package skills

import "github.com/tormenta-bot/internal/models"

// PaladinSkills returns 15 new paladin skills (class had ZERO before).
// IDs use "pal2_" prefix.
//
// Branches: "sagrado" (holy damage), "protecao" (shield/tank), "julgamento" (hybrid+smite)
// Synergy: Holy Mark → sacred_mark → Divine Judgment ×2
// Builds: Holy Avenger (offense+holy), Shield Guardian (tank), Judge (debuff+execute)
func PaladinSkills() []models.Skill {
	return []models.Skill{
		// ── SAGRADO (dano sagrado) ─────────────────────────────────────────
		{
			ID: "pal2_smite", Class: "paladin", Branch: "sagrado", Tier: 1,
			Name: "Golpe Divino", Emoji: "✨", PointCost: 1, RequiredLevel: 1,
			MPCost: 10, Damage: 15, DamageType: "holy",
			Role: RoleDirect, Scaling: 0.4,
			Description: "Golpe imbuído de luz sagrada: 15 + 40% STR/SAB. Dano dobrado contra mortos-vivos e demônios.",
		},
		{
			ID: "pal2_holy_flame", Class: "paladin", Branch: "sagrado", Tier: 2,
			Name: "Chama Sagrada", Emoji: "🕯️", PointCost: 1, RequiredLevel: 8,
			Requires: "pal2_smite", MPCost: 20, Damage: 25, DamageType: "holy",
			Role: RoleDoT, Scaling: 0.5,
			Description: "Chama divina: 25 + 50% SAB. Aplica queimadura sagrada por 3 turnos (dano extra contra mal).",
			AppliesStatus: "burn", AppliesStatusTurns: 3,
		},
		{
			ID: "pal2_holy_sword", Class: "paladin", Branch: "sagrado", Tier: 3,
			Name: "Espada da Luz", Emoji: "⚔️", PointCost: 2, RequiredLevel: 28,
			Requires: "pal2_holy_flame", MPCost: 30, Damage: 45, DamageType: "holy",
			Role: RoleDirect, Scaling: 0.7,
			Description: "Imbuiu a espada com luz: 45 + 70% STR. Se alvo em queimadura sagrada: dano +60% e cura você por 20% do dano.",
			RequiresStatus: "burn", SynergyMult: 0.6,
		},
		{
			ID: "pal2_divine_judgment", Class: "paladin", Branch: "sagrado", Tier: 4,
			Name: "Julgamento Divino", Emoji: "☀️", PointCost: 3, RequiredLevel: 50,
			Requires: "pal2_holy_sword", MPCost: 50, Damage: 80, DamageType: "holy",
			Role: RoleDirect, Scaling: 1.0, Cooldown: 2,
			Description: "Julgamento sagrado: 80 + 100% SAB. Dobrado contra malditos/mortos-vivos. Remove uma maldição do alvo.",
		},
		{
			ID: "pal2_armageddon", Class: "paladin", Branch: "sagrado", Tier: 5,
			Name: "Armagedom", Emoji: "🌤️", PointCost: 4, RequiredLevel: 72,
			Requires: "pal2_divine_judgment", MPCost: 80, EnergyCost: 20,
			Damage: 120, DamageType: "holy",
			Role: RoleUlt, Scaling: 1.3, Cooldown: 8,
			Description: "Ultimate: luz divina em área — 120 + 130% SAB. Todos os aliados recebem 30% HP cura. Inimigos malditos são eliminados se HP < 20%.",
		},

		// ── PROTEÇÃO (escudo + tank) ───────────────────────────────────────
		{
			ID: "pal2_holy_shield", Class: "paladin", Branch: "protecao", Tier: 1,
			Name: "Escudo Sagrado", Emoji: "🛡️", PointCost: 1, RequiredLevel: 1,
			MPCost: 15, Role: RoleBuff, Cooldown: 3,
			Description: "Cria escudo sagrado: absorve próximo dano até 40 HP. Dura até ser absorvido ou 2 turnos.",
			AppliesStatus: "shield", AppliesStatusTurns: 2,
		},
		{
			ID: "pal2_lay_on_hands", Class: "paladin", Branch: "protecao", Tier: 2,
			Name: "Imposição de Mãos", Emoji: "🙏", PointCost: 1, RequiredLevel: 10,
			Requires: "pal2_holy_shield", MPCost: 25,
			Role: RoleHeal,
			Description: "Cura poderosa: restaura 3d8 + SAB HP. Remove um efeito negativo (veneno, maldição ou queimadura).",
		},
		{
			ID: "pal2_aura_of_protection", Class: "paladin", Branch: "protecao", Tier: 3,
			Name: "Aura Protetora", Emoji: "💛", PointCost: 2, RequiredLevel: 28,
			Requires: "pal2_lay_on_hands", MPCost: 0, Passive: true,
			Role: RolePassive,
			Description: "Passivo: reduz 10% de todo dano recebido. Aliados próximos recebem 5% de redução de dano.",
		},
		{
			ID: "pal2_sacred_ground", Class: "paladin", Branch: "protecao", Tier: 4,
			Name: "Solo Sagrado", Emoji: "✳️", PointCost: 3, RequiredLevel: 48,
			Requires: "pal2_aura_of_protection", MPCost: 40, Cooldown: 4,
			Role: RoleBuff,
			Description: "Consagra o terreno por 3 turnos: você e aliados regeneram 15 HP/turno e são imunes a veneno.",
		},
		{
			ID: "pal2_divine_shield", Class: "paladin", Branch: "protecao", Tier: 5,
			Name: "Escudo Divino", Emoji: "🌟", PointCost: 4, RequiredLevel: 75,
			Requires: "pal2_sacred_ground", MPCost: 60, EnergyCost: 10,
			Role: RoleUlt, Cooldown: 8,
			Description: "Ultimate: torna-se completamente invulnerável por 2 turnos. Ao sair, explode com luz sagrada causando 50% HP máximo em dano a todos inimigos.",
		},

		// ── JULGAMENTO (híbrido + debuff) ──────────────────────────────────
		{
			ID: "pal2_mark_of_justice", Class: "paladin", Branch: "julgamento", Tier: 1,
			Name: "Marca da Justiça", Emoji: "⚖️", PointCost: 1, RequiredLevel: 5,
			MPCost: 15,
			Role: RoleDebuff,
			Description: "Marca o alvo com justiça divina: -20% DEF e DEX por 3 turnos. Acumula com outras marcas.",
			AppliesStatus: "curse", AppliesStatusTurns: 3,
		},
		{
			ID: "pal2_holy_wrath", Class: "paladin", Branch: "julgamento", Tier: 2,
			Name: "Ira Sagrada", Emoji: "😤", PointCost: 1, RequiredLevel: 12,
			Requires: "pal2_mark_of_justice", MPCost: 25, Damage: 35, DamageType: "holy",
			Role: RoleDirect, Scaling: 0.6,
			Description: "Canaliza ira: 35 + 60% SAB. Se alvo maldito: +50% dano e aplica silêncio 1 turno.",
			RequiresStatus: "curse", SynergyMult: 0.5,
			AppliesStatus: "silence", AppliesStatusTurns: 1,
		},
		{
			ID: "pal2_crusader_strike", Class: "paladin", Branch: "julgamento", Tier: 3,
			Name: "Golpe do Cruzado", Emoji: "🗡️", PointCost: 2, RequiredLevel: 30,
			Requires: "pal2_holy_wrath", MPCost: 30, Damage: 50, DamageType: "physical",
			Role: RoleDirect, Scaling: 0.75,
			Description: "Golpe em cruz: 50 + 75% STR. Cura aliado com menos HP por 20% do dano causado.",
		},
		{
			ID: "pal2_consecration", Class: "paladin", Branch: "julgamento", Tier: 4,
			Name: "Consagração", Emoji: "🌊", PointCost: 3, RequiredLevel: 52,
			Requires: "pal2_crusader_strike", MPCost: 45, Damage: 40, DamageType: "holy",
			Role: RoleAoE, Scaling: 0.7, Cooldown: 3,
			Description: "Consagra área: 40 + 70% SAB em todos inimigos. Malditos recebem ×1.5. Aliados na área curam 20 HP/turno.",
			RequiresStatus: "curse", SynergyMult: 0.5,
		},
		{
			ID: "pal2_resurrection", Class: "paladin", Branch: "julgamento", Tier: 5,
			Name: "Ressurreição Divina", Emoji: "🌈", PointCost: 4, RequiredLevel: 78,
			Requires: "pal2_consecration", MPCost: 100, EnergyCost: 30,
			Role: RoleUlt, Cooldown: 10,
			Description: "Ultimate sagrado: remove TODOS os efeitos negativos de você, cura 100% HP e causa 80 + 100% SAB em todos os inimigos. Uma vez por combate.",
		},
	}
}
