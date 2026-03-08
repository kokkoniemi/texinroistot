# Technical architecture

## Monorepo layout

- `texinroistot-server/`
  - Go Fiber HTTP API
  - Postgres data access layer
  - Excel importer
- `texinroistot-ui/`
  - SvelteKit application
  - Server-side API proxy routes to backend
- `scripts/`
  - local helper scripts (for example import + active version switch)
- `.github/workflows/`
  - CI and GHCR image publishing

## Runtime components

### Backend (Go + Fiber)

- Entry point: `texinroistot-server/cmd/server/server.go`
- Listens on `:6969`
- Routes are under `/api`
- Data access through repository layer in `internal/db`

Main backend packages:

- `internal/stories`: story listing and story->villain listing handlers
- `internal/villains`: villain listing handler
- `internal/versions`: active version + stats endpoint
- `internal/auth`: login/logout/me and protected route helper
- `internal/admin`: admin-only handlers
- `internal/importer`: spreadsheet parsing and persistence logic

### Frontend (SvelteKit)

- Main pages:
  - `/`
  - `/roistot`
  - `/tarinat`
  - `/hallinta`
  - `/julkaisematon`
- Proxy endpoints:
  - `/api/roistot` -> backend `/api/villains`
  - `/api/tarinat` -> backend `/api/stories`
  - `/api/tarinat/[storyHash]/roistot` -> backend `/api/stories/:storyHash/villains`
  - `/api/version/active` -> backend `/api/version/active`

Backend base URL for proxy is resolved at runtime via `BACKEND_HOST` (`$env/dynamic/private`) with fallback `http://backend:6969`.

## Database

- Engine: PostgreSQL
- Schema file: `texinroistot-server/internal/db/schema.sql`
- Key tables:
  - `versions`
  - `villains`, `villains_in_stories`
  - `stories`, `stories_in_publications`
  - `publications`
  - `authors`, `authors_in_stories`
  - `users`

All content entities are versioned via `version` foreign keys.

## Request flow (read path)

1. Browser requests page route (`/tarinat`, `/roistot`, etc.).
2. SvelteKit load fetches corresponding internal `/api/...` route.
3. SvelteKit server endpoint proxies request to backend service.
4. Backend resolves active version, validates query params, queries Postgres, returns JSON.
5. Frontend renders data, filters, and pagination from response.

## Import flow (write path)

1. Importer reads `Texinroistot.xlsx` (`Taul1`).
2. Rows are normalized into in-memory entities:
  - authors
  - publications
  - stories (+ story-publication links + story-author links)
  - villains (+ villain-story appearance links)
3. Importer creates a new inactive version.
4. Data is bulk inserted in dependency-safe order.
5. Activation script marks newest version as only active version.

## Deployment boundary

This repository is an image producer:

- backend image
- frontend image
- importer image

Infrastructure composition and rollout is expected in a separate operations repository.

## Current constraints and notes

- Active version assumes exactly one `versions.is_active = true`.
- Authentication subsystem exists but admin UI is minimal (`/hallinta` is a placeholder page).
- Unpublished password gate is intentionally lightweight; it is not a substitute for robust authz.
