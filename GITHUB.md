# GitHub - Descrição Profissional do Repositório

Use os textos abaixo no campo **About** do GitHub.

## Nome sugerido

`tormenta-rpg-bot`

## Descrição curta (campo Description)

Bot de RPG multiplayer para Telegram em Go, com combate por turnos, masmorras, PvP, árvore de habilidades, economia, VIP com auto-caça e pagamentos Pix.

Versão atualizada (com progressão avançada):

Bot de RPG multiplayer para Telegram em Go, com combate por turnos, masmorras, PvP, forja de equipamentos (+1 a +10), crafting, loot tables modulares, VIP com auto-caça e pagamentos Pix.

## About (descrição estendida)

Tormenta RPG Bot é um backend de jogo para Telegram, desenvolvido em Go, com foco em progressão de personagem, combate tático e operação em produção com Docker + PostgreSQL. O projeto inclui sistema de inventário/equipamentos, skills por classe, masmorras, ranking PvP, economia com moedas in-game, VIP com auto-caça e integração de pagamentos via Pix.

Versão atualizada (com novos sistemas):

Tormenta RPG Bot é um backend de RPG para Telegram em Go, com arquitetura modular e foco em escalabilidade. Além de PvE/PvP, inventário e economia, o projeto agora inclui progressão avançada de itens com separação de template/instância (`player_items`), sistema de forja com risco de quebra, crafting baseado em materiais, loot tables por contexto (normal/dungeon/explore/auto_hunt) e base de eventos aleatórios de exploração.

## Topics recomendados

`go`  
`telegram-bot`  
`rpg`  
`turn-based-combat`  
`postgresql`  
`docker`  
`pix`  
`game-backend`  
`pvp`  
`dungeon-crawler`

## Visibilidade recomendada

- Privado: se ainda estiver em evolução interna.
- Público: se quiser portfólio e colaboração.

## Primeira release (sugestão)

Tag: `v1.0.0`  
Título: `v1.0.0 - Core Gameplay + PvP + VIP + Pix`  
Resumo:
- Sistema completo de personagens, combate e progressão.
- PvE (exploração e masmorras) + PvP com ranking.
- Inventário, equipamentos, loja e venda.
- VIP com auto-caça.
- Pagamentos Pix integrados.

## Publicação no GitHub (passo a passo)

### 1) Inicializar e versionar localmente

```bash
git init
git add .
git commit -m "chore: initial release of tormenta rpg bot"
git branch -M main
```

### 2) Criar repositório no GitHub

- Acesse: `https://github.com/new`
- Nome sugerido: `tormenta-rpg-bot`
- Visibilidade: pública ou privada
- Não marque README/.gitignore/licença (já existem localmente)

### 3) Conectar remoto e fazer push

```bash
git remote add origin https://github.com/SEU_USUARIO/tormenta-rpg-bot.git
git push -u origin main
```

### 4) Configurar About

- Description: use a seção "Descrição curta" deste arquivo.
- About: use a seção "About (descrição estendida)".
- Topics: use a lista recomendada.
