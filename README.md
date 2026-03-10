# ⚔️ Tormenta RPG Bot

Bot MMORPG multiplayer para Telegram, inspirado no sistema Tormenta 20. Combate por turnos com sistema d20, masmorras procedurais, PvP ranqueado, guildas, economia controlada, bosses mundiais, raids, temporadas e pagamentos via Pix (AbacatePay).

Arquitetura modular escrita em Go, projetada para escalar de centenas a **1 milhão de jogadores**.

> **Conteúdo gerado:** 210 habilidades · 750+ itens · 59+ monstros · 8 classes · 7 raças — do nível 1 ao 100 · 20 combos · maestria em 6 níveis.

---

## Índice

1. [Stack e dependências](#stack-e-dependências)
2. [Configuração rápida](#configuração-rápida)
3. [Variáveis de ambiente](#variáveis-de-ambiente)
4. [Banco de dados](#banco-de-dados)
5. [Estrutura do projeto](#estrutura-do-projeto)
6. [Gameplay](#gameplay)
7. [Sistemas avançados](#sistemas-avançados)
8. [Dados RPG gerados (rpgdata)](#dados-rpg-gerados-rpgdata)
9. [Comandos de GM](#comandos-de-gm)
10. [Pagamentos Pix](#pagamentos-pix-abacatepay)
11. [Arquitetura técnica](#arquitetura-técnica)
12. [Deploy com Docker](#deploy-com-docker)
13. [Deploy manual](#deploy-manual)

---

## Stack e dependências

| Componente | Tecnologia |
|---|---|
| Linguagem | Go 1.21 |
| Telegram API | `go-telegram-bot-api/v5` |
| Banco de dados | PostgreSQL 16 |
| Cache | In-memory (MemCache) com interface para Redis |
| Pagamentos | AbacatePay (Pix) |
| Geração de imagens | Go nativo (`image/draw`, `image/png`) |
| Containerização | Docker + Docker Compose |

---

## Configuração rápida

```bash
git clone https://github.com/seu-usuario/tormenta-bot.git
cd tormenta-bot
cp .env.example .env
# edite o .env com suas credenciais
docker compose up -d
docker compose logs -f bot
```

O banco é inicializado automaticamente pelo Docker com as migrações em `./migrations/`.

---

## Variáveis de ambiente

```env
# OBRIGATÓRIO
TELEGRAM_TOKEN=123456789:ABCdef...
DATABASE_URL=postgres://tormenta:troque_esta_senha@localhost:5432/tormenta_rpg?sslmode=disable
GM_IDS=123456789,987654321           # IDs Telegram dos GMs, separados por vírgula

# OPCIONAL
ABACATEPAY_TOKEN=acp_...             # Habilita pagamentos Pix
MP_WEBHOOK_PORT=8080                 # Porta HTTP para webhooks (padrão: 8080)
GM_LOG_CHAT=-100123456789            # Grupo para logs de ações GM
ASSETS_DIR=./assets/images           # Diretório de imagens geradas
DUNGEON_PROCEDURAL_ENABLED=false     # true = dungeons procedurais (5-10 salas)
ABACATEPAY_WEBHOOK_SECRET=change-me  # valida header X-AbacatePay-Secret
ECONOMY_DYNAMIC_PRICING=false        # preço dinâmico progressivo na loja NPC
TELEMETRY_ENABLED=false              # analytics opcionais
ENERGY_FIXED_CAP=true                # cap fixo 100/200 no worker central de energia
FORGE_PROFILE=legacy10               # legacy10 | classic5
UPDATE_WORKERS=1                     # workers de processamento de updates do Telegram
REDIS_ADDR=localhost:6379            # opcional: ativa Redis para cache distribuído
```

> **Importante:** nunca faça commit do `.env`. Se tokens reais forem expostos, faça **rotação** imediata nos painéis correspondentes.

---

## Banco de dados

O projeto usa **migrações incrementais** em `migrations/001_init.sql` até `migrations/021_energy_tick_index.sql`.

```bash
# Subida manual (sem Docker)
psql -U postgres -c "CREATE DATABASE tormenta_rpg;"
for f in migrations/*.sql; do psql -U tormenta -d tormenta_rpg -f "$f"; done
```

Com Docker, o PostgreSQL executa automaticamente todos os `.sql` de `./migrations/` em ordem alfabética.

### Tabelas principais

| Tabela | Descrição |
|---|---|
| `players` | Conta Telegram (id, ban, VIP) |
| `characters` | Personagem (stats, localização, estado) |
| `inventory` | Itens por personagem |
| `character_skills` | Habilidades aprendidas |
| `dungeon_runs` / `dungeon_best` | Histórico e recordes de masmorras |
| `pvp_challenges` / `pvp_stats` | Duelos e ranking Elo |
| `pix_payments` | Pagamentos Pix |
| `auto_hunt_sessions` | Sessões de caça automática VIP |
| `player_timers` | Timers/cooldowns genéricos |
| `item_usage_stats` | Estatísticas de uso para precificação dinâmica |
| `analytics_events` | Eventos opcionais de telemetria |
| `gm_action_logs` | Auditoria de ações administrativas |
| `player_items` | Instâncias de itens equipáveis (forja/progressão) |
| `guilds` | Guildas (nome, tag, líder, nível, XP, banco, território) |
| `guild_members` | Membros por guilda (rank, contribuição, última vez online) |
| `market_listings` | Listagens ativas do mercado entre jogadores |
| `market_auctions` | Leilões em andamento e histórico |
| `pvp_ratings` | Rating Elo por jogador e temporada |
| `world_boss_log` | Participação e dano por jogador em cada boss mundial |
| `seasons` | Histórico de temporadas (data início/fim, tema) |
| `season_rewards` | Recompensas concedidas ao encerrar temporada |

---

## Estrutura do projeto

```
tormenta-bot/
├── cmd/bot/
│   └── main.go                    # Entrypoint: bot, webhook HTTP, workers
├── internal/
│   ├── engine/                    # Motor de combate modular (sem efeitos colaterais)
│   │   ├── state_guard.go         # Validação de transição de estado
│   │   ├── combat_engine.go       # CombatEngine: orquestra um turno completo
│   │   ├── damage_calculator.go   # Fórmula de dano, elementos, críticos, D20
│   │   ├── status_engine.go       # StatusSet, 12 efeitos (veneno, burn, stun…)
│   │   ├── effect_processor.go    # Combatant interface, ProcessEffect/Effects
│   │   ├── skill_engine.go        # Resolução de habilidades, PassiveRegistry
│   │   ├── ai_engine.go           # AITier, MonsterAdaptation, SelectAction
│   │   ├── combo_engine.go        # GlobalCombos: 20 combos de 3 passos com janela temporal
│   │   └── mastery_engine.go      # GlobalMasteryStore: maestria em 6 níveis (0–300 usos)
│   ├── rpg/                       # Sistema RPG completo (extensão sem quebrar game/)
│   │   ├── race.go                # 7 raças: Humano, Elfo, Anão, Goblin, Qareen, Minotauro, Meio-Orc
│   │   ├── class.go               # 8 classes: Guerreiro, Arcanista, Ladino, Caçador,
│   │   │                          #            Paladino, Clérigo, Bárbaro, Bardo
│   │   ├── attributes.go          # Atributos, modificadores D&D, stats derivados
│   │   ├── xp.go                  # Curva XP nível 1–100, milestones, recompensas
│   │   ├── skill_tree.go          # SkillNode, Branch, SkillTree, Trees map
│   │   ├── skill_trees_warrior.go # 27 skills: Espada & Escudo · Maestria · Determinação
│   │   ├── skill_trees_mage.go    # 27 skills: Arcanismo · Elementalismo · Transmutação
│   │   ├── skill_trees_rogue.go   # 25 skills: Sombras · Venenos · Esperteza
│   │   ├── skill_trees_archer.go  # 27 skills: Precisão · Natureza · Sobrevivência
│   │   ├── skill_trees_extended.go# Extensão barb/pal/cler/bard + 30+ passivas novas
│   │   ├── skill_trees_extra.go   # Skills finais de completamento
│   │   ├── talents.go             # Talentos (desbloqueados nos níveis 10/25/50/75/100)
│   │   └── passives.go            # Registro de passivas de raça/classe no engine
│   ├── economy/                   # Economia controlada anti-inflação
│   │   ├── economy_manager.go     # EconomyManager global, gold in circulation
│   │   ├── inflation_controller.go# Multiplicadores dinâmicos de drop/custo
│   │   └── tax_system.go          # Taxas de mercado/leilão/comércio direto
│   ├── config/                    # Constantes de jogo centralizadas
│   │   └── game_config.go         # EnergyMaxFree/VIP, taxas, intervalos, Elo K, etc.
│   ├── state/                     # Gerenciador de sessão por usuário
│   │   └── user_state.go          # Manager mutex-safe para estado efêmero de handlers
│   ├── market/                    # Mercado entre jogadores
│   │   ├── listing.go             # Listagens de itens, ListingStore
│   │   ├── auction.go             # Leilão, lances, buy-it-now, liquidação
│   │   ├── market_service.go      # MarketService: post/buy/cancel/bid/settle
│   │   ├── db_store.go            # DBListingStore + DBAuctionStore (PostgreSQL)
│   │   └── globals.go             # InitDB() substitui MemStore por DBStore na startup
│   ├── guild/                     # Sistema completo de guildas
│   │   ├── guild.go               # Guild model, Store interface, MemStore
│   │   ├── guild_members.go       # GuildService: criar/convidar/expulsar/promover
│   │   ├── guild_bank.go          # Banco da guilda, depósito/saque, XP de guilda
│   │   ├── guild_perks.go         # Perks por nível, buffs temporários de guilda
│   │   ├── guild_war.go           # Territórios, guerras agendadas, pontuação
│   │   ├── db_store.go            # DBStore — implementação PostgreSQL do guild.Store
│   │   └── globals.go             # InitDB() substitui MemStore por DBStore na startup
│   ├── world/                     # Bosses mundiais e raids
│   │   ├── world_boss.go          # BossManager global, spawn a cada 12h, ranking de dano
│   │   └── raid_system.go         # RaidManager, sessões multi-fase, 5–20 jogadores
│   ├── ai/                        # IA adaptativa de monstros
│   │   ├── player_behavior_tracker.go # BehaviourProfile por jogador
│   │   ├── monster_ai.go          # AdaptationStore, Advise, RecordPlayerAction
│   │   └── behavior_tree.go       # Árvore de comportamento (Sequence/Selector)
│   ├── pvp/                       # Arena PvP
│   │   ├── ranking.go             # Sistema Elo, divisões, leaderboard, reset de temporada
│   │   ├── matchmaking.go         # Fila com janela de rating crescente
│   │   └── arena.go               # ArenaManager, partidas 1v1, forfeit, atualização de ranking
│   ├── season/                    # Sistema de temporadas (3 meses)
│   │   ├── season_manager.go      # Manager, StartSeason, CheckAndRollOver
│   │   └── season_rewards.go      # 7 tiers de recompensa, SeasonEndSummary
│   ├── eventbus/                  # Barramento de eventos global
│   │   ├── events.go              # Event struct, 30+ Kind constants, construtores fluentes
│   │   └── eventbus.go            # Bus com pool de 8 goroutines, 10k slots de fila
│   ├── cache/                     # Camada de cache unificada
│   │   ├── ttl_cache.go           # Cache TTL genérico (existente)
│   │   ├── game_cache.go          # Cache de ranking/loja/dungeon (existente)
│   │   └── redis_cache.go         # Interface Cache + MemCache fallback + helpers de chave
│   ├── workers/                   # Workers de background escaláveis
│   │   ├── combat_worker.go       # CombatPool: 16 goroutines, 1000 jobs assíncronos
│   │   ├── economy_worker.go      # Snapshot + eventos de inflação a cada 10 min
│   │   ├── event_worker.go        # Reage ao eventbus, broadcast Telegram
│   │   └── raid_worker.go         # Spawn/expiração de boss mundial
│   ├── rpgdata/                   # Biblioteca de dados RPG nível 1–100 (pura, sem I/O)
│   │   ├── tiers.go               # 5 tiers de item, 4 de habilidade, funções de escala
│   │   ├── xp_curve.go            # Tabela XP lv1-100, milestones, fórmula suave
│   │   ├── races.go               # 7 raças com stats e traços
│   │   ├── classes.go             # 10 classes com HP/MP/nível e traços
│   │   ├── skills.go              # 120 habilidades geradas (10 classes × 3 ramos × 4 tiers)
│   │   ├── skill_tree.go          # SkillTree/Branch/SkillNode, CanUnlock()
│   │   ├── items.go               # 750+ itens gerados (25 templates × 5 tiers × 6 raridades)
│   │   └── monsters.go            # 59+ monstros gerados (13 arquétipos × bandas de nível)
│   ├── game/                      # Lógica de jogo principal (existente, preservado)
│   │   ├── combat.go              # Engine de combate por turnos (d20)
│   │   ├── data.go                # Raças, classes, monstros, mapas, itens (legado)
│   │   ├── rpgdata_loader.go      # init(): mescla rpgdata nos mapas game.Items/Skills/Monsters
│   │   ├── dungeon_logic.go       # Lógica de masmorras
│   │   ├── energy.go              # Sistema de energia e regeneração
│   │   ├── pvp_game.go            # Lógica de PvP
│   │   ├── extended_items.go      # Itens estendidos (materiais/crafting)
│   │   └── safe_lookup.go         # GetMonster/Race/Class/Skill/Item — nil-safe com erro
│   │
│   ├── game/skills/               # Auditoria e validação do sistema de habilidades
│   │   └── validator.go           # ValidateSkillTrees(): IDs duplicados, mecânicas iguais,
│   │                              #   requisitos órfãos, anomalias de balanceamento (mean+2σ)
│   │
│   ├── ui/                        # Motor de menus Telegram profissional
│   │   └── menu_engine.go         # MenuState, Engine global, debounce 500ms, cache TTL 5s,
│   │                              #   NavStack, PaginateItems[T], CompressCallback ≤64 bytes
│   │
│   ├── handlers/                  # Handlers Telegram
│   │   ├── safe_state.go          # RWMutex + acessores para 6 mapas de sessão (thread-safe)
│   │   └── pvp_handler.go         # pvpTargetGet/Set/Clear — race condition corrigida
│   ├── menu/                      # Menus de teclado inline
│   ├── router/                    # Roteador de mensagens/callbacks
│   ├── services/                  # Serviços de negócio
│   ├── repository/                # Contratos e adapters de persistência
│   ├── database/                  # Queries SQL e migrações
│   ├── models/                    # Structs: Character, Player, Item, Monster…
│   ├── items/ forge/ drops/ crafting/ dungeon/ energy/ explore/
│   ├── systems/ timers/ bot/ assets/ gmtools/ telemetry/ utils/
│   └── service/ worker/ cache/    # Camadas de suporte (existentes)
├── migrations/                    # 021 migrações SQL incrementais
├── scripts/
│   ├── update.ps1                 # Atualização com backup (Windows)
│   └── update.sh                  # Atualização com backup (Linux/macOS)
├── assets/images/                 # Imagens geradas (não versionado)
├── Dockerfile
├── docker-compose.yml
└── .env.example
```

---

## Gameplay

### Raças jogáveis

| Raça | Traço passivo | Destaques |
|---|---|---|
| 👤 Humano | +10% XP ganho | Atributos equilibrados, versátil |
| 🧝 Elfo | +20% dano mágico | DEX +3, INT +3, MP +20 |
| ⛏️ Anão | -15% dano recebido | CON +4, HP +25, imune a Freeze |
| 👹 Meio-Orc | +25% dano físico | STR +4, HP +20, Berserk sob 20% HP |
| 👺 Goblin | +15% chance crítico | DEX +4, venenos +1 turno, HP -10 |
| 🧞 Qareen | +30% dano de Fogo | INT +4, CHA +4, imune a Burn, MP +30 |
| 🐂 Minotauro | Ignora 10% de armadura | STR +5, CON +4, HP +40, resistente a físico |

### Classes jogáveis

| Classe | Função | HP base | MP base | Traço |
|---|---|---|---|---|
| ⚔️ Guerreiro | Tanque | 80 | 20 | Maestria em Armas |
| 🧙 Arcanista | Conjurador | 45 | 80 | Surto Arcano (a cada 4 feitiços, 1 grátis) |
| 🗡️ Ladino | DPS | 55 | 40 | Facada pelas Costas (+40% dano em flanqueamento) |
| 🏹 Caçador | Distância | 60 | 35 | Olho de Águia (+10% crit à distância) |
| ⚜️ Paladino | Tanque/Suporte | 75 | 50 | Aura Sagrada (cura +10% para aliados) |
| ✝️ Clérigo | Curandeiro | 55 | 75 | Graça Divina (10% curas duplas, imune a Curse) |
| 🪓 Bárbaro | DPS brutal | 90 | 10 | Fúria Bárbara (+60% ATK em Berserk) |
| 🎵 Bardo | Suporte | 58 | 60 | Inspiração (+15% XP da party pós-batalha) |
| 🌿 Druida | Suporte/DPS | 60 | 70 | Forma Selvagem (metamorfose em combate) |
| 💀 Necromante | Conjurador | 48 | 95 | Senhor dos Mortos (invoca mortos-vivos por turno) |

### Progressão (nível 1–100)

| Milestone | Nível | Recompensa extra |
|---|---|---|
| Aventureiro | 10 | +1 ponto de habilidade bônus |
| Veterano | 25 | +2 pontos de habilidade |
| Campeão | 50 | +3 pts, +50 HP, +25 MP |
| Lendário | 75 | +4 pontos de habilidade |
| Imortal | 100 | +5 pts, +100 HP, +50 MP |

A curva XP é suave: \~150k XP total no nível 20 (antigo cap), \~1.2M no nível 50, \~8M no nível 100.

A cada **4 níveis** os dois atributos primários da classe sobem +1. A cada **10 níveis** todos os atributos sobem +1.

### Árvore de habilidades

Cada classe possui **3 ramos** com **5 tiers** de habilidades (T1 Lv1 → T5 Ultimate Lv75-80). Total: **210 habilidades** balanceadas do nível 1 ao 100, com pré-requisitos, custo em pontos de habilidade e efeitos de engine.

| Classe | Ramos | Skills | Ultimas |
|---|---|---|---|
| ⚔️ Guerreiro | Espada & Escudo · Maestria de Armas · Determinação | 27 | Avatar da Guerra |
| 🧙 Arcanista | Arcanismo · Elementalismo · Transmutação | 27 | Aniquilação Arcana · Raio Prismático · Ruptura da Realidade |
| 🗡️ Ladino | Sombras · Venenos · Esperteza | 25 | Assassinar · Toxina Viral · Golpe de Misericórdia |
| 🏹 Caçador | Precisão · Natureza · Sobrevivência | 27 | Flecha Divina · Avatar da Natureza · Tiro Fatal |
| ⚜️ Paladino | Sagrado · Proteção · Julgamento | 27 | Ressurreição Divina · Escudo Divino · Armagedom |
| ✝️ Clérigo | Cura · Luz Divina · Proteção Divina | 26 | Milagre · Avatar da Luz · Escudo dos Deuses |
| 🪓 Bárbaro | Fúria · Resistência · Guerreiro Tribal | 25 | Senhor da Guerra · Postura da Montanha · Impacto do Titã |
| 🎵 Bardo | Música · Conhecimento Bárdico · Ilusionismo | 26 | Canção das Lendas · Final da Ópera · Grande Ilusão |

**Tiers de progressão:**

| Tier | Nível | Custo de pontos | Custo MP |
|---|---|---|---|
| T1 | 1–8 | 1 pt | 5–15 MP |
| T2 | 10–22 | 1–2 pts | 15–35 MP |
| T3 | 25–40 | 2 pts | 30–60 MP |
| T4 | 40–60 | 3 pts | 50–90 MP |
| T5 Ultimate | 75–80 | 5 pts | 90–130 MP |

### Sistema de Combos

Sequências específicas de habilidades na ordem correta ativam combos que causam bônus de dano e efeitos extras.

| Combo | Classe | Sequência | Bônus |
|---|---|---|---|
| Trindade de Ferro | Guerreiro | Golpe Firme → Corte Poderoso → Executar | +50% dano + 80 físico |
| Tri-Elemental | Arcanista | Bola de Fogo → Raio → Lança de Gelo | +60% dano + AoE mágica |
| Arte do Assassino | Ladino | Ataque Pelas Costas → Marca da Morte → Assassinar | ×2 dano + 100 físico |
| Abate Perfeito | Caçador | Marca do Caçador → Tiro na Cabeça → Tiro Fatal | ×2 dano + 95 físico |
| Retribuição Divina | Paladino | Marca da Justiça → Ira Sagrada → Julgamento Divino | +80% dano + 100 sagrado |
| Frenesi de Sangue | Bárbaro | Fúria → Sede de Sangue → Devastação | Berserk 5 turnos |
| Sinfonia da Destruição | Bardo | Hino de Batalha → Discordância → Final da Ópera | AoE sombra + 80 |

20 combos no total — 3 por classe principal, 2 por suporte.

### Sistema de Maestria

Usar uma habilidade repetidamente desenvolve maestria, desbloqueando bônus progressivos:

| Nível | Usos | Bônus de dano | Redução de MP | Crítico extra |
|---|---|---|---|---|
| ⚪ Novato | 0 | — | — | — |
| 🟢 Aprendiz | 10 | +8% | -3% | +1% |
| 🔵 Adepto | 30 | +18% | -8% | +3% |
| 🟣 Especialista | 75 | +32% | -15% | +6% |
| 🟡 Mestre | 150 | +50% | -25% | +10% |
| 🔴 Grão-Mestre | 300 | +75% | -40% | +15% |

### Talentos

Aos níveis 10, 25, 50, 75 e 100 o jogador escolhe um talento. Exemplos:

| Talento | Nível | Efeito |
|---|---|---|
| 💪 Durão | 10 | +50 HP máximo |
| 📚 Aprendizado Rápido | 10 | +15% XP |
| 💎 Caçador de Tesouros | 25 | +10% drop rate |
| 🌟 Proeza Lendária | 75 | +5% crit, +30 ATK, +20 DEF |
| 👑 Paragão | 100 | +200 HP, +100 MP, +50 ATK, +10% XP, +20% ouro |

### Zonas e progressão de mapa

| Zona | Nível | Monstros |
|---|---|---|
| 🏘️ Vila de Trifort | Qualquer | Hub: loja e estalagem |
| 🌾 Arredores da Vila | 1–5 | Rato, Goblin, Slime |
| 🌲 Floresta Sombria | 4–9 | Lobo, Orc, Troll, Harpia |
| 💎 Caverna de Cristal | 8–13 | Morcego, Aranha, Golem, Morto-Vivo |
| 🏚️ Masmorra Antiga | 13–18 | Demônio, Necromante, Vampiro, Lich |
| 🏔️ Pico dos Dragões | 17–100 | Dragão Jovem, Ancião, Wyvern, Fênix |

### Sistema de energia ⚡

| Ação | Custo |
|---|---|
| Viajar entre zonas | 1⚡ |
| Iniciar combate | 1⚡ |
| Andar em masmorra | 1⚡ |
| Tick de caça automática | 1⚡ |

| Status | Máximo | Regeneração |
|---|---|---|
| Normal | 100 + (nível−1)×2 | 1⚡ / 10 min |
| 👑 VIP | 200 + (nível−1)×4 | 1⚡ / 5 min |

### Combate (sistema d20 modular)

1. Explorar → 1 monstro sorteado na zona
2. Por turno: **Atacar** / **Habilidade** / **Item** / **Fugir**
3. Vitória: XP + ouro + chance de drop + efeito de talento/raça
4. Derrota: perde XP e ouro proporcionais, retorna à Vila

**Efeitos de status (engine/status_engine.go):**
Veneno, Queimadura, Congelamento, Atordoamento, Cegueira, Berserk, Escudo, Pressa, Regeneração, Maldição, Silêncio, Proteção — com DoT/HoT, duração em turnos e modificadores de dano/precisão.

**IA adaptativa dos monstros:**
Monstros aprendem com o comportamento do jogador. Se a maioria das ações for mágica, o monstro ganha resistência mágica. Se for física, ganha counterataque. Monstros veteranos usam árvore de comportamento (Behavior Tree) com enrage, cura e habilidades especiais.

---

## Sistemas avançados

### Economia controlada

O `EconomyManager` rastreia todo o ouro em circulação. Quando detecta inflação:

| Nível | Gatilho (ouro/jogador) | Efeito |
|---|---|---|
| Normal | < 50.000 | Drop e preços padrão |
| ⚠️ Aviso | ≥ 50.000 | Drop −30%, preços +30% |
| 🚨 Crítico | ≥ 200.000 | Drop −60%, preços +80%, taxas dobradas |

Todo ouro destruído (reparos, taxas, leilões) é registrado como **gold sink** e reduz a circulação automaticamente.

### Mercado entre jogadores

- **Listagem:** postar item à venda (máx. 10 listagens ativas, taxa de 1% na postagem)
- **Compra direta:** o comprador paga o preço anunciado menos 5% de imposto
- **Leilão:** lances progressivos com buy-it-now opcional; 8% de imposto sobre o lance vencedor
- **Taxa durante inflação:** duplica automaticamente para destruir ouro mais rapidamente

### Guildas

| Recurso | Descrição |
|---|---|
| Criação | Qualquer jogador pode fundar uma guilda |
| Hierarquia | Líder → Oficial → Membro → Recruta |
| Banco | Depósito/saque com rastreamento de contribuição por membro |
| Nível (1–10) | Sobe com XP de guilda; desbloqueia +membros, perks e slots de território |
| Perks passivos | +3% XP/nível, +2% ouro/nível, +1% drop rate/nível |
| Buffs temporários | Bênção do Sábio (+30% XP 2h), Febre do Ouro (+40% gold 2h), Sinfonia da Vitória (+30% ATK 5t) |

### Guerras de guilda e territórios

4 territórios capturáveis com bônus de renda passiva:

| Território | Renda/hora | Bônus XP | Multiplicador de recurso |
|---|---|---|---|
| 🌲 Floresta Sombria | 50🪙 | +10% | ×1.2 |
| 🏰 Fortaleza de Ferro | 100🪙 | +5% | ×1.5 |
| 🏛️ Templo Submerso | 80🪙 | +15% | ×1.1 |
| 🐉 Pico do Dragão | 200🪙 | +20% | ×2.0 |

Guerra tem aviso de 30 minutos, duração de 1 hora e liquidação automática.

### Bosses Mundiais

Spawn automático a cada **12 horas** (janela de 30 minutos para derrota):

| Boss | Nível | Elemento | Fraqueza | Drop garantido |
|---|---|---|---|---|
| 🌀 Tormenta | 50 | Trevas | Sagrado | tormenta_shard |
| 🦂 Rei Escorpião | 30 | Veneno | Fogo | poison_gland |
| 🐲 Dragão de Gelo | 45 | Gelo | Fogo | dragon_scale |

O HP do boss **escala** com o número de participantes (+60% por jogador adicional, máx. ×10). As recompensas são ranqueadas pelo dano causado — top 3 têm chance de item lendário.

### Raids

Raid de múltiplos estágios para grupos de 5–20 jogadores:

- **Raid: Núcleo da Tormenta** — 3 fases (Rei Escorpião → Dragão de Gelo → Tormenta)
- Cada fase tem mecânica especial (veneno em área, onda de gelo, dano dobrado)
- Recompensas ranqueadas por dano + token de raid garantido

### PvP Arena

- **Matchmaking:** fila com janela de rating ±100 (expande +25 a cada 30s de espera)
- **Sistema Elo:** K=32, rating inicial 1000
- **Divisões:** Bronze → Prata → Ouro → Platina → Mestre → Lendário → 🏆 Imortal
- **Forfeit:** jogador pode se render a qualquer momento

### Temporadas

Cada temporada dura **3 meses** com tema diferente e boss em destaque. Ao encerrar:
- Rankings e ratings PvP resetam (soft-reset: metade da distância ao padrão)
- Jogadores recebem recompensas baseadas no tier final

| Tier | Rating mín. | 💎 Diamantes | Título |
|---|---|---|---|
| 🏆 Imortal | 2400 | 500 | Imortal + coroa exclusiva |
| 💎 Lendário | 2000 | 300 | Lendário + capa exclusiva |
| 🥇 Mestre | 1600 | 200 | Mestre + distintivo |
| 🥈 Platina | 1300 | 100 | Platina + anel |
| 🥉 Ouro | 1100 | 50 | Ouro + amuleto |
| ⚔️ Prata | 900 | 25 | Prata + skin de espada |
| 🗡️ Bronze | 0 | 10 | Bronze + skin de escudo |

### Forja e Crafting

Equipamentos podem ser aprimorados de `+1` até `+10`.

| Nível alvo | Chance | Risco de quebra |
|---|---|---|
| +1 a +4 | 70–100% | Nenhum |
| +5 a +7 | 40–60% | Sim |
| +8 a +10 | 10–30% | Alto |

Crafting usa receitas com consumo de materiais (Pedra de Forja, Metal Negro, Essência Arcana, etc.).

### Sistema VIP

| Plano | Custo | Duração |
|---|---|---|
| 30 dias | 500 💎 | 30 dias |
| 90 dias | 1.200 💎 | 90 dias |
| Permanente | 3.000 💎 | Vitalício |

**Benefícios:** energia máxima dobrada, regeneração 2× mais rápida, 🤖 caça automática offline.

---

## Dados RPG gerados (rpgdata)

O pacote `internal/rpgdata` é uma biblioteca de dados pura (sem I/O, sem BD) que gera programaticamente todo o conteúdo escalável do jogo. Os dados são mesclados nos mapas do `game/` via `init()` em `rpgdata_loader.go`, sem modificar nenhum arquivo existente.

### Escala de conteúdo gerado

| Categoria | Quantidade | Método |
|---|---|---|
| Habilidades | **210** | 8 classes × 3 ramos × T1–T5 (engine/rpg) |
| Combos | **20** | 3 por classe principal, 2 por suporte |
| Níveis de Maestria | **6** | Novato → Grão-Mestre (engine/mastery) |
| Itens | **750+** | 25 templates × 5 tiers × 6 raridades (rpgdata) |
| Monstros | **59+** | 13 arquétipos × bandas de nível (rpgdata) |
| Raças | 7 | Definidas manualmente |
| Classes | 8 | Definidas manualmente |

### Raridades de itens

| Raridade | Multiplicador de stat | Peso de drop |
|---|---|---|
| ⚪ Comum | ×1,00 | 60 |
| 🟢 Incomum | ×1,30 | 25 |
| 🔵 Raro | ×1,70 | 10 |
| 🟣 Épico | ×2,30 | 4 |
| 🟡 Lendário | ×3,20 | 1 |
| 🔴 Mítico | ×4,50 | 0 (drop manual) |

### Tiers de item vs. nível

| Tier | Nome | Faixa | Multiplicador |
|---|---|---|---|
| 1 | Aprendiz 🪨 | lv 1–20 | ×1,0 |
| 2 | Veterano 🔩 | lv 21–40 | ×2,5 |
| 3 | Mestre ⚙️ | lv 41–60 | ×5,0 |
| 4 | Élite 💎 | lv 61–80 | ×9,0 |
| 5 | Lendário 🌌 | lv 81–100 | ×14,0 |

---

## Comandos de GM

Apenas IDs listados em `GM_IDS` têm acesso. O comando `/gm` abre painel interativo.

| Comando | Descrição |
|---|---|
| `/gm buscar <nome>` | Busca personagem por nome |
| `/gm info <nome>` | Ficha completa do personagem |
| `/gm id <telegramID>` | Lookup por ID Telegram |
| `/gm ban <nome> [razão]` | Bane jogador |
| `/gm unban <nome>` | Desbane jogador |
| `/gm diamond <nome> <+N/-N>` | Adiciona/remove diamantes |
| `/gm gold <nome> <+N/-N>` | Adiciona/remove ouro |
| `/gm vip <nome> <dias>` | Concede VIP (0=permanente, -1=revogar) |
| `/gm pix` | Lista pagamentos Pix recentes |

Todas as ações são auditadas em `gm_action_logs`.

---

## Pagamentos Pix (AbacatePay)

**Fluxo:**
1. Jogador escolhe pacote de diamantes
2. Bot gera cobrança via AbacatePay e exibe QR Code Pix
3. Jogador paga → AbacatePay envia `POST /pix/webhook`
4. Bot confirma e credita diamantes (idempotente — sem crédito duplicado)

**Fallback:** polling automático a cada 15s se webhook não estiver configurado.

| Endpoint | Descrição |
|---|---|
| `GET /health` | Health check |
| `POST /pix/webhook` | Notificações AbacatePay |
| `POST /mp/webhook` | Alias de compatibilidade |

---

## Arquitetura técnica

### Princípios

- **Funções puras** no `engine/` e `rpg/` — sem acesso a BD, fáceis de testar
- **Separação de domínio** — cada pacote tem responsabilidade única
- **Extensível sem quebrar** — todos os novos módulos são aditivos (nenhum arquivo existente foi modificado na expansão MMORPG)
- **Thread-safe** — mapas de sessão protegidos por `sync.RWMutex` em `handlers/safe_state.go`; `pvpChallengeTarget` migrado para accessors seguros; singletons globais são thread-safe
- **Event-driven** — sistemas comunicam via `eventbus.Global` (pub/sub assíncrono)
- **Constantes centralizadas** — `internal/config/game_config.go` elimina valores hardcoded espalhados pelo código
- **Skill roles obrigatórios** — todo `SkillNode` declara um `SkillRole` (10 valores: `DIRECT_DAMAGE`, `AOE`, `DOT`, `BUFF`, `DEBUFF`, `CONTROL`, `HEAL`, `UTILITY`, `SUMMON`, `PASSIVE`)
- **Fórmula anti-inflação** — `engine.CalculateDamageLog()` usa `log(lv+1)` limitando dano máx. a ×2.3 no lv 100 (vs ×10 linear)
- **Validação de skills ao startup** — `game/skills.ValidateSkillTrees()` detecta IDs duplicados, mecânicas idênticas, requisitos órfãos e outliers de balanceamento
- **Safe lookups** — `game.GetMonster/Race/Class/Skill/Item` retornam `(T, error)` eliminando nil-pointer panics em produção

### Motor de menus (`internal/ui/`)

O `MenuEngine` centraliza toda a navegação de teclados inline:

```go
// Registrar uma tela
ui.Global.Register(&ui.MenuState{
    ID: "inv_home",
    Render: func(userID int64, args string) ui.Menu {
        items := loadInventory(userID)
        page, _ := strconv.Atoi(args)
        window, totalPages, _ := ui.PaginateItems(items, page)
        // ... build rows
        return ui.Menu{Caption: caption, Rows: rows}
    },
    Handle: func(userID int64, callback string) { /* ... */ },
})

// Navegar com back stack
ui.Global.NavigateTo(userID, "main_menu", "inv_home")
prev := ui.Global.NavigateBack(userID) // retorna "main_menu"
```

| Recurso | Detalhe |
|---|---|
| Debounce | 500 ms por usuário — bloqueia double-tap |
| Cache | TTL 5 s por (state+user+args) — evita DB em navegação rápida |
| Callback limit | `CompressCallback` / `SafeCallback` — enforce ≤ 64 bytes (limite Telegram) |
| Paginação | `PaginateItems[T]` genérico + `NavRow()` para ◀️/▶️ consistentes |
| Back button | `BackButton(callback)` padronizado → "⬅️ Voltar" em toda tela |

### Sistema de roles de habilidades

Cada `SkillNode` agora declara um `SkillRole` obrigatório:

| Role | Mecânica |
|---|---|
| `DIRECT_DAMAGE` | Dano único em alvo |
| `AOE` | Dano em área |
| `DOT` | Dano por turno (veneno, queimadura) |
| `BUFF` | Aumenta stat de aliado/self |
| `DEBUFF` | Reduz stat do inimigo |
| `CONTROL` | Atordoa, congela, silencia |
| `HEAL` | Restaura HP ou remove status |
| `UTILITY` | Mobilidade, escape, gerenciamento de recurso |
| `SUMMON` | Invoca minions ou totens |
| `PASSIVE` | Bônus permanente de stat |

O validador `game/skills.ValidateSkillTrees()` detecta automaticamente:
- **[ERROR]** IDs duplicados entre classes
- **[ERROR]** Pré-requisitos apontando para skills inexistentes
- **[WARN]** Skills sem `Role` definido
- **[WARN]** Mecânicas idênticas (mesmo fingerprint de effects + role)
- **[WARN]** BasePower acima de mean + 2σ do mesmo tier (outlier de balanceamento)
- **[WARN]** IsPassive / Role inconsistentes

### Barramento de eventos (`internal/eventbus/`)

```go
// Publicar
eventbus.Pub(eventbus.NewEvent(eventbus.KindPlayerLevelUp).
    WithPlayer(charID).
    WithInt(newLevel))

// Subscrever
eventbus.Sub(eventbus.KindBossKilled, func(e eventbus.Event) {
    // reagir ao boss morto
})
```

30+ eventos: `PLAYER_LEVEL_UP`, `BOSS_KILLED`, `GUILD_WAR_ENDED`, `INFLATION_CRITICAL`, `SEASON_ENDED`, `TRADE_COMPLETED`…

Pool de **8 goroutines**, fila de **10.000 slots** — sem bloqueio do loop principal.

### Cache (`internal/cache/`)

Interface `Cache` com `Set/Get/Delete/Incr/ZSet/ZTopN/Lock/Unlock`.

- Sem Redis: `MemCache` em memória (padrão, zero dependências extras)
- Com Redis: definir `REDIS_ADDR` e injetar um adapter Redis no `cache.Global`

### Workers de background

| Worker | Intervalo | Responsabilidade |
|---|---|---|
| `EconomyWorker` | 10 min | Snapshot, eventos de inflação, cache |
| `RaidWorker` | 5 min | Spawn/expiração de boss mundial |
| `EventWorker` | Reativo | Reage ao eventbus, broadcast Telegram |
| `CombatPool` | Assíncrono | 16 goroutines para combate pesado |

### Escalabilidade

| Componente | Capacidade atual | Caminho para 1M jogadores |
|---|---|---|
| Update workers | Configurável (`UPDATE_WORKERS`) | Horizontal (múltiplas instâncias) |
| Combat pool | 16 workers, 1000 jobs | Aumentar via env ou Redis queue |
| Event bus | 8 workers, 10k fila | Redis pub/sub para multi-instância |
| Cache | MemCache in-process | Trocar por Redis (interface já existe) |
| DB | PostgreSQL | Connection pool, read replicas, sharding |

---

## Deploy com Docker

```bash
# Build e start
docker compose up -d --build

# Health check
curl http://localhost:8080/health

# Logs ao vivo
docker compose logs -f bot

# Restart apenas do bot
docker compose restart bot

# Parar tudo
docker compose down
```

Volumes persistentes: `postgres_data` (banco) e `bot_assets` (imagens geradas).

### Atualizar sem perder dados

```bash
# Linux/macOS
chmod +x scripts/update.sh
./scripts/update.sh                                          # padrão (com backup)
./scripts/update.sh --migrate-latest                         # aplica última migration
./scripts/update.sh --migration migrations/021_energy_tick_index.sql
```

```powershell
# Windows PowerShell
.\scripts\update.ps1
.\scripts\update.ps1 -MigrateLatest
```

Backups ficam em `./backups/` no formato `pg_<database>_YYYYMMDD_HHMMSS.sql`.

### Restaurar backup

```bash
cat backups/pg_tormenta_rpg_20260308_120000.sql | \
  docker compose exec -T postgres psql -U tormenta -d tormenta_rpg
```

---

## Deploy manual

```bash
# Pré-requisitos: Go 1.21+ e PostgreSQL 16+

go mod download
go build -o tormenta-bot ./cmd/bot/main.go

psql -U postgres -c "CREATE USER tormenta WITH PASSWORD 'tormenta123';"
psql -U postgres -c "CREATE DATABASE tormenta_rpg OWNER tormenta;"
for f in migrations/*.sql; do psql -U tormenta -d tormenta_rpg -f "$f"; done

cp .env.example .env
./tormenta-bot
```

### Systemd

```ini
# /etc/systemd/system/tormenta-bot.service
[Unit]
Description=Tormenta RPG Bot
After=network.target postgresql.service

[Service]
Type=simple
User=tormenta
WorkingDirectory=/opt/tormenta-bot
EnvironmentFile=/opt/tormenta-bot/.env
ExecStart=/opt/tormenta-bot/tormenta-bot
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl enable --now tormenta-bot
```
