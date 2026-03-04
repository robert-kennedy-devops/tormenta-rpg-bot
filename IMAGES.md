# Sistema de Imagens - Tormenta RPG Bot

## Visão geral

O bot gera imagens automaticamente em Go no startup e usa cache de `file_id` do Telegram para acelerar envios.

Fluxo:
1. `assets.Init()` chama `GenerateAllImages()`.
2. Arquivos PNG faltantes são criados em `assets/images/`.
3. No primeiro envio, o bot faz upload do arquivo.
4. O Telegram retorna `file_id` e o bot salva em `image_cache`.
5. Nos próximos envios, usa `file_id` (mais rápido).

Observação importante:
- O gerador não sobrescreve arquivos existentes. Se você customizar um PNG, ele será preservado.

---

## Estrutura atual

```text
assets/images/
├── ui/
│   ├── welcome.png
│   ├── menu.png
│   ├── status.png
│   ├── inventory.png
│   ├── skills.png
│   ├── shop.png
│   ├── travel.png
│   ├── combat.png
│   ├── victory.png
│   ├── defeat.png
│   └── rest.png
├── races/
│   ├── human.png
│   ├── elf.png
│   ├── dwarf.png
│   └── halforc.png
├── classes/
│   ├── warrior.png
│   ├── mage.png
│   ├── rogue.png
│   └── archer.png
├── monsters/
│   ├── rat.png
│   ├── goblin.png
│   ├── slime.png
│   ├── wolf.png
│   ├── orc.png
│   ├── troll.png
│   ├── bandit_leader.png
│   ├── bat.png
│   ├── spider.png
│   ├── golem.png
│   ├── undead_knight.png
│   ├── demon.png
│   ├── necromancer.png
│   ├── vampire_lord.png
│   ├── dragon_young.png
│   ├── dragon_elder.png
│   ├── mushroom.png
│   ├── crow.png
│   ├── harpy.png
│   ├── werewolf.png
│   ├── stone_golem_shard.png
│   ├── crystal_wraith.png
│   ├── shadow_assassin.png
│   ├── lich.png
│   ├── wyvern.png
│   └── phoenix.png
├── maps/
│   ├── village.png
│   ├── village_outskirts.png
│   ├── dark_forest.png
│   ├── forest_camp.png
│   ├── crystal_cave.png
│   ├── ancient_dungeon.png
│   ├── dungeon_outpost.png
│   └── dragon_peak.png
└── items/
    ├── weapons.png
    ├── armors.png
    └── consumables.png
```

Total atual: **56 imagens**.

---

## Chaves de imagem usadas no código

Definidas em `internal/assets/manager.go` (map `ImageKey`):

- UI: `welcome`, `menu`, `status`, `inventory`, `skills`, `shop`, `travel`, `combat`, `victory`, `defeat`, `rest`
- Raças: `race_human`, `race_elf`, `race_dwarf`, `race_halforc`
- Classes: `class_warrior`, `class_mage`, `class_rogue`, `class_archer`
- Monstros: `monster_*` (26 no total)
- Mapas: `map_*` (8 no total)
- Itens: `item_weapon`, `item_armor`, `item_consumable`

Helpers importantes:
- `assets.MonsterImageKey(monsterID)` -> `monster_<id>`
- `assets.MapImageKey(mapID)` -> `map_<id>`
- `assets.RaceImageKey(raceID)` -> `race_<id>`
- `assets.ClassImageKey(classID)` -> `class_<id>`
- `assets.ItemTypeImageKey(itemType)` -> `item_<type>`

---

## Como substituir por arte personalizada

1. Gere a base uma vez:

```bash
go run ./cmd/bot/main.go
```

2. Substitua os arquivos em `assets/images/` mantendo:
- formato PNG
- nome exato do arquivo
- tamanho recomendado 512x512

3. Limpe o cache da imagem alterada para forçar reupload:

```sql
DELETE FROM image_cache WHERE key = 'monster_dragon_elder';
```

4. Reinicie o bot.

---

## Cache e fallback

Comportamento em `internal/handlers/media.go`:
- Se existir `file_id` em cache, envia por `FileID`.
- Se não existir, envia por `FilePath` e salva `file_id`.
- Se imagem/chave não existir, cai para mensagem de texto.
- Se tentar editar mídia em mensagem que não é foto, o bot recria a mensagem como foto.

---

## Debug rápido

Ver imagens cacheadas:

```sql
SELECT key, LEFT(file_id, 20) || '...' AS file_id, updated_at
FROM image_cache
ORDER BY updated_at DESC;
```

Forçar reupload de uma imagem:

```sql
DELETE FROM image_cache WHERE key = 'monster_rat';
```

Forçar reupload de todas:

```sql
TRUNCATE image_cache;
```

Logs de imagem (Linux/macOS):

```bash
docker compose logs -f bot | grep "🖼️"
```

Logs de imagem (PowerShell):

```powershell
docker compose logs -f bot | Select-String "🖼️"
```

---

## Como adicionar uma nova imagem

1. Adicione a chave em `internal/assets/manager.go`:

```go
"monster_dragon_god": "monsters/dragon_god.png",
```

2. Adicione o `imageSpec` em `internal/assets/generator.go` para fallback gerado.

3. Use no handler:

```go
editPhoto(chatID, msgID, "monster_dragon_god", caption, &kb)
```

4. Reinicie o bot para gerar o novo PNG automaticamente (se ainda não existir).
