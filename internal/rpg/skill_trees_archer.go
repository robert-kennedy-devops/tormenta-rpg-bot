package rpg

func init() {
	Trees["archer"] = archerTree()
}

// archerTree returns the full skill tree for the archer class.
// 26 skills across 3 branches: Precisão, Natureza, Sobrevivência.
func archerTree() *SkillTree {
	return &SkillTree{
		ClassID: "archer",
		Branches: []Branch{
			{
				ID: "precisao", Name: "Precisão", Emoji: "🎯",
				Nodes: []SkillNode{
					// T1 — Nível 1
					{
						ID: "arch_aimed_shot", Class: "archer", Branch: "precisao",
						Name: "Tiro Mirado", Emoji: "🎯", Tier: 1, PointCost: 1, RequiredLevel: 1,
						MPCost: 8, EnergyCost: 2,
						Description: "Tiro cuidadosamente mirado: 2d8+DEX de dano com +20% precisão. Nunca erra em alvos parados.",
						Effects: []engine_Effect{
							{TypeName: "damage", Element: "physical", BasePower: 26},
						},
					},
					{
						ID: "arch_eagle_eye", Class: "archer", Branch: "precisao",
						Name: "Olho de Águia", Emoji: "🦅", Tier: 1, PointCost: 1, RequiredLevel: 3,
						IsPassive: true,
						Description: "Passivo: +15% de chance de crítico, +6 DEX efetivo e pode atingir alvos com Escudo sem penalidade.",
						Requires: []string{"arch_aimed_shot"},
					},
					{
						ID: "arch_piercing_arrow", Class: "archer", Branch: "precisao",
						Name: "Flecha Perfurante", Emoji: "🏹", Tier: 1, PointCost: 1, RequiredLevel: 6,
						MPCost: 14, EnergyCost: 3,
						Description: "Flecha com ponta de diamante que ignora 60% da armadura do alvo. 2d10+DEX de dano.",
						Requires: []string{"arch_eagle_eye"},
						Effects: []engine_Effect{
							{TypeName: "damage", Element: "physical", BasePower: 32},
							{TypeName: "stat_debuff", StatName: "defense", StatDelta: -8},
						},
					},
					// T2 — Nível 10
					{
						ID: "arch_power_shot", Class: "archer", Branch: "precisao",
						Name: "Tiro Poderoso", Emoji: "💥", Tier: 2, PointCost: 1, RequiredLevel: 10,
						MPCost: 22, EnergyCost: 5,
						Description: "Tiro com força máxima: 4d10+DEX de dano e empurra o alvo (stun 1 turno).",
						Requires: []string{"arch_piercing_arrow"},
						Effects: []engine_Effect{
							{TypeName: "damage", Element: "physical", BasePower: 50},
							{TypeName: "apply_status", StatusKind: "stun", StatusTurns: 1},
						},
					},
					{
						ID: "arch_multi_shot", Class: "archer", Branch: "precisao",
						Name: "Tiro Múltiplo", Emoji: "🌟", Tier: 2, PointCost: 2, RequiredLevel: 15,
						MPCost: 28, EnergyCost: 6,
						Description: "Dispara 3 flechas simultaneamente causando 1d10+DEX cada. Pode atingir alvos diferentes.",
						Requires: []string{"arch_power_shot"},
						Effects: []engine_Effect{
							{TypeName: "damage", Element: "physical", BasePower: 26},
							{TypeName: "damage", Element: "physical", BasePower: 26},
							{TypeName: "damage", Element: "physical", BasePower: 26},
						},
					},
					{
						ID: "arch_sniper", Class: "archer", Branch: "precisao",
						Name: "Franco-Atirador", Emoji: "🔭", Tier: 2, PointCost: 2, RequiredLevel: 20,
						IsPassive: true,
						Description: "Passivo: +25% dano em ataques de longa distância, +10% crit e nunca penalidade de distância.",
						Requires: []string{"arch_multi_shot"},
					},
					// T3 — Nível 28
					{
						ID: "arch_barrage", Class: "archer", Branch: "precisao",
						Name: "Bombardeio", Emoji: "🌧️", Tier: 3, PointCost: 2, RequiredLevel: 28,
						MPCost: 50, EnergyCost: 10,
						Description: "Chuva de flechas: 5 flechas rápidas de 1d8+DEX cada em alvos aleatórios.",
						Requires: []string{"arch_sniper"},
						Effects: []engine_Effect{
							{TypeName: "damage", Element: "physical", BasePower: 22},
							{TypeName: "damage", Element: "physical", BasePower: 22},
							{TypeName: "damage", Element: "physical", BasePower: 22},
							{TypeName: "damage", Element: "physical", BasePower: 22},
							{TypeName: "damage", Element: "physical", BasePower: 22},
						},
					},
					{
						ID: "arch_headshot", Class: "archer", Branch: "precisao",
						Name: "Tiro na Cabeça", Emoji: "🎯", Tier: 3, PointCost: 2, RequiredLevel: 35,
						MPCost: 55, EnergyCost: 12,
						Description: "Tiro precisíssimo na cabeça: 5d12+DEX de dano + Stun por 2 turnos. Sempre crítico.",
						Requires: []string{"arch_barrage"},
						Effects: []engine_Effect{
							{TypeName: "damage", Element: "physical", BasePower: 80},
							{TypeName: "apply_status", StatusKind: "stun", StatusTurns: 2},
						},
					},
					{
						ID: "arch_divine_arrow", Class: "archer", Branch: "precisao",
						Name: "Flecha Divina", Emoji: "✨", Tier: 4, PointCost: 3, RequiredLevel: 45,
						MPCost: 85, EnergyCost: 20, IsUltimate: true,
						Description: "Flecha abençoada pelos deuses: 8d12+DEX de dano sagrado. Crítico garantido. Dano triplo em mortos-vivos.",
						Requires: []string{"arch_headshot"},
						Effects: []engine_Effect{
							{TypeName: "damage", Element: "holy", BasePower: 115},
						},
					},
				},
			},
			{
				ID: "natureza", Name: "Natureza", Emoji: "🌿",
				Nodes: []SkillNode{
					// T1 — Nível 1
					{
						ID: "arch_nature_arrow", Class: "archer", Branch: "natureza",
						Name: "Flecha da Natureza", Emoji: "🌿", Tier: 1, PointCost: 1, RequiredLevel: 1,
						MPCost: 10, EnergyCost: 2,
						Description: "Flecha impregnada com veneno de planta: 1d8+DEX de dano + veneno por 4 turnos (4/turno).",
						Effects: []engine_Effect{
							{TypeName: "damage", Element: "poison", BasePower: 18},
							{TypeName: "apply_status", StatusKind: "poison", StatusTurns: 4, StatusDmgPT: 4},
						},
					},
					{
						ID: "arch_vine_trap", Class: "archer", Branch: "natureza",
						Name: "Armadilha de Vinhas", Emoji: "🌱", Tier: 1, PointCost: 1, RequiredLevel: 5,
						MPCost: 14, EnergyCost: 3,
						Description: "Planta armadilha de vinhas: quando ativada, aplica Stun por 2 turnos e causa 1d6 de dano.",
						Requires: []string{"arch_nature_arrow"},
						Effects: []engine_Effect{
							{TypeName: "damage", Element: "physical", BasePower: 16},
							{TypeName: "apply_status", StatusKind: "stun", StatusTurns: 2},
						},
					},
					{
						ID: "arch_wind_arrow", Class: "archer", Branch: "natureza",
						Name: "Flecha do Vento", Emoji: "💨", Tier: 2, PointCost: 1, RequiredLevel: 12,
						MPCost: 22, EnergyCost: 4,
						Description: "Flecha carregada com vento que causa 2d8+DEX de dano e AoE de impacto empurrando inimigos.",
						Requires: []string{"arch_vine_trap"},
						Effects: []engine_Effect{
							{TypeName: "damage", Element: "physical", BasePower: 30},
							{TypeName: "aoe", Element: "physical", BasePower: 15},
						},
					},
					{
						ID: "arch_storm_arrow", Class: "archer", Branch: "natureza",
						Name: "Flecha da Tempestade", Emoji: "⛈️", Tier: 2, PointCost: 2, RequiredLevel: 18,
						MPCost: 32, EnergyCost: 7,
						Description: "Flecha elétrica da tempestade: 3d8+DEX de dano relâmpago e Stun por 1 turno.",
						Requires: []string{"arch_wind_arrow"},
						Effects: []engine_Effect{
							{TypeName: "damage", Element: "lightning", BasePower: 38},
							{TypeName: "apply_status", StatusKind: "stun", StatusTurns: 1},
						},
					},
					{
						ID: "arch_entangle", Class: "archer", Branch: "natureza",
						Name: "Enredar", Emoji: "🌳", Tier: 2, PointCost: 2, RequiredLevel: 22,
						MPCost: 30, EnergyCost: 6,
						Description: "Raízes mágicas brotam do chão enredando o alvo: Stun por 3 turnos + -40% velocidade.",
						Requires: []string{"arch_storm_arrow"},
						Effects: []engine_Effect{
							{TypeName: "apply_status", StatusKind: "stun", StatusTurns: 3},
							{TypeName: "stat_debuff", StatName: "speed", StatDelta: -8},
						},
					},
					{
						ID: "arch_call_of_wild", Class: "archer", Branch: "natureza",
						Name: "Chamado Selvagem", Emoji: "🐺", Tier: 3, PointCost: 2, RequiredLevel: 30,
						MPCost: 50, EnergyCost: 10,
						Description: "Invoca o espírito da natureza: aplica Haste em si mesmo e veneno em AoE (6/turno por 4 turnos).",
						Requires: []string{"arch_entangle"},
						Effects: []engine_Effect{
							{TypeName: "apply_status", StatusKind: "haste", StatusTurns: 3},
							{TypeName: "aoe", Element: "poison", BasePower: 20},
							{TypeName: "apply_status", StatusKind: "poison", StatusTurns: 4, StatusDmgPT: 6},
						},
					},
					{
						ID: "arch_nature_wrath", Class: "archer", Branch: "natureza",
						Name: "Ira da Natureza", Emoji: "🌪️", Tier: 3, PointCost: 2, RequiredLevel: 35,
						MPCost: 60, EnergyCost: 12,
						Description: "A natureza vinga: 6d10+DEX de dano de todos os elementos naturais em AoE.",
						Requires: []string{"arch_call_of_wild"},
						Effects: []engine_Effect{
							{TypeName: "aoe", Element: "physical", BasePower: 40},
							{TypeName: "aoe", Element: "poison", BasePower: 40},
							{TypeName: "aoe", Element: "lightning", BasePower: 25},
						},
					},
					{
						ID: "arch_avatar_nature", Class: "archer", Branch: "natureza",
						Name: "Avatar da Natureza", Emoji: "🌟", Tier: 4, PointCost: 3, RequiredLevel: 50,
						MPCost: 90, EnergyCost: 22, IsUltimate: true,
						Description: "Canaliza o poder da natureza: 8d12+DEX de dano, veneno por 8 turnos (15/turno), Haste por 4 turnos.",
						Requires: []string{"arch_nature_wrath"},
						Effects: []engine_Effect{
							{TypeName: "aoe", Element: "poison", BasePower: 80},
							{TypeName: "apply_status", StatusKind: "poison", StatusTurns: 8, StatusDmgPT: 15},
							{TypeName: "apply_status", StatusKind: "haste", StatusTurns: 4},
						},
					},
				},
			},
			{
				ID: "sobrevivencia", Name: "Sobrevivência", Emoji: "🏕️",
				Nodes: []SkillNode{
					// T1 — Nível 1
					{
						ID: "arch_camouflage", Class: "archer", Branch: "sobrevivencia",
						Name: "Camuflagem", Emoji: "🌿", Tier: 1, PointCost: 1, RequiredLevel: 1,
						IsPassive: true,
						Description: "Passivo: +18% de esquiva e +10% de crit quando o arqueiro não se moveu no turno anterior.",
					},
					{
						ID: "arch_trap", Class: "archer", Branch: "sobrevivencia",
						Name: "Armadilha", Emoji: "⚙️", Tier: 1, PointCost: 1, RequiredLevel: 8,
						MPCost: 15, EnergyCost: 4,
						Description: "Coloca armadilha explosiva no campo: quando ativada, causa 3d8+DEX e aplica Stun por 2 turnos.",
						Requires: []string{"arch_camouflage"},
						Effects: []engine_Effect{
							{TypeName: "damage", Element: "physical", BasePower: 38},
							{TypeName: "apply_status", StatusKind: "stun", StatusTurns: 2},
						},
					},
					{
						ID: "arch_quick_reflexes", Class: "archer", Branch: "sobrevivencia",
						Name: "Reflexos Rápidos", Emoji: "⚡", Tier: 2, PointCost: 2, RequiredLevel: 15,
						IsPassive: true,
						Description: "Passivo: +25% de esquiva, +5 velocidade e contra-ataca automaticamente ataques físicos (1d6).",
						Requires: []string{"arch_trap"},
					},
					{
						ID: "arch_hunter_mark", Class: "archer", Branch: "sobrevivencia",
						Name: "Marca do Caçador", Emoji: "🎯", Tier: 2, PointCost: 2, RequiredLevel: 20,
						MPCost: 20, EnergyCost: 5,
						Description: "Marca o alvo para abate: +35% de dano contra ele por 5 turnos e revela fraquezas.",
						Requires: []string{"arch_quick_reflexes"},
						Effects: []engine_Effect{
							{TypeName: "stat_debuff", StatName: "defense", StatDelta: -10},
						},
					},
					{
						ID: "arch_escape", Class: "archer", Branch: "sobrevivencia",
						Name: "Escapar", Emoji: "🏃", Tier: 3, PointCost: 2, RequiredLevel: 28,
						MPCost: 30, EnergyCost: 8,
						Description: "Salta para trás evitando golpe + contra-ataca com 3d10+DEX de dano.",
						Requires: []string{"arch_hunter_mark"},
						Effects: []engine_Effect{
							{TypeName: "apply_status", StatusKind: "haste", StatusTurns: 1},
							{TypeName: "damage", Element: "physical", BasePower: 45},
						},
					},
					{
						ID: "arch_wilderness_lore", Class: "archer", Branch: "sobrevivencia",
						Name: "Lore Selvagem", Emoji: "📜", Tier: 3, PointCost: 2, RequiredLevel: 35,
						IsPassive: true,
						Description: "Passivo: Resistência a veneno, gelo e fogo (+20% cada). Regenera 5 HP/turno em combate.",
						Requires: []string{"arch_escape"},
					},
					{
						ID: "arch_kill_shot", Class: "archer", Branch: "sobrevivencia",
						Name: "Tiro Fatal", Emoji: "💀", Tier: 4, PointCost: 3, RequiredLevel: 45,
						MPCost: 70, EnergyCost: 15, IsUltimate: true,
						Description: "O tiro perfeito: 7d12+DEX de dano físico. Inimigo abaixo de 25% HP → dano duplo e Stun automático.",
						Requires: []string{"arch_wilderness_lore"},
						Effects: []engine_Effect{
							{TypeName: "damage", Element: "physical", BasePower: 100},
						},
					},
					{
						ID: "arch_perfect_hunt", Class: "archer", Branch: "sobrevivencia",
						Name: "Caça Perfeita", Emoji: "🏆", Tier: 4, PointCost: 2, RequiredLevel: 55,
						IsPassive: true,
						Description: "Passivo: Mestre da caça — +20% dano total, +15% crit, +20% ouro de recompensas. Nunca perde flechas especiais.",
						Requires: []string{"arch_kill_shot"},
					},
				},
			},
		},
	}
}
