# Codex Configuration

## Project Overview
- Monorepo for a comic database.
- Backend service is in `texinroistot-server` (referred to by the team as `texinroistot-backend`).
- Frontend service is in `texinroistot-ui` (SvelteKit).
- Local stack is started with `docker-compose.yaml` at repo root.

## Working Rules
- Keep changes minimal and scoped to the requested task.
- Do not edit secrets or local environment files: `**/.env`, `**/.env.*`, except `*.env-example`.
- Never run destructive git commands (`git reset --hard`, `git checkout --`, force-push) unless explicitly requested.
- Prefer `rg` for search and `apply_patch` for small targeted edits.
- Preserve existing coding style and naming in touched files.

## Backend (Go) Defaults
- Working directory: `texinroistot-server`.
- Preferred validation order for backend-only changes:
  1. `go test ./...`
  2. `go test -race ./...` (when tests exist and runtime is acceptable)
  3. `go build ./...`
- If database schema or importer logic changes, also check:
  - `internal/db/schema.sql`
  - `cmd/importer/importer.go`

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

## Skills
- `texinroistot-go-backend`
  - Path: `.codex/skills/texinroistot-go-backend/SKILL.md`
  - Use for backend Go handlers/services/repositories/schema/importer work in `texinroistot-server`.
- `texinroistot-sveltekit-frontend`
  - Path: `.codex/skills/texinroistot-sveltekit-frontend/SKILL.md`
  - Use for SvelteKit pages/routes/API wiring/build-lint-check work in `texinroistot-ui`.
- For backend/frontend API contract changes, use both skills together.
