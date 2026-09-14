# Backend Map

## Entry Points
- `cmd/server/server.go`: backend service startup.
- `cmd/importer/importer.go`: data import entrypoint.

## Core Packages
- `internal/auth/`: authentication handlers and service.
- `internal/stories/`: story-related handlers.
- `internal/admin/`: admin handlers.
- `internal/db/`: models, repositories, DB wiring, schema snapshot, and versioned SQL migrations.
- `internal/config/config.go`: environment-backed configuration.
- `internal/importer/`: importer modules for authors, stories, villains, publications.

## Common Checks
- Unit tests and race tests: `go test ./...`, `go test -race ./...`
- Build verification: `go build ./...`
- Migration verification from repo root: `python3 scripts/check_migration_history.py --base-ref <base-commit>`, `bash scripts/check_migrations.sh`
