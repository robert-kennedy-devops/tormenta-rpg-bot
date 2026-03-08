package game

import (
	"math/rand"

	"github.com/tormenta-bot/internal/models"
)

// ── RACES ─────────────────────────────────────────────────
var Races = map[string]models.Race{
	"human": {
		ID: "human", Name: "Humano", Emoji: "👤",
		Description: "Adaptáveis e versáteis. Bônus em todas as classes.",
		BonusHP:     10, BonusMP: 5,
		BonusStr: 1, BonusDex: 1, BonusCon: 1, BonusInt: 1, BonusWis: 1, BonusCha: 2,
		Trait: "Versátil: +10% de XP ganha",
	},
	"elf": {
		ID: "elf", Name: "Elfo", Emoji: "🧝",
		Description: "Ágeis e conectados à magia.",
		BonusHP:     5, BonusMP: 20,
		BonusStr: 0, BonusDex: 3, BonusCon: 0, BonusInt: 3, BonusWis: 2, BonusCha: 1,
		Trait: "Visão Élfica: +20% dano com magia",
	},
	"dwarf": {
		ID: "dwarf", Name: "Anão", Emoji: "⛏️",
		Description: "Resistentes como as pedras que moldam.",
		BonusHP:     25, BonusMP: 0,
		BonusStr: 2, BonusDex: 0, BonusCon: 4, BonusInt: 0, BonusWis: 1, BonusCha: 0,
		Trait: "Pele de Pedra: -15% de dano recebido",
	},
	"halforc": {
		ID: "halforc", Name: "Meio-Orc", Emoji: "👹",
		Description: "Brutal em combate, com força descomunal.",
		BonusHP:     20, BonusMP: 0,
		BonusStr: 4, BonusDex: 0, BonusCon: 3, BonusInt: -1, BonusWis: 0, BonusCha: -1,
		Trait: "Fúria do Sangue: +25% de dano físico",
	},
	"goblin": {
		ID: "goblin", Name: "Goblin", Emoji: "👺",
		Description: "Pequenos e astutos, especialistas em emboscadas.",
		BonusHP:     5, BonusMP: 10,
		BonusStr: -1, BonusDex: 4, BonusCon: 0, BonusInt: 2, BonusWis: 0, BonusCha: 0,
		Trait: "Esperteza: +15% de ouro encontrado",
	},
	"qareen": {
		ID: "qareen", Name: "Qareen", Emoji: "🧞",
		Description: "Seres mágicos do deserto, ligados a forças elementais.",
		BonusHP:     10, BonusMP: 25,
		BonusStr: 0, BonusDex: 1, BonusCon: 0, BonusInt: 2, BonusWis: 2, BonusCha: 3,
		Trait: "Magia Elemental: +20% de dano mágico",
	},
	"minotaur": {
		ID: "minotaur", Name: "Minotauro", Emoji: "🐂",
		Description: "Guerreiros poderosos com chifres devastadores.",
		BonusHP:     35, BonusMP: 0,
		BonusStr: 5, BonusDex: -1, BonusCon: 4, BonusInt: -1, BonusWis: 0, BonusCha: -1,
		Trait: "Investida: +30% de dano no primeiro ataque do combate",
	},
}

// ── CLASSES ───────────────────────────────────────────────
var Classes = map[string]models.Class{
	"warrior": {
		ID: "warrior", Name: "Guerreiro", Emoji: "⚔️",
		Description: "Mestre das armas e armaduras pesadas.",
		BaseHP:      80, BaseMP: 20, HPPerLevel: 12, MPPerLevel: 3,
		BaseAttack: 12, BaseDefense: 10,
		PrimaryStats: []string{"strength", "constitution"}, Role: "tank",
	},
	"mage": {
		ID: "mage", Name: "Mago", Emoji: "🧙",
		Description: "Manipula as forças arcanas com maestria.",
		BaseHP:      45, BaseMP: 80, HPPerLevel: 5, MPPerLevel: 12,
		BaseAttack: 4, BaseDefense: 3,
		PrimaryStats: []string{"intelligence", "wisdom"}, Role: "mage",
	},
	"rogue": {
		ID: "rogue", Name: "Ladino", Emoji: "🗡️",
		Description: "Ágil e letal nas sombras.",
		BaseHP:      55, BaseMP: 40, HPPerLevel: 7, MPPerLevel: 6,
		BaseAttack: 10, BaseDefense: 5,
		PrimaryStats: []string{"dexterity", "charisma"}, Role: "dps",
	},
	"archer": {
		ID: "archer", Name: "Arqueiro", Emoji: "🏹",
		Description: "Precisão e velocidade a longa distância.",
		BaseHP:      60, BaseMP: 35, HPPerLevel: 8, MPPerLevel: 5,
		BaseAttack: 9, BaseDefense: 6,
		PrimaryStats: []string{"dexterity", "wisdom"}, Role: "ranged",
	},
	"paladin": {
		ID: "paladin", Name: "Paladino", Emoji: "🛡️",
		Description: "Guerreiro sagrado que combina força e magia divina.",
		BaseHP:      75, BaseMP: 45, HPPerLevel: 10, MPPerLevel: 7,
		BaseAttack: 10, BaseDefense: 9,
		PrimaryStats: []string{"strength", "charisma"}, Role: "tank",
	},
	"cleric": {
		ID: "cleric", Name: "Clérigo", Emoji: "✝️",
		Description: "Canal divino de cura e proteção.",
		BaseHP:      55, BaseMP: 70, HPPerLevel: 7, MPPerLevel: 10,
		BaseAttack: 6, BaseDefense: 7,
		PrimaryStats: []string{"wisdom", "constitution"}, Role: "support",
	},
	"barbarian": {
		ID: "barbarian", Name: "Bárbaro", Emoji: "🪓",
		Description: "Guerreiro selvagem movido por fúria incontrolável.",
		BaseHP:      95, BaseMP: 15, HPPerLevel: 14, MPPerLevel: 2,
		BaseAttack: 14, BaseDefense: 7,
		PrimaryStats: []string{"strength", "constitution"}, Role: "dps",
	},
	"bard": {
		ID: "bard", Name: "Bardo", Emoji: "🎵",
		Description: "Artista versátil que inspira aliados e confunde inimigos.",
		BaseHP:      50, BaseMP: 55, HPPerLevel: 6, MPPerLevel: 8,
		BaseAttack: 7, BaseDefense: 5,
		PrimaryStats: []string{"charisma", "dexterity"}, Role: "support",
	},
}

// ── SKILLS ────────────────────────────────────────────────
var Skills = map[string]models.Skill{

	// ══════════════════════════════════════════════════════
	// GUERREIRO — 3 ramos: Protetor | Berserker | Campeão
	//
	// Protetor: tank, CA alta, dano reduzido, sustain
	// Berserker: máximo dano, sem defesa, instável
	// Campeão: equilíbrio força + liderança
	//
	// Custo de pontos: T1=1  T2=1  T3=2  T4=3
	// Total possível a gastar: 19 pts (lv1→20)
	// Custo para completar 1 ramo: 1+1+2+3 = 7 pts
	// ══════════════════════════════════════════════════════

	// ── PROTETOR ──────────────────────────────────────────
	"w_iron_skin": {
		ID: "w_iron_skin", Class: "warrior", Branch: "protetor", Tier: 1, PointCost: 1,
		Name: "Pele de Ferro", Emoji: "🛡️", RequiredLevel: 1,
		MPCost: 0, Damage: 0, DamageType: "passive", Passive: true,
		Description: "+2 CA permanente. Sua pele endurece com o treinamento.",
	},
	"w_shield_bash": {
		ID: "w_shield_bash", Class: "warrior", Branch: "protetor", Tier: 2, PointCost: 1,
		Name: "Bate-Escudo", Emoji: "🪃", RequiredLevel: 5, Requires: "w_iron_skin",
		MPCost: 10, Damage: 12, DamageType: "physical",
		Description: "Golpe com o escudo. Dano físico + reduz CA do inimigo por 1 rodada.",
	},
	"w_fortress": {
		ID: "w_fortress", Class: "warrior", Branch: "protetor", Tier: 3, PointCost: 2,
		Name: "Postura de Fortaleza", Emoji: "🏰", RequiredLevel: 10, Requires: "w_shield_bash",
		MPCost: 20, Damage: 0, DamageType: "buff",
		Description: "+4 CA e -30% dano recebido por 3 rodadas. Você vira uma muralha.",
	},
	"w_divine_guard": {
		ID: "w_divine_guard", Class: "warrior", Branch: "protetor", Tier: 4, PointCost: 3,
		Name: "Guarda Divina", Emoji: "⛪", RequiredLevel: 16, Requires: "w_fortress",
		MPCost: 35, Damage: 0, DamageType: "buff",
		Description: "Por 2 rodadas, dano recebido reduzido a 1. Imunidade temporária.",
	},

	// ── BERSERKER ─────────────────────────────────────────
	"w_power_strike": {
		ID: "w_power_strike", Class: "warrior", Branch: "berserker", Tier: 1, PointCost: 1,
		Name: "Golpe Brutal", Emoji: "⚔️", RequiredLevel: 1,
		MPCost: 8, Damage: 18, DamageType: "physical",
		Description: "Golpe pesado com bônus de FOR. Base do caminho da destruição.",
	},
	"w_reckless": {
		ID: "w_reckless", Class: "warrior", Branch: "berserker", Tier: 2, PointCost: 1,
		Name: "Ataque Imprudente", Emoji: "💥", RequiredLevel: 5, Requires: "w_power_strike",
		MPCost: 12, Damage: 28, DamageType: "physical",
		Description: "+40% dano mas -2 CA até próximo turno. Alto risco, alta recompensa.",
	},
	"w_blood_rage": {
		ID: "w_blood_rage", Class: "warrior", Branch: "berserker", Tier: 3, PointCost: 2,
		Name: "Fúria Sangrenta", Emoji: "🩸", RequiredLevel: 10, Requires: "w_reckless",
		MPCost: 25, Damage: 0, DamageType: "buff",
		Description: "Quanto menos HP, mais forte: abaixo de 30% HP, +80% dano por 3 rodadas.",
	},
	"w_rampage": {
		ID: "w_rampage", Class: "warrior", Branch: "berserker", Tier: 4, PointCost: 3,
		Name: "Destruição Total", Emoji: "🌪️", RequiredLevel: 16, Requires: "w_blood_rage",
		MPCost: 50, Damage: 80, DamageType: "physical",
		Description: "Golpe devastador que ignora toda a defesa. Crítico em 18-20.",
	},

	// ── CAMPEÃO ───────────────────────────────────────────
	"w_battle_cry": {
		ID: "w_battle_cry", Class: "warrior", Branch: "campiao", Tier: 1, PointCost: 1,
		Name: "Grito de Guerra", Emoji: "📣", RequiredLevel: 1,
		MPCost: 10, Damage: 0, DamageType: "buff",
		Description: "+25% dano nos próximos 3 ataques. Você inspira terror.",
	},
	"w_cleave": {
		ID: "w_cleave", Class: "warrior", Branch: "campiao", Tier: 2, PointCost: 1,
		Name: "Corte em Arco", Emoji: "🌀", RequiredLevel: 6, Requires: "w_battle_cry",
		MPCost: 18, Damage: 22, DamageType: "physical",
		Description: "Golpe amplo e veloz. +15% chance de acerto crítico.",
	},
	"w_champion_aura": {
		ID: "w_champion_aura", Class: "warrior", Branch: "campiao", Tier: 3, PointCost: 2,
		Name: "Aura do Campeão", Emoji: "👑", RequiredLevel: 11, Requires: "w_cleave",
		MPCost: 20, Damage: 0, DamageType: "passive", Passive: true,
		Description: "Passiva: +3 no bônus de ataque d20 e +1 CA permanentes.",
	},
	"w_titan_blow": {
		ID: "w_titan_blow", Class: "warrior", Branch: "campiao", Tier: 4, PointCost: 3,
		Name: "Golpe do Titã", Emoji: "🏔️", RequiredLevel: 17, Requires: "w_champion_aura",
		MPCost: 45, Damage: 65, DamageType: "physical",
		Description: "O golpe mais poderoso do campeão. Ignora 50% da defesa.",
	},

	// ══════════════════════════════════════════════════════
	// MAGO — 3 ramos: Piromante | Crionita | Arcanista
	//
	// Piromante: DPS explosivo, dano em área
	// Crionita: controle + dano, reduz velocidade/CA
	// Arcanista: versátil, escudo mágico + burst
	// ══════════════════════════════════════════════════════

	// ── PIROMANTE ─────────────────────────────────────────
	"m_fireball": {
		ID: "m_fireball", Class: "mage", Branch: "piromante", Tier: 1, PointCost: 1,
		Name: "Bola de Fogo", Emoji: "🔥", RequiredLevel: 1,
		MPCost: 12, Damage: 20, DamageType: "magic",
		Description: "Projétil de fogo explosivo. Pilar do piromante.",
	},
	"m_flame_wave": {
		ID: "m_flame_wave", Class: "mage", Branch: "piromante", Tier: 2, PointCost: 1,
		Name: "Onda de Chamas", Emoji: "🌊", RequiredLevel: 5, Requires: "m_fireball",
		MPCost: 20, Damage: 30, DamageType: "magic",
		Description: "Onda de fogo que queima o inimigo por 2 rodadas (+5 dano/rodada).",
	},
	"m_fire_mastery": {
		ID: "m_fire_mastery", Class: "mage", Branch: "piromante", Tier: 3, PointCost: 2,
		Name: "Maestria do Fogo", Emoji: "♨️", RequiredLevel: 10, Requires: "m_flame_wave",
		MPCost: 0, Damage: 0, DamageType: "passive", Passive: true,
		Description: "Passiva: +25% dano em todas as habilidades de fogo.",
	},
	"m_meteor": {
		ID: "m_meteor", Class: "mage", Branch: "piromante", Tier: 4, PointCost: 3,
		Name: "Meteoro", Emoji: "☄️", RequiredLevel: 16, Requires: "m_fire_mastery",
		MPCost: 55, Damage: 90, DamageType: "magic",
		Description: "Invoca um meteoro devastador. Dano massivo. Crítico em 18-20.",
	},

	// ── CRIONITA ──────────────────────────────────────────
	"m_ice_shard": {
		ID: "m_ice_shard", Class: "mage", Branch: "crionita", Tier: 1, PointCost: 1,
		Name: "Fragmento de Gelo", Emoji: "❄️", RequiredLevel: 1,
		MPCost: 10, Damage: 15, DamageType: "magic",
		Description: "Cristal de gelo. Reduz CA do inimigo em 1 por 2 rodadas.",
	},
	"m_frost_nova": {
		ID: "m_frost_nova", Class: "mage", Branch: "crionita", Tier: 2, PointCost: 1,
		Name: "Nova Glacial", Emoji: "🌨️", RequiredLevel: 5, Requires: "m_ice_shard",
		MPCost: 18, Damage: 22, DamageType: "magic",
		Description: "Explosão de gelo. -3 CA do inimigo e -2 no bônus de ataque dele.",
	},
	"m_blizzard": {
		ID: "m_blizzard", Class: "mage", Branch: "crionita", Tier: 3, PointCost: 2,
		Name: "Nevasca", Emoji: "🌪️", RequiredLevel: 10, Requires: "m_frost_nova",
		MPCost: 30, Damage: 40, DamageType: "magic",
		Description: "Tempestade de gelo. Dano alto + inimigo perde 1 turno de ataque.",
	},
	"m_absolute_zero": {
		ID: "m_absolute_zero", Class: "mage", Branch: "crionita", Tier: 4, PointCost: 3,
		Name: "Zero Absoluto", Emoji: "🧊", RequiredLevel: 16, Requires: "m_blizzard",
		MPCost: 60, Damage: 75, DamageType: "magic",
		Description: "Congela o inimigo completamente. Dano massivo + CA -5 no próximo turno.",
	},

	// ── ARCANISTA ─────────────────────────────────────────
	"m_arcane_bolt": {
		ID: "m_arcane_bolt", Class: "mage", Branch: "arcanista", Tier: 1, PointCost: 1,
		Name: "Raio Arcano", Emoji: "⚡", RequiredLevel: 1,
		MPCost: 8, Damage: 16, DamageType: "magic",
		Description: "Projétil de energia pura. Sem resistência elemental.",
	},
	"m_arcane_shield": {
		ID: "m_arcane_shield", Class: "mage", Branch: "arcanista", Tier: 2, PointCost: 1,
		Name: "Escudo Arcano", Emoji: "🔮", RequiredLevel: 5, Requires: "m_arcane_bolt",
		MPCost: 15, Damage: 0, DamageType: "buff",
		Description: "+4 CA por 2 rodadas. Absorve o próximo ataque mágico.",
	},
	"m_chain_lightning": {
		ID: "m_chain_lightning", Class: "mage", Branch: "arcanista", Tier: 3, PointCost: 2,
		Name: "Raio em Cadeia", Emoji: "🌩️", RequiredLevel: 10, Requires: "m_arcane_shield",
		MPCost: 35, Damage: 50, DamageType: "magic",
		Description: "Relâmpago que salta e acerta com +5 no bônus de ataque.",
	},
	"m_arcane_burst": {
		ID: "m_arcane_burst", Class: "mage", Branch: "arcanista", Tier: 4, PointCost: 3,
		Name: "Explosão Arcana", Emoji: "💫", RequiredLevel: 16, Requires: "m_chain_lightning",
		MPCost: 50, Damage: 70, DamageType: "magic",
		Description: "Libera toda a energia acumulada. Dano dobrado se escudo estiver ativo.",
	},

	// ══════════════════════════════════════════════════════
	// LADINO — 3 ramos: Assassino | Envenenador | Sombra
	//
	// Assassino: burst de dano único, críticos
	// Envenenador: dano ao longo do tempo, debuffs
	// Sombra: esquiva, furtividade, utilitário
	// ══════════════════════════════════════════════════════

	// ── ASSASSINO ─────────────────────────────────────────
	"r_backstab": {
		ID: "r_backstab", Class: "rogue", Branch: "assassino", Tier: 1, PointCost: 1,
		Name: "Facada nas Costas", Emoji: "🗡️", RequiredLevel: 1,
		MPCost: 8, Damage: 22, DamageType: "physical",
		Description: "Ataque surpresa. +50% dano se inimigo não atacou no turno anterior.",
	},
	"r_vital_strike": {
		ID: "r_vital_strike", Class: "rogue", Branch: "assassino", Tier: 2, PointCost: 1,
		Name: "Golpe Vital", Emoji: "🎯", RequiredLevel: 5, Requires: "r_backstab",
		MPCost: 14, Damage: 30, DamageType: "physical",
		Description: "Mira um ponto vital. Crítico em 17-20 ao invés de só 20.",
	},
	"r_expose": {
		ID: "r_expose", Class: "rogue", Branch: "assassino", Tier: 3, PointCost: 2,
		Name: "Expor Fraqueza", Emoji: "🔍", RequiredLevel: 10, Requires: "r_vital_strike",
		MPCost: 18, Damage: 15, DamageType: "physical",
		Description: "Dano moderado + reduz CA do inimigo em 3 por 2 rodadas.",
	},
	"r_death_blow": {
		ID: "r_death_blow", Class: "rogue", Branch: "assassino", Tier: 4, PointCost: 3,
		Name: "Golpe Mortal", Emoji: "💀", RequiredLevel: 16, Requires: "r_expose",
		MPCost: 40, Damage: 70, DamageType: "physical",
		Description: "Golpe letal. Se inimigo tem menos de 25% HP, causa dano x3.",
	},

	// ── ENVENENADOR ───────────────────────────────────────
	"r_poison": {
		ID: "r_poison", Class: "rogue", Branch: "envenenador", Tier: 1, PointCost: 1,
		Name: "Veneno", Emoji: "☠️", RequiredLevel: 1,
		MPCost: 10, Damage: 14, DamageType: "poison",
		PoisonDmgPerTurn: 6, PoisonTurnsCount: 3,
		Description: "Causa dano imediato + veneno: 6/turno por 3 turnos.",
	},
	"r_acid_blade": {
		ID: "r_acid_blade", Class: "rogue", Branch: "envenenador", Tier: 2, PointCost: 1,
		Name: "Lâmina Ácida", Emoji: "🧪", RequiredLevel: 5, Requires: "r_poison",
		MPCost: 15, Damage: 20, DamageType: "poison",
		PoisonDmgPerTurn: 8, PoisonTurnsCount: 3,
		Description: "Causa dano imediato + veneno corrosivo: 8/turno por 3 turnos.",
	},
	"r_plague": {
		ID: "r_plague", Class: "rogue", Branch: "envenenador", Tier: 3, PointCost: 2,
		Name: "Praga", Emoji: "🦠", RequiredLevel: 10, Requires: "r_acid_blade",
		MPCost: 25, Damage: 18, DamageType: "poison",
		PoisonDmgPerTurn: 12, PoisonTurnsCount: 4,
		Description: "Veneno devastador: 12/turno por 4 turnos. Sobrescreve venenos anteriores.",
	},
	"r_toxic_cloud": {
		ID: "r_toxic_cloud", Class: "rogue", Branch: "envenenador", Tier: 4, PointCost: 3,
		Name: "Nuvem Tóxica", Emoji: "💚", RequiredLevel: 16, Requires: "r_plague",
		MPCost: 40, Damage: 35, DamageType: "poison",
		PoisonDmgPerTurn: 15, PoisonTurnsCount: 4,
		Description: "Nuvem venenosa: dano imediato pesado + 15/turno por 4 turnos.",
	},

	// ── SOMBRA ────────────────────────────────────────────
	"r_shadow_step": {
		ID: "r_shadow_step", Class: "rogue", Branch: "sombra", Tier: 1, PointCost: 1,
		Name: "Passo das Sombras", Emoji: "👤", RequiredLevel: 1,
		MPCost: 10, Damage: 0, DamageType: "buff",
		Description: "Desaparece nas sombras. Próximo ataque é crítico garantido.",
	},
	"r_smoke_bomb": {
		ID: "r_smoke_bomb", Class: "rogue", Branch: "sombra", Tier: 2, PointCost: 1,
		Name: "Bomba de Fumaça", Emoji: "💨", RequiredLevel: 5, Requires: "r_shadow_step",
		MPCost: 12, Damage: 0, DamageType: "buff",
		Description: "-4 no bônus de ataque do inimigo por 2 rodadas. Dificulta ser acertado.",
	},
	"r_evasion": {
		ID: "r_evasion", Class: "rogue", Branch: "sombra", Tier: 3, PointCost: 2,
		Name: "Evasão", Emoji: "🌫️", RequiredLevel: 10, Requires: "r_smoke_bomb",
		MPCost: 0, Damage: 0, DamageType: "passive", Passive: true,
		Description: "Passiva: +3 CA permanente. Sua agilidade natural dificulta ser acertado.",
	},
	"r_phantom": {
		ID: "r_phantom", Class: "rogue", Branch: "sombra", Tier: 4, PointCost: 3,
		Name: "Forma Fantasma", Emoji: "👻", RequiredLevel: 16, Requires: "r_evasion",
		MPCost: 35, Damage: 45, DamageType: "physical",
		Description: "Ataque imaterial: ignora CA do inimigo. Acerto automático garantido.",
	},

	// ══════════════════════════════════════════════════════
	// ARQUEIRO — 3 ramos: Atirador | Caçador | Arcano
	//
	// Atirador: precisão, dano alto por tiro único
	// Caçador: múltiplos projéteis, velocidade
	// Arcano: flechas mágicas, elemental
	// ══════════════════════════════════════════════════════

	// ── ATIRADOR ──────────────────────────────────────────
	"a_aimed_shot": {
		ID: "a_aimed_shot", Class: "archer", Branch: "atirador", Tier: 1, PointCost: 1,
		Name: "Tiro Preciso", Emoji: "🎯", RequiredLevel: 1,
		MPCost: 10, Damage: 20, DamageType: "physical",
		Description: "Disparo focado. +3 no bônus de ataque d20 neste tiro.",
	},
	"a_headshot": {
		ID: "a_headshot", Class: "archer", Branch: "atirador", Tier: 2, PointCost: 1,
		Name: "Tiro na Cabeça", Emoji: "🎖️", RequiredLevel: 5, Requires: "a_aimed_shot",
		MPCost: 15, Damage: 32, DamageType: "physical",
		Description: "Mira a cabeça. Crítico em 18-20. Dano máximo no dado.",
	},
	"a_eagle_eye": {
		ID: "a_eagle_eye", Class: "archer", Branch: "atirador", Tier: 3, PointCost: 2,
		Name: "Olho de Águia", Emoji: "🦅", RequiredLevel: 10, Requires: "a_headshot",
		MPCost: 0, Damage: 0, DamageType: "passive", Passive: true,
		Description: "Passiva: +2 no bônus de ataque d20 permanente. Visão sobrenatural.",
	},
	"a_deadeye": {
		ID: "a_deadeye", Class: "archer", Branch: "atirador", Tier: 4, PointCost: 3,
		Name: "Olho de Falcão", Emoji: "🌟", RequiredLevel: 16, Requires: "a_eagle_eye",
		MPCost: 30, Damage: 70, DamageType: "physical",
		Description: "Tiro perfeito. +6 no ataque d20 + dano dobrado se crítico.",
	},

	// ── CAÇADOR ───────────────────────────────────────────
	"a_quick_shot": {
		ID: "a_quick_shot", Class: "archer", Branch: "cacador", Tier: 1, PointCost: 1,
		Name: "Tiro Rápido", Emoji: "🏹", RequiredLevel: 1,
		MPCost: 6, Damage: 14, DamageType: "physical",
		Description: "Disparo veloz com 30% de chance de atacar uma segunda vez.",
	},
	"a_multishot": {
		ID: "a_multishot", Class: "archer", Branch: "cacador", Tier: 2, PointCost: 1,
		Name: "Chuva de Flechas", Emoji: "🌧️", RequiredLevel: 5, Requires: "a_quick_shot",
		MPCost: 18, Damage: 18, DamageType: "physical",
		Description: "3 flechas consecutivas. Cada uma faz rolagem d20 independente.",
	},
	"a_tracking": {
		ID: "a_tracking", Class: "archer", Branch: "cacador", Tier: 3, PointCost: 2,
		Name: "Rastreamento", Emoji: "🐾", RequiredLevel: 10, Requires: "a_multishot",
		MPCost: 0, Damage: 0, DamageType: "passive", Passive: true,
		Description: "Passiva: +20% dano contra monstros com HP abaixo de 50%.",
	},
	"a_volley": {
		ID: "a_volley", Class: "archer", Branch: "cacador", Tier: 4, PointCost: 3,
		Name: "Saraivada", Emoji: "⛈️", RequiredLevel: 16, Requires: "a_tracking",
		MPCost: 45, Damage: 55, DamageType: "physical",
		Description: "5 flechas de uma vez. Cada acerto aplica dano completo.",
	},

	// ── ARCANO ────────────────────────────────────────────
	"a_magic_arrow": {
		ID: "a_magic_arrow", Class: "archer", Branch: "arcano", Tier: 1, PointCost: 1,
		Name: "Flecha Mágica", Emoji: "✨", RequiredLevel: 1,
		MPCost: 10, Damage: 18, DamageType: "magic",
		Description: "Flecha imbuída de energia arcana. Ignora resistência física.",
	},
	"a_frost_arrow": {
		ID: "a_frost_arrow", Class: "archer", Branch: "arcano", Tier: 2, PointCost: 1,
		Name: "Flecha de Gelo", Emoji: "🧊", RequiredLevel: 5, Requires: "a_magic_arrow",
		MPCost: 15, Damage: 24, DamageType: "magic",
		Description: "Flecha glacial. -2 CA do inimigo por 2 rodadas.",
	},
	"a_arcane_quiver": {
		ID: "a_arcane_quiver", Class: "archer", Branch: "arcano", Tier: 3, PointCost: 2,
		Name: "Aljava Arcana", Emoji: "🪄", RequiredLevel: 10, Requires: "a_frost_arrow",
		MPCost: 0, Damage: 0, DamageType: "passive", Passive: true,
		Description: "Passiva: +20% dano mágico em todas as flechas encantadas.",
	},
	"a_dragon_arrow": {
		ID: "a_dragon_arrow", Class: "archer", Branch: "arcano", Tier: 4, PointCost: 3,
		Name: "Flecha do Dragão", Emoji: "🐉", RequiredLevel: 16, Requires: "a_arcane_quiver",
		MPCost: 50, Damage: 80, DamageType: "magic",
		Description: "Flecha lendária de fogo dracônico. Maior dano mágico do arqueiro.",
	},
}

