# Frontend Map

## Core Files
- `src/routes/+layout.svelte`: app-level layout shell.
- `src/routes/+page.svelte`: root page.
- `src/app.css`: global styles.

## Route Areas
- `src/routes/tarinat/`: story pages and loaders.
- `src/routes/roistot/`: villain pages.
- `src/routes/hallinta/`: admin-related page.
- `src/routes/api/tarinat/+server.ts`: SvelteKit endpoint for stories.
- `src/routes/api/roistot/+server.ts`: SvelteKit endpoint for villains.

## Common Checks
- Type and Svelte checks: `npm run check`
- Formatting/lint checks: `npm run lint`
- Build verification: `npm run build`
