TODO: Paholaiskolo näyttää väärin julkaisutiedot
TODO: Kursiivilla kirjoitetut osuudet Excelin tekstissä tulisi näyttää kursiivilla käyttöliittymässä
TODO: tarkista filtterit. Suomen perussarjassa ei näytä esim. kuolma tulee taivaalta
TODO: Salanimi Roistot-sivun Järjestys-filtteriin.
TODO: Aakkosten mukainen sivutus
TODO: Migraatiot. Miten tietokantamuutokset ajetaan sisään helpommin?
TODO: Ohjelmaversio ei toimi. Sen voisi ottaa IMAGE_TAG-muuttujasta ja paikallisesti syöttää aikaleiman

TODO: Svelte 5 Migration
The app installs Svelte 5 (`^5.53.7`) but runs entirely in legacy Svelte 4 mode. Migration involves:
- Replace `$:` reactive statements with `$derived` / `$effect`
- Replace `let` mutable state with `$state`
- Replace `on:event` handlers with `onevent` attributes
- Replace `export let data` props with `$props()`
- Enable runes mode per component

Do this after Phases 1–5 are complete and tests are passing.
