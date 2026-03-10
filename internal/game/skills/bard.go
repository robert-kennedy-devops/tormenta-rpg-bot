package skills

import "github.com/tormenta-bot/internal/models"

// BardSkills returns 15 new bard skills (class had ZERO before).
// IDs use "bard2_" prefix.
//
// Branches: "musica" (buffs/support), "conhecimento" (debuffs/utility), "ilusao" (control/misdirection)
// Synergy: Dissonance → confused → Grand Illusion triples effect
// Builds: Maestro (full buff support), Dissonant (debuff+control), Illusionist (confusion+stun)
func BardSkills() []models.Skill {
	return []models.Skill{
		// ── MÚSICA (buff + suporte) ────────────────────────────────────────
		{
			ID: "bard2_inspire", Class: "bard", Branch: "musica", Tier: 1,
			Name: "Inspiração", Emoji: "🎵", PointCost: 1, RequiredLevel: 1,
			MPCost: 10, Role: RoleBuff,
			Description: "Toca nota inspiradora: aliados recebem +20% ATK por 2 turnos. Recupera 5 MP por turno por 2 turnos.",
			AppliesStatus: "shield", AppliesStatusTurns: 2,
		},
		{
			ID: "bard2_battle_hymn", Class: "bard", Branch: "musica", Tier: 2,
			Name: "Hino de Batalha", Emoji: "🎶", PointCost: 1, RequiredLevel: 8,
			Requires: "bard2_inspire", MPCost: 20,
			Role: RoleBuff,
			Description: "Hino que energiza: +30% ATK e +20% DEF a todos aliados por 3 turnos. Inspira cura de 10 HP/turno.",
		},
		{
			ID: "bard2_power_chord", Class: "bard", Branch: "musica", Tier: 3,
			Name: "Acorde de Poder", Emoji: "🎸", PointCost: 2, RequiredLevel: 28,
			Requires: "bard2_battle_hymn", MPCost: 30, Damage: 30, DamageType: "magic",
			Role: RoleDirect, Scaling: 0.5,
			Description: "Acorde devastador: 30 + 50% CHA como dano sônico. Aliados em buffados atacam junto (30% extra).",
		},
		{
			ID: "bard2_anthem_victory", Class: "bard", Branch: "musica", Tier: 4,
			Name: "Anthem da Vitória", Emoji: "🎺", PointCost: 3, RequiredLevel: 50,
			Requires: "bard2_power_chord", MPCost: 45,
			Role: RoleBuff, Cooldown: 4,
			Description: "Anthem poderoso por 4 turnos: +50% ATK, +30% DEF, +25% velocidade a TODOS aliados. Recupera 15 HP/turno.",
		},
		{
			ID: "bard2_legendary_song", Class: "bard", Branch: "musica", Tier: 5,
			Name: "Canção das Lendas", Emoji: "🎼", PointCost: 4, RequiredLevel: 72,
			Requires: "bard2_anthem_victory", MPCost: 80, EnergyCost: 20,
			Role: RoleUlt, Cooldown: 8,
			Description: "Ultimate: canção lendária por 5 turnos — todos aliados têm ATK/DEF/HP MAX dobrados e recebem +100% XP ao fim da batalha.",
		},

		// ── CONHECIMENTO (debuff + utility) ───────────────────────────────
		{
			ID: "bard2_mock", Class: "bard", Branch: "conhecimento", Tier: 1,
			Name: "Escárnio", Emoji: "😝", PointCost: 1, RequiredLevel: 3,
			MPCost: 12, Role: RoleDebuff,
			Description: "Debilita o inimigo com sarcasmo: -20% ATK por 3 turnos. Aplica confusão (10% chance de errar ataque).",
			AppliesStatus: "blind", AppliesStatusTurns: 1,
		},
		{
			ID: "bard2_discordance", Class: "bard", Branch: "conhecimento", Tier: 2,
			Name: "Discordância", Emoji: "🔊", PointCost: 1, RequiredLevel: 10,
			Requires: "bard2_mock", MPCost: 20, Damage: 20, DamageType: "magic",
			Role: RoleDebuff, Scaling: 0.4,
			Description: "Som dissonante: 20 + 40% CHA. Silencia o alvo 2 turnos e reduz -30% todos stats.",
			AppliesStatus: "silence", AppliesStatusTurns: 2,
		},
		{
			ID: "bard2_lore_master", Class: "bard", Branch: "conhecimento", Tier: 3,
			Name: "Mestre do Saber", Emoji: "📚", PointCost: 2, RequiredLevel: 28,
			Requires: "bard2_discordance", MPCost: 0, Passive: true,
			Role: RolePassive,
			Description: "Passivo: +15% de todo XP ganho. Identifica pontos fracos — +10% dano em alvos silenciados.",
		},
		{
			ID: "bard2_tale_of_weakness", Class: "bard", Branch: "conhecimento", Tier: 4,
			Name: "Conto da Fraqueza", Emoji: "📖", PointCost: 3, RequiredLevel: 52,
			Requires: "bard2_lore_master", MPCost: 35,
			Role: RoleDebuff, Cooldown: 2,
			Description: "Recita fraquezas: alvo fica com -40% de todos stats e -50% resistências por 3 turnos. Silenciados recebem dobrado.",
			RequiresStatus: "silence", SynergyMult: 1.0,
			AppliesStatus: "curse", AppliesStatusTurns: 3,
		},
		{
			ID: "bard2_finale", Class: "bard", Branch: "conhecimento", Tier: 5,
			Name: "Final da Ópera", Emoji: "🎭", PointCost: 4, RequiredLevel: 75,
			Requires: "bard2_tale_of_weakness", MPCost: 70, EnergyCost: 15,
			Damage: 80, DamageType: "magic",
			Role: RoleUlt, Scaling: 1.1, Cooldown: 7,
			Description: "Ultimate: nota final: 80 + 110% CHA. Malditos recebem ×2. Atordoa todos por 2 turnos.",
			RequiresStatus: "curse", SynergyMult: 1.0,
			AppliesStatus: "stun", AppliesStatusTurns: 2,
		},

		// ── ILUSÃO (controle + misdirection) ──────────────────────────────
		{
			ID: "bard2_mirage", Class: "bard", Branch: "ilusao", Tier: 1,
			Name: "Miragem", Emoji: "🪄", PointCost: 1, RequiredLevel: 5,
			MPCost: 15, Role: RoleControl,
			Description: "Cria ilusão que engana o inimigo: 30% chance de inimigo atacar ilusão (erro). Dura 2 turnos.",
			AppliesStatus: "blind", AppliesStatusTurns: 2,
		},
		{
			ID: "bard2_lullaby", Class: "bard", Branch: "ilusao", Tier: 2,
			Name: "Cantiga do Sono", Emoji: "😴", PointCost: 1, RequiredLevel: 12,
			Requires: "bard2_mirage", MPCost: 25,
			Role: RoleControl, Cooldown: 3,
			Description: "Melodia sonolenta: atordoa o inimigo por 1 turno (resistência: SAB). Em combate, 50% chance de adicionar mais 1 turno.",
			AppliesStatus: "stun", AppliesStatusTurns: 1,
		},
		{
			ID: "bard2_illusion_clone", Class: "bard", Branch: "ilusao", Tier: 3,
			Name: "Clone Ilusório", Emoji: "👥", PointCost: 2, RequiredLevel: 30,
			Requires: "bard2_lullaby", MPCost: 35,
			Role: RoleUtility, Cooldown: 4,
			Description: "Cria clone que absorve próximo ataque e explode causando 40% do dano em área.",
		},
		{
			ID: "bard2_mass_confusion", Class: "bard", Branch: "ilusao", Tier: 4,
			Name: "Confusão em Massa", Emoji: "🌀", PointCost: 3, RequiredLevel: 52,
			Requires: "bard2_illusion_clone", MPCost: 50,
			Role: RoleControl, Cooldown: 4,
			Description: "Ilusão coletiva: confunde TODOS os inimigos por 2 turnos (50% chance de atacar aliado ou a si mesmo).",
			AppliesStatus: "blind", AppliesStatusTurns: 2,
		},
		{
			ID: "bard2_grand_illusion", Class: "bard", Branch: "ilusao", Tier: 5,
			Name: "Grande Ilusão", Emoji: "🎆", PointCost: 4, RequiredLevel: 78,
			Requires: "bard2_mass_confusion", MPCost: 90, EnergyCost: 25,
			Role: RoleUlt, Cooldown: 9,
			Description: "Ultimate: cria realidade alternativa por 3 turnos — inimigos ficam completamente confusos (atacam aleatoriamente), você fica invisível e regenera 20% HP/turno.",
		},
	}
}
