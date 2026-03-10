package skills

import "github.com/tormenta-bot/internal/models"

// ArcherSkills returns 15 new archer/caçador skills.
// Legacy: a_precise_shot, a_multishot, a_trap, etc. (data.go)
// IDs use "ac_" prefix.
//
// Branches: "atirador" (precision), "cacador" (traps/control), "arcano" (magic arrows)
// Synergy: Mark → Marked → Power Shot ×1.5
// Builds: Sniper (execute+crit), Trapper (control+DoT), Arcane Archer (magic+debuff)
func ArcherSkills() []models.Skill {
	return []models.Skill{
		// ── ATIRADOR (precisão + execute) ─────────────────────────────────
		{
			ID: "ac_piercing_shot", Class: "archer", Branch: "atirador", Tier: 3,
			Name: "Tiro Perfurante", Emoji: "🎯", PointCost: 2, RequiredLevel: 30,
			Requires: "a_precise_shot", MPCost: 25, Damage: 55, DamageType: "physical",
			Role: RoleDirect, Scaling: 0.8,
			Description: "Tiro que atravessa armadura: 55 + 80% DEX. Ignora 40% da defesa do alvo. Aplica sangramento 2 turnos.",
			AppliesStatus: "bleed", AppliesStatusTurns: 2,
		},
		{
			// Substituído: "Tiro na Cabeça" era duplicata do legado a_headshot.
			// Novo: Execução Balística — tiro preciso no ponto vital com execute threshold.
			ID: "ac_headshot", Class: "archer", Branch: "atirador", Tier: 4,
			Name: "Execução Balística", Emoji: "💥", PointCost: 3, RequiredLevel: 50,
			Requires: "ac_piercing_shot", MPCost: 40, Damage: 80, DamageType: "physical",
			Role: RoleDirect, Scaling: 1.0, Cooldown: 2,
			Description: "Mira calculada no ponto vital: 80 + 100% DEX. Alvo em sangramento: sempre crítico + dano +50%. +25% dano por turno de sangramento restante.",
			RequiresStatus: "bleed", SynergyMult: 0.5,
		},
		{
			ID: "ac_divine_arrow", Class: "archer", Branch: "atirador", Tier: 5,
			Name: "Flecha Divina", Emoji: "✨", PointCost: 4, RequiredLevel: 75,
			Requires: "ac_headshot", MPCost: 70, EnergyCost: 15,
			Damage: 120, DamageType: "physical",
			Role: RoleUlt, Scaling: 1.4, Cooldown: 7,
			Description: "Ultimate: flecha carregada de luz — 120 + 140% DEX. Dobra o dano se inimigo com < 50% HP. Não pode errar.",
		},

		// ── CAÇADOR (armadilhas + controle) ───────────────────────────────
		{
			ID: "ac_mark_prey", Class: "archer", Branch: "cacador", Tier: 3,
			Name: "Marcar Presa", Emoji: "🔍", PointCost: 2, RequiredLevel: 30,
			Requires: "a_trap", MPCost: 15,
			Role: RoleDebuff, Cooldown: 0,
			Description: "Marca o alvo: reduz 20% DEF e evasão por 3 turnos. Próximos ataques físicos recebem +15% dano.",
			AppliesStatus: "curse", AppliesStatusTurns: 3,
		},
		{
			ID: "ac_net_trap", Class: "archer", Branch: "cacador", Tier: 3,
			Name: "Rede Armadilha", Emoji: "🕸️", PointCost: 2, RequiredLevel: 35,
			Requires: "ac_mark_prey", MPCost: 25, Damage: 20, DamageType: "physical",
			Role: RoleControl, Scaling: 0.3, Cooldown: 3,
			Description: "Dispara rede: 20 + 30% DEX. Imobiliza o alvo por 2 turnos (não pode atacar ou fugir).",
			AppliesStatus: "stun", AppliesStatusTurns: 2,
		},
		{
			ID: "ac_predator_instinct", Class: "archer", Branch: "cacador", Tier: 4,
			Name: "Instinto de Predador", Emoji: "🐆", PointCost: 3, RequiredLevel: 55,
			Requires: "ac_net_trap", MPCost: 35,
			Role: RoleBuff, Passive: false, Cooldown: 4,
			Description: "Postura de caçador por 3 turnos: +40% DEX, +25% crit, recupera 5 MP por acerto.",
		},
		{
			ID: "ac_fatal_hunt", Class: "archer", Branch: "cacador", Tier: 5,
			Name: "Caçada Fatal", Emoji: "🏹", PointCost: 4, RequiredLevel: 78,
			Requires: "ac_predator_instinct", MPCost: 80, EnergyCost: 20,
			Damage: 60, DamageType: "physical",
			Role: RoleUlt, Scaling: 1.1, Cooldown: 8,
			Description: "Ultimate: persegue o alvo com 4 tiros rápidos (60 + 110% DEX cada). Alvo marcado recebe ×1.5 dano total.",
			RequiresStatus: "curse", SynergyMult: 0.5,
		},

		// ── ARCANO (flechas mágicas + debuffs) ────────────────────────────
		{
			// Substituído: "Flecha de Gelo" era duplicata do legado a_frost_arrow.
			// Novo: Flecha do Vácuo — flecha dimensional que silencia e ignora barreiras.
			ID: "ac_frost_arrow", Class: "archer", Branch: "arcano", Tier: 3,
			Name: "Flecha do Vácuo", Emoji: "🌌", PointCost: 2, RequiredLevel: 30,
			Requires: "a_arcane_shot", MPCost: 30, Damage: 45, DamageType: "magic",
			Role: RoleDirect, Scaling: 0.7,
			Description: "Flecha infundida com energia do vácuo dimensional: 45 + 70% DEX/INT. Atravessa barreiras mágicas e silencia o alvo por 1 turno (impede habilidades mágicas).",
			AppliesStatus: "silence", AppliesStatusTurns: 1,
		},
		{
			ID: "ac_arcane_volley", Class: "archer", Branch: "arcano", Tier: 4,
			Name: "Saraivada Arcana", Emoji: "💫", PointCost: 3, RequiredLevel: 55,
			Requires: "ac_frost_arrow", MPCost: 50, Damage: 30, DamageType: "magic",
			Role: RoleAoE, Scaling: 0.6, Cooldown: 2,
			Description: "Dispara saraivada de flechas mágicas em área: 30 + 60% DEX cada. Silencia 1 turno.",
			AppliesStatus: "silence", AppliesStatusTurns: 1,
		},
		{
			ID: "ac_tiro_fatal", Class: "archer", Branch: "arcano", Tier: 5,
			Name: "Tiro Fatal", Emoji: "🌠", PointCost: 4, RequiredLevel: 80,
			Requires: "ac_arcane_volley", MPCost: 90, EnergyCost: 20,
			Damage: 100, DamageType: "magic",
			Role: RoleUlt, Scaling: 1.3, Cooldown: 9,
			Description: "Ultimate: concentra toda energia arcana em uma única flecha — 100 + 130% DEX/INT (maior stat). Inimigos silenciados recebem ×2.",
			RequiresStatus: "silence", SynergyMult: 1.0,
		},
	}
}
