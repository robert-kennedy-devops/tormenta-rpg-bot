Param(
    [string]$Migration = "",
    [switch]$MigrateLatest,
    [switch]$SkipBackup
)

$ErrorActionPreference = "Stop"

function Run-Step {
    param([string]$Message, [scriptblock]$Action)
    Write-Host "==> $Message"
    & $Action
}

if (-not (Get-Command docker -ErrorAction SilentlyContinue)) {
    throw "Docker não encontrado no PATH."
}

$projectRoot = Resolve-Path (Join-Path $PSScriptRoot "..")
Set-Location $projectRoot

$pgUser = if ($env:POSTGRES_USER) { $env:POSTGRES_USER } else { "tormenta" }
$pgDb = if ($env:POSTGRES_DB) { $env:POSTGRES_DB } else { "tormenta_rpg" }

if (-not $SkipBackup) {
    Run-Step "Gerando backup do PostgreSQL" {
        $backupDir = Join-Path $projectRoot "backups"
        New-Item -ItemType Directory -Force -Path $backupDir | Out-Null
        $timestamp = Get-Date -Format "yyyyMMdd_HHmmss"
        $backupFile = Join-Path $backupDir "pg_${pgDb}_${timestamp}.sql"
        $cmd = "docker compose exec -T postgres pg_dump -U $pgUser -d $pgDb > `"$backupFile`""
        cmd /c $cmd | Out-Null
        Write-Host "Backup salvo em: $backupFile"
    }
}

Run-Step "Atualizando imagem do PostgreSQL" {
    docker compose pull postgres
}

Run-Step "Subindo PostgreSQL" {
    docker compose up -d postgres
}

Run-Step "Rebuild e atualização do bot" {
    docker compose up -d --build bot
}

if ($MigrateLatest -and $Migration -ne "") {
    throw "Use apenas um: -Migration ou -MigrateLatest."
}

if ($MigrateLatest) {
    $latest = Get-ChildItem -Path (Join-Path $projectRoot "migrations") -Filter "*.sql" |
        Sort-Object Name |
        Select-Object -Last 1
    if ($null -eq $latest) {
        throw "Nenhum arquivo .sql encontrado em ./migrations."
    }
    $Migration = "migrations/$($latest.Name)"
}

if ($Migration -ne "") {
    if (-not (Test-Path $Migration)) {
        throw "Migration não encontrada: $Migration"
    }
    Run-Step "Aplicando migration $Migration" {
        docker compose exec -T postgres psql -U $pgUser -d $pgDb -f $Migration
    }
}

Run-Step "Status dos containers" {
    docker compose ps
}

Write-Host "`nAtualização concluída."
