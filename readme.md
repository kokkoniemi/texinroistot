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

Create `/texinroistot-server/.env` (example values are shown in `configuration.md`).

### 3. Start local stack

```bash
docker compose up
```

This starts:

- Postgres at `localhost:5432`
- Backend at `localhost:6969`
- Frontend dev server at `localhost:5173`

### 4. Initialize schema (first time)

```bash
psql -h localhost -p 5432 -d tex -U tex < texinroistot-server/internal/db/schema.sql
```

### 5. Import data and activate newest version

```bash
./scripts/import_excel_and_activate_latest.sh
```

Importer reads:

- file: `texinroistot-server/Texinroistot.xlsx`
- sheet: `Taul1`

## CI and image publishing

### CI

Workflow: `.github/workflows/ci.yml`

- backend: `go test ./...` and `go build ./...`
- frontend: `npm ci`, `npm run check`, `npm run lint`, `npm run build`

### Production image publishing (GHCR)

Workflow: `.github/workflows/images.yml`

This repository publishes production-ready images only.
Hosting infrastructure (Compose/Kubernetes/reverse proxy/Terraform) should live in a separate repository.

Images:

- `ghcr.io/<owner>/<repo>-backend`
- `ghcr.io/<owner>/<repo>-frontend`
- `ghcr.io/<owner>/<repo>-importer`

Tags include:

- `latest` (default branch)
- git branch/tag refs
- semver aliases for release tags (`1`, `1.2`, `1.2.3`)
- commit SHA

Required repository setting:

- GitHub Actions `GITHUB_TOKEN` must have package write permissions (`packages: write`).

## Unpublished mode gate (frontend)

Set frontend runtime env vars:

- `UNPUBLISHED_MODE=true`
- `UNPUBLISHED_PASSWORD=<shared-password>`

When enabled:

- every non-static route is blocked until password is entered on `/julkaisematon`
- backend-proxy API routes under frontend return `401` without access cookie
- successful password submit sets an HTTP-only cookie and allows browsing
