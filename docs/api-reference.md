# API reference

This document covers the backend HTTP API exposed by `texinroistot-server` and the frontend proxy API exposed by `texinroistot-ui`.

## Backend base

- Local default: `http://localhost:6969`
- Prefix: `/api`

## Health

- `GET /healthz`
  - Returns `200` when server process is alive.

## Public data endpoints

### `GET /api/version/active`

Returns active version and aggregate stats used by front page/footer.

Response shape:

```json
{
  "version": {
    "id": 123,
    "createdAt": "2026-03-08T10:00:00Z",
    "isActive": true
  },
  "stats": {
    "villains": 5800,
    "stories": 4100,
    "drawers": 120,
    "writers": 95,
    "translators": 60
  }
}
```

### `GET /api/stories`

Story list with filtering, search, sort, and page-based pagination.

Query params:

- `publication`: `all|perus_fi|perus_it|suur|maxi|kirjasto|kronikka|special`
  - default: `perus_fi`
- `sort`: `alpha|fi_pub_date|it_pub_date`
  - default: `fi_pub_date`
- `q`: free text search
  - default: empty
- `page`: positive integer
  - default: `1`
- `pageSize`: positive integer, max `100`
  - default: `25`

Response shape:

```json
{
  "stories": [/* Story[] */],
  "meta": {
    "total": 4100,
    "page": 1,
    "pageSize": 25,
    "totalPages": 164
  },
  "filters": {
    "publication": "perus_fi",
    "sort": "fi_pub_date",
    "q": ""
  }
}
```

### `GET /api/stories/:storyHash/villains`

Lists villains for a single story hash in active version.

Path params:

- `storyHash`: required

Response shape:

```json
{
  "storyHash": "....",
  "villains": [/* Villain[] */],
  "meta": {
    "total": 4
  }
}
```

Errors:

- `400` if `storyHash` missing/empty
- `404` if story does not exist in active version

### `GET /api/villains`

Villain list with filtering, search, sort, and page-based pagination.

Query params:

- `publication`: `all|fi|it`
  - default: `fi`
- `sort`: `first_name|last_name|nickname|rank|fi_pub_date|it_pub_date`
  - default: `fi_pub_date`
- `q`: free text search
  - default: empty
- `page`: positive integer
  - default: `1`
- `pageSize`: positive integer, max `100`
  - default: `25`

Response shape:

```json
{
  "villains": [/* Villain[] */],
  "meta": {
    "total": 5800,
    "page": 1,
    "pageSize": 25,
    "totalPages": 232
  },
  "filters": {
    "publication": "fi",
    "sort": "fi_pub_date",
    "q": ""
  }
}
```

## Auth-related endpoints

### `POST /api/login`

- Expects form-urlencoded payload with Google credential token and CSRF token.
- Sets auth cookies on success.
- Redirects to `/hallinta` on success.

### `POST /api/logout`

- Clears auth cookies.
- Returns:

```json
{ "loggedOut": true }
```

### `GET /api/me`

- Returns logged-in status, user email, and admin flag (if authenticated).
- Returns `{ "loggedIn": false, "email": "" }` when access token is missing, invalid, or expired.

### `DELETE /api/me`

- Deletes currently logged-in user account.
- Clears auth cookies.

### `GET /api/admin/users`

- Protected by backend middleware (`auth.ProtectedRoute`).
- Returns users list.

### `POST /api/admin/users/grant-admin`

- Protected by backend middleware (`auth.ProtectedRoute`).
- Expects JSON body:

```json
{ "email": "user@example.com" }
```

- Grants admin rights for an existing logged-in user matching the email hash.

### `GET /api/admin/versions`

- Protected by backend middleware (`auth.ProtectedRoute`).
- Returns all versions sorted by creation time.

### `POST /api/admin/versions/:versionID/activate`

- Protected by backend middleware (`auth.ProtectedRoute`).
- Sets the given version as active.
- Returns `404` if version does not exist.

### `DELETE /api/admin/versions/:versionID`

- Protected by backend middleware (`auth.ProtectedRoute`).
- Deletes a non-active version.
- Returns `409` when trying to delete active version.
- Returns `404` if version does not exist.

### `POST /api/admin/versions/import`

- Protected by backend middleware (`auth.ProtectedRoute`).
- Downloads spreadsheet from configured import URL and creates a new inactive version.
- Returns `409` if another import is already running.
- Returns `400` if source URL is invalid, unavailable, or does not resolve to a valid `.xlsx` payload.

## Error behavior

Common patterns:

- `400` for invalid query/path values
- `401` for protected or gated access failures
- `404` for missing resources
- `500` for server/database failures

Error body is usually:

```json
{ "error": "..." }
```

## Frontend proxy API

SvelteKit exposes same-shape proxy routes under frontend origin:

- `POST /api/login`
- `POST /api/logout`
- `GET /api/me`
- `DELETE /api/me`
- `GET /api/admin/users`
- `POST /api/admin/users/grant-admin`
- `GET /api/admin/versions`
- `POST /api/admin/versions/import`
- `POST /api/admin/versions/:versionID/activate`
- `DELETE /api/admin/versions/:versionID`
- `GET /api/version/active`
- `GET /api/tarinat`
- `GET /api/tarinat/:storyHash/roistot`
- `GET /api/roistot`

These routes forward requests to backend host from `BACKEND_HOST` runtime env (fallback `http://backend:6969`).
