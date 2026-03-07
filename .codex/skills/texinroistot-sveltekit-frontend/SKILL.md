---
name: texinroistot-sveltekit-frontend
description: Implement and maintain the SvelteKit frontend in texinroistot-ui. Use when work touches Svelte pages/layouts, +page.ts loaders, +server.ts API routes, frontend build/lint/check flows, or backend-frontend API contract wiring.
---

# Texinroistot Sveltekit Frontend

Apply this workflow for frontend work in `texinroistot-ui`.

## Workflow
1. Identify affected routes/components and related API calls before editing.
2. Keep edits minimal and consistent with existing SvelteKit patterns.
3. Validate frontend changes in order:
   - `cd texinroistot-ui`
   - `npm run check`
   - `npm run lint`
   - `npm run build`
4. Report changed files, command results, and skipped checks.

## Guardrails
- Keep route structure and naming conventions (`+layout.svelte`, `+page.svelte`, `+page.ts`, `+server.ts`) consistent.
- For API changes, update frontend fetch usage and backend handlers together.
- Avoid adding dependencies unless requested or clearly necessary.
- Preserve existing formatting and lint expectations.

## References
- Use [references/frontend-map.md](references/frontend-map.md) for key UI paths and route map.