// ── ITEMS ─────────────────────────────────────────────────
// Rarity: Common=0, Uncommon=1, Rare=2, Epic=3, Legendary=4
var Items = map[string]models.Item{

	// ── CONSUMABLES ────────────────────────────────────────
	"potion_small": {
		ID: "potion_small", Name: "Poção Pequena", Emoji: "🧪", Type: "consumable", Rarity: models.RarityCommon,
		Description: "Restaura 30 HP.", Price: 15, SellPrice: 7, HealHP: 30, MinLevel: 1, DropWeight: 50,
	},
	"potion_medium": {
		ID: "potion_medium", Name: "Poção Média", Emoji: "🧪", Type: "consumable", Rarity: models.RarityCommon,
		Description: "Restaura 80 HP.", Price: 35, SellPrice: 15, HealHP: 80, MinLevel: 5, DropWeight: 30,
	},
	"potion_large": {
		ID: "potion_large", Name: "Poção Grande", Emoji: "⚗️", Type: "consumable", Rarity: models.RarityUncommon,
		Description: "Restaura 200 HP.", Price: 80, SellPrice: 35, HealHP: 200, MinLevel: 10, DropWeight: 15,
	},
	"potion_supreme": {
		ID: "potion_supreme", Name: "Poção Suprema", Emoji: "🌟", Type: "consumable", Rarity: models.RarityRare,
		Description: "Restaura 500 HP.", Price: 200, SellPrice: 80, HealHP: 500, MinLevel: 15, DropWeight: 5,
	},
	"mana_potion_small": {
		ID: "mana_potion_small", Name: "Poção de Mana P.", Emoji: "💧", Type: "consumable", Rarity: models.RarityCommon,
		Description: "Restaura 20 MP.", Price: 20, SellPrice: 8, HealMP: 20, MinLevel: 1, DropWeight: 40,
	},
	"mana_potion_large": {
		ID: "mana_potion_large", Name: "Poção de Mana G.", Emoji: "💙", Type: "consumable", Rarity: models.RarityUncommon,
		Description: "Restaura 60 MP.", Price: 60, SellPrice: 25, HealMP: 60, MinLevel: 6, DropWeight: 20,
	},
	"mana_potion_supreme": {
		ID: "mana_potion_supreme", Name: "Elixir de Mana", Emoji: "🌀", Type: "consumable", Rarity: models.RarityRare,
		Description: "Restaura 150 MP.", Price: 150, SellPrice: 60, HealMP: 150, MinLevel: 12, DropWeight: 8,
	},
	"elixir": {
		ID: "elixir", Name: "Elixir Completo", Emoji: "✨", Type: "consumable", Rarity: models.RarityUncommon,
		Description: "Restaura 150 HP e 50 MP.", Price: 120, SellPrice: 50, HealHP: 150, HealMP: 50, MinLevel: 10, DropWeight: 15,
	},
	"elixir_divine": {
		ID: "elixir_divine", Name: "Elixir Divino", Emoji: "💫", Type: "consumable", Rarity: models.RarityEpic,
		Description: "Restaura 600 HP e 200 MP.", Price: 0, SellPrice: 300, HealHP: 600, HealMP: 200, MinLevel: 15, DropWeight: 3,
	},
	"energy_drink": {
		ID: "energy_drink", Name: "Bebida Energética", Emoji: "⚡", Type: "consumable", Rarity: models.RarityCommon,
		Description: "Restaura 10 de Energia.", Price: 25, SellPrice: 10, RestoreEnergy: 10, MinLevel: 1, DropWeight: 30,
	},
	"energy_potion": {
		ID: "energy_potion", Name: "Poção de Energia", Emoji: "🔋", Type: "consumable", Rarity: models.RarityUncommon,
		Description: "Restaura 30 de Energia.", Price: 70, SellPrice: 28, RestoreEnergy: 30, MinLevel: 5, DropWeight: 15,
	},
	"energy_elixir": {
		ID: "energy_elixir", Name: "Elixir de Energia", Emoji: "💡", Type: "consumable", Rarity: models.RarityRare,
		Description: "Restaura 60 Energia e 50 HP.", Price: 150, SellPrice: 60, HealHP: 50, RestoreEnergy: 60, MinLevel: 10, DropWeight: 8, DiamondPrice: 5,
	},
	"antidote": {
		ID: "antidote", Name: "Antídoto", Emoji: "💊", Type: "consumable", Rarity: models.RarityCommon,
		Description: "Cura envenenamento.", Price: 30, SellPrice: 12, MinLevel: 1, DropWeight: 20, CurePoison: true,
	},
	"revive_token": {
		ID: "revive_token", Name: "Token de Reviver", Emoji: "🔮", Type: "consumable", Rarity: models.RarityEpic,
		Description: "Revive com HP/MP cheio ao morrer.", Price: 0, SellPrice: 150, MinLevel: 1, DropWeight: 1, DiamondPrice: 20,
	},
	"xp_boost": {
		ID: "xp_boost", Name: "Bênção do Sábio", Emoji: "📖", Type: "consumable", Rarity: models.RarityRare,
		Description: "+50% de XP por 30 minutos. Use antes de lutar!", Price: 0, SellPrice: 30, MinLevel: 1, DropWeight: 0, DiamondPrice: 25, XPBoostMinutes: 30,
	},
	"skill_reset": {
		ID: "skill_reset", Name: "Tomo do Esquecimento", Emoji: "📜", Type: "consumable", Rarity: models.RarityEpic,
		Description: "Apaga todas as habilidades aprendidas e devolve 100% dos pontos gastos.", Price: 0, SellPrice: 0, MinLevel: 1, DropWeight: 0, DiamondPrice: 50,
	},

	// ── CHESTS ──────────────────────────────────────────────
	"chest_wooden": {
		ID: "chest_wooden", Name: "Baú de Madeira", Emoji: "📦", Type: "chest", Rarity: models.RarityCommon,
		Description: "Contém 20-80 moedas de ouro.", Price: 0, SellPrice: 0, MinLevel: 1, DropWeight: 20,
	},
	"chest_iron": {
		ID: "chest_iron", Name: "Baú de Ferro", Emoji: "🗃️", Type: "chest", Rarity: models.RarityUncommon,
		Description: "Contém 80-200 moedas + item aleatório.", Price: 0, SellPrice: 0, MinLevel: 5, DropWeight: 10,
	},
	"chest_gold": {
		ID: "chest_gold", Name: "Baú Dourado", Emoji: "💰", Type: "chest", Rarity: models.RarityRare,
		Description: "Contém 200-500 moedas + item raro.", Price: 0, SellPrice: 0, MinLevel: 10, DropWeight: 5,
	},
	"chest_dragon": {
		ID: "chest_dragon", Name: "Baú do Dragão", Emoji: "🐲", Type: "chest", Rarity: models.RarityLegendary,
		Description: "Contém 500-1500 moedas + item épico/lendário.", Price: 0, SellPrice: 0, MinLevel: 15, DropWeight: 1,
	},

	// ── WEAPONS: WARRIOR ────────────────────────────────────
	"sword_iron": {
		ID: "sword_iron", Name: "Espada de Ferro", Emoji: "⚔️", Type: "weapon", Rarity: models.RarityCommon,
		Description: "Espada básica forjada em ferro.", Price: 80, SellPrice: 30, AttackBonus: 8, MinLevel: 1, ClassReq: "warrior", DropWeight: 20, Slot: "weapon", HitBonus: 1,
	},
	"sword_steel": {
		ID: "sword_steel", Name: "Espada de Aço", Emoji: "🗡️", Type: "weapon", Rarity: models.RarityUncommon,
		Description: "Espada afiada de alta qualidade.", Price: 250, SellPrice: 100, AttackBonus: 18, MinLevel: 5, ClassReq: "warrior", DropWeight: 12, Slot: "weapon", HitBonus: 1,
	},
	"sword_silver": {
		ID: "sword_silver", Name: "Espada de Prata", Emoji: "✴️", Type: "weapon", Rarity: models.RarityRare,
		Description: "Eficaz contra mortos-vivos.", Price: 600, SellPrice: 250, AttackBonus: 30, MinLevel: 10, ClassReq: "warrior", DropWeight: 7, Slot: "weapon",
	},
	"sword_darksteel": {
		ID: "sword_darksteel", Name: "Espada do Abismo", Emoji: "🌑", Type: "weapon", Rarity: models.RarityEpic,
		Description: "Forjada nas profundezas do abismo.", Price: 1500, SellPrice: 600, AttackBonus: 45, MinLevel: 14, ClassReq: "warrior", DropWeight: 3, Slot: "weapon",
	},
	"sword_dragonslayer": {
		ID: "sword_dragonslayer", Name: "Matador de Dragões", Emoji: "🔥", Type: "weapon", Rarity: models.RarityLegendary,
		Description: "A lâmina mais poderosa já forjada.", Price: 0, SellPrice: 2000, AttackBonus: 65, MinLevel: 18, ClassReq: "warrior", DropWeight: 1, Slot: "weapon", HitBonus: 3,
	},

	// ── WEAPONS: MAGE ───────────────────────────────────────
	"staff_oak": {
		ID: "staff_oak", Name: "Cajado de Carvalho", Emoji: "🪄", Type: "weapon", Rarity: models.RarityCommon,
		Description: "Cajado básico que amplifica magias.", Price: 70, SellPrice: 28, MagicAtkBonus: 10, MinLevel: 1, ClassReq: "mage", DropWeight: 20, Slot: "weapon", HitBonus: 1,
	},
	"staff_arcane": {
		ID: "staff_arcane", Name: "Cajado Arcano", Emoji: "🔮", Type: "weapon", Rarity: models.RarityUncommon,
		Description: "Imbuído com energia arcana.", Price: 280, SellPrice: 110, MagicAtkBonus: 22, MinLevel: 6, ClassReq: "mage", DropWeight: 12, Slot: "weapon",
	},
	"staff_crystal": {
		ID: "staff_crystal", Name: "Cajado Cristalino", Emoji: "💎", Type: "weapon", Rarity: models.RarityRare,
		Description: "Cristal que amplifica magia elementar.", Price: 700, SellPrice: 280, MagicAtkBonus: 35, MinLevel: 10, ClassReq: "mage", DropWeight: 7, Slot: "weapon", HitBonus: 2,
	},
	"staff_void": {
		ID: "staff_void", Name: "Cajado do Vazio", Emoji: "🌌", Type: "weapon", Rarity: models.RarityEpic,
		Description: "Canaliza energia do vazio dimensional.", Price: 1800, SellPrice: 720, MagicAtkBonus: 52, MinLevel: 14, ClassReq: "mage", DropWeight: 3, Slot: "weapon",
	},
	"staff_dragon": {
		ID: "staff_dragon", Name: "Cajado do Dragão", Emoji: "🐲", Type: "weapon", Rarity: models.RarityLegendary,
		Description: "Feito de osso de dragão ancião.", Price: 0, SellPrice: 2500, MagicAtkBonus: 75, MinLevel: 18, ClassReq: "mage", DropWeight: 1, Slot: "weapon", HitBonus: 3,
	},

	// ── WEAPONS: ROGUE ──────────────────────────────────────
	"dagger_iron": {
		ID: "dagger_iron", Name: "Adaga de Ferro", Emoji: "🗡️", Type: "weapon", Rarity: models.RarityCommon,
		Description: "Adaga leve para golpes rápidos.", Price: 60, SellPrice: 24, AttackBonus: 7, MinLevel: 1, ClassReq: "rogue", DropWeight: 20, Slot: "weapon", HitBonus: 1,
	},
	"dagger_venom": {
		ID: "dagger_venom", Name: "Adaga Venenosa", Emoji: "☠️", Type: "weapon", Rarity: models.RarityUncommon,
		Description: "Envenena o alvo a cada golpe.", Price: 220, SellPrice: 88, AttackBonus: 16, MinLevel: 5, ClassReq: "rogue", DropWeight: 12, Slot: "weapon",
	},
	"dagger_shadow": {
		ID: "dagger_shadow", Name: "Adaga das Sombras", Emoji: "🌑", Type: "weapon", Rarity: models.RarityRare,
		Description: "Drena vida do alvo.", Price: 600, SellPrice: 240, AttackBonus: 28, MinLevel: 10, ClassReq: "rogue", DropWeight: 7, Slot: "weapon", HitBonus: 2,
	},
	"dagger_assassin": {
		ID: "dagger_assassin", Name: "Adaga do Assassino", Emoji: "💀", Type: "weapon", Rarity: models.RarityEpic,
		Description: "Ignora 30% da defesa do oponente.", Price: 1400, SellPrice: 560, AttackBonus: 42, MinLevel: 14, ClassReq: "rogue", DropWeight: 3, Slot: "weapon",
	},
	"dagger_void": {
		ID: "dagger_void", Name: "Adaga do Vazio", Emoji: "🌌", Type: "weapon", Rarity: models.RarityLegendary,
		Description: "Toca a própria alma do inimigo.", Price: 0, SellPrice: 2200, AttackBonus: 60, MinLevel: 18, ClassReq: "rogue", DropWeight: 1, Slot: "weapon",
	},

	// ── WEAPONS: ARCHER ─────────────────────────────────────
	"bow_short": {
		ID: "bow_short", Name: "Arco Curto", Emoji: "🏹", Type: "weapon", Rarity: models.RarityCommon,
		Description: "Arco básico para iniciantes.", Price: 65, SellPrice: 26, AttackBonus: 7, MinLevel: 1, ClassReq: "archer", DropWeight: 20, Slot: "weapon", HitBonus: 1,
	},
	"bow_long": {
		ID: "bow_long", Name: "Arco Longo", Emoji: "🏹", Type: "weapon", Rarity: models.RarityUncommon,
		Description: "Maior alcance e poder.", Price: 220, SellPrice: 88, AttackBonus: 17, MinLevel: 5, ClassReq: "archer", DropWeight: 12, Slot: "weapon", HitBonus: 1,
	},
	"bow_elven": {
		ID: "bow_elven", Name: "Arco Élfico", Emoji: "🌿", Type: "weapon", Rarity: models.RarityRare,
		Description: "Forjado pelos elfos com madeira sagrada.", Price: 700, SellPrice: 280, AttackBonus: 32, MinLevel: 12, ClassReq: "archer", DropWeight: 7, Slot: "weapon", HitBonus: 2,
	},
	"bow_storm": {
		ID: "bow_storm", Name: "Arco da Tempestade", Emoji: "⛈️", Type: "weapon", Rarity: models.RarityEpic,
		Description: "Dispara flechas imbuídas com raios.", Price: 1600, SellPrice: 640, AttackBonus: 48, MinLevel: 15, ClassReq: "archer", DropWeight: 3, Slot: "weapon", HitBonus: 2,
	},
	"bow_dragon": {
		ID: "bow_dragon", Name: "Arco do Dragão", Emoji: "🐉", Type: "weapon", Rarity: models.RarityLegendary,
		Description: "Flechas de energia de dragão.", Price: 0, SellPrice: 2300, AttackBonus: 68, MinLevel: 18, ClassReq: "archer", DropWeight: 1, Slot: "weapon",
	},

	// ── ARMORS ──────────────────────────────────────────────
	"leather_armor": {
		ID: "leather_armor", Name: "Armadura de Couro", Emoji: "🥋", Type: "armor", Rarity: models.RarityCommon,
		Description: "Proteção básica de couro curtido.", Price: 50, SellPrice: 20, DefenseBonus: 5, MinLevel: 1, DropWeight: 20, Slot: "chest",
	},
	"chain_mail": {
		ID: "chain_mail", Name: "Cota de Malha", Emoji: "🛡️", Type: "armor", Rarity: models.RarityUncommon,
		Description: "Proteção média de anéis de metal.", Price: 180, SellPrice: 72, DefenseBonus: 12, MinLevel: 4, DropWeight: 12, Slot: "chest",
	},
	"plate_armor": {
		ID: "plate_armor", Name: "Armadura de Placas", Emoji: "⚙️", Type: "armor", Rarity: models.RarityRare,
		Description: "Armadura pesada de aço.", Price: 500, SellPrice: 200, DefenseBonus: 22, MinLevel: 8, ClassReq: "warrior", DropWeight: 7, Slot: "chest",
	},
	"mage_robe": {
		ID: "mage_robe", Name: "Manto do Mago", Emoji: "👘", Type: "armor", Rarity: models.RarityUncommon,
		Description: "Manto com proteção mágica.", Price: 160, SellPrice: 64, DefenseBonus: 4, MagicDefBonus: 15, MinLevel: 3, ClassReq: "mage", DropWeight: 12, Slot: "chest",
	},
	"shadow_cloak": {
		ID: "shadow_cloak", Name: "Manto das Sombras", Emoji: "🌑", Type: "armor", Rarity: models.RarityRare,
		Description: "Aumenta evasão do portador.", Price: 350, SellPrice: 140, DefenseBonus: 10, SpeedBonus: 3, MinLevel: 6, ClassReq: "rogue", DropWeight: 7, Slot: "chest",
	},
	"ranger_vest": {
		ID: "ranger_vest", Name: "Colete do Ranger", Emoji: "🧥", Type: "armor", Rarity: models.RarityUncommon,
		Description: "Leve e resistente para arqueiros.", Price: 220, SellPrice: 88, DefenseBonus: 8, SpeedBonus: 2, MinLevel: 4, ClassReq: "archer", DropWeight: 12, Slot: "chest", CABonus: 2,
	},
	"mystic_robe": {
		ID: "mystic_robe", Name: "Manto Místico", Emoji: "🔮", Type: "armor", Rarity: models.RarityRare,
		Description: "Amplia poder mágico e resistência.", Price: 650, SellPrice: 260, DefenseBonus: 6, MagicDefBonus: 25, MinLevel: 10, ClassReq: "mage", DropWeight: 7, Slot: "chest",
	},
	"dark_plate": {
		ID: "dark_plate", Name: "Armadura Sombria", Emoji: "🖤", Type: "armor", Rarity: models.RarityEpic,
		Description: "Forjada com metal das profundezas.", Price: 1400, SellPrice: 560, DefenseBonus: 35, MinLevel: 13, ClassReq: "warrior", DropWeight: 3, Slot: "chest",
	},
	"void_cloak": {
		ID: "void_cloak", Name: "Manto do Vazio", Emoji: "🌌", Type: "armor", Rarity: models.RarityEpic,
		Description: "Torna o usuário semitransparente.", Price: 1300, SellPrice: 520, DefenseBonus: 18, SpeedBonus: 5, MinLevel: 13, ClassReq: "rogue", DropWeight: 3, Slot: "chest",
	},
	"storm_robe": {
		ID: "storm_robe", Name: "Manto da Tempestade", Emoji: "⛈️", Type: "armor", Rarity: models.RarityEpic,
		Description: "Canaliza energia elétrica.", Price: 1350, SellPrice: 540, DefenseBonus: 8, MagicDefBonus: 35, MinLevel: 13, ClassReq: "mage", DropWeight: 3, Slot: "chest", CABonus: 3,
	},
	"dragon_armor": {
		ID: "dragon_armor", Name: "Armadura de Dragão", Emoji: "🐉", Type: "armor", Rarity: models.RarityLegendary,
		Description: "Escamas de dragão ancião. Proteção máxima.", Price: 0, SellPrice: 3000, DefenseBonus: 50, MagicDefBonus: 30, MinLevel: 17, DropWeight: 1, Slot: "chest", CABonus: 7,
	},
	"arcane_mantle": {
		ID: "arcane_mantle", Name: "Manto Arcano", Emoji: "💫", Type: "armor", Rarity: models.RarityLegendary,
		Description: "Tecido com linhas de força mágica puras.", Price: 0, SellPrice: 3000, DefenseBonus: 15, MagicDefBonus: 50, SpeedBonus: 8, MinLevel: 17, ClassReq: "mage", DropWeight: 1, Slot: "chest", CABonus: 4,
	},
	"phantom_cloak": {
		ID: "phantom_cloak", Name: "Manto Fantasma", Emoji: "👻", Type: "armor", Rarity: models.RarityLegendary,
		Description: "Permite mover-se como uma sombra.", Price: 0, SellPrice: 3000, DefenseBonus: 25, SpeedBonus: 15, MinLevel: 17, ClassReq: "rogue", DropWeight: 1, Slot: "chest", CABonus: 5,
	},

	// ── NOVAS ARMAS — GUERREIRO ─────────────────────────────
	"axe_iron": {
		ID: "axe_iron", Name: "Machado de Ferro", Emoji: "🪓", Type: "weapon", Rarity: models.RarityCommon,
		Description: "Machado pesado com golpes devastadores.", Price: 90, SellPrice: 35, AttackBonus: 10, MinLevel: 2, ClassReq: "warrior", DropWeight: 25, Slot: "weapon",
	},
	"mace_spiked": {
		ID: "mace_spiked", Name: "Maça Cravejada", Emoji: "🔨", Type: "weapon", Rarity: models.RarityUncommon,
		Description: "Maça com pontas de aço que perfuram armaduras.", Price: 200, SellPrice: 80, AttackBonus: 15, MinLevel: 4, ClassReq: "warrior", DropWeight: 18, Slot: "weapon",
	},
	"halberd": {
		ID: "halberd", Name: "Alabarda", Emoji: "⚔️", Type: "weapon", Rarity: models.RarityUncommon,
		Description: "Arma de haste com lâmina e gancho.", Price: 320, SellPrice: 128, AttackBonus: 22, MinLevel: 7, ClassReq: "warrior", DropWeight: 15, Slot: "weapon", HitBonus: 2,
	},
	"greatsword": {
		ID: "greatsword", Name: "Espadão dos Campeões", Emoji: "🗡️", Type: "weapon", Rarity: models.RarityRare,
		Description: "Espada de dois gumes usada pelos campeões da arena.", Price: 750, SellPrice: 300, AttackBonus: 34, MinLevel: 11, ClassReq: "warrior", DropWeight: 10, Slot: "weapon",
	},
	"war_hammer": {
		ID: "war_hammer", Name: "Martelo de Guerra", Emoji: "🔨", Type: "weapon", Rarity: models.RarityEpic,
		Description: "Esmaga escudos e ossos com igual facilidade.", Price: 1600, SellPrice: 640, AttackBonus: 48, MinLevel: 15, ClassReq: "warrior", DropWeight: 5, Slot: "weapon", HitBonus: 2,
	},

	// ── NOVAS ARMAS — MAGO ──────────────────────────────────
	"wand_basic": {
		ID: "wand_basic", Name: "Varinha Iniciante", Emoji: "✨", Type: "weapon", Rarity: models.RarityCommon,
		Description: "Varinha simples para iniciantes na magia.", Price: 55, SellPrice: 22, MagicAtkBonus: 8, MinLevel: 1, ClassReq: "mage", DropWeight: 25, Slot: "weapon", HitBonus: 1,
	},
	"tome_fire": {
		ID: "tome_fire", Name: "Grimório do Fogo", Emoji: "📕", Type: "weapon", Rarity: models.RarityUncommon,
		Description: "Contém feitiços de fogo elementar.", Price: 240, SellPrice: 96, MagicAtkBonus: 20, MinLevel: 5, ClassReq: "mage", DropWeight: 18, Slot: "weapon",
	},
	"orb_lightning": {
		ID: "orb_lightning", Name: "Orbe de Relâmpago", Emoji: "⚡", Type: "weapon", Rarity: models.RarityRare,
		Description: "Esfera que amplifica magia elétrica.", Price: 680, SellPrice: 272, MagicAtkBonus: 32, MinLevel: 9, ClassReq: "mage", DropWeight: 10, Slot: "weapon",
	},
	"staff_inferno": {
		ID: "staff_inferno", Name: "Cajado do Inferno", Emoji: "🔥", Type: "weapon", Rarity: models.RarityEpic,
		Description: "Forjado nas chamas do submundo.", Price: 1700, SellPrice: 680, MagicAtkBonus: 55, MinLevel: 15, ClassReq: "mage", DropWeight: 5, Slot: "weapon",
	},
	"tome_ancient": {
		ID: "tome_ancient", Name: "Grimório Antigo", Emoji: "📖", Type: "weapon", Rarity: models.RarityRare,
		Description: "Grimório com segredos de magia ancestral.", Price: 800, SellPrice: 320, MagicAtkBonus: 38, MinLevel: 12, ClassReq: "mage", DropWeight: 8, Slot: "weapon",
	},

	// ── NOVAS ARMAS — LADINO ────────────────────────────────
	"knife_bone": {
		ID: "knife_bone", Name: "Faca de Osso", Emoji: "🦴", Type: "weapon", Rarity: models.RarityCommon,
		Description: "Faca improvisada, mas letal nas mãos certas.", Price: 45, SellPrice: 18, AttackBonus: 6, MinLevel: 1, ClassReq: "rogue", DropWeight: 25, Slot: "weapon",
	},
	"dagger_curved": {
		ID: "dagger_curved", Name: "Adaga Curva", Emoji: "🗡️", Type: "weapon", Rarity: models.RarityUncommon,
		Description: "Lâmina curva que causa ferimentos profundos.", Price: 190, SellPrice: 76, AttackBonus: 14, MinLevel: 4, ClassReq: "rogue", DropWeight: 18, Slot: "weapon",
	},
	"blade_dual": {
		ID: "blade_dual", Name: "Lâminas Gêmeas", Emoji: "⚔️", Type: "weapon", Rarity: models.RarityRare,
		Description: "Par de lâminas para ataques em combinação.", Price: 580, SellPrice: 232, AttackBonus: 26, MinLevel: 8, ClassReq: "rogue", DropWeight: 10, Slot: "weapon",
	},
	"stiletto": {
		ID: "stiletto", Name: "Estilete do Eclipse", Emoji: "🌒", Type: "weapon", Rarity: models.RarityEpic,
		Description: "Lâmina fina que penetra qualquer armadura.", Price: 1350, SellPrice: 540, AttackBonus: 40, MinLevel: 13, ClassReq: "rogue", DropWeight: 5, Slot: "weapon",
	},
	"serpent_fang": {
		ID: "serpent_fang", Name: "Presa de Serpente", Emoji: "🐍", Type: "weapon", Rarity: models.RarityRare,
		Description: "Arma revestida com veneno de serpente rara.", Price: 650, SellPrice: 260, AttackBonus: 30, MinLevel: 11, ClassReq: "rogue", DropWeight: 8, Slot: "weapon",
	},

	// ── NOVAS ARMAS — ARQUEIRO ──────────────────────────────
	"sling": {
		ID: "sling", Name: "Funda de Couro", Emoji: "🪃", Type: "weapon", Rarity: models.RarityCommon,
		Description: "Arma simples de longa distância.", Price: 30, SellPrice: 12, AttackBonus: 5, MinLevel: 1, ClassReq: "archer", DropWeight: 25, Slot: "weapon",
	},
	"crossbow": {
		ID: "crossbow", Name: "Besta", Emoji: "🏹", Type: "weapon", Rarity: models.RarityUncommon,
		Description: "Besta compacta com tiro poderoso.", Price: 200, SellPrice: 80, AttackBonus: 15, MinLevel: 4, ClassReq: "archer", DropWeight: 18, Slot: "weapon", HitBonus: 1,
	},
	"bow_hunter": {
		ID: "bow_hunter", Name: "Arco do Caçador", Emoji: "🌿", Type: "weapon", Rarity: models.RarityRare,
		Description: "Arco reforçado preferido por caçadores veteranos.", Price: 580, SellPrice: 232, AttackBonus: 28, MinLevel: 8, ClassReq: "archer", DropWeight: 10, Slot: "weapon",
	},
	"bow_shadow": {
		ID: "bow_shadow", Name: "Arco das Sombras", Emoji: "🌑", Type: "weapon", Rarity: models.RarityEpic,
		Description: "Dispara flechas que desaparecem nas sombras.", Price: 1500, SellPrice: 600, AttackBonus: 45, MinLevel: 13, ClassReq: "archer", DropWeight: 5, Slot: "weapon",
	},
	"bow_ancient": {
		ID: "bow_ancient", Name: "Arco Ancestral", Emoji: "🏺", Type: "weapon", Rarity: models.RarityRare,
		Description: "Arco élfico antigo com entalhes de poder.", Price: 720, SellPrice: 288, AttackBonus: 34, MinLevel: 11, ClassReq: "archer", DropWeight: 8, Slot: "weapon",
	},

	// ── NOVAS ARMADURAS — GUERREIRO ─────────────────────────
	"iron_shield": {
		ID: "iron_shield", Name: "Escudo de Ferro", Emoji: "🛡️", Type: "armor", Rarity: models.RarityCommon,
		Description: "Escudo básico que bloqueia golpes inimigos.", Price: 60, SellPrice: 24, DefenseBonus: 6, MinLevel: 1, ClassReq: "warrior", DropWeight: 25, Slot: "offhand",
	},
	"scale_armor": {
		ID: "scale_armor", Name: "Armadura de Escamas", Emoji: "🐊", Type: "armor", Rarity: models.RarityUncommon,
		Description: "Feita de escamas de réptil resistente.", Price: 220, SellPrice: 88, DefenseBonus: 14, MinLevel: 5, ClassReq: "warrior", DropWeight: 15, Slot: "chest", CABonus: 2,
	},
	"knights_armor": {
		ID: "knights_armor", Name: "Armadura do Cavaleiro", Emoji: "⚔️", Type: "armor", Rarity: models.RarityRare,
		Description: "Armadura completa usada pelos cavaleiros do reino.", Price: 650, SellPrice: 260, DefenseBonus: 26, MinLevel: 9, ClassReq: "warrior", DropWeight: 8, Slot: "chest", CABonus: 4,
	},
	"fortress_plate": {
		ID: "fortress_plate", Name: "Placa da Fortaleza", Emoji: "🏰", Type: "armor", Rarity: models.RarityEpic,
		Description: "Armadura pesada que transforma o guerreiro em uma fortaleza.", Price: 1600, SellPrice: 640, DefenseBonus: 40, MinLevel: 14, ClassReq: "warrior", DropWeight: 4, Slot: "chest",
	},
	"titan_armor": {
		ID: "titan_armor", Name: "Armadura do Titã", Emoji: "🗿", Type: "armor", Rarity: models.RarityLegendary,
		Description: "Lenda viva da forja divina. Quase indestrutível.", Price: 0, SellPrice: 3500, DefenseBonus: 55, MinLevel: 18, ClassReq: "warrior", DropWeight: 1, Slot: "chest", CABonus: 6,
	},

	// ── NOVAS ARMADURAS — MAGO ──────────────────────────────
	"apprentice_robe": {
		ID: "apprentice_robe", Name: "Manto de Aprendiz", Emoji: "👕", Type: "armor", Rarity: models.RarityCommon,
		Description: "Manto básico com leve proteção mágica.", Price: 40, SellPrice: 16, DefenseBonus: 2, MagicDefBonus: 8, MinLevel: 1, ClassReq: "mage", DropWeight: 25, Slot: "chest", CABonus: 1,
	},
	"enchanted_robe": {
		ID: "enchanted_robe", Name: "Manto Encantado", Emoji: "🌀", Type: "armor", Rarity: models.RarityRare,
		Description: "Manto com encantamentos de proteção.", Price: 580, SellPrice: 232, DefenseBonus: 7, MagicDefBonus: 28, MinLevel: 8, ClassReq: "mage", DropWeight: 9, Slot: "chest",
	},
	"void_robe": {
		ID: "void_robe", Name: "Manto do Vazio", Emoji: "🌌", Type: "armor", Rarity: models.RarityEpic,
		Description: "Tecido com fragmentos da dimensão do vazio.", Price: 1400, SellPrice: 560, DefenseBonus: 10, MagicDefBonus: 40, MinLevel: 14, ClassReq: "mage", DropWeight: 4, Slot: "chest", CABonus: 4,
	},
	"celestial_robe": {
		ID: "celestial_robe", Name: "Manto Celestial", Emoji: "✨", Type: "armor", Rarity: models.RarityLegendary,
		Description: "Trazido dos planos celestiais. Máxima resistência mágica.", Price: 0, SellPrice: 3200, DefenseBonus: 18, MagicDefBonus: 55, MinLevel: 18, ClassReq: "mage", DropWeight: 1, Slot: "chest", CABonus: 5,
	},

	// ── NOVAS ARMADURAS — LADINO ────────────────────────────
	"light_vest": {
		ID: "light_vest", Name: "Colete Leve", Emoji: "🧥", Type: "armor", Rarity: models.RarityCommon,
		Description: "Colete leve que não limita os movimentos.", Price: 45, SellPrice: 18, DefenseBonus: 4, SpeedBonus: 1, MinLevel: 1, ClassReq: "rogue", DropWeight: 25, Slot: "chest", CABonus: 2,
	},
	"studded_leather": {
		ID: "studded_leather", Name: "Couro Tachonado", Emoji: "🥋", Type: "armor", Rarity: models.RarityUncommon,
		Description: "Couro reforçado com tachas de metal.", Price: 200, SellPrice: 80, DefenseBonus: 9, SpeedBonus: 2, MinLevel: 4, ClassReq: "rogue", DropWeight: 15, Slot: "chest", CABonus: 3,
	},
	"assassin_garb": {
		ID: "assassin_garb", Name: "Traje do Assassino", Emoji: "🖤", Type: "armor", Rarity: models.RarityRare,
		Description: "Traje negro que absorve luz e som.", Price: 560, SellPrice: 224, DefenseBonus: 13, SpeedBonus: 4, MinLevel: 9, ClassReq: "rogue", DropWeight: 8, Slot: "chest",
	},
	"nightmare_cloak": {
		ID: "nightmare_cloak", Name: "Manto do Pesadelo", Emoji: "😱", Type: "armor", Rarity: models.RarityEpic,
		Description: "Aterroriza os inimigos ao ser visto.", Price: 1350, SellPrice: 540, DefenseBonus: 20, SpeedBonus: 7, MinLevel: 14, ClassReq: "rogue", DropWeight: 4, Slot: "chest",
	},

	// ── NOVAS ARMADURAS — ARQUEIRO ──────────────────────────
	"scout_vest": {
		ID: "scout_vest", Name: "Colete de Escoteiro", Emoji: "🎽", Type: "armor", Rarity: models.RarityCommon,
		Description: "Colete leve usado por batedores e escoteiros.", Price: 50, SellPrice: 20, DefenseBonus: 5, SpeedBonus: 1, MinLevel: 1, ClassReq: "archer", DropWeight: 25, Slot: "chest", CABonus: 3,
	},
	"hunter_vest": {
		ID: "hunter_vest", Name: "Colete do Caçador", Emoji: "🦺", Type: "armor", Rarity: models.RarityRare,
		Description: "Colete robusto de caçadores experientes.", Price: 540, SellPrice: 216, DefenseBonus: 12, SpeedBonus: 3, MinLevel: 8, ClassReq: "archer", DropWeight: 8, Slot: "chest",
	},
	"wind_vest": {
		ID: "wind_vest", Name: "Colete do Vento", Emoji: "💨", Type: "armor", Rarity: models.RarityEpic,
		Description: "Leve como o vento, veloz como uma flecha.", Price: 1300, SellPrice: 520, DefenseBonus: 16, SpeedBonus: 8, MinLevel: 14, ClassReq: "archer", DropWeight: 4, Slot: "chest", CABonus: 4,
	},
	"eagle_mantle": {
		ID: "eagle_mantle", Name: "Manto da Águia Real", Emoji: "🦅", Type: "armor", Rarity: models.RarityLegendary,
		Description: "Manto encantado com penas de águia sagrada.", Price: 0, SellPrice: 3200, DefenseBonus: 28, SpeedBonus: 12, MinLevel: 18, ClassReq: "archer", DropWeight: 1, Slot: "chest", CABonus: 5,
	},

	// ── ARMADURAS GENÉRICAS (todas as classes) ───────────────
	"reinforced_leather": {
		ID: "reinforced_leather", Name: "Couro Reforçado", Emoji: "🥋", Type: "armor", Rarity: models.RarityCommon,
		Description: "Couro curtido com reforço duplo.", Price: 75, SellPrice: 30, DefenseBonus: 7, MinLevel: 2, DropWeight: 22, Slot: "chest",
	},
	"battle_vest": {
		ID: "battle_vest", Name: "Colete de Batalha", Emoji: "🧥", Type: "armor", Rarity: models.RarityUncommon,
		Description: "Proteção confiável para aventureiros.", Price: 280, SellPrice: 112, DefenseBonus: 16, MinLevel: 6, DropWeight: 14, Slot: "chest", CABonus: 3,
	},
	"warded_armor": {
		ID: "warded_armor", Name: "Armadura Guardada", Emoji: "🔒", Type: "armor", Rarity: models.RarityRare,
		Description: "Encantada com runa de proteção.", Price: 750, SellPrice: 300, DefenseBonus: 24, MagicDefBonus: 10, MinLevel: 10, DropWeight: 7, Slot: "chest", CABonus: 5,
	},
	"eldritch_plate": {
		ID: "eldritch_plate", Name: "Placa Élfica", Emoji: "🌟", Type: "armor", Rarity: models.RarityEpic,
		Description: "Armadura leve e resistente dos elfos.", Price: 1500, SellPrice: 600, DefenseBonus: 32, MagicDefBonus: 18, MinLevel: 14, DropWeight: 4, Slot: "chest",
	},

	// ── ACESSÓRIOS (todas as classes) ───────────────────────
	"ring_iron": {
		ID: "ring_iron", Name: "Anel de Ferro", Emoji: "💍", Type: "accessory", Rarity: models.RarityCommon,
		Description: "Anel simples com um encantamento básico.", Price: 60, SellPrice: 24, AttackBonus: 3, DefenseBonus: 2, MinLevel: 1, DropWeight: 22,
	},
	"amulet_luck": {
		ID: "amulet_luck", Name: "Amuleto da Sorte", Emoji: "🍀", Type: "accessory", Rarity: models.RarityUncommon,
		Description: "Dizem que traz boa sorte ao portador.", Price: 180, SellPrice: 72, AttackBonus: 4, DefenseBonus: 3, SpeedBonus: 1, MinLevel: 3, DropWeight: 15,
	},
	"ring_strength": {
		ID: "ring_strength", Name: "Anel da Força", Emoji: "💪", Type: "accessory", Rarity: models.RarityUncommon,
		Description: "Aumenta a força do portador.", Price: 220, SellPrice: 88, AttackBonus: 6, MinLevel: 5, ClassReq: "warrior", DropWeight: 14,
	},
	"pendant_arcane": {
		ID: "pendant_arcane", Name: "Pingente Arcano", Emoji: "🔮", Type: "accessory", Rarity: models.RarityUncommon,
		Description: "Amplifica o poder mágico de quem o usa.", Price: 230, SellPrice: 92, MagicAtkBonus: 7, MinLevel: 5, ClassReq: "mage", DropWeight: 14,
	},
	"ring_agility": {
		ID: "ring_agility", Name: "Anel da Agilidade", Emoji: "🌀", Type: "accessory", Rarity: models.RarityUncommon,
		Description: "Aumenta a velocidade do portador.", Price: 210, SellPrice: 84, SpeedBonus: 3, AttackBonus: 4, MinLevel: 5, ClassReq: "rogue", DropWeight: 14,
	},
	"bracers_precision": {
		ID: "bracers_precision", Name: "Bráçadeiras de Precisão", Emoji: "🎯", Type: "accessory", Rarity: models.RarityUncommon,
		Description: "Melhora a pontaria do arqueiro.", Price: 210, SellPrice: 84, AttackBonus: 5, SpeedBonus: 2, MinLevel: 5, ClassReq: "archer", DropWeight: 14,
	},
	"ring_protection": {
		ID: "ring_protection", Name: "Anel de Proteção", Emoji: "🛡️", Type: "accessory", Rarity: models.RarityRare,
		Description: "Encantado com runa de proteção.", Price: 500, SellPrice: 200, DefenseBonus: 8, MagicDefBonus: 6, MinLevel: 8, DropWeight: 9,
	},
	"necklace_war": {
		ID: "necklace_war", Name: "Colar do Guerreiro", Emoji: "⚔️", Type: "accessory", Rarity: models.RarityRare,
		Description: "Colar usado pelos grandes guerreiros.", Price: 600, SellPrice: 240, AttackBonus: 10, DefenseBonus: 5, MinLevel: 10, ClassReq: "warrior", DropWeight: 7,
	},
	"ring_mage": {
		ID: "ring_mage", Name: "Anel do Arquimago", Emoji: "✨", Type: "accessory", Rarity: models.RarityRare,
		Description: "Concede controle sobre magia elementar.", Price: 580, SellPrice: 232, MagicAtkBonus: 12, MagicDefBonus: 8, MinLevel: 10, ClassReq: "mage", DropWeight: 7,
	},
	"shadow_ring": {
		ID: "shadow_ring", Name: "Anel das Sombras", Emoji: "🌑", Type: "accessory", Rarity: models.RarityRare,
		Description: "Torna os golpes mais rápidos e silenciosos.", Price: 560, SellPrice: 224, AttackBonus: 8, SpeedBonus: 5, MinLevel: 10, ClassReq: "rogue", DropWeight: 7,
	},
	"eagle_eye": {
		ID: "eagle_eye", Name: "Olho de Águia", Emoji: "🦅", Type: "accessory", Rarity: models.RarityRare,
		Description: "Amuleto que aguça a visão do arqueiro.", Price: 570, SellPrice: 228, AttackBonus: 9, SpeedBonus: 3, MinLevel: 10, ClassReq: "archer", DropWeight: 7,
	},
	"amulet_power": {
		ID: "amulet_power", Name: "Amuleto do Poder", Emoji: "💫", Type: "accessory", Rarity: models.RarityEpic,
		Description: "Emana uma aura de poder sobrenatural.", Price: 1300, SellPrice: 520, AttackBonus: 12, MagicAtkBonus: 12, DefenseBonus: 8, MinLevel: 13, DropWeight: 4,
	},
	"ring_champion": {
		ID: "ring_champion", Name: "Anel do Campeão", Emoji: "🏆", Type: "accessory", Rarity: models.RarityEpic,
		Description: "Portado pelos campeões da arena.", Price: 1400, SellPrice: 560, AttackBonus: 15, DefenseBonus: 10, MinLevel: 15, ClassReq: "warrior", DropWeight: 3,
	},
	"archmage_focus": {
		ID: "archmage_focus", Name: "Foco do Arquimago", Emoji: "🌟", Type: "accessory", Rarity: models.RarityEpic,
		Description: "Amplificador de magia da mais alta ordem.", Price: 1350, SellPrice: 540, MagicAtkBonus: 18, MagicDefBonus: 12, MinLevel: 15, ClassReq: "mage", DropWeight: 3, HitBonus: 3,
	},
	"void_emblem": {
		ID: "void_emblem", Name: "Emblema do Vazio", Emoji: "🌌", Type: "accessory", Rarity: models.RarityEpic,
		Description: "Símbolo da guilda dos assassinos lendários.", Price: 1300, SellPrice: 520, AttackBonus: 14, SpeedBonus: 8, MinLevel: 15, ClassReq: "rogue", DropWeight: 3,
	},
	"hawkeye_charm": {
		ID: "hawkeye_charm", Name: "Amuleto Olho-de-Falcão", Emoji: "🎯", Type: "accessory", Rarity: models.RarityEpic,
		Description: "Nunca erra o alvo.", Price: 1300, SellPrice: 520, AttackBonus: 13, SpeedBonus: 6, MinLevel: 15, ClassReq: "archer", DropWeight: 3,
	},
	"amulet_ancients": {
		ID: "amulet_ancients", Name: "Amuleto dos Anciões", Emoji: "🏺", Type: "accessory", Rarity: models.RarityLegendary,
		Description: "Relíquia dos tempos mais antigos. Poder ilimitado.", Price: 0, SellPrice: 4000, AttackBonus: 20, MagicAtkBonus: 20, DefenseBonus: 15, MagicDefBonus: 15, SpeedBonus: 5, MinLevel: 18, DropWeight: 1,
	},

	// ── ACESSÓRIOS COMUNS (lv 1-2, todas as classes) ────────
	"lucky_charm": {
		ID: "lucky_charm", Name: "Amuleto de Osso", Emoji: "🦴", Type: "accessory", Rarity: models.RarityCommon,
		Description: "Amuleto primitivo que traz pequena proteção.", Price: 30, SellPrice: 12, DefenseBonus: 2, MinLevel: 1, DropWeight: 30,
	},
	"worn_ring": {
		ID: "worn_ring", Name: "Anel Gasto", Emoji: "💍", Type: "accessory", Rarity: models.RarityCommon,
		Description: "Anel surrado com leve encantamento.", Price: 25, SellPrice: 10, AttackBonus: 2, MinLevel: 1, DropWeight: 30,
	},

	// ── ACESSÓRIOS COMUNS POR CLASSE (lv 1-2) ───────────────
	"warrior_token": {
		ID: "warrior_token", Name: "Ficha do Soldado", Emoji: "🪙", Type: "accessory", Rarity: models.RarityCommon,
		Description: "Ficha dada a todo recruta do exército.", Price: 35, SellPrice: 14, AttackBonus: 3, MinLevel: 1, ClassReq: "warrior", DropWeight: 28,
	},
	"apprentice_focus": {
		ID: "apprentice_focus", Name: "Foco de Aprendiz", Emoji: "🔮", Type: "accessory", Rarity: models.RarityCommon,
		Description: "Cristal básico para canalizar magia.", Price: 35, SellPrice: 14, MagicAtkBonus: 3, MinLevel: 1, ClassReq: "mage", DropWeight: 28,
	},
	"thief_token": {
		ID: "thief_token", Name: "Moeda do Ladrão", Emoji: "🪙", Type: "accessory", Rarity: models.RarityCommon,
		Description: "Moeda falsa usada como talismã pelos ladinos.", Price: 30, SellPrice: 12, SpeedBonus: 2, AttackBonus: 1, MinLevel: 1, ClassReq: "rogue", DropWeight: 28,
	},
	"hunter_feather": {
		ID: "hunter_feather", Name: "Pena do Caçador", Emoji: "🪶", Type: "accessory", Rarity: models.RarityCommon,
		Description: "Pena encantada que melhora a mira.", Price: 30, SellPrice: 12, AttackBonus: 2, SpeedBonus: 1, MinLevel: 1, ClassReq: "archer", DropWeight: 28,
	},

	// ── ACESSÓRIOS RAROS (lv 12-13) — gap existente ─────────
	"berserker_band": {
		ID: "berserker_band", Name: "Braçadeira do Berserk", Emoji: "💢", Type: "accessory", Rarity: models.RarityRare,
		Description: "Libera a fúria interior do guerreiro.", Price: 620, SellPrice: 248, AttackBonus: 11, DefenseBonus: 4, MinLevel: 12, ClassReq: "warrior", DropWeight: 8,
	},
	"lich_seal": {
		ID: "lich_seal", Name: "Selo do Lich", Emoji: "💀", Type: "accessory", Rarity: models.RarityRare,
		Description: "Selo necromântico que amplifica magia sombria.", Price: 600, SellPrice: 240, MagicAtkBonus: 13, MagicDefBonus: 7, MinLevel: 12, ClassReq: "mage", DropWeight: 8,
	},
	"venom_ring": {
		ID: "venom_ring", Name: "Anel de Veneno", Emoji: "☠️", Type: "accessory", Rarity: models.RarityRare,
		Description: "Anel encharcado com veneno de aranha.", Price: 590, SellPrice: 236, AttackBonus: 9, SpeedBonus: 4, MinLevel: 12, ClassReq: "rogue", DropWeight: 8,
	},
	"quiver_charm": {
		ID: "quiver_charm", Name: "Amuleto da Aljava", Emoji: "🏹", Type: "accessory", Rarity: models.RarityRare,
		Description: "Aumenta a velocidade e precisão das flechas.", Price: 590, SellPrice: 236, AttackBonus: 10, SpeedBonus: 4, MinLevel: 12, ClassReq: "archer", DropWeight: 8,
	},

	// ── ARMAS / ARMADURAS NÍVEL 16 (gap) ────────────────────
	"sword_runic": {
		ID: "sword_runic", Name: "Espada Rúnica", Emoji: "⚡", Type: "weapon", Rarity: models.RarityEpic,
		Description: "Gravada com runas de batalha ancestrais.", Price: 1700, SellPrice: 680, AttackBonus: 50, MinLevel: 16, ClassReq: "warrior", DropWeight: 4, Slot: "weapon",
	},
	"staff_elder": {
		ID: "staff_elder", Name: "Cajado dos Anciões", Emoji: "🌿", Type: "weapon", Rarity: models.RarityEpic,
		Description: "Cajado usado pelos primeiros magos do mundo.", Price: 1750, SellPrice: 700, MagicAtkBonus: 58, MinLevel: 16, ClassReq: "mage", DropWeight: 4, Slot: "weapon", HitBonus: 2,
	},
	"dagger_reaper": {
		ID: "dagger_reaper", Name: "Adaga da Morte", Emoji: "💀", Type: "weapon", Rarity: models.RarityEpic,
		Description: "Cada golpe drena a força vital do inimigo.", Price: 1600, SellPrice: 640, AttackBonus: 44, MinLevel: 16, ClassReq: "rogue", DropWeight: 4, Slot: "weapon",
	},
	"bow_celestial": {
		ID: "bow_celestial", Name: "Arco Celestial", Emoji: "⭐", Type: "weapon", Rarity: models.RarityEpic,
		Description: "Arco forjado com aço estelar.", Price: 1650, SellPrice: 660, AttackBonus: 50, MinLevel: 16, ClassReq: "archer", DropWeight: 4, Slot: "weapon", HitBonus: 3,
	},
	"guardian_plate": {
		ID: "guardian_plate", Name: "Armadura do Guardião", Emoji: "🛡️", Type: "armor", Rarity: models.RarityEpic,
		Description: "Armadura dos guardiões do reino.", Price: 1550, SellPrice: 620, DefenseBonus: 38, MinLevel: 16, ClassReq: "warrior", DropWeight: 4, Slot: "chest", CABonus: 6,
	},
	"arcane_robe": {
		ID: "arcane_robe", Name: "Manto Arcano Supremo", Emoji: "🌟", Type: "armor", Rarity: models.RarityEpic,
		Description: "Manto imbuído com magia pura dos planos superiores.", Price: 1500, SellPrice: 600, DefenseBonus: 12, MagicDefBonus: 44, MinLevel: 16, ClassReq: "mage", DropWeight: 4, Slot: "chest", CABonus: 2,
	},
	"shadow_shroud": {
		ID: "shadow_shroud", Name: "Véu das Sombras", Emoji: "🌑", Type: "armor", Rarity: models.RarityEpic,
		Description: "Véu que dobra a luz ao redor do portador.", Price: 1450, SellPrice: 580, DefenseBonus: 22, SpeedBonus: 9, MinLevel: 16, ClassReq: "rogue", DropWeight: 4, Slot: "chest", CABonus: 4,
	},
	"storm_mantle": {
		ID: "storm_mantle", Name: "Manto da Tempestade Eterna", Emoji: "⛈️", Type: "armor", Rarity: models.RarityEpic,
		Description: "Carregado com a fúria de mil tempestades.", Price: 1450, SellPrice: 580, DefenseBonus: 18, SpeedBonus: 10, MinLevel: 16, ClassReq: "archer", DropWeight: 4, Slot: "chest", CABonus: 5,
	},

	// ── ACESSÓRIOS LENDÁRIOS POR CLASSE ─────────────────────
	"crown_warlord": {
		ID: "crown_warlord", Name: "Coroa do Senhor da Guerra", Emoji: "👑", Type: "accessory", Rarity: models.RarityLegendary,
		Description: "Usada pelo maior guerreiro que já existiu.", Price: 0, SellPrice: 4500, AttackBonus: 22, DefenseBonus: 18, MinLevel: 18, ClassReq: "warrior", DropWeight: 1,
	},
	"eye_archmage": {
		ID: "eye_archmage", Name: "Olho do Arquimago", Emoji: "👁️", Type: "accessory", Rarity: models.RarityLegendary,
		Description: "Orbe que contém o conhecimento de todos os magos.", Price: 0, SellPrice: 4500, MagicAtkBonus: 25, MagicDefBonus: 18, MinLevel: 18, ClassReq: "mage", DropWeight: 1,
	},
	"phantom_seal": {
		ID: "phantom_seal", Name: "Selo Fantasma", Emoji: "👻", Type: "accessory", Rarity: models.RarityLegendary,
		Description: "Torna o portador quase impossível de rastrear.", Price: 0, SellPrice: 4500, AttackBonus: 18, SpeedBonus: 15, MinLevel: 18, ClassReq: "rogue", DropWeight: 1,
	},
	"starshot_pendant": {
		ID: "starshot_pendant", Name: "Pingente Estrela-Cadente", Emoji: "🌠", Type: "accessory", Rarity: models.RarityLegendary,
		Description: "Canaliza o poder de estrelas distantes.", Price: 0, SellPrice: 4500, AttackBonus: 20, SpeedBonus: 12, MinLevel: 18, ClassReq: "archer", DropWeight: 1,
	},

	// ── CAPACETES / ELMOS ──────────────────────────────────

	// Warrior helmets
	"helmet_iron": {
		ID: "helmet_iron", Name: "Elmo de Ferro", Emoji: "⛑️", Type: "armor", Rarity: models.RarityCommon,
		Description: "Elmo básico que protege a cabeça de golpes.", Price: 55, SellPrice: 22, DefenseBonus: 4, MinLevel: 1, ClassReq: "warrior", DropWeight: 28, Slot: "head", CABonus: 1,
	},
	"helmet_steel": {
		ID: "helmet_steel", Name: "Elmo de Aço", Emoji: "🪖", Type: "armor", Rarity: models.RarityUncommon,
		Description: "Elmo forjado em aço temperado.", Price: 180, SellPrice: 72, DefenseBonus: 10, MinLevel: 5, ClassReq: "warrior", DropWeight: 16, Slot: "head", CABonus: 1,
	},
	"helmet_knights": {
		ID: "helmet_knights", Name: "Elmo do Cavaleiro", Emoji: "🛡️", Type: "armor", Rarity: models.RarityRare,
		Description: "Elmo completo com viseira abatível.", Price: 520, SellPrice: 208, DefenseBonus: 18, MinLevel: 9, ClassReq: "warrior", DropWeight: 9, Slot: "head", CABonus: 2,
	},
	"helmet_warlord": {
		ID: "helmet_warlord", Name: "Elmo do Senhor da Guerra", Emoji: "⚔️", Type: "armor", Rarity: models.RarityEpic,
		Description: "Elmo usado pelos generais que nunca perderam batalha.", Price: 1400, SellPrice: 560, DefenseBonus: 28, MinLevel: 14, ClassReq: "warrior", DropWeight: 4, Slot: "head", CABonus: 2,
	},
	"helmet_dragon": {
		ID: "helmet_dragon", Name: "Elmo do Dragão", Emoji: "🐉", Type: "armor", Rarity: models.RarityLegendary,
		Description: "Forjado com escamas de dragão ancião. Praticamente indestrutível.", Price: 0, SellPrice: 3800, DefenseBonus: 42, MinLevel: 18, ClassReq: "warrior", DropWeight: 1, Slot: "head", CABonus: 3,
	},
	// Mage hoods
	"hood_linen": {
		ID: "hood_linen", Name: "Capuz de Linho", Emoji: "🧢", Type: "armor", Rarity: models.RarityCommon,
		Description: "Capuz simples que facilita a concentração.", Price: 40, SellPrice: 16, MagicDefBonus: 5, MinLevel: 1, ClassReq: "mage", DropWeight: 28, Slot: "head",
	},
	"hood_arcane": {
		ID: "hood_arcane", Name: "Capuz Arcano", Emoji: "🎩", Type: "armor", Rarity: models.RarityUncommon,
		Description: "Capuz imbuído com proteção contra magia.", Price: 170, SellPrice: 68, DefenseBonus: 2, MagicDefBonus: 14, MinLevel: 5, ClassReq: "mage", DropWeight: 16, Slot: "head", CABonus: 1,
	},
	"hood_crystal": {
		ID: "hood_crystal", Name: "Capuz Cristalino", Emoji: "💠", Type: "armor", Rarity: models.RarityRare,
		Description: "Capuz adornado com fragmentos de cristal mágico.", Price: 500, SellPrice: 200, DefenseBonus: 4, MagicDefBonus: 24, MinLevel: 9, ClassReq: "mage", DropWeight: 9, Slot: "head", CABonus: 1,
	},
	"hood_void": {
		ID: "hood_void", Name: "Capuz do Vazio", Emoji: "🌌", Type: "armor", Rarity: models.RarityEpic,
		Description: "Tecido com fios da dimensão do vazio.", Price: 1350, SellPrice: 540, DefenseBonus: 7, MagicDefBonus: 38, MinLevel: 14, ClassReq: "mage", DropWeight: 4, Slot: "head", CABonus: 2,
	},
	"hood_celestial": {
		ID: "hood_celestial", Name: "Capuz Celestial", Emoji: "✨", Type: "armor", Rarity: models.RarityLegendary,
		Description: "Capuz brilhante trazido dos planos celestiais.", Price: 0, SellPrice: 3800, DefenseBonus: 12, MagicDefBonus: 55, MinLevel: 18, ClassReq: "mage", DropWeight: 1, Slot: "head", CABonus: 2,
	},
	// Rogue hoods/masks
	"mask_leather": {
		ID: "mask_leather", Name: "Máscara de Couro", Emoji: "🎭", Type: "armor", Rarity: models.RarityCommon,
		Description: "Máscara simples que oculta a identidade.", Price: 45, SellPrice: 18, DefenseBonus: 3, SpeedBonus: 1, MinLevel: 1, ClassReq: "rogue", DropWeight: 28, Slot: "head", CABonus: 1,
	},
	"mask_shadow": {
		ID: "mask_shadow", Name: "Máscara das Sombras", Emoji: "🌑", Type: "armor", Rarity: models.RarityUncommon,
		Description: "Máscara escura que absorve a luz.", Price: 175, SellPrice: 70, DefenseBonus: 6, SpeedBonus: 2, MinLevel: 5, ClassReq: "rogue", DropWeight: 16, Slot: "head", CABonus: 1,
	},
	"mask_assassin": {
		ID: "mask_assassin", Name: "Máscara do Assassino", Emoji: "💀", Type: "armor", Rarity: models.RarityRare,
		Description: "Máscara usada pelos assassinos de elite.", Price: 480, SellPrice: 192, DefenseBonus: 10, SpeedBonus: 4, MinLevel: 9, ClassReq: "rogue", DropWeight: 9, Slot: "head", CABonus: 2,
	},
	"mask_void": {
		ID: "mask_void", Name: "Máscara do Vazio", Emoji: "🌀", Type: "armor", Rarity: models.RarityEpic,
		Description: "Máscara que distorce a percepção do inimigo.", Price: 1300, SellPrice: 520, DefenseBonus: 16, SpeedBonus: 7, MinLevel: 14, ClassReq: "rogue", DropWeight: 4, Slot: "head", CABonus: 2,
	},
	"mask_phantom": {
		ID: "mask_phantom", Name: "Máscara Fantasma", Emoji: "👻", Type: "armor", Rarity: models.RarityLegendary,
		Description: "Torna o portador invisível nas sombras.", Price: 0, SellPrice: 3800, DefenseBonus: 24, SpeedBonus: 14, MinLevel: 18, ClassReq: "rogue", DropWeight: 1, Slot: "head", CABonus: 3,
	},
	// Archer hoods
	"hood_ranger": {
		ID: "hood_ranger", Name: "Capuz do Ranger", Emoji: "🍃", Type: "armor", Rarity: models.RarityCommon,
		Description: "Capuz leve que não atrapalha a mira.", Price: 40, SellPrice: 16, DefenseBonus: 3, SpeedBonus: 1, MinLevel: 1, ClassReq: "archer", DropWeight: 28, Slot: "head", CABonus: 1,
	},
	"hood_hunter": {
		ID: "hood_hunter", Name: "Capuz do Caçador", Emoji: "🎯", Type: "armor", Rarity: models.RarityUncommon,
		Description: "Capuz reforçado que protege sem limitar os sentidos.", Price: 165, SellPrice: 66, DefenseBonus: 7, SpeedBonus: 2, MinLevel: 5, ClassReq: "archer", DropWeight: 16, Slot: "head", CABonus: 1,
	},
	"hood_elven": {
		ID: "hood_elven", Name: "Capuz Élfico", Emoji: "🌿", Type: "armor", Rarity: models.RarityRare,
		Description: "Capuz leve dos elfos batedores.", Price: 470, SellPrice: 188, DefenseBonus: 11, SpeedBonus: 3, MinLevel: 9, ClassReq: "archer", DropWeight: 9, Slot: "head", CABonus: 2,
	},
	"hood_storm": {
		ID: "hood_storm", Name: "Capuz da Tempestade", Emoji: "⛈️", Type: "armor", Rarity: models.RarityEpic,
		Description: "Protege contra ventos e relâmpagos durante a caça.", Price: 1300, SellPrice: 520, DefenseBonus: 17, SpeedBonus: 6, MinLevel: 14, ClassReq: "archer", DropWeight: 4, Slot: "head", CABonus: 2,
	},
	"hood_eagle": {
		ID: "hood_eagle", Name: "Capuz da Águia Real", Emoji: "🦅", Type: "armor", Rarity: models.RarityLegendary,
		Description: "Capuz com penas sagradas que aguçam todos os sentidos.", Price: 0, SellPrice: 3800, DefenseBonus: 26, SpeedBonus: 13, MinLevel: 18, ClassReq: "archer", DropWeight: 1, Slot: "head", CABonus: 3,
	},

	// ── BOTAS ──────────────────────────────────────────────

	// Warrior boots
	"boots_iron": {
		ID: "boots_iron", Name: "Botas de Ferro", Emoji: "👢", Type: "armor", Rarity: models.RarityCommon,
		Description: "Botas pesadas de ferro que protegem os pés.", Price: 50, SellPrice: 20, DefenseBonus: 3, MinLevel: 1, ClassReq: "warrior", DropWeight: 28, Slot: "feet", CABonus: 1,
	},
	"boots_steel": {
		ID: "boots_steel", Name: "Botas de Aço", Emoji: "🥾", Type: "armor", Rarity: models.RarityUncommon,
		Description: "Botas robustas de aço para longas marchas.", Price: 170, SellPrice: 68, DefenseBonus: 8, MinLevel: 5, ClassReq: "warrior", DropWeight: 16, Slot: "feet", CABonus: 1,
	},
	"boots_knights": {
		ID: "boots_knights", Name: "Botas do Cavaleiro", Emoji: "👢", Type: "armor", Rarity: models.RarityRare,
		Description: "Botas reforçadas do arsenal do cavaleiro.", Price: 490, SellPrice: 196, DefenseBonus: 15, MinLevel: 9, ClassReq: "warrior", DropWeight: 9, Slot: "feet", CABonus: 1,
	},
	"boots_titan": {
		ID: "boots_titan", Name: "Botas do Titã", Emoji: "🦵", Type: "armor", Rarity: models.RarityEpic,
		Description: "Botas pesadas que fazem o chão tremer a cada passo.", Price: 1350, SellPrice: 540, DefenseBonus: 25, MinLevel: 14, ClassReq: "warrior", DropWeight: 4, Slot: "feet", CABonus: 2,
	},
	"boots_dragon": {
		ID: "boots_dragon", Name: "Botas do Dragão", Emoji: "🐉", Type: "armor", Rarity: models.RarityLegendary,
		Description: "Botas cobertas com escamas de dragão.", Price: 0, SellPrice: 3600, DefenseBonus: 38, MinLevel: 18, ClassReq: "warrior", DropWeight: 1, Slot: "feet", CABonus: 2,
	},
	// Mage boots
	"boots_cloth": {
		ID: "boots_cloth", Name: "Sapatos de Tecido", Emoji: "👟", Type: "armor", Rarity: models.RarityCommon,
		Description: "Calçados leves que facilitam a movimentação do mago.", Price: 38, SellPrice: 15, MagicDefBonus: 4, MinLevel: 1, ClassReq: "mage", DropWeight: 28, Slot: "feet",
	},
	"boots_arcane": {
		ID: "boots_arcane", Name: "Botas Arcanas", Emoji: "🥿", Type: "armor", Rarity: models.RarityUncommon,
		Description: "Imbuídas com runas de levitação.", Price: 160, SellPrice: 64, DefenseBonus: 2, MagicDefBonus: 12, MinLevel: 5, ClassReq: "mage", DropWeight: 16, Slot: "feet",
	},
	"boots_mystic": {
		ID: "boots_mystic", Name: "Botas Místicas", Emoji: "✨", Type: "armor", Rarity: models.RarityRare,
		Description: "Botas que sussurram feitiços a cada passo.", Price: 460, SellPrice: 184, DefenseBonus: 4, MagicDefBonus: 20, MinLevel: 9, ClassReq: "mage", DropWeight: 9, Slot: "feet",
	},
	"boots_void_mage": {
		ID: "boots_void_mage", Name: "Botas do Vazio", Emoji: "🌌", Type: "armor", Rarity: models.RarityEpic,
		Description: "Pisam entre as dimensões a cada passo.", Price: 1300, SellPrice: 520, DefenseBonus: 6, MagicDefBonus: 34, MinLevel: 14, ClassReq: "mage", DropWeight: 4, Slot: "feet",
	},
	"boots_celestial": {
		ID: "boots_celestial", Name: "Botas Celestiais", Emoji: "☁️", Type: "armor", Rarity: models.RarityLegendary,
		Description: "Andam sobre o ar como se fosse terra sólida.", Price: 0, SellPrice: 3600, DefenseBonus: 10, MagicDefBonus: 50, MinLevel: 18, ClassReq: "mage", DropWeight: 1, Slot: "feet",
	},
	// Rogue boots
	"boots_soft": {
		ID: "boots_soft", Name: "Botas Silenciosas", Emoji: "🧦", Type: "armor", Rarity: models.RarityCommon,
		Description: "Botas que não fazem barulho ao caminhar.", Price: 42, SellPrice: 17, DefenseBonus: 2, SpeedBonus: 2, MinLevel: 1, ClassReq: "rogue", DropWeight: 28, Slot: "feet",
	},
	"boots_swift": {
		ID: "boots_swift", Name: "Botas Velozes", Emoji: "👟", Type: "armor", Rarity: models.RarityUncommon,
		Description: "Botas encantadas que aumentam a velocidade.", Price: 165, SellPrice: 66, DefenseBonus: 5, SpeedBonus: 3, MinLevel: 5, ClassReq: "rogue", DropWeight: 16, Slot: "feet",
	},
	"boots_shadow": {
		ID: "boots_shadow", Name: "Botas das Sombras", Emoji: "🌑", Type: "armor", Rarity: models.RarityRare,
		Description: "Passam por qualquer superfície sem deixar rastro.", Price: 450, SellPrice: 180, DefenseBonus: 8, SpeedBonus: 5, MinLevel: 9, ClassReq: "rogue", DropWeight: 9, Slot: "feet",
	},
	"boots_phantom": {
		ID: "boots_phantom", Name: "Botas Fantasma", Emoji: "👻", Type: "armor", Rarity: models.RarityEpic,
		Description: "Atravessam obstáculos como fantasmas.", Price: 1250, SellPrice: 500, DefenseBonus: 14, SpeedBonus: 9, MinLevel: 14, ClassReq: "rogue", DropWeight: 4, Slot: "feet",
	},
	"boots_void_rogue": {
		ID: "boots_void_rogue", Name: "Botas do Vazio Sombrio", Emoji: "🌀", Type: "armor", Rarity: models.RarityLegendary,
		Description: "Pés que pisam no vazio entre os planos.", Price: 0, SellPrice: 3600, DefenseBonus: 22, SpeedBonus: 16, MinLevel: 18, ClassReq: "rogue", DropWeight: 1, Slot: "feet",
	},
	// Archer boots
	"boots_light": {
		ID: "boots_light", Name: "Botas Leves", Emoji: "👡", Type: "armor", Rarity: models.RarityCommon,
		Description: "Botas leves que permitem movimentação rápida.", Price: 42, SellPrice: 17, DefenseBonus: 3, SpeedBonus: 2, MinLevel: 1, ClassReq: "archer", DropWeight: 28, Slot: "feet",
	},
	"boots_tracker": {
		ID: "boots_tracker", Name: "Botas do Rastreador", Emoji: "👟", Type: "armor", Rarity: models.RarityUncommon,
		Description: "Permitem seguir rastros sem ser notado.", Price: 160, SellPrice: 64, DefenseBonus: 6, SpeedBonus: 3, MinLevel: 5, ClassReq: "archer", DropWeight: 16, Slot: "feet",
	},
	"boots_wind": {
		ID: "boots_wind", Name: "Botas do Vento", Emoji: "💨", Type: "armor", Rarity: models.RarityRare,
		Description: "Correm velozes como o vento da floresta.", Price: 445, SellPrice: 178, DefenseBonus: 9, SpeedBonus: 5, MinLevel: 9, ClassReq: "archer", DropWeight: 9, Slot: "feet",
	},
	"boots_tempest": {
		ID: "boots_tempest", Name: "Botas da Tempestade", Emoji: "⛈️", Type: "armor", Rarity: models.RarityEpic,
		Description: "A cada passo, deixam rastros de eletricidade.", Price: 1250, SellPrice: 500, DefenseBonus: 15, SpeedBonus: 8, MinLevel: 14, ClassReq: "archer", DropWeight: 4, Slot: "feet",
	},
	"boots_starwalker": {
		ID: "boots_starwalker", Name: "Botas Caminhante-Estelar", Emoji: "🌠", Type: "armor", Rarity: models.RarityLegendary,
		Description: "Caminham entre as estrelas.", Price: 0, SellPrice: 3600, DefenseBonus: 23, SpeedBonus: 14, MinLevel: 18, ClassReq: "archer", DropWeight: 1, Slot: "feet",
	},

	// ── LUVAS / BRÁÇADEIRAS ─────────────────────────────────

	// Warrior gauntlets
	"gauntlets_iron": {
		ID: "gauntlets_iron", Name: "Manoplas de Ferro", Emoji: "🥊", Type: "armor", Rarity: models.RarityCommon,
		Description: "Manoplas simples que protegem as mãos.", Price: 45, SellPrice: 18, DefenseBonus: 3, AttackBonus: 2, MinLevel: 1, ClassReq: "warrior", DropWeight: 28, Slot: "hands", CABonus: 1,
	},
	"gauntlets_steel": {
		ID: "gauntlets_steel", Name: "Manoplas de Aço", Emoji: "🤜", Type: "armor", Rarity: models.RarityUncommon,
		Description: "Manoplas reforçadas que aumentam o impacto.", Price: 160, SellPrice: 64, DefenseBonus: 7, AttackBonus: 5, MinLevel: 5, ClassReq: "warrior", DropWeight: 16, Slot: "hands", CABonus: 1,
	},
	"gauntlets_knights": {
		ID: "gauntlets_knights", Name: "Manoplas do Cavaleiro", Emoji: "⚔️", Type: "armor", Rarity: models.RarityRare,
		Description: "Manoplas pesadas com espinhos de aço.", Price: 460, SellPrice: 184, DefenseBonus: 12, AttackBonus: 10, MinLevel: 9, ClassReq: "warrior", DropWeight: 9, Slot: "hands", CABonus: 2,
	},
	"gauntlets_warlord": {
		ID: "gauntlets_warlord", Name: "Manoplas do Senhor da Guerra", Emoji: "💥", Type: "armor", Rarity: models.RarityEpic,
		Description: "Manoplas que triplicam a força do soco.", Price: 1280, SellPrice: 512, DefenseBonus: 20, AttackBonus: 18, MinLevel: 14, ClassReq: "warrior", DropWeight: 4, Slot: "hands", CABonus: 2,
	},
	"gauntlets_dragon": {
		ID: "gauntlets_dragon", Name: "Manoplas do Dragão", Emoji: "🐉", Type: "armor", Rarity: models.RarityLegendary,
		Description: "Forjadas com garras de dragão. Golpes que rasgam o ar.", Price: 0, SellPrice: 3600, DefenseBonus: 32, AttackBonus: 28, MinLevel: 18, ClassReq: "warrior", DropWeight: 1, Slot: "hands", CABonus: 3,
	},
	// Mage gloves
	"gloves_cloth": {
		ID: "gloves_cloth", Name: "Luvas de Tecido", Emoji: "🧤", Type: "armor", Rarity: models.RarityCommon,
		Description: "Luvas simples que protegem as mãos durante feitiços.", Price: 35, SellPrice: 14, MagicAtkBonus: 3, MinLevel: 1, ClassReq: "mage", DropWeight: 28, Slot: "hands",
	},
	"gloves_arcane": {
		ID: "gloves_arcane", Name: "Luvas Arcanas", Emoji: "✨", Type: "armor", Rarity: models.RarityUncommon,
		Description: "Luvas que amplificam o toque mágico.", Price: 155, SellPrice: 62, MagicAtkBonus: 9, MinLevel: 5, ClassReq: "mage", DropWeight: 16, Slot: "hands",
	},
	"gloves_crystal": {
		ID: "gloves_crystal", Name: "Luvas Cristalinas", Emoji: "💠", Type: "armor", Rarity: models.RarityRare,
		Description: "Luvas incrustadas com fragmentos de cristal mágico.", Price: 440, SellPrice: 176, MagicAtkBonus: 17, MagicDefBonus: 8, MinLevel: 9, ClassReq: "mage", DropWeight: 9, Slot: "hands",
	},
	"gloves_void_mage": {
		ID: "gloves_void_mage", Name: "Luvas do Vazio", Emoji: "🌌", Type: "armor", Rarity: models.RarityEpic,
		Description: "Tocam a realidade e dobram sua forma.", Price: 1250, SellPrice: 500, MagicAtkBonus: 28, MagicDefBonus: 15, MinLevel: 14, ClassReq: "mage", DropWeight: 4, Slot: "hands",
	},
	"gloves_archmage": {
		ID: "gloves_archmage", Name: "Luvas do Arquimago", Emoji: "🔮", Type: "armor", Rarity: models.RarityLegendary,
		Description: "Canalizaram a magia de mil feitiços.", Price: 0, SellPrice: 3600, MagicAtkBonus: 42, MagicDefBonus: 25, MinLevel: 18, ClassReq: "mage", DropWeight: 1, Slot: "hands",
	},
	// Rogue gloves
	"gloves_leather_rogue": {
		ID: "gloves_leather_rogue", Name: "Luvas de Couro", Emoji: "🧤", Type: "armor", Rarity: models.RarityCommon,
		Description: "Luvas leves que não atrapalham a destreza.", Price: 38, SellPrice: 15, AttackBonus: 2, SpeedBonus: 1, MinLevel: 1, ClassReq: "rogue", DropWeight: 28, Slot: "hands",
	},
	"gloves_swift": {
		ID: "gloves_swift", Name: "Luvas Velozes", Emoji: "⚡", Type: "armor", Rarity: models.RarityUncommon,
		Description: "Aumentam a velocidade dos golpes.", Price: 155, SellPrice: 62, AttackBonus: 5, SpeedBonus: 2, MinLevel: 5, ClassReq: "rogue", DropWeight: 16, Slot: "hands",
	},
	"gloves_poison": {
		ID: "gloves_poison", Name: "Luvas Venenosas", Emoji: "☠️", Type: "armor", Rarity: models.RarityRare,
		Description: "Impregnadas com veneno de aranha.", Price: 430, SellPrice: 172, AttackBonus: 9, SpeedBonus: 4, MinLevel: 9, ClassReq: "rogue", DropWeight: 9, Slot: "hands",
	},
	"gloves_shadow": {
		ID: "gloves_shadow", Name: "Luvas das Sombras", Emoji: "🌑", Type: "armor", Rarity: models.RarityEpic,
		Description: "Dedos que tocam sem deixar rastro.", Price: 1200, SellPrice: 480, AttackBonus: 15, SpeedBonus: 7, MinLevel: 14, ClassReq: "rogue", DropWeight: 4, Slot: "hands",
	},
	"gloves_reaper": {
		ID: "gloves_reaper", Name: "Luvas da Morte", Emoji: "💀", Type: "armor", Rarity: models.RarityLegendary,
		Description: "Cada toque drena a vida do inimigo.", Price: 0, SellPrice: 3600, AttackBonus: 24, SpeedBonus: 14, MinLevel: 18, ClassReq: "rogue", DropWeight: 1, Slot: "hands",
	},
	// Archer bracers
	"bracers_basic": {
		ID: "bracers_basic", Name: "Bráçadeiras Básicas", Emoji: "🤜", Type: "armor", Rarity: models.RarityCommon,
		Description: "Bráçadeiras simples para proteger os braços.", Price: 38, SellPrice: 15, AttackBonus: 2, SpeedBonus: 1, MinLevel: 1, ClassReq: "archer", DropWeight: 28, Slot: "hands",
	},
	"bracers_hunter": {
		ID: "bracers_hunter", Name: "Bráçadeiras do Caçador", Emoji: "🎯", Type: "armor", Rarity: models.RarityUncommon,
		Description: "Guiam as flechas com precisão cirúrgica.", Price: 155, SellPrice: 62, AttackBonus: 6, SpeedBonus: 2, MinLevel: 5, ClassReq: "archer", DropWeight: 16, Slot: "hands",
	},
	"bracers_elven": {
		ID: "bracers_elven", Name: "Bráçadeiras Élficas", Emoji: "🌿", Type: "armor", Rarity: models.RarityRare,
		Description: "Bráçadeiras usadas pelos arqueiros élficos de elite.", Price: 430, SellPrice: 172, AttackBonus: 11, SpeedBonus: 4, MinLevel: 9, ClassReq: "archer", DropWeight: 9, Slot: "hands",
	},
	"bracers_storm": {
		ID: "bracers_storm", Name: "Bráçadeiras da Tempestade", Emoji: "⛈️", Type: "armor", Rarity: models.RarityEpic,
		Description: "Carregadas com estática que guia as flechas.", Price: 1200, SellPrice: 480, AttackBonus: 18, SpeedBonus: 6, MinLevel: 14, ClassReq: "archer", DropWeight: 4, Slot: "hands",
	},
	"bracers_dragon": {
		ID: "bracers_dragon", Name: "Bráçadeiras do Dragão", Emoji: "🐉", Type: "armor", Rarity: models.RarityLegendary,
		Description: "Flechas disparadas nunca erram o alvo.", Price: 0, SellPrice: 3600, AttackBonus: 28, SpeedBonus: 12, MinLevel: 18, ClassReq: "archer", DropWeight: 1, Slot: "hands",
	},

	// ── ESCUDOS (Warrior) ───────────────────────────────────
	"shield_wooden": {
		ID: "shield_wooden", Name: "Escudo de Madeira", Emoji: "🪵", Type: "armor", Rarity: models.RarityCommon,
		Description: "Escudo simples de madeira reforçada.", Price: 35, SellPrice: 14, DefenseBonus: 4, MinLevel: 1, ClassReq: "warrior", DropWeight: 30, Slot: "offhand", CABonus: 1,
	},
	"shield_steel": {
		ID: "shield_steel", Name: "Escudo de Aço", Emoji: "🛡️", Type: "armor", Rarity: models.RarityUncommon,
		Description: "Escudo de aço temperado com brasão do reino.", Price: 190, SellPrice: 76, DefenseBonus: 13, MinLevel: 6, ClassReq: "warrior", DropWeight: 16, Slot: "offhand", CABonus: 2,
	},
	"shield_tower": {
		ID: "shield_tower", Name: "Escudo Torre", Emoji: "🏰", Type: "armor", Rarity: models.RarityRare,
		Description: "Escudo enorme que protege o corpo inteiro.", Price: 560, SellPrice: 224, DefenseBonus: 24, MinLevel: 10, ClassReq: "warrior", DropWeight: 9, Slot: "offhand", CABonus: 3,
	},
	"shield_blessed": {
		ID: "shield_blessed", Name: "Escudo Abençoado", Emoji: "⚔️", Type: "armor", Rarity: models.RarityEpic,
		Description: "Abençoado por um clérigo de alta ordem.", Price: 1400, SellPrice: 560, DefenseBonus: 36, MagicDefBonus: 15, MinLevel: 15, ClassReq: "warrior", DropWeight: 4, Slot: "offhand", CABonus: 4,
	},
	"shield_aegis": {
		ID: "shield_aegis", Name: "Égide Divina", Emoji: "🌟", Type: "armor", Rarity: models.RarityLegendary,
		Description: "Escudo lendário que repele qualquer ataque.", Price: 0, SellPrice: 4000, DefenseBonus: 52, MagicDefBonus: 30, MinLevel: 18, ClassReq: "warrior", DropWeight: 1, Slot: "offhand", CABonus: 5,
	},

	// ── ARMAS ADICIONAIS ────────────────────────────────────

	// Warrior – spear line
	"spear_iron": {
		ID: "spear_iron", Name: "Lança de Ferro", Emoji: "🏹", Type: "weapon", Rarity: models.RarityCommon,
		Description: "Lança leve e versátil.", Price: 75, SellPrice: 30, AttackBonus: 9, MinLevel: 2, ClassReq: "warrior", DropWeight: 26, Slot: "weapon", HitBonus: 1,
	},
	"spear_steel": {
		ID: "spear_steel", Name: "Lança de Aço", Emoji: "⚔️", Type: "weapon", Rarity: models.RarityUncommon,
		Description: "Lança de alcance médio com ponta afiada.", Price: 240, SellPrice: 96, AttackBonus: 20, MinLevel: 6, ClassReq: "warrior", DropWeight: 16, Slot: "weapon", HitBonus: 1,
	},
	"spear_thunder": {
		ID: "spear_thunder", Name: "Lança do Trovão", Emoji: "⚡", Type: "weapon", Rarity: models.RarityRare,
		Description: "Vibra com energia elétrica a cada golpe.", Price: 680, SellPrice: 272, AttackBonus: 36, MinLevel: 11, ClassReq: "warrior", DropWeight: 9, Slot: "weapon", HitBonus: 2,
	},
	// Mage – orb/tome line
	"orb_frost": {
		ID: "orb_frost", Name: "Orbe de Gelo", Emoji: "❄️", Type: "weapon", Rarity: models.RarityUncommon,
		Description: "Orbe que amplifica magias de gelo.", Price: 250, SellPrice: 100, MagicAtkBonus: 21, MinLevel: 6, ClassReq: "mage", DropWeight: 16, Slot: "weapon", HitBonus: 2,
	},
	"tome_shadow": {
		ID: "tome_shadow", Name: "Grimório das Sombras", Emoji: "📕", Type: "weapon", Rarity: models.RarityRare,
		Description: "Contém feitiços de magia sombria proibida.", Price: 720, SellPrice: 288, MagicAtkBonus: 36, MinLevel: 11, ClassReq: "mage", DropWeight: 9, Slot: "weapon", HitBonus: 2,
	},
	"tome_arcane": {
		ID: "tome_arcane", Name: "Grimório Arcano", Emoji: "📘", Type: "weapon", Rarity: models.RarityEpic,
		Description: "Grimório com feitiços de magia pura concentrada.", Price: 1750, SellPrice: 700, MagicAtkBonus: 57, MinLevel: 16, ClassReq: "mage", DropWeight: 5, Slot: "weapon", HitBonus: 2,
	},
	// Rogue – katar/claw line
	"katar_iron": {
		ID: "katar_iron", Name: "Katar de Ferro", Emoji: "🗡️", Type: "weapon", Rarity: models.RarityCommon,
		Description: "Punhal de haste usado em combate corpo a corpo.", Price: 65, SellPrice: 26, AttackBonus: 8, MinLevel: 2, ClassReq: "rogue", DropWeight: 26, Slot: "weapon", HitBonus: 1,
	},
	"claw_steel": {
		ID: "claw_steel", Name: "Garra de Aço", Emoji: "🐾", Type: "weapon", Rarity: models.RarityUncommon,
		Description: "Garra metálica que rasga armaduras.", Price: 210, SellPrice: 84, AttackBonus: 17, MinLevel: 6, ClassReq: "rogue", DropWeight: 16, Slot: "weapon", HitBonus: 1,
	},
	"claw_shadow": {
		ID: "claw_shadow", Name: "Garra das Sombras", Emoji: "🌑", Type: "weapon", Rarity: models.RarityRare,
		Description: "Garra que envenena a cada arranhão.", Price: 620, SellPrice: 248, AttackBonus: 32, MinLevel: 11, ClassReq: "rogue", DropWeight: 9, Slot: "weapon", HitBonus: 2,
	},
	// Archer – repeating crossbow / magic quiver
	"repeating_crossbow": {
		ID: "repeating_crossbow", Name: "Besta de Repetição", Emoji: "🏹", Type: "weapon", Rarity: models.RarityUncommon,
		Description: "Dispara múltiplos dardos em sequência.", Price: 230, SellPrice: 92, AttackBonus: 19, MinLevel: 6, ClassReq: "archer", DropWeight: 16, Slot: "weapon", HitBonus: 2,
	},
	"quiver_magic": {
		ID: "quiver_magic", Name: "Aljava Mágica", Emoji: "✨", Type: "weapon", Rarity: models.RarityRare,
		Description: "Flechas nunca acabam e são imbuídas com magia.", Price: 640, SellPrice: 256, AttackBonus: 31, MinLevel: 10, ClassReq: "archer", DropWeight: 9, Slot: "weapon", HitBonus: 2,
	},
	"bow_phantom": {
		ID: "bow_phantom", Name: "Arco Fantasma", Emoji: "👻", Type: "weapon", Rarity: models.RarityEpic,
		Description: "As flechas atravessam obstáculos como fantasmas.", Price: 1650, SellPrice: 660, AttackBonus: 52, MinLevel: 16, ClassReq: "archer", DropWeight: 5, HitBonus: 3,
	},

	// ════════════════════════════════════════════════════════
	// SLOT ITEMS — cabeça, peito, mãos, pernas, pés, acessórios
	// Cada slot tem 5 itens por classe (Comum→Lendário)
	// Warrior=guerreiro só equipa warrior, etc.
	// Arcos são universais (ClassReq="")
	// ════════════════════════════════════════════════════════

	// ── HEAD: WARRIOR ────────────────────────────────────────
	"helm_iron": {
		ID: "helm_iron", Name: "Elmo de Ferro", Emoji: "⛑️", Type: "armor", Rarity: models.RarityCommon,
		Description: "Elmo básico de ferro que protege a cabeça.", Price: 70, SellPrice: 28, DefenseBonus: 3, CABonus: 1, MinLevel: 1, ClassReq: "warrior", DropWeight: 20, Slot: "head",
	},
	"helm_steel": {
		ID: "helm_steel", Name: "Elmo de Aço", Emoji: "⛑️", Type: "armor", Rarity: models.RarityUncommon,
		Description: "Elmo forjado com aço temperado.", Price: 200, SellPrice: 80, DefenseBonus: 6, CABonus: 2, MinLevel: 5, ClassReq: "warrior", DropWeight: 12, Slot: "head",
	},
	"helm_knight": {
		ID: "helm_knight", Name: "Elmo do Cavaleiro", Emoji: "🪖", Type: "armor", Rarity: models.RarityRare,
		Description: "Elmo completo com viseira articulada.", Price: 500, SellPrice: 200, DefenseBonus: 10, CABonus: 3, MinLevel: 9, ClassReq: "warrior", DropWeight: 7, Slot: "head",
	},
	"helm_warlord": {
		ID: "helm_warlord", Name: "Elmo do Senhor da Guerra", Emoji: "👑", Type: "armor", Rarity: models.RarityEpic,
		Description: "Elmo lendário que inspira medo nos inimigos.", Price: 1400, SellPrice: 560, DefenseBonus: 16, CABonus: 4, MinLevel: 13, ClassReq: "warrior", DropWeight: 3, Slot: "head",
	},
	"helm_titan": {
		ID: "helm_titan", Name: "Elmo do Titã", Emoji: "🏆", Type: "armor", Rarity: models.RarityLegendary,
		Description: "Forjado com oricálco. Quase impossível de romper.", Price: 0, SellPrice: 2500, DefenseBonus: 22, CABonus: 6, HPBonus: 30, MinLevel: 17, ClassReq: "warrior", DropWeight: 1, Slot: "head",
	},

	// ── HEAD: MAGE ───────────────────────────────────────────
	"hood_apprentice": {
		ID: "hood_apprentice", Name: "Capuz do Aprendiz", Emoji: "🎓", Type: "armor", Rarity: models.RarityCommon,
		Description: "Capuz básico com fio de prata para concentração.", Price: 60, SellPrice: 24, MagicDefBonus: 5, CABonus: 1, MinLevel: 1, ClassReq: "mage", DropWeight: 20, Slot: "head",
	},
	"circlet_arcane": {
		ID: "circlet_arcane", Name: "Diadema Arcano", Emoji: "💍", Type: "armor", Rarity: models.RarityUncommon,
		Description: "Diadema que amplifica a concentração mágica.", Price: 220, SellPrice: 88, MagicDefBonus: 12, MagicAtkBonus: 5, CABonus: 1, MinLevel: 5, ClassReq: "mage", DropWeight: 12, Slot: "head",
	},
	"cowl_mystic": {
		ID: "cowl_mystic", Name: "Capuz Místico", Emoji: "🔮", Type: "armor", Rarity: models.RarityRare,
		Description: "Capuz tecido com fios de magia pura.", Price: 550, SellPrice: 220, MagicDefBonus: 20, MagicAtkBonus: 10, CABonus: 2, MinLevel: 9, ClassReq: "mage", DropWeight: 7, Slot: "head",
	},
	"crown_sorcerer": {
		ID: "crown_sorcerer", Name: "Coroa do Feiticeiro", Emoji: "👸", Type: "armor", Rarity: models.RarityEpic,
		Description: "Coroa que amplifica drasticamente o poder arcano.", Price: 1500, SellPrice: 600, MagicDefBonus: 30, MagicAtkBonus: 18, CABonus: 2, MinLevel: 13, ClassReq: "mage", DropWeight: 3, Slot: "head",
	},
	"crown_archmage": {
		ID: "crown_archmage", Name: "Coroa do Arquimago", Emoji: "👑", Type: "armor", Rarity: models.RarityLegendary,
		Description: "Pertencia ao mais poderoso mago da era antiga.", Price: 0, SellPrice: 2800, MagicDefBonus: 45, MagicAtkBonus: 28, CABonus: 3, MPBonus: 40, MinLevel: 17, ClassReq: "mage", DropWeight: 1, Slot: "head",
	},

	// ── HEAD: ROGUE ──────────────────────────────────────────
	"hood_leather": {
		ID: "hood_leather", Name: "Capuz de Couro", Emoji: "🧢", Type: "armor", Rarity: models.RarityCommon,
		Description: "Capuz simples que esconde o rosto.", Price: 50, SellPrice: 20, DefenseBonus: 2, CABonus: 1, HitBonus: 0, MinLevel: 1, ClassReq: "rogue", DropWeight: 20, Slot: "head",
	},
	"cowl_assassin": {
		ID: "cowl_assassin", Name: "Capuz do Assassino", Emoji: "😶‍🌫️", Type: "armor", Rarity: models.RarityRare,
		Description: "Permite se mover nas sombras quase invisível.", Price: 500, SellPrice: 200, DefenseBonus: 7, CABonus: 3, HitBonus: 1, MinLevel: 9, ClassReq: "rogue", DropWeight: 7, Slot: "head",
	},
	"visor_eclipse": {
		ID: "visor_eclipse", Name: "Viseira do Eclipse", Emoji: "🌒", Type: "armor", Rarity: models.RarityEpic,
		Description: "Viseira imbuída com energia das trevas.", Price: 1300, SellPrice: 520, DefenseBonus: 11, CABonus: 3, HitBonus: 2, MinLevel: 13, ClassReq: "rogue", DropWeight: 3, Slot: "head",
	},

	// ── HEAD: ARCHER ─────────────────────────────────────────
	"helm_scout": {
		ID: "helm_scout", Name: "Elmo do Escoteiro", Emoji: "⛑️", Type: "armor", Rarity: models.RarityUncommon,
		Description: "Proteção leve com lentes de longa distância.", Price: 200, SellPrice: 80, DefenseBonus: 5, CABonus: 2, HitBonus: 1, MinLevel: 5, ClassReq: "archer", DropWeight: 12, Slot: "head",
	},
	"crown_hunter": {
		ID: "crown_hunter", Name: "Coroa do Caçador", Emoji: "🌿", Type: "armor", Rarity: models.RarityRare,
		Description: "Coroa de galhos encantados que aguça os sentidos.", Price: 500, SellPrice: 200, DefenseBonus: 7, CABonus: 2, HitBonus: 2, MinLevel: 9, ClassReq: "archer", DropWeight: 7, Slot: "head",
	},
	"visor_eagle": {
		ID: "visor_eagle", Name: "Viseira da Águia", Emoji: "🦅", Type: "armor", Rarity: models.RarityEpic,
		Description: "Melhora a precisão drasticamente.", Price: 1300, SellPrice: 520, DefenseBonus: 10, CABonus: 3, HitBonus: 2, MinLevel: 13, ClassReq: "archer", DropWeight: 3, Slot: "head",
	},
	"crown_wind": {
		ID: "crown_wind", Name: "Coroa dos Ventos", Emoji: "💨", Type: "armor", Rarity: models.RarityLegendary,
		Description: "Concedida pela deusa dos ventos aos melhores arqueiros.", Price: 0, SellPrice: 2500, DefenseBonus: 14, CABonus: 4, HitBonus: 3, SpeedBonus: 5, MinLevel: 17, ClassReq: "archer", DropWeight: 1, Slot: "head",
	},

	// ── CHEST: WARRIOR ───────────────────────────────────────
	"breastplate_iron": {
		ID: "breastplate_iron", Name: "Peitoral de Ferro", Emoji: "🛡️", Type: "armor", Rarity: models.RarityCommon,
		Description: "Proteção torácica básica de ferro.", Price: 100, SellPrice: 40, DefenseBonus: 8, CABonus: 2, MinLevel: 1, ClassReq: "warrior", DropWeight: 18, Slot: "chest",
	},
	"breastplate_steel": {
		ID: "breastplate_steel", Name: "Peitoral de Aço", Emoji: "⚙️", Type: "armor", Rarity: models.RarityUncommon,
		Description: "Aço forjado para máxima proteção torácica.", Price: 300, SellPrice: 120, DefenseBonus: 18, CABonus: 3, MinLevel: 5, ClassReq: "warrior", DropWeight: 12, Slot: "chest",
	},
	"breastplate_heavy": {
		ID: "breastplate_heavy", Name: "Couraça Pesada", Emoji: "⚔️", Type: "armor", Rarity: models.RarityRare,
		Description: "Armadura de peito que resiste a qualquer golpe.", Price: 700, SellPrice: 280, DefenseBonus: 30, CABonus: 4, MinLevel: 9, ClassReq: "warrior", DropWeight: 7, Slot: "chest",
	},
	"cuirass_darksteel": {
		ID: "cuirass_darksteel", Name: "Couraça do Abismo", Emoji: "🖤", Type: "armor", Rarity: models.RarityEpic,
		Description: "Forjada com metal das profundezas, quase intransponível.", Price: 1800, SellPrice: 720, DefenseBonus: 45, CABonus: 5, MinLevel: 13, ClassReq: "warrior", DropWeight: 3, Slot: "chest",
	},
	"cuirass_dragon": {
		ID: "cuirass_dragon", Name: "Couraça de Dragão", Emoji: "🐉", Type: "armor", Rarity: models.RarityLegendary,
		Description: "Escamas do ventre do dragão. Proteção absoluta.", Price: 0, SellPrice: 4000, DefenseBonus: 60, CABonus: 7, HPBonus: 50, MinLevel: 17, ClassReq: "warrior", DropWeight: 1, Slot: "chest",
	},

	// ── CHEST: MAGE ──────────────────────────────────────────
	"robe_basic": {
		ID: "robe_basic", Name: "Robe Básico", Emoji: "👘", Type: "armor", Rarity: models.RarityCommon,
		Description: "Robe de linho com fios mágicos sutis.", Price: 80, SellPrice: 32, MagicDefBonus: 8, CABonus: 1, MinLevel: 1, ClassReq: "mage", DropWeight: 18, Slot: "chest",
	},
	"robe_scholar": {
		ID: "robe_scholar", Name: "Robe do Estudioso", Emoji: "📚", Type: "armor", Rarity: models.RarityUncommon,
		Description: "Proteção moderada com amplificação de magias.", Price: 280, SellPrice: 112, MagicDefBonus: 18, MagicAtkBonus: 8, CABonus: 2, MinLevel: 5, ClassReq: "mage", DropWeight: 12, Slot: "chest",
	},
	"robe_arcane": {
		ID: "robe_arcane", Name: "Robe Arcano", Emoji: "🌀", Type: "armor", Rarity: models.RarityRare,
		Description: "Robe tecido com fios de ley-lines.", Price: 680, SellPrice: 272, MagicDefBonus: 30, MagicAtkBonus: 15, CABonus: 3, MinLevel: 9, ClassReq: "mage", DropWeight: 7, Slot: "chest",
	},
	"robe_void": {
		ID: "robe_void", Name: "Robe do Vazio", Emoji: "🌌", Type: "armor", Rarity: models.RarityEpic,
		Description: "Trazido de dimensão paralela. Absorve magia hostil.", Price: 1700, SellPrice: 680, MagicDefBonus: 45, MagicAtkBonus: 22, CABonus: 4, MinLevel: 13, ClassReq: "mage", DropWeight: 3, Slot: "chest",
	},
	"robe_celestial": {
		ID: "robe_celestial", Name: "Robe Celestial", Emoji: "✨", Type: "armor", Rarity: models.RarityLegendary,
		Description: "Tece proteção divina ao redor do portador.", Price: 0, SellPrice: 3800, MagicDefBonus: 60, MagicAtkBonus: 32, CABonus: 5, MPBonus: 60, MinLevel: 17, ClassReq: "mage", DropWeight: 1, Slot: "chest",
	},

	// ── CHEST: ROGUE ─────────────────────────────────────────
	"vest_thief": {
		ID: "vest_thief", Name: "Colete do Ladrão", Emoji: "🥷", Type: "armor", Rarity: models.RarityCommon,
		Description: "Colete leve cheio de bolsos escondidos.", Price: 75, SellPrice: 30, DefenseBonus: 5, CABonus: 2, SpeedBonus: 1, MinLevel: 1, ClassReq: "rogue", DropWeight: 18, Slot: "chest",
	},
	"vest_shadow": {
		ID: "vest_shadow", Name: "Colete das Sombras", Emoji: "🌑", Type: "armor", Rarity: models.RarityUncommon,
		Description: "Tecido que absorve luz para passar despercebido.", Price: 260, SellPrice: 104, DefenseBonus: 12, CABonus: 3, SpeedBonus: 2, MinLevel: 5, ClassReq: "rogue", DropWeight: 12, Slot: "chest",
	},
	"jacket_assassin": {
		ID: "jacket_assassin", Name: "Jaqueta do Assassino", Emoji: "🖤", Type: "armor", Rarity: models.RarityRare,
		Description: "Silenciosa e resistente para missões letais.", Price: 620, SellPrice: 248, DefenseBonus: 20, CABonus: 4, SpeedBonus: 3, MinLevel: 9, ClassReq: "rogue", DropWeight: 7, Slot: "chest",
	},
	"shroud_night": {
		ID: "shroud_night", Name: "Manto da Noite", Emoji: "🌙", Type: "armor", Rarity: models.RarityEpic,
		Description: "Funde o portador com a escuridão.", Price: 1600, SellPrice: 640, DefenseBonus: 30, CABonus: 5, SpeedBonus: 5, MinLevel: 13, ClassReq: "rogue", DropWeight: 3, Slot: "chest",
	},
	"shroud_phantom": {
		ID: "shroud_phantom", Name: "Manto do Fantasma", Emoji: "👻", Type: "armor", Rarity: models.RarityLegendary,
		Description: "Torna o portador parcialmente intangível.", Price: 0, SellPrice: 3500, DefenseBonus: 40, CABonus: 6, SpeedBonus: 8, HPBonus: 20, MinLevel: 17, ClassReq: "rogue", DropWeight: 1, Slot: "chest",
	},

	// ── CHEST: ARCHER ────────────────────────────────────────
	"tunic_ranger": {
		ID: "tunic_ranger", Name: "Túnica do Ranger", Emoji: "🎽", Type: "armor", Rarity: models.RarityCommon,
		Description: "Túnica leve que não restringe os braços.", Price: 75, SellPrice: 30, DefenseBonus: 5, CABonus: 2, SpeedBonus: 1, MinLevel: 1, ClassReq: "archer", DropWeight: 18, Slot: "chest",
	},
	"vest_forest": {
		ID: "vest_forest", Name: "Colete da Floresta", Emoji: "🌿", Type: "armor", Rarity: models.RarityUncommon,
		Description: "Camuflado com padrões de floresta.", Price: 260, SellPrice: 104, DefenseBonus: 12, CABonus: 3, SpeedBonus: 2, MinLevel: 5, ClassReq: "archer", DropWeight: 12, Slot: "chest",
	},
	"jacket_hunt": {
		ID: "jacket_hunt", Name: "Jaqueta de Caça", Emoji: "🦺", Type: "armor", Rarity: models.RarityRare,
		Description: "Reforçada com couro de besta para suportar batalhas longas.", Price: 600, SellPrice: 240, DefenseBonus: 18, CABonus: 3, SpeedBonus: 3, HitBonus: 1, MinLevel: 9, ClassReq: "archer", DropWeight: 7, Slot: "chest",
	},
	"vest_storm": {
		ID: "vest_storm", Name: "Colete da Tempestade", Emoji: "⛈️", Type: "armor", Rarity: models.RarityEpic,
		Description: "Imbuído com estática elétrica que deflecte projéteis.", Price: 1600, SellPrice: 640, DefenseBonus: 26, CABonus: 4, SpeedBonus: 5, HitBonus: 1, MinLevel: 13, ClassReq: "archer", DropWeight: 3, Slot: "chest",
	},
	"mantle_eagle": {
		ID: "mantle_eagle", Name: "Manto da Águia", Emoji: "🦅", Type: "armor", Rarity: models.RarityLegendary,
		Description: "Tecido com penas de águia sagrada. Velocidade lendária.", Price: 0, SellPrice: 3500, DefenseBonus: 35, CABonus: 5, SpeedBonus: 10, HitBonus: 2, MinLevel: 17, ClassReq: "archer", DropWeight: 1, Slot: "chest",
	},

	// ── HANDS: WARRIOR ───────────────────────────────────────
	"gauntlets_spiked": {
		ID: "gauntlets_spiked", Name: "Manoplas Cravejadas", Emoji: "🔩", Type: "armor", Rarity: models.RarityRare,
		Description: "Espinhos de aço amplificam o dano corpo a corpo.", Price: 550, SellPrice: 220, AttackBonus: 15, DefenseBonus: 6, HitBonus: 1, MinLevel: 9, ClassReq: "warrior", DropWeight: 7, Slot: "hands",
	},
	"gauntlets_titan": {
		ID: "gauntlets_titan", Name: "Punhos do Titã", Emoji: "🏔️", Type: "armor", Rarity: models.RarityLegendary,
		Description: "Capazes de partir montanhas. Força sobre-humana.", Price: 0, SellPrice: 3000, AttackBonus: 38, DefenseBonus: 12, HitBonus: 2, MinLevel: 17, ClassReq: "warrior", DropWeight: 1, Slot: "hands",
	},

	// ── HANDS: MAGE ──────────────────────────────────────────
	"gloves_wool": {
		ID: "gloves_wool", Name: "Luvas de Lã Arcana", Emoji: "🧤", Type: "armor", Rarity: models.RarityCommon,
		Description: "Luvas que estabilizam fluxos de mana.", Price: 55, SellPrice: 22, MagicAtkBonus: 4, MagicDefBonus: 3, MinLevel: 1, ClassReq: "mage", DropWeight: 20, Slot: "hands",
	},
	"gloves_silk": {
		ID: "gloves_silk", Name: "Luvas de Seda Mística", Emoji: "🤲", Type: "armor", Rarity: models.RarityUncommon,
		Description: "Seda encantada que amplifica feitiços.", Price: 200, SellPrice: 80, MagicAtkBonus: 10, MagicDefBonus: 6, MinLevel: 5, ClassReq: "mage", DropWeight: 12, Slot: "hands",
	},
	"gloves_void": {
		ID: "gloves_void", Name: "Luvas do Vazio", Emoji: "🌌", Type: "armor", Rarity: models.RarityEpic,
		Description: "Tocam dimensões paralelas para amplificar poder.", Price: 1600, SellPrice: 640, MagicAtkBonus: 30, MagicDefBonus: 15, HitBonus: 1, MinLevel: 13, ClassReq: "mage", DropWeight: 3, Slot: "hands",
	},

	// ── HANDS: ROGUE ─────────────────────────────────────────
	"gloves_leather": {
		ID: "gloves_leather", Name: "Luvas de Couro", Emoji: "🧤", Type: "armor", Rarity: models.RarityCommon,
		Description: "Aderência perfeita para golpes precisos.", Price: 50, SellPrice: 20, AttackBonus: 3, HitBonus: 0, MinLevel: 1, ClassReq: "rogue", DropWeight: 20, Slot: "hands",
	},
	"gloves_grip": {
		ID: "gloves_grip", Name: "Luvas de Agarre", Emoji: "🤜", Type: "armor", Rarity: models.RarityUncommon,
		Description: "Maximizam a velocidade dos ataques.", Price: 190, SellPrice: 76, AttackBonus: 7, HitBonus: 1, MinLevel: 5, ClassReq: "rogue", DropWeight: 12, Slot: "hands",
	},
	"gloves_assassin": {
		ID: "gloves_assassin", Name: "Luvas do Assassino", Emoji: "💀", Type: "armor", Rarity: models.RarityEpic,
		Description: "Cobertos de veneno paralisante.", Price: 1400, SellPrice: 560, AttackBonus: 22, HitBonus: 2, SpeedBonus: 2, MinLevel: 13, ClassReq: "rogue", DropWeight: 3, Slot: "hands",
	},
	"gloves_void_rogue": {
		ID: "gloves_void_rogue", Name: "Garras do Vazio", Emoji: "🌌", Type: "armor", Rarity: models.RarityLegendary,
		Description: "Atravessam armaduras como névoa.", Price: 0, SellPrice: 2800, AttackBonus: 34, HitBonus: 3, SpeedBonus: 3, MinLevel: 17, ClassReq: "rogue", DropWeight: 1, Slot: "hands",
	},

	// ── HANDS: ARCHER ────────────────────────────────────────
	"bracer_basic": {
		ID: "bracer_basic", Name: "Braçadeira de Couro", Emoji: "💪", Type: "armor", Rarity: models.RarityCommon,
		Description: "Protege o braço do arco e melhora a precisão.", Price: 50, SellPrice: 20, AttackBonus: 3, HitBonus: 1, MinLevel: 1, ClassReq: "archer", DropWeight: 20, Slot: "hands",
	},
	"bracer_hunter": {
		ID: "bracer_hunter", Name: "Braçadeira do Caçador", Emoji: "🏹", Type: "armor", Rarity: models.RarityUncommon,
		Description: "Melhora precisão e velocidade de disparo.", Price: 190, SellPrice: 76, AttackBonus: 7, HitBonus: 1, SpeedBonus: 1, MinLevel: 5, ClassReq: "archer", DropWeight: 12, Slot: "hands",
	},
	"bracer_elven": {
		ID: "bracer_elven", Name: "Braçadeira Élfica", Emoji: "🌿", Type: "armor", Rarity: models.RarityRare,
		Description: "Encantada para precisão sobrenatural.", Price: 520, SellPrice: 208, AttackBonus: 14, HitBonus: 2, SpeedBonus: 1, MinLevel: 9, ClassReq: "archer", DropWeight: 7, Slot: "hands",
	},
	"bracer_storm": {
		ID: "bracer_storm", Name: "Braçadeira da Tempestade", Emoji: "⛈️", Type: "armor", Rarity: models.RarityEpic,
		Description: "Carregada com energia para flechas explosivas.", Price: 1400, SellPrice: 560, AttackBonus: 24, HitBonus: 2, SpeedBonus: 2, MinLevel: 13, ClassReq: "archer", DropWeight: 3, Slot: "hands",
	},
	"bracer_dragon": {
		ID: "bracer_dragon", Name: "Braçadeira do Dragão", Emoji: "🐉", Type: "armor", Rarity: models.RarityLegendary,
		Description: "Energia de dragão em cada flecha.", Price: 0, SellPrice: 2800, AttackBonus: 36, HitBonus: 3, SpeedBonus: 3, MinLevel: 17, ClassReq: "archer", DropWeight: 1, Slot: "hands",
	},

	// ── LEGS: WARRIOR ────────────────────────────────────────
	"greaves_iron": {
		ID: "greaves_iron", Name: "Grevas de Ferro", Emoji: "🦵", Type: "armor", Rarity: models.RarityCommon,
		Description: "Proteção básica para as pernas.", Price: 65, SellPrice: 26, DefenseBonus: 4, CABonus: 1, MinLevel: 1, ClassReq: "warrior", DropWeight: 20, Slot: "legs",
	},
	"greaves_steel": {
		ID: "greaves_steel", Name: "Grevas de Aço", Emoji: "⚙️", Type: "armor", Rarity: models.RarityUncommon,
		Description: "Grevas resistentes que suportam cargas pesadas.", Price: 220, SellPrice: 88, DefenseBonus: 10, CABonus: 2, MinLevel: 5, ClassReq: "warrior", DropWeight: 12, Slot: "legs",
	},
	"greaves_knight": {
		ID: "greaves_knight", Name: "Grevas do Cavaleiro", Emoji: "🏇", Type: "armor", Rarity: models.RarityRare,
		Description: "Proteção completa de pernas para cavaleiros.", Price: 550, SellPrice: 220, DefenseBonus: 18, CABonus: 3, MinLevel: 9, ClassReq: "warrior", DropWeight: 7, Slot: "legs",
	},
	"greaves_titan": {
		ID: "greaves_titan", Name: "Grevas do Titã", Emoji: "🗿", Type: "armor", Rarity: models.RarityEpic,
		Description: "Impossível de romper. Cada passo abala o chão.", Price: 1500, SellPrice: 600, DefenseBonus: 28, CABonus: 4, MinLevel: 13, ClassReq: "warrior", DropWeight: 3, Slot: "legs",
	},
	"greaves_dragon": {
		ID: "greaves_dragon", Name: "Grevas de Dragão", Emoji: "🐉", Type: "armor", Rarity: models.RarityLegendary,
		Description: "Escamas do flanco do dragão. Máxima proteção de pernas.", Price: 0, SellPrice: 2800, DefenseBonus: 38, CABonus: 5, HPBonus: 20, MinLevel: 17, ClassReq: "warrior", DropWeight: 1, Slot: "legs",
	},

	// ── LEGS: MAGE ───────────────────────────────────────────
	"leggings_arcane": {
		ID: "leggings_arcane", Name: "Calças Arcanas", Emoji: "👖", Type: "armor", Rarity: models.RarityCommon,
		Description: "Calças básicas com fios de proteção mágica.", Price: 55, SellPrice: 22, MagicDefBonus: 5, CABonus: 1, MinLevel: 1, ClassReq: "mage", DropWeight: 20, Slot: "legs",
	},
	"leggings_scholar": {
		ID: "leggings_scholar", Name: "Calças do Estudioso", Emoji: "📜", Type: "armor", Rarity: models.RarityUncommon,
		Description: "Com bolsos para pergaminhos e reagentes.", Price: 190, SellPrice: 76, MagicDefBonus: 12, MagicAtkBonus: 4, CABonus: 1, MinLevel: 5, ClassReq: "mage", DropWeight: 12, Slot: "legs",
	},
	"leggings_mystic": {
		ID: "leggings_mystic", Name: "Calças Místicas", Emoji: "🌀", Type: "armor", Rarity: models.RarityRare,
		Description: "Padrões rúnicos amplificam o fluxo de mana.", Price: 520, SellPrice: 208, MagicDefBonus: 20, MagicAtkBonus: 8, CABonus: 2, MinLevel: 9, ClassReq: "mage", DropWeight: 7, Slot: "legs",
	},
	"leggings_void": {
		ID: "leggings_void", Name: "Calças do Vazio", Emoji: "🌌", Type: "armor", Rarity: models.RarityEpic,
		Description: "Dimensionais — amplificam magia passiva.", Price: 1400, SellPrice: 560, MagicDefBonus: 30, MagicAtkBonus: 14, CABonus: 2, MinLevel: 13, ClassReq: "mage", DropWeight: 3, Slot: "legs",
	},
	"leggings_celestial": {
		ID: "leggings_celestial", Name: "Calças Celestiais", Emoji: "✨", Type: "armor", Rarity: models.RarityLegendary,
		Description: "Trazidas dos planos etéreos.", Price: 0, SellPrice: 2800, MagicDefBonus: 42, MagicAtkBonus: 20, CABonus: 3, MPBonus: 30, MinLevel: 17, ClassReq: "mage", DropWeight: 1, Slot: "legs",
	},

	// ── LEGS: ROGUE ──────────────────────────────────────────
	"pants_light": {
		ID: "pants_light", Name: "Calças Leves", Emoji: "👖", Type: "armor", Rarity: models.RarityCommon,
		Description: "Calças leves que facilitam movimentos rápidos.", Price: 50, SellPrice: 20, DefenseBonus: 3, SpeedBonus: 1, MinLevel: 1, ClassReq: "rogue", DropWeight: 20, Slot: "legs",
	},
	"pants_shadow": {
		ID: "pants_shadow", Name: "Calças das Sombras", Emoji: "🌑", Type: "armor", Rarity: models.RarityUncommon,
		Description: "Silenciosas e ágeis.", Price: 180, SellPrice: 72, DefenseBonus: 7, SpeedBonus: 2, CABonus: 1, MinLevel: 5, ClassReq: "rogue", DropWeight: 12, Slot: "legs",
	},
	"pants_assassin": {
		ID: "pants_assassin", Name: "Calças do Assassino", Emoji: "🥷", Type: "armor", Rarity: models.RarityRare,
		Description: "Permitem movimentos acrobáticos em combate.", Price: 480, SellPrice: 192, DefenseBonus: 12, SpeedBonus: 4, CABonus: 2, MinLevel: 9, ClassReq: "rogue", DropWeight: 7, Slot: "legs",
	},
	"pants_night": {
		ID: "pants_night", Name: "Calças da Noite", Emoji: "🌙", Type: "armor", Rarity: models.RarityEpic,
		Description: "Fundem portador com a escuridão.", Price: 1300, SellPrice: 520, DefenseBonus: 18, SpeedBonus: 6, CABonus: 3, MinLevel: 13, ClassReq: "rogue", DropWeight: 3, Slot: "legs",
	},
	"pants_phantom": {
		ID: "pants_phantom", Name: "Calças Fantasma", Emoji: "👻", Type: "armor", Rarity: models.RarityLegendary,
		Description: "Quase intangíveis. Velocidade máxima.", Price: 0, SellPrice: 2600, DefenseBonus: 25, SpeedBonus: 10, CABonus: 4, MinLevel: 17, ClassReq: "rogue", DropWeight: 1, Slot: "legs",
	},

	// ── LEGS: ARCHER ─────────────────────────────────────────
	"leggings_ranger": {
		ID: "leggings_ranger", Name: "Perneiras do Ranger", Emoji: "🌿", Type: "armor", Rarity: models.RarityCommon,
		Description: "Leves e duráveis para longas marchas.", Price: 50, SellPrice: 20, DefenseBonus: 3, SpeedBonus: 1, MinLevel: 1, ClassReq: "archer", DropWeight: 20, Slot: "legs",
	},
	"leggings_hunt": {
		ID: "leggings_hunt", Name: "Perneiras de Caça", Emoji: "🦺", Type: "armor", Rarity: models.RarityUncommon,
		Description: "Camufladas e confortáveis para esperar a presa.", Price: 180, SellPrice: 72, DefenseBonus: 7, SpeedBonus: 2, CABonus: 1, MinLevel: 5, ClassReq: "archer", DropWeight: 12, Slot: "legs",
	},
	"leggings_wind": {
		ID: "leggings_wind", Name: "Perneiras dos Ventos", Emoji: "💨", Type: "armor", Rarity: models.RarityRare,
		Description: "Encantadas com velocidade do vento.", Price: 480, SellPrice: 192, DefenseBonus: 11, SpeedBonus: 4, CABonus: 2, HitBonus: 1, MinLevel: 9, ClassReq: "archer", DropWeight: 7, Slot: "legs",
	},
	"leggings_storm": {
		ID: "leggings_storm", Name: "Perneiras da Tempestade", Emoji: "⛈️", Type: "armor", Rarity: models.RarityEpic,
		Description: "Velocidade de raio. Impossível de alcançar.", Price: 1300, SellPrice: 520, DefenseBonus: 16, SpeedBonus: 7, CABonus: 2, HitBonus: 1, MinLevel: 13, ClassReq: "archer", DropWeight: 3, Slot: "legs",
	},
	"leggings_eagle": {
		ID: "leggings_eagle", Name: "Perneiras da Águia", Emoji: "🦅", Type: "armor", Rarity: models.RarityLegendary,
		Description: "Penas sagradas permitem movimentos impossíveis.", Price: 0, SellPrice: 2600, DefenseBonus: 22, SpeedBonus: 10, CABonus: 3, HitBonus: 2, MinLevel: 17, ClassReq: "archer", DropWeight: 1, Slot: "legs",
	},

	// ── FEET: WARRIOR ────────────────────────────────────────
	"boots_knight": {
		ID: "boots_knight", Name: "Botas do Cavaleiro", Emoji: "⚔️", Type: "armor", Rarity: models.RarityRare,
		Description: "Botas de placa com esporões.", Price: 520, SellPrice: 208, DefenseBonus: 12, CABonus: 3, MinLevel: 9, ClassReq: "warrior", DropWeight: 7, Slot: "feet",
	},
	"boots_warlord": {
		ID: "boots_warlord", Name: "Botas do Senhor da Guerra", Emoji: "🏔️", Type: "armor", Rarity: models.RarityEpic,
		Description: "Cada passo esmaga o chão sob seus pés.", Price: 1400, SellPrice: 560, DefenseBonus: 18, CABonus: 3, AttackBonus: 5, MinLevel: 13, ClassReq: "warrior", DropWeight: 3, Slot: "feet",
	},

	// ── FEET: MAGE ───────────────────────────────────────────
	"slippers_arcane": {
		ID: "slippers_arcane", Name: "Sapatilhas Arcanas", Emoji: "🥿", Type: "armor", Rarity: models.RarityCommon,
		Description: "Sapatilhas com runas de proteção.", Price: 50, SellPrice: 20, MagicDefBonus: 4, CABonus: 1, MinLevel: 1, ClassReq: "mage", DropWeight: 20, Slot: "feet",
	},
	"boots_levitation": {
		ID: "boots_levitation", Name: "Botas de Levitação", Emoji: "🌀", Type: "armor", Rarity: models.RarityUncommon,
		Description: "Levitam levemente, melhorando evasão.", Price: 200, SellPrice: 80, MagicDefBonus: 9, CABonus: 2, SpeedBonus: 2, MinLevel: 5, ClassReq: "mage", DropWeight: 12, Slot: "feet",
	},
	"boots_void": {
		ID: "boots_void", Name: "Botas do Vazio", Emoji: "🌌", Type: "armor", Rarity: models.RarityEpic,
		Description: "Pisam em dimensões paralelas para escapar de ataques.", Price: 1400, SellPrice: 560, MagicDefBonus: 24, MagicAtkBonus: 12, CABonus: 3, SpeedBonus: 3, MinLevel: 13, ClassReq: "mage", DropWeight: 3, Slot: "feet",
	},

	// ── FEET: ROGUE ──────────────────────────────────────────
	"boots_silent": {
		ID: "boots_silent", Name: "Botas do Silêncio", Emoji: "👟", Type: "armor", Rarity: models.RarityCommon,
		Description: "Não fazem nenhum som ao caminhar.", Price: 50, SellPrice: 20, SpeedBonus: 2, CABonus: 1, MinLevel: 1, ClassReq: "rogue", DropWeight: 20, Slot: "feet",
	},
	"boots_shadow_step": {
		ID: "boots_shadow_step", Name: "Botas do Passo Sombrio", Emoji: "🌑", Type: "armor", Rarity: models.RarityUncommon,
		Description: "Deixam rastros de sombra no chão.", Price: 190, SellPrice: 76, SpeedBonus: 3, CABonus: 2, DefenseBonus: 3, MinLevel: 5, ClassReq: "rogue", DropWeight: 12, Slot: "feet",
	},
	"boots_phantom_step": {
		ID: "boots_phantom_step", Name: "Botas do Passo Fantasma", Emoji: "👻", Type: "armor", Rarity: models.RarityRare,
		Description: "Permite reposicionamento instantâneo.", Price: 500, SellPrice: 200, SpeedBonus: 5, CABonus: 3, DefenseBonus: 5, MinLevel: 9, ClassReq: "rogue", DropWeight: 7, Slot: "feet",
	},
	"boots_void_step": {
		ID: "boots_void_step", Name: "Botas do Passo do Vazio", Emoji: "🌌", Type: "armor", Rarity: models.RarityEpic,
		Description: "Atravessam pequenas distâncias instantaneamente.", Price: 1400, SellPrice: 560, SpeedBonus: 7, CABonus: 3, DefenseBonus: 8, HitBonus: 1, MinLevel: 13, ClassReq: "rogue", DropWeight: 3, Slot: "feet",
	},
	"boots_night": {
		ID: "boots_night", Name: "Botas da Noite Eterna", Emoji: "🌙", Type: "armor", Rarity: models.RarityLegendary,
		Description: "Desaparecem no escuro. Velocidade impossível.", Price: 0, SellPrice: 2500, SpeedBonus: 10, CABonus: 4, DefenseBonus: 12, HitBonus: 2, MinLevel: 17, ClassReq: "rogue", DropWeight: 1, Slot: "feet",
	},

	// ── FEET: ARCHER ─────────────────────────────────────────
	"boots_scout": {
		ID: "boots_scout", Name: "Botas do Escoteiro", Emoji: "🥾", Type: "armor", Rarity: models.RarityCommon,
		Description: "Leves e confortáveis para longas jornadas.", Price: 50, SellPrice: 20, SpeedBonus: 2, CABonus: 1, MinLevel: 1, ClassReq: "archer", DropWeight: 20, Slot: "feet",
	},
	"boots_forest": {
		ID: "boots_forest", Name: "Botas da Floresta", Emoji: "🌿", Type: "armor", Rarity: models.RarityRare,
		Description: "Se movem em silêncio pela floresta.", Price: 500, SellPrice: 200, SpeedBonus: 4, CABonus: 2, DefenseBonus: 5, HitBonus: 1, MinLevel: 9, ClassReq: "archer", DropWeight: 7, Slot: "feet",
	},
	"boots_storm": {
		ID: "boots_storm", Name: "Botas da Tempestade", Emoji: "⛈️", Type: "armor", Rarity: models.RarityEpic,
		Description: "Carregadas de eletricidade. Velocidade de raio.", Price: 1400, SellPrice: 560, SpeedBonus: 6, CABonus: 3, DefenseBonus: 8, HitBonus: 1, MinLevel: 13, ClassReq: "archer", DropWeight: 3, Slot: "feet",
	},
	"boots_dragon_wind": {
		ID: "boots_dragon_wind", Name: "Botas do Vento do Dragão", Emoji: "🐉", Type: "armor", Rarity: models.RarityLegendary,
		Description: "Sopro de dragão propulsiona cada passo.", Price: 0, SellPrice: 2500, SpeedBonus: 9, CABonus: 3, DefenseBonus: 12, HitBonus: 2, MinLevel: 17, ClassReq: "archer", DropWeight: 1, Slot: "feet",
	},

	// ── ACCESSORIES (universais — qualquer classe) ───────────
	// Anéis (accessory1 preferencial)
	"ring_bronze": {
		ID: "ring_bronze", Name: "Anel de Bronze", Emoji: "💍", Type: "armor", Rarity: models.RarityCommon,
		Description: "Anel simples com leve proteção mágica.", Price: 80, SellPrice: 32, MagicDefBonus: 4, MinLevel: 1, DropWeight: 15, Slot: "accessory1",
	},
	"ring_silver": {
		ID: "ring_silver", Name: "Anel de Prata", Emoji: "💍", Type: "armor", Rarity: models.RarityUncommon,
		Description: "Prata pura que repele energias negativas.", Price: 250, SellPrice: 100, MagicDefBonus: 9, DefenseBonus: 4, MinLevel: 5, DropWeight: 10, Slot: "accessory1",
	},
	"ring_gold": {
		ID: "ring_gold", Name: "Anel de Ouro", Emoji: "💍", Type: "armor", Rarity: models.RarityRare,
		Description: "Símbolo de poder. Aumenta atributos gerais.", Price: 700, SellPrice: 280, AttackBonus: 5, MagicAtkBonus: 5, DefenseBonus: 5, MagicDefBonus: 5, MinLevel: 9, DropWeight: 6, Slot: "accessory1",
	},
	"ring_dragon": {
		ID: "ring_dragon", Name: "Anel do Dragão", Emoji: "🐉", Type: "armor", Rarity: models.RarityEpic,
		Description: "Escama de dragão montada em ouro. Poder imenso.", Price: 1800, SellPrice: 720, AttackBonus: 10, MagicAtkBonus: 10, DefenseBonus: 8, MagicDefBonus: 8, HitBonus: 1, MinLevel: 13, DropWeight: 3, Slot: "accessory1",
	},
	"ring_void": {
		ID: "ring_void", Name: "Anel do Vazio", Emoji: "🌌", Type: "armor", Rarity: models.RarityLegendary,
		Description: "Conecta o portador às forças primordiais do universo.", Price: 0, SellPrice: 4000, AttackBonus: 18, MagicAtkBonus: 18, DefenseBonus: 12, MagicDefBonus: 12, HitBonus: 2, CABonus: 2, MinLevel: 17, DropWeight: 1, Slot: "accessory1",
	},

	// Colares/Amuletos (accessory2 preferencial)
	"necklace_bone": {
		ID: "necklace_bone", Name: "Colar de Ossos", Emoji: "🦴", Type: "armor", Rarity: models.RarityCommon,
		Description: "Ossos de monstros trazem pequena proteção.", Price: 70, SellPrice: 28, DefenseBonus: 3, HPBonus: 10, MinLevel: 1, DropWeight: 15, Slot: "accessory2",
	},
	"necklace_crystal": {
		ID: "necklace_crystal", Name: "Colar de Cristal", Emoji: "🔮", Type: "armor", Rarity: models.RarityUncommon,
		Description: "Cristal mágico que amplifica as energias vitais.", Price: 240, SellPrice: 96, HPBonus: 20, MPBonus: 15, MinLevel: 5, DropWeight: 10, Slot: "accessory2",
	},
	"necklace_gold": {
		ID: "necklace_gold", Name: "Colar Dourado", Emoji: "📿", Type: "armor", Rarity: models.RarityRare,
		Description: "Símbolo de nobreza com proteção abrangente.", Price: 680, SellPrice: 272, HPBonus: 30, MPBonus: 20, DefenseBonus: 5, MagicDefBonus: 8, MinLevel: 9, DropWeight: 6, Slot: "accessory2",
	},
	"amulet_eternity": {
		ID: "amulet_eternity", Name: "Amuleto da Eternidade", Emoji: "♾️", Type: "armor", Rarity: models.RarityLegendary,
		Description: "Forjado com tempo cristalizado. Poder absoluto.", Price: 0, SellPrice: 4500, AttackBonus: 15, MagicAtkBonus: 15, HPBonus: 60, MPBonus: 50, HitBonus: 2, CABonus: 1, MinLevel: 17, DropWeight: 1, Slot: "accessory2",
	},
}

