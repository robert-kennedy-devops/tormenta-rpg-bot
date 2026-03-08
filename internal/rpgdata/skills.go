package rpgdata

import (
	"fmt"
	"strings"

	"github.com/tormenta-bot/internal/models"
)

// ─── Skill system ─────────────────────────────────────────────────────────────
//
// 10 classes × 3 branches × 4 tiers = 120 core skills.
// Additional utility/passive skills bring the total above 130.
//
// Skill tier → required level:
//   Tier 1 → levels  1–20  (PointCost 1)
//   Tier 2 → levels 21–40  (PointCost 2)
//   Tier 3 → levels 41–70  (PointCost 3)
//   Tier 4 → levels 71–100 (PointCost 4)
//
// Damage formula used by combat:
//   finalDmg = skill.Damage + (statScaling × primaryStat) + weaponBonus
//   where statScaling and primaryStat are class-specific.
//
// To add a new skill:
//   1. Add a skillDef entry to the class's 'skills' slice inside skillClassDefs.
//   2. Re-run `go build ./...` — AllSkills is populated automatically.

// skillDef is the intermediate definition used by the generator before it
// produces a models.Skill value.
type skillDef struct {
	id          string
	name        string
	emoji       string
	branch      string
	tier        int    // 1–4
	mpCost      int
	damage      int
	damageType  string // "physical" | "magic" | "buff" | "passive" | "utility"
	passive     bool
	description string
	// optional
	poisonDmg   int
	poisonTurns int
}

// skillClassDef groups a class's skill definitions.
type skillClassDef struct {
	classID string
	skills  []skillDef
}

// firstRequiredLevel maps skill tier to the minimum required level.
var firstRequiredLevel = [5]int{0, 1, 21, 41, 71}

// tierPointCost maps skill tier to the skill-point cost to learn it.
var tierPointCost = [5]int{0, 1, 2, 3, 4}

