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

	<h2>Selaa roistoja</h2>

	<p>
		<a href="/roistot?publication=all&sort=first_name">Etunimen mukaan</a> |
		<a href="/roistot?publication=all&sort=last_name">Sukunimen mukaan</a> |
		<a href="/roistot?publication=all&sort=nickname">Lempinimen mukaan</a> |
		<a href="/roistot?publication=all&sort=other_name">Etnisen nimen mukaan</a> |
		<a href="/roistot?publication=all&sort=rank">Arvon mukaan</a> |
		<a href="/roistot?publication=all&sort=fi_pub_date">Näytä kaikki</a> |
		<a href="/roistot">Hae hakusanalla</a>
	</p>
	<p>
		Etninen nimi on sama kuin "Nimi", eli "Ei etunimi-sukunimi-muotoiset nimet (esim.
		intiaani- ja kiinalaisten nimet)".
	</p>

	<h2>Selaa tarinoita</h2>

	<p>
		<a href="/tarinat?publication=all&sort=alpha">Aakkosjärjestyksessä</a> |
		<a href="/tarinat?publication=perus_fi&sort=fi_pub_date">Suomen julkaisupäivän mukaan</a> |
		<a href="/tarinat?publication=perus_it&sort=it_pub_date">Alkuperäisessä ilmestymisjärjestyksessä (Italia)</a>
		| <a href="/tarinat?publication=all&sort=fi_pub_date">Näytä kaikki</a> |
		<a href="/tarinat?publication=suur&sort=alpha">Suuralbumit</a> |
		<a href="/tarinat?publication=maxi&sort=alpha">Maxi Texit</a> |
		<a href="/tarinat?publication=kirjasto&sort=alpha">Kirjastot</a>
		<!-- TODO: Poista kronikat etusivulta ja listaussivujen filttereistä -->
		<!-- TODO: Filtteriksi "Muut erikoisjulkaisut" -->
		<!-- I-->
		<a href="/tarinat">Hae hakusanalla</a>
	</p>

	<p>
				<a href="/tarinat?publication=serie_extra&sort=alpha">Serie extra</a> |
				<a href="/tarinat?publication=texone&sort=alpha">Texone</a> |
				<a href="/tarinat?publication=mini_texone_maxi_tex&sort=alpha">Mini Texone & Maxi Tex</a> |
				<a href="/tarinat?publication=almanacco_del_west&sort=alpha">Almanacco del West</a> |
				<a href="/tarinat?publication=color_tex&sort=alpha">Color Tex</a> |
				<a href="/tarinat?publication=tex_romanzi_a_fumetti&sort=alpha">Tex romanzi a fumetti</a> |
				<a href="/tarinat?publication=tex_magazine&sort=alpha">Tex Magazine</a>
	</p>

	<h2>Selaa tekijöitä</h2>

	<p>
		<a href="/tekijat?type=writer&sort=last_name">Kertojat</a> |
		<a href="/tekijat?type=drawer&sort=last_name">Piirtäjät</a>
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