// ── MONSTERS ──────────────────────────────────────────────
var Monsters = map[string]models.Monster{
	// Village area (lvl 1-3)
	"rat": {
		ID: "rat", Name: "Rato Gigante", Emoji: "🐀", Level: 1, HP: 15, CA: 11, Attack: 4, Defense: 2, MagicAtk: 0, MagicDef: 1, Speed: 6,
		ExpReward: 15, GoldReward: 3, DiamondChance: 0, MapID: "village_outskirts", Weakness: "physical",
		DropTable: map[string]int{
			"potion_small": 30, "gold_coins": 12,
			"lucky_charm": 14, "worn_ring": 12,
			"warrior_token": 6, "apprentice_focus": 6, "thief_token": 6, "hunter_feather": 6,
			"boots_iron": 5, "boots_cloth": 5, "boots_soft": 5, "boots_light": 5,
			"helmet_iron": 4, "hood_linen": 4, "mask_leather": 4, "hood_ranger": 4,
		},
	},
	"goblin": {
		ID: "goblin", Name: "Goblin Pilhador", Emoji: "👺", Level: 2, HP: 25, CA: 12, Attack: 7, Defense: 3, MagicAtk: 0, MagicDef: 2, Speed: 7,
		ExpReward: 30, GoldReward: 8, DiamondChance: 0, MapID: "village_outskirts", Weakness: "magic",
		DropTable: map[string]int{
			"potion_small": 22, "chest_wooden": 14,
			"dagger_iron": 12, "axe_iron": 12, "knife_bone": 10, "katar_iron": 10, "spear_iron": 10,
			"ring_iron": 10, "lucky_charm": 8, "worn_ring": 8, "iron_shield": 8, "shield_wooden": 8,
			"warrior_token": 6, "thief_token": 6,
			"gauntlets_iron": 6, "gloves_leather_rogue": 6, "bracers_basic": 6, "gloves_cloth": 6,
			"boots_iron": 5, "boots_soft": 5, "boots_light": 5, "boots_cloth": 5,
		},
	},
	"slime": {
		ID: "slime", Name: "Gosma Corrosiva", Emoji: "🟢", Level: 3, HP: 35, CA: 10, Attack: 5, Defense: 8, MagicAtk: 3, MagicDef: 1, Speed: 2, PoisonChance: 20, PoisonDmg: 4, PoisonTurns: 2,
		ExpReward: 35, GoldReward: 10, DiamondChance: 0, MapID: "village_outskirts", Weakness: "physical",
		DropTable: map[string]int{
			"potion_small": 18, "mana_potion_small": 18, "antidote": 14,
			"reinforced_leather": 12, "wand_basic": 10, "ring_iron": 10, "lucky_charm": 8, "apprentice_focus": 8,
			"scout_vest": 8, "light_vest": 8, "helmet_iron": 6, "hood_linen": 6, "mask_leather": 6, "hood_ranger": 6,
			"gauntlets_iron": 6, "gloves_cloth": 6, "gloves_leather_rogue": 6, "bracers_basic": 6,
			"boots_iron": 5, "boots_cloth": 5, "boots_soft": 5, "boots_light": 5,
		},
	},
	// Forest (lvl 4-7)
	"wolf": {
		ID: "wolf", Name: "Lobo das Sombras", Emoji: "🐺", Level: 4, HP: 50, CA: 13, Attack: 14, Defense: 5, MagicAtk: 0, MagicDef: 4, Speed: 12,
		ExpReward: 60, GoldReward: 15, DiamondChance: 2, MapID: "dark_forest", Weakness: "magic",
		DropTable: map[string]int{
			"potion_small": 16, "potion_medium": 14, "chest_wooden": 10,
			"leather_armor": 12, "mace_spiked": 10, "dagger_curved": 10, "crossbow": 10, "studded_leather": 10,
			"claw_steel": 8, "orb_frost": 8, "repeating_crossbow": 8,
			"amulet_luck": 8, "ring_iron": 7, "hunter_feather": 7,
			"helmet_steel": 7, "hood_arcane": 7, "mask_shadow": 7, "hood_hunter": 7,
			"gauntlets_steel": 6, "gloves_arcane": 6, "gloves_swift": 6, "bracers_hunter": 6,
			"boots_steel": 6, "boots_arcane": 6, "boots_swift": 6, "boots_tracker": 6,
		},
	},
	"orc": {
		ID: "orc", Name: "Orc Selvagem", Emoji: "👹", Level: 5, HP: 70, CA: 13, Attack: 18, Defense: 8, MagicAtk: 0, MagicDef: 3, Speed: 8,
		ExpReward: 80, GoldReward: 20, DiamondChance: 3, MapID: "dark_forest", Weakness: "magic",
		DropTable: map[string]int{
			"potion_medium": 14, "chest_wooden": 12,
			"sword_iron": 13, "mace_spiked": 12, "halberd": 10, "spear_steel": 10, "scale_armor": 10,
			"ring_strength": 10, "tome_fire": 10, "orb_frost": 10, "pendant_arcane": 9, "bracers_precision": 8, "ring_agility": 8,
			"helmet_steel": 8, "hood_arcane": 8, "mask_shadow": 8, "hood_hunter": 8,
			"gauntlets_steel": 7, "gloves_arcane": 7, "gloves_swift": 7, "bracers_hunter": 7,
			"boots_steel": 7, "boots_arcane": 7, "boots_swift": 7, "boots_tracker": 7,
			"shield_steel": 8,
		},
	},
	"troll": {
		ID: "troll", Name: "Troll da Floresta", Emoji: "🧌", Level: 6, HP: 90, CA: 12, Attack: 20, Defense: 12, MagicAtk: 0, MagicDef: 5, Speed: 5,
		ExpReward: 100, GoldReward: 30, DiamondChance: 5, MapID: "dark_forest", Weakness: "fire",
		DropTable: map[string]int{
			"potion_medium": 14, "chest_iron": 8,
			"chain_mail": 12, "battle_vest": 12, "halberd": 10, "spear_steel": 10,
			"ring_strength": 8, "ring_agility": 8, "bracers_precision": 8, "pendant_arcane": 8, "amulet_luck": 6,
			"helmet_steel": 9, "hood_arcane": 9, "mask_shadow": 9, "hood_hunter": 9,
			"gauntlets_steel": 8, "gloves_arcane": 8, "gloves_swift": 8, "bracers_hunter": 8,
			"boots_steel": 8, "boots_arcane": 8, "boots_swift": 8, "boots_tracker": 8,
			"shield_steel": 9,
		},
	},
	"bandit_leader": {
		ID: "bandit_leader", Name: "Líder Bandido", Emoji: "🦹", Level: 7, HP: 110, CA: 14, Attack: 22, Defense: 14, MagicAtk: 5, MagicDef: 8, Speed: 9, PoisonChance: 18, PoisonDmg: 6, PoisonTurns: 2,
		ExpReward: 130, GoldReward: 45, DiamondChance: 8, MapID: "dark_forest", Weakness: "physical",
		DropTable: map[string]int{
			"potion_medium": 12, "chest_iron": 10,
			"dagger_venom": 12, "halberd": 10, "mage_robe": 10, "ranger_vest": 10, "studded_leather": 10,
			"claw_steel": 9, "orb_frost": 9, "repeating_crossbow": 9,
			"amulet_luck": 7, "ring_agility": 7, "bracers_precision": 6,
			"helmet_steel": 8, "hood_arcane": 8, "mask_shadow": 8, "hood_hunter": 8,
			"gauntlets_steel": 7, "gloves_swift": 7, "bracers_hunter": 7, "gloves_arcane": 7,
			"boots_swift": 7, "boots_arcane": 7, "boots_tracker": 7, "boots_steel": 7,
		},
	},
	// Cave (lvl 8-12)
	"bat": {
		ID: "bat", Name: "Morcego Vampiro", Emoji: "🦇", Level: 8, HP: 120, CA: 13, Attack: 25, Defense: 10, MagicAtk: 8, MagicDef: 10, Speed: 15,
		ExpReward: 150, GoldReward: 35, DiamondChance: 5, MapID: "crystal_cave", Weakness: "physical",
		DropTable: map[string]int{
			"potion_large": 16, "mana_potion_large": 16,
			"shadow_cloak": 10, "blade_dual": 10, "bow_hunter": 10, "hunter_vest": 10,
			"ring_protection": 10, "enchanted_robe": 8, "knights_armor": 8,
			"tome_shadow": 8, "quiver_magic": 8, "claw_shadow": 8, "spear_thunder": 7,
			"helmet_knights": 8, "hood_crystal": 8, "mask_assassin": 8, "hood_elven": 8,
			"gauntlets_knights": 7, "gloves_crystal": 7, "gloves_poison": 7, "bracers_elven": 7,
			"boots_knights": 7, "boots_mystic": 7, "boots_shadow": 7, "boots_wind": 7,
			"shield_tower": 8,
		},
	},
	"spider": {
		ID: "spider", Name: "Aranha Gigante", Emoji: "🕷️", Level: 9, HP: 140, CA: 14, Attack: 28, Defense: 12, MagicAtk: 10, MagicDef: 8, Speed: 13, PoisonChance: 35, PoisonDmg: 8, PoisonTurns: 3,
		ExpReward: 175, GoldReward: 40, DiamondChance: 7, MapID: "crystal_cave", Weakness: "fire",
		DropTable: map[string]int{
			"antidote": 18, "chest_iron": 8,
			"dagger_shadow": 12, "orb_lightning": 10, "assassin_garb": 10, "knights_armor": 10,
			"ring_protection": 8, "venom_ring": 8, "quiver_charm": 8,
			"claw_shadow": 9, "tome_shadow": 9, "quiver_magic": 8,
			"helmet_knights": 8, "hood_crystal": 8, "mask_assassin": 8, "hood_elven": 8,
			"gauntlets_knights": 7, "gloves_crystal": 7, "gloves_poison": 7, "bracers_elven": 7,
			"boots_knights": 7, "boots_mystic": 7, "boots_shadow": 7, "boots_wind": 7,
		},
	},
	"golem": {
		ID: "golem", Name: "Golem de Pedra", Emoji: "🗿", Level: 11, HP: 200, CA: 16, Attack: 32, Defense: 25, MagicAtk: 0, MagicDef: 15, Speed: 3,
		ExpReward: 220, GoldReward: 60, DiamondChance: 10, MapID: "crystal_cave", Weakness: "magic",
		DropTable: map[string]int{
			"chest_iron": 14, "elixir": 10,
			"plate_armor": 12, "greatsword": 10, "warded_armor": 10, "serpent_fang": 9, "bow_ancient": 9,
			"spear_thunder": 9, "tome_shadow": 8, "quiver_magic": 8, "claw_shadow": 8,
			"berserker_band": 8, "lich_seal": 8, "venom_ring": 7, "quiver_charm": 7,
			"helmet_knights": 8, "hood_crystal": 8, "mask_assassin": 8, "hood_elven": 8,
			"gauntlets_knights": 7, "gloves_crystal": 7, "gloves_poison": 7, "bracers_elven": 7,
			"boots_knights": 7, "boots_mystic": 7, "boots_shadow": 7, "boots_wind": 7,
			"shield_tower": 8,
		},
	},
	"undead_knight": {
		ID: "undead_knight", Name: "Cavaleiro Morto", Emoji: "💀", Level: 12, HP: 180, CA: 17, Attack: 35, Defense: 20, MagicAtk: 15, MagicDef: 18, Speed: 7,
		ExpReward: 250, GoldReward: 70, DiamondChance: 12, MapID: "crystal_cave", Weakness: "holy",
		DropTable: map[string]int{
			"chest_gold":   8,
			"sword_silver": 10, "dark_plate": 8, "tome_ancient": 9, "necklace_war": 8,
			"ring_mage": 8, "shadow_ring": 8, "eagle_eye": 8,
			"berserker_band": 8, "lich_seal": 8, "venom_ring": 8, "quiver_charm": 8,
			"spear_thunder": 7, "tome_shadow": 7, "quiver_magic": 7,
			"helmet_knights": 8, "hood_crystal": 8, "mask_assassin": 8, "hood_elven": 8,
			"gauntlets_knights": 7, "gloves_crystal": 7, "gloves_poison": 7, "bracers_elven": 7,
			"boots_knights": 7, "boots_mystic": 7, "boots_shadow": 7, "boots_wind": 7,
			"shield_blessed": 5,
		},
	},
	// Dungeon (lvl 13-18)
	"demon": {
		ID: "demon", Name: "Demônio Menor", Emoji: "😈", Level: 13, HP: 250, CA: 16, Attack: 40, Defense: 22, MagicAtk: 30, MagicDef: 20, Speed: 11,
		ExpReward: 300, GoldReward: 90, DiamondChance: 15, MapID: "ancient_dungeon", Weakness: "holy",
		DropTable: map[string]int{
			"chest_gold": 14, "elixir": 12,
			"stiletto": 9, "staff_inferno": 9, "bow_shadow": 9, "fortress_plate": 8, "void_robe": 8,
			"tome_arcane": 7, "bow_phantom": 7, "sword_runic": 6,
			"amulet_power": 8, "nightmare_cloak": 8, "wind_vest": 8,
			"helmet_warlord": 7, "hood_void": 7, "mask_void": 7, "hood_storm": 7,
			"gauntlets_warlord": 6, "gloves_void_mage": 6, "gloves_shadow": 6, "bracers_storm": 6,
			"boots_titan": 6, "boots_void_mage": 6, "boots_phantom": 6, "boots_tempest": 6,
			"shield_blessed": 5,
		},
	},
	"necromancer": {
		ID: "necromancer", Name: "Necromante", Emoji: "🧟", Level: 15, HP: 200, CA: 14, Attack: 25, Defense: 15, MagicAtk: 45, MagicDef: 28, Speed: 8,
		ExpReward: 380, GoldReward: 110, DiamondChance: 18, MapID: "ancient_dungeon", Weakness: "physical",
		DropTable: map[string]int{
			"mana_potion_supreme": 10, "chest_gold": 10,
			"staff_crystal": 9, "void_robe": 8, "archmage_focus": 8, "ring_champion": 8, "eldritch_plate": 8,
			"tome_arcane": 7, "bow_phantom": 7,
			"sword_runic": 6, "staff_elder": 6, "dagger_reaper": 6, "bow_celestial": 6,
			"helmet_warlord": 7, "hood_void": 7, "mask_void": 7, "hood_storm": 7,
			"gauntlets_warlord": 6, "gloves_void_mage": 6, "gloves_shadow": 6, "bracers_storm": 6,
			"boots_titan": 6, "boots_void_mage": 6, "boots_phantom": 6, "boots_tempest": 6,
		},
	},
	"vampire_lord": {
		ID: "vampire_lord", Name: "Lorde Vampiro", Emoji: "🧛", Level: 17, HP: 320, CA: 17, Attack: 50, Defense: 30, MagicAtk: 40, MagicDef: 35, Speed: 14, PoisonChance: 25, PoisonDmg: 14, PoisonTurns: 3,
		ExpReward: 480, GoldReward: 140, DiamondChance: 22, MapID: "ancient_dungeon", Weakness: "holy",
		DropTable: map[string]int{
			"chest_gold": 10, "elixir_divine": 4,
			"sword_darksteel": 8, "void_cloak": 8, "war_hammer": 8, "nightmare_cloak": 8, "void_emblem": 7, "hawkeye_charm": 7,
			"guardian_plate": 6, "arcane_robe": 6, "shadow_shroud": 6, "storm_mantle": 6,
			"tome_arcane": 6, "bow_phantom": 6,
			"helmet_warlord": 6, "hood_void": 6, "mask_void": 6, "hood_storm": 6,
			"gauntlets_warlord": 5, "gloves_void_mage": 5, "gloves_shadow": 5, "bracers_storm": 5,
			"boots_titan": 5, "boots_void_mage": 5, "boots_phantom": 5, "boots_tempest": 5,
			"crown_warlord": 2, "eye_archmage": 2, "phantom_seal": 2, "starshot_pendant": 2,
		},
	},
	// Dragon Peak (lvl 18-20)
	"dragon_young": {
		ID: "dragon_young", Name: "Dragão Jovem", Emoji: "🐉", Level: 18, HP: 450, CA: 18, Attack: 65, Defense: 40, MagicAtk: 50, MagicDef: 42, Speed: 10,
		ExpReward: 650, GoldReward: 200, DiamondChance: 30, MapID: "dragon_peak", Weakness: "ice",
		DropTable: map[string]int{
			"chest_dragon": 15, "elixir_divine": 5,
			"bow_storm": 8, "storm_robe": 8, "titan_armor": 5, "celestial_robe": 5, "eagle_mantle": 5, "bow_phantom": 6,
			"helmet_dragon": 5, "hood_celestial": 5, "mask_phantom": 5, "hood_eagle": 5,
			"gauntlets_dragon": 4, "gloves_archmage": 4, "gloves_reaper": 4, "bracers_dragon": 4,
			"boots_dragon": 4, "boots_celestial": 4, "boots_void_rogue": 4, "boots_starwalker": 4,
			"shield_aegis":  4,
			"crown_warlord": 4, "eye_archmage": 4, "phantom_seal": 4, "starshot_pendant": 4, "amulet_ancients": 3,
		},
	},
	"dragon_elder": {
		ID: "dragon_elder", Name: "Dragão Ancião", Emoji: "🐲", Level: 20, HP: 600, CA: 20, Attack: 80, Defense: 55, MagicAtk: 65, MagicDef: 55, Speed: 12,
		ExpReward: 1000, GoldReward: 350, DiamondChance: 50, MapID: "dragon_peak", Weakness: "ice",
		DropTable: map[string]int{
			"chest_dragon": 20, "elixir_divine": 5,
			"sword_dragonslayer": 6, "staff_dragon": 6, "dragon_armor": 6,
			"arcane_mantle": 5, "phantom_cloak": 5, "titan_armor": 4, "celestial_robe": 4, "eagle_mantle": 4,
			"helmet_dragon": 6, "hood_celestial": 6, "mask_phantom": 6, "hood_eagle": 6,
			"gauntlets_dragon": 5, "gloves_archmage": 5, "gloves_reaper": 5, "bracers_dragon": 5,
			"boots_dragon": 5, "boots_celestial": 5, "boots_void_rogue": 5, "boots_starwalker": 5,
			"shield_aegis":  5,
			"crown_warlord": 4, "eye_archmage": 4, "phantom_seal": 4, "starshot_pendant": 4, "amulet_ancients": 3,
		},
	},

	// ── NOVOS MONSTROS (5 por zona) ─────────────────────────

	// Village (novos - lvl 1-3)
	"mushroom": {
		ID: "mushroom", Name: "Cogumelo Venenoso", Emoji: "🍄", Level: 2, HP: 20, CA: 10, Attack: 5, Defense: 4, MagicAtk: 4, MagicDef: 2, Speed: 1, PoisonChance: 25, PoisonDmg: 3, PoisonTurns: 2,
		ExpReward: 20, GoldReward: 5, DiamondChance: 0, MapID: "village_outskirts", Weakness: "fire",
		DropTable: map[string]int{
			"antidote": 28, "potion_small": 14, "mana_potion_small": 10,
			"wand_basic": 10, "apprentice_robe": 10, "apprentice_focus": 8, "lucky_charm": 8, "worn_ring": 7,
			"hood_linen": 7, "mask_leather": 7, "hood_ranger": 7, "helmet_iron": 7,
			"gloves_cloth": 6, "gloves_leather_rogue": 6, "bracers_basic": 6, "gauntlets_iron": 6,
			"boots_cloth": 5, "boots_soft": 5, "boots_light": 5, "boots_iron": 5,
		},
	},
	"crow": {
		ID: "crow", Name: "Corvo das Trevas", Emoji: "🐦‍⬛", Level: 3, HP: 28, CA: 12, Attack: 8, Defense: 3, MagicAtk: 2, MagicDef: 3, Speed: 14,
		ExpReward: 32, GoldReward: 7, DiamondChance: 1, MapID: "village_outskirts", Weakness: "physical",
		DropTable: map[string]int{
			"potion_small": 22, "gold_coins": 18,
			"light_vest": 12, "scout_vest": 12, "hunter_feather": 10, "thief_token": 8, "amulet_luck": 7,
			"hood_linen": 8, "mask_leather": 8, "hood_ranger": 8, "helmet_iron": 7,
			"gloves_cloth": 6, "gloves_leather_rogue": 6, "bracers_basic": 6, "gauntlets_iron": 6,
			"boots_cloth": 5, "boots_soft": 5, "boots_light": 5, "boots_iron": 5,
		},
	},

	// Forest (novos - lvl 4-7)
	"harpy": {
		ID: "harpy", Name: "Harpia Selvagem", Emoji: "🦅", Level: 5, HP: 60, CA: 13, Attack: 16, Defense: 6, MagicAtk: 5, MagicDef: 5, Speed: 16,
		ExpReward: 75, GoldReward: 18, DiamondChance: 3, MapID: "dark_forest", Weakness: "physical",
		DropTable: map[string]int{
			"potion_medium": 18, "chest_wooden": 14,
			"bow_long": 12, "crossbow": 12, "ranger_vest": 10, "repeating_crossbow": 10,
			"ring_agility": 9, "bracers_precision": 9, "amulet_luck": 6,
			"hood_hunter": 8, "hood_arcane": 8, "mask_shadow": 8, "helmet_steel": 8,
			"bracers_hunter": 7, "gloves_swift": 7, "gloves_arcane": 7, "gauntlets_steel": 7,
			"boots_tracker": 7, "boots_swift": 7, "boots_arcane": 7, "boots_steel": 7,
		},
	},
	"werewolf": {
		ID: "werewolf", Name: "Lobisomem", Emoji: "🐺", Level: 7, HP: 100, CA: 14, Attack: 24, Defense: 10, MagicAtk: 0, MagicDef: 6, Speed: 14,
		ExpReward: 120, GoldReward: 40, DiamondChance: 6, MapID: "dark_forest", Weakness: "holy",
		DropTable: map[string]int{
			"potion_large": 12, "chest_iron": 10,
			"halberd": 12, "battle_vest": 12, "studded_leather": 10, "spear_steel": 10,
			"ring_strength": 8, "ring_agility": 8, "bracers_precision": 7, "claw_steel": 8,
			"helmet_steel": 8, "hood_arcane": 8, "mask_shadow": 8, "hood_hunter": 8,
			"gauntlets_steel": 7, "gloves_swift": 7, "gloves_arcane": 7, "bracers_hunter": 7,
			"boots_steel": 7, "boots_swift": 7, "boots_arcane": 7, "boots_tracker": 7,
		},
	},

	// Cave (novos - lvl 8-12)
	"stone_golem_shard": {
		ID: "stone_golem_shard", Name: "Fragmento de Golem", Emoji: "🪨", Level: 9, HP: 130, CA: 15, Attack: 26, Defense: 22, MagicAtk: 0, MagicDef: 12, Speed: 2,
		ExpReward: 160, GoldReward: 38, DiamondChance: 6, MapID: "crystal_cave", Weakness: "magic",
		DropTable: map[string]int{
			"chest_iron": 12, "elixir": 8, "potion_large": 10,
			"knights_armor": 10, "warded_armor": 10, "ring_protection": 9,
			"tome_shadow": 8, "quiver_magic": 8, "claw_shadow": 8, "spear_thunder": 7,
			"berserker_band": 8, "lich_seal": 7, "venom_ring": 7, "quiver_charm": 7,
			"helmet_knights": 8, "hood_crystal": 8, "mask_assassin": 8, "hood_elven": 8,
			"gauntlets_knights": 7, "gloves_crystal": 7, "gloves_poison": 7, "bracers_elven": 7,
			"boots_knights": 7, "boots_mystic": 7, "boots_shadow": 7, "boots_wind": 7,
			"shield_tower": 8,
		},
	},
	"crystal_wraith": {
		ID: "crystal_wraith", Name: "Espectro de Cristal", Emoji: "👻", Level: 11, HP: 155, CA: 15, Attack: 30, Defense: 8, MagicAtk: 28, MagicDef: 20, Speed: 11,
		ExpReward: 200, GoldReward: 55, DiamondChance: 9, MapID: "crystal_cave", Weakness: "holy",
		DropTable: map[string]int{
			"mana_potion_large": 14, "chest_iron": 8,
			"staff_crystal": 10, "mystic_robe": 9, "tome_ancient": 9, "ring_mage": 8,
			"tome_shadow": 8, "quiver_magic": 8, "claw_shadow": 7,
			"berserker_band": 7, "lich_seal": 7, "venom_ring": 7, "quiver_charm": 7,
			"helmet_knights": 8, "hood_crystal": 9, "mask_assassin": 8, "hood_elven": 8,
			"gauntlets_knights": 7, "gloves_crystal": 8, "gloves_poison": 7, "bracers_elven": 7,
			"boots_knights": 7, "boots_mystic": 8, "boots_shadow": 7, "boots_wind": 7,
		},
	},

	// Dungeon (novos - lvl 13-18)
	"shadow_assassin": {
		ID: "shadow_assassin", Name: "Assassino das Sombras", Emoji: "🗡️", Level: 14, HP: 220, CA: 16, Attack: 45, Defense: 18, MagicAtk: 15, MagicDef: 22, Speed: 18, PoisonChance: 30, PoisonDmg: 12, PoisonTurns: 3,
		ExpReward: 340, GoldReward: 100, DiamondChance: 16, MapID: "ancient_dungeon", Weakness: "holy",
		DropTable: map[string]int{
			"chest_gold": 12, "elixir": 8,
			"dagger_shadow": 10, "stiletto": 9, "nightmare_cloak": 9, "void_emblem": 8, "shadow_ring": 8,
			"tome_arcane": 6, "bow_phantom": 6,
			"sword_runic": 5, "dagger_reaper": 5, "shadow_shroud": 5, "storm_mantle": 5,
			"helmet_warlord": 7, "hood_void": 7, "mask_void": 8, "hood_storm": 7,
			"gauntlets_warlord": 6, "gloves_void_mage": 6, "gloves_shadow": 7, "bracers_storm": 6,
			"boots_titan": 6, "boots_void_mage": 6, "boots_phantom": 7, "boots_tempest": 6,
		},
	},
	"lich": {
		ID: "lich", Name: "Lich Ancestral", Emoji: "💀", Level: 16, HP: 280, CA: 16, Attack: 30, Defense: 20, MagicAtk: 55, MagicDef: 35, Speed: 7,
		ExpReward: 430, GoldReward: 125, DiamondChance: 20, MapID: "ancient_dungeon", Weakness: "physical",
		DropTable: map[string]int{
			"chest_gold": 12, "mana_potion_supreme": 10, "elixir_divine": 5,
			"staff_crystal": 9, "storm_robe": 8, "void_robe": 8, "archmage_focus": 7,
			"tome_arcane": 8, "bow_phantom": 6,
			"staff_elder": 6, "guardian_plate": 5, "arcane_robe": 5, "shadow_shroud": 5,
			"helmet_warlord": 7, "hood_void": 8, "mask_void": 7, "hood_storm": 7,
			"gauntlets_warlord": 6, "gloves_void_mage": 7, "gloves_shadow": 6, "bracers_storm": 6,
			"boots_titan": 6, "boots_void_mage": 7, "boots_phantom": 6, "boots_tempest": 6,
			"eye_archmage": 3, "phantom_seal": 3,
		},
	},

	// Dragon Peak (novos - lvl 18-20)
	"wyvern": {
		ID: "wyvern", Name: "Wyvern de Chamas", Emoji: "🦎", Level: 18, HP: 420, CA: 17, Attack: 60, Defense: 35, MagicAtk: 40, MagicDef: 38, Speed: 13,
		ExpReward: 600, GoldReward: 180, DiamondChance: 28, MapID: "dragon_peak", Weakness: "ice",
		DropTable: map[string]int{
			"chest_dragon": 12, "elixir_divine": 6,
			"storm_robe": 8, "war_hammer": 8, "wind_vest": 8, "ring_champion": 7,
			"guardian_plate": 7, "storm_mantle": 7, "bow_celestial": 7, "bow_phantom": 6,
			"helmet_dragon": 4, "hood_celestial": 4, "mask_phantom": 4, "hood_eagle": 4,
			"gauntlets_dragon": 3, "gloves_archmage": 3, "gloves_reaper": 3, "bracers_dragon": 3,
			"boots_dragon": 3, "boots_celestial": 3, "boots_void_rogue": 3, "boots_starwalker": 3,
			"shield_aegis":  3,
			"crown_warlord": 3, "phantom_seal": 3, "starshot_pendant": 3,
		},
	},
	"phoenix": {
		ID: "phoenix", Name: "Fênix Imortal", Emoji: "🔥", Level: 19, HP: 500, CA: 18, Attack: 70, Defense: 45, MagicAtk: 60, MagicDef: 50, Speed: 15,
		ExpReward: 800, GoldReward: 280, DiamondChance: 40, MapID: "dragon_peak", Weakness: "ice",
		DropTable: map[string]int{
			"chest_dragon": 18, "elixir_divine": 8,
			"arcane_mantle": 8, "bow_storm": 8, "celestial_robe": 7, "eagle_mantle": 7, "bow_phantom": 6,
			"helmet_dragon": 5, "hood_celestial": 5, "mask_phantom": 5, "hood_eagle": 5,
			"gauntlets_dragon": 4, "gloves_archmage": 4, "gloves_reaper": 4, "bracers_dragon": 4,
			"boots_dragon": 4, "boots_celestial": 4, "boots_void_rogue": 4, "boots_starwalker": 4,
			"shield_aegis":  4,
			"crown_warlord": 4, "eye_archmage": 4, "phantom_seal": 4, "starshot_pendant": 4, "amulet_ancients": 3,
		},
	},
}

