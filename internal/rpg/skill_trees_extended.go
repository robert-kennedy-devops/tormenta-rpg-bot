package rpg

import engine "github.com/tormenta-bot/internal/engine"

// init runs after skill_tree.go init() (alphabetical file order).
// Extends existing trees and registers all passive bonuses.
func init() {
	extendBarbarian()
	extendPaladin()
	extendCleric()
	extendBard()
	registerAllPassives()
}

// ─── Barbarian extension ──────────────────────────────────────────────────────
// Adds 17 new skills: 5 to Fúria, 4 to Resistência, 8 new Tribal branch.

func extendBarbarian() {
	t, ok := Trees["barbarian"]
	if !ok {
		return
	}
	// Extend branch 0 (Fúria) — 5 new nodes
	t.Branches[0].Nodes = append(t.Branches[0].Nodes,
		SkillNode{
			ID: "barb_berserker_fury", Class: "barbarian", Branch: "furia",
			Name: "Fúria Berserker", Emoji: "🔥", Tier: 3, PointCost: 2, RequiredLevel: 20,
			MPCost: 10, EnergyCost: 12,
			Description: "Fúria total: Berserk por 4 turnos e cada golpe crítico regenera 8 HP.",
			Requires: []string{"barb_blood_thirst"},
			Effects: []engine_Effect{
				{TypeName: "apply_status", StatusKind: "berserk", StatusTurns: 4},
			},
		},
		SkillNode{
			ID: "barb_wild_swing", Class: "barbarian", Branch: "furia",
			Name: "Pancada Selvagem", Emoji: "🌪️", Tier: 3, PointCost: 2, RequiredLevel: 25,
			MPCost: 20, EnergyCost: 14,
			Description: "Pancada descontrolada: 4d12+FOR de dano físico. 20% de chance de auto-stun por 1 turno.",
			Requires: []string{"barb_berserker_fury"},
			Effects: []engine_Effect{
				{TypeName: "damage", Element: "physical", BasePower: 65},
			},
		},
		SkillNode{
			ID: "barb_carnage", Class: "barbarian", Branch: "furia",
			Name: "Carnificina", Emoji: "🩸", Tier: 4, PointCost: 2, RequiredLevel: 40,
			MPCost: 30, EnergyCost: 16,
			Description: "Ataca 2 vezes em AoE causando 3d10+FOR cada. Sangramento em todos os atingidos por 4 turnos.",
			Requires: []string{"barb_rampage"},
			Effects: []engine_Effect{
				{TypeName: "aoe", Element: "physical", BasePower: 45},
				{TypeName: "aoe", Element: "physical", BasePower: 45},
				{TypeName: "apply_status", StatusKind: "poison", StatusTurns: 4, StatusDmgPT: 5},
			},
		},
		SkillNode{
			ID: "barb_blood_frenzy", Class: "barbarian", Branch: "furia",
			Name: "Frenesi Sanguinário", Emoji: "💀", Tier: 5, PointCost: 3, RequiredLevel: 65,
			MPCost: 0, EnergyCost: 20,
			Description: "Quando HP cair abaixo de 30%, entra em Frenesi: +150% ATK, regenera 20 HP/turno por 3 turnos.",
			Requires: []string{"barb_carnage"},
			IsPassive: true,
		},
		SkillNode{
			ID: "barb_warlord", Class: "barbarian", Branch: "furia",
			Name: "Senhor da Guerra", Emoji: "👑", Tier: 5, PointCost: 5, RequiredLevel: 80,
			MPCost: 50, EnergyCost: 30, IsUltimate: true,
			Description: "Forma final do Bárbaro: 10d20+FOR de dano físico. Por 1 turno, é invencível (0 dano recebido).",
			Requires: []string{"barb_blood_frenzy"},
			Effects: []engine_Effect{
				{TypeName: "damage", Element: "physical", BasePower: 150},
				{TypeName: "apply_status", StatusKind: "protect", StatusTurns: 1},
				{TypeName: "apply_status", StatusKind: "berserk", StatusTurns: 3},
			},
		},
	)

	// Extend branch 1 (Resistência) — 4 new nodes
	t.Branches[1].Nodes = append(t.Branches[1].Nodes,
		SkillNode{
			ID: "barb_battle_hardened", Class: "barbarian", Branch: "resistencia",
			Name: "Endurecido em Batalha", Emoji: "🪖", Tier: 2, PointCost: 1, RequiredLevel: 15,
			IsPassive: true,
			Description: "Passivo: +10 DEF e -15% de todos os danos recebidos.",
			Requires: []string{"barb_stone_skin"},
		},
		SkillNode{
			ID: "barb_immortal_flesh", Class: "barbarian", Branch: "resistencia",
			Name: "Carne Imortal", Emoji: "🧱", Tier: 3, PointCost: 2, RequiredLevel: 28,
			IsPassive: true,
			Description: "Passivo: +80 HP máximo. Regenera 8 HP/turno fora de combate e 3 HP/turno em combate.",
			Requires: []string{"barb_battle_hardened"},
		},
		SkillNode{
			ID: "barb_pain_defiance", Class: "barbarian", Branch: "resistencia",
			Name: "Desafiar a Dor", Emoji: "💪", Tier: 4, PointCost: 2, RequiredLevel: 45,
			MPCost: 0, EnergyCost: 10,
			Description: "Resiste a controles: reduz duração de Stun/Freeze/Silêncio pela metade. Passivo: +5 de todos os resistências.",
			Requires: []string{"barb_immortal_flesh"},
			IsPassive: true,
		},
		SkillNode{
			ID: "barb_mountain_stance", Class: "barbarian", Branch: "resistencia",
			Name: "Postura da Montanha", Emoji: "⛰️", Tier: 5, PointCost: 4, RequiredLevel: 60,
			MPCost: 40, EnergyCost: 20, IsUltimate: true,
			Description: "Postura rochosa por 4 turnos: -80% de todo dano recebido, imune a CC, +50% ATK e regenera 25 HP/turno.",
			Requires: []string{"barb_pain_defiance"},
			Effects: []engine_Effect{
				{TypeName: "apply_status", StatusKind: "protect", StatusTurns: 4},
				{TypeName: "apply_status", StatusKind: "regen", StatusTurns: 4},
			},
		},
	)

	// Add branch 2 — Guerreiro Tribal (8 nodes)
	t.Branches = append(t.Branches, Branch{
		ID: "guerreiro_tribal", Name: "Guerreiro Tribal", Emoji: "🥁",
		Nodes: []SkillNode{
			{
				ID: "barb_war_paint", Class: "barbarian", Branch: "guerreiro_tribal",
				Name: "Pintura de Guerra", Emoji: "🎨", Tier: 1, PointCost: 1, RequiredLevel: 5,
				MPCost: 0, EnergyCost: 5,
				Description: "Pinta o corpo com símbolos tribais: +25% ATK e intimida o inimigo (-20% ATK deles) por 3 turnos.",
				Effects: []engine_Effect{
					{TypeName: "apply_status", StatusKind: "berserk", StatusTurns: 2},
					{TypeName: "stat_debuff", StatName: "attack", StatDelta: -6},
				},
			},
			{
				ID: "barb_tribal_drums", Class: "barbarian", Branch: "guerreiro_tribal",
				Name: "Tambores Tribais", Emoji: "🥁", Tier: 1, PointCost: 1, RequiredLevel: 10,
				MPCost: 15, EnergyCost: 6,
				Description: "Ritmando os tambores ancestrais: Haste por 3 turnos para si e Berserk por 2 turnos.",
				Requires: []string{"barb_war_paint"},
				Effects: []engine_Effect{
					{TypeName: "apply_status", StatusKind: "haste", StatusTurns: 3},
					{TypeName: "apply_status", StatusKind: "berserk", StatusTurns: 2},
				},
			},
			{
				ID: "barb_totem_strike", Class: "barbarian", Branch: "guerreiro_tribal",
				Name: "Golpe do Totem", Emoji: "🗿", Tier: 2, PointCost: 1, RequiredLevel: 18,
				MPCost: 25, EnergyCost: 8,
				Description: "Canal energia do totem natural: 3d10+FOR de dano de veneno e aplica Maldição por 3 turnos.",
				Requires: []string{"barb_tribal_drums"},
				Effects: []engine_Effect{
					{TypeName: "damage", Element: "poison", BasePower: 45},
					{TypeName: "apply_status", StatusKind: "curse", StatusTurns: 3},
				},
			},
			{
				ID: "barb_blood_ritual", Class: "barbarian", Branch: "guerreiro_tribal",
				Name: "Ritual de Sangue", Emoji: "🩸", Tier: 2, PointCost: 2, RequiredLevel: 25,
				MPCost: 0, EnergyCost: 15,
				Description: "Sacrifica 15% do HP atual: ganhe +60% ATK por 4 turnos e Berserk imediatamente.",
				Requires: []string{"barb_totem_strike"},
				Effects: []engine_Effect{
					{TypeName: "apply_status", StatusKind: "berserk", StatusTurns: 4},
				},
			},
			{
				ID: "barb_totemic_call", Class: "barbarian", Branch: "guerreiro_tribal",
				Name: "Chamado Totêmico", Emoji: "🦁", Tier: 3, PointCost: 2, RequiredLevel: 35,
				MPCost: 35, EnergyCost: 12,
				Description: "Chama os espíritos dos ancestrais: Regen por 5 turnos (10 HP/turno) e +40% DEF.",
				Requires: []string{"barb_blood_ritual"},
				Effects: []engine_Effect{
					{TypeName: "apply_status", StatusKind: "regen", StatusTurns: 5},
					{TypeName: "apply_status", StatusKind: "protect", StatusTurns: 3},
				},
			},
			{
				ID: "barb_ancestral_rage", Class: "barbarian", Branch: "guerreiro_tribal",
				Name: "Fúria Ancestral", Emoji: "👻", Tier: 3, PointCost: 2, RequiredLevel: 45,
				MPCost: 40, EnergyCost: 16,
				Description: "Canaliza fúria de todos os ancestrais: 5d12+FOR de dano físico em AoE + Berserk por 3 turnos.",
				Requires: []string{"barb_totemic_call"},
				Effects: []engine_Effect{
					{TypeName: "aoe", Element: "physical", BasePower: 75},
					{TypeName: "apply_status", StatusKind: "berserk", StatusTurns: 3},
				},
			},
			{
				ID: "barb_spirit_warrior", Class: "barbarian", Branch: "guerreiro_tribal",
				Name: "Guerreiro Espiritual", Emoji: "💫", Tier: 4, PointCost: 3, RequiredLevel: 60,
				MPCost: 60, EnergyCost: 20,
				Description: "Fundição com o espírito ancestral por 4 turnos: Berserk+Haste+Regen simultâneos, +80% a todo dano.",
				Requires: []string{"barb_ancestral_rage"},
				Effects: []engine_Effect{
					{TypeName: "apply_status", StatusKind: "berserk", StatusTurns: 4},
					{TypeName: "apply_status", StatusKind: "haste", StatusTurns: 4},
					{TypeName: "apply_status", StatusKind: "regen", StatusTurns: 4},
				},
			},
			{
				ID: "barb_titan_slam", Class: "barbarian", Branch: "guerreiro_tribal",
				Name: "Impacto do Titã", Emoji: "💥", Tier: 5, PointCost: 5, RequiredLevel: 80,
				MPCost: 70, EnergyCost: 30, IsUltimate: true,
				Description: "O golpe mais devastador da tribo: 12d20+FOR de dano físico em AoE massiva. Stun por 3 turnos a todos.",
				Requires: []string{"barb_spirit_warrior"},
				Effects: []engine_Effect{
					{TypeName: "aoe", Element: "physical", BasePower: 180},
					{TypeName: "apply_status", StatusKind: "stun", StatusTurns: 3},
				},
			},
		},
	})
}

