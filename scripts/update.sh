#!/usr/bin/env bash
set -euo pipefail

MIGRATION=""
MIGRATE_LATEST="false"
SKIP_BACKUP="false"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --migration)
      MIGRATION="${2:-}"
      shift 2
      ;;
    --migrate-latest)
      MIGRATE_LATEST="true"
      shift
      ;;
    --skip-backup)
      SKIP_BACKUP="true"
      shift
      ;;
    *)
      echo "Uso: $0 [--migration migrations/014_x.sql] [--migrate-latest] [--skip-backup]"
      exit 1
      ;;
  esac
done

if [[ "$MIGRATE_LATEST" == "true" && -n "$MIGRATION" ]]; then
  echo "Use apenas um: --migration ou --migrate-latest"
  exit 1
fi

if ! command -v docker >/dev/null 2>&1; then
  echo "Docker não encontrado no PATH."
  exit 1
fi

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

PG_USER="${POSTGRES_USER:-tormenta}"
PG_DB="${POSTGRES_DB:-tormenta_rpg}"

step() {
  echo "==> $1"
}

if [[ "$SKIP_BACKUP" != "true" ]]; then
  step "Gerando backup do PostgreSQL"
  mkdir -p backups
  TS="$(date +%Y%m%d_%H%M%S)"
  BACKUP_FILE="backups/pg_${PG_DB}_${TS}.sql"
  docker compose exec -T postgres pg_dump -U "$PG_USER" -d "$PG_DB" > "$BACKUP_FILE"
  echo "Backup salvo em: $BACKUP_FILE"
fi

step "Atualizando imagem do PostgreSQL"
docker compose pull postgres

step "Subindo PostgreSQL"
docker compose up -d postgres

step "Rebuild e atualização do bot"
docker compose up -d --build bot

if [[ "$MIGRATE_LATEST" == "true" ]]; then
  MIGRATION="$(ls -1 migrations/*.sql | sort | tail -n 1)"
fi

if [[ -n "$MIGRATION" ]]; then
  if [[ ! -f "$MIGRATION" ]]; then
    echo "Migration não encontrada: $MIGRATION"
    exit 1
  fi
  step "Aplicando migration $MIGRATION"
  docker compose exec -T postgres psql -U "$PG_USER" -d "$PG_DB" -f "$MIGRATION"
fi

step "Status dos containers"
docker compose ps

echo
echo "Atualização concluída."
