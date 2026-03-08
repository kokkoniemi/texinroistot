# Project

* This is a comic catalog project for the villains in Tex Willer series.

## Startup

Copy .env-example to .env and populate values

```
docker-compose up
psql -h localhost -p 5432 -d tex -U tex < texinroistot-server/internal/db/schema.sql
```

Helper script for import + activate newest version:

```
./scripts/import_excel_and_activate_latest.sh
```

## Production Images (GHCR)

This repository only builds and publishes production images.
Hosting infrastructure (Compose/Kubernetes/Caddy/Traefik/Terraform) should live in a separate repo.

### Published images

On push to `main`, tag push (`v*`), or manual dispatch, workflow `.github/workflows/images.yml` publishes:
- `ghcr.io/<owner>/<repo>-backend`
- `ghcr.io/<owner>/<repo>-frontend`
- `ghcr.io/<owner>/<repo>-importer`

Tag strategy:
- `latest` from default branch
- branch/tag refs
- commit SHA tags

### Requirements

- The workflow uses `GITHUB_TOKEN` and needs repository permission `packages: write`.
- For private images, your infra repo deployment runner must authenticate to GHCR.

### Example pull

```
docker pull ghcr.io/<owner>/<repo>-backend:latest
docker pull ghcr.io/<owner>/<repo>-frontend:latest
docker pull ghcr.io/<owner>/<repo>-importer:latest
```
