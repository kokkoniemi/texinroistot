# Development workflow

## Local stack (Docker Compose)

```bash
docker compose up
```

Services:

- database: Postgres `localhost:5432`
- backend API: `localhost:6969`
- frontend dev server: `localhost:5173`

## Schema bootstrap

```bash
./scripts/init_schema.sh
```

## Import latest spreadsheet

```bash
./scripts/import_excel_and_activate_latest.sh
```

Both scripts use Docker Compose services:

- schema init runs `psql` in `db`
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

GitHub workflow `.github/workflows/images.yml` builds and pushes three GHCR images:

- backend
- frontend
- importer

No infrastructure deployment is performed in this repository.
