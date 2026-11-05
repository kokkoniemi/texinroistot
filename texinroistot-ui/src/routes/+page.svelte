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

<section class="page-front">
	<img src="/favicon.png" class="front-logo" alt="Kuva Tex Willeristä" />
	<h1>Texin roistot-tietokanta</h1>

	<p>
		Texin roistot-tietokanta sisältää tiedot
		{formatCount(data.activeVersionStats?.villains)}:sta tex willer -sarjakuvassa esiintyvästä
		roistosta, {formatCount(data.activeVersionStats?.stories)}:sta tarinasta,
		{formatCount(data.activeVersionStats?.drawers)}:sta piirtäjästä, sekä
		{formatCount(data.activeVersionStats?.writers)}:sta kirjoittajasta.
	</p>
</section>

<style>
	.page-front {
		text-align: center;
	}
	.front-logo {
		margin-top: 20px;
		width: 150px;
		height: 150px;
		border-radius: 50%;
		border: 2px solid black;
	}
</style>
