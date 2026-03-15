# Functional overview

## Purpose

Texin roistot is a searchable, filterable reference for data from Tex Willer comics.

The core product goal is to answer questions such as:

- Which villains appear in a given story?
- In which publications has a story appeared?
- Which stories match a name, role, title, or publication clue?
- Which writers, drawers, and translators are associated with a story?

## Core concepts (domain)

- `Roisto` (villain): a canonical villain identity with one or more story appearances.
- `Tarina` (story): a story entity that may appear in many publication contexts.
- `Julkaisu` (publication): a release context, for example Finnish perussarja, Italian perussarja, Maxi-Tex, etc.
- `Tekijat` (authors): writers, drawers, and translators attached to stories.
- `Versio` (version): an import snapshot; all rows are tied to one version, and one version is active at a time.

## User-facing pages

### Front page (`/`)

- Shows high-level dataset counts (villains, stories, drawers, writers) from active version.
- Provides quick links to common list filters in both villain and story tabs.

### Villains (`/roistot`)

- Backend-driven list with URL-stateful filters.
- Filters:
  - publication scope (`all`, `fi`, `it`)
  - sort options (first name, last name, nickname, rank, FI pub date, IT pub date)
  - keyword search (`q`)
- Pagination:
  - page and page size
  - first/last and truncated middle page links
- Card content:
  - villain title (name + nickname handling)
  - rank, role, destiny
  - linked story and publication summary
  - story creators (writer, drawer, translator)

### Stories (`/tarinat`)

- Backend-driven list with URL-stateful filters.
- Filters:
  - publication category (`all`, `perus_fi`, `perus_it`, `suur`, `maxi`, `kirjasto`, `kronikka`, `special`)
  - sort (`alpha`, `fi_pub_date`, `it_pub_date`)
  - keyword search (`q`)
- Pagination:
  - page and page size
  - first/last and truncated middle page links
- Card content:
  - title (with Italian title in parentheses when available)
  - creators (writer, drawer, translator)
  - grouped publication summary
- Expand behavior:
  - "Nayta tarinan roistot" toggles villains list for the selected story
  - data fetched from backend by story hash and cached per visible result set

### Hallinta (`/hallinta`)

- Uses Google Sign-In for authentication.
- Logged-out users see Google login.
- Logged-in non-admin users see message: `Sinulla ei ole oikeuksia hallintaan` and can delete their account.
- Logged-in admin users see placeholder admin message, user list, and admin-right granting action.

## Unpublished access gate

The UI supports a global unpublished mode:

- route: `/julkaisematon`
- controlled by frontend env vars:
  - `UNPUBLISHED_MODE`
  - `UNPUBLISHED_PASSWORD`

If unpublished mode is enabled and visitor has not passed the gate:

- all pages redirect to `/julkaisematon`
- frontend API proxy endpoints return `401`

This is intended as lightweight pre-release access control, not full authentication/authorization.

## Data freshness model

- Data is refreshed by importing a new Excel snapshot.
- Import creates a new `version`.
- One version is marked active.
- All browsing and counts read from active version.
