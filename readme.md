# Texin roistot

Texin roistot is an alpha-stage catalog application for Tex Willer comics.
It stores and serves structured data about:

- villains (`Roistot`)
- stories (`Tarinat`)
- publications (Finnish and Italian release contexts)
- story creators (writers, drawers, translators)

The repository contains:

- Go backend (`texinroistot-server`)
- SvelteKit frontend (`texinroistot-ui`)
- Excel importer that creates versioned snapshots of the dataset

## Project status

- Alpha / unpublished.
- Frontend includes optional password gate (`/julkaisematon`) to block public access in pre-release environments.

## Documentation map

- [Documentation index](docs/README.md)
- [Functional overview](docs/functional-overview.md)
- [Technical architecture](docs/technical-architecture.md)
- [API reference](docs/api-reference.md)
- [Configuration reference](docs/configuration.md)
- [Data import and versioning](docs/data-import-and-versioning.md)
- [Development workflow](docs/development-workflow.md)

## Quick start (local development)

### 1. Prerequisites

- Docker + Docker Compose plugin
- Optional local tooling:
  - Go (for running backend directly)
  - Node.js + npm (for running frontend directly)

### 2. Configure backend env

Create `texinroistot-server/.env` using the [configuration reference](docs/configuration.md).

### 3. Start local stack

```bash
docker compose up
```

This starts:

- Postgres at `localhost:5432`
- Backend at `localhost:6969`
- Frontend dev server at `localhost:5173`

### 4. Apply database migrations

Compose applies migrations before starting the backend. To rebuild and run the migrator separately:

```bash
./scripts/init_schema.sh
```

### 5. Import data and activate newest version

```bash
./scripts/import_excel_and_activate_latest.sh
```

Both commands use Docker Compose services:

- schema init builds and runs the `migrate` service
- data import runs the dedicated `import` image/container

Importer reads:

- file: `texinroistot-server/Texinroistot.xlsx`
- sheet: `Taul1`

## CI and image publishing

### CI

Workflow: `.github/workflows/ci.yml`

- backend: `go test ./...` and `go build ./...`
- frontend: `npm ci`, `npm run check`, `npm run lint`, `npm run build`
- migrations: append-only history, disposable PostgreSQL checks, and schema snapshot comparison

### Image publishing (GHCR)

Workflow: `.github/workflows/images.yml`

After checks pass on `main`, CI publishes backend, frontend, importer, and migrator images with `sha-<full-commit-sha>` and `vX.Y.Z` tags, then creates a GitHub Release. Pull requests run checks only. There are no branch, `latest`, or semver-alias image tags published by this workflow.

See [releases and retries](docs/releases-and-deployment.md).

## Unpublished mode gate (frontend)

Set frontend runtime env vars:

- `UNPUBLISHED_MODE=true`
- `UNPUBLISHED_PASSWORD=<shared-password>`

When enabled:

- every non-static route is blocked until password is entered on `/julkaisematon`
- backend-proxy API routes under frontend return `401` without access cookie
- successful password submit sets an HTTP-only cookie and allows browsing
