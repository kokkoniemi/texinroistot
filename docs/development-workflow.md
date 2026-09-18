# Development workflow

## Coding agent instructions

Follow [AGENTS.md](../AGENTS.md) for repository rules, safety constraints, and required checks.

## Local stack (Docker Compose)

```bash
docker compose up --build
```

Services:

- database: Postgres `localhost:5432`
- backend API: `localhost:6969`
- frontend dev server: `localhost:5173`

## Schema bootstrap

```bash
./scripts/init_schema.sh
```

The pinned `golang-migrate` v4.20.1 image applies all forward migrations from `texinroistot-server/internal/db/migrations/`. Repeating bootstrap is a no-op when current. Compose waits for database health and successful migration completion before starting the backend or importer. Rebuild the migrator when adding SQL; the bootstrap script does this automatically.

Existing databases created from the old `schema.sql` are not automatically adopted or erased. Migration 1 fails on their existing objects and records dirty state. Use a fresh local database for migration testing; do not use `force` to hide a failed migration.

To inspect the local migration version:

```bash
docker compose run --rm -T migrate -database 'postgres://tex:willer@db:5432/tex?sslmode=disable' version
```

## Migration development and validation

Add paired `000002_description.up.sql` / `000002_description.down.sql` files using the next unused six-digit version. Never edit or renumber merged migrations. Keep changes compatible with the previous application release, use transactions where PostgreSQL supports them, and document recovery for irreversible changes.

Update the reviewed `schema.sql` snapshot alongside new migrations. Validation creates a separate empty database from the snapshot and compares normalized schema-only `pg_dump` output, including columns, defaults, identity sequences, enums, constraints, indexes, and comments. It ignores only migration-tool metadata, ownership/grants, and random dump restriction markers. Never generate this snapshot from production data.

From the repository root:

```bash
python3 -B -m unittest discover -s scripts -p 'test_*.py'
python3 scripts/check_migration_history.py --base-ref origin/main
bash scripts/check_migrations.sh
```

The history guard also checks new versions are newer than merged versions. Its default base is `HEAD`, useful for uncommitted changes. CI compares against the pull request base or the previous `main` commit and fails closed if the base cannot be resolved.

The container check needs Docker, Bash, and standard Unix tools. It builds the migrator from an explicit SQL-only context, starts disposable PostgreSQL 17 with temporary storage and no published ports, checks up/down/up, repeated up, schema parity, relational behavior, and rejection of an unversioned legacy schema. It removes its own containers and temporary image on exit. It never uses the normal Compose stack or runtime env files. Rootless Podman is supported:

```bash
CONTAINER_ENGINE=podman bash scripts/check_migrations.sh
```

`down` is exercised only on this disposable database. CI builds the backend and importer after migration validation succeeds.

## Import latest spreadsheet

```bash
./scripts/import_excel_and_activate_latest.sh
```

Both scripts use Docker Compose services:

- schema init builds and runs the one-shot `migrate` service
- import runs `docker compose --profile tools run --rm import`

## Backend commands

From `texinroistot-server`:

- tests: `go test ./...`
- build: `go build ./...`
- run api: `go run cmd/server/server.go`
- run importer: `go run cmd/importer/importer.go`

## Frontend commands

From `texinroistot-ui`:

- install: `npm ci`
- type check: `npm run check`
- lint: `npm run lint`
- build: `npm run build`
- dev: `npm run dev -- --host 0.0.0.0`

## CI parity checks before push

Backend:

```bash
cd texinroistot-server
go test ./...
go build ./...
```

Frontend:

```bash
cd texinroistot-ui
npm ci
npm run check
npm run lint
npm run build
```

## Image publish workflow

After application and migration checks pass on `main`, `.github/workflows/ci.yml` calls the reusable `.github/workflows/images.yml` to publish four GHCR images:

- backend
- frontend
- importer
- migrator

All four receive matching source-SHA and immutable release tags before GitHub Release creation. Pull requests run checks only. See [releases and retries](releases-and-deployment.md).
