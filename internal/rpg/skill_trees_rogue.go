package rpg

func init() {
	Trees["rogue"] = rogueTree()
}

// rogueTree returns the full skill tree for the rogue class.
// 26 skills across 3 branches: Sombras, Venenos, Esperteza.
func rogueTree() *SkillTree {
	return &SkillTree{
		ClassID: "rogue",
		Branches: []Branch{
			{
				ID: "sombras", Name: "Sombras", Emoji: "🌑",
				Nodes: []SkillNode{
					// T1 — Nível 1
					{
						ID: "rog_backstab", Class: "rogue", Branch: "sombras",
						Name: "Ataque Pelas Costas", Emoji: "🗡️", Tier: 1, PointCost: 1, RequiredLevel: 1,
						MPCost: 8, EnergyCost: 3,
						Description: "Ataque sorrateiro pelas costas: 2d8+DEX de dano com +30% de chance de crítico.",
						Effects: []engine_Effect{
							{TypeName: "damage", Element: "physical", BasePower: 28},
						},
					},
					{
						ID: "rog_shadow_step", Class: "rogue", Branch: "sombras",
						Name: "Passo nas Sombras", Emoji: "👤", Tier: 1, PointCost: 1, RequiredLevel: 4,
						MPCost: 12, EnergyCost: 4,
						Description: "Move-se pelas sombras, evitando o próximo ataque e reposicionando para um golpe de 1d10+DEX.",
						Requires: []string{"rog_backstab"},
						Effects: []engine_Effect{
							{TypeName: "apply_status", StatusKind: "haste", StatusTurns: 1},
							{TypeName: "damage", Element: "physical", BasePower: 22},
						},
					},
					{
						ID: "rog_mark", Class: "rogue", Branch: "sombras",
						Name: "Marcar Alvo", Emoji: "🎯", Tier: 2, PointCost: 1, RequiredLevel: 10,
						MPCost: 15, EnergyCost: 3,
						Description: "Marca o alvo como presa: todos os ataques subsequentes causam +25% de dano por 4 turnos.",
						Requires: []string{"rog_shadow_step"},
						Effects: []engine_Effect{
							{TypeName: "stat_debuff", StatName: "defense", StatDelta: -6},
						},
					},
					{
						ID: "rog_shadow_strike", Class: "rogue", Branch: "sombras",
						Name: "Golpe das Sombras", Emoji: "🌑", Tier: 2, PointCost: 2, RequiredLevel: 14,
						MPCost: 25, EnergyCost: 6,
						Description: "Ataque invisível da escuridão: 3d10+DEX de dano. Sempre crítico se o alvo estiver Cego.",
						Requires: []string{"rog_mark"},
						Effects: []engine_Effect{
							{TypeName: "damage", Element: "physical", BasePower: 45},
						},
					},
					{
						ID: "rog_evade", Class: "rogue", Branch: "sombras",
						Name: "Evasão", Emoji: "💨", Tier: 2, PointCost: 2, RequiredLevel: 18,
						IsPassive: true,
						Description: "Passivo: +20% de esquiva, +8 DEX efetivo e -15% de chance de ser acertado por ataques.",
						Requires: []string{"rog_shadow_strike"},
					},
					{
						ID: "rog_death_mark", Class: "rogue", Branch: "sombras",
						Name: "Marca da Morte", Emoji: "💀", Tier: 3, PointCost: 2, RequiredLevel: 25,
						MPCost: 35, EnergyCost: 8,
						Description: "Condena o alvo: amplifica todo dano recebido em 40% por 5 turnos e aplica Amaldiçoado.",
						Requires: []string{"rog_evade"},
						Effects: []engine_Effect{
							{TypeName: "apply_status", StatusKind: "curse", StatusTurns: 5},
							{TypeName: "stat_debuff", StatName: "defense", StatDelta: -12},
						},
					},
					{
						ID: "rog_shadow_clone", Class: "rogue", Branch: "sombras",
						Name: "Clone de Sombra", Emoji: "👥", Tier: 3, PointCost: 2, RequiredLevel: 30,
						MPCost: 45, EnergyCost: 10,
						Description: "Cria um clone de sombra que ataca junto: dois golpes de 2d8+DEX cada.",
						Requires: []string{"rog_death_mark"},
						Effects: []engine_Effect{
							{TypeName: "damage", Element: "physical", BasePower: 35},
							{TypeName: "damage", Element: "physical", BasePower: 35},
						},
					},
					{
						ID: "rog_assassinate", Class: "rogue", Branch: "sombras",
						Name: "Assassinar", Emoji: "⚰️", Tier: 4, PointCost: 3, RequiredLevel: 40,
						MPCost: 70, EnergyCost: 15, IsUltimate: true,
						Description: "Golpe de assassino perfeito: 8d10+DEX de dano. Se o alvo tiver menos de 35% HP, causa dano triplo.",
						Requires: []string{"rog_shadow_clone"},
						Effects: []engine_Effect{
							{TypeName: "damage", Element: "physical", BasePower: 100},
						},
					},
				},
			},
			{
				ID: "venenos", Name: "Venenos", Emoji: "☠️",
				Nodes: []SkillNode{
					// T1 — Nível 1
					{
						ID: "rog_poison_blade", Class: "rogue", Branch: "venenos",
						Name: "Lâmina Envenenada", Emoji: "☠️", Tier: 1, PointCost: 1, RequiredLevel: 1,
						MPCost: 10, EnergyCost: 3,
						Description: "Envenena a lâmina: próximo ataque causa 1d6 de dano e aplica veneno por 4 turnos (3 dano/turno).",
						Effects: []engine_Effect{
							{TypeName: "damage", Element: "poison", BasePower: 16},
							{TypeName: "apply_status", StatusKind: "poison", StatusTurns: 4, StatusDmgPT: 3},
						},
					},
					{
						ID: "rog_crippling_poison", Class: "rogue", Branch: "venenos",
						Name: "Veneno Paralisante", Emoji: "🧪", Tier: 1, PointCost: 1, RequiredLevel: 5,
						MPCost: 14, EnergyCost: 3,
						Description: "Aplica veneno que paralisa: veneno 3/turno por 3 turnos + Lentidão por 2 turnos (-30% velocidade).",
						Requires: []string{"rog_poison_blade"},
						Effects: []engine_Effect{
							{TypeName: "apply_status", StatusKind: "poison", StatusTurns: 3, StatusDmgPT: 3},
							{TypeName: "stat_debuff", StatName: "speed", StatDelta: -5},
						},
					},
					{
						ID: "rog_toxin", Class: "rogue", Branch: "venenos",
						Name: "Toxina", Emoji: "💀", Tier: 2, PointCost: 1, RequiredLevel: 10,
						MPCost: 22, EnergyCost: 5,
						Description: "Injecta toxina virulenta: 2d6 de dano imediato + veneno pesado por 5 turnos (7 dano/turno).",
						Requires: []string{"rog_crippling_poison"},
						Effects: []engine_Effect{
							{TypeName: "damage", Element: "poison", BasePower: 22},
							{TypeName: "apply_status", StatusKind: "poison", StatusTurns: 5, StatusDmgPT: 7},
						},
					},
					{
						ID: "rog_poison_cloud", Class: "rogue", Branch: "venenos",
						Name: "Nuvem de Veneno", Emoji: "🌫️", Tier: 2, PointCost: 2, RequiredLevel: 15,
						MPCost: 30, EnergyCost: 7,
						Description: "Cria nuvem venenosa em AoE que aplica veneno por 4 turnos (5 dano/turno) a todos os inimigos.",
						Requires: []string{"rog_toxin"},
						Effects: []engine_Effect{
							{TypeName: "aoe", Element: "poison", BasePower: 15},
							{TypeName: "apply_status", StatusKind: "poison", StatusTurns: 4, StatusDmgPT: 5},
						},
					},
					{
						ID: "rog_lethal_venom", Class: "rogue", Branch: "venenos",
						Name: "Veneno Letal", Emoji: "⚗️", Tier: 2, PointCost: 2, RequiredLevel: 20,
						MPCost: 35, EnergyCost: 8,
						Description: "Veneno mortal destilado: aplica veneno que escalona (5→7→10→14 dano/turno) por 4 turnos.",
						Requires: []string{"rog_poison_cloud"},
						Effects: []engine_Effect{
							{TypeName: "damage", Element: "poison", BasePower: 30},
							{TypeName: "apply_status", StatusKind: "poison", StatusTurns: 4, StatusDmgPT: 10},
						},
					},
					{
						ID: "rog_plague_blade", Class: "rogue", Branch: "venenos",
						Name: "Lâmina da Praga", Emoji: "🦠", Tier: 3, PointCost: 2, RequiredLevel: 28,
						MPCost: 45, EnergyCost: 10,
						Description: "Lâmina infectada com praga: 3d8+DEX de dano, veneno por 6 turnos (8/turno) e Cegueira por 2 turnos.",
						Requires: []string{"rog_lethal_venom"},
						Effects: []engine_Effect{
							{TypeName: "damage", Element: "poison", BasePower: 40},
							{TypeName: "apply_status", StatusKind: "poison", StatusTurns: 6, StatusDmgPT: 8},
							{TypeName: "apply_status", StatusKind: "blind", StatusTurns: 2},
						},
					},
					{
						ID: "rog_venomous_strikes", Class: "rogue", Branch: "venenos",
						Name: "Golpes Venenosos", Emoji: "🐍", Tier: 3, PointCost: 2, RequiredLevel: 35,
						IsPassive: true,
						Description: "Passivo: Todo ataque básico tem 40% de chance de aplicar veneno por 3 turnos (4/turno).",
						Requires: []string{"rog_plague_blade"},
					},
					{
						ID: "rog_viral_toxin", Class: "rogue", Branch: "venenos",
						Name: "Toxina Viral", Emoji: "☠️", Tier: 4, PointCost: 3, RequiredLevel: 45,
						MPCost: 80, EnergyCost: 18, IsUltimate: true,
						Description: "Toxina viral extremamente potente: 6d10+DEX de dano de veneno + veneno+maldição por 8 turnos (12/turno).",
						Requires: []string{"rog_venomous_strikes"},
						Effects: []engine_Effect{
							{TypeName: "damage", Element: "poison", BasePower: 85},
							{TypeName: "apply_status", StatusKind: "poison", StatusTurns: 8, StatusDmgPT: 12},
							{TypeName: "apply_status", StatusKind: "curse", StatusTurns: 4},
						},
					},
				},
			},
			{
				ID: "esperteza", Name: "Esperteza", Emoji: "🎲",
				Nodes: []SkillNode{
					// T1 — Nível 1
					{
						ID: "rog_cheap_shot", Class: "rogue", Branch: "esperteza",
						Name: "Golpe Sujo", Emoji: "👊", Tier: 1, PointCost: 1, RequiredLevel: 1,
						MPCost: 8, EnergyCost: 2,
						Description: "Ataque traiçoeiro em ponto fraco: 1d8+DEX de dano e Stun por 1 turno.",
						Effects: []engine_Effect{
							{TypeName: "damage", Element: "physical", BasePower: 18},
							{TypeName: "apply_status", StatusKind: "stun", StatusTurns: 1},
						},
					},
					{
						ID: "rog_distraction", Class: "rogue", Branch: "esperteza",
						Name: "Distração", Emoji: "🎭", Tier: 1, PointCost: 1, RequiredLevel: 6,
						MPCost: 12, EnergyCost: 3,
						Description: "Distrai o inimigo com truque: aplica Cegueira por 2 turnos e reduz ATK em 25%.",
						Requires: []string{"rog_cheap_shot"},
						Effects: []engine_Effect{
							{TypeName: "apply_status", StatusKind: "blind", StatusTurns: 2},
							{TypeName: "stat_debuff", StatName: "attack", StatDelta: -6},
						},
					},
					{
						ID: "rog_smoke_bomb", Class: "rogue", Branch: "esperteza",
						Name: "Bomba de Fumaça", Emoji: "💨", Tier: 2, PointCost: 2, RequiredLevel: 18,
						MPCost: 25, EnergyCost: 6,
						Description: "Joga bomba de fumaça em AoE: Cegueira por 3 turnos a todos + +30% esquiva pessoal.",
						Requires: []string{"rog_distraction"},
						Effects: []engine_Effect{
							{TypeName: "aoe", Element: "physical", BasePower: 5},
							{TypeName: "apply_status", StatusKind: "blind", StatusTurns: 3},
						},
					},
					{
						ID: "rog_dirty_fighting", Class: "rogue", Branch: "esperteza",
						Name: "Luta Suja", Emoji: "😈", Tier: 3, PointCost: 2, RequiredLevel: 25,
						MPCost: 35, EnergyCost: 8,
						Description: "Ataque com múltiplos truques sujos: Stun + Cegueira + -30% ATK do inimigo, tudo por 2 turnos.",
						Requires: []string{"rog_smoke_bomb"},
						Effects: []engine_Effect{
							{TypeName: "damage", Element: "physical", BasePower: 25},
							{TypeName: "apply_status", StatusKind: "stun", StatusTurns: 1},
							{TypeName: "apply_status", StatusKind: "blind", StatusTurns: 2},
							{TypeName: "stat_debuff", StatName: "attack", StatDelta: -8},
						},
					},
					{
						ID: "rog_vanish", Class: "rogue", Branch: "esperteza",
						Name: "Sumir", Emoji: "🌑", Tier: 3, PointCost: 2, RequiredLevel: 32,
						MPCost: 40, EnergyCost: 10,
						Description: "Desaparece completamente: evita todo dano por 1 turno, reinicia cooldowns e reposiciona.",
						Requires: []string{"rog_dirty_fighting"},
						Effects: []engine_Effect{
							{TypeName: "apply_status", StatusKind: "haste", StatusTurns: 2},
							{TypeName: "apply_status", StatusKind: "shield", StatusTurns: 1},
						},
					},
					{
						ID: "rog_coup_de_grace", Class: "rogue", Branch: "esperteza",
						Name: "Golpe de Misericórdia", Emoji: "🏆", Tier: 4, PointCost: 3, RequiredLevel: 45,
						MPCost: 75, EnergyCost: 16, IsUltimate: true,
						Description: "O golpe final perfeito: 7d12+DEX de dano físico. Inimigo atordoado → dano dobrado. Inimigo abaixo de 20% HP → instakill.",
						Requires: []string{"rog_vanish"},
						Effects: []engine_Effect{
							{TypeName: "damage", Element: "physical", BasePower: 105},
						},
					},
					{
						ID: "rog_pickpocket", Class: "rogue", Branch: "esperteza",
						Name: "Bater Carteira", Emoji: "💰", Tier: 2, PointCost: 1, RequiredLevel: 12,
						MPCost: 14, EnergyCost: 4,
						Description: "Rouba ouro do inimigo durante o combate: 10–30% do ouro de recompensa antecipado e aplica Cegueira por 1 turno.",
						Requires: []string{"rog_distraction"},
						Effects: []engine_Effect{
							{TypeName: "apply_status", StatusKind: "blind", StatusTurns: 1},
						},
					},
					{
						ID: "rog_shadow_arts", Class: "rogue", Branch: "esperteza",
						Name: "Artes das Sombras", Emoji: "🌑", Tier: 3, PointCost: 2, RequiredLevel: 38,
						IsPassive: true,
						Description: "Passivo: +12% dano total, +15% crit sob Haste e +10% esquiva global.",
						Requires: []string{"rog_vanish"},
					},
					{
						ID: "rog_master_thief", Class: "rogue", Branch: "esperteza",
						Name: "Ladrão Mestre", Emoji: "💎", Tier: 4, PointCost: 2, RequiredLevel: 50,
						IsPassive: true,
						Description: "Passivo: +15% ouro de recompensas, +10% crit global e 25% de chance de roubar um item ao matar inimigo.",
						Requires: []string{"rog_coup_de_grace"},
					},
				},
			},
		},
	}
}