// skillClassDefs holds the raw definitions for all 10 classes.
// The generator transforms these into models.Skill entries with automatic
// RequiredLevel, PointCost, Requires chain, and Class fields.
var skillClassDefs = []skillClassDef{

	// ── Guerreiro ──────────────────────────────────────────────────────────────
	{classID: "warrior", skills: []skillDef{
		// Branch: Protetor (defensive)
		{id: "w_iron_skin", name: "Pele de Ferro", emoji: "🛡️", branch: "protetor", tier: 1, mpCost: 0, damage: 0, damageType: "passive", passive: true, description: "Passiva: +3 CA permanente. Seu corpo endurece com o combate."},
		{id: "w_shield_bash", name: "Escudada", emoji: "🪃", branch: "protetor", tier: 2, mpCost: 12, damage: 20, damageType: "physical", description: "Golpe com o escudo. Dano físico + reduz CA do inimigo por 2 rodadas."},
		{id: "w_fortress", name: "Fortaleza", emoji: "🏰", branch: "protetor", tier: 3, mpCost: 25, damage: 0, damageType: "buff", description: "+5 CA e -30% dano recebido por 3 rodadas."},
		{id: "w_divine_guard", name: "Guarda Divina", emoji: "⛪", branch: "protetor", tier: 4, mpCost: 40, damage: 0, damageType: "buff", description: "Imunidade a dano por 2 rodadas. 1 uso por combate."},
		// Branch: Berserker (damage)
		{id: "w_power_strike", name: "Golpe Brutal", emoji: "⚔️", branch: "berserker", tier: 1, mpCost: 8, damage: 22, damageType: "physical", description: "Golpe pesado com bônus de FOR."},
		{id: "w_whirlwind", name: "Redemoinho", emoji: "🌀", branch: "berserker", tier: 2, mpCost: 20, damage: 35, damageType: "physical", description: "Ataca todos os inimigos próximos em arco."},
		{id: "w_berserker_rage", name: "Fúria Berserker", emoji: "😡", branch: "berserker", tier: 3, mpCost: 30, damage: 0, damageType: "buff", description: "+50% dano e -15% defesa por 3 rodadas."},
		{id: "w_titans_blow", name: "Golpe do Titã", emoji: "💥", branch: "berserker", tier: 4, mpCost: 50, damage: 120, damageType: "physical", description: "Ataque devastador. Ignora 30% da defesa do inimigo."},
		// Branch: Duelist (balanced)
		{id: "w_riposte", name: "Riposte", emoji: "🤺", branch: "duelista", tier: 1, mpCost: 6, damage: 14, damageType: "physical", description: "Contra-ataque rápido após esquivar. +15% chance de crítico."},
		{id: "w_parry", name: "Aparar", emoji: "🛡️", branch: "duelista", tier: 2, mpCost: 10, damage: 0, damageType: "buff", description: "Bloqueia próximo ataque e contra-ataca por 50% do dano."},
		{id: "w_blade_mastery", name: "Maestria da Lâmina", emoji: "⚔️", branch: "duelista", tier: 3, mpCost: 0, damage: 0, damageType: "passive", passive: true, description: "Passiva: +15% dano com espadas e machados."},
		{id: "w_champions_strike", name: "Golpe do Campeão", emoji: "🏆", branch: "duelista", tier: 4, mpCost: 55, damage: 100, damageType: "physical", description: "Ataque preciso que garante acerto crítico se o d20 ≥ 15."},
	}},

	// ── Paladino ───────────────────────────────────────────────────────────────
	{classID: "paladin", skills: []skillDef{
		// Branch: Sagrado (holy damage)
		{id: "pa_holy_strike", name: "Golpe Sagrado", emoji: "✨", branch: "sagrado", tier: 1, mpCost: 10, damage: 18, damageType: "magic", description: "Golpe abençoado que causa dano sagrado."},
		{id: "pa_smite", name: "Martelo da Justiça", emoji: "⚡", branch: "sagrado", tier: 2, mpCost: 22, damage: 40, damageType: "magic", description: "Chama poder divino. +100% dano contra mortos-vivos."},
		{id: "pa_consecration", name: "Consagração", emoji: "🕯️", branch: "sagrado", tier: 3, mpCost: 35, damage: 25, damageType: "magic", description: "Área sagrada. Dano contínuo em inimigos por 3 rodadas."},
		{id: "pa_divine_wrath", name: "Ira Divina", emoji: "☀️", branch: "sagrado", tier: 4, mpCost: 60, damage: 130, damageType: "magic", description: "Canaliza toda a fé. Dano sagrado massivo."},
		// Branch: Proteção (tank/heal)
		{id: "pa_lay_hands", name: "Imposição de Mãos", emoji: "🙏", branch: "protecao", tier: 1, mpCost: 12, damage: 0, damageType: "buff", description: "Cura 40 HP. Aumenta para 80 HP com WIS alta."},
		{id: "pa_divine_shield", name: "Escudo Divino", emoji: "🛡️", branch: "protecao", tier: 2, mpCost: 20, damage: 0, damageType: "buff", description: "Absorve próximos 60 pontos de dano."},
		{id: "pa_aura_of_valor", name: "Aura de Valor", emoji: "🌟", branch: "protecao", tier: 3, mpCost: 0, damage: 0, damageType: "passive", passive: true, description: "Passiva: +10% CA e +5% resistência mágica."},
		{id: "pa_sacred_armor", name: "Armadura Sagrada", emoji: "⛩️", branch: "protecao", tier: 4, mpCost: 45, damage: 0, damageType: "buff", description: "Durante 5 rodadas: +8 CA, reflete 20% do dano recebido."},
		// Branch: Redenção (hybrid)
		{id: "pa_blessed_weapon", name: "Arma Abençoada", emoji: "🗡️", branch: "redencao", tier: 1, mpCost: 8, damage: 0, damageType: "buff", description: "Envolve a arma em luz sagrada por 3 rodadas (+20% dano)."},
		{id: "pa_holy_aura", name: "Aura Santa", emoji: "😇", branch: "redencao", tier: 2, mpCost: 18, damage: 0, damageType: "buff", description: "Regenera 5 HP por rodada por 4 rodadas."},
		{id: "pa_crusader", name: "Cruzado", emoji: "⚔️", branch: "redencao", tier: 3, mpCost: 30, damage: 50, damageType: "physical", description: "Ataque físico + sagrado combinado. Bonus em criaturas do mal."},
		{id: "pa_avatar", name: "Avatar Divino", emoji: "👼", branch: "redencao", tier: 4, mpCost: 70, damage: 0, damageType: "buff", description: "Por 4 rodadas: +30% todos os atributos, HP regenera 10% por turno."},
	}},

	// ── Bárbaro ────────────────────────────────────────────────────────────────
	{classID: "barbarian", skills: []skillDef{
		// Branch: Fúria (pure damage)
		{id: "ba_rage", name: "Fúria", emoji: "😡", branch: "furia", tier: 1, mpCost: 5, damage: 0, damageType: "buff", description: "Ativa Fúria: +30% dano físico, -10% defesa por 4 rodadas."},
		{id: "ba_cleave", name: "Clivagem", emoji: "🪓", branch: "furia", tier: 2, mpCost: 15, damage: 45, damageType: "physical", description: "Golpe amplo que acerta em arco. Alta eficácia vs grupos."},
		{id: "ba_bloodthirst", name: "Sede de Sangue", emoji: "🩸", branch: "furia", tier: 3, mpCost: 0, damage: 0, damageType: "passive", passive: true, description: "Passiva: cada acerto durante Fúria cura 5 HP."},
		{id: "ba_rampage", name: "Amok", emoji: "🌪️", branch: "furia", tier: 4, mpCost: 0, damage: 80, damageType: "physical", description: "Série de 4 ataques rápidos durante Fúria. Cada um rola d20 independente."},
		// Branch: Selvagem (primal)
		{id: "ba_war_cry", name: "Grito de Guerra", emoji: "📢", branch: "selvagem", tier: 1, mpCost: 8, damage: 0, damageType: "buff", description: "Grito aterrorizante: reduz CA do inimigo em 3 por 2 rodadas."},
		{id: "ba_reckless", name: "Ataque Imprudente", emoji: "⚡", branch: "selvagem", tier: 2, mpCost: 0, damage: 55, damageType: "physical", description: "Ataque poderoso: vantagem no d20, mas próximo ataque do inimigo tem vantagem."},
		{id: "ba_primal_force", name: "Força Primordial", emoji: "🦍", branch: "selvagem", tier: 3, mpCost: 25, damage: 0, damageType: "buff", description: "+20 STR por 3 rodadas (bônus em dano físico)."},
		{id: "ba_titan_slam", name: "Slam do Titã", emoji: "💢", branch: "selvagem", tier: 4, mpCost: 40, damage: 150, damageType: "physical", description: "O ataque físico mais poderoso do jogo. Atordoa o inimigo por 1 rodada."},
		// Branch: Resistência (defensive)
		{id: "ba_toughness", name: "Robustez", emoji: "🏋️", branch: "resistencia", tier: 1, mpCost: 0, damage: 0, damageType: "passive", passive: true, description: "Passiva: +50 HP máximo. Aumenta 10 HP a cada 10 níveis."},
		{id: "ba_endurance", name: "Resistência", emoji: "💪", branch: "resistencia", tier: 2, mpCost: 10, damage: 0, damageType: "buff", description: "Reduz todo dano recebido em 15 por 3 rodadas."},
		{id: "ba_relentless", name: "Implacável", emoji: "♾️", branch: "resistencia", tier: 3, mpCost: 0, damage: 0, damageType: "passive", passive: true, description: "Passiva: sobrevive com 1 HP uma vez por combate (CD 5 turnos)."},
		{id: "ba_immortal_rage", name: "Fúria Imortal", emoji: "🔥", branch: "resistencia", tier: 4, mpCost: 30, damage: 0, damageType: "buff", description: "Fúria especial: +80% HP temporário, +40% dano, dura 5 rodadas."},
	}},

	// ── Ladino ─────────────────────────────────────────────────────────────────
	{classID: "rogue", skills: []skillDef{
		// Branch: Assassino
		{id: "ro_backstab", name: "Apunhalada pelas Costas", emoji: "🗡️", branch: "assassino", tier: 1, mpCost: 8, damage: 28, damageType: "physical", description: "Ataque furtivo. +100% dano se o inimigo não atacou na última rodada."},
		{id: "ro_eviscerate", name: "Evisceração", emoji: "🩸", branch: "assassino", tier: 2, mpCost: 18, damage: 45, damageType: "physical", description: "Golpe profundo. Aplica sangramento: 5 dano por rodada por 3 turnos.", poisonDmg: 5, poisonTurns: 3},
		{id: "ro_shadow_strike", name: "Golpe das Sombras", emoji: "🌑", branch: "assassino", tier: 3, mpCost: 28, damage: 70, damageType: "physical", description: "Emerge das sombras para um golpe letal. Ignora 40% da defesa."},
		{id: "ro_death_mark", name: "Marca da Morte", emoji: "☠️", branch: "assassino", tier: 4, mpCost: 45, damage: 110, damageType: "physical", description: "Marca o inimigo: por 3 rodadas recebe +30% dano de todas as fontes."},
		// Branch: Esquiva
		{id: "ro_dodge", name: "Esquivar", emoji: "💨", branch: "esquiva", tier: 1, mpCost: 5, damage: 0, damageType: "buff", description: "+30% chance de esquivar por 2 rodadas."},
		{id: "ro_smoke_bomb", name: "Bomba de Fumaça", emoji: "💣", branch: "esquiva", tier: 2, mpCost: 15, damage: 0, damageType: "utility", description: "Reduz acerto do inimigo em 50% por 2 rodadas."},
		{id: "ro_phantomstep", name: "Passo Fantasma", emoji: "👻", branch: "esquiva", tier: 3, mpCost: 22, damage: 0, damageType: "buff", description: "Se esquivar, contra-ataca automaticamente por 50% do ataque base."},
		{id: "ro_untouchable", name: "Intocável", emoji: "🌫️", branch: "esquiva", tier: 4, mpCost: 40, damage: 0, damageType: "buff", description: "Por 3 rodadas, cada esquiva cura 15 HP e garante próximo ataque crítico."},
		// Branch: Veneno
		{id: "ro_poison_blade", name: "Lâmina Envenenada", emoji: "☠️", branch: "veneno", tier: 1, mpCost: 6, damage: 10, damageType: "physical", description: "Aplica veneno: 8 dano por rodada por 3 turnos.", poisonDmg: 8, poisonTurns: 3},
		{id: "ro_toxic_cloud", name: "Nuvem Tóxica", emoji: "🌫️", branch: "veneno", tier: 2, mpCost: 16, damage: 0, damageType: "utility", description: "Área de veneno denso. Inimigo sofre 12 dano/rodada por 4 turnos.", poisonDmg: 12, poisonTurns: 4},
		{id: "ro_master_toxin", name: "Toxina Mestre", emoji: "🧪", branch: "veneno", tier: 3, mpCost: 0, damage: 0, damageType: "passive", passive: true, description: "Passiva: todos os venenos duram 2 rodadas a mais."},
		{id: "ro_deadly_venom", name: "Veneno Mortal", emoji: "💀", branch: "veneno", tier: 4, mpCost: 35, damage: 20, damageType: "physical", description: "Veneno imparável: 25 dano/rodada por 6 turnos. Não pode ser curado.", poisonDmg: 25, poisonTurns: 6},
	}},

	// ── Arcanista ──────────────────────────────────────────────────────────────
	{classID: "mage", skills: []skillDef{
		// Branch: Elementalista
		{id: "mg_fireball", name: "Bola de Fogo", emoji: "🔥", branch: "elementalista", tier: 1, mpCost: 14, damage: 32, damageType: "magic", description: "Projétil de fogo explosivo. Dano de área ao impactar."},
		{id: "mg_chain_lightning", name: "Relâmpago em Cadeia", emoji: "⚡", branch: "elementalista", tier: 2, mpCost: 25, damage: 50, damageType: "magic", description: "Relâmpago que salta entre alvos. Cada pulo perde 10% de dano."},
		{id: "mg_blizzard", name: "Ventania", emoji: "❄️", branch: "elementalista", tier: 3, mpCost: 40, damage: 40, damageType: "magic", description: "Tempestade de gelo por 3 rodadas. Cada rodada causa dano e reduz velocidade."},
		{id: "mg_meteor", name: "Meteoro", emoji: "☄️", branch: "elementalista", tier: 4, mpCost: 70, damage: 160, damageType: "magic", description: "Convoca um meteoro. Maior dano único do jogo. Atordoa por 1 turno."},
		// Branch: Arcanista
		{id: "mg_arcane_bolt", name: "Projétil Arcano", emoji: "🔮", branch: "arcanista", tier: 1, mpCost: 10, damage: 24, damageType: "magic", description: "Projétil puro de energia arcana. Ignora resistência elemental."},
		{id: "mg_arcane_surge", name: "Surto Arcano", emoji: "💫", branch: "arcanista", tier: 2, mpCost: 20, damage: 0, damageType: "buff", description: "+40% dano mágico por 3 rodadas. Consome MP extra a cada turno."},
		{id: "mg_spell_mastery", name: "Maestria Mágica", emoji: "📚", branch: "arcanista", tier: 3, mpCost: 0, damage: 0, damageType: "passive", passive: true, description: "Passiva: -20% custo de MP em todas as magias."},
		{id: "mg_arcane_nova", name: "Nova Arcana", emoji: "🌟", branch: "arcanista", tier: 4, mpCost: 80, damage: 140, damageType: "magic", description: "Explosão arcana em área. Reduz CA de todos inimigos atingidos em 5."},
		// Branch: Ilusionista
		{id: "mg_mirror_image", name: "Imagem Espelho", emoji: "🪞", branch: "ilusionista", tier: 1, mpCost: 12, damage: 0, damageType: "buff", description: "Cria uma cópia ilusória. 30% chance de redirecionar ataques para a cópia."},
		{id: "mg_confusion", name: "Confusão", emoji: "😵", branch: "ilusionista", tier: 2, mpCost: 22, damage: 0, damageType: "utility", description: "Confunde o inimigo: 40% chance de ele atacar a si mesmo por 2 rodadas."},
		{id: "mg_phantasm", name: "Fantasma", emoji: "👻", branch: "ilusionista", tier: 3, mpCost: 35, damage: 55, damageType: "magic", description: "Ilusão assassina. Causa dano psíquico que ignora armadura completamente."},
		{id: "mg_grand_illusion", name: "Grande Ilusão", emoji: "🎭", branch: "ilusionista", tier: 4, mpCost: 60, damage: 0, damageType: "buff", description: "Torna-se intocável por 3 rodadas e cria 3 cópias que atacam o inimigo."},
	}},

	// ── Clérigo ────────────────────────────────────────────────────────────────
	{classID: "cleric", skills: []skillDef{
		// Branch: Curandeiro
		{id: "cl_heal", name: "Curar", emoji: "💚", branch: "curandeiro", tier: 1, mpCost: 12, damage: 0, damageType: "buff", description: "Restaura 50 HP. Aumenta com WIS alta."},
		{id: "cl_mass_heal", name: "Cura em Massa", emoji: "💛", branch: "curandeiro", tier: 2, mpCost: 25, damage: 0, damageType: "buff", description: "Recupera 80 HP e remove 1 efeito negativo."},
		{id: "cl_rejuvenation", name: "Rejuvenescimento", emoji: "🌱", branch: "curandeiro", tier: 3, mpCost: 30, damage: 0, damageType: "buff", description: "Regenera 15 HP por rodada por 4 turnos."},
		{id: "cl_miracle", name: "Milagre", emoji: "✨", branch: "curandeiro", tier: 4, mpCost: 60, damage: 0, damageType: "buff", description: "Restaura HP para 75% do máximo. Remove todos os efeitos negativos."},
		// Branch: Sagrado
		{id: "cl_holy_smite", name: "Punição Sagrada", emoji: "☀️", branch: "sagrado", tier: 1, mpCost: 10, damage: 22, damageType: "magic", description: "Dano sagrado. +50% dano vs criaturas malignas."},
		{id: "cl_turn_undead", name: "Expulsar Mortos-Vivos", emoji: "⚰️", branch: "sagrado", tier: 2, mpCost: 20, damage: 60, damageType: "magic", description: "Devastador vs mortos-vivos. Afasta criaturas de nível menor."},
		{id: "cl_divine_favor", name: "Favor Divino", emoji: "🙏", branch: "sagrado", tier: 3, mpCost: 28, damage: 0, damageType: "buff", description: "+25% a todos os atributos por 3 rodadas."},
		{id: "cl_holy_storm", name: "Tempestade Sagrada", emoji: "⛈️", branch: "sagrado", tier: 4, mpCost: 65, damage: 120, damageType: "magic", description: "Chuva de luz divina. Dano sagrado massivo em área."},
		// Branch: Protetor
		{id: "cl_bless", name: "Abençoar", emoji: "✝️", branch: "protetor", tier: 1, mpCost: 8, damage: 0, damageType: "buff", description: "+3 em todos os testes de ataque por 3 rodadas."},
		{id: "cl_sanctuary", name: "Santuário", emoji: "🕍", branch: "protetor", tier: 2, mpCost: 18, damage: 0, damageType: "buff", description: "Escudo sagrado que absorve até 50 de dano."},
		{id: "cl_aegis", name: "Égide", emoji: "🛡️", branch: "protetor", tier: 3, mpCost: 32, damage: 0, damageType: "passive", passive: true, description: "Passiva: +10% a todas as resistências."},
		{id: "cl_divine_bastion", name: "Bastião Divino", emoji: "🏰", branch: "protetor", tier: 4, mpCost: 50, damage: 0, damageType: "buff", description: "Por 4 rodadas: imune a morte, absorve 100 de dano, reflete 25%."},
	}},

	// ── Bardo ──────────────────────────────────────────────────────────────────
	{classID: "bard", skills: []skillDef{
		// Branch: Inspiração
		{id: "bd_inspire", name: "Inspirar", emoji: "🎵", branch: "inspiracao", tier: 1, mpCost: 8, damage: 0, damageType: "buff", description: "+15% dano e +5% acerto por 3 rodadas."},
		{id: "bd_battle_hymn", name: "Hino de Batalha", emoji: "🎶", branch: "inspiracao", tier: 2, mpCost: 18, damage: 0, damageType: "buff", description: "Múscula de combate: +20% ATK, velocidade e CA por 3 rodadas."},
		{id: "bd_crescendo", name: "Crescendo", emoji: "📯", branch: "inspiracao", tier: 3, mpCost: 28, damage: 0, damageType: "buff", description: "Buff crescente: +10% dano cumulativo por rodada (máx 4 rodadas)."},
		{id: "bd_magnum_opus", name: "Magnum Opus", emoji: "🎻", branch: "inspiracao", tier: 4, mpCost: 55, damage: 0, damageType: "buff", description: "A peça final: +50% a todos os atributos por 5 rodadas."},
		// Branch: Encantamento
		{id: "bd_lullaby", name: "Canção de Ninar", emoji: "😴", branch: "encantamento", tier: 1, mpCost: 10, damage: 0, damageType: "utility", description: "Pode adormecer o inimigo por 2 rodadas (falha se sofrer dano)."},
		{id: "bd_charm", name: "Charme", emoji: "💕", branch: "encantamento", tier: 2, mpCost: 20, damage: 0, damageType: "utility", description: "50% chance de fascinar o inimigo: ele perde 1 ação."},
		{id: "bd_dissonance", name: "Dissonância", emoji: "🔔", branch: "encantamento", tier: 3, mpCost: 30, damage: 35, damageType: "magic", description: "Som dissonante causa dano psíquico e reduz INT do inimigo em 5."},
		{id: "bd_siren_song", name: "Canção da Sereia", emoji: "🧜", branch: "encantamento", tier: 4, mpCost: 50, damage: 0, damageType: "utility", description: "Encanta completamente: inimigo fica sob controle por 3 rodadas."},
		// Branch: Performance
		{id: "bd_taunt", name: "Provocar", emoji: "😜", branch: "performance", tier: 1, mpCost: 6, damage: 8, damageType: "physical", description: "Golpe leve + provoca: inimigo só ataca o Bardo por 1 rodada."},
		{id: "bd_acrobatics", name: "Acrobacia", emoji: "🤸", branch: "performance", tier: 2, mpCost: 12, damage: 0, damageType: "buff", description: "+40% esquiva por 2 rodadas."},
		{id: "bd_showstopper", name: "Parar o Espetáculo", emoji: "🎪", branch: "performance", tier: 3, mpCost: 25, damage: 45, damageType: "magic", description: "Golpe de magia de performance. Atordoa se d20 ≥ 14."},
		{id: "bd_grand_finale", name: "Grande Finale", emoji: "🎆", branch: "performance", tier: 4, mpCost: 60, damage: 100, damageType: "magic", description: "Explosão de energia artística. Afeta todos na área com dano e buff aliado."},
	}},

	// ── Caçador ────────────────────────────────────────────────────────────────
	{classID: "archer", skills: []skillDef{
		// Branch: Atirador
		{id: "ar_aimed_shot", name: "Disparo Mira", emoji: "🎯", branch: "atirador", tier: 1, mpCost: 8, damage: 26, damageType: "physical", description: "Mira cuidadosa: +20% dano, +10% acerto."},
		{id: "ar_multishot", name: "Chuva de Flechas", emoji: "🌧️", branch: "atirador", tier: 2, mpCost: 20, damage: 20, damageType: "physical", description: "3 flechas consecutivas. Cada uma rola d20 independente."},
		{id: "ar_sniper", name: "Sniper", emoji: "🔭", branch: "atirador", tier: 3, mpCost: 30, damage: 80, damageType: "physical", description: "Disparo de longo alcance: garante acerto crítico se d20 ≥ 16."},
		{id: "ar_rain_of_arrows", name: "Chuvarada", emoji: "⛈️", branch: "atirador", tier: 4, mpCost: 55, damage: 55, damageType: "physical", description: "5 flechas simultâneas. Cada flecha com dano completo + chance de crítico."},
		// Branch: Rastreador
		{id: "ar_track", name: "Rastrear", emoji: "🐾", branch: "rastreador", tier: 1, mpCost: 0, damage: 0, damageType: "passive", passive: true, description: "Passiva: +15% dano contra inimigos com HP abaixo de 50%."},
		{id: "ar_hunters_mark", name: "Marca do Caçador", emoji: "🎯", branch: "rastreador", tier: 2, mpCost: 14, damage: 0, damageType: "utility", description: "Marca o inimigo: +25% dano de todas as fontes contra ele por 4 turnos."},
		{id: "ar_beast_slayer", name: "Matador de Bestas", emoji: "🐺", branch: "rastreador", tier: 3, mpCost: 0, damage: 0, damageType: "passive", passive: true, description: "Passiva: +35% dano contra monstros de tipo 'besta'."},
		{id: "ar_kill_shot", name: "Tiro Final", emoji: "💀", branch: "rastreador", tier: 4, mpCost: 45, damage: 120, damageType: "physical", description: "Executa inimigo com HP < 25% com um tiro garantido."},
		// Branch: Armadilha
		{id: "ar_snare", name: "Armadilha", emoji: "🪤", branch: "armadilha", tier: 1, mpCost: 6, damage: 0, damageType: "utility", description: "Prende o inimigo: perde próxima ação."},
		{id: "ar_explosive_trap", name: "Armadilha Explosiva", emoji: "💥", branch: "armadilha", tier: 2, mpCost: 18, damage: 40, damageType: "physical", description: "Armadilha com dano ao ativar."},
		{id: "ar_poison_trap", name: "Armadilha de Veneno", emoji: "☠️", branch: "armadilha", tier: 3, mpCost: 22, damage: 15, damageType: "physical", description: "Armadilha que aplica veneno intenso por 5 rodadas.", poisonDmg: 15, poisonTurns: 5},
		{id: "ar_death_trap", name: "Armadilha da Morte", emoji: "⚰️", branch: "armadilha", tier: 4, mpCost: 40, damage: 100, damageType: "physical", description: "Armadilha colossal com dano massivo e atordoamento."},
	}},

	// ── Druida ─────────────────────────────────────────────────────────────────
	{classID: "druid", skills: []skillDef{
		// Branch: Natureza
		{id: "dr_entangle", name: "Enredar", emoji: "🌿", branch: "natureza", tier: 1, mpCost: 10, damage: 12, damageType: "magic", description: "Raízes imobilizam o inimigo por 2 rodadas."},
		{id: "dr_thorns", name: "Espinhos", emoji: "🌵", branch: "natureza", tier: 2, mpCost: 18, damage: 0, damageType: "buff", description: "Aura de espinhos: reflete 30% do dano recebido por 4 rodadas."},
		{id: "dr_call_lightning", name: "Invocar Relâmpago", emoji: "⚡", branch: "natureza", tier: 3, mpCost: 35, damage: 75, damageType: "magic", description: "Relâmpago natural. Dano dobrado em inimigos aquáticos."},
		{id: "dr_wrath_of_nature", name: "Fúria da Natureza", emoji: "🌋", branch: "natureza", tier: 4, mpCost: 65, damage: 130, damageType: "magic", description: "Terremoto e tempestade combinados. Dano massivo de natureza."},
		// Branch: Metamorfose
		{id: "dr_beast_form", name: "Forma Besta", emoji: "🐺", branch: "metamorfose", tier: 1, mpCost: 12, damage: 0, damageType: "buff", description: "Transforma-se: +20% ATK, +15% velocidade por 3 rodadas."},
		{id: "dr_bear_form", name: "Forma de Urso", emoji: "🐻", branch: "metamorfose", tier: 2, mpCost: 22, damage: 0, damageType: "buff", description: "Transforma-se em urso: +40% HP, +20% defesa por 4 rodadas."},
		{id: "dr_eagle_form", name: "Forma de Águia", emoji: "🦅", branch: "metamorfose", tier: 3, mpCost: 30, damage: 60, damageType: "physical", description: "Mergulho de águia: ataque veloz que ignora 25% da armadura."},
		{id: "dr_primal_beast", name: "Besta Primordial", emoji: "🦁", branch: "metamorfose", tier: 4, mpCost: 55, damage: 100, damageType: "physical", description: "Forma ancestral máxima: +50% todos os stats por 5 rodadas."},
		// Branch: Cura Natural
		{id: "dr_rejuvenate", name: "Rejuvenescer", emoji: "🌱", branch: "cura_natural", tier: 1, mpCost: 10, damage: 0, damageType: "buff", description: "Regenera 10 HP por rodada por 4 turnos."},
		{id: "dr_wild_growth", name: "Crescimento Selvagem", emoji: "🌳", branch: "cura_natural", tier: 2, mpCost: 20, damage: 0, damageType: "buff", description: "Restaura 60 HP e cura 1 efeito negativo."},
		{id: "dr_barkskin", name: "Casca de Árvore", emoji: "🪵", branch: "cura_natural", tier: 3, mpCost: 25, damage: 0, damageType: "buff", description: "+6 CA e resistência a dano físico por 3 rodadas."},
		{id: "dr_heart_of_forest", name: "Coração da Floresta", emoji: "💚", branch: "cura_natural", tier: 4, mpCost: 50, damage: 0, damageType: "buff", description: "Cura total de HP, remove todos os debuffs, gera escudo de 100 HP."},
	}},

	// ── Necromante ─────────────────────────────────────────────────────────────
	{classID: "necromancer", skills: []skillDef{
		// Branch: Morte
		{id: "nc_death_bolt", name: "Dardo da Morte", emoji: "💀", branch: "morte", tier: 1, mpCost: 12, damage: 28, damageType: "magic", description: "Projétil de energia necrótica."},
		{id: "nc_wither", name: "Murchar", emoji: "🥀", branch: "morte", tier: 2, mpCost: 22, damage: 35, damageType: "magic", description: "Drena vitalidade: causa dano e cura 50% do dano como HP."},
		{id: "nc_death_coil", name: "Espiral da Morte", emoji: "🌀", branch: "morte", tier: 3, mpCost: 35, damage: 65, damageType: "magic", description: "Espiral necrótica que ignora resistência mágica."},
		{id: "nc_death_wave", name: "Onda da Morte", emoji: "🌊", branch: "morte", tier: 4, mpCost: 70, damage: 140, damageType: "magic", description: "Onda de energia mortal. Inimigos com HP < 30% morrem instantaneamente."},
		// Branch: Dreno
		{id: "nc_life_drain", name: "Dreno de Vida", emoji: "🩸", branch: "dreno", tier: 1, mpCost: 10, damage: 18, damageType: "magic", description: "Suga vida: causa 18 dano e cura 18 HP."},
		{id: "nc_soul_rip", name: "Rasgar Alma", emoji: "👻", branch: "dreno", tier: 2, mpCost: 20, damage: 30, damageType: "magic", description: "Arranca fragmento de alma. Causa dano e drena 20 MP do inimigo."},
		{id: "nc_vampiric_aura", name: "Aura Vampírica", emoji: "🦇", branch: "dreno", tier: 3, mpCost: 0, damage: 0, damageType: "passive", passive: true, description: "Passiva: cura 10% do dano causado por habilidades mágicas."},
		{id: "nc_soul_feast", name: "Banquete de Almas", emoji: "💜", branch: "dreno", tier: 4, mpCost: 55, damage: 110, damageType: "magic", description: "Consome a alma do inimigo. Cura 100% do dano causado."},
		// Branch: Sombra
		{id: "nc_decay", name: "Decomposição", emoji: "☠️", branch: "sombra", tier: 1, mpCost: 8, damage: 8, damageType: "magic", description: "Veneno necrótico: 10 dano/rodada por 4 turnos.", poisonDmg: 10, poisonTurns: 4},
		{id: "nc_bone_shield", name: "Escudo de Ossos", emoji: "🦴", branch: "sombra", tier: 2, mpCost: 18, damage: 0, damageType: "buff", description: "Escudo de fragmentos ósseos que absorve 45 de dano."},
		{id: "nc_shadow_bolt", name: "Raio das Sombras", emoji: "🌑", branch: "sombra", tier: 3, mpCost: 30, damage: 60, damageType: "magic", description: "Raio das trevas: -3 CA do alvo por 3 rodadas."},
		{id: "nc_lich_form", name: "Forma do Lich", emoji: "💀", branch: "sombra", tier: 4, mpCost: 60, damage: 0, damageType: "buff", description: "Transforma-se em lich por 5 rodadas: +50% dano mágico, imune a veneno, regenera 15 MP/rodada."},
	}},
}

