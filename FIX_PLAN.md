# texinroistot-ui Refactor & Fix Plan

## Phase 1 — Critical Bugs & Accessibility

| # | Issue | File(s) | Status |
|---|-------|---------|--------|
| 1 | `lang="en"` on a Finnish app — WCAG Level A failure | `src/app.html` | ✅ Done |
| 2 | No `<title>` on any page — WCAG 2.4.2 Level A failure | all routes | ✅ Done |
| 3 | Race condition: popup villain fetch from story A can leak into story B if user switches stories fast | `tekijat/+page.svelte`, `roistot/+page.svelte` | ✅ Done |
| 4 | Front page "Hae hakusanalla" links to same URL as "Näytä kaikki" | `routes/+page.svelte` | ✅ Done |
| 5 | `publicationItem()` returns `year/issue`, but `formatBaseSeriesRange()` returns `issue/year` — inconsistent in same module | `lib/listing/shared.ts` | ✅ Done |
| 6 | `{#each}` blocks have no keys — stale DOM after navigation; eslint rule re-enabled | all three listing pages | ✅ Done |

## Phase 2 — UI Inconsistencies

| # | Issue | Detail | Status |
|---|-------|--------|--------|
| 7 | Popup "Sulje" button: **black bg/white text** on Tekijät, **white bg/dark text** on Roistot | CSS in both pages | ✅ Done |
| 8 | No CSS custom properties — error colors alone have 3 different hex values (`#8b0000`, `#8a0000`, `#8f0000`) | global styles / all pages | ✅ Done |
| 9 | "Rooli" row in villain card always renders even when empty ("Rooli: -"), while other rows are conditional | `roistot/+page.svelte:406` | ✅ Done |
| 10 | "Danger" button visually identical to normal button — no red/destructive styling | `hallinta/+page.svelte` | ✅ Done |
| 11 | Front page section links are `<p>` tags, should be `<h2>` | `routes/+page.svelte` | ✅ Done |

## Phase 3 — Shared Components & Deduplication ✅ Done

Highest-leverage structural work. The app has zero shared components — nearly everything is copy-pasted across `tekijat`, `roistot`, and `tarinat`.

### Components extracted into `src/lib/components/`

| Component | What gets deduplicated |
|-----------|------------------------|
| `StoryPopup.svelte` | Entire popup HTML + backdrop + close logic (2 copies → 1) |
| `Pagination.svelte` | Top+bottom pagination nav (6 total copies across 3 pages → 1) |
| `VillainCard.svelte` | Villain card inside popups/expanded rows (3 copies → 1) |
| `FilterForm.svelte` | Submit button, reset link, loading state, result-row/total/page (3 copies) |

### Functions moved into `$lib/listing/shared.ts`

| Function | Previously duplicated in |
|----------|------------------------|
| `italianOriginalPublication()` | All 3 listing pages (identical) |
| `storyVillainTitle()` (was `villainTitle`/`popupVillainTitle`) | Same function, different names across pages |
| `storyVillainForStory()` | All 3 listing pages (identical) |

## Phase 4 — TypeScript & Type Safety

| # | Issue | Status |
|---|-------|--------|
| 12 | `Villain`, `StoryVillain`, `StoryVillainsResponse` types declared locally in each page — move to `$lib/types.ts` | ⬜ Todo |
| 13 | `StoryVillain.hash` and `Villain.hash` are `string` in tekijat/tarinat but `string \| undefined` in roistot — pick one truth | ⬜ Todo |
| 14 | `+page.ts` load functions return `any` — define typed response shapes | ⬜ Todo |
| 15 | `AdminUser`/`AdminVersion` types duplicated between `+page.server.ts` and `+page.svelte` in hallinta | ⬜ Todo |
| 16 | `pageHref()` in tekijat hardcodes `sort: 'last_name'`, ignoring `filters.sort` | ⬜ Todo |

## Phase 5 — Accessibility Polish

| # | Issue | Status |
|---|-------|--------|
| 17 | Modal dialog doesn't move focus on open — screen readers must navigate in manually | ⬜ Todo |
| 18 | Toggle buttons missing `aria-expanded` (villain expand, author expand) | ⬜ Todo |
| 19 | "Palauta oletukset" reset link uses `aria-disabled` without `tabindex="-1"` — keyboard users can still activate it | ⬜ Todo |

## Phase 6 — Svelte 5 Migration (largest effort, do last)

The app installs Svelte 5 (`^5.53.7`) but runs entirely in legacy Svelte 4 mode. Migration involves:
- Replace `$:` reactive statements with `$derived` / `$effect`
- Replace `let` mutable state with `$state`
- Replace `on:event` handlers with `onevent` attributes
- Replace `export let data` props with `$props()`
- Enable runes mode per component

Do this after Phases 1–5 are complete and tests are passing.
