<script lang="ts">
	import { navigating } from '$app/stores';
	import type { PageData } from './$types';

	export let data: PageData;

	type Author = {
		firstName: string;
		lastName: string;
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
		inventedBy?: Author[] | null;
		publications?: StoryPublication[] | null;
	};

	type StoryVillain = {
		nicknames?: string[] | null;
		aliases?: string[] | null;
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

	const publicationOptions = [
		{ value: 'all', label: 'Näytä kaikki' },
		{ value: 'fi', label: 'Suomen julkaisut' },
		{ value: 'it', label: 'Italian julkaisut' }
	];

	const sortOptions = [
		{ value: 'first_name', label: 'Etunimen mukaan' },
		{ value: 'last_name', label: 'Sukunimen mukaan' },
		{ value: 'nickname', label: 'Lempinimen mukaan' },
		{ value: 'rank', label: 'Arvon mukaan' },
		{ value: 'fi_pub_date', label: 'Suomen julkaisupäivän mukaan' },
		{ value: 'it_pub_date', label: 'Italian julkaisupäivän mukaan' }
	];

	const publicationTypeLabels: Record<string, string> = {
		perus: 'Suomen perussarja',
		italia_perus: 'Italian perussarja',
		suur: 'Suuralbumit',
		maxi: 'Maxi-Tex',
		kirjasto: 'Kirjasto',
		kronikka: 'Kronikka',
		muu_erikois: 'Muut erikoiset',
		italia_erikois: 'Italian erikoiset'
	};

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

	function joinValues(values?: string[] | null, fallback = '-'): string {
		if (!values || values.length === 0) return fallback;
		return values.filter(Boolean).join(', ');
	}

	function hasValues(values?: string[] | null): boolean {
		return Boolean(values && values.some((value) => Boolean(value)));
	}

	function authorList(authors?: Author[] | null): string {
		if (!authors || authors.length === 0) return '-';
		return authors.map((author) => `${author.firstName} ${author.lastName}`.trim()).join(', ');
	}

	function villainRealName(villain: Villain): string {
		const firstNames = joinValues(villain.firstNames, '').trim();
		const lastName = (villain.lastName ?? '').trim();
		return `${firstNames} ${lastName}`.trim();
	}

	function villainNicknames(villain: Villain): string[] {
		return (villain.as ?? [])
			.flatMap((appearance) => appearance.nicknames ?? [])
			.filter((nickname, index, values) => Boolean(nickname) && values.indexOf(nickname) === index);
	}

	function villainTitle(villain: Villain): string {
		const realName = villainRealName(villain);
		const nicknames = villainNicknames(villain);
		const cleanNicknames = nicknames.map((nickname) => nickname.trim()).filter(Boolean);
		const nicknamesInParentheses =
			cleanNicknames.length > 0 ? `(${cleanNicknames.join(', ')})` : '';
		const quotedNicknames = cleanNicknames.map((nickname) => `"${nickname}"`).join(', ');

		if (realName && nicknamesInParentheses) return `${realName} ${nicknamesInParentheses}`;
		if (realName) return realName;
		if (quotedNicknames) return quotedNicknames;
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

	function publicationItem(publication: StoryPublication): string {
		const pub = publication.in;
		if (!pub) return publication.title;

		if (pub.year && pub.issue) return `${pub.year}/${pub.issue}`;
		if (pub.issue) return pub.issue;
		if (pub.year) return `${pub.year}`;
		return publication.title;
	}

	function publicationSummary(story?: Story | null): string {
		if (!story) return '-';

		const groups: Record<string, string[]> = {};
		for (const publication of story.publications ?? []) {
			const pType = publication.in?.type ?? 'muu_erikois';
			const item = publicationItem(publication);
			if (!groups[pType]) groups[pType] = [];
			if (!groups[pType].includes(item)) groups[pType].push(item);
		}

		const order = [
			'perus',
			'italia_perus',
			'suur',
			'maxi',
			'kirjasto',
			'kronikka',
			'muu_erikois',
			'italia_erikois'
		];
		const parts = order
			.filter((type) => groups[type]?.length)
			.map((type) => `${publicationTypeLabels[type] ?? type} ${groups[type].join(', ')}`);

		return parts.length > 0 ? parts.join('; ') : '-';
	}

	function authorsSummary(story?: Story | null): string {
		if (!story) return '-';
		return `kertoi: ${authorList(story.writtenBy)} | piirsi: ${authorList(story.drawnBy)} | ideoi: ${authorList(story.inventedBy)}`;
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
			<span>Julkaisu</span>
			<select name="publication" disabled={isFilterLoading}>
				{#each publicationOptions as option}
					<option value={option.value} selected={filters.publication === option.value}
						>{option.label}</option
					>
				{/each}
			</select>
		</label>

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

	<section class="result-header">
		<p>Roistoja yhteensä {meta.total}</p>
		<p>
			Sivu {meta.totalPages === 0 ? 0 : meta.page} / {meta.totalPages === 0 ? 0 : meta.totalPages}
		</p>
	</section>

	<nav class="pagination pagination-top">
		{#if meta.totalPages > 0 && meta.page > 1 && !isFilterLoading}
			<a href={pageHref(1)}>Ensimmäinen</a>
		{:else}
			<span class="disabled">Ensimmäinen</span>
		{/if}

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

		{#if meta.totalPages > 0 && meta.page < meta.totalPages && !isFilterLoading}
			<a href={pageHref(meta.totalPages)}>Viimeinen</a>
		{:else}
			<span class="disabled">Viimeinen</span>
		{/if}
	</nav>

	{#if villains.length === 0}
		<p class="empty">Ei tuloksia valituilla hakuehdoilla.</p>
	{:else}
		<div class="villain-list">
			{#each villains as villain}
				{@const appearance = primaryAppearance(villain)}
				<article class="villain-card">
					<h3>{villainTitle(villain)}</h3>
					{#if hasValues(villain.ranks)}
						<p><strong>Arvo:</strong> {joinValues(villain.ranks)}</p>
					{/if}
					<p><strong>Kohtalo:</strong> {joinValues(appearance?.destiny)}</p>
					<p><strong>Rooli:</strong> {joinValues(appearance?.roles)}</p>
					<p><strong>Tarina:</strong> {storyTitle(appearance?.story)}</p>
					<p><strong>Julkaisut:</strong> {publicationSummary(appearance?.story)}</p>
					<p><strong>Tekijät:</strong> {authorsSummary(appearance?.story)}</p>
				</article>
			{/each}
		</div>
	{/if}

	<nav class="pagination">
		{#if meta.totalPages > 0 && meta.page > 1 && !isFilterLoading}
			<a href={pageHref(1)}>Ensimmäinen</a>
		{:else}
			<span class="disabled">Ensimmäinen</span>
		{/if}

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

		{#if meta.totalPages > 0 && meta.page < meta.totalPages && !isFilterLoading}
			<a href={pageHref(meta.totalPages)}>Viimeinen</a>
		{:else}
			<span class="disabled">Viimeinen</span>
		{/if}
	</nav>
</section>

<style>
	.roistot-page {
		margin: 1rem 1.5rem 0;
	}

	.filters {
		display: grid;
		grid-template-columns: 220px 280px minmax(260px, 1fr);
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

	.field span {
		font-size: 0.95rem;
	}

	select,
	input {
		font-size: 1rem;
		padding: 0.45rem 0.5rem;
		border: 1px solid black;
		background: #fff;
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

	.result-header {
		display: flex;
		justify-content: space-between;
		margin: 1rem 0;
		gap: 0.75rem;
		flex-wrap: wrap;
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
		margin: 0 0 1rem;
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
			justify-content: flex-start;
			gap: 0.55rem;
		}

		.result-header {
			flex-direction: column;
			align-items: flex-start;
			margin: 0.85rem 0;
		}
	}
</style>