// AllSkills is the complete generated skill map keyed by skill ID.
// Populated during init() and merged into game.Skills by the loader.
var AllSkills map[string]models.Skill

func init() {
	AllSkills = make(map[string]models.Skill, 130)
	for _, cd := range skillClassDefs {
		// Group by branch to build requires chains
		byBranch := map[string][]skillDef{}
		for _, s := range cd.skills {
			byBranch[s.branch] = append(byBranch[s.branch], s)
		}
		for _, branchSkills := range byBranch {
			for i, s := range branchSkills {
				ms := buildModelSkill(s, cd.classID)
				// Auto-wire requires: each skill requires previous in same branch
				if i > 0 {
					ms.Requires = branchSkills[i-1].id
				}
				AllSkills[ms.ID] = ms
			}
		}
	}
}

func buildModelSkill(s skillDef, classID string) models.Skill {
	reqLevel := firstRequiredLevel[s.tier]
	// Add an offset within the tier so T1 unlocks at 1, 6, 11, 16
	// (up to 4 tiers per branch in the 1-20 band)
	return models.Skill{
		ID:               s.id,
		Name:             s.name,
		Emoji:            s.emoji,
		Class:            classID,
		Branch:           s.branch,
		Tier:             s.tier,
		PointCost:        tierPointCost[s.tier],
		MPCost:           s.mpCost,
		Damage:           s.damage,
		DamageType:       s.damageType,
		RequiredLevel:    reqLevel,
		Passive:          s.passive,
		Description:      s.description,
		PoisonDmgPerTurn: s.poisonDmg,
		PoisonTurnsCount: s.poisonTurns,
	}
}

