package skills

import "github.com/tormenta-bot/internal/models"

// ClericSkills returns 15 new cleric skills (class had ZERO before).
// IDs use "cler2_" prefix.
//
// Branches: "cura" (healing+support), "luz" (holy offense), "protecao_divina" (shielding)
// Synergy: Bless → blessed → Mass Heal heals bonus
// Builds: Battle Healer (offense+sustain), Holy Warrior (damage dealer), Shielder (defensive)
func ClericSkills() []models.Skill {
	return []models.Skill{
		// ── CURA (cura + suporte) ──────────────────────────────────────────
		{
			ID: "cler2_heal", Class: "cleric", Branch: "cura", Tier: 1,
			Name: "Cura", Emoji: "💚", PointCost: 1, RequiredLevel: 1,
			MPCost: 15, Role: RoleHeal,
			Description: "Restaura 2d8 + SAB HP. Remova efeito de veneno do alvo se presente.",
		},
		{
			ID: "cler2_mass_heal", Class: "cleric", Branch: "cura", Tier: 2,
			Name: "Cura em Área", Emoji: "💖", PointCost: 1, RequiredLevel: 8,
			Requires: "cler2_heal", MPCost: 30,
			Role: RoleHeal,
			Description: "Cura todos aliados por 2d6 + SAB. Aliados abençoados recebem +50% de cura.",
			RequiresStatus: "shield", SynergyMult: 0.5,
		},
		{
			ID: "cler2_regen", Class: "cleric", Branch: "cura", Tier: 3,
			Name: "Regeneração", Emoji: "🔄", PointCost: 2, RequiredLevel: 25,
			Requires: "cler2_mass_heal", MPCost: 35,
			Role: RoleHeal, Cooldown: 2,
			Description: "Aplica regeneração poderosa: recupera 25 HP/turno por 4 turnos. Remove maldições.",
			AppliesStatus: "shield", AppliesStatusTurns: 4,
		},
		{
			ID: "cler2_divine_grace", Class: "cleric", Branch: "cura", Tier: 4,
			Name: "Graça Divina", Emoji: "⭐", PointCost: 3, RequiredLevel: 45,
			Requires: "cler2_regen", MPCost: 50,
			Role: RoleHeal, Cooldown: 3,
			Description: "Cura massiva: restaura 50% do HP máximo. Remove TODOS os status negativos.",
		},
		{
			ID: "cler2_miracle", Class: "cleric", Branch: "cura", Tier: 5,
			Name: "Milagre", Emoji: "🌟", PointCost: 4, RequiredLevel: 70,
			Requires: "cler2_divine_grace", MPCost: 80, EnergyCost: 25,
			Role: RoleUlt, Cooldown: 10,
			Description: "Ultimate: milagre divino — se abaixo de 20% HP, restaura para 100% HP instantaneamente. Cura aliados por 40% do HP máximo.",
		},

		// ── LUZ DIVINA (dano sagrado + debuff) ────────────────────────────
		{
			ID: "cler2_smite_evil", Class: "cleric", Branch: "luz", Tier: 1,
			Name: "Punir o Mal", Emoji: "⚡", PointCost: 1, RequiredLevel: 3,
			MPCost: 15, Damage: 20, DamageType: "holy",
			Role: RoleDirect, Scaling: 0.5,
			Description: "Feixe de luz: 20 + 50% SAB. Dano triplo contra mortos-vivos e demônios.",
		},
		{
			ID: "cler2_holy_bolt", Class: "cleric", Branch: "luz", Tier: 2,
			Name: "Raio Sagrado", Emoji: "🌩️", PointCost: 1, RequiredLevel: 10,
			Requires: "cler2_smite_evil", MPCost: 25, Damage: 35, DamageType: "holy",
			Role: RoleDirect, Scaling: 0.65,
			Description: "Raio de luz concentrado: 35 + 65% SAB. Silencia o alvo por 1 turno.",
			AppliesStatus: "silence", AppliesStatusTurns: 1,
		},
		{
			ID: "cler2_solar_flare", Class: "cleric", Branch: "luz", Tier: 3,
			Name: "Chama Solar", Emoji: "☀️", PointCost: 2, RequiredLevel: 28,
			Requires: "cler2_holy_bolt", MPCost: 40, Damage: 50, DamageType: "holy",
			Role: RoleAoE, Scaling: 0.8, Cooldown: 2,
			Description: "Explosão de luz solar em área: 50 + 80% SAB. Cega todos os inimigos por 2 turnos.",
			AppliesStatus: "blind", AppliesStatusTurns: 2,
		},
		{
			ID: "cler2_word_of_pain", Class: "cleric", Branch: "luz", Tier: 4,
			Name: "Palavra da Dor", Emoji: "💬", PointCost: 3, RequiredLevel: 50,
			Requires: "cler2_solar_flare", MPCost: 50, Damage: 65, DamageType: "holy",
			Role: RoleDebuff, Scaling: 0.9, Cooldown: 2,
			Description: "Palavras sagradas: 65 + 90% SAB. Maldiz o alvo por 4 turnos (-30% todos stats). Silenciados recebem ×1.5.",
			RequiresStatus: "silence", SynergyMult: 0.5,
			AppliesStatus: "curse", AppliesStatusTurns: 4,
		},
		{
			ID: "cler2_avatar_of_light", Class: "cleric", Branch: "luz", Tier: 5,
			Name: "Avatar da Luz", Emoji: "🌈", PointCost: 4, RequiredLevel: 75,
			Requires: "cler2_word_of_pain", MPCost: 80, EnergyCost: 20,
			Damage: 100, DamageType: "holy",
			Role: RoleUlt, Scaling: 1.2, Cooldown: 8,
			Description: "Ultimate: incorpora avatar de luz — 100 + 120% SAB em área. Aliados recebem +50% ATK por 3 turnos e são curados por 30% HP.",
		},

		// ── PROTEÇÃO DIVINA (escudos + utility) ───────────────────────────
		{
			ID: "cler2_bless", Class: "cleric", Branch: "protecao_divina", Tier: 1,
			Name: "Abençoar", Emoji: "✝️", PointCost: 1, RequiredLevel: 1,
			MPCost: 10, Role: RoleBuff,
			Description: "Abençoa o alvo: +20% ATK e DEF por 3 turnos. Base das bênçãos divinas.",
			AppliesStatus: "shield", AppliesStatusTurns: 3,
		},
		{
			ID: "cler2_divine_wall", Class: "cleric", Branch: "protecao_divina", Tier: 2,
			Name: "Muro Divino", Emoji: "🏛️", PointCost: 1, RequiredLevel: 10,
			Requires: "cler2_bless", MPCost: 25,
			Role: RoleBuff, Cooldown: 3,
			Description: "Cria barreira divina: absorve próximos 50 HP de dano. Dura 3 turnos.",
			AppliesStatus: "shield", AppliesStatusTurns: 3,
		},
		{
			ID: "cler2_sanctuary", Class: "cleric", Branch: "protecao_divina", Tier: 3,
			Name: "Santuário", Emoji: "🕍", PointCost: 2, RequiredLevel: 28,
			Requires: "cler2_divine_wall", MPCost: 35,
			Role: RoleBuff, Cooldown: 4,
			Description: "Cria zona sagrada por 3 turnos: imune a veneno e maldições, regenera 10 HP/turno.",
		},
		{
			ID: "cler2_aegis", Class: "cleric", Branch: "protecao_divina", Tier: 4,
			Name: "Égide", Emoji: "⚜️", PointCost: 3, RequiredLevel: 52,
			Requires: "cler2_sanctuary", MPCost: 45,
			Role: RoleBuff, Cooldown: 5,
			Description: "Escudo divino supremo: reduz 50% de todo dano por 2 turnos. Ao expirar, reflete 30% do dano absorvido.",
		},
		{
			ID: "cler2_gods_shield", Class: "cleric", Branch: "protecao_divina", Tier: 5,
			Name: "Escudo dos Deuses", Emoji: "🌠", PointCost: 4, RequiredLevel: 78,
			Requires: "cler2_aegis", MPCost: 70, EnergyCost: 20,
			Role: RoleUlt, Cooldown: 9,
			Description: "Ultimate: o próprio deus o protege — você e aliados ficam invulneráveis por 1 turno. Na expiração: golpe sagrado em área por 70% HP max de todos aliados.",
		},
	}
}
