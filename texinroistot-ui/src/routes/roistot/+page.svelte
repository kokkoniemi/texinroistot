<script lang="ts">
	import { navigating } from '$app/stores';
	import {
		authorList,
		buildPageHref,
		hasValues,
		joinValues,
		nonItalianTitlesByFirstPublication,
		paginationTokens
	} from '$lib/listing/shared';
	import type { Meta, PaginationToken, StoryBase } from '$lib/listing/shared';
	import type { PageData } from './$types';

	export let data: PageData;

	type Story = StoryBase;

	type StoryVillain = {
		nicknames?: string[] | null;
		otherNames?: string[] | null;
		codeNames?: string[] | null;
		roles?: string[] | null;
		destiny?: string[] | null;
		story?: Story | null;
	};

	type Villain = {
		firstNames?: string[] | null;
		lastName?: string | null;
		ranks?: string[] | null;
		as?: StoryVillain[] | null;
	};

	type Filters = {
		publication: string;
		sort: string;
		q: string;
	};

	const sortOptions = [
		{ value: 'first_name', label: 'Etunimen mukaan' },
		{ value: 'last_name', label: 'Sukunimen mukaan' },
		{ value: 'nickname', label: 'Lempinimen mukaan' },
		{ value: 'other_name', label: 'Etnisen nimen mukaan' },
		{ value: 'rank', label: 'Arvon mukaan' },
		{ value: 'fi_pub_date', label: 'Suomen julkaisupäivän mukaan' },
		{ value: 'it_pub_date', label: 'Alkuperäisessä ilmestymisjärjestyksessä (Italia)' }
	];

	let villains: Villain[] = [];
	let meta: Meta = { total: 0, page: 1, pageSize: 25, totalPages: 0 };
	let filters: Filters = { publication: 'fi', sort: 'fi_pub_date', q: '' };
	let hasPrev = false;
	let hasNext = false;
	let isFilterLoading = false;
	let pageTokens: PaginationToken[] = [];

	$: villains = data.villains ?? [];
	$: meta = data.meta ?? { total: 0, page: 1, pageSize: 25, totalPages: 0 };
	$: filters = data.filters ?? { publication: 'fi', sort: 'fi_pub_date', q: '' };
	$: hasPrev = meta.page > 1;
	$: hasNext = meta.page < meta.totalPages;
	$: isFilterLoading = Boolean($navigating) && $navigating?.to?.url.pathname === '/roistot';
	$: pageTokens = paginationTokens(meta.page, meta.totalPages);

	function villainRealName(villain: Villain): string {
		const firstNames = joinValues(villain.firstNames, '').trim();
		const lastName = (villain.lastName ?? '').trim();
		return `${firstNames} ${lastName}`.trim();
	}

	function villainNicknames(villain: Villain): string[] {
		return (villain.as ?? [])
			.flatMap((appearance) => appearance.nicknames ?? [])
			.map((nickname) => nickname.trim())
			.filter((nickname, index, values) => Boolean(nickname) && values.indexOf(nickname) === index);
	}

	function appearanceCodeNames(appearance?: StoryVillain | null): string[] {
		return (appearance?.codeNames ?? [])
			.map((codeName) => codeName.trim())
			.filter((codeName, index, values) => Boolean(codeName) && values.indexOf(codeName) === index);
	}

	type VillainTitle = {
		baseName: string;
		nicknames: string[];
	};

	function villainTitle(villain: Villain): VillainTitle {
		const rank = joinValues(villain.ranks, '').trim();
		const realName = villainRealName(villain);
		const nicknames = villainNicknames(villain);
		const baseName = `${rank} ${realName}`.trim();
		return { baseName, nicknames };
	}

	function primaryAppearance(villain: Villain): StoryVillain | null {
		const appearances = villain.as ?? [];
		return appearances.length > 0 ? appearances[0] : null;
	}

	function storyTitle(story?: Story | null): string {
		if (!story) return '-';
		const publications = story.publications ?? [];
		const baseTitles = nonItalianTitlesByFirstPublication(
			publications.filter((publication) => publication.in?.type === 'perus')
		);
		if (baseTitles.length > 0) return baseTitles[0];

		const nonItalianTitles = nonItalianTitlesByFirstPublication(publications);
		if (nonItalianTitles.length > 0) return nonItalianTitles[0];

		const anyTitle = publications.find((publication) => Boolean(publication.title?.trim()))?.title?.trim();
		return anyTitle || 'Nimetön tarina';
	}

	function authorsSummary(story?: Story | null): string {
		if (!story) return '-';
		return `kertoi: ${authorList(story.writtenBy)} | piirsi: ${authorList(story.drawnBy)} | suomensi: ${authorList(story.translatedBy)}`;
	}

	function pageHref(page: number): string {
		return buildPageHref('/roistot', {
			publication: filters.publication,
			sort: filters.sort,
			page,
			pageSize: meta.pageSize,
			q: filters.q || undefined
		});
	}