// ── MAPS ──────────────────────────────────────────────────
var Maps = map[string]models.GameMap{
	"village": {
		ID: "village", Name: "Vila de Trifort", Emoji: "🏘️",
		Description: "Uma pequena vila no interior do reino. Ponto de partida para aventureiros.",
		MinLevel:    1, MaxLevel: 99, ConnectsTo: []string{"village_outskirts", "dark_forest"},
		HasShop: true, HasInn: true,
	},
	"village_outskirts": {
		ID: "village_outskirts", Name: "Arredores da Vila", Emoji: "🌾",
		Description: "Campos e bosques ao redor da vila. Criaturas fracas habitam aqui.",
		MinLevel:    1, MaxLevel: 5, ConnectsTo: []string{"village", "dark_forest"},
		Monsters: []string{"rat", "goblin", "slime", "mushroom", "crow"},
		HasShop:  false, HasInn: false,
	},
	"dark_forest": {
		ID: "dark_forest", Name: "Floresta Sombria", Emoji: "🌲",
		Description: "Floresta densa onde lobos e orcs espreitam nas sombras.",
		MinLevel:    4, MaxLevel: 9, ConnectsTo: []string{"village", "crystal_cave"},
		Monsters: []string{"wolf", "orc", "troll", "bandit_leader", "harpy", "werewolf"},
		HasShop:  false, HasInn: false,
	},
	"crystal_cave": {
		ID: "crystal_cave", Name: "Caverna de Cristal", Emoji: "💎",
		Description: "Caverna iluminada por cristais. Habitada por criaturas mortais.",
		MinLevel:    8, MaxLevel: 13, ConnectsTo: []string{"dark_forest", "ancient_dungeon"},
		Monsters: []string{"bat", "spider", "golem", "undead_knight", "stone_golem_shard", "crystal_wraith"},
		HasShop:  true, HasInn: false,
	},
	"ancient_dungeon": {
		ID: "ancient_dungeon", Name: "Masmorra Antiga", Emoji: "🏚️",
		Description: "Ruínas de uma fortaleza tomada por demônios.",
		MinLevel:    13, MaxLevel: 18, ConnectsTo: []string{"crystal_cave", "dragon_peak"},
		Monsters: []string{"demon", "necromancer", "vampire_lord", "shadow_assassin", "lich"},
		HasShop:  false, HasInn: true,
	},
	"dragon_peak": {
		ID: "dragon_peak", Name: "Pico dos Dragões", Emoji: "🏔️",
		Description: "O cume onde os dragões mais antigos habitam.",
		MinLevel:    17, MaxLevel: 20, ConnectsTo: []string{"ancient_dungeon"},
		Monsters: []string{"dragon_young", "dragon_elder", "wyvern", "phoenix"},
		HasShop:  false, HasInn: false,
	},
}

