package rpgdata

import "github.com/tormenta-bot/internal/models"

// ─── Races ────────────────────────────────────────────────────────────────────
//
// Each Race provides flat attribute bonuses applied at character creation and a
// unique Trait string that game systems can reference.  The ID matches the key
// used in game.Races so that the loader can update existing entries or add new
// ones without duplication.
//
// Stat philosophy:
//
//   Human      – balanced, small bonus everywhere, XP advantage
//   Elf        – DEX + INT, fragile but magical
//   Dwarf      – CON king, negligible DEX/INT
//   Meio-Orc   – raw STR/CON powerhouse, negative INT/CHA
//   Goblin     – DEX scout, low STR, bonus gold
//   Qareen     – magic-aligned, high INT/WIS/CHA, no melee bonus
//   Minotauro  – heaviest tank in the roster, negative everything else

// AllRaces contains all playable races keyed by their canonical ID.
// These entries are MERGED into game.Races by the loader in game/rpgdata_loader.go.
var AllRaces = map[string]models.Race{
	"human": {
		ID: "human", Name: "Humano", Emoji: "👤",
		Description: "Adaptáveis e versáteis. Bônus em todas as classes.",
		BonusHP:     10, BonusMP: 5,
		BonusStr: 1, BonusDex: 1, BonusCon: 1,
		BonusInt: 1, BonusWis: 1, BonusCha: 2,
		Trait: "Versátil: +10% de XP ganha; +1 em todos os atributos no nível 50",
	},
	"elf": {
		ID: "elf", Name: "Elfo", Emoji: "🧝",
		Description: "Ágeis e conectados à magia. Poucos superam a graça élfica.",
		BonusHP:     5, BonusMP: 20,
		BonusStr: 0, BonusDex: 3, BonusCon: 0,
		BonusInt: 3, BonusWis: 2, BonusCha: 1,
		Trait: "Visão Élfica: +20% dano com magia; +5% esquiva em combate",
	},
	"dwarf": {
		ID: "dwarf", Name: "Anão", Emoji: "⛏️",
		Description: "Resistentes como as pedras que moldam. Incomparáveis nas forjas.",
		BonusHP:     30, BonusMP: 0,
		BonusStr: 2, BonusDex: 0, BonusCon: 5,
		BonusInt: 0, BonusWis: 1, BonusCha: 0,
		Trait: "Pele de Pedra: -15% dano recebido; +10% eficiência na forja",
	},
	"halforc": {
		ID: "halforc", Name: "Meio-Orc", Emoji: "👹",
		Description: "Brutal em combate, com força descomunal e sangue de orc.",
		BonusHP:     20, BonusMP: 0,
		BonusStr: 5, BonusDex: 0, BonusCon: 3,
		BonusInt: -1, BonusWis: 0, BonusCha: -1,
		Trait: "Fúria do Sangue: +25% dano físico; ressurge com 1 HP uma vez por combate",
	},
	"goblin": {
		ID: "goblin", Name: "Goblin", Emoji: "👺",
		Description: "Pequenos, astutos e incrivelmente sortudos na hora de negociar.",
		BonusHP:     5, BonusMP: 10,
		BonusStr: -1, BonusDex: 4, BonusCon: 0,
		BonusInt: 2, BonusWis: 0, BonusCha: 0,
		Trait: "Esperteza: +15% ouro encontrado; +10% chance de encontrar item raro",
	},
	"qareen": {
		ID: "qareen", Name: "Qareen", Emoji: "🧞",
		Description: "Seres mágicos do deserto profundo, ligados a forças elementais.",
		BonusHP:     10, BonusMP: 30,
		BonusStr: 0, BonusDex: 1, BonusCon: 0,
		BonusInt: 3, BonusWis: 3, BonusCha: 4,
		Trait: "Magia Elemental: +20% dano mágico; resistência a fogo e trovão",
	},
	"minotaur": {
		ID: "minotaur", Name: "Minotauro", Emoji: "🐂",
		Description: "Guerreiros imponentes com chifres capazes de derrubar muralhas.",
		BonusHP:     40, BonusMP: 0,
		BonusStr: 6, BonusDex: -1, BonusCon: 5,
		BonusInt: -2, BonusWis: 0, BonusCha: -2,
		Trait: "Investida: +30% dano no primeiro ataque por combate; imune a knockback",
	},
}
