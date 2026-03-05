# ⚔️ Tormenta RPG Bot

Bot de RPG multiplayer para Telegram, inspirado no sistema Tormenta. Criação de personagens, combate por turnos, masmorras, PvP, loja, sistema VIP com caça automática e pagamentos via Pix (AbacatePay).

---

## Índice

1. [Stack e dependências](#stack-e-dependências)
2. [Configuração rápida](#configuração-rápida)
3. [Variáveis de ambiente](#variáveis-de-ambiente)
4. [Banco de dados](#banco-de-dados)
5. [Atualizações recentes](#atualizações-recentes)
6. [Estrutura do projeto](#estrutura-do-projeto)
7. [Gameplay](#gameplay)
8. [Comandos de GM](#comandos-de-gm)
9. [Pagamentos Pix](#pagamentos-pix-abacatepay)
10. [Deploy com Docker](#deploy-com-docker)
11. [Deploy manual](#deploy-manual)

---

## Stack e dependências

| Componente | Tecnologia |
|---|---|
| Linguagem | Go 1.21 |
| Telegram API | `go-telegram-bot-api/v5` |
| Banco de dados | PostgreSQL 16 |
| Pagamentos | AbacatePay (Pix) |
| Geração de imagens | Go nativo (`image/draw`, `image/png`) |
| Containerização | Docker + Docker Compose |

---

## Configuração rápida

```bash
git clone https://github.com/seu-usuario/tormenta-bot.git
cd tormenta-bot
cp .env.example .env
# edite o .env
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
DUNGEON_PROCEDURAL_ENABLED=false     # true=gera dungeons procedurais (5-10 salas)
ABACATEPAY_WEBHOOK_SECRET=change-me  # valida header X-AbacatePay-Secret (opcional)
ECONOMY_DYNAMIC_PRICING=false        # preço dinâmico progressivo na loja NPC
```

**Importante:** nunca faça commit do arquivo `.env`. Se você chegou a compartilhar tokens reais (Telegram/AbacatePay), faça **rotação** imediata no painel correspondente.

---

## Banco de dados

O projeto usa **migrações incrementais** em `migrations/001_init.sql` até `migrations/018_economy_usage_stats.sql`.

Para subir do zero (manual), aplique os `.sql` em ordem:

```bash
psql -U postgres -c "CREATE DATABASE tormenta_rpg;"
for f in migrations/*.sql; do psql -U tormenta -d tormenta_rpg -f "$f"; done
```

Com Docker, o PostgreSQL executa automaticamente todos os arquivos `.sql` de `./migrations/` em ordem alfabética.

Evoluções recentes de schema:
- `013_backfill_equipped_slots.sql`: normalização/backfill de slots equipados.
- `014_shield_offhand_slot.sql`: cria e ajusta o slot dedicado `offhand` (escudo).
- `015_player_items.sql`: adiciona instâncias de item por personagem para progressão de forja (`upgrade_level`, quebra e equip por instância).
- `016_energy_timestamp.sql`: energia com cálculo por timestamp (`last_energy_update`) sem worker de regen.
- `017_player_timers.sql`: timers genéricos por jogador (`player_timers`) para cooldowns/sistemas.
- `018_economy_usage_stats.sql`: estatísticas de compra para economia dinâmica (`item_usage_stats`).

## Atualizações recentes

- Progressão avançada (Fase 2):
  - Novo domínio de itens em `internal/items` com separação entre **template** e **instância de jogador**.
  - Sistema de **forja** em `internal/forge` com upgrade de `+1` até `+10`, chance configurável e risco de quebra a partir de `+5`.
  - Sistema de **crafting** em `internal/crafting` com receitas e consumo de materiais.
  - Sistema de **drops modulares** em `internal/drops` e serviço em `internal/services/drop_service.go`.
  - Eventos de **exploração aleatória** em `internal/explore`.
  - Integração inicial no jogo:
    - Menus `🔨 Forja` e `🧰 Crafting` no menu principal.
    - Drops de materiais ativos em combate normal, dungeon e auto-caça.
    - Auto-caça com chance reduzida por multiplicador (`5%` base -> `1.5%` efetivo).
- Arquitetura e escalabilidade:
  - Novo processamento concorrente de updates em `internal/bot/update_worker.go`.
  - Rate limiting de chamadas Telegram em `internal/bot/rate_limiter.go`.
  - Roteamento dedicado de mensagens/callbacks em `internal/router/`.
  - Engine de menus reutilizável em `internal/menu/` para reduzir duplicação de teclado inline.
  - Camada inicial de serviços em `internal/services/` para separar lógica de negócio dos handlers.
  - Auto-caça em modo offline: processamento por ciclos com base em `last_tick_at` quando o jogador retorna.
  - Worker manager central em `internal/systems/workers` (pix, eventos, limpeza e manutenção).
  - Serviço de pagamento idempotente em `internal/services/payment`.
  - Camada anti-cheat em `internal/services/anti_cheat` (duplicidade de callback e transições inválidas).
  - Eventos globais em `internal/systems/events` (Blood Moon, Tormenta Storm, Double Drop).
- Navegação:
  - Botões `⬅️ Voltar` padronizados com destino contextual (inventário, loja, habilidades, vender).
  - `🛒 Loja` e `💰 Vender` disponíveis de forma fixa no menu principal.
- Inventário e equipamentos:
  - Fluxo de equipar acessórios sem abrir janela secundária; mantém o usuário no fluxo principal.
  - Suporte a slot `offhand` para escudos na tela/equipamento e no cálculo de status.
- Combate e efeitos:
  - Sistema de efeitos temporários em `internal/handlers/effects.go` integrado ao PvE, PvP, masmorra e auto-caça.
  - Suporte a buffs/debuffs de CA, penalidade de ataque, redução/aumento de dano, crítico forçado/ampliado, queimadura e perda de turno.
  - Veneno aplicado por monstros e por skills do jogador (linha Envenenador do Ladino), com DoT em PvE, PvP, masmorra e auto-caça.
- Operação:
  - Scripts de atualização: `scripts/update.ps1` e `scripts/update.sh` com backup de banco e migração opcional.

### Tabelas

| Tabela | Descrição |
|---|---|
| `players` | Conta Telegram (id, ban, VIP) |
| `characters` | Personagem (stats, localização, estado) |
| `inventory` | Itens por personagem |
| `character_skills` | Habilidades aprendidas |
| `dungeon_runs` / `dungeon_best` | Histórico e recordes de masmorras |
| `pvp_challenges` / `pvp_stats` | Duelos e ranking ELO |
| `pix_payments` | Pagamentos Pix |
| `auto_hunt_sessions` | Sessões de caça automática VIP |
| `player_timers` | Timers/cooldowns genéricos por jogador |
| `item_usage_stats` | Estatística de uso para precificação dinâmica |
| `diamond_log` | Log de transações de diamantes |
| `combat_log` | Histórico de combates |
| `daily_bonus` | Controle de bônus diário |
| `image_cache` | Cache de `file_id` do Telegram |
| `player_items` | Instâncias de itens equipáveis por personagem (forja/progressão) |

---

## Estrutura do projeto

```
tormenta-bot/
├── cmd/bot/
│   └── main.go                 # Entrypoint: bot, webhook HTTP, workers
├── internal/
│   ├── bot/
│   │   ├── rate_limiter.go     # Limitador de chamadas para API do Telegram
│   │   └── update_worker.go    # Worker pool para processamento de updates
│   ├── assets/
│   │   ├── generator.go        # Geração procedural de imagens de personagem
│   │   └── manager.go          # Cache de imagens no Telegram
│   ├── database/
│   │   ├── database.go         # Conexão e queries SQL
│   │   ├── image_cache.go      # Implementação de cache de file_id no banco
│   │   └── player_items.go     # Persistência de instâncias de itens (forja)
│   ├── game/
│   │   ├── combat.go           # Engine de combate por turnos
│   │   ├── data.go             # Raças, classes, monstros, mapas, itens, drops
│   │   ├── dungeon_logic.go    # Lógica de masmorras
│   │   ├── energy.go           # Sistema de energia e regeneração
│   │   ├── game_pix.go         # Integração AbacatePay
│   │   ├── pvp_game.go         # Lógica de PvP
│   │   └── extended_items.go   # Itens estendidos (materiais/crafting)
│   ├── handlers/
│   │   ├── handlers.go         # Handler principal: mensagens e callbacks
│   │   ├── dungeon_handler.go  # Handlers de dungeon
│   │   ├── effects.go          # Estado/efeitos temporários de combate
│   │   ├── gm.go               # Painel e comandos de GM
│   │   ├── media.go            # Envio de imagens com fallback texto
│   │   ├── pix_handler.go      # Compra, polling e webhook de Pix
│   │   ├── pvp_handler.go      # Handlers de PvP
│   │   ├── rank.go             # Ranking global
│   │   ├── vip.go              # VIP: painel, compra, caça automática
│   │   ├── drop_materials.go   # Aplicação de drops de materiais por modo
│   │   └── progression.go      # Menus e fluxo de Forja/Crafting
│   ├── items/
│   │   ├── types.go            # ItemTemplate, PlayerItem e stat blocks
│   │   └── materials.go        # Catálogo de materiais de forja/crafting
│   ├── forge/
│   │   └── forge.go            # Regras de sucesso/falha/quebra da forja
│   ├── drops/
│   │   ├── loot.go             # Loot tables por modo (normal/dungeon/auto)
│   │   └── default_tables.go   # Tabelas padrão de materiais
│   ├── crafting/
│   │   └── crafting.go         # Receitas e consumo de materiais
│   ├── dungeon/
│   │   └── generator.go        # Geração procedural de salas (5-10)
│   ├── systems/
│   │   ├── workers/
│   │   │   └── manager.go      # Workers centralizados de manutenção/pix/eventos
│   │   └── events/
│   │       └── world_events.go # Eventos globais temporários
│   ├── explore/
│   │   └── events.go           # Eventos aleatórios de exploração
│   ├── menu/
│   │   ├── engine.go           # Helpers de botões/linhas/teclados inline
│   │   ├── main_menu.go        # Menu principal
│   │   ├── shop_menu.go        # Menus da loja
│   │   ├── inventory_menu.go   # Menus do inventário
│   │   └── ...                 # Menus de VIP, PvP, ranking, pix e dungeon
│   ├── router/
│   │   ├── callback_router.go  # Roteador de ações de callback
│   │   └── message_router.go   # Roteador de mensagens de texto/comando
│   ├── services/
│   │   ├── player_service.go   # Regras de jogador
│   │   ├── shop_service.go     # Regras de loja/compra
│   │   ├── combat_service.go   # Regras de combate
│   │   ├── autohunt_service.go # Processamento offline de ciclos da auto-caça
│   │   ├── drop_service.go     # Serviço de drop por contexto de jogo
│   │   ├── payment/
│   │   │   └── service.go      # Confirmação PIX idempotente + validação antifraude
│   │   └── anti_cheat/
│   │       └── guard.go        # Guards de callback e transição de estado
│   ├── timers/
│   │   └── store.go            # Persistência de timers/cooldowns
│   └── models/
│       └── models.go           # Structs: Character, Player, Item, Monster...
├── migrations/
│   ├── 001_init.sql
│   ├── ...
│   ├── 013_backfill_equipped_slots.sql
│   ├── 014_shield_offhand_slot.sql
│   ├── 015_player_items.sql
│   ├── 016_energy_timestamp.sql
│   ├── 017_player_timers.sql
│   └── 018_economy_usage_stats.sql
├── scripts/
│   ├── update.ps1              # Atualização com backup (Windows)
│   └── update.sh               # Atualização com backup (Linux/macOS)
├── assets/images/              # Imagens geradas (não versionado)
├── Dockerfile
├── docker-compose.yml
└── .env.example
```

---

## Gameplay

### Raças e Classes

Ao criar um personagem o jogador escolhe uma raça e uma classe.

**Raças:**

| Raça | Traço | Bônus principal |
|---|---|---|
| 👤 Humano | +10% de XP ganho | Atributos equilibrados |
| 🧝 Elfo | +20% dano mágico | DEX +3, INT +3 |
| ⛏️ Anão | -15% dano recebido | CON +4, HP +25 |
| 👹 Meio-Orc | +25% dano físico | STR +4, HP +20 |

**Classes:**

| Classe | Função | HP base | MP base |
|---|---|---|---|
| ⚔️ Guerreiro | Tanque | 80 | 20 |
| 🧙 Mago | Conjurador | 45 | 80 |
| 🗡️ Ladino | DPS | 55 | 40 |
| 🏹 Arqueiro | Distância | 60 | 35 |

---

### Zonas e progressão

O mapa é linear — viajar entre zonas custa 1⚡.

| Zona | Nível | Monstros |
|---|---|---|
| 🏘️ Vila de Trifort | Qualquer | Hub com loja e estalagem |
| 🌾 Arredores da Vila | 1–5 | Rato, Goblin, Slime, Cogumelo, Corvo |
| 🌲 Floresta Sombria | 4–9 | Lobo, Orc, Troll, Harpia, Lobisomem |
| 💎 Caverna de Cristal | 8–13 | Morcego, Aranha, Golem, Cavaleiro Morto-Vivo |
| 🏚️ Masmorra Antiga | 13–18 | Demônio, Necromante, Lorde Vampiro, Lich |
| 🏔️ Pico dos Dragões | 17–20 | Dragão Jovem, Dragão Ancião, Wyvern, Fênix |

---

### Sistema de energia ⚡

Energia é o recurso central — toda ação consome 1⚡.

| Ação | Custo |
|---|---|
| Viajar para outra zona | 1⚡ |
| Iniciar combate | 1⚡ |
| Andar em masmorra | 1⚡ |
| Tick de caça automática | 1⚡ |

**Regeneração passiva** — calculada automaticamente na próxima interação:

| Status | Máximo | Regeneração |
|---|---|---|
| Normal | 100 + (nível − 1) × 2 | 1⚡ / 10 min |
| 👑 VIP | 200 + (nível − 1) × 4 | 1⚡ / 5 min |

---

### Combate

Combate por turnos contra monstros sorteados na zona.

1. Acessa "🗺️ Explorar" → 1 monstro sorteado entre os disponíveis na zona
2. Por turno: **Atacar** / **Habilidade** / **Item** / **Fugir**
3. Vitória: XP + ouro + chance de drop
4. Derrota: perde XP e ouro proporcionais ao nível, retorna à Vila

Efeitos de combate ativos:
- Veneno (DoT em player/monstro), queimadura, debuffs de CA/ataque, buffs defensivos/ofensivos e manipulação de crítico.
- Aplicação consistente em exploração, PvP, masmorra e auto-caça.

---

### Masmorras

Sequências de andares com dificuldade crescente, cada andar custa 1⚡. A recompensa final inclui ouro escalado e itens raros. Os recordes são salvos em `dungeon_best`.
Com `DUNGEON_PROCEDURAL_ENABLED=true`, cada run gera automaticamente entre 5 e 10 salas (`monster/treasure/trap/elite/boss`) de forma determinística por sessão.

---

### PvP

Duelos 1v1 com aposta de ouro. O desafiado tem 5 minutos para aceitar. Combate automático sem gasto de ⚡. Rating Elo atualizado em `pvp_stats`.

---

### Economia

**Ouro 🪙** — obtido em combate e venda de itens. Usado na loja para equipamentos.

**Diamantes 💎** — moeda premium. Obtida via bônus diário, drops raros ou compra por Pix. Usada para VIP e loja premium.

**Materiais de Progressão 🧱** — usados em forja e crafting:
- Pedra de Forja
- Pedra Refinada
- Essência Arcana
- Fragmento de Monstro
- Metal Negro

---

### Forja

Equipamentos podem ser aprimorados de `+1` até `+10`.

| Nível alvo | Chance |
|---|---|
| +1 | 100% |
| +2 | 90% |
| +3 | 80% |
| +4 | 70% |
| +5 | 60% |
| +6 | 50% |
| +7 | 40% |
| +8 | 30% |
| +9 | 20% |
| +10 | 10% |

Regras de falha:
- Até `+4`: falha segura (não quebra).
- De `+5` em diante: falha pode quebrar item.

---

### Crafting

Sistema de receitas com consumo de materiais.

Exemplo:
- **Espada Negra**
  - `3x Metal Negro`
  - `1x Essência Arcana`

---

### Drops e exploração

- Drops usam loot tables por contexto (`normal`, `dungeon`, `explore`, `auto_hunt`).
- Auto-caça aplica multiplicador reduzido para manter balanceamento de economia.
- Base para exploração aleatória pronta com eventos:
  - monstro
  - tesouro
  - evento raro
  - nada

---

### Sistema VIP

| Plano | Custo | Duração |
|---|---|---|
| 30 dias | 500 💎 | 30 dias |
| 90 dias | 1.200 💎 | 90 dias |
| Permanente | 3.000 💎 | Vitalício |

**Benefícios VIP:**
- Energia máxima dobrada
- Regeneração 2× mais rápida
- 🤖 Caça automática (offline hunting)

**Caça automática:** o jogador seleciona uma zona e inicia a sessão. O progresso é processado em modo offline por ciclos de 60s com base no `last_tick_at` quando o jogador abre painel/relatório/stop. Cada ciclo consome 1⚡, sorteia monstro, simula combate e credita recompensas. Para automaticamente se energia zerar, personagem morrer ou VIP expirar.

---

## Comandos de GM

Apenas usuários com ID listado em `GM_IDS` têm acesso. O comando `/gm` sem argumentos abre um painel interativo com botões.

| Comando | Descrição |
|---|---|
| `/gm buscar <nome>` | Busca personagem por nome |
| `/gm info <nome>` | Ficha completa do personagem |
| `/gm id <telegramID>` | Lookup por ID do Telegram |
| `/gm ban <nome> [razão]` | Bane jogador |
| `/gm unban <nome>` | Desbane jogador |
| `/gm diamond <nome> <+N/-N>` | Adiciona ou remove diamantes |
| `/gm gold <nome> <+N/-N>` | Adiciona ou remove ouro |
| `/gm vip <nome> <dias>` | Concede VIP (`0` = permanente, `-1` = revogar) |
| `/gm pix` | Lista pagamentos Pix recentes |

---

## Pagamentos Pix (AbacatePay)

**Fluxo:**
1. Jogador escolhe pacote de diamantes
2. Bot gera cobrança via AbacatePay e exibe QR Code Pix
3. Jogador paga → AbacatePay envia webhook `POST /pix/webhook` (opcionalmente validado por `X-AbacatePay-Secret`)
4. Bot confirma em `pix_payments` e credita diamantes

**Fallback:** sem webhook configurado, o bot faz polling a cada 15 segundos automaticamente.

**Webhook no painel AbacatePay:** `https://seu-dominio.com/pix/webhook`

**Endpoints HTTP:**

| Endpoint | Descrição |
|---|---|
| `GET /health` | Health check |
| `POST /pix/webhook` | Notificações AbacatePay |
| `POST /mp/webhook` | Alias de compatibilidade |

---

## Deploy com Docker

```bash
# Build e start
docker compose up -d --build

# Health check
curl http://localhost:8080/health

# Logs
docker compose logs -f bot

# Restart apenas do bot
docker compose restart bot

# Parar tudo
docker compose down
```

Volumes persistentes: `postgres_data` (banco) e `bot_assets` (imagens geradas).

### Atualizar sem apagar dados

Os dados do banco ficam no volume `postgres_data`, então você **não** precisa apagar containers/images para atualizar.

Foram adicionados scripts de atualização com:
- backup automático do PostgreSQL
- pull/restart do PostgreSQL
- rebuild do bot
- execução opcional de migração

#### Windows (PowerShell)

```powershell
# Atualização padrão (com backup)
.\scripts\update.ps1

# Atualizar e aplicar migration específica
.\scripts\update.ps1 -Migration migrations/014_shield_offhand_slot.sql

# Atualizar e aplicar o arquivo .sql mais recente
.\scripts\update.ps1 -MigrateLatest
```

#### Linux/macOS (Bash)

```bash
# Dar permissão uma vez
chmod +x scripts/update.sh

# Atualização padrão (com backup)
./scripts/update.sh

# Atualizar e aplicar migration específica
./scripts/update.sh --migration migrations/014_shield_offhand_slot.sql

# Atualizar e aplicar o .sql mais recente
./scripts/update.sh --migrate-latest
```

#### Onde ficam os backups

- Pasta: `./backups`
- Formato: `pg_<database>_YYYYMMDD_HHMMSS.sql`

#### Como restaurar um backup

```bash
# Exemplo Linux/macOS
cat backups/pg_tormenta_rpg_20260304_120000.sql | docker compose exec -T postgres psql -U tormenta -d tormenta_rpg
```

```powershell
# Exemplo Windows PowerShell
Get-Content backups\pg_tormenta_rpg_20260304_120000.sql | docker compose exec -T postgres psql -U tormenta -d tormenta_rpg
```

---

## Deploy manual

### Pré-requisitos: Go 1.21+ e PostgreSQL 16+

```bash
# Build
go mod download
go build -o tormenta-bot ./cmd/bot/main.go

# Banco
psql -U postgres -c "CREATE USER tormenta WITH PASSWORD 'tormenta123';"
psql -U postgres -c "CREATE DATABASE tormenta_rpg OWNER tormenta;"
for f in migrations/*.sql; do psql -U tormenta -d tormenta_rpg -f "$f"; done

# Executar
cp .env.example .env && ./tormenta-bot
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