</script>

<section class="roistot-page">
	<h1>Roistot</h1>

	<form method="GET" class="filters">
		<label class="field">
			<span>Järjestys</span>
			<select name="sort" disabled={isFilterLoading}>
				{#each sortOptions as option}
					<option value={option.value} selected={filters.sort === option.value}
						>{option.label}</option
					>
				{/each}
			</select>
		</label>

		<label class="field search">
			<span>Hae hakusanalla</span>
			<input
				name="q"
				type="text"
				value={filters.q ?? ''}
				placeholder="Kirjoita hakusana..."
				disabled={isFilterLoading}
			/>
		</label>

		<input type="hidden" name="publication" value={filters.publication} />
		<input type="hidden" name="page" value="1" />
		<input type="hidden" name="pageSize" value={meta.pageSize} />

		<div class="actions">
			<button type="submit" disabled={isFilterLoading}>
				{isFilterLoading ? 'Ladataan...' : 'Hae'}
			</button>
			<a
				href="/roistot"
				class:loading-link-disabled={isFilterLoading}
				aria-disabled={isFilterLoading}>Palauta oletukset</a
			>
		</div>
	</form>

	<div class="result-row">
		<p class="result-total">Roistoja yhteensä {meta.total}</p>

		<nav class="pagination pagination-top">
			{#if hasPrev && !isFilterLoading}
				<a href={pageHref(meta.page - 1)}>Edellinen</a>
			{:else}
				<span class="disabled">Edellinen</span>
			{/if}

			{#each pageTokens as token}
				{#if token === 'ellipsis'}
					<span class="ellipsis">...</span>
				{:else if token === meta.page}
					<span class="current-page">{token}</span>
				{:else if !isFilterLoading}
					<a href={pageHref(token)}>{token}</a>
				{:else}
					<span class="disabled">{token}</span>
				{/if}
			{/each}

			{#if hasNext && !isFilterLoading}
				<a href={pageHref(meta.page + 1)}>Seuraava</a>
			{:else}
				<span class="disabled">Seuraava</span>
			{/if}
		</nav>

		<p class="result-page">
			Sivu {meta.totalPages === 0 ? 0 : meta.page} / {meta.totalPages === 0 ? 0 : meta.totalPages}
		</p>
	</div>

	{#if villains.length === 0}
		<p class="empty">Ei tuloksia valituilla hakuehdoilla.</p>
	{:else}
		<div class="villain-list">
				{#each villains as villain}
					{@const appearance = primaryAppearance(villain)}
					{@const title = villainTitle(villain)}
					{@const displayName = joinValues(appearance?.otherNames, '').trim()}
					{@const codeNames = appearanceCodeNames(appearance)}
					{@const nicknameTitle = title.nicknames.map((nickname) => `"${nickname}"`).join(', ')}
					{@const cardTitle = [title.baseName, nicknameTitle, displayName].filter(Boolean).join(', ')}
					{@const hasCodeNames = codeNames.length > 0}
					<article class="villain-card">
						<h3>{#if cardTitle}{cardTitle}{/if}{#if hasCodeNames}{cardTitle ? ', ' : ''}{#each codeNames as codeName, index}{#if index > 0}, {/if}<em>{codeName}</em>{/each}{:else if !cardTitle}Nimetön roisto{/if}</h3>
						<p><strong>Rooli:</strong> {joinValues(appearance?.roles, '-', '; ')}</p>
						{#if hasValues(appearance?.destiny)}
							<p><strong>Kohtalo:</strong> {joinValues(appearance?.destiny)}</p>
						{/if}
						{#if hasValues(appearance?.codeNames)}
							<p><strong>Salanimi:</strong> {joinValues(appearance?.codeNames)}</p>
						{/if}
						<p><strong>Tarina:</strong> {storyTitle(appearance?.story)}</p>
						<p><strong>Tekijät:</strong> {authorsSummary(appearance?.story)}</p>
					</article>
				{/each}
			</div>
	{/if}

	<nav class="pagination">
		{#if hasPrev && !isFilterLoading}
			<a href={pageHref(meta.page - 1)}>Edellinen</a>
		{:else}
			<span class="disabled">Edellinen</span>
		{/if}

		{#each pageTokens as token}
			{#if token === 'ellipsis'}
				<span class="ellipsis">...</span>
			{:else if token === meta.page}
				<span class="current-page">{token}</span>
			{:else if !isFilterLoading}
				<a href={pageHref(token)}>{token}</a>
			{:else}
				<span class="disabled">{token}</span>
			{/if}
		{/each}

		{#if hasNext && !isFilterLoading}
			<a href={pageHref(meta.page + 1)}>Seuraava</a>
		{:else}
			<span class="disabled">Seuraava</span>
		{/if}
	</nav>
</section>

<style>
	.roistot-page {
		margin: 1rem 1.5rem 0;
	}

	.filters {
		display: grid;
		grid-template-columns: minmax(300px, 360px) minmax(320px, 1fr) auto;
		gap: 0.75rem;
		align-items: end;
		padding: 0.75rem;
		border: 1px solid black;
		background-color: #f7f7f7;
	}

	.field {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
	}

	.field.search {
		min-width: 0;
	}

	.field span {
		font-size: 0.95rem;
	}

	select,
	input[type='text'] {
		font-size: 1rem;
		padding: 0.45rem 0.5rem;
		border: 1px solid black;
		background: #fff;
		width: 100%;
		box-sizing: border-box;
	}

	.actions {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		grid-column: 1 / -1;
		justify-self: start;
	}

	button {
		font-size: 1rem;
		padding: 0.45rem 1rem;
		border: 1px solid black;
		background: black;
		color: white;
		cursor: pointer;
	}

	button:disabled {
		opacity: 0.65;
		cursor: wait;
	}

	.loading-link-disabled {
		pointer-events: none;
		opacity: 0.65;
	}

	.result-row {
		display: grid;
		grid-template-columns: minmax(0, 1fr) auto minmax(0, 1fr);
		align-items: center;
		gap: 0.75rem 1rem;
		margin: 1rem 0;
	}

	.result-total,
	.result-page {
		margin: 0;
	}

	.result-page {
		text-align: right;
		justify-self: end;
	}

	.villain-list {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	.villain-card {
		border: 1px solid black;
		background-color: #f7f7f7;
		padding: 1rem;
		box-shadow: 0 4px 10px rgba(0, 0, 0, 0.15);
	}

	.villain-card h3 {
		margin: 0 0 0.25rem;
	}

	.villain-card p {
		margin: 0.25rem 0;
	}

	.pagination {
		margin: 1.25rem 0 0.5rem;
		display: flex;
		justify-content: center;
		gap: 1rem;
		flex-wrap: wrap;
	}

	.pagination-top {
		grid-column: 2;
		margin: 0;
		justify-content: center;
	}

	.pagination a,
	.pagination span {
		white-space: nowrap;
	}

	.disabled {
		color: #666;
		text-decoration: none;
	}

	.current-page {
		font-weight: 700;
		text-decoration: underline;
	}

	.ellipsis {
		color: #333;
	}

	.empty {
		border: 1px solid black;
		padding: 1rem;
		background-color: #f7f7f7;
	}

	@media (max-width: 900px) {
		.filters {
			grid-template-columns: 1fr;
		}
	}

	@media (max-width: 640px) {
		.roistot-page {
			margin: 0.75rem 0.75rem 0;
		}

		.pagination {
			justify-content: center;
			gap: 0.55rem;
		}

		.result-row {
			display: flex;
			flex-wrap: wrap;
			align-items: baseline;
			margin: 0.85rem 0;
		}

		.result-page {
			order: 2;
			margin-left: auto;
			text-align: right;
			justify-self: auto;
		}

		.result-total {
			order: 1;
		}

		.pagination-top {
			order: 3;
			width: 100%;
			justify-content: center;
		}
	}
</style>