// ─── Paladin extension ────────────────────────────────────────────────────────
// Adds 23 new skills: 5 to Sagrado, 9 new Proteção, 9 new Julgamento.

func extendPaladin() {
	t, ok := Trees["paladin"]
	if !ok {
		return
	}
	// Extend branch 0 (Sagrado) — 5 new nodes
	t.Branches[0].Nodes = append(t.Branches[0].Nodes,
		SkillNode{
			ID: "pal_bless", Class: "paladin", Branch: "sagrado",
			Name: "Abençoar", Emoji: "✨", Tier: 2, PointCost: 1, RequiredLevel: 8,
			MPCost: 15, EnergyCost: 3,
			Description: "Abençoa a si mesmo: +20% ATK e +15% DEF por 4 turnos. Remove 1 status negativo.",
			Requires: []string{"pal_smite"},
			Effects: []engine_Effect{
				{TypeName: "apply_status", StatusKind: "protect", StatusTurns: 2},
				{TypeName: "remove_status", StatusKind: "curse"},
			},
		},
		SkillNode{
			ID: "pal_consecrate", Class: "paladin", Branch: "sagrado",
			Name: "Consagrar", Emoji: "⭐", Tier: 3, PointCost: 2, RequiredLevel: 18,
			MPCost: 35, EnergyCost: 8,
			Description: "Consagra o campo de batalha com luz sagrada: 3d8+SAB de dano sagrado em AoE. Dano dobrado contra mortos-vivos.",
			Requires: []string{"pal_bless"},
			Effects: []engine_Effect{
				{TypeName: "aoe", Element: "holy", BasePower: 38},
			},
		},
		SkillNode{
			ID: "pal_aura_protection", Class: "paladin", Branch: "sagrado",
			Name: "Aura de Proteção", Emoji: "🌟", Tier: 3, PointCost: 2, RequiredLevel: 20,
			IsPassive: true,
			Description: "Passivo: Aura sagrada permanente reduz dano recebido em 10% e +8 DEF.",
			Requires: []string{"pal_consecrate"},
		},
		SkillNode{
			ID: "pal_holy_nova", Class: "paladin", Branch: "sagrado",
			Name: "Nova Sagrada", Emoji: "☀️", Tier: 4, PointCost: 2, RequiredLevel: 40,
			MPCost: 60, EnergyCost: 14,
			Description: "Explosão de luz sagrada em AoE: 5d10+SAB de dano sagrado e cura 2d8 HP do Paladino.",
			Requires: []string{"pal_aura_protection"},
			Effects: []engine_Effect{
				{TypeName: "aoe", Element: "holy", BasePower: 65},
				{TypeName: "heal", BasePower: 25},
			},
		},
		SkillNode{
			ID: "pal_resurrection", Class: "paladin", Branch: "sagrado",
			Name: "Ressurreição Divina", Emoji: "🌟", Tier: 5, PointCost: 5, RequiredLevel: 75,
			MPCost: 100, EnergyCost: 30, IsUltimate: true,
			Description: "Milagre supremo: ressuscita com 70% HP, remove todos os status negativos e aplica Escudo e Proteção por 3 turnos.",
			Requires: []string{"pal_holy_nova"},
			Effects: []engine_Effect{
				{TypeName: "heal", BasePower: 100},
				{TypeName: "remove_status", StatusKind: "poison"},
				{TypeName: "apply_status", StatusKind: "shield", StatusTurns: 3},
				{TypeName: "apply_status", StatusKind: "protect", StatusTurns: 3},
			},
		},
	)

	// Add branch 1 — Proteção (9 nodes)
	t.Branches = append(t.Branches, Branch{
		ID: "protecao", Name: "Proteção", Emoji: "🛡️",
		Nodes: []SkillNode{
			{
				ID: "pal_guard", Class: "paladin", Branch: "protecao",
				Name: "Guardar", Emoji: "🛡️", Tier: 1, PointCost: 1, RequiredLevel: 1,
				MPCost: 12, EnergyCost: 3,
				Description: "Assume posição defensiva: aplica Escudo por 2 turnos e contra-ataca com golpe de 1d6+FOR.",
				Effects: []engine_Effect{
					{TypeName: "apply_status", StatusKind: "shield", StatusTurns: 2},
					{TypeName: "damage", Element: "holy", BasePower: 14},
				},
			},
			{
				ID: "pal_shield_of_faith", Class: "paladin", Branch: "protecao",
				Name: "Escudo da Fé", Emoji: "✝️", Tier: 1, PointCost: 1, RequiredLevel: 5,
				IsPassive: true,
				Description: "Passivo: +12 DEF e +30 HP máximo. Fé inabalável reduz dano mágico em 15%.",
				Requires: []string{"pal_guard"},
			},
			{
				ID: "pal_intervene", Class: "paladin", Branch: "protecao",
				Name: "Intervir", Emoji: "🙅", Tier: 2, PointCost: 1, RequiredLevel: 10,
				MPCost: 20, EnergyCost: 4,
				Description: "Intercepta o próximo ataque recebido, reduzindo-o em 70% e refletindo 20% do dano.",
				Requires: []string{"pal_shield_of_faith"},
				Effects: []engine_Effect{
					{TypeName: "apply_status", StatusKind: "shield", StatusTurns: 1},
				},
			},
			{
				ID: "pal_divine_aegis", Class: "paladin", Branch: "protecao",
				Name: "Égide Divina", Emoji: "⚜️", Tier: 2, PointCost: 2, RequiredLevel: 15,
				MPCost: 30, EnergyCost: 6,
				Description: "Escudo divino massivo que absorve até 80 pontos de dano por 3 turnos.",
				Requires: []string{"pal_intervene"},
				Effects: []engine_Effect{
					{TypeName: "apply_status", StatusKind: "shield", StatusTurns: 3},
					{TypeName: "apply_status", StatusKind: "protect", StatusTurns: 2},
				},
			},
			{
				ID: "pal_blessed_armor", Class: "paladin", Branch: "protecao",
				Name: "Armadura Abençoada", Emoji: "✨", Tier: 2, PointCost: 2, RequiredLevel: 20,
				IsPassive: true,
				Description: "Passivo: +20 DEF, +15 DEF mágica e resistência sagrada (+30% resistência a magia das trevas).",
				Requires: []string{"pal_divine_aegis"},
			},
			{
				ID: "pal_sanctuary", Class: "paladin", Branch: "protecao",
				Name: "Santuário", Emoji: "🏛️", Tier: 3, PointCost: 2, RequiredLevel: 28,
				MPCost: 45, EnergyCost: 10,
				Description: "Cria zona sagrada: imunidade total por 1 turno, depois Escudo+Proteção por 2 turnos.",
				Requires: []string{"pal_blessed_armor"},
				Effects: []engine_Effect{
					{TypeName: "apply_status", StatusKind: "protect", StatusTurns: 3},
					{TypeName: "apply_status", StatusKind: "shield", StatusTurns: 2},
				},
			},
			{
				ID: "pal_iron_fortress", Class: "paladin", Branch: "protecao",
				Name: "Fortaleza de Ferro", Emoji: "🏰", Tier: 3, PointCost: 2, RequiredLevel: 35,
				MPCost: 50, EnergyCost: 12,
				Description: "Transforma em fortaleza viva: +60% DEF, +60% DEF mágica, regenera 12 HP/turno por 4 turnos.",
				Requires: []string{"pal_sanctuary"},
				Effects: []engine_Effect{
					{TypeName: "apply_status", StatusKind: "protect", StatusTurns: 4},
					{TypeName: "apply_status", StatusKind: "regen", StatusTurns: 4},
				},
			},
			{
				ID: "pal_martyr", Class: "paladin", Branch: "protecao",
				Name: "Mártir", Emoji: "💔", Tier: 4, PointCost: 3, RequiredLevel: 45,
				MPCost: 60, EnergyCost: 15,
				Description: "Sacrifício épico: recebe todo o dano destinado a aliados e reflete 50% como dano sagrado.",
				Requires: []string{"pal_iron_fortress"},
				Effects: []engine_Effect{
					{TypeName: "damage", Element: "holy", BasePower: 60},
					{TypeName: "apply_status", StatusKind: "shield", StatusTurns: 2},
				},
			},
			{
				ID: "pal_divine_shield", Class: "paladin", Branch: "protecao",
				Name: "Escudo Divino", Emoji: "🌟", Tier: 5, PointCost: 5, RequiredLevel: 60,
				MPCost: 90, EnergyCost: 25, IsUltimate: true,
				Description: "Invulnerabilidade total por 2 turnos. Ao terminar, libera explosão sagrada de 6d10+SAB em AoE.",
				Requires: []string{"pal_martyr"},
				Effects: []engine_Effect{
					{TypeName: "apply_status", StatusKind: "protect", StatusTurns: 2},
					{TypeName: "apply_status", StatusKind: "shield", StatusTurns: 2},
					{TypeName: "aoe", Element: "holy", BasePower: 80},
				},
			},
		},
	})

	// Add branch 2 — Julgamento (9 nodes)
	t.Branches = append(t.Branches, Branch{
		ID: "julgamento", Name: "Julgamento", Emoji: "⚖️",
		Nodes: []SkillNode{
			{
				ID: "pal_mark_of_justice", Class: "paladin", Branch: "julgamento",
				Name: "Marca da Justiça", Emoji: "⚖️", Tier: 1, PointCost: 1, RequiredLevel: 1,
				MPCost: 12, EnergyCost: 3,
				Description: "Marca o alvo como culpado: -20% DEF e -20% ATK por 4 turnos.",
				Effects: []engine_Effect{
					{TypeName: "stat_debuff", StatName: "defense", StatDelta: -6},
					{TypeName: "stat_debuff", StatName: "attack", StatDelta: -5},
				},
			},
			{
				ID: "pal_expose", Class: "paladin", Branch: "julgamento",
				Name: "Expor", Emoji: "🔍", Tier: 2, PointCost: 1, RequiredLevel: 10,
				MPCost: 18, EnergyCost: 4,
				Description: "Expõe as fraquezas do alvo: -30% DEF e todos os próximos 3 ataques são críticos.",
				Requires: []string{"pal_mark_of_justice"},
				Effects: []engine_Effect{
					{TypeName: "stat_debuff", StatName: "defense", StatDelta: -10},
				},
			},
			{
				ID: "pal_holy_wrath", Class: "paladin", Branch: "julgamento",
				Name: "Ira Sagrada", Emoji: "😤", Tier: 2, PointCost: 2, RequiredLevel: 18,
				MPCost: 30, EnergyCost: 7,
				Description: "Descarga de ira sagrada: 3d8+SAB de dano sagrado + Cegueira por 2 turnos.",
				Requires: []string{"pal_expose"},
				Effects: []engine_Effect{
					{TypeName: "damage", Element: "holy", BasePower: 40},
					{TypeName: "apply_status", StatusKind: "blind", StatusTurns: 2},
				},
			},
			{
				ID: "pal_seals_judgment", Class: "paladin", Branch: "julgamento",
				Name: "Selos do Julgamento", Emoji: "📜", Tier: 3, PointCost: 2, RequiredLevel: 25,
				MPCost: 40, EnergyCost: 9,
				Description: "Aplica Silêncio + Maldição por 3 turnos. Causa 2d10+SAB de dano sagrado.",
				Requires: []string{"pal_holy_wrath"},
				Effects: []engine_Effect{
					{TypeName: "apply_status", StatusKind: "silence", StatusTurns: 3},
					{TypeName: "apply_status", StatusKind: "curse", StatusTurns: 3},
					{TypeName: "damage", Element: "holy", BasePower: 30},
				},
			},
			{
				ID: "pal_conviction", Class: "paladin", Branch: "julgamento",
				Name: "Convicção", Emoji: "💪", Tier: 3, PointCost: 2, RequiredLevel: 32,
				MPCost: 45, EnergyCost: 10,
				Description: "Fé absoluta no julgamento: 5d10+SAB de dano sagrado puro. Dano dobrado se alvo for maligno.",
				Requires: []string{"pal_seals_judgment"},
				Effects: []engine_Effect{
					{TypeName: "damage", Element: "holy", BasePower: 65},
				},
			},
			{
				ID: "pal_void_smite", Class: "paladin", Branch: "julgamento",
				Name: "Golpe do Vazio", Emoji: "⚡", Tier: 3, PointCost: 2, RequiredLevel: 40,
				MPCost: 55, EnergyCost: 12,
				Description: "Combinação de sagrado e trevas: 4d10 de dano sagrado + 4d10 de dano das trevas. Ignora armadura.",
				Requires: []string{"pal_conviction"},
				Effects: []engine_Effect{
					{TypeName: "damage", Element: "holy", BasePower: 55},
					{TypeName: "damage", Element: "dark", BasePower: 55},
				},
			},
			{
				ID: "pal_eternal_judgment", Class: "paladin", Branch: "julgamento",
				Name: "Julgamento Eterno", Emoji: "⚜️", Tier: 4, PointCost: 3, RequiredLevel: 50,
				MPCost: 70, EnergyCost: 16,
				Description: "Remove todos os buffs do alvo e causa 6d10+SAB de dano sagrado. Aplica Maldição permanente (5 turnos).",
				Requires: []string{"pal_void_smite"},
				Effects: []engine_Effect{
					{TypeName: "remove_status", StatusKind: "shield"},
					{TypeName: "damage", Element: "holy", BasePower: 80},
					{TypeName: "apply_status", StatusKind: "curse", StatusTurns: 5},
				},
			},
			{
				ID: "pal_crusader_strike", Class: "paladin", Branch: "julgamento",
				Name: "Golpe do Cruzado", Emoji: "⚔️", Tier: 4, PointCost: 3, RequiredLevel: 60,
				MPCost: 80, EnergyCost: 18,
				Description: "Investida sagrada devastadora: move e ataca com 7d12+SAB de dano sagrado. Stun por 2 turnos.",
				Requires: []string{"pal_eternal_judgment"},
				Effects: []engine_Effect{
					{TypeName: "damage", Element: "holy", BasePower: 105},
					{TypeName: "apply_status", StatusKind: "stun", StatusTurns: 2},
				},
			},
			{
				ID: "pal_armageddon", Class: "paladin", Branch: "julgamento",
				Name: "Armagedom", Emoji: "☀️", Tier: 5, PointCost: 5, RequiredLevel: 80,
				MPCost: 120, EnergyCost: 35, IsUltimate: true,
				Description: "Julgamento final divino: 12d20+SAB de dano sagrado em AoE massiva. Stun + Maldição + Silêncio por 4 turnos em todos.",
				Requires: []string{"pal_crusader_strike"},
				Effects: []engine_Effect{
					{TypeName: "aoe", Element: "holy", BasePower: 200},
					{TypeName: "apply_status", StatusKind: "stun", StatusTurns: 4},
					{TypeName: "apply_status", StatusKind: "curse", StatusTurns: 4},
					{TypeName: "apply_status", StatusKind: "silence", StatusTurns: 4},
				},
			},
		},
	})
}

