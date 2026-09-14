# Development workflow

## Coding agent instructions

Maintain shared rules in the root `AGENTS.md`. `CLAUDE.md` and `GEMINI.md` import it; keep backend/frontend checks, migration policy, and deployment authorization rules in that canonical file.

Claude file-read exclusions are in `.claude/settings.json`; Gemini exclusions are in `.geminiignore`. These cover common env, key, VPN, token, dump, and production-data paths. They do not identify every possible secret filename or provide an OS-level boundary for arbitrary shell commands. The secret-handling rules in `AGENTS.md` apply to every tool.

Gemini allows placeholder `*.example` and `*.env-example` templates. Claude deny rules have no allow exceptions, so env templates such as `backend.env.example` are also denied; use the existing `.env-example` naming convention for templates Claude needs to read.

Validate agent setup from a clean clone without runtime files:

1. Start Codex at the repository root and ask it to summarize the loaded repository instructions and required checks.
2. In Claude Code, use `/context` to inspect memory files and `/permissions` to inspect the shared deny rules. Ask it to summarize the imported repository instructions.
3. In Gemini CLI, use `/memory show` to verify the imported `AGENTS.md`. Restart after changing `.geminiignore`.
4. Confirm each agent identifies Go tests/race/build, Svelte check/lint/build, disposable migration validation, and deployment dry-run/failure checks. Test exclusions using synthetic files only; never probe with real secrets.

Migration checks are available below. Live deployment scripts, operational plans, and deployment checks are maintained in the private infrastructure repository; see [Release and deployment boundary](releases-and-deployment.md).

Native configuration references: [Codex instructions](https://learn.chatgpt.com/docs/agent-configuration/agents-md), [Claude imports](https://code.claude.com/docs/en/memory#agentsmd), [Claude permissions](https://code.claude.com/docs/en/permissions#read-and-edit), [Gemini imports](https://google-gemini.github.io/gemini-cli/docs/cli/gemini-md.html), and [Gemini ignore patterns](https://google-gemini.github.io/gemini-cli/docs/cli/gemini-ignore.html).

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

Existing databases created from the old `schema.sql` are not automatically adopted or erased. Migration 1 fails on their existing objects and records dirty state. Use a fresh database for this transition; do not use `force` to hide a failed migration. The one-time development database rebuild and Excel reimport are a separate operator step in the private deployment plan.

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

`down` is exercised only on this disposable database; live rollback remains forward-only. CI builds the backend and importer after migration validation succeeds. A master Excel reimport and live login/admin smoke test remain part of the separate rebuild procedure.

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

All four receive matching source-SHA and immutable release tags before GitHub Release creation. Pull requests run checks only. Private deployment handoff is separately gated and disabled by default; no host-side deployment runs in GitHub. See [release publication, retries, and operator setup](releases-and-deployment.md).