// ── HELPER FUNCTIONS ──────────────────────────────────────

// GetAvailableMapsForHunt returns maps with monsters that the player can hunt in.
func GetAvailableMapsForHunt(level int) []models.GameMap {
	order := []string{"village_outskirts", "dark_forest", "crystal_cave", "ancient_dungeon", "dragon_peak"}
	var result []models.GameMap
	for _, id := range order {
		m := Maps[id]
		if len(m.Monsters) > 0 && level >= m.MinLevel {
			result = append(result, m)
		}
	}
	return result
}

func GetSkillsForClass(class string) []models.Skill {
	var result []models.Skill
	for _, s := range Skills {
		if s.Class == class {
			result = append(result, s)
		}
	}
	return result
}

func GetMonstersForMap(mapID string) []models.Monster {
	m, ok := Maps[mapID]
	if !ok {
		return nil
	}
	var result []models.Monster
	for _, mID := range m.Monsters {
		if monster, ok := Monsters[mID]; ok {
			result = append(result, monster)
		}
	}
	return result
}

func GetShopItemsForMap(_ string, level int) []models.Item {
	var result []models.Item
	for _, item := range Items {
		if item.Price > 0 && item.MinLevel <= level {
			result = append(result, item)
		}
	}
	return result
}

func ExperienceForLevel(level int) int {
	return level * level * 50
}

