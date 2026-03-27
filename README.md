# ⚔️ Tormenta RPG Bot

[![Go](https://img.shields.io/badge/Go-1.21-00ADD8?style=flat&logo=go&logoColor=white)](https://go.dev/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-336791?style=flat&logo=postgresql&logoColor=white)](https://www.postgresql.org/)
[![Docker](https://img.shields.io/badge/Docker-Compose-2496ED?style=flat&logo=docker&logoColor=white)](https://www.docker.com/)
[![License](https://img.shields.io/badge/License-MIT-green?style=flat)](LICENSE)

Bot MMORPG multiplayer para Telegram, inspirado no sistema Tormenta 20.  
Arquitetura modular em Go, projetada para escalar de centenas a **1 milhão de jogadores**.

---

## Stack técnica

| Componente | Tecnologia |
|---|---|
| Linguagem | Go 1.21 |
| Telegram API | `go-telegram-bot-api/v5` |
| Banco de dados | PostgreSQL 16 com 21 migrações incrementais |
| Cache | In-memory com interface para Redis (`internal/cache/`) |
| Pagamentos | AbacatePay (Pix) via webhook HTTP |
| Imagens | Go nativo (`image/draw`, `image/png`) |
| Containerização | Docker + Docker Compose |
| Hot reload (dev) | Air (`.air.toml`) |

---

## Arquitetura

O projeto segue separação clara de responsabilidades com pacotes independentes e interfaces bem definidas:

```
tormenta-rpg-bot/
├── cmd/bot/            # Entrypoint: inicialização, bot, webhook HTTP, workers
├── internal/
│   ├── engine/         # Motor de combate — sem efeitos colaterais, 100% testável
│   ├── rpg/            # Sistema RPG completo (raças, classes, habilidades, XP)
│   ├── rpgdata/        # Biblioteca de dados pura (sem I/O) — 750+ itens, 59+ monstros
│   ├── game/           # Lógica de jogo principal e handlers de estado
│   ├── economy/        # Economia controlada anti-inflação
│   ├── market/         # Mercado player-to-player e sistema de leilão
│   ├── guild/          # Sistema completo de guildas (banco, guerras, territórios)
│   ├── world/          # Bosses mundiais e raids multi-fase
│   ├── pvp/            # Arena PvP com sistema Elo e matchmaking
│   ├── season/         # Temporadas com reset de ranking e recompensas por tier
│   ├── eventbus/       # Bus assíncrono com pool de 8 goroutines e fila de 10k slots
│   ├── workers/        # Workers de background (combate, economia, eventos, raids)
│   ├── cache/          # Camada unificada: TTL cache → Redis fallback
│   ├── security/       # Validação, rate limiting, anti-exploit, detecção de anomalias
│   ├── ui/             # Motor de menus Telegram com debounce e cache de estado
│   ├── repository/     # Contratos e adapters de persistência (interface segregation)
│   └── models/         # Structs compartilhadas: Character, Player, Item, Monster…
├── migrations/         # 21 migrações SQL incrementais e versionadas
├── Dockerfile
├── docker-compose.yml
└── .env.example
```

### Decisões de design

**Engine de combate desacoplada** — `internal/engine/` não importa nada de `game/`, `handlers/` ou banco de dados. Recebe e retorna structs simples, facilitando testes unitários isolados.

**Event Bus** — subsistemas (combate, economia, raids) se comunicam via eventos assíncronos em vez de chamadas diretas. Pool fixo de goroutines evita criação desenfreada de goroutines.

**Camada de segurança dedicada** — `internal/security/` cobre: rate limiting por ação por usuário, deduplicação de callbacks Telegram (idempotência), detecção de speed-hack e farm de dungeon, validação de integridade econômica e logger JSON estruturado.

**Repository pattern** — toda persistência passa por interfaces em `internal/repository/`, facilitando troca entre MemStore (testes) e DBStore (PostgreSQL em produção).

---

## Como rodar

### Com Docker (recomendado)

```bash
git clone https://github.com/robert-kennedy-devops/tormenta-rpg-bot.git
cd tormenta-rpg-bot
cp .env.example .env
# edite .env com seu token do Telegram e credenciais do banco
docker compose up -d
docker compose logs -f bot
```

O PostgreSQL inicializa automaticamente com todas as migrações em `./migrations/` em ordem.

### Sem Docker

```bash
# Pré-requisitos: Go 1.21+, PostgreSQL 16
psql -U postgres -c "CREATE DATABASE tormenta_rpg;"
for f in migrations/*.sql; do psql -U tormenta -d tormenta_rpg -f "$f"; done
go run ./cmd/bot
```

---

## Variáveis de ambiente

```env
# Obrigatórias
TELEGRAM_TOKEN=123456789:ABCdef...
DATABASE_URL=postgres://user:senha@localhost:5432/tormenta_rpg?sslmode=disable
GM_IDS=123456789,987654321          # IDs Telegram dos admins, separados por vírgula

# Opcionais
ABACATEPAY_TOKEN=acp_...            # Habilita pagamentos via Pix
MP_WEBHOOK_PORT=8080                # Porta HTTP para webhooks (padrão: 8080)
REDIS_ADDR=localhost:6379           # Ativa cache distribuído Redis
UPDATE_WORKERS=1                    # Workers de processamento de updates
DUNGEON_PROCEDURAL_ENABLED=false    # true = dungeons procedurais (5–10 salas)
ECONOMY_DYNAMIC_PRICING=false       # Precificação dinâmica progressiva na loja NPC
```

> Nunca faça commit do `.env`. Use `.env.example` como referência. Se tokens reais forem expostos, faça rotação imediata.

---

## Banco de dados

Schema versionado com 21 migrações incrementais (`migrations/001_init.sql` → `migrations/021_energy_tick_index.sql`).

**Tabelas principais:**

| Tabela | Descrição |
|---|---|
| `players` | Conta Telegram (id, ban, VIP) |
| `characters` | Personagem ativo (stats, localização, estado) |
| `inventory` / `player_items` | Itens e instâncias de equipamentos |
| `character_skills` | Árvore de habilidades desbloqueadas |
| `dungeon_runs` / `dungeon_best` | Histórico e recordes de masmorras |
| `pvp_challenges` / `pvp_ratings` | Duelos, histórico Elo por temporada |
| `pix_payments` | Pagamentos Pix com idempotência |
| `auto_hunt_sessions` | Caça automática VIP |
| `guilds` / `guild_members` | Estrutura de guildas e membros |
| `market_listings` / `market_auctions` | Mercado player-to-player e leilões |
| `world_boss_log` | Participação e dano por jogador em bosses |
| `seasons` / `season_rewards` | Histórico de temporadas e recompensas |
| `gm_action_logs` | Auditoria de ações administrativas |
| `analytics_events` | Telemetria opcional |

---

## Segurança

Camada dedicada em `internal/security/`:

- **Rate limiting por ação** — cada tipo de ação (combate, loja, dungeon) tem seu próprio bucket de rate limit por usuário
- **Deduplicação de callbacks** — callbacks Telegram duplicados (rede instável, double-tap) são descartados sem processar duas vezes
- **Anti-exploit de transações** — operações econômicas usam idempotência para evitar duplicação de gold/itens
- **Detecção de comportamento anômalo** — identificação de bot-pattern (ações uniformemente espaçadas) e dungeon farm excessivo
- **Validação de economia** — verificação de integridade antes de qualquer transferência de recursos
- **Logger JSON estruturado** — todos os eventos de segurança são logados com contexto (user_id, action, timestamp)

---

## Escalabilidade

- **Workers de background** com goroutines e canais de capacidade fixa — sem goroutine leak
- **Event bus** assíncrono desacopla subsistemas sem bloquear o loop principal do bot
- **Cache em camadas** — hot data em memória com interface para escalar para Redis distribuído
- **Repository pattern** permite troca de storage sem alterar lógica de negócio
- **Modelagem de schema** compatível com read replicas e particionamento futuro

---

## Conteúdo gerado (rpgdata)

Todos os dados RPG são gerados programaticamente via templates — sem JSON/CSV estático:

- 8 classes · 7 raças · progressão do nível 1 ao 100
- 210 habilidades em `internal/rpg/` (árvores com pré-requisitos, tiers T1–T5)
- 102 habilidades adicionais em `internal/game/skills/` com validação automática (IDs duplicados, anomalias de balanceamento por mean+2σ)
- 750+ itens gerados (25 templates × 5 tiers × 6 raridades)
- 59+ monstros (13 arquétipos × bandas de nível)
- 20 combos de 3 passos com janela temporal
- Maestria em 6 níveis (0–300 usos acumulados)

---

## Gameplay (resumo)

<details>
<summary>Ver raças, classes e progressão</summary>

### Raças

| Raça | Traço passivo |
|---|---|
| 👤 Humano | +10% XP ganho |
| 🧝 Elfo | +20% dano mágico |
| ⛏️ Anão | -15% dano recebido, imune a Freeze |
| 👹 Meio-Orc | +25% dano físico, Berserk sob 20% HP |
| 👺 Goblin | +15% chance crítico |
| 🧞 Qareen | +30% dano de Fogo, imune a Burn |
| 🐂 Minotauro | Ignora 10% de armadura |

### Classes

| Classe | Função | Traço |
|---|---|---|
| ⚔️ Guerreiro | Tanque | Maestria em Armas |
| 🧙 Arcanista | Conjurador | Surto Arcano (a cada 4 feitiços, 1 grátis) |
| 🗡️ Ladino | DPS | Facada pelas Costas (+40% dano em flanqueamento) |
| 🏹 Caçador | Distância | Olho de Águia (+10% crit à distância) |
| ⚜️ Paladino | Tanque/Suporte | Aura Sagrada |
| ✝️ Clérigo | Curandeiro | Graça Divina (10% curas duplas) |
| 🪓 Bárbaro | DPS brutal | Fúria Bárbara (+60% ATK em Berserk) |
| 🎵 Bardo | Suporte | Inspiração (+15% XP pós-batalha) |

### Progressão

Nível 1–100 com curva XP suave. Milestones em 10, 25, 50, 75 e 100 com recompensas extras.

</details>

---

## Autor

**Robert Kennedy** — Desenvolvedor Backend · Go · Manaus, AM  
[LinkedIn](https://www.linkedin.com/in/robert-kennedy-034687369/) · [GitHub](https://github.com/robert-kennedy-devops)
