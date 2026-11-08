<script lang="ts">
	import type { PageData } from './$types';

	export let data: PageData;

	function formatCount(value: number | null | undefined): string {
		if (value === null || value === undefined) return '...';
		return new Intl.NumberFormat('fi-FI').format(value);
	}
</script>

<svelte:head>
	<title>Texin roistot-tietokanta</title>
</svelte:head>

<div class="front-poster">
	<!-- The paper: a clipped sheet (torn corners) frayed by the #ragged-edge filter. -->
	<div class="poster-paper" aria-hidden="true"><div class="poster-sheet"></div></div>

	<section class="page-front">
		<img src="/tex-logo.jpg" class="front-logo" alt="Kuva Tex Willeristä" />
		<h1>Texin roistot-tietokanta</h1>

		<p>
			Texin roistot-tietokanta sisältää tiedot
			{formatCount(data.activeVersionStats?.villains)}:sta tex willer -sarjakuvassa esiintyvästä
			roistosta, {formatCount(data.activeVersionStats?.stories)}:sta tarinasta,
			{formatCount(data.activeVersionStats?.drawers)}:sta piirtäjästä, sekä
			{formatCount(data.activeVersionStats?.writers)}:sta kirjoittajasta.
		</p>
	</section>
</div>

<style>
	/* A notice on the page, like a poster on a saloon wall. */
	.front-poster {
		position: relative;
		isolation: isolate;
		max-width: 40rem;
		margin: 2.5rem auto 0;
		/* the shadow follows the torn, frayed outline of the paper */
		filter: drop-shadow(0 5px 8px var(--shadow));
	}

	.poster-paper {
		position: absolute;
		inset: 0;
		z-index: -1;
		filter: url(#ragged-edge);
	}

	/* Rough, gritty paper with browned edges; a chunk torn off two corners. */
	.poster-sheet {
		position: absolute;
		inset: 0;
		background:
			url('/paper-grit.svg') 0 0 / 420px 420px,
			var(--card-bg);
		box-shadow:
			inset 0 0 40px rgba(120, 85, 35, 0.14),
			inset 0 0 6px rgba(100, 70, 30, 0.25);
		clip-path: polygon(
			0 0,
			calc(100% - 70px) 0,
			calc(100% - 52px) 10px,
			calc(100% - 38px) 7px,
			calc(100% - 24px) 22px,
			calc(100% - 10px) 26px,
			100% 48px,
			100% 100%,
			58px 100%,
			44px calc(100% - 12px),
			26px calc(100% - 8px),
			12px calc(100% - 22px),
			0 calc(100% - 30px)
		);
	}

	.page-front {
		padding: 2rem 2rem 1.8rem;
		text-align: center;
	}

	.front-logo {
		width: 150px;
		height: 150px;
		border-radius: 50%;
		border: 3px solid var(--gold);
		box-shadow: 0 0 0 4px var(--leather);
		object-fit: cover;
		/* blend the white drawing paper into the notice's paper */
		filter: sepia(0.35);
	}

	h1 {
		margin: 1.2rem 0 0.8rem;
		font-size: 2.1rem;
	}

	p {
		margin: 0 auto;
		max-width: 32rem;
		font-size: 1.1rem;
	}

	@media (max-width: 640px) {
		.front-poster {
			margin: 1.25rem 0.75rem 0;
		}

		.page-front {
			padding: 1.5rem 1rem 1.4rem;
		}

		h1 {
			font-size: 1.6rem;
		}
	}
</style>
