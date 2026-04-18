<script lang="ts">
	import { navigating } from '$app/stores';
	import {
		authorList,
		buildPageHref,
		hasValues,
		joinValues,
		nonItalianTitlesByFirstPublication,
		paginationTokens,
		publicationSummaryFromPublications
	} from '$lib/listing/shared';
	import type { Author, Meta, PaginationToken, StoryBase } from '$lib/listing/shared';
	import type { PageData } from './$types';

	export let data: PageData;

	type ListedAuthor = Author & {
		hash?: string;
		isWriter?: boolean;
		isDrawer?: boolean;
	};

	type Story = StoryBase & {
		hash: string;
	};

	type AuthorStoriesResponse = {
		authorHash: string;
		stories: Story[];
		meta?: {
			total: number;
		};
	};

	type StoryVillain = {
		hash: string;
		nicknames?: string[] | null;
		otherNames?: string[] | null;
		codeNames?: string[] | null;
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

	type ItalianOriginalPublication = {
		title: string;
		details: string;
	};

	type Filters = {
		type: string;
		sort: string;
		q: string;
	};

	const typeOptions = [
		{ value: 'writer', label: 'Kertojat' },
		{ value: 'drawer', label: 'Piirtäjät' }
	];

	let authors: ListedAuthor[] = [];
	let meta: Meta = { total: 0, page: 1, pageSize: 25, totalPages: 0 };
	let filters: Filters = { type: 'writer', sort: 'last_name', q: '' };
	let hasPrev = false;
	let hasNext = false;
	let isFilterLoading = false;
	let pageTokens: PaginationToken[] = [];
	let expandedAuthorHashes: Record<string, boolean> = {};
	let loadingAuthorHashes: Record<string, boolean> = {};
	let errorByAuthorHash: Record<string, string> = {};
	let storiesByAuthorHash: Record<string, Story[]> = {};
	let authorListSignature = '';
	let selectedStory: Story | null = null;
	let popupStoryVillainsExpanded = false;
	let popupStoryVillainsLoading = false;
	let popupStoryVillainsError = '';
	let popupStoryVillains: Villain[] = [];
	let popupStoryVillainsLoadedHash = '';

	$: authors = data.authors ?? [];
	$: meta = data.meta ?? { total: 0, page: 1, pageSize: 25, totalPages: 0 };
	$: filters = data.filters ?? { type: 'writer', sort: 'last_name', q: '' };
	$: hasPrev = meta.page > 1;
	$: hasNext = meta.page < meta.totalPages;
	$: isFilterLoading = Boolean($navigating) && $navigating?.to?.url.pathname === '/tekijat';
	$: pageTokens = paginationTokens(meta.page, meta.totalPages);
	$: {
		const nextSignature = authors
			.map((author) => (author.hash ?? '').trim())
			.filter(Boolean)
			.join('|');
		if (nextSignature !== authorListSignature) {
			authorListSignature = nextSignature;
			expandedAuthorHashes = {};
			loadingAuthorHashes = {};
			errorByAuthorHash = {};
			storiesByAuthorHash = {};
		}
	}

	function authorName(author: ListedAuthor): string {
		return `${author.firstName} ${author.lastName}`.trim() || '-';
	}

	function resultLabel(authorType: string): string {
		return authorType === 'drawer' ? 'Piirtäjiä yhteensä' : 'Kertojia yhteensä';
	}

	function pageHref(page: number): string {
		return buildPageHref('/tekijat', {
			type: filters.type,
			sort: 'last_name',
			page,
			pageSize: meta.pageSize,
			q: filters.q || undefined
		});
	}

	function normalizeAuthorHash(raw?: string): string {
		return (raw ?? '').trim();
	}

	function hasFetchedAuthorStories(authorHash: string): boolean {
		return Object.prototype.hasOwnProperty.call(storiesByAuthorHash, authorHash);
	}

	function authorStories(authorHash: string): Story[] {
		return storiesByAuthorHash[authorHash] ?? [];
	}

	function finnishPublications(story: Story) {
		return (story.publications ?? []).filter(
			(publication) => !publication.in?.type?.startsWith('italia_')
		);
	}

	function storyTitle(story: Story): string {
		const publications = finnishPublications(story);
		const baseTitles = nonItalianTitlesByFirstPublication(
			publications.filter((publication) => publication.in?.type === 'perus')
		);
		if (baseTitles.length > 0) return baseTitles[0];

		const nonItalianTitles = nonItalianTitlesByFirstPublication(publications);
		if (nonItalianTitles.length > 0) return nonItalianTitles[0];

		const anyTitle = publications.find((publication) => Boolean(publication.title?.trim()))?.title?.trim();
		return anyTitle || 'Nimetön tarina';
	}

	function publicationSummary(story: Story): string {
		return publicationSummaryFromPublications(finnishPublications(story), 'Ei julkaisutietoja');
	}

	function normalizeStoryHash(raw?: string | null): string {
		return (raw ?? '').trim();
	}

	function italianOriginalPublication(story: Story): ItalianOriginalPublication | null {
		const italianPublications = (story.publications ?? []).filter((publication) =>
			publication.in?.type?.startsWith('italia_')
		);
		if (italianPublications.length === 0) return null;

		const titles = italianPublications
			.map((publication) => publication.title.trim())
			.filter((title, index, values) => Boolean(title) && values.indexOf(title) === index);
		const titlePart = titles.join('; ');

		const issues = italianPublications
			.map((publication) => {
				const issue = (publication.in?.issue ?? '').trim();
				const year = publication.in?.year ?? 0;
				const sortIssue = Number.parseInt(issue.replace(/\D/g, ''), 10);
				return {
					issue,
					year,
					sortIssue: Number.isNaN(sortIssue) ? Number.MAX_SAFE_INTEGER : sortIssue
				};
			})
			.filter((entry) => entry.issue && entry.year > 0)
			.filter(
				(entry, index, values) =>
					values.findIndex((other) => other.issue === entry.issue && other.year === entry.year) ===
					index
			)
			.sort((a, b) => {
				if (a.year !== b.year) return a.year - b.year;
				if (a.sortIssue !== b.sortIssue) return a.sortIssue - b.sortIssue;
				return a.issue.localeCompare(b.issue);
			});

		let issuePart = '';
		if (issues.length === 1) {
			issuePart = `${issues[0].issue}/${issues[0].year}`;
		} else if (issues.length > 1) {
			const first = issues[0];
			const last = issues[issues.length - 1];
			issuePart = `${first.issue}/${first.year}-${last.issue}/${last.year}`;
		}

		let details = issuePart;
		if (story.orderNumber > 0) {
			const storyNumberPart = `(tarina nro ${story.orderNumber})`;
			details = details ? `${details} ${storyNumberPart}` : storyNumberPart;
		}

		if (!titlePart && !details) return null;
		return { title: titlePart, details };
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
		const codeNames = (storyVillainForStory(villain, storyHash)?.codeNames ?? [])
			.map((codeName) => codeName.trim())
			.filter((codeName, index, values) => Boolean(codeName) && values.indexOf(codeName) === index);
		const quotedNicknames = nicknames.map((nickname) => `"${nickname}"`);

		if (realName && quotedNicknames.length > 0) return [realName, ...quotedNicknames].join(', ');
		if (realName) return realName;
		if (quotedNicknames.length > 0) return quotedNicknames.join(', ');
		if (codeNames.length > 0) return codeNames.join(', ');
		return 'Nimetön roisto';
	}

	function openStoryPopup(story: Story): void {
		selectedStory = story;
		popupStoryVillainsExpanded = false;
		popupStoryVillainsLoading = false;
		popupStoryVillainsError = '';
		popupStoryVillains = [];
		popupStoryVillainsLoadedHash = '';
	}

	function closeStoryPopup(): void {
		selectedStory = null;
		popupStoryVillainsExpanded = false;
		popupStoryVillainsLoading = false;
		popupStoryVillainsError = '';
		popupStoryVillains = [];
		popupStoryVillainsLoadedHash = '';
	}

	function handleWindowKeydown(event: KeyboardEvent): void {
		if (event.key === 'Escape' && selectedStory) {
			closeStoryPopup();
		}
	}

	async function togglePopupStoryVillains(): Promise<void> {
		if (!selectedStory) return;
		const storyHash = normalizeStoryHash(selectedStory.hash);
		if (!storyHash) return;

		if (popupStoryVillainsExpanded) {
			popupStoryVillainsExpanded = false;
			return;
		}

		popupStoryVillainsExpanded = true;
		if (popupStoryVillainsLoadedHash === storyHash || popupStoryVillainsLoading) {
			return;
		}

		popupStoryVillainsLoading = true;
		popupStoryVillainsError = '';

		try {
			const response = await fetch(`/api/tarinat/${encodeURIComponent(storyHash)}/roistot`);
			if (!response.ok) {
				throw new Error(`Roistojen haku epäonnistui (${response.status})`);
			}
			const payload = (await response.json()) as StoryVillainsResponse;
			if (normalizeStoryHash(selectedStory?.hash) === storyHash) {
				popupStoryVillains = payload.villains ?? [];
				popupStoryVillainsLoadedHash = storyHash;
			}
		} catch (error) {
			if (normalizeStoryHash(selectedStory?.hash) === storyHash) {
				popupStoryVillainsError =
					error instanceof Error ? error.message : 'Roistojen haku epäonnistui';
			}
		} finally {
			if (normalizeStoryHash(selectedStory?.hash) === storyHash) {
				popupStoryVillainsLoading = false;
			}
		}
	}

	async function toggleAuthorStories(author: ListedAuthor): Promise<void> {
		const authorHash = normalizeAuthorHash(author.hash);
		if (!authorHash) {
			return;
		}

		const isCurrentlyExpanded = Boolean(expandedAuthorHashes[authorHash]);
		if (isCurrentlyExpanded) {
			expandedAuthorHashes = { ...expandedAuthorHashes, [authorHash]: false };
			return;
		}

		expandedAuthorHashes = { ...expandedAuthorHashes, [authorHash]: true };
		if (hasFetchedAuthorStories(authorHash) || loadingAuthorHashes[authorHash]) {
			return;
		}

		loadingAuthorHashes = { ...loadingAuthorHashes, [authorHash]: true };
		errorByAuthorHash = { ...errorByAuthorHash, [authorHash]: '' };

		try {
			const response = await fetch(
				`/api/tekijat/${encodeURIComponent(authorHash)}/tarinat?type=${encodeURIComponent(filters.type)}`
			);
			if (!response.ok) {
				throw new Error(`Tarinoiden haku epäonnistui (${response.status})`);
			}
			const payload = (await response.json()) as AuthorStoriesResponse;
			storiesByAuthorHash = { ...storiesByAuthorHash, [authorHash]: payload.stories ?? [] };
		} catch (error) {
			errorByAuthorHash = {
				...errorByAuthorHash,
				[authorHash]: error instanceof Error ? error.message : 'Tarinoiden haku epäonnistui'
			};
		} finally {
			loadingAuthorHashes = { ...loadingAuthorHashes, [authorHash]: false };
		}
	}
</script>

<svelte:window on:keydown={handleWindowKeydown} />

<svelte:head>
	<title>Tekijät – Texin roistot</title>
</svelte:head>

<section class="tekijat-page">
	<h1>Tekijät</h1>

	<form method="GET" class="filters">
		<label class="field">
			<span>Ryhmä</span>
			<select name="type" disabled={isFilterLoading}>
				{#each typeOptions as option (option.value)}
					<option value={option.value} selected={filters.type === option.value}>{option.label}</option>
				{/each}
			</select>
		</label>

		<label class="field search">
			<span>Hae nimellä</span>
			<input
				name="q"
				type="text"
				value={filters.q ?? ''}
				placeholder="Kirjoita nimi..."
				disabled={isFilterLoading}
			/>
		</label>

		<input type="hidden" name="page" value="1" />
		<input type="hidden" name="pageSize" value={meta.pageSize} />
		<input type="hidden" name="sort" value="last_name" />

		<div class="actions">
			<button type="submit" disabled={isFilterLoading}>
				{isFilterLoading ? 'Ladataan...' : 'Hae'}
			</button>
			<a
				href="/tekijat"
				class:loading-link-disabled={isFilterLoading}
				aria-disabled={isFilterLoading}>Palauta oletukset</a
			>
		</div>
	</form>

	<div class="result-row">
		<p class="result-total">{resultLabel(filters.type)} {meta.total}</p>

		<nav class="pagination pagination-top">
			{#if hasPrev && !isFilterLoading}
				<a href={pageHref(meta.page - 1)}>Edellinen</a>
			{:else}
				<span class="disabled">Edellinen</span>
			{/if}

			{#each pageTokens as token, i (i)}
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

	{#if authors.length === 0}
		<p class="empty">Ei tuloksia valituilla hakuehdoilla.</p>
	{:else}
		<div class="author-list">
			{#each authors as author (author.hash ?? authorName(author))}
				{@const authorHash = normalizeAuthorHash(author.hash)}
				<article class="author-card">
					<h3>
						<button
							type="button"
							class="author-link"
							on:click={() => toggleAuthorStories(author)}
							disabled={!authorHash}
						>
							{authorName(author)}
						</button>
					</h3>

					{#if expandedAuthorHashes[authorHash]}
						<section class="author-stories">
							{#if loadingAuthorHashes[authorHash]}
								<p>Haetaan tarinoita...</p>
							{:else if errorByAuthorHash[authorHash]}
								<p class="author-stories-error">{errorByAuthorHash[authorHash]}</p>
							{:else if authorStories(authorHash).length === 0}
								<p>Tekijälle ei löytynyt tarinoita.</p>
							{:else}
								<ul class="author-story-list">
									{#each authorStories(authorHash) as story (story.hash)}
										<li>
											<button
												type="button"
												class="author-story-link"
												on:click={() => openStoryPopup(story)}
											>
												{storyTitle(story)}
											</button>
											<br />
											<span>{publicationSummary(story)}</span>
										</li>
									{/each}
								</ul>
							{/if}
						</section>
					{/if}
				</article>
			{/each}
		</div>
	{/if}

	{#if selectedStory}
		{@const selectedStoryHash = normalizeStoryHash(selectedStory.hash)}
		{@const italianOriginal = italianOriginalPublication(selectedStory)}
		<div class="story-popup-backdrop" role="presentation" on:click|self={closeStoryPopup}>
			<div class="story-popup" role="dialog" aria-modal="true" aria-labelledby="story-popup-title">
				<div class="story-popup-actions">
					<button type="button" class="story-popup-close" on:click={closeStoryPopup}>Sulje</button>
				</div>

				<article class="story-card popup-story-card">
					<h3 id="story-popup-title">{storyTitle(selectedStory)}</h3>
					<p><strong>Kertoi:</strong> {authorList(selectedStory.writtenBy, '; ')}</p>
					<p><strong>Piirsi:</strong> {authorList(selectedStory.drawnBy, '; ')}</p>
					<p><strong>Suomensi:</strong> {authorList(selectedStory.translatedBy, '; ')}</p>
					{#if italianOriginal}
						<p>
							<strong>Alkuperäisjulkaisu (Italia):</strong>
							{#if italianOriginal.title}
								<em>{italianOriginal.title}</em>
							{/if}
							{#if italianOriginal.details}
								{italianOriginal.title ? ', ' : ''}{italianOriginal.details}
							{/if}
						</p>
					{/if}
					<p><strong>Ilmestynyt Suomessa:</strong> {publicationSummary(selectedStory)}</p>

					<button
						type="button"
						class="toggle-villains"
						on:click={togglePopupStoryVillains}
						disabled={!selectedStoryHash}
					>
						{#if popupStoryVillainsExpanded}
							Piilota tarinan roistot
						{:else}
							Näytä tarinan roistot
						{/if}
					</button>

					{#if popupStoryVillainsExpanded}
						<section class="story-villains">
							{#if popupStoryVillainsLoading}
								<p>Haetaan roistoja...</p>
							{:else if popupStoryVillainsError}
								<p class="villain-error">{popupStoryVillainsError}</p>
							{:else if popupStoryVillains.length === 0}
								<p>Tarinalle ei löytynyt roistoja.</p>
							{:else}
								<div class="story-villains-list">
									{#each popupStoryVillains as villain (villain.hash)}
										{@const appearance = storyVillainForStory(villain, selectedStoryHash)}
										{@const baseTitle = villainTitle(villain, selectedStoryHash)}
										{@const displayName = joinValues(appearance?.otherNames, '').trim()}
										{@const cardTitle =
											displayName && baseTitle === 'Nimetön roisto'
												? displayName
												: displayName
													? `${baseTitle}, ${displayName}`
													: baseTitle}
										<article class="story-villain-card">
											<h4>{cardTitle}</h4>
											{#if hasValues(villain.ranks)}
												<p><strong>Arvo:</strong> {joinValues(villain.ranks)}</p>
											{/if}
											{#if hasValues(appearance?.roles)}
												<p><strong>Rooli:</strong> {joinValues(appearance?.roles, '-', '; ')}</p>
											{/if}
											{#if hasValues(appearance?.destiny)}
												<p><strong>Kohtalo:</strong> {joinValues(appearance?.destiny, '-', '; ')}</p>
											{/if}
											{#if hasValues(appearance?.codeNames)}
												<p><strong>Salanimi:</strong> {joinValues(appearance?.codeNames)}</p>
											{/if}
										</article>
									{/each}
								</div>
							{/if}
						</section>
					{/if}
				</article>
			</div>
		</div>
	{/if}

	<nav class="pagination">
		{#if hasPrev && !isFilterLoading}
			<a href={pageHref(meta.page - 1)}>Edellinen</a>
		{:else}
			<span class="disabled">Edellinen</span>
		{/if}

		{#each pageTokens as token, i (i)}
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
	.tekijat-page {
		margin: 1rem 1.5rem 0;
	}

	.filters {
		display: grid;
		grid-template-columns: minmax(220px, 260px) minmax(300px, 1fr) auto;
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

	.author-list {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	.author-card {
		border: 1px solid black;
		background-color: #f7f7f7;
		padding: 1rem;
		box-shadow: 0 4px 10px rgba(0, 0, 0, 0.15);
	}

	.author-card h3 {
		margin: 0 0 0.25rem;
	}

	.author-link {
		border: 0;
		padding: 0;
		background: transparent;
		font: inherit;
		font-weight: 700;
		color: inherit;
		text-align: left;
		cursor: pointer;
		text-decoration: underline;
	}

	.author-link:disabled {
		text-decoration: none;
		color: #777;
		cursor: default;
	}

	.author-stories {
		border: 1px solid black;
		background: #fff;
		margin-top: 0.55rem;
		padding: 0.65rem 0.8rem;
	}

	.author-story-list {
		margin: 0;
		padding-left: 1.1rem;
		display: grid;
		gap: 0.4rem;
	}

	.author-story-link {
		border: 0;
		padding: 0;
		background: transparent;
		color: inherit;
		font: inherit;
		font-weight: 700;
		text-align: left;
		text-decoration: underline;
		cursor: pointer;
	}

	.story-popup-backdrop {
		position: fixed;
		inset: 0;
		z-index: 1000;
		background: rgba(0, 0, 0, 0.5);
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 1rem;
	}

	.story-popup {
		width: min(920px, 100%);
		max-height: calc(100vh - 2rem);
		overflow: auto;
		background: #ffffed;
		border: 1px solid black;
		box-shadow: 0 10px 30px rgba(0, 0, 0, 0.35);
		padding: 0.9rem;
	}

	.story-popup-actions {
		display: flex;
		justify-content: flex-end;
		margin-bottom: 0.6rem;
	}

	.story-popup-close {
		padding: 0.35rem 0.7rem;
		font-size: 0.95rem;
	}

	.popup-story-card {
		margin: 0;
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

	.author-stories-error {
		margin: 0;
		color: #8b0000;
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
		.tekijat-page {
			margin: 0.75rem 0.75rem 0;
		}

		.story-popup-backdrop {
			padding: 0.5rem;
		}

		.story-popup {
			max-height: calc(100vh - 1rem);
			padding: 0.55rem;
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