// ─── Cleric extension ─────────────────────────────────────────────────────────
// Adds 23 new skills: 5 to Cura, 9 new Luz Divina, 9 new Proteção Divina.

func extendCleric() {
	t, ok := Trees["cleric"]
	if !ok {
		return
	}
	// Extend branch 0 (Cura) — 5 new nodes
	t.Branches[0].Nodes = append(t.Branches[0].Nodes,
		SkillNode{
			ID: "cler_minor_heal", Class: "cleric", Branch: "cura",
			Name: "Cura Menor", Emoji: "💚", Tier: 1, PointCost: 1, RequiredLevel: 3,
			MPCost: 8, EnergyCost: 2,
			Description: "Cura rápida e barata: recupera 1d6+SAB HP. Custo mínimo de mana.",
			Effects: []engine_Effect{
				{TypeName: "heal", BasePower: 14},
			},
		},
		SkillNode{
			ID: "cler_dispel_magic", Class: "cleric", Branch: "cura",
			Name: "Dissipar Magia", Emoji: "✨", Tier: 2, PointCost: 1, RequiredLevel: 12,
			MPCost: 25, EnergyCost: 4,
			Description: "Remove todos os status negativos do alvo (veneno, maldição, silêncio, cegueira).",
			Requires: []string{"cler_heal"},
			Effects: []engine_Effect{
				{TypeName: "remove_status", StatusKind: "poison"},
				{TypeName: "remove_status", StatusKind: "curse"},
				{TypeName: "remove_status", StatusKind: "silence"},
				{TypeName: "remove_status", StatusKind: "blind"},
			},
		},
		SkillNode{
			ID: "cler_greater_heal", Class: "cleric", Branch: "cura",
			Name: "Cura Maior", Emoji: "💖", Tier: 3, PointCost: 2, RequiredLevel: 25,
			MPCost: 45, EnergyCost: 8,
			Description: "Cura massiva: recupera 5d8+SAB HP. Aplica Regen por 3 turnos (5 HP/turno).",
			Requires: []string{"cler_dispel_magic"},
			Effects: []engine_Effect{
				{TypeName: "heal", BasePower: 55},
				{TypeName: "apply_status", StatusKind: "regen", StatusTurns: 3},
			},
		},
		SkillNode{
			ID: "cler_divine_intervention", Class: "cleric", Branch: "cura",
			Name: "Intervenção Divina", Emoji: "👼", Tier: 4, PointCost: 3, RequiredLevel: 50,
			MPCost: 70, EnergyCost: 15,
			Description: "Intervenção dos deuses: quando HP cair a 0, cura automaticamente 40% do HP máximo (1x por combate).",
			Requires: []string{"cler_greater_heal"},
			Effects: []engine_Effect{
				{TypeName: "heal", BasePower: 60},
				{TypeName: "apply_status", StatusKind: "regen", StatusTurns: 5},
			},
		},
		SkillNode{
			ID: "cler_miracle", Class: "cleric", Branch: "cura",
			Name: "Milagre", Emoji: "🌟", Tier: 5, PointCost: 5, RequiredLevel: 80,
			MPCost: 120, EnergyCost: 30, IsUltimate: true,
			Description: "Milagre sagrado: cura COMPLETAMENTE o HP de todos os aliados e remove todos os status negativos.",
			Requires: []string{"cler_divine_intervention"},
			Effects: []engine_Effect{
				{TypeName: "heal", BasePower: 200},
				{TypeName: "remove_status", StatusKind: "poison"},
				{TypeName: "remove_status", StatusKind: "curse"},
				{TypeName: "remove_status", StatusKind: "silence"},
				{TypeName: "apply_status", StatusKind: "regen", StatusTurns: 5},
				{TypeName: "apply_status", StatusKind: "protect", StatusTurns: 3},
			},
		},
	)

	// Add branch 1 — Luz Divina (9 nodes)
	t.Branches = append(t.Branches, Branch{
		ID: "luz_divina", Name: "Luz Divina", Emoji: "☀️",
		Nodes: []SkillNode{
			{
				ID: "cler_holy_smite", Class: "cleric", Branch: "luz_divina",
				Name: "Punição Sagrada", Emoji: "☀️", Tier: 1, PointCost: 1, RequiredLevel: 1,
				MPCost: 12, EnergyCost: 3,
				Description: "Punição sagrada: 2d6+SAB de dano sagrado. Dano +50% contra mortos-vivos e demônios.",
				Effects: []engine_Effect{
					{TypeName: "damage", Element: "holy", BasePower: 22},
				},
			},
			{
				ID: "cler_radiance", Class: "cleric", Branch: "luz_divina",
				Name: "Radiância", Emoji: "✨", Tier: 1, PointCost: 1, RequiredLevel: 5,
				MPCost: 15, EnergyCost: 3,
				Description: "Flash de luz sagrada: 1d8+SAB de dano e Cegueira por 2 turnos.",
				Requires: []string{"cler_holy_smite"},
				Effects: []engine_Effect{
					{TypeName: "damage", Element: "holy", BasePower: 16},
					{TypeName: "apply_status", StatusKind: "blind", StatusTurns: 2},
				},
			},
			{
				ID: "cler_light_burst", Class: "cleric", Branch: "luz_divina",
				Name: "Explosão de Luz", Emoji: "💥", Tier: 2, PointCost: 1, RequiredLevel: 12,
				MPCost: 25, EnergyCost: 5,
				Description: "Explosão de luz sagrada em AoE: 3d8+SAB de dano sagrado a todos os inimigos.",
				Requires: []string{"cler_radiance"},
				Effects: []engine_Effect{
					{TypeName: "aoe", Element: "holy", BasePower: 38},
				},
			},
			{
				ID: "cler_blessing", Class: "cleric", Branch: "luz_divina",
				Name: "Bênção", Emoji: "🙏", Tier: 2, PointCost: 2, RequiredLevel: 18,
				MPCost: 30, EnergyCost: 6,
				Description: "Bênção divina: +20% ATK, +20% DEF, +20% MAG por 4 turnos. Aplica Regen.",
				Requires: []string{"cler_light_burst"},
				Effects: []engine_Effect{
					{TypeName: "apply_status", StatusKind: "regen", StatusTurns: 4},
					{TypeName: "apply_status", StatusKind: "protect", StatusTurns: 2},
				},
			},
			{
				ID: "cler_purge", Class: "cleric", Branch: "luz_divina",
				Name: "Purgar", Emoji: "🔥", Tier: 2, PointCost: 2, RequiredLevel: 22,
				MPCost: 35, EnergyCost: 8,
				Description: "Purga com fogo sagrado: 4d10+SAB de dano de fogo sagrado. Instakill mortos-vivos abaixo de 30% HP.",
				Requires: []string{"cler_blessing"},
				Effects: []engine_Effect{
					{TypeName: "damage", Element: "holy", BasePower: 55},
					{TypeName: "damage", Element: "fire", BasePower: 25},
				},
			},
			{
				ID: "cler_divine_fire", Class: "cleric", Branch: "luz_divina",
				Name: "Fogo Divino", Emoji: "🔥", Tier: 3, PointCost: 2, RequiredLevel: 30,
				MPCost: 50, EnergyCost: 10,
				Description: "Colunas de fogo divino: 5d10+SAB de dano em AoE + queimadura por 4 turnos (8/turno).",
				Requires: []string{"cler_purge"},
				Effects: []engine_Effect{
					{TypeName: "aoe", Element: "fire", BasePower: 65},
					{TypeName: "apply_status", StatusKind: "burn", StatusTurns: 4, StatusDmgPT: 8},
				},
			},
			{
				ID: "cler_holy_storm", Class: "cleric", Branch: "luz_divina",
				Name: "Tempestade Sagrada", Emoji: "⛈️", Tier: 3, PointCost: 2, RequiredLevel: 38,
				MPCost: 60, EnergyCost: 12,
				Description: "Tempestade de luz divina por 3 turnos: 3d8+SAB de dano sagrado/turno em AoE + Cegueira.",
				Requires: []string{"cler_divine_fire"},
				Effects: []engine_Effect{
					{TypeName: "aoe", Element: "holy", BasePower: 40},
					{TypeName: "apply_status", StatusKind: "blind", StatusTurns: 3},
				},
			},
			{
				ID: "cler_crusade", Class: "cleric", Branch: "luz_divina",
				Name: "Cruzada", Emoji: "⚔️", Tier: 4, PointCost: 3, RequiredLevel: 50,
				MPCost: 75, EnergyCost: 16,
				Description: "Bênção de cruzada em massa: todos os aliados ganham +30% ATK, +20% DEF e Berserk por 4 turnos.",
				Requires: []string{"cler_holy_storm"},
				Effects: []engine_Effect{
					{TypeName: "apply_status", StatusKind: "berserk", StatusTurns: 4},
					{TypeName: "apply_status", StatusKind: "protect", StatusTurns: 3},
				},
			},
			{
				ID: "cler_avatar_of_light", Class: "cleric", Branch: "luz_divina",
				Name: "Avatar da Luz", Emoji: "🌟", Tier: 5, PointCost: 5, RequiredLevel: 80,
				MPCost: 130, EnergyCost: 35, IsUltimate: true,
				Description: "Transcende para Avatar da Luz: 10d20+SAB de dano sagrado em AoE + cura 100 HP de todos + todos os buffs por 5 turnos.",
				Requires: []string{"cler_crusade"},
				Effects: []engine_Effect{
					{TypeName: "aoe", Element: "holy", BasePower: 180},
					{TypeName: "heal", BasePower: 100},
					{TypeName: "apply_status", StatusKind: "protect", StatusTurns: 5},
					{TypeName: "apply_status", StatusKind: "regen", StatusTurns: 5},
				},
			},
		},
	})

	// Add branch 2 — Proteção Divina (9 nodes)
	t.Branches = append(t.Branches, Branch{
		ID: "protecao_divina", Name: "Proteção Divina", Emoji: "🛡️",
		Nodes: []SkillNode{
			{
				ID: "cler_shield_prayer", Class: "cleric", Branch: "protecao_divina",
				Name: "Oração Escudo", Emoji: "🙏", Tier: 1, PointCost: 1, RequiredLevel: 1,
				MPCost: 12, EnergyCost: 3,
				Description: "Oração protetora: aplica Escudo por 2 turnos e Regen por 2 turnos.",
				Effects: []engine_Effect{
					{TypeName: "apply_status", StatusKind: "shield", StatusTurns: 2},
					{TypeName: "apply_status", StatusKind: "regen", StatusTurns: 2},
				},
			},
			{
				ID: "cler_refuge", Class: "cleric", Branch: "protecao_divina",
				Name: "Refúgio", Emoji: "🏛️", Tier: 1, PointCost: 1, RequiredLevel: 6,
				MPCost: 16, EnergyCost: 3,
				Description: "Refugia-se na graça divina: aplica Proteção por 2 turnos e cura 15 HP.",
				Requires: []string{"cler_shield_prayer"},
				Effects: []engine_Effect{
					{TypeName: "apply_status", StatusKind: "protect", StatusTurns: 2},
					{TypeName: "heal", BasePower: 15},
				},
			},
			{
				ID: "cler_sanctuary_d", Class: "cleric", Branch: "protecao_divina",
				Name: "Santuário Divino", Emoji: "☀️", Tier: 2, PointCost: 1, RequiredLevel: 12,
				MPCost: 28, EnergyCost: 5,
				Description: "Cria santuário de luz: reduz todo dano recebido em 40% por 3 turnos.",
				Requires: []string{"cler_refuge"},
				Effects: []engine_Effect{
					{TypeName: "apply_status", StatusKind: "protect", StatusTurns: 3},
					{TypeName: "apply_status", StatusKind: "shield", StatusTurns: 2},
				},
			},
			{
				ID: "cler_divine_armor", Class: "cleric", Branch: "protecao_divina",
				Name: "Armadura Divina", Emoji: "⚜️", Tier: 2, PointCost: 2, RequiredLevel: 18,
				IsPassive: true,
				Description: "Passivo: +25 DEF mágica, +15 DEF física e -10% de todo dano recebido.",
				Requires: []string{"cler_sanctuary_d"},
			},
			{
				ID: "cler_blessed_ground", Class: "cleric", Branch: "protecao_divina",
				Name: "Solo Abençoado", Emoji: "🌟", Tier: 3, PointCost: 2, RequiredLevel: 25,
				MPCost: 40, EnergyCost: 9,
				Description: "Consagra o solo: cria campo de Regen (+8 HP/turno) por 4 turnos para aliados na área.",
				Requires: []string{"cler_divine_armor"},
				Effects: []engine_Effect{
					{TypeName: "apply_status", StatusKind: "regen", StatusTurns: 4},
				},
			},
			{
				ID: "cler_ward", Class: "cleric", Branch: "protecao_divina",
				Name: "Proteção Mágica", Emoji: "💎", Tier: 3, PointCost: 2, RequiredLevel: 32,
				MPCost: 45, EnergyCost: 10,
				Description: "Imunidade a status por 3 turnos: nem veneno, stun, cegueira, silêncio ou maldição afetam.",
				Requires: []string{"cler_blessed_ground"},
				Effects: []engine_Effect{
					{TypeName: "remove_status", StatusKind: "poison"},
					{TypeName: "remove_status", StatusKind: "curse"},
					{TypeName: "remove_status", StatusKind: "stun"},
					{TypeName: "apply_status", StatusKind: "protect", StatusTurns: 3},
				},
			},
			{
				ID: "cler_absolution", Class: "cleric", Branch: "protecao_divina",
				Name: "Absolvição", Emoji: "✨", Tier: 3, PointCost: 2, RequiredLevel: 40,
				MPCost: 55, EnergyCost: 12,
				Description: "Remove TODOS os status negativos, cura 3d8+SAB e aplica Escudo+Regen por 4 turnos.",
				Requires: []string{"cler_ward"},
				Effects: []engine_Effect{
					{TypeName: "remove_status", StatusKind: "poison"},
					{TypeName: "heal", BasePower: 35},
					{TypeName: "apply_status", StatusKind: "shield", StatusTurns: 4},
					{TypeName: "apply_status", StatusKind: "regen", StatusTurns: 4},
				},
			},
			{
				ID: "cler_celestial_guard", Class: "cleric", Branch: "protecao_divina",
				Name: "Guarda Celestial", Emoji: "👼", Tier: 4, PointCost: 3, RequiredLevel: 55,
				MPCost: 80, EnergyCost: 18,
				Description: "Proteção celestial por 5 turnos: Escudo+Proteção+Regen e qualquer golpe fatal é anulado (1x).",
				Requires: []string{"cler_absolution"},
				Effects: []engine_Effect{
					{TypeName: "apply_status", StatusKind: "shield", StatusTurns: 5},
					{TypeName: "apply_status", StatusKind: "protect", StatusTurns: 5},
					{TypeName: "apply_status", StatusKind: "regen", StatusTurns: 5},
				},
			},
			{
				ID: "cler_godly_shield", Class: "cleric", Branch: "protecao_divina",
				Name: "Escudo dos Deuses", Emoji: "🌟", Tier: 5, PointCost: 5, RequiredLevel: 80,
				MPCost: 120, EnergyCost: 30, IsUltimate: true,
				Description: "O escudo definitivo: invulnerabilidade absoluta por 3 turnos. Ao fim, cura 80 HP e lança Nova Sagrada em AoE.",
				Requires: []string{"cler_celestial_guard"},
				Effects: []engine_Effect{
					{TypeName: "apply_status", StatusKind: "protect", StatusTurns: 3},
					{TypeName: "apply_status", StatusKind: "shield", StatusTurns: 3},
					{TypeName: "heal", BasePower: 80},
					{TypeName: "aoe", Element: "holy", BasePower: 70},
				},
			},
		},
	})
}

