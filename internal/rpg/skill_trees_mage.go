package rpg

func init() {
	Trees["mage"] = mageTree()
}

// mageTree returns the full skill tree for the mage (arcanista) class.
// 27 skills across 3 branches: Arcanismo, Elementalismo, Transmutação.
func mageTree() *SkillTree {
	return &SkillTree{
		ClassID: "mage",
		Branches: []Branch{
			{
				ID: "arcanismo", Name: "Arcanismo", Emoji: "✨",
				Nodes: []SkillNode{
					// T1 — Nível 1
					{
						ID: "mage_magic_missile", Class: "mage", Branch: "arcanismo",
						Name: "Míssil Mágico", Emoji: "✨", Tier: 1, PointCost: 1, RequiredLevel: 1,
						MPCost: 10, EnergyCost: 2,
						Description: "Dispara 3 projéteis arcanos que causam 1d4+INT de dano cada. Sempre acerta.",
						Effects: []engine_Effect{
							{TypeName: "damage", Element: "magic", BasePower: 12},
							{TypeName: "damage", Element: "magic", BasePower: 12},
							{TypeName: "damage", Element: "magic", BasePower: 12},
						},
					},
					{
						ID: "mage_arcane_shield", Class: "mage", Branch: "arcanismo",
						Name: "Escudo Arcano", Emoji: "🔮", Tier: 1, PointCost: 1, RequiredLevel: 3,
						MPCost: 15, EnergyCost: 3,
						Description: "Cria uma barreira mágica que absorve até 35 pontos de dano por 2 turnos.",
						Requires: []string{"mage_magic_missile"},
						Effects: []engine_Effect{
							{TypeName: "apply_status", StatusKind: "shield", StatusTurns: 2},
						},
					},
					{
						ID: "mage_mana_surge", Class: "mage", Branch: "arcanismo",
						Name: "Surto de Mana", Emoji: "💫", Tier: 1, PointCost: 1, RequiredLevel: 5,
						IsPassive: true,
						Description: "Passivo: +30 MP máximo e reduz o custo de MP de todas as magias em 5%.",
						Requires: []string{"mage_arcane_shield"},
					},
					// T2 — Nível 10
					{
						ID: "mage_force_blast", Class: "mage", Branch: "arcanismo",
						Name: "Explosão de Força", Emoji: "💥", Tier: 2, PointCost: 1, RequiredLevel: 10,
						MPCost: 25, EnergyCost: 5,
						Description: "Onda de força arcana que causa 3d8+INT de dano e aplica Stun por 1 turno.",
						Requires: []string{"mage_mana_surge"},
						Effects: []engine_Effect{
							{TypeName: "damage", Element: "magic", BasePower: 38},
							{TypeName: "apply_status", StatusKind: "stun", StatusTurns: 1},
						},
					},
					{
						ID: "mage_arcane_torrent", Class: "mage", Branch: "arcanismo",
						Name: "Torrente Arcana", Emoji: "🌊", Tier: 2, PointCost: 1, RequiredLevel: 14,
						MPCost: 30, EnergyCost: 5,
						Description: "Canaliza um fluxo de energia arcana por 3 golpes de 1d10+INT cada.",
						Requires: []string{"mage_force_blast"},
						Effects: []engine_Effect{
							{TypeName: "damage", Element: "magic", BasePower: 22},
							{TypeName: "damage", Element: "magic", BasePower: 22},
							{TypeName: "damage", Element: "magic", BasePower: 22},
						},
					},
					{
						ID: "mage_spell_power", Class: "mage", Branch: "arcanismo",
						Name: "Poder das Magias", Emoji: "⚡", Tier: 2, PointCost: 2, RequiredLevel: 18,
						IsPassive: true,
						Description: "Passivo: +8 MAG e todas as magias de dano causam +15% de dano.",
						Requires: []string{"mage_arcane_torrent"},
					},
					// T3 — Nível 25
					{
						ID: "mage_singularity", Class: "mage", Branch: "arcanismo",
						Name: "Singularidade", Emoji: "🌑", Tier: 3, PointCost: 2, RequiredLevel: 25,
						MPCost: 50, EnergyCost: 10,
						Description: "Cria um ponto de singularidade que puxa e causa 5d10+INT de dano arcano em AoE.",
						Requires: []string{"mage_spell_power"},
						Effects: []engine_Effect{
							{TypeName: "aoe", Element: "magic", BasePower: 65},
						},
					},
					{
						ID: "mage_arcane_mastery", Class: "mage", Branch: "arcanismo",
						Name: "Maestria Arcana", Emoji: "📚", Tier: 3, PointCost: 2, RequiredLevel: 30,
						IsPassive: true,
						Description: "Passivo: Reduz custo de MP em 20%, +12 MAG e +10% de chance de crítico mágico.",
						Requires: []string{"mage_singularity"},
					},
					{
						ID: "mage_arcane_annihilation", Class: "mage", Branch: "arcanismo",
						Name: "Aniquilação Arcana", Emoji: "💠", Tier: 4, PointCost: 3, RequiredLevel: 45,
						MPCost: 90, EnergyCost: 20, IsUltimate: true,
						Description: "Canaliza toda energia arcana em um raio devastador: 8d12+INT de dano mágico puro. Ignora resistências.",
						Requires: []string{"mage_arcane_mastery"},
						Effects: []engine_Effect{
							{TypeName: "damage", Element: "magic", BasePower: 110},
						},
					},
				},
			},
			{
				ID: "elementalismo", Name: "Elementalismo", Emoji: "🔥",
				Nodes: []SkillNode{
					// T1 — Nível 1
					{
						ID: "mage_fireball", Class: "mage", Branch: "elementalismo",
						Name: "Bola de Fogo", Emoji: "🔥", Tier: 1, PointCost: 1, RequiredLevel: 1,
						MPCost: 12, EnergyCost: 3,
						Description: "Lança uma esfera de fogo que causa 2d6+INT de dano e aplica queimadura por 2 turnos.",
						Effects: []engine_Effect{
							{TypeName: "damage", Element: "fire", BasePower: 22},
							{TypeName: "apply_status", StatusKind: "burn", StatusTurns: 2, StatusDmgPT: 4},
						},
					},
					{
						ID: "mage_ice_lance", Class: "mage", Branch: "elementalismo",
						Name: "Lança de Gelo", Emoji: "❄️", Tier: 1, PointCost: 1, RequiredLevel: 4,
						MPCost: 14, EnergyCost: 3,
						Description: "Projétil de gelo que causa 2d8+INT de dano e aplica Lentidão por 2 turnos.",
						Requires: []string{"mage_fireball"},
						Effects: []engine_Effect{
							{TypeName: "damage", Element: "ice", BasePower: 24},
							{TypeName: "stat_debuff", StatName: "speed", StatDelta: -4},
						},
					},
					{
						ID: "mage_lightning_bolt", Class: "mage", Branch: "elementalismo",
						Name: "Raio", Emoji: "⚡", Tier: 2, PointCost: 1, RequiredLevel: 10,
						MPCost: 20, EnergyCost: 4,
						Description: "Raio que causa 2d10+INT de dano elétrico e atordoa o alvo por 1 turno.",
						Requires: []string{"mage_ice_lance"},
						Effects: []engine_Effect{
							{TypeName: "damage", Element: "lightning", BasePower: 30},
							{TypeName: "apply_status", StatusKind: "stun", StatusTurns: 1},
						},
					},
					{
						ID: "mage_inferno", Class: "mage", Branch: "elementalismo",
						Name: "Inferno", Emoji: "🌋", Tier: 2, PointCost: 2, RequiredLevel: 15,
						MPCost: 35, EnergyCost: 7,
						Description: "Cria um campo de fogo em AoE que causa 3d8+INT e aplica queimadura em todos.",
						Requires: []string{"mage_lightning_bolt"},
						Effects: []engine_Effect{
							{TypeName: "aoe", Element: "fire", BasePower: 42},
							{TypeName: "apply_status", StatusKind: "burn", StatusTurns: 3, StatusDmgPT: 6},
						},
					},
					{
						ID: "mage_blizzard", Class: "mage", Branch: "elementalismo",
						Name: "Nevasca", Emoji: "🌨️", Tier: 2, PointCost: 2, RequiredLevel: 18,
						MPCost: 40, EnergyCost: 8,
						Description: "Desencadeia uma nevasca em AoE: 3d8+INT de dano de gelo e Freeze por 2 turnos.",
						Requires: []string{"mage_inferno"},
						Effects: []engine_Effect{
							{TypeName: "aoe", Element: "ice", BasePower: 40},
							{TypeName: "apply_status", StatusKind: "freeze", StatusTurns: 2},
						},
					},
					{
						ID: "mage_chain_lightning", Class: "mage", Branch: "elementalismo",
						Name: "Raio em Cadeia", Emoji: "⚡", Tier: 3, PointCost: 2, RequiredLevel: 25,
						MPCost: 50, EnergyCost: 10,
						Description: "Raio que salta entre inimigos causando 3d10+INT no primeiro e 60% nas ricochetes.",
						Requires: []string{"mage_blizzard"},
						Effects: []engine_Effect{
							{TypeName: "damage", Element: "lightning", BasePower: 55},
							{TypeName: "damage", Element: "lightning", BasePower: 33},
						},
					},
					{
						ID: "mage_meteor", Class: "mage", Branch: "elementalismo",
						Name: "Meteoro", Emoji: "☄️", Tier: 3, PointCost: 2, RequiredLevel: 30,
						MPCost: 60, EnergyCost: 12,
						Description: "Invoca um meteoro que causa 6d10+INT de dano de fogo em uma área enorme.",
						Requires: []string{"mage_chain_lightning"},
						Effects: []engine_Effect{
							{TypeName: "aoe", Element: "fire", BasePower: 80},
						},
					},
					{
						ID: "mage_frozen_tomb", Class: "mage", Branch: "elementalismo",
						Name: "Túmulo de Gelo", Emoji: "🧊", Tier: 3, PointCost: 2, RequiredLevel: 35,
						MPCost: 55, EnergyCost: 10,
						Description: "Encobre o alvo em gelo sólido: 4d10+INT de dano e Freeze por 3 turnos.",
						Requires: []string{"mage_meteor"},
						Effects: []engine_Effect{
							{TypeName: "damage", Element: "ice", BasePower: 65},
							{TypeName: "apply_status", StatusKind: "freeze", StatusTurns: 3},
						},
					},
					{
						ID: "mage_prismatic_ray", Class: "mage", Branch: "elementalismo",
						Name: "Raio Prismático", Emoji: "🌈", Tier: 4, PointCost: 3, RequiredLevel: 50,
						MPCost: 100, EnergyCost: 22, IsUltimate: true,
						Description: "Raio de todos os elementos: Fogo+Gelo+Raio em AoE. 10d12+INT de dano elementar combinado.",
						Requires: []string{"mage_frozen_tomb"},
						Effects: []engine_Effect{
							{TypeName: "aoe", Element: "fire", BasePower: 50},
							{TypeName: "aoe", Element: "ice", BasePower: 50},
							{TypeName: "aoe", Element: "lightning", BasePower: 50},
						},
					},
				},
			},
			{
				ID: "transmutacao", Name: "Transmutação", Emoji: "🔄",
				Nodes: []SkillNode{
					// T1 — Nível 1
					{
						ID: "mage_haste", Class: "mage", Branch: "transmutacao",
						Name: "Pressa", Emoji: "💨", Tier: 1, PointCost: 1, RequiredLevel: 1,
						MPCost: 15, EnergyCost: 3,
						Description: "Acelera o tempo ao redor do alvo: aplica Haste por 3 turnos (+30% velocidade e esquiva).",
						Effects: []engine_Effect{
							{TypeName: "apply_status", StatusKind: "haste", StatusTurns: 3},
						},
					},
					{
						ID: "mage_weakness", Class: "mage", Branch: "transmutacao",
						Name: "Fraqueza", Emoji: "💔", Tier: 1, PointCost: 1, RequiredLevel: 5,
						MPCost: 18, EnergyCost: 3,
						Description: "Aplica Fraqueza no alvo: -20% DEF e -15% ATK por 3 turnos.",
						Requires: []string{"mage_haste"},
						Effects: []engine_Effect{
							{TypeName: "stat_debuff", StatName: "defense", StatDelta: -7},
							{TypeName: "stat_debuff", StatName: "attack", StatDelta: -5},
						},
					},
					{
						ID: "mage_time_warp", Class: "mage", Branch: "transmutacao",
						Name: "Distorção Temporal", Emoji: "⏰", Tier: 2, PointCost: 1, RequiredLevel: 10,
						MPCost: 30, EnergyCost: 6,
						Description: "Dobra o tempo: age duas vezes neste turno (usa a próxima habilidade sem custo extra).",
						Requires: []string{"mage_weakness"},
						Effects: []engine_Effect{
							{TypeName: "apply_status", StatusKind: "haste", StatusTurns: 1},
						},
					},
					{
						ID: "mage_polymorph", Class: "mage", Branch: "transmutacao",
						Name: "Polimorfismo", Emoji: "🐸", Tier: 2, PointCost: 2, RequiredLevel: 15,
						MPCost: 35, EnergyCost: 7,
						Description: "Transforma o inimigo temporariamente: Silêncio + -40% ATK por 3 turnos.",
						Requires: []string{"mage_time_warp"},
						Effects: []engine_Effect{
							{TypeName: "apply_status", StatusKind: "silence", StatusTurns: 3},
							{TypeName: "stat_debuff", StatName: "attack", StatDelta: -12},
						},
					},
					{
						ID: "mage_mana_void", Class: "mage", Branch: "transmutacao",
						Name: "Vácuo de Mana", Emoji: "🌀", Tier: 2, PointCost: 2, RequiredLevel: 20,
						MPCost: 40, EnergyCost: 8,
						Description: "Drena o MP do alvo causando 1 de dano para cada 3 MP drenados. Drena até 60 MP.",
						Requires: []string{"mage_polymorph"},
						Effects: []engine_Effect{
							{TypeName: "damage", Element: "magic", BasePower: 30},
						},
					},
					{
						ID: "mage_time_stop", Class: "mage", Branch: "transmutacao",
						Name: "Parar o Tempo", Emoji: "⏸️", Tier: 3, PointCost: 2, RequiredLevel: 28,
						MPCost: 60, EnergyCost: 12,
						Description: "Para o tempo ao redor do inimigo: Stun por 2 turnos e -50% DEF enquanto dura.",
						Requires: []string{"mage_mana_void"},
						Effects: []engine_Effect{
							{TypeName: "apply_status", StatusKind: "stun", StatusTurns: 2},
							{TypeName: "stat_debuff", StatName: "defense", StatDelta: -15},
						},
					},
					{
						ID: "mage_blink", Class: "mage", Branch: "transmutacao",
						Name: "Piscar", Emoji: "👁️", Tier: 3, PointCost: 2, RequiredLevel: 32,
						MPCost: 45, EnergyCost: 8,
						Description: "Teletransporte instantâneo que evita o próximo ataque e responde com dano arcano de 3d8+INT.",
						Requires: []string{"mage_time_stop"},
						Effects: []engine_Effect{
							{TypeName: "apply_status", StatusKind: "haste", StatusTurns: 1},
							{TypeName: "damage", Element: "magic", BasePower: 35},
						},
					},
					{
						ID: "mage_dispel", Class: "mage", Branch: "transmutacao",
						Name: "Dissipar", Emoji: "✨", Tier: 3, PointCost: 2, RequiredLevel: 36,
						MPCost: 40, EnergyCost: 7,
						Description: "Remove todos os efeitos positivos do alvo e aplica Silêncio por 2 turnos.",
						Requires: []string{"mage_blink"},
						Effects: []engine_Effect{
							{TypeName: "remove_status", StatusKind: "shield"},
							{TypeName: "remove_status", StatusKind: "protect"},
							{TypeName: "apply_status", StatusKind: "silence", StatusTurns: 2},
						},
					},
					{
						ID: "mage_reality_break", Class: "mage", Branch: "transmutacao",
						Name: "Ruptura da Realidade", Emoji: "💀", Tier: 4, PointCost: 3, RequiredLevel: 55,
						MPCost: 110, EnergyCost: 25, IsUltimate: true,
						Description: "Rasga a realidade ao redor do alvo: 8d12+INT de dano mágico puro + todos os debuffs + Silêncio por 3 turnos.",
						Requires: []string{"mage_dispel"},
						Effects: []engine_Effect{
							{TypeName: "damage", Element: "magic", BasePower: 120},
							{TypeName: "apply_status", StatusKind: "silence", StatusTurns: 3},
							{TypeName: "apply_status", StatusKind: "curse", StatusTurns: 3},
							{TypeName: "stat_debuff", StatName: "defense", StatDelta: -20},
						},
					},
				},
			},
		},
	}
}
