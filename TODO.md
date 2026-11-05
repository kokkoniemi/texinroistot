TODO: Aakkosten mukainen sivutus

TODO: Svelte 5 Migration
The app installs Svelte 5 (`^5.53.7`) but runs entirely in legacy Svelte 4 mode. Migration involves:
- Replace `$:` reactive statements with `$derived` / `$effect`
- Replace `let` mutable state with `$state`
- Replace `on:event` handlers with `onevent` attributes
- Replace `export let data` props with `$props()`
- Enable runes mode per component