// ─── Bard extension ───────────────────────────────────────────────────────────
// Adds 23 new skills: 5 to Música, 9 new Conhecimento Bárdico, 9 new Ilusionismo.

func extendBard() {
	t, ok := Trees["bard"]
	if !ok {
		return
	}
	// Extend branch 0 (Música) — 5 new nodes
	t.Branches[0].Nodes = append(t.Branches[0].Nodes,
		SkillNode{
			ID: "bard_ballad", Class: "bard", Branch: "musica",
			Name: "Balada", Emoji: "🎶", Tier: 2, PointCost: 1, RequiredLevel: 8,
			MPCost: 20, EnergyCost: 5,
			Description: "Balada restauradora: aplica Regen por 4 turnos (6 HP/turno) e cura 15 HP imediatamente.",
			Requires: []string{"bard_inspire"},
			Effects: []engine_Effect{
				{TypeName: "heal", BasePower: 15},
				{TypeName: "apply_status", StatusKind: "regen", StatusTurns: 4},
			},
		},
		SkillNode{
			ID: "bard_battle_hymn", Class: "bard", Branch: "musica",
			Name: "Hino de Batalha", Emoji: "⚔️", Tier: 2, PointCost: 2, RequiredLevel: 15,
			MPCost: 30, EnergyCost: 7,
			Description: "Hino guerreiro: aplica Berserk+Haste por 3 turnos em si mesmo.",
			Requires: []string{"bard_ballad"},
			Effects: []engine_Effect{
				{TypeName: "apply_status", StatusKind: "berserk", StatusTurns: 3},
				{TypeName: "apply_status", StatusKind: "haste", StatusTurns: 3},
			},
		},
		SkillNode{
			ID: "bard_requiem", Class: "bard", Branch: "musica",
			Name: "Réquiem", Emoji: "💀", Tier: 3, PointCost: 2, RequiredLevel: 28,
			MPCost: 45, EnergyCost: 10,
			Description: "Música da morte em AoE: Stun + Maldição + -30% ATK por 3 turnos em todos os inimigos.",
			Requires: []string{"bard_battle_hymn"},
			Effects: []engine_Effect{
				{TypeName: "aoe", Element: "dark", BasePower: 20},
				{TypeName: "apply_status", StatusKind: "curse", StatusTurns: 3},
				{TypeName: "stat_debuff", StatName: "attack", StatDelta: -8},
			},
		},
		SkillNode{
			ID: "bard_epic_saga", Class: "bard", Branch: "musica",
			Name: "Saga Épica", Emoji: "📖", Tier: 4, PointCost: 3, RequiredLevel: 50,
			MPCost: 70, EnergyCost: 16,
			Description: "Narra uma saga épica: todos os aliados ganham Berserk+Haste+Regen por 4 turnos.",
			Requires: []string{"bard_requiem"},
			Effects: []engine_Effect{
				{TypeName: "apply_status", StatusKind: "berserk", StatusTurns: 4},
				{TypeName: "apply_status", StatusKind: "haste", StatusTurns: 4},
				{TypeName: "apply_status", StatusKind: "regen", StatusTurns: 4},
			},
		},
		SkillNode{
			ID: "bard_song_of_legends", Class: "bard", Branch: "musica",
			Name: "Canção das Lendas", Emoji: "🌟", Tier: 5, PointCost: 5, RequiredLevel: 80,
			MPCost: 120, EnergyCost: 35, IsUltimate: true,
			Description: "A canção mais poderosa já entoada: todos os buffs por 6 turnos e 8d10+CHA de dano mágico em AoE.",
			Requires: []string{"bard_epic_saga"},
			Effects: []engine_Effect{
				{TypeName: "apply_status", StatusKind: "berserk", StatusTurns: 6},
				{TypeName: "apply_status", StatusKind: "haste", StatusTurns: 6},
				{TypeName: "apply_status", StatusKind: "regen", StatusTurns: 6},
				{TypeName: "apply_status", StatusKind: "protect", StatusTurns: 6},
				{TypeName: "aoe", Element: "magic", BasePower: 100},
			},
		},
	)

	// Add branch 1 — Conhecimento Bárdico (9 nodes)
	t.Branches = append(t.Branches, Branch{
		ID: "conhecimento_bardico", Name: "Conhecimento Bárdico", Emoji: "📚",
		Nodes: []SkillNode{
			{
				ID: "bard_knowledge_blast", Class: "bard", Branch: "conhecimento_bardico",
				Name: "Explosão de Conhecimento", Emoji: "📚", Tier: 1, PointCost: 1, RequiredLevel: 1,
				MPCost: 12, EnergyCost: 3,
				Description: "Canaliza conhecimento como energia: 2d6+INT de dano mágico arcano.",
				Effects: []engine_Effect{
					{TypeName: "damage", Element: "magic", BasePower: 22},
				},
			},
			{
				ID: "bard_mock", Class: "bard", Branch: "conhecimento_bardico",
				Name: "Zombar", Emoji: "😂", Tier: 1, PointCost: 1, RequiredLevel: 5,
				MPCost: 14, EnergyCost: 3,
				Description: "Zomba do inimigo humilhando-o: -25% ATK e -20% DEF por 3 turnos.",
				Requires: []string{"bard_knowledge_blast"},
				Effects: []engine_Effect{
					{TypeName: "stat_debuff", StatName: "attack", StatDelta: -7},
					{TypeName: "stat_debuff", StatName: "defense", StatDelta: -6},
				},
			},
			{
				ID: "bard_tale_of_weakness", Class: "bard", Branch: "conhecimento_bardico",
				Name: "Conto da Fraqueza", Emoji: "📜", Tier: 2, PointCost: 1, RequiredLevel: 12,
				MPCost: 22, EnergyCost: 5,
				Description: "Narra a fraqueza do inimigo: -40% DEF por 5 turnos e revela todas as fraquezas elementais.",
				Requires: []string{"bard_mock"},
				Effects: []engine_Effect{
					{TypeName: "stat_debuff", StatName: "defense", StatDelta: -14},
				},
			},
			{
				ID: "bard_dissonance", Class: "bard", Branch: "conhecimento_bardico",
				Name: "Dissonância", Emoji: "🎵", Tier: 2, PointCost: 2, RequiredLevel: 18,
				MPCost: 30, EnergyCost: 7,
				Description: "Nota dissonante em AoE: causa confusão (-30% precisão) e 2d6+INT de dano mágico a todos.",
				Requires: []string{"bard_tale_of_weakness"},
				Effects: []engine_Effect{
					{TypeName: "aoe", Element: "magic", BasePower: 22},
					{TypeName: "apply_status", StatusKind: "blind", StatusTurns: 2},
				},
			},
			{
				ID: "bard_dark_melody", Class: "bard", Branch: "conhecimento_bardico",
				Name: "Melodia Sombria", Emoji: "🎶", Tier: 2, PointCost: 2, RequiredLevel: 22,
				MPCost: 35, EnergyCost: 8,
				Description: "Melodia que amaldiçoa: aplica Maldição + Silêncio por 3 turnos.",
				Requires: []string{"bard_dissonance"},
				Effects: []engine_Effect{
					{TypeName: "apply_status", StatusKind: "curse", StatusTurns: 3},
					{TypeName: "apply_status", StatusKind: "silence", StatusTurns: 2},
				},
			},
			{
				ID: "bard_shatter_morale", Class: "bard", Branch: "conhecimento_bardico",
				Name: "Destruir Moral", Emoji: "💔", Tier: 3, PointCost: 2, RequiredLevel: 28,
				MPCost: 45, EnergyCost: 10,
				Description: "Destrói a moral: -50% ATK, -40% DEF e Maldição por 4 turnos. Em AoE.",
				Requires: []string{"bard_dark_melody"},
				Effects: []engine_Effect{
					{TypeName: "aoe", Element: "dark", BasePower: 15},
					{TypeName: "stat_debuff", StatName: "attack", StatDelta: -14},
					{TypeName: "stat_debuff", StatName: "defense", StatDelta: -12},
					{TypeName: "apply_status", StatusKind: "curse", StatusTurns: 4},
				},
			},
			{
				ID: "bard_echo_strike", Class: "bard", Branch: "conhecimento_bardico",
				Name: "Golpe Eco", Emoji: "🔊", Tier: 3, PointCost: 2, RequiredLevel: 35,
				MPCost: 50, EnergyCost: 11,
				Description: "Ataque mágico que ecoa: 2x (3d8+INT) de dano mágico em sequência rápida.",
				Requires: []string{"bard_shatter_morale"},
				Effects: []engine_Effect{
					{TypeName: "damage", Element: "magic", BasePower: 40},
					{TypeName: "damage", Element: "magic", BasePower: 40},
				},
			},
			{
				ID: "bard_discordance", Class: "bard", Branch: "conhecimento_bardico",
				Name: "Discordância", Emoji: "🎼", Tier: 4, PointCost: 3, RequiredLevel: 48,
				MPCost: 70, EnergyCost: 15,
				Description: "Sinfonia caótica em AoE: Stun+Silêncio+Maldição+Cegueira por 2 turnos em todos os inimigos.",
				Requires: []string{"bard_echo_strike"},
				Effects: []engine_Effect{
					{TypeName: "aoe", Element: "magic", BasePower: 35},
					{TypeName: "apply_status", StatusKind: "stun", StatusTurns: 2},
					{TypeName: "apply_status", StatusKind: "silence", StatusTurns: 2},
					{TypeName: "apply_status", StatusKind: "curse", StatusTurns: 2},
					{TypeName: "apply_status", StatusKind: "blind", StatusTurns: 2},
				},
			},
			{
				ID: "bard_finale", Class: "bard", Branch: "conhecimento_bardico",
				Name: "Final da Ópera", Emoji: "🎭", Tier: 5, PointCost: 5, RequiredLevel: 75,
				MPCost: 120, EnergyCost: 32, IsUltimate: true,
				Description: "O grande final: 10d12+INT de dano mágico em AoE + todos os debuffs por 5 turnos. Silencia todos.",
				Requires: []string{"bard_discordance"},
				Effects: []engine_Effect{
					{TypeName: "aoe", Element: "magic", BasePower: 140},
					{TypeName: "apply_status", StatusKind: "stun", StatusTurns: 3},
					{TypeName: "apply_status", StatusKind: "silence", StatusTurns: 5},
					{TypeName: "apply_status", StatusKind: "curse", StatusTurns: 5},
				},
			},
		},
	})

	// Add branch 2 — Ilusionismo (9 nodes)
	t.Branches = append(t.Branches, Branch{
		ID: "ilusionismo", Name: "Ilusionismo", Emoji: "🌀",
		Nodes: []SkillNode{
			{
				ID: "bard_mirror_image", Class: "bard", Branch: "ilusionismo",
				Name: "Imagem Espelho", Emoji: "🪞", Tier: 1, PointCost: 1, RequiredLevel: 1,
				MPCost: 14, EnergyCost: 3,
				Description: "Cria imagem ilusória: aplica Haste e +25% esquiva por 2 turnos.",
				Effects: []engine_Effect{
					{TypeName: "apply_status", StatusKind: "haste", StatusTurns: 2},
				},
			},
			{
				ID: "bard_phantasmal_blade", Class: "bard", Branch: "ilusionismo",
				Name: "Lâmina Fantasmal", Emoji: "👻", Tier: 1, PointCost: 1, RequiredLevel: 6,
				MPCost: 18, EnergyCost: 4,
				Description: "Ataca com lâmina ilusória que causa dano mágico: 2d8+INT. Ignora armadura física.",
				Requires: []string{"bard_mirror_image"},
				Effects: []engine_Effect{
					{TypeName: "damage", Element: "magic", BasePower: 26},
				},
			},
			{
				ID: "bard_fear", Class: "bard", Branch: "ilusionismo",
				Name: "Medo", Emoji: "😱", Tier: 2, PointCost: 1, RequiredLevel: 12,
				MPCost: 22, EnergyCost: 5,
				Description: "Induz medo no alvo: Cegueira por 2 turnos e -35% ATK por 3 turnos.",
				Requires: []string{"bard_phantasmal_blade"},
				Effects: []engine_Effect{
					{TypeName: "apply_status", StatusKind: "blind", StatusTurns: 2},
					{TypeName: "stat_debuff", StatName: "attack", StatDelta: -10},
				},
			},
			{
				ID: "bard_mass_confusion", Class: "bard", Branch: "ilusionismo",
				Name: "Confusão em Massa", Emoji: "😵", Tier: 2, PointCost: 2, RequiredLevel: 18,
				MPCost: 32, EnergyCost: 7,
				Description: "Ilusão coletiva confunde todos os inimigos: Stun por 1 turno + Cegueira por 2 turnos em AoE.",
				Requires: []string{"bard_fear"},
				Effects: []engine_Effect{
					{TypeName: "aoe", Element: "magic", BasePower: 10},
					{TypeName: "apply_status", StatusKind: "stun", StatusTurns: 1},
					{TypeName: "apply_status", StatusKind: "blind", StatusTurns: 2},
				},
			},
			{
				ID: "bard_glamour", Class: "bard", Branch: "ilusionismo",
				Name: "Glamour", Emoji: "✨", Tier: 2, PointCost: 2, RequiredLevel: 22,
				MPCost: 35, EnergyCost: 8,
				Description: "Sedução encantadora: força o inimigo a perder turno (Stun) por 2 turnos. Ele fica paralisado pelo encanto.",
				Requires: []string{"bard_mass_confusion"},
				Effects: []engine_Effect{
					{TypeName: "apply_status", StatusKind: "stun", StatusTurns: 2},
				},
			},
			{
				ID: "bard_phantasm", Class: "bard", Branch: "ilusionismo",
				Name: "Fantasma", Emoji: "👻", Tier: 3, PointCost: 2, RequiredLevel: 30,
				MPCost: 48, EnergyCost: 10,
				Description: "Cria fantasma que ataca 3 vezes com 2d8+INT de dano mágico cada.",
				Requires: []string{"bard_glamour"},
				Effects: []engine_Effect{
					{TypeName: "damage", Element: "magic", BasePower: 32},
					{TypeName: "damage", Element: "magic", BasePower: 32},
					{TypeName: "damage", Element: "magic", BasePower: 32},
				},
			},
			{
				ID: "bard_dream_walk", Class: "bard", Branch: "ilusionismo",
				Name: "Caminhar no Sonho", Emoji: "💤", Tier: 3, PointCost: 2, RequiredLevel: 38,
				MPCost: 55, EnergyCost: 12,
				Description: "Entra no reino dos sonhos: evita próximo ataque e responde com 4d10+INT de dano mental.",
				Requires: []string{"bard_phantasm"},
				Effects: []engine_Effect{
					{TypeName: "apply_status", StatusKind: "haste", StatusTurns: 1},
					{TypeName: "damage", Element: "magic", BasePower: 55},
					{TypeName: "apply_status", StatusKind: "stun", StatusTurns: 1},
				},
			},
			{
				ID: "bard_nightmare", Class: "bard", Branch: "ilusionismo",
				Name: "Pesadelo", Emoji: "😨", Tier: 4, PointCost: 3, RequiredLevel: 50,
				MPCost: 75, EnergyCost: 16,
				Description: "Pesadelo induzido: 6d10+INT de dano mágico + todos os status negativos por 2 turnos.",
				Requires: []string{"bard_dream_walk"},
				Effects: []engine_Effect{
					{TypeName: "damage", Element: "dark", BasePower: 80},
					{TypeName: "apply_status", StatusKind: "stun", StatusTurns: 2},
					{TypeName: "apply_status", StatusKind: "blind", StatusTurns: 2},
					{TypeName: "apply_status", StatusKind: "curse", StatusTurns: 2},
					{TypeName: "apply_status", StatusKind: "silence", StatusTurns: 1},
				},
			},
			{
				ID: "bard_grand_illusion", Class: "bard", Branch: "ilusionismo",
				Name: "Grande Ilusão", Emoji: "🌌", Tier: 5, PointCost: 5, RequiredLevel: 80,
				MPCost: 130, EnergyCost: 35, IsUltimate: true,
				Description: "A ilusão suprema: invulnerabilidade por 2 turnos + 10d12+INT de dano em AoE + todos os debuffs por 4 turnos.",
				Requires: []string{"bard_nightmare"},
				Effects: []engine_Effect{
					{TypeName: "apply_status", StatusKind: "protect", StatusTurns: 2},
					{TypeName: "apply_status", StatusKind: "haste", StatusTurns: 2},
					{TypeName: "aoe", Element: "magic", BasePower: 150},
					{TypeName: "apply_status", StatusKind: "stun", StatusTurns: 4},
					{TypeName: "apply_status", StatusKind: "curse", StatusTurns: 4},
				},
			},
		},
	})
}

