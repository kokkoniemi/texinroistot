<script lang="ts">
	import { navigating } from '$app/stores';
	import {
		buildPageHref,
		nonItalianTitlesByFirstPublication,
		paginationTokens,
		publicationSummaryFromPublications
	} from '$lib/listing/shared';
	import type { Meta, PaginationToken } from '$lib/listing/shared';
	import { authorStoryList } from '$lib/listing/storyLists';
	import type { AuthorRole, StoryListSpec } from '$lib/listing/storyLists';
	import type { Story, ListedAuthor } from '$lib/types';
	import FilterForm from '$lib/components/FilterForm.svelte';
	import Pagination from '$lib/components/Pagination.svelte';
	import StoryListPopup from '$lib/components/StoryListPopup.svelte';
	import StoryPopup from '$lib/components/StoryPopup.svelte';
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
	// The author card (the author's stories and Wikipedia link) opened from the list.
	let openList: StoryListSpec | null = null;
	// A story chosen from the author card, and the card it was chosen from, so
	// the story popup's "Edellinen" (or the browser's back) returns to the card.
	let selectedStory: Story | null = null;
	let selectedFromList: StoryListSpec | null = null;

	$: authors = data.authors ?? [];
	$: meta = data.meta ?? { total: 0, page: 1, pageSize: 25, totalPages: 0 };
	$: filters = data.filters ?? { type: 'writer', sort: 'last_name', q: '' };
	$: hasPrev = meta.page > 1;
	$: hasNext = meta.page < meta.totalPages;
	$: isFilterLoading = Boolean($navigating) && $navigating?.to?.url.pathname === '/tekijat';
	$: pageTokens = paginationTokens(meta.page, meta.totalPages);
	// The translators in one column; the writers and drawers in two: the first
	// half in the left column, the rest in the right.
	$: singleColumn = filters.type === 'translator';
	$: authorColumns = singleColumn
		? [authors]
		: [
				authors.slice(0, Math.ceil(authors.length / 2)),
				authors.slice(Math.ceil(authors.length / 2))
			];

	function authorName(author: ListedAuthor): string {
		return `${author.firstName} ${author.lastName}`.trim() || '-';
	}

	// The name in the list, surname first: "Bonelli, Giovanni Luigi".
	function listedName(author: ListedAuthor): string {
		const firstName = (author.firstName ?? '').trim();
		const lastName = (author.lastName ?? '').trim();
		if (firstName && lastName) return `${lastName}, ${firstName}`;
		return lastName || firstName || '-';
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

	function authorRole(type: string): AuthorRole {
		return type === 'drawer' || type === 'translator' ? type : 'writer';
	}

	function openAuthor(author: ListedAuthor): void {
		const authorHash = (author.hash ?? '').trim();
		if (!authorHash) return;
		openList = authorStoryList(authorName(author), authorHash, authorRole(filters.type));
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

	function publicationSummary(story: Story): string {
		return publicationSummaryFromPublications(finnishPublications(story), 'Ei julkaisutietoja');
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
		<!-- All authors on one sheet of paper, in two columns (translators in
		     one). A name opens the author card: the author's stories and
		     Wikipedia link. -->
		<div class="author-list author-sheet" class:single-column={singleColumn}>
			{#each authorColumns as column, columnIndex (columnIndex)}
				<ul class="author-column">
					{#each column as author (author.hash ?? authorName(author))}
						<li class="author-entry">
							<button
								type="button"
								class="author-link"
								on:click={() => openAuthor(author)}
								disabled={!(author.hash ?? '').trim()}
							>
								{listedName(author)}
							</button>
						</li>
					{/each}
				</ul>
			{/each}
		</div>
	{/if}

	{#if openList}
		{@const list = openList}
		<StoryListPopup
			{...list}
			on:close={() => (openList = null)}
			on:selectStory={(e) => {
				// remember the card before closing it ({@const} follows openList)
				selectedFromList = list;
				openList = null;
				selectedStory = e.detail;
			}}
		/>
	{/if}

	{#if selectedStory}
		<StoryPopup
			story={selectedStory}
			title={storyTitle(selectedStory)}
			publicationSummary={publicationSummary(selectedStory)}
			fromList={selectedFromList}
			on:close={() => {
				selectedStory = null;
				selectedFromList = null;
			}}
			on:backToList={(e) => {
				selectedStory = null;
				selectedFromList = null;
				openList = e.detail;
			}}
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

	/* All authors on one sheet of paper (the paper comes from app.css), one
	   short row per author, in two columns. */
	.author-list {
		display: grid;
		grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
		border: 1px solid var(--border);
		padding: 1rem 1.25rem;
		box-shadow: 0 4px 10px var(--shadow);
	}

	.author-list.single-column {
		grid-template-columns: minmax(0, 1fr);
	}

	.author-column {
		min-width: 0;
		margin: 0;
		padding: 0 1.25rem 0 0;
		list-style: none;
	}

	/* a thin line between the columns */
	.author-column + .author-column {
		padding: 0 0 0 1.25rem;
		border-left: 1px solid var(--border-soft);
	}

	.author-entry {
		padding: 0.2rem 0;
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

	.author-link:disabled {
		text-decoration: none;
		color: var(--muted);
		cursor: default;
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

		.author-list {
			padding: 0.75rem 0.85rem;
		}

		.author-column {
			padding-right: 0.6rem;
		}

		.author-column + .author-column {
			padding: 0 0 0 0.6rem;
		}

		.author-entry {
			font-size: 0.98rem;
		}
	}
</style>