func CalculateBaseStats(race string, class string) (hp, mp, atk, def, matk, mdef, speed int) {
	r := Races[race]
	c := Classes[class]
	hp = c.BaseHP + r.BonusHP + r.BonusCon*5
	mp = c.BaseMP + r.BonusMP + r.BonusInt*3
	atk = c.BaseAttack + r.BonusStr*2
	def = c.BaseDefense + r.BonusCon
	matk = 4 + r.BonusInt*2
	mdef = 4 + r.BonusWis
	speed = 5 + r.BonusDex
	return
}

// ── CA (CLASSE DE ARMADURA) ───────────────────────────────
//
// Fórmula Tormenta 20 adaptada:
//   CA = 10 + modificador de atributo defensivo + bônus de armadura
//
// Por classe:
//   Guerreiro: 10 + mod(CON)  — armaduras pesadas, mais robusto
//   Mago:      10 + mod(INT)  — escudos arcanos, usa inteligência
//   Ladino:    10 + mod(DEX)  — esquiva ágil, usa destreza
//   Arqueiro:  10 + mod(DEX)  — mobilidade e distância
//
// Modificador de atributo = (atributo - 10) / 2  (mínimo -2)

func attrMod(attr int) int {
	mod := (attr - 10) / 2
	if mod < -2 {
		mod = -2
	}
	return mod
}