// ─── Passive registry for all new skills ─────────────────────────────────────

func registerAllPassives() {
	// ── Warrior passives ──
	engine.PassiveRegistry["warr_parry"] = engine.PassiveBonus{DefenseBonus: 4, CritChance: 5}
	engine.PassiveRegistry["warr_weapon_focus"] = engine.PassiveBonus{AttackBonus: 5, CritChance: 8}
	engine.PassiveRegistry["warr_iron_will"] = engine.PassiveBonus{HPBonus: 15}
	engine.PassiveRegistry["warr_grit"] = engine.PassiveBonus{DefenseBonus: 6, HPBonus: 20}
	engine.PassiveRegistry["warr_endurance"] = engine.PassiveBonus{HPBonus: 40}
	engine.PassiveRegistry["warr_indomable"] = engine.PassiveBonus{AttackBonus: 3, HPBonus: 10}

	// ── Mage passives ──
	engine.PassiveRegistry["mage_mana_surge"] = engine.PassiveBonus{MPBonus: 30}
	engine.PassiveRegistry["mage_spell_power"] = engine.PassiveBonus{MPBonus: 10, CritChance: 5}
	engine.PassiveRegistry["mage_arcane_mastery"] = engine.PassiveBonus{MPBonus: 20, CritChance: 10}

	// ── Rogue passives ──
	engine.PassiveRegistry["rog_evade"] = engine.PassiveBonus{SpeedBonus: 3, CritChance: 5}
	engine.PassiveRegistry["rog_venomous_strikes"] = engine.PassiveBonus{AttackBonus: 3, CritChance: 8}
	engine.PassiveRegistry["rog_master_thief"] = engine.PassiveBonus{CritChance: 10, AttackBonus: 5}

	// ── Archer passives ──
	engine.PassiveRegistry["arch_eagle_eye"] = engine.PassiveBonus{CritChance: 15}
	engine.PassiveRegistry["arch_sniper"] = engine.PassiveBonus{AttackBonus: 5, CritChance: 10}
	engine.PassiveRegistry["arch_camouflage"] = engine.PassiveBonus{SpeedBonus: 2, CritChance: 5}
	engine.PassiveRegistry["arch_quick_reflexes"] = engine.PassiveBonus{SpeedBonus: 5, DefenseBonus: 3}
	engine.PassiveRegistry["arch_wilderness_lore"] = engine.PassiveBonus{DefenseBonus: 5, HPBonus: 20}
	engine.PassiveRegistry["arch_perfect_hunt"] = engine.PassiveBonus{AttackBonus: 8, CritChance: 15}

	// ── Barbarian extended passives ──
	engine.PassiveRegistry["barb_blood_frenzy"] = engine.PassiveBonus{AttackBonus: 8, HPBonus: 20}
	engine.PassiveRegistry["barb_battle_hardened"] = engine.PassiveBonus{DefenseBonus: 10, HPBonus: 15}
	engine.PassiveRegistry["barb_immortal_flesh"] = engine.PassiveBonus{HPBonus: 80}
	engine.PassiveRegistry["barb_pain_defiance"] = engine.PassiveBonus{DefenseBonus: 8, HPBonus: 25}

	// ── Paladin extended passives ──
	engine.PassiveRegistry["pal_aura_protection"] = engine.PassiveBonus{DefenseBonus: 8, HPBonus: 15}
	engine.PassiveRegistry["pal_shield_of_faith"] = engine.PassiveBonus{DefenseBonus: 12, HPBonus: 30}
	engine.PassiveRegistry["pal_blessed_armor"] = engine.PassiveBonus{DefenseBonus: 20, HPBonus: 20}

	// ── Cleric extended passives ──
	engine.PassiveRegistry["cler_divine_armor"] = engine.PassiveBonus{DefenseBonus: 15, MPBonus: 20}

	// ── Bard extended — no new passives (mostly active skills) ──
}