// SkillsForClass returns all skills for a given class ID, sorted by tier.
func SkillsForClass(classID string) []models.Skill {
	var out []models.Skill
	for _, s := range AllSkills {
		if s.Class == classID {
			out = append(out, s)
		}
	}
	return out
}

// SkillsByTier returns all skills at a given tier across all classes.
func SkillsByTier(tier int) []models.Skill {
	var out []models.Skill
	for _, s := range AllSkills {
		if s.Tier == tier {
			out = append(out, s)
		}
	}
	return out
}

// SkillID formats a canonical prefix so users of this package can build skill
// IDs programmatically: fmt.Sprintf(SkillIDFmt, classPrefix, branch, tier).
var SkillIDFmt = "%s_%s_t%d"

// SkillCount returns the total number of generated skills.
func SkillCount() int { return len(AllSkills) }

// ClassesWithSkills returns a de-duplicated list of class IDs present in AllSkills.
func ClassesWithSkills() []string {
	seen := map[string]bool{}
	for _, s := range AllSkills {
		seen[s.Class] = true
	}
	out := make([]string, 0, len(seen))
	for k := range seen {
		out = append(out, k)
	}
	return out
}

// FormatSkillID builds a normalised skill ID from components.
func FormatSkillID(classID, branch string, tier int) string {
	return fmt.Sprintf("%s_%s_t%d", classID, strings.ToLower(branch), tier)
}
