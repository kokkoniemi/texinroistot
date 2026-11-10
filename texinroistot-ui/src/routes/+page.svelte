<script lang="ts">
	import PaperNotice from '$lib/components/PaperNotice.svelte';
	import type { PageData } from './$types';

	export let data: PageData;

	function formatCount(value: number | null | undefined): string {
		if (value === null || value === undefined) return '...';
		return new Intl.NumberFormat('fi-FI').format(value);
	}

	// What the database holds, counted from the active data version.
	$: counts = [
		{ value: data.activeVersionStats?.villains, label: 'roistoa' },
		{ value: data.activeVersionStats?.stories, label: 'tarinaa' },
		{ value: data.activeVersionStats?.writers, label: 'kertojaa' },
		{ value: data.activeVersionStats?.drawers, label: 'piirtäjää' },
		{ value: data.activeVersionStats?.translators, label: 'suomentajaa' }
	];
</script>

<svelte:head>
	<title>Texin roistot -tietokanta</title>
</svelte:head>

<PaperNotice>
	<section class="page-front">
		<img src="/tex-logo.jpg" class="front-logo" alt="Kuva Tex Willeristä" />
		<h1>Texin roistot -tietokanta</h1>

		<p class="counts-intro">Tex Willer -sarjakuvan tietokannassa on</p>
		<ul class="counts">
			{#each counts as count (count.label)}
				<li><strong>{formatCount(count.value)}</strong> <span>{count.label}</span></li>
			{/each}
		</ul>
		<p class="counts-note">
			Tilasto sisältää kaikki Suomessa ennen vuotta 2020 julkaistut Tex Willer -tarinat roistoineen.
		</p>
	</section>
</PaperNotice>

<style>
	.page-front {
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

	.counts-intro {
		margin: 0 0 0.6rem;
		font-size: 1.1rem;
	}

	/* Two columns: the counts right-aligned so they line up, labels beside them. */
	.counts {
		display: grid;
		grid-template-columns: auto auto;
		justify-content: center;
		column-gap: 0.6rem;
		row-gap: 0.15rem;
		margin: 0 auto;
		padding: 0;
		list-style: none;
		font-size: 1.15rem;
	}

	.counts li {
		display: contents;
	}

	.counts strong {
		text-align: right;
		font-variant-numeric: tabular-nums;
	}

	.counts span {
		text-align: left;
	}

	.counts-note {
		margin: 1rem auto 0;
		max-width: 30rem;
		font-size: 0.95rem;
		font-style: italic;
		color: var(--ink-soft);
	}

	@media (max-width: 640px) {
		h1 {
			font-size: 1.6rem;
		}
	}
</style>
