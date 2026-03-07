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
4. Run backend validations in order:
   - `cd texinroistot-server`
   - `go test ./...`
   - `go test -race ./...` (skip only if clearly too slow)
   - `go build ./...`
5. Report changed files, command results, and any skipped checks.

## Guardrails
- Keep API response shapes and field names stable unless explicitly asked to break compatibility.
- Update SQL and repository code together when modifying persistence behavior.
- Do not edit `.env` or secrets; use existing config loading patterns.
- For schema/importer changes, call out manual follow-up steps for:
  - `internal/db/schema.sql`
  - `cmd/importer/importer.go`

## References
- Use [references/backend-map.md](references/backend-map.md) for key backend paths and fast entry points.
