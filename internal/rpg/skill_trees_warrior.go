package rpg

func init() {
	Trees["warrior"] = warriorTree()
}

// warriorTree returns the full skill tree for the warrior class.
// 27 skills across 3 branches: Espada & Escudo, Maestria de Armas, Determinação.
func warriorTree() *SkillTree {
	return &SkillTree{
		ClassID: "warrior",
		Branches: []Branch{
			{
				ID: "espada_escudo", Name: "Espada & Escudo", Emoji: "⚔️",
				Nodes: []SkillNode{
					// T1 — Nível 1
					{
						ID: "warr_firm_strike", Class: "warrior", Branch: "espada_escudo",
						Name: "Golpe Firme", Emoji: "⚔️", Tier: 1, PointCost: 1, RequiredLevel: 1,
						MPCost: 5, EnergyCost: 2,
						Description: "Golpe preciso que adiciona +1d6 ao dano e reduz a CA do inimigo por 1 turno.",
						Effects: []engine_Effect{
							{TypeName: "damage", Element: "physical", BasePower: 18},
							{TypeName: "stat_debuff", StatName: "defense", StatDelta: -3},
						},
					},
					{
						ID: "warr_shield_bash", Class: "warrior", Branch: "espada_escudo",
						Name: "Pancada de Escudo", Emoji: "🛡️", Tier: 1, PointCost: 1, RequiredLevel: 3,
						MPCost: 8, EnergyCost: 3,
						Description: "Golpeia com o escudo causando dano físico e atordoando o alvo por 1 turno.",
						Requires: []string{"warr_firm_strike"},
						Effects: []engine_Effect{
							{TypeName: "damage", Element: "physical", BasePower: 12},
							{TypeName: "apply_status", StatusKind: "stun", StatusTurns: 1},
						},
					},
					{
						ID: "warr_parry", Class: "warrior", Branch: "espada_escudo",
						Name: "Aparar", Emoji: "🔰", Tier: 1, PointCost: 1, RequiredLevel: 5,
						IsPassive: true,
						Description: "Passivo: +4 de DEF e +5% de chance de deflexão em combate.",
						Requires: []string{"warr_shield_bash"},
					},
					// T2 — Nível 10
					{
						ID: "warr_power_slash", Class: "warrior", Branch: "espada_escudo",
						Name: "Corte Poderoso", Emoji: "🗡️", Tier: 2, PointCost: 1, RequiredLevel: 10,
						MPCost: 18, EnergyCost: 5,
						Description: "Golpe de espada carregado que causa 2d10+FOR de dano físico.",
						Requires: []string{"warr_parry"},
						Effects: []engine_Effect{
							{TypeName: "damage", Element: "physical", BasePower: 35},
						},
					},
					{
						ID: "warr_bulwark", Class: "warrior", Branch: "espada_escudo",
						Name: "Baluarte", Emoji: "🏰", Tier: 2, PointCost: 1, RequiredLevel: 12,
						MPCost: 20, EnergyCost: 4,
						Description: "Eleva um escudo de proteção que absorve 40% do próximo dano recebido.",
						Requires: []string{"warr_power_slash"},
						Effects: []engine_Effect{
							{TypeName: "apply_status", StatusKind: "shield", StatusTurns: 2},
						},
					},
					{
						ID: "warr_counter_strike", Class: "warrior", Branch: "espada_escudo",
						Name: "Contra-Golpe", Emoji: "↩️", Tier: 2, PointCost: 2, RequiredLevel: 15,
						MPCost: 22, EnergyCost: 5,
						Description: "Contra-ataque imediato que causa 1d8+FOR de dano após receber um golpe.",
						Requires: []string{"warr_bulwark"},
						Effects: []engine_Effect{
							{TypeName: "damage", Element: "physical", BasePower: 28},
						},
					},
					// T3 — Nível 20
					{
						ID: "warr_whirlwind", Class: "warrior", Branch: "espada_escudo",
						Name: "Redemoinho", Emoji: "🌀", Tier: 3, PointCost: 2, RequiredLevel: 20,
						MPCost: 35, EnergyCost: 8,
						Description: "Gira com a espada causando dano em área de 3d8+FOR a todos os inimigos adjacentes.",
						Requires: []string{"warr_counter_strike"},
						Effects: []engine_Effect{
							{TypeName: "aoe", Element: "physical", BasePower: 40},
						},
					},
					{
						ID: "warr_iron_will", Class: "warrior", Branch: "espada_escudo",
						Name: "Vontade de Ferro", Emoji: "🔩", Tier: 3, PointCost: 2, RequiredLevel: 25,
						IsPassive: true,
						Description: "Passivo: Regenera 5 HP por turno em combate e +15 HP máximo.",
						Requires: []string{"warr_whirlwind"},
					},
					{
						ID: "warr_bladestorm", Class: "warrior", Branch: "espada_escudo",
						Name: "Tempestade de Lâminas", Emoji: "⚡", Tier: 4, PointCost: 3, RequiredLevel: 35,
						MPCost: 55, EnergyCost: 12, IsUltimate: true,
						Description: "Ativa modo berserker e ataca 4 vezes causando 1d12+FOR em cada golpe. Dura 2 turnos.",
						Requires: []string{"warr_iron_will"},
						Effects: []engine_Effect{
							{TypeName: "apply_status", StatusKind: "berserk", StatusTurns: 2},
							{TypeName: "damage", Element: "physical", BasePower: 50},
							{TypeName: "damage", Element: "physical", BasePower: 50},
							{TypeName: "damage", Element: "physical", BasePower: 50},
							{TypeName: "damage", Element: "physical", BasePower: 50},
						},
					},
				},
			},
			{
				ID: "maestria_armas", Name: "Maestria de Armas", Emoji: "🗡️",
				Nodes: []SkillNode{
					// T1 — Nível 1
					{
						ID: "warr_quick_strike", Class: "warrior", Branch: "maestria_armas",
						Name: "Ataque Rápido", Emoji: "💨", Tier: 1, PointCost: 1, RequiredLevel: 1,
						MPCost: 6, EnergyCost: 2,
						Description: "Dois golpes rápidos que causam 1d6+DEX cada, com +15% de chance de acerto crítico.",
						Effects: []engine_Effect{
							{TypeName: "damage", Element: "physical", BasePower: 14},
							{TypeName: "damage", Element: "physical", BasePower: 14},
						},
					},
					{
						ID: "warr_rend", Class: "warrior", Branch: "maestria_armas",
						Name: "Rasgar", Emoji: "🩸", Tier: 1, PointCost: 1, RequiredLevel: 4,
						MPCost: 10, EnergyCost: 3,
						Description: "Causa ferida sangrante: 1d6 de dano imediato + 3 de dano por turno por 4 turnos.",
						Requires: []string{"warr_quick_strike"},
						Effects: []engine_Effect{
							{TypeName: "damage", Element: "physical", BasePower: 14},
							{TypeName: "apply_status", StatusKind: "poison", StatusTurns: 4, StatusDmgPT: 3},
						},
					},
					{
						ID: "warr_weapon_focus", Class: "warrior", Branch: "maestria_armas",
						Name: "Foco na Arma", Emoji: "🎯", Tier: 1, PointCost: 1, RequiredLevel: 6,
						IsPassive: true,
						Description: "Passivo: +5 de ATK e +8% de chance de acerto crítico.",
						Requires: []string{"warr_rend"},
					},
					// T2 — Nível 10
					{
						ID: "warr_savage_blow", Class: "warrior", Branch: "maestria_armas",
						Name: "Golpe Selvagem", Emoji: "💪", Tier: 2, PointCost: 1, RequiredLevel: 10,
						MPCost: 20, EnergyCost: 5,
						Description: "Golpe brutal com toda a força: 3d10+FOR de dano físico. Pode derrubar o alvo.",
						Requires: []string{"warr_weapon_focus"},
						Effects: []engine_Effect{
							{TypeName: "damage", Element: "physical", BasePower: 48},
						},
					},
					{
						ID: "warr_armor_pierce", Class: "warrior", Branch: "maestria_armas",
						Name: "Perfurar Armadura", Emoji: "🔱", Tier: 2, PointCost: 1, RequiredLevel: 14,
						MPCost: 25, EnergyCost: 5,
						Description: "Golpe preciso que ignora 50% da armadura do alvo. Causa 2d8+FOR de dano.",
						Requires: []string{"warr_savage_blow"},
						Effects: []engine_Effect{
							{TypeName: "damage", Element: "physical", BasePower: 38},
							{TypeName: "stat_debuff", StatName: "defense", StatDelta: -8},
						},
					},
					{
						ID: "warr_battle_rage", Class: "warrior", Branch: "maestria_armas",
						Name: "Fúria de Batalha", Emoji: "💢", Tier: 2, PointCost: 2, RequiredLevel: 18,
						MPCost: 0, EnergyCost: 8,
						Description: "Entra em estado de fúria por 3 turnos: +40% ATK, velocidade de ataque dobrada.",
						Requires: []string{"warr_armor_pierce"},
						Effects: []engine_Effect{
							{TypeName: "apply_status", StatusKind: "berserk", StatusTurns: 3},
						},
					},
					// T3 — Nível 25
					{
						ID: "warr_execute", Class: "warrior", Branch: "maestria_armas",
						Name: "Executar", Emoji: "☠️", Tier: 3, PointCost: 2, RequiredLevel: 25,
						MPCost: 40, EnergyCost: 8,
						Description: "Golpe de execução: causa 4d12+FOR de dano. Dano dobrado se inimigo tiver menos de 30% HP.",
						Requires: []string{"warr_battle_rage"},
						Effects: []engine_Effect{
							{TypeName: "damage", Element: "physical", BasePower: 60},
						},
					},
					{
						ID: "warr_cleave", Class: "warrior", Branch: "maestria_armas",
						Name: "Clivar", Emoji: "⚔️", Tier: 3, PointCost: 2, RequiredLevel: 28,
						MPCost: 35, EnergyCost: 8,
						Description: "Corte em arco que atinge todos os inimigos na frente por 3d8+FOR de dano.",
						Requires: []string{"warr_execute"},
						Effects: []engine_Effect{
							{TypeName: "aoe", Element: "physical", BasePower: 45},
						},
					},
					{
						ID: "warr_colossus_smash", Class: "warrior", Branch: "maestria_armas",
						Name: "Golpe do Colosso", Emoji: "🏔️", Tier: 4, PointCost: 3, RequiredLevel: 40,
						MPCost: 65, EnergyCost: 15, IsUltimate: true,
						Description: "Smash devastador: 6d12+FOR de dano físico. Destroça armadura por 5 turnos (-50% DEF do alvo).",
						Requires: []string{"warr_cleave"},
						Effects: []engine_Effect{
							{TypeName: "damage", Element: "physical", BasePower: 85},
							{TypeName: "stat_debuff", StatName: "defense", StatDelta: -20},
						},
					},
				},
			},
			{
				ID: "determinacao", Name: "Determinação", Emoji: "🔥",
				Nodes: []SkillNode{
					// T1 — Nível 1
					{
						ID: "warr_second_wind", Class: "warrior", Branch: "determinacao",
						Name: "Segundo Fôlego", Emoji: "💨", Tier: 1, PointCost: 1, RequiredLevel: 1,
						MPCost: 0, EnergyCost: 5,
						Description: "Recupera 2d8+CON de HP instantaneamente em combate.",
						Effects: []engine_Effect{
							{TypeName: "heal", BasePower: 25},
						},
					},
					{
						ID: "warr_grit", Class: "warrior", Branch: "determinacao",
						Name: "Garra", Emoji: "💪", Tier: 1, PointCost: 1, RequiredLevel: 5,
						IsPassive: true,
						Description: "Passivo: +6 de DEF e +20 HP máximo. Reduz dano recebido em 5%.",
						Requires: []string{"warr_second_wind"},
					},
					{
						ID: "warr_taunt", Class: "warrior", Branch: "determinacao",
						Name: "Provocar", Emoji: "😤", Tier: 1, PointCost: 1, RequiredLevel: 7,
						MPCost: 12, EnergyCost: 3,
						Description: "Provoca o inimigo fazendo-o focar nos próximos ataques no guerreiro, reduzindo seu ATK em 20%.",
						Requires: []string{"warr_grit"},
						Effects: []engine_Effect{
							{TypeName: "stat_debuff", StatName: "attack", StatDelta: -5},
						},
					},
					// T2 — Nível 12
					{
						ID: "warr_endurance", Class: "warrior", Branch: "determinacao",
						Name: "Endurance", Emoji: "🛡️", Tier: 2, PointCost: 1, RequiredLevel: 12,
						IsPassive: true,
						Description: "Passivo: +40 HP máximo e regenera 3 HP por turno fora de combate.",
						Requires: []string{"warr_taunt"},
					},
					{
						ID: "warr_stand_firm", Class: "warrior", Branch: "determinacao",
						Name: "Firme no Posto", Emoji: "⚓", Tier: 2, PointCost: 2, RequiredLevel: 16,
						MPCost: 25, EnergyCost: 6,
						Description: "Postura defensiva por 2 turnos: reduz todo dano recebido em 50% e aplica proteção.",
						Requires: []string{"warr_endurance"},
						Effects: []engine_Effect{
							{TypeName: "apply_status", StatusKind: "protect", StatusTurns: 2},
						},
					},
					{
						ID: "warr_indomable", Class: "warrior", Branch: "determinacao",
						Name: "Indomável", Emoji: "🔱", Tier: 2, PointCost: 2, RequiredLevel: 20,
						IsPassive: true,
						Description: "Passivo: Se o HP cair abaixo de 20%, ganha Berserk automaticamente por 2 turnos (1x por combate).",
						Requires: []string{"warr_stand_firm"},
					},
					// T3 — Nível 25
					{
						ID: "warr_war_cry", Class: "warrior", Branch: "determinacao",
						Name: "Grito de Guerra", Emoji: "📢", Tier: 3, PointCost: 2, RequiredLevel: 25,
						MPCost: 30, EnergyCost: 8,
						Description: "Grito intimidador: aplica Berserk em si mesmo e causa medo no inimigo (-30% ATK) por 3 turnos.",
						Requires: []string{"warr_indomable"},
						Effects: []engine_Effect{
							{TypeName: "apply_status", StatusKind: "berserk", StatusTurns: 3},
							{TypeName: "stat_debuff", StatName: "attack", StatDelta: -8},
						},
					},
					{
						ID: "warr_last_stand", Class: "warrior", Branch: "determinacao",
						Name: "Última Resistência", Emoji: "🔥", Tier: 3, PointCost: 2, RequiredLevel: 32,
						MPCost: 45, EnergyCost: 10,
						Description: "Concentra toda a força restante. Quanto menor o HP, maior o dano: 5d10 + (HP_max - HP_atual) / 2.",
						Requires: []string{"warr_war_cry"},
						Effects: []engine_Effect{
							{TypeName: "damage", Element: "physical", BasePower: 55},
						},
					},
					{
						ID: "warr_avatar_of_war", Class: "warrior", Branch: "determinacao",
						Name: "Avatar da Guerra", Emoji: "⚔️", Tier: 4, PointCost: 3, RequiredLevel: 50,
						MPCost: 80, EnergyCost: 20, IsUltimate: true,
						Description: "Transcende os limites mortais por 4 turnos: +100% ATK, +60% DEF, regenera 15 HP/turno e é imune a controle.",
						Requires: []string{"warr_last_stand"},
						Effects: []engine_Effect{
							{TypeName: "apply_status", StatusKind: "berserk", StatusTurns: 4},
							{TypeName: "apply_status", StatusKind: "protect", StatusTurns: 4},
							{TypeName: "apply_status", StatusKind: "regen", StatusTurns: 4},
						},
					},
				},
			},
		},
	}
}