// CharacterCA calcula a CA total do personagem.
// charAttr = atributo defensivo primário (CON para guerreiro, INT para mago, DEX para ladino/arqueiro)
// equipCA = soma dos CABonus dos equipamentos vestidos
func CharacterCA(class string, charAttr int, equipCA int) int {
	base := 10
	mod := attrMod(charAttr)
	// Bônus base por classe (representa proficiência com armadura)
	classMod := map[string]int{
		"warrior": 3, // armaduras pesadas
		"mage":    0, // sem armadura física
		"rogue":   2, // armaduras leves
		"archer":  2, // armaduras leves
	}
	return base + mod + classMod[class] + equipCA
}

// DefensiveAttr retorna o atributo defensivo principal de cada classe.
func DefensiveAttr(class string, con, dex, intel int) int {
	switch class {
	case "warrior":
		return con
	case "mage":
		return intel
	case "rogue", "archer":
		return dex
	}
	return con
}

// AttackBonus calcula o bônus de ataque no d20 do personagem.
// Representa proficiência + modificador de atributo ofensivo.
// Guerreiro/Ladino: FOR; Mago: INT; Arqueiro: DES
// profBonus cresce com o nível (como Tormenta 20)
func CharacterAttackBonus(class string, level, str, dex, intel int) int {
	var attrB int
	switch class {
	case "warrior":
		attrB = attrMod(str)
	case "mage":
		attrB = attrMod(intel)
	case "rogue", "archer":
		attrB = attrMod(dex)
	}
	// Bônus de proficiência: +2 nos níveis 1-4, +3 nos 5-8, +4 nos 9-12, +5 nos 13-16, +6 nos 17-20
	prof := 1 + (level-1)/4
	return prof + attrB
}

