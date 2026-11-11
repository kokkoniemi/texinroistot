<script lang="ts">
	import { navigating } from '$app/stores';
	import {
		buildPageHref,
		nonItalianTitlesByFirstPublication,
		openInOwnWindow,
		paginationTokens,
		publicationSummaryFromPublications,
		translatedPartTitles
	} from '$lib/listing/shared';
	import type { Meta, PaginationToken } from '$lib/listing/shared';
	import type { Story, ListedAuthor, AuthorStoriesResponse } from '$lib/types';
	import { authorWikipediaUrl } from '$lib/data/authorWikipedia';
	import FilterForm from '$lib/components/FilterForm.svelte';
	import Pagination from '$lib/components/Pagination.svelte';
	import StoryPopup from '$lib/components/StoryPopup.svelte';
	import StoryTitle from '$lib/components/StoryTitle.svelte';
	import type { PageData } from './$types';

	export let data: PageData;

	type Filters = {
		type: string;
		sort: string;
		q: string;
	};

	const typeOptions = [
		{ value: 'writer', label: 'Kertojat' },
		{ value: 'drawer', label: 'Piirtäjät' },
		{ value: 'translator', label: 'Suomentajat' }
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
		if (authorType === 'drawer') return 'Piirtäjiä yhteensä';
		if (authorType === 'translator') return 'Suomentajia yhteensä';
		return 'Kertojia yhteensä';
	}

	function pageHref(page: number): string {
		return buildPageHref('/tekijat', {
			type: filters.type,
			sort: filters.sort,
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
		const nonItalianTitles = nonItalianTitlesByFirstPublication(publications);
		if (nonItalianTitles.length > 0) return nonItalianTitles.join('; ');

		const anyTitle = publications
			.find((publication) => Boolean(publication.title?.trim()))
			?.title?.trim();
		return anyTitle || 'Nimetön tarina';
	}

	// In the Suomentajat list, show only the titles of the parts the translator
	// translated; other groups show all titles.
	function authorStoryTitle(story: Story, authorHash: string): string {
		if (filters.type !== 'translator') return storyTitle(story);
		const titles = translatedPartTitles(story, authorHash);
		return titles.length > 0 ? titles.join('; ') : storyTitle(story);
	}

	function publicationSummary(story: Story): string {
		return publicationSummaryFromPublications(finnishPublications(story), 'Ei julkaisutietoja');
	}

	async function toggleAuthorStories(author: ListedAuthor): Promise<void> {
		const authorHash = normalizeAuthorHash(author.hash);
		if (!authorHash) return;

		const isCurrentlyExpanded = Boolean(expandedAuthorHashes[authorHash]);
		if (isCurrentlyExpanded) {
			expandedAuthorHashes = { ...expandedAuthorHashes, [authorHash]: false };
			return;
		}

		expandedAuthorHashes = { ...expandedAuthorHashes, [authorHash]: true };
		if (hasFetchedAuthorStories(authorHash) || loadingAuthorHashes[authorHash]) return;

		loadingAuthorHashes = { ...loadingAuthorHashes, [authorHash]: true };
		errorByAuthorHash = { ...errorByAuthorHash, [authorHash]: '' };

		try {
			const response = await fetch(
				`/api/tekijat/${encodeURIComponent(authorHash)}/tarinat?type=${encodeURIComponent(filters.type)}`
			);
			if (!response.ok) throw new Error(`Tarinoiden haku epäonnistui (${response.status})`);
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

<svelte:head>
	<title>Tekijät – Texin roistot</title>
</svelte:head>

<section class="tekijat-page">
	<h1>Tekijät</h1>

	<FilterForm
		isLoading={isFilterLoading}
		resetHref="/tekijat"
		totalLabel="{resultLabel(filters.type)} {meta.total}"
		{meta}
		{pageTokens}
		{hasPrev}
		{hasNext}
		{pageHref}
		filterColumns="minmax(220px, 260px) minmax(300px, 1fr) auto"
	>
		<label class="field">
			<span>Ryhmä</span>
			<!-- Changing the group runs the search right away. -->
			<select
				name="type"
				disabled={isFilterLoading}
				on:change={(e) => e.currentTarget.form?.requestSubmit()}
			>
				{#each typeOptions as option (option.value)}
					<option value={option.value} selected={filters.type === option.value}
						>{option.label}</option
					>
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
	</FilterForm>

	{#if authors.length === 0}
		<p class="empty">Ei tuloksia valituilla hakuehdoilla.</p>
	{:else}
		<div class="author-list">
			{#each authors as author (author.hash ?? authorName(author))}
				{@const authorHash = normalizeAuthorHash(author.hash)}
				{@const wikipediaUrl = authorWikipediaUrl(authorName(author))}
				<article class="author-card">
					<h3>
						<button
							type="button"
							class="author-link"
							on:click={() => toggleAuthorStories(author)}
							disabled={!authorHash}
							aria-expanded={Boolean(expandedAuthorHashes[authorHash])}
						>
							{authorName(author)}
						</button>
						{#if wikipediaUrl}
							<a
								class="wiki-link"
								href={wikipediaUrl}
								target="_blank"
								rel="noopener noreferrer"
								on:click={openInOwnWindow}>Wikipedia</a
							>
						{/if}
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
												on:click={() => (selectedStory = story)}
											>
												<StoryTitle title={authorStoryTitle(story, authorHash)} />
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
		<StoryPopup
			story={selectedStory}
			title={storyTitle(selectedStory)}
			publicationSummary={publicationSummary(selectedStory)}
			on:close={() => (selectedStory = null)}
			on:selectStory={(e) => (selectedStory = e.detail)}
		/>
	{/if}

	<Pagination {meta} {pageTokens} {hasPrev} {hasNext} isLoading={isFilterLoading} {pageHref} />
</section>

<style>
	.tekijat-page {
		margin: 1rem 1.5rem 0;
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
		border: 1px solid var(--border);
		background: var(--panel-bg);
		width: 100%;
		box-sizing: border-box;
	}

	/* A compact list: one short row per author, opening to their stories. */
	.author-list {
		display: flex;
		flex-direction: column;
		gap: 0.4rem;
	}

	/* the paper background comes from app.css */
	.author-card {
		border: 1px solid var(--border);
		padding: 0.5rem 0.9rem;
		box-shadow: 0 2px 6px var(--shadow);
	}

	.author-card h3 {
		margin: 0;
		font-size: 1.05rem;
		line-height: 1.4;
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

	.wiki-link {
		margin-left: 0.6rem;
		font-size: 0.85rem;
		font-weight: 400;
	}

	.author-link:disabled {
		text-decoration: none;
		color: var(--muted);
		cursor: default;
	}

	.author-stories {
		border: 1px solid var(--border);
		background: var(--panel-bg);
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

	.author-stories-error {
		margin: 0;
		color: var(--color-error);
	}

	.empty {
		border: 1px solid var(--border);
		padding: 1rem;
		background-color: var(--card-bg);
	}

	@media (max-width: 640px) {
		.tekijat-page {
			margin: 0.75rem 0.75rem 0;
		}
	}
</style>
