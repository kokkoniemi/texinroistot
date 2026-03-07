# Backend Map

## Entry Points
- `cmd/server/server.go`: backend service startup.
- `cmd/importer/importer.go`: data import entrypoint.

## Core Packages
- `internal/auth/`: authentication handlers and service.
- `internal/stories/`: story-related handlers.
- `internal/admin/`: admin handlers.
- `internal/db/`: models, repositories, DB wiring, schema.
- `internal/config/config.go`: environment-backed configuration.
- `internal/importer/`: importer modules for authors, stories, villains, publications.

## Common Checks
- Unit tests and race tests: `go test ./...`, `go test -race ./...`
- Build verification: `go build ./...`