// DamageRoll calcula o dado de dano base por classe/arma.
// Retorna (diceCount, diceSides, attrMod) para exibição clara no log.
// Guerreiro: 1d10+FOR;  Mago: 1d6+INT (mágico);  Ladino: 1d8+DES;  Arqueiro: 1d8+DES
func ClassDamageRoll(class string, str, dex, intel int) (dice, sides, mod int) {
	switch class {
	case "warrior":
		return 1, 10, attrMod(str)
	case "mage":
		return 1, 6, attrMod(intel)
	case "rogue":
		return 1, 8, attrMod(dex)
	case "archer":
		return 1, 8, attrMod(dex)
	}
	return 1, 6, 0
}

// MonsterAttackBonus retorna o bônus de ataque do monstro no d20.
// Baseia-se no nível do monstro como proxy de CR.
func MonsterAttackBonus(monster *models.Monster) int {
	return 1 + monster.Level/3
}

// MonsterDamageRoll rola o dado de dano do monstro.
// Monstros mais fortes têm dados maiores.
func MonsterDamageDice(monster *models.Monster) (dice, sides int) {
	switch {
	case monster.Level <= 3:
		return 1, 4
	case monster.Level <= 6:
		return 1, 6
	case monster.Level <= 10:
		return 1, 8
	case monster.Level <= 14:
		return 2, 6
	case monster.Level <= 17:
		return 2, 8
	default:
		return 2, 10
	}
}

// ── DROP SYSTEM ───────────────────────────────────────────

type DropResult struct {
	ItemID   string
	Quantity int
	Gold     int // bonus gold from chests
}

// RollDrops determines what items drop from a monster.
// Returns 0-2 items based on luck and the monster's drop table.
func RollDrops(monster *models.Monster, charLevel int) []DropResult {
	if len(monster.DropTable) == 0 {
		return nil
	}

	var results []DropResult

	// Base drop chance: 55% + 3% per level above monster
	levelBonus := (charLevel - monster.Level) * 3
	dropChance := 55 + levelBonus
	if dropChance > 85 {
		dropChance = 85
	}
	if dropChance < 25 {
		dropChance = 25
	}

	if rand.Intn(100) >= dropChance {
		return nil
	}

	// Pick item using weighted random
	totalWeight := 0
	for _, w := range monster.DropTable {
		totalWeight += w
	}
	if totalWeight == 0 {
		return nil
	}

	roll := rand.Intn(totalWeight)
	cumulative := 0
	var pickedItem string
	for itemID, weight := range monster.DropTable {
		cumulative += weight
		if roll < cumulative {
			pickedItem = itemID
			break
		}
	}
	if pickedItem == "" {
		return nil
	}

	item := Items[pickedItem]
	qty := 1
	if item.Type == "consumable" {
		qty = rand.Intn(2) + 1 // 1-2 consumables
	}

	results = append(results, DropResult{ItemID: pickedItem, Quantity: qty})
	return results
}

// OpenChest returns gold and optional item from a chest
func OpenChest(chestID string, _ int) (gold int, itemID string) {
	switch chestID {
	case "chest_wooden":
		gold = 20 + rand.Intn(61) // 20-80
		// 40% chance for a common/uncommon item
		pool := []string{"ring_iron", "amulet_luck", "reinforced_leather", "axe_iron", "knife_bone", "sling", "wand_basic", "light_vest", "scout_vest"}
		if rand.Intn(100) < 40 {
			itemID = pool[rand.Intn(len(pool))]
		}
	case "chest_iron":
		gold = 80 + rand.Intn(121) // 80-200
		// 65% chance for an uncommon/rare item
		pool := []string{
			"potion_medium", "mana_potion_large", "antidote", "elixir",
			"mace_spiked", "dagger_curved", "crossbow", "tome_fire",
			"scale_armor", "mage_robe", "studded_leather", "ranger_vest", "battle_vest",
			"ring_strength", "pendant_arcane", "ring_agility", "bracers_precision", "amulet_luck",
		}
		if rand.Intn(100) < 65 {
			itemID = pool[rand.Intn(len(pool))]
		}
	case "chest_gold":
		gold = 200 + rand.Intn(301) // 200-500
		// 75% chance for a rare/epic item
		pool := []string{
			"elixir", "plate_armor", "mystic_robe", "shadow_cloak", "chain_mail",
			"halberd", "orb_lightning", "blade_dual", "bow_hunter",
			"knights_armor", "enchanted_robe", "assassin_garb", "hunter_vest", "warded_armor",
			"ring_protection", "necklace_war", "ring_mage", "shadow_ring", "eagle_eye", "amulet_power",
		}
		if rand.Intn(100) < 75 {
			itemID = pool[rand.Intn(len(pool))]
		}
	case "chest_dragon":
		gold = 500 + rand.Intn(1001) // 500-1500
		// 85% chance for epic/legendary item
		pool := []string{
			"elixir_divine", "dark_plate", "void_cloak", "storm_robe", "sword_darksteel", "staff_void", "dagger_assassin", "bow_storm",
			"war_hammer", "staff_inferno", "stiletto", "bow_shadow",
			"fortress_plate", "void_robe", "nightmare_cloak", "wind_vest", "eldritch_plate",
			"ring_champion", "archmage_focus", "void_emblem", "hawkeye_charm",
		}
		legendary := []string{
			"dragon_armor", "arcane_mantle", "phantom_cloak", "sword_dragonslayer", "staff_dragon", "bow_dragon", "dagger_void",
			"titan_armor", "celestial_robe", "eagle_mantle", "amulet_ancients",
		}
		if rand.Intn(100) < 25 {
			itemID = legendary[rand.Intn(len(legendary))]
		} else if rand.Intn(100) < 85 {
			itemID = pool[rand.Intn(len(pool))]
		}
	}
	return
}
