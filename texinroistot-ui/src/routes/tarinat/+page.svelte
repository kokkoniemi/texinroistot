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
		hash: string;
		orderNumber: number;
		writtenBy?: Author[] | null;
		drawnBy?: Author[] | null;
		inventedBy?: Author[] | null;
		publications?: StoryPublication[] | null;
	};

	type StoryVillain = {
		hash: string;
		nicknames?: string[] | null;
		aliases?: string[] | null;
		roles?: string[] | null;
		destiny?: string[] | null;
		story?: { hash: string } | null;
	};

	type Villain = {
		hash: string;
		ranks?: string[] | null;
		firstNames?: string[] | null;
		lastName?: string | null;
		as?: StoryVillain[] | null;
	};

	type StoryVillainsResponse = {
		storyHash: string;
		villains: Villain[];
		meta?: {
			total: number;
		};
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
		{ value: 'perus_fi', label: 'Suomen perussarja' },
		{ value: 'perus_it', label: 'Italian perussarja' },
		{ value: 'suur', label: 'Suuralbumit' },
		{ value: 'maxi', label: 'Maxi-Tex' },
		{ value: 'kirjasto', label: 'Kirjasto' },
		{ value: 'kronikka', label: 'Kronikka' },
		{ value: 'special', label: 'Muut erikoiset' }
	];

	const sortOptions = [
		{ value: 'fi_pub_date', label: 'Suomen julkaisupäivän mukaan' },
		{ value: 'it_pub_date', label: 'Italian julkaisupäivän mukaan' },
		{ value: 'alpha', label: 'Aakkosjärjestyksessä' }
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

	let stories: Story[] = [];
	let meta: Meta = { total: 0, page: 1, pageSize: 25, totalPages: 0 };
	let filters: Filters = { publication: 'perus_fi', sort: 'fi_pub_date', q: '' };
	let hasPrev = false;
	let hasNext = false;
	let expandedStoryHashes: Record<string, boolean> = {};
	let loadingStoryHashes: Record<string, boolean> = {};
	let errorByStoryHash: Record<string, string> = {};
	let villainsByStoryHash: Record<string, Villain[]> = {};
	let storyListSignature = '';
	let isFilterLoading = false;
	let pageTokens: PaginationToken[] = [];

	function normalizeStoryHash(raw?: string | null): string {
		return (raw ?? '').trim();
	}

	$: stories = data.stories ?? [];
	$: meta = data.meta ?? { total: 0, page: 1, pageSize: 25, totalPages: 0 };
	$: filters = data.filters ?? { publication: 'perus_fi', sort: 'fi_pub_date', q: '' };
	$: hasPrev = meta.page > 1;
	$: hasNext = meta.page < meta.totalPages;
	$: isFilterLoading = Boolean($navigating) && $navigating?.to?.url.pathname === '/tarinat';
	$: pageTokens = paginationTokens(meta.page, meta.totalPages);
	$: {
		const nextSignature = stories.map((story) => normalizeStoryHash(story.hash)).join('|');
		if (nextSignature !== storyListSignature) {
			storyListSignature = nextSignature;
			expandedStoryHashes = {};
			loadingStoryHashes = {};
			errorByStoryHash = {};
			villainsByStoryHash = {};
		}
	}

	function authorList(authors?: Author[] | null): string {
		if (!authors || authors.length === 0) return '-';
		return authors.map((author) => `${author.firstName} ${author.lastName}`.trim()).join(', ');
	}

	function joinValues(values?: string[] | null, fallback = '-'): string {
		if (!values || values.length === 0) return fallback;
		return values.filter(Boolean).join(', ');
	}

	function hasValues(values?: string[] | null): boolean {
		return Boolean(values && values.some((value) => Boolean(value)));
	}

	function storyTitle(story: Story): string {
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

	function italianTitles(story: Story): string {
		const titles = (story.publications ?? [])
			.filter((publication) => publication.in?.type?.startsWith('italia_') && publication.title)
			.map((publication) => publication.title)
			.filter((title, index, list) => list.indexOf(title) === index);

		return titles.join(', ');
	}

	function cardTitle(story: Story): string {
		const title = storyTitle(story);
		const italian = italianTitles(story);
		return italian ? `${title} (${italian})` : title;
	}

	function publicationItem(publication: StoryPublication): string {
		const pub = publication.in;
		if (!pub) return publication.title;

		if (pub.year && pub.issue) return `${pub.year}/${pub.issue}`;
		if (pub.issue) return pub.issue;
		if (pub.year) return `${pub.year}`;
		return publication.title;
	}

	function publicationSummary(story: Story): string {
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

		return parts.length > 0 ? parts.join('; ') : 'Ei julkaisutietoja';
	}

	function storyVillainForStory(villain: Villain, storyHash: string): StoryVillain | null {
		const appearances = villain.as ?? [];
		const matchingStory = appearances.find(
			(appearance) => normalizeStoryHash(appearance.story?.hash) === storyHash
		);
		if (matchingStory) return matchingStory;
		return appearances.length > 0 ? appearances[0] : null;
	}

	function villainRealName(villain: Villain): string {
		const firstNames = joinValues(villain.firstNames, '').trim();
		const lastName = (villain.lastName ?? '').trim();
		return `${firstNames} ${lastName}`.trim();
	}

	function villainNicknames(villain: Villain, storyHash: string): string[] {
		return (storyVillainForStory(villain, storyHash)?.nicknames ?? [])
			.map((nickname) => nickname.trim())
			.filter((nickname, index, values) => Boolean(nickname) && values.indexOf(nickname) === index);
	}

	function villainTitle(villain: Villain, storyHash: string): string {
		const realName = villainRealName(villain);
		const nicknames = villainNicknames(villain, storyHash);

		if (realName && nicknames.length > 0) return `${realName} (${nicknames.join(', ')})`;
		if (realName) return realName;
		if (nicknames.length > 0) return nicknames.map((nickname) => `"${nickname}"`).join(', ');
		return 'Nimetön roisto';
	}

	function storyVillains(storyHash: string): Villain[] {
		return villainsByStoryHash[storyHash] ?? [];
	}

	function hasFetchedStoryVillains(storyHash: string): boolean {
		return Object.prototype.hasOwnProperty.call(villainsByStoryHash, storyHash);
	}

	async function toggleStoryVillains(storyHash: string): Promise<void> {
		if (!storyHash) return;
		const isCurrentlyExpanded = Boolean(expandedStoryHashes[storyHash]);

		if (isCurrentlyExpanded) {
			expandedStoryHashes = { ...expandedStoryHashes, [storyHash]: false };
			return;
		}

		expandedStoryHashes = { ...expandedStoryHashes, [storyHash]: true };
		if (hasFetchedStoryVillains(storyHash) || loadingStoryHashes[storyHash]) {
			return;
		}

		loadingStoryHashes = { ...loadingStoryHashes, [storyHash]: true };
		errorByStoryHash = { ...errorByStoryHash, [storyHash]: '' };

		try {
			const response = await fetch(`/api/tarinat/${encodeURIComponent(storyHash)}/roistot`);
			if (!response.ok) {
				throw new Error(`Roistojen haku epäonnistui (${response.status})`);
			}
			const payload = (await response.json()) as StoryVillainsResponse;
			villainsByStoryHash = { ...villainsByStoryHash, [storyHash]: payload.villains ?? [] };
		} catch (error) {
			errorByStoryHash = {
				...errorByStoryHash,
				[storyHash]: error instanceof Error ? error.message : 'Roistojen haku epäonnistui'
			};
		} finally {
			loadingStoryHashes = { ...loadingStoryHashes, [storyHash]: false };
		}
	}

	function pageHref(page: number): string {
		const params = new URLSearchParams();
		params.set('publication', filters.publication);
		params.set('sort', filters.sort);
		params.set('page', String(page));
		params.set('pageSize', String(meta.pageSize));
		if (filters.q) params.set('q', filters.q);
		return `/tarinat?${params.toString()}`;
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

<section class="tarinat-page">
	<h1>Tarinat</h1>

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
				href="/tarinat"
				class:loading-link-disabled={isFilterLoading}
				aria-disabled={isFilterLoading}>Palauta oletukset</a
			>
		</div>
	</form>

	<section class="result-header">
		<p>Tarinoita yhteensä {meta.total}</p>
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

	{#if stories.length === 0}
		<p class="empty">Ei tuloksia valituilla hakuehdoilla.</p>
	{:else}
		<div class="story-list">
			{#each stories as story}
				{@const storyHash = normalizeStoryHash(story.hash)}
				<article class="story-card">
					<h3>{cardTitle(story)}</h3>
					<p><strong>Kirjoitti:</strong> {authorList(story.writtenBy)}</p>
					<p><strong>Piirsi:</strong> {authorList(story.drawnBy)}</p>
					<p><strong>Ideoi:</strong> {authorList(story.inventedBy)}</p>
					<p><strong>Julkaisut:</strong> {publicationSummary(story)}</p>

					<button
						type="button"
						class="toggle-villains"
						on:click={() => toggleStoryVillains(storyHash)}
						disabled={!storyHash}
					>
						{#if expandedStoryHashes[storyHash]}
							Piilota tarinan roistot
						{:else}
							Näytä tarinan roistot
						{/if}
					</button>

					{#if expandedStoryHashes[storyHash]}
						<section class="story-villains">
							{#if loadingStoryHashes[storyHash]}
								<p>Haetaan roistoja...</p>
							{:else if errorByStoryHash[storyHash]}
								<p class="villain-error">{errorByStoryHash[storyHash]}</p>
							{:else if storyVillains(storyHash).length === 0}
								<p>Tarinalle ei löytynyt roistoja.</p>
							{:else}
								<div class="story-villains-list">
									{#each storyVillains(storyHash) as villain}
										{@const appearance = storyVillainForStory(villain, storyHash)}
										<article class="story-villain-card">
											<h4>{villainTitle(villain, storyHash)}</h4>
											{#if hasValues(villain.ranks)}
												<p><strong>Arvo:</strong> {joinValues(villain.ranks)}</p>
											{/if}
											{#if hasValues(appearance?.roles)}
												<p><strong>Rooli:</strong> {joinValues(appearance?.roles)}</p>
											{/if}
											{#if hasValues(appearance?.destiny)}
												<p><strong>Kohtalo:</strong> {joinValues(appearance?.destiny)}</p>
											{/if}
											{#if hasValues(appearance?.aliases)}
												<p><strong>Alias:</strong> {joinValues(appearance?.aliases)}</p>
											{/if}
										</article>
									{/each}
								</div>
							{/if}
						</section>
					{/if}
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
	.tarinat-page {
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

	.actions button {
		font-size: 1rem;
		padding: 0.45rem 1rem;
		border: 1px solid black;
		background: black;
		color: white;
		cursor: pointer;
	}

	.actions button:disabled {
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

	.story-list {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	.story-card {
		border: 1px solid black;
		background-color: #f7f7f7;
		padding: 1rem;
		box-shadow: 0 4px 10px rgba(0, 0, 0, 0.15);
	}

	.story-card h3 {
		margin: 0 0 0.5rem;
	}

	.story-card p {
		margin: 0.25rem 0;
	}

	.toggle-villains {
		margin-top: 0.75rem;
		padding: 0.35rem 0.7rem;
		font-size: 0.95rem;
		border: 1px solid black;
		background: #fff;
		color: #111;
		cursor: pointer;
	}

	.toggle-villains:disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}

	.story-villains {
		margin-top: 0.8rem;
		padding-top: 0.75rem;
		border-top: 1px solid black;
	}

	.story-villains-list {
		display: flex;
		flex-direction: column;
		gap: 0.6rem;
	}

	.story-villain-card {
		border: 1px solid black;
		background: #fff;
		padding: 0.6rem 0.75rem;
	}

	.story-villain-card h4 {
		margin: 0 0 0.25rem;
	}

	.story-villain-card p {
		margin: 0.2rem 0;
	}

	.villain-error {
		color: #8a0000;
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
		.tarinat-page {
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
