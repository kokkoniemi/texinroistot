# Configuration reference

This document lists environment variables and runtime settings used by backend and frontend.

## Backend (`texinroistot-server`)

Read in `internal/config/config.go` via `godotenv` autoload.

### Required secrets

- `ROISTOT_SECRET`
- `ROISTOT_SALT`
- `ROISTOT_COOKIE_ACCESS_SECRET`
- `ROISTOT_COOKIE_REFRESH_SECRET`

### Auth and cookie behavior

- `ROISTOT_COOKIE_SECURE`
  - `true|false`
  - enables secure cookie behavior in backend auth cookie creation
- `GOOGLE_OAUTH2_CLIENT_ID`
  - audience for Google ID token validation in login flow
- `ROISTOT_ADMIN_EMAILS`
  - comma-separated admin email list (for example `admin@example.com,second@example.com`)
  - applied when users log in; matching users are marked as admin

### Database

- `DB_CONNECTION_STRING`
  - example:
  - `postgresql://tex:willer@db:5432/tex?sslmode=disable`

### Other backend vars

- `ROISTOT_LOGIN_EXPIRES_AFTER_MINUTES`
  - currently present in `.env` but not actively consumed in backend code paths.

## Frontend (`texinroistot-ui`)

### Backend proxy routing

- `BACKEND_HOST`
  - backend base URL for SvelteKit server-side proxy endpoints
  - fallback default: `http://backend:6969`

### Google login (frontend)

- `PUBLIC_GOOGLE_OAUTH2_CLIENT_ID`
  - Google OAuth2 client id used by `/hallinta` page Google Sign-In widget
  - should match backend `GOOGLE_OAUTH2_CLIENT_ID` audience

### Unpublished gate

- `UNPUBLISHED_MODE`
  - truthy values: `1`, `true`, `yes`, `on` (case-insensitive)
  - when enabled, users are redirected to `/julkaisematon` unless gate cookie exists
- `UNPUBLISHED_PASSWORD`
  - shared password accepted at unpublished gate page

### Typical node runtime

- `HOST`
  - server bind host (commonly `0.0.0.0`)
- `PORT`
  - server port (commonly `3000`)

### Local Docker Compose defaults

Development compose file: `docker-compose.yaml`.

Current defaults include:

- Postgres credentials:
  - user: `tex`
  - password: `willer`
  - db: `tex`
- Frontend unpublished gate enabled by default:
  - `UNPUBLISHED_MODE=true`
  - `UNPUBLISHED_PASSWORD=tex`

Adjust these values for your local workflow as needed.

## CI/CD settings

### GitHub Actions workflow permissions

Image publish workflow requires:

- job permission: `packages: write`
- repository Actions setting: `GITHUB_TOKEN` with read/write permissions

### GHCR image names

Derived in `.github/workflows/images.yml` as:

- `ghcr.io/<owner>/<repo>-backend`
- `ghcr.io/<owner>/<repo>-frontend`
- `ghcr.io/<owner>/<repo>-importer`
