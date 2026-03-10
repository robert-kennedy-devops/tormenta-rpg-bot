package skills

import "github.com/tormenta-bot/internal/models"

// MageSkills returns 15 new mage/arcanist skills.
// Legacy: m_fireball, m_ice_lance, m_arcane_blast, etc. (data.go)
// IDs use "mg_" prefix.
//
// Branches: "piromante" (fire), "crionita" (ice), "arcanista" (arcane)
// Synergy chain: Lâmina de Gelo → freeze → Lança Potencializada +50% dmg
// Builds: Pyromancer (DoT/burst), Cryomancer (control/execute), Arcane (scaling/mana)
func MageSkills() []models.Skill {
	return []models.Skill{
		// ── PIROMANTE (fire — DoT + burst) ────────────────────────────────
		{
			ID: "mg_combustion", Class: "mage", Branch: "piromante", Tier: 3,
			Name: "Combustão", Emoji: "🔥", PointCost: 2, RequiredLevel: 30,
			Requires: "m_fireball", MPCost: 30, Damage: 40, DamageType: "fire",
			Role: RoleDoT, Scaling: 0.6,
			Description: "Incendeia o alvo por 4 turnos: 40 + 60% INT inicial + 15 por turno. Alvos com queimadura recebem +30% de dano de fogo.",
			AppliesStatus: "burn", AppliesStatusTurns: 4,
		},
		{
			// Substituído: "Meteoro" era duplicata do legado m_meteor.
			// Novo: Chuva de Meteoros — 4 impactos aleatórios em área (multi-hit AoE).
			ID: "mg_meteor", Class: "mage", Branch: "piromante", Tier: 4,
			Name: "Chuva de Meteoros", Emoji: "☄️", PointCost: 3, RequiredLevel: 50,
			Requires: "mg_combustion", MPCost: 50, Damage: 35, DamageType: "fire",
			Role: RoleAoE, Scaling: 0.4, Cooldown: 3,
			Description: "Invoca 4 meteoros em posições aleatórias: cada um causa 35 + 40% INT em área. Cada impacto aplica queimadura 2 turnos. Alvos já em combustão recebem +50% por impacto.",
			RequiresStatus: "burn", SynergyMult: 0.5,
			AppliesStatus: "burn", AppliesStatusTurns: 2,
		},
		{
			ID: "mg_phoenix_fire", Class: "mage", Branch: "piromante", Tier: 5,
			Name: "Chama da Fênix", Emoji: "🦅", PointCost: 4, RequiredLevel: 75,
			Requires: "mg_meteor", MPCost: 80, EnergyCost: 15,
			Damage: 100, DamageType: "fire",
			Role: RoleUlt, Scaling: 1.2, Cooldown: 7,
			Description: "Ultimate: explosão fênix 100 + 120% INT. Se você estiver abaixo de 30% HP, também recupera 40% do seu HP máximo.",
		},

		// ── CRIONITA (ice — controle + synergy) ───────────────────────────
		{
			// Substituído: "Nova de Gelo" era duplicata do legado m_frost_nova.
			// Novo: Lâmina de Gelo — projétil de gelo concentrado de curto alcance.
			ID: "mg_frost_nova", Class: "mage", Branch: "crionita", Tier: 3,
			Name: "Lâmina de Gelo", Emoji: "🔪", PointCost: 2, RequiredLevel: 30,
			Requires: "m_ice_lance", MPCost: 30, Damage: 45, DamageType: "ice",
			Role: RoleDirect, Scaling: 0.6, Cooldown: 2,
			Description: "Conjura lâmina de gelo cristalizado: 45 + 60% INT. 50% de chance de congelar por 1 turno. Alvo congelado recebe +40% de dano de gelo nesse turno.",
			AppliesStatus: "freeze", AppliesStatusTurns: 1,
		},
		{
			ID: "mg_ice_lance_ex", Class: "mage", Branch: "crionita", Tier: 3,
			Name: "Lança de Gelo Potencializada", Emoji: "🧊", PointCost: 2, RequiredLevel: 35,
			Requires: "mg_frost_nova", MPCost: 35, Damage: 60, DamageType: "ice",
			Role: RoleDirect, Scaling: 0.85,
			Description: "Lança de gelo concentrada: 60 + 85% INT. Se alvo congelado: +50% dano e stun adicional por 1 turno.",
			RequiresStatus: "freeze", SynergyMult: 0.5,
		},
		{
			// Substituído: "Nevasca" era duplicata do legado m_blizzard.
			// Novo: Tempestade de Granizo — DoT em área com lentidão cumulativa.
			ID: "mg_blizzard", Class: "mage", Branch: "crionita", Tier: 4,
			Name: "Tempestade de Granizo", Emoji: "🌨️", PointCost: 3, RequiredLevel: 55,
			Requires: "mg_ice_lance_ex", MPCost: 60, Damage: 25, DamageType: "ice",
			Role: RoleAoE, Scaling: 0.5, Cooldown: 3,
			Description: "Cobre a área com granizo por 2 turnos: 25 + 50% INT por turno em todos. Cada turno aplica lentidão cumulativa (-20% velocidade por turno). Garante congelamento no 2° turno.",
			AppliesStatus: "freeze", AppliesStatusTurns: 2,
		},
		{
			// Substituído: "Zero Absoluto" era duplicata do legado m_absolute_zero.
			// Novo: Extinção Glacial — reduz temperatura ao zero absoluto em área.
			ID: "mg_absolute_zero", Class: "mage", Branch: "crionita", Tier: 5,
			Name: "Extinção Glacial", Emoji: "🌌", PointCost: 4, RequiredLevel: 78,
			Requires: "mg_blizzard", MPCost: 90, EnergyCost: 20,
			Damage: 80, DamageType: "ice",
			Role: RoleUlt, Scaling: 1.0, Cooldown: 8,
			Description: "Ultimate: congela tudo ao redor — 80 + 100% INT em área. Alvos congelados recebem ×2.5 dano e ficam paralisados por 3 turnos. Pode congelar chefes que normalmente são imunes.",
			RequiresStatus: "freeze", SynergyMult: 1.5,
		},

		// ── ARCANISTA (arcane — scaling + mana) ───────────────────────────
		{
			ID: "mg_mana_surge", Class: "mage", Branch: "arcanista", Tier: 3,
			Name: "Surto de Mana", Emoji: "💜", PointCost: 2, RequiredLevel: 30,
			Requires: "m_arcane_blast", MPCost: 0,
			Role: RoleBuff, Passive: false,
			Description: "Canaliza por 1 turno: recupera 40% do MP máximo e dobra o dano do próximo feitiço.",
			AppliesStatus: "shield", AppliesStatusTurns: 1,
		},
		{
			ID: "mg_arcane_barrage", Class: "mage", Branch: "arcanista", Tier: 3,
			Name: "Barragem Arcana", Emoji: "✨", PointCost: 2, RequiredLevel: 35,
			Requires: "mg_mana_surge", MPCost: 40, Damage: 50, DamageType: "magic",
			Role: RoleDirect, Scaling: 0.9,
			Description: "Dispara 3 projéteis arcanos: cada um causa 50 + 90% INT. Multiplicados se em surto de mana.",
		},
		{
			ID: "mg_arcane_rupture", Class: "mage", Branch: "arcanista", Tier: 4,
			Name: "Ruptura Arcana", Emoji: "🌀", PointCost: 3, RequiredLevel: 55,
			Requires: "mg_arcane_barrage", MPCost: 55, Damage: 70, DamageType: "magic",
			Role: RoleDebuff, Scaling: 0.8, Cooldown: 2,
			Description: "Rasga as defesas mágicas do alvo: 70 + 80% INT. Reduz resistência mágica do alvo em 40% por 3 turnos.",
			AppliesStatus: "curse", AppliesStatusTurns: 3,
		},
		{
			ID: "mg_annihilation", Class: "mage", Branch: "arcanista", Tier: 5,
			Name: "Aniquilação Arcana", Emoji: "💥", PointCost: 4, RequiredLevel: 80,
			Requires: "mg_arcane_rupture", MPCost: 100, EnergyCost: 20,
			Damage: 150, DamageType: "magic",
			Role: RoleUlt, Scaling: 1.5, Cooldown: 10,
			Description: "Ultimate: canaliza poder arcano absoluto — 150 + 150% INT. Dano adicional +1% por cada ponto de MP atual. Ignora resistências.",
		},
	}
}
