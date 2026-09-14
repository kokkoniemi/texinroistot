# Repository Agent Instructions

`AGENTS.md` is the canonical instruction source for all coding agents. Keep vendor adapters as imports rather than copies of these rules.

## Project Overview
- Monorepo for a comic database.
- Backend service is in `texinroistot-server` (referred to by the team as `texinroistot-backend`).
- Frontend service is in `texinroistot-ui` (SvelteKit).
- Local stack is started with `docker-compose.yaml` at repo root.

## Working Rules
- Keep changes minimal and scoped to the requested task.
- Never read, print, edit, or commit secrets, runtime env files (including `backend.env`, `frontend.env`, and `stack.env`), VPN configs, SSH private keys, tokens, database dumps, or production data. Committed templates such as `*.example` and `*.env-example` are allowed only when they contain placeholders.
- Never run destructive git commands (`git reset --hard`, `git checkout --`, force-push) unless explicitly requested.
- Prefer `rg` for search and `apply_patch` for small targeted edits.
- Preserve existing coding style and naming in touched files.
- Preserve unrelated worktree changes and generated/data files.

## Backend (Go) Defaults
- Working directory: `texinroistot-server`.
- Preferred validation order for backend-only changes:
  1. `go test ./...`
  2. `go test -race ./...` (when tests exist and runtime is acceptable)
  3. `go build ./...`
- If database schema or importer logic changes, also check:
  - `internal/db/migrations/`
  - `internal/db/schema.sql`
  - `cmd/importer/importer.go`

## Database Migrations
- SQL migrations in `texinroistot-server/internal/db/migrations/` are the schema source of truth; keep `internal/db/schema.sql` as a reviewed snapshot during transition.
- Migration files are append-only after merge. Never edit, delete, or renumber a merged/applied migration; add a new migration instead.
- Every schema change includes forward SQL and reverse SQL where safe. Document irreversible changes and their recovery procedure.
- Check repositories, models, importer behavior, and the schema snapshot whenever persistence changes.
- From the repository root, run `python3 scripts/check_migration_history.py --base-ref <base-commit>` and `bash scripts/check_migrations.sh` to verify append-only history, PostgreSQL 17 up/down/up, required objects, and snapshot parity. Use only the disposable database created by the check script.
- Use expand/contract migrations so the previous application release remains compatible with the migrated database. Never automatically run down migrations in production.

## Frontend (SvelteKit) Defaults
- Working directory: `texinroistot-ui`.
- Install dependencies with `npm ci` (not `npm install`) when needed.
- Preferred validation order for frontend-only changes:
  1. `npm run check`
  2. `npm run lint`
  3. `npm run build`

## Full-Stack Change Checklist
- For API contract changes, update both:
  - Go handlers/services in `texinroistot-server/internal/...`
  - SvelteKit API routes and UI usage in `texinroistot-ui/src/routes/...`
- Keep backend endpoints and frontend fetch paths aligned.
- Mention any skipped checks in the final report.

## Safe Execution Defaults
- Use read-only inspection first, edit second, run validations last.
- Ask before network-heavy or long-running commands if they are not strictly required.
- Do not add new dependencies unless necessary for the task.

## Release and Deployment
- Never run production deployments, OpenTofu, or infrastructure apply commands without explicit operator authorization. A plan or code change does not authorize production execution, secret access, or WireGuard changes.
- Release workflow changes require least-privilege permissions, immutable release image tags, concurrency control, and a documented rollback path.
- Deployment scripts must support a local dry run and fail closed on preflight, backup, migration, or health-check errors. Validate dry-run behavior and failure paths on disposable targets before production use.
- Backup and migration must succeed before replacing application containers. On failed post-deploy health checks, restore the previous application image tag while keeping the database migrated forward.

## Skills
- `texinroistot-go-backend`
  - Path: `.codex/skills/texinroistot-go-backend/SKILL.md`
  - Use for backend Go handlers/services/repositories/schema/importer work in `texinroistot-server`.
- `texinroistot-sveltekit-frontend`
  - Path: `.codex/skills/texinroistot-sveltekit-frontend/SKILL.md`
  - Use for SvelteKit pages/routes/API wiring/build-lint-check work in `texinroistot-ui`.
- For backend/frontend API contract changes, use both skills together.
