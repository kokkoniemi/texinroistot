# Data import and versioning

## Overview

Dataset updates are driven by Excel imports.
Each import creates a new database `version` snapshot.
Application reads only from the active version.

Importer entrypoint:

- `texinroistot-server/cmd/importer/importer.go`

Input file assumptions:

- filename: `Texinroistot.xlsx`
- sheet name: `Taul1`
- first row contains column titles

## Version model

All content entities point to a `versions.id`:

- `authors.version`
- `stories.version`
- `publications.version`
- `villains.version`

Only one row in `versions` should be active (`is_active = true`) at a time.

The helper script `scripts/import_excel_and_activate_latest.sh`:

1. runs importer in the dedicated import image/container
2. sets newest version as active
3. prints current active version

## Import pipeline

High-level sequence in importer:

1. parse spreadsheet rows
2. build in-memory entities and cross-references
3. create new inactive version
4. persist entities in dependency order:
   - authors
   - publications
   - stories (+ authors_in_stories + stories_in_publications)
   - villains (+ villains_in_stories)

Bulk insert helpers use Postgres `COPY` to reduce insert overhead.

## Required spreadsheet columns

Importer matches exact header names and maps them to internal keys.
Current required logical keys include:

- villain identity:
  - rank, first names, last name
  - nicknames, other names, code names
  - role, destiny
  - external villain id
- story and creators:
  - story title
  - writers
  - drawers
  - translators
  - story order number
- publication dimensions:
  - Finnish base publication year/from/to
  - Finnish re-publication year/from/to
  - Finnish special publication, kronikka, kirjasto
  - Italian base year/from/to
  - Italian special publication
  - Italian story title

If required columns are missing, importer returns an error listing missing keys.

## Translator parsing details

Translator field (`Suomensi`) supports semicolon-separated names and optional detail notes.

Handled forms include:

- `Surname, Firstname`
- `Surname, Firstname (2. - 3. p)`
- `Surname, Firstname 2. - 3. p`
- parenthesized segments in source rows

Importer stores translator detail text in `authors_in_stories.details`.
UI then renders translators in format like:

- `Firstname Lastname (2. - 3. p)`

## Publication mapping

Publication records are normalized into `publications` with enum `publication_type`:

- `perus`
- `maxi`
- `suur`
- `muu_erikois`
- `kronikka`
- `kirjasto`
- `italia_perus`
- `italia_erikois`

Story-publication titles are stored in `stories_in_publications.title`.

## Import execution options

### Via docker-compose helpers (recommended locally)

Initialize schema (first time):

```bash
./scripts/init_schema.sh
```

```bash
./scripts/import_excel_and_activate_latest.sh
```

These scripts use:

- `docker compose exec -T db psql -U tex -d tex ... < schema.sql`
- `docker compose --profile tools run --rm -T import`
- `docker compose exec -T db psql -U tex -d tex -c "UPDATE versions ..."`

### Direct importer run (inside backend context)

```bash
go run cmd/importer/importer.go
```

## Common failure modes

- missing/renamed Excel column titles
- wrong sheet name
- malformed numeric fields (year/issue/order)
- no active version after manual DB operations

When troubleshooting:

1. verify `versions` table has exactly one active row
2. verify expected tables have rows for active version id
3. re-run importer and activation script in sequence
