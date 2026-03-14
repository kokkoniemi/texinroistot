<script lang="ts">
	import { navigating } from '$app/stores';
	import type { PageData } from './$types';

	export let data: PageData;

	type Author = {
		firstName: string;
		lastName: string;
		details?: string | null;
	};

	type Publication = {
		type: string;
		year: number;
		issue: string;
	};

	type StoryPublication = {
		title: string;
		in?: Publication;
	};

	type Story = {
		orderNumber: number;
		writtenBy?: Author[] | null;
		drawnBy?: Author[] | null;
		translatedBy?: Author[] | null;
		publications?: StoryPublication[] | null;
	};

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

	type Meta = {
		total: number;
		page: number;
		pageSize: number;
		totalPages: number;
	};

	type Filters = {
		publication: string;
		sort: string;
		q: string;
	};

	type PaginationToken = number | 'ellipsis';

	const sortOptions = [
		{ value: 'first_name', label: 'Etunimen mukaan' },
		{ value: 'last_name', label: 'Sukunimen mukaan' },
		{ value: 'nickname', label: 'Lempinimen mukaan' },
		{ value: 'rank', label: 'Arvon mukaan' },
		{ value: 'fi_pub_date', label: 'Suomen julkaisupäivän mukaan' },
		{ value: 'it_pub_date', label: 'Italian julkaisupäivän mukaan' }
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

	function joinValues(values?: string[] | null, fallback = '-', separator = ', '): string {
		if (!values || values.length === 0) return fallback;
		return values.filter(Boolean).join(separator);
	}

	function hasValues(values?: string[] | null): boolean {
		return Boolean(values && values.some((value) => Boolean(value)));
	}

	function authorList(authors?: Author[] | null): string {
		if (!authors || authors.length === 0) return '-';
		return authors
			.map((author) => {
				const base = `${author.firstName} ${author.lastName}`.trim();
				const details = (author.details ?? '').trim();
				if (details) return `${base} (${details})`.trim();
				return base;
			})
			.filter(Boolean)
			.join(', ');
	}

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

	function villainAlternativeNames(villain: Villain): string[] {
		return (villain.as ?? [])
			.flatMap((appearance) => [...(appearance.otherNames ?? []), ...(appearance.codeNames ?? [])])
			.map((name) => name.trim())
			.filter((name, index, values) => Boolean(name) && values.indexOf(name) === index);
	}

	function villainTitle(villain: Villain): string {
		const rank = joinValues(villain.ranks, '').trim();
		const realName = villainRealName(villain);
		const nicknames = villainNicknames(villain);
		const alternativeNames = villainAlternativeNames(villain);
		const aliases = [...nicknames, ...alternativeNames].filter(
			(name, index, values) => Boolean(name) && values.indexOf(name) === index
		);
		const baseName = `${rank} ${realName}`.trim();
		const quotedAliases = aliases.map((name) => `"${name}"`);
		const fullName = [baseName, ...quotedAliases].filter(Boolean).join(', ');

		if (fullName) return fullName;
		return 'Nimetön roisto';
	}

	function primaryAppearance(villain: Villain): StoryVillain | null {
		const appearances = villain.as ?? [];
		return appearances.length > 0 ? appearances[0] : null;
	}

	function storyTitle(story?: Story | null): string {
		if (!story) return '-';
		const publications = story.publications ?? [];
		const finBase = publications.find(
			(publication) => publication.in?.type === 'perus' && publication.title
		);
		if (finBase?.title) return finBase.title;

		const nonItalian = publications.find(
			(publication) => !publication.in?.type?.startsWith('italia_') && publication.title
		);
		if (nonItalian?.title) return nonItalian.title;

		return publications[0]?.title ?? 'Nimetön tarina';
	}

	function publicationSummary(story?: Story | null): string {
		if (!story) return '-';

		const uniqueTitles = (story.publications ?? [])
			.map((publication) => publication.title.trim())
			.filter((title, index, values) => Boolean(title) && values.indexOf(title) === index);

		return uniqueTitles.length > 0 ? uniqueTitles.join('; ') : '-';
	}

	function authorsSummary(story?: Story | null): string {
		if (!story) return '-';
		return `kertoi: ${authorList(story.writtenBy)} | piirsi: ${authorList(story.drawnBy)} | suomensi: ${authorList(story.translatedBy)}`;
	}

	function pageHref(page: number): string {
		const params = new URLSearchParams();
		params.set('publication', filters.publication);
		params.set('sort', filters.sort);
		params.set('page', String(page));
		params.set('pageSize', String(meta.pageSize));
		if (filters.q) params.set('q', filters.q);
		return `/roistot?${params.toString()}`;
	}

	function paginationTokens(currentPage: number, totalPages: number): PaginationToken[] {
		if (totalPages <= 0) return [];

		const visiblePages = new Set<number>([1, totalPages]);
		for (let page = currentPage - 1; page <= currentPage + 1; page++) {
			if (page >= 1 && page <= totalPages) visiblePages.add(page);
		}
		if (currentPage <= 3) {
			visiblePages.add(2);
			visiblePages.add(3);
		}
		if (currentPage >= totalPages - 2) {
			visiblePages.add(totalPages - 1);
			visiblePages.add(totalPages - 2);
		}

		const orderedPages = [...visiblePages]
			.filter((page) => page >= 1 && page <= totalPages)
			.sort((a, b) => a - b);
		const tokens: PaginationToken[] = [];
		let previousPage = 0;
		for (const page of orderedPages) {
			if (previousPage > 0 && page - previousPage > 1) {
				tokens.push('ellipsis');
			}
			tokens.push(page);
			previousPage = page;
		}
		return tokens;
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
				<article class="villain-card">
					<h3>{villainTitle(villain)}</h3>
					<p><strong>Rooli:</strong> {joinValues(appearance?.roles, '-', '; ')}</p>
					<p><strong>Kohtalo:</strong> {joinValues(appearance?.destiny)}</p>
					{#if hasValues(appearance?.otherNames)}
						<p><strong>Nimi:</strong> {joinValues(appearance?.otherNames)}</p>
					{/if}
					{#if hasValues(appearance?.codeNames)}
						<p><strong>Salanimi:</strong> {joinValues(appearance?.codeNames)}</p>
					{/if}
					<p><strong>Tarina:</strong> {storyTitle(appearance?.story)}</p>
					<p><strong>Julkaisut:</strong> {publicationSummary(appearance?.story)}</p>
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
