---
name: texinroistot-go-backend
description: Implement and maintain the Go backend in texinroistot-server. Use when work touches Go handlers, services, auth, importer, database repositories/schema, backend API behavior, or backend test/build workflows.
---

# Texinroistot Go Backend

Apply this workflow for backend work in `texinroistot-server`.

## Workflow
1. Identify affected packages and files before editing.
2. Read nearby handler/service/repository code and preserve existing style and naming.
3. Make minimal edits; avoid unrelated refactors.
   - For persistence changes, follow the database migration rules in the root `AGENTS.md`. `internal/db/migrations/` is the schema source of truth; `schema.sql` is the transitional snapshot.
   - From the repository root, run `python3 scripts/check_migration_history.py --base-ref <base-commit>` and `bash scripts/check_migrations.sh`. The latter creates disposable PostgreSQL 17 containers for up/down/up, required-object and snapshot checks; never substitute a production database. Rootless Podman is supported with `CONTAINER_ENGINE=podman`.
4. Run backend validations in order:
   - `cd texinroistot-server`
   - `go test ./...`
   - `go test -race ./...` (skip only if clearly too slow)
   - `go build ./...`
5. Report changed files, command results, and any skipped checks.

## Guardrails
- Keep API response shapes and field names stable unless explicitly asked to break compatibility.
- Review migrations, the schema snapshot, models, repositories, and importer behavior together when modifying persistence behavior.
- Do not edit `.env` or secrets; use existing config loading patterns.
- For schema/importer changes, call out manual follow-up steps for:
  - `internal/db/migrations/`
  - `internal/db/schema.sql`
  - `cmd/importer/importer.go`

## References
- Use [references/backend-map.md](references/backend-map.md) for key backend paths and fast entry points.
