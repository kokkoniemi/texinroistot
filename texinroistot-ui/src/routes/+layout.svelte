<script lang="ts">
	import '../app.css';
	import { page } from '$app/stores';
	import type { LayoutData } from './$types';

	export let data: LayoutData;

	function formatLastUpdated(rawDate: string | null | undefined): string {
		if (!rawDate) return 'ei tiedossa';

		const date = new Date(rawDate);
		if (Number.isNaN(date.getTime())) return 'ei tiedossa';

		return new Intl.DateTimeFormat('fi-FI', {
			day: '2-digit',
			month: '2-digit',
			year: 'numeric',
			hour: '2-digit',
			minute: '2-digit'
		}).format(date);
	}
</script>

{#if $page.url.pathname !== '/julkaisematon'}
	<ul class="top-menu">
		<li class:active={$page.url.pathname === '/'}>
			<a href="/">Etusivu</a>
		</li>
		<li class:active={$page.url.pathname === '/roistot'}>
			<a href="/roistot?sort=first_name&publication=fi">Roistot</a>
		</li>
		<li class:active={$page.url.pathname === '/tarinat'}>
			<a href="/tarinat">Tarinat</a>
		</li>
		<li class:active={$page.url.pathname === '/tekijat'}>
			<a href="/tekijat">Tekijät</a>
		</li>
	</ul>
{/if}

<slot />

{#if $page.url.pathname !== '/julkaisematon'}
	<hr />
	<p>Sisältö päivitetty: {formatLastUpdated(data.activeVersionCreatedAt)}</p>
	<p>Järjestelmän versio: {data.systemVersion}</p>
	<p>Kaikki oikeudet pidätetään | <a href="/hallinta">Hallinta</a></p>
{/if}

<style>
	.top-menu {
		list-style-type: none;
		margin: 0;
		padding: 0 0 0 20px;
		display: block;
		border-bottom: 1px solid black;
	}
	.top-menu li {
		display: inline-block;
		text-align: center;
		width: 100px;
		padding: 10px 0;
		border-width: 1px 1px 0 1px;
		border-color: black;
		border-style: solid;
		margin-bottom: -1px;
	}
	.top-menu li.active {
		border-bottom: 1px solid #ffffed;
		font-weight: bold;
	}
	.top-menu li a,
	.top-menu li a:active,
	.top-menu li a:visited {
		color: black;
		text-decoration: none;
	}
</style>
