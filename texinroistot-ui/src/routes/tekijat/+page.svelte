<script lang="ts">
	import { navigating } from '$app/stores';
	import {
		authorList,
		buildPageHref,
		joinValues,
		nonItalianTitlesByFirstPublication,
		paginationTokens,
		publicationSummaryFromPublications
	} from '$lib/listing/shared';
	import type { Author, Meta, PaginationToken, StoryBase } from '$lib/listing/shared';
	import FilterForm from '$lib/components/FilterForm.svelte';
	import Pagination from '$lib/components/Pagination.svelte';
	import StoryPopup from '$lib/components/StoryPopup.svelte';
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
	</FilterForm>

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
												on:click={() => (selectedStory = story)}
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
		<StoryPopup
			story={selectedStory}
			title={storyTitle(selectedStory)}
			publicationSummary={publicationSummary(selectedStory)}
			on:close={() => (selectedStory = null)}
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
		border: 1px solid black;
		background: #fff;
		width: 100%;
		box-sizing: border-box;
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

	.author-stories-error {
		margin: 0;
		color: var(--color-error);
	}

	.empty {
		border: 1px solid black;
		padding: 1rem;
		background-color: #f7f7f7;
	}

	@media (max-width: 640px) {
		.tekijat-page {
			margin: 0.75rem 0.75rem 0;
		}
	}
</style>
