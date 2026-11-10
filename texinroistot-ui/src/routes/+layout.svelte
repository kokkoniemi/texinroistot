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
	<header class="site-header">
		<a class="brand" href="/">
			<img src="/tex-logo.jpg" alt="" class="brand-logo" />
			<span class="brand-text">
				<span class="brand-title">Texin roistot</span>
				<span class="brand-subtitle">Tex Willerin vastustajat, tarinat ja tekijät</span>
			</span>
		</a>
		<nav aria-label="Päävalikko">
			<ul class="top-menu">
				<li class:active={$page.url.pathname === '/'}>
					<a href="/">Etusivu</a>
				</li>
				<li class:active={$page.url.pathname === '/roistot'}>
					<a href="/roistot?sort=default&publication=fi">Roistot</a>
				</li>
				<li class:active={$page.url.pathname === '/tarinat'}>
					<a href="/tarinat">Tarinat</a>
				</li>
				<li class:active={$page.url.pathname === '/tekijat'}>
					<a href="/tekijat">Tekijät</a>
				</li>
			</ul>
		</nav>
	</header>
{/if}

<!-- Shared SVG filter: frays the edges of paper elements (filter: url(#ragged-edge)). -->
<svg class="svg-defs" width="0" height="0" aria-hidden="true">
	<filter id="ragged-edge" x="-5%" y="-5%" width="110%" height="110%">
		<feTurbulence type="fractalNoise" baseFrequency="0.04" numOctaves="4" seed="3" result="n" />
		<feDisplacementMap
			in="SourceGraphic"
			in2="n"
			scale="9"
			xChannelSelector="R"
			yChannelSelector="G"
		/>
	</filter>
</svg>

<div class="paper-folds" aria-hidden="true"></div>

<main class="site-main">
	<slot />
</main>

{#if $page.url.pathname !== '/julkaisematon'}
	<footer class="site-footer">
		<p>Sisältö päivitetty: {formatLastUpdated(data.activeVersionCreatedAt)}</p>
		<p>Järjestelmän versio: {data.systemVersion}</p>
		<p>Kaikki oikeudet pidätetään | <a href="/hallinta">Hallinta</a></p>
	</footer>
{/if}

<style>
	/* Dark saddle-leather band with the site name, and leather tabs below it. */
	.site-header {
		background: linear-gradient(180deg, var(--leather-light), var(--leather));
		border-bottom: 3px solid var(--gold);
		box-shadow: 0 3px 10px var(--shadow);
		padding: 0.9rem 1.5rem 0;
	}

	.brand {
		display: inline-flex;
		align-items: center;
		gap: 0.85rem;
		color: var(--on-accent);
		text-decoration: none;
	}

	.brand:hover {
		color: var(--on-accent);
	}

	.brand-logo {
		width: 56px;
		height: 56px;
		border-radius: 50%;
		border: 2px solid var(--gold);
		background: var(--card-bg);
		object-fit: cover;
	}

	.brand-text {
		display: grid;
	}

	.brand-title {
		font-family: var(--font-display);
		font-size: 1.9rem;
		line-height: 1.1;
		letter-spacing: 0.03em;
		color: var(--card-bg);
		text-shadow: 0 2px 0 rgba(0, 0, 0, 0.35);
	}

	.brand-subtitle {
		font-size: 0.9rem;
		font-style: italic;
		color: var(--border-soft);
	}

	.top-menu {
		list-style-type: none;
		margin: 0.9rem 0 0;
		padding: 0;
		display: flex;
		flex-wrap: wrap;
		gap: 0.35rem;
	}

	.top-menu li {
		margin-bottom: -3px;
	}

	.top-menu li a {
		display: block;
		min-width: 6.5rem;
		padding: 0.55rem 1rem;
		text-align: center;
		font-family: var(--font-display);
		letter-spacing: 0.03em;
		color: var(--border-soft);
		text-decoration: none;
		border: 1px solid rgba(201, 174, 134, 0.35);
		border-bottom: 0;
		border-radius: 6px 6px 0 0;
		background: rgba(0, 0, 0, 0.18);
	}

	.top-menu li a:hover {
		color: var(--card-bg);
		background: rgba(0, 0, 0, 0.3);
	}

	.top-menu li.active a {
		color: var(--leather);
		background: var(--page-bg);
		border-color: var(--gold);
		border-bottom: 3px solid var(--page-bg);
	}

	.svg-defs {
		position: absolute;
	}

	.site-main {
		min-height: 60vh;
	}

	.site-footer {
		margin-top: 2.5rem;
		padding: 1rem 1.5rem 1.5rem;
		border-top: 3px double var(--border-soft);
		color: var(--ink-soft);
		font-size: 0.9rem;
		text-align: center;
	}

	.site-footer p {
		margin: 0.2rem 0;
	}

	@media (max-width: 640px) {
		.site-header {
			padding: 0.75rem 0.75rem 0;
		}

		.brand-logo {
			width: 44px;
			height: 44px;
		}

		.brand-title {
			font-size: 1.5rem;
		}

		.top-menu {
			gap: 0.2rem;
			flex-wrap: nowrap;
		}

		.top-menu li a {
			min-width: 0;
			padding: 0.45rem 0.55rem;
			font-size: 0.85rem;
		}
	}
</style>
