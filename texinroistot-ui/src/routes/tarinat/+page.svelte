<script lang="ts">
	import { setContext, tick } from 'svelte';
	import { afterNavigate } from '$app/navigation';
	import { navigating } from '$app/stores';
	import {
		buildPageHref,
		nonItalianTitlesByFirstPublication,
		paginationTokens,
		publicationSummaryFromPublications,
		scrollToListItem,
		italianOriginalPublication,
		storyVillainForStory,
		otherStoriesVillainHash,
		storyVillainTitle,
		storyVillainCodeNameTitle,
		storyMentionTargets,
		showMentionedVillain
	} from '$lib/listing/shared';
	import type { Meta, PaginationToken } from '$lib/listing/shared';
	import type { Story, Villain, StoryVillainsResponse } from '$lib/types';
	import FilterForm from '$lib/components/FilterForm.svelte';
	import AuthorLinks from '$lib/components/AuthorLinks.svelte';
	import LetterBar from '$lib/components/LetterBar.svelte';
	import YearBar from '$lib/components/YearBar.svelte';
	import Pagination from '$lib/components/Pagination.svelte';
	import StoryPopup from '$lib/components/StoryPopup.svelte';
	import StoryListPopup from '$lib/components/StoryListPopup.svelte';
	import StoryWikiLink from '$lib/components/StoryWikiLink.svelte';
	import { REMEMBER_STORY_LIST } from '$lib/listing/storyLists';
	import type { RememberStoryList, StoryListSpec } from '$lib/listing/storyLists';
	import StoryTitle from '$lib/components/StoryTitle.svelte';
	import VillainCard from '$lib/components/VillainCard.svelte';
	import type { PageData } from './$types';

	export let data: PageData;

	type Filters = {
		publication: string;
		sort: string;
		q: string;
		year: number;
		// The random sort's seed: page links keep it, so paging stays in the same
		// shuffle; a new visit (no seed) gets a new one from the backend.
		seed?: number;
	};
	const defaultFilters: Filters = { publication: 'all', sort: 'random', q: '', year: 0 };

	const publicationOptions = [
		{ value: 'all', label: 'Näytä kaikki' },
		{ value: 'perus_fi', label: 'Suomen perussarja' },
		{ value: 'perus_it', label: 'Italian perussarja' },
		{ value: 'suur', label: 'Suuralbumit' },
		{ value: 'maxi', label: 'Maxi Tex' },
		{ value: 'kirjasto', label: 'Kirjasto' },
		{ value: 'kronikka', label: 'Kronikka' },
		{ value: 'serie_extra', label: 'Serie extra' },
		{ value: 'texone', label: 'Texone' },
		{ value: 'mini_texone_maxi_tex', label: 'Mini Texone & Maxi Tex' },
		{ value: 'almanacco_del_west', label: 'Almanacco del West' },
		{ value: 'color_tex', label: 'Color Tex' },
		{ value: 'tex_romanzi_a_fumetti', label: 'Tex romanzi a fumetti' },
		{ value: 'tex_magazine', label: 'Tex Magazine' }
	];

	const baseSortOptions = [
		{ value: 'random', label: 'Sattumanvarainen' },
		{ value: 'fi_pub_date', label: 'Ilmestymisjärjestyksessä (Suomi)' },
		{ value: 'it_pub_date', label: 'Alkuperäisessä ilmestymisjärjestyksessä (Italia)' },
		{ value: 'alpha', label: 'Aakkosjärjestyksessä' }
	];
	// For the Italian publications (base series and specials) both alphabetical
	// sorts are offered, so they name their language.
	const italianSortOptions = [
		...baseSortOptions.filter((option) => option.value !== 'alpha'),
		{ value: 'alpha', label: 'Aakkosjärjestyksessä (suomi)' },
		{ value: 'alpha_it', label: 'Aakkosjärjestyksessä (italia)' }
	];
	const italianPublications = [
		'perus_it',
		'serie_extra',
		'texone',
		'mini_texone_maxi_tex',
		'almanacco_del_west',
		'color_tex',
		'tex_romanzi_a_fumetti',
		'tex_magazine'
	];

	let stories: Story[] = [];
	let meta: Meta = { total: 0, page: 1, pageSize: 25, totalPages: 0 };
	let filters: Filters = defaultFilters;
	let hasPrev = false;
	let hasNext = false;
	let expandedStoryHashes: Record<string, boolean> = {};
	// Story chosen from an author popup (Kertoi/Piirsi), shown in a story popup.
	let selectedStory: Story | null = null;
	// The story list (an author's or a villain's) the story was chosen from, so
	// the story popup's "Edellinen" (or the browser's back) returns to it.
	let chosenFromList: StoryListSpec | null = null;
	let selectedFromList: StoryListSpec | null = null;
	let reopenedList: StoryListSpec | null = null;
	setContext<RememberStoryList>(REMEMBER_STORY_LIST, (list) => (chosenFromList = list));

	function openStory(story: Story): void {
		selectedFromList = chosenFromList;
		chosenFromList = null;
		selectedStory = story;
	}
	let storyList: HTMLElement | undefined;

	// After a letter jump, bring the letter's first story to the top.
	afterNavigate(() => scrollToListItem(storyList, meta.letterIndex));
	let loadingStoryHashes: Record<string, boolean> = {};
	let errorByStoryHash: Record<string, string> = {};
	let villainsByStoryHash: Record<string, Villain[]> = {};
	let storyListSignature = '';
	let isFilterLoading = false;
	let pageTokens: PaginationToken[] = [];

	$: stories = data.stories ?? [];
	$: meta = data.meta ?? { total: 0, page: 1, pageSize: 25, totalPages: 0 };
	$: filters = data.filters ?? defaultFilters;
	$: seed = filters.sort === 'random' && filters.seed ? filters.seed : undefined;
	$: sortOptions =
		italianPublications.includes(filters.publication) || filters.sort === 'alpha_it'
			? italianSortOptions
			: baseSortOptions;
	$: hasPrev = meta.page > 1;
	$: hasNext = meta.page < meta.totalPages;
	$: isFilterLoading = Boolean($navigating) && $navigating?.to?.url.pathname === '/tarinat';
	$: pageTokens = paginationTokens(meta.page, meta.totalPages);
	$: {
		const nextSignature = stories.map((story) => story.hash.trim()).join('|');
		if (nextSignature !== storyListSignature) {
			storyListSignature = nextSignature;
			expandedStoryHashes = {};
			loadingStoryHashes = {};
			errorByStoryHash = {};
			villainsByStoryHash = {};
		}
	}

	function storyTitle(story: Story): string {
		const uniqueTitles = nonItalianTitlesByFirstPublication(story.publications);
		return uniqueTitles.length > 0 ? uniqueTitles.join('; ') : 'Nimetön tarina';
	}

	function publicationSummary(story: Story): string {
		return publicationSummaryFromPublications(story.publications, 'Ei julkaisutietoja');
	}

	function storyVillains(storyHash: string): Villain[] {
		return villainsByStoryHash[storyHash] ?? [];
	}

	function hasFetchedStoryVillains(storyHash: string): boolean {
		return Object.prototype.hasOwnProperty.call(villainsByStoryHash, storyHash);
	}

	// Only one story's villains are open at a time: opening one closes the
	// previous. Closing a list above the button would shift the page, so the
	// clicked button is kept where it was on screen.
	async function toggleStoryVillains(storyHash: string, button?: HTMLElement): Promise<void> {
		if (!storyHash) return;
		const isCurrentlyExpanded = Boolean(expandedStoryHashes[storyHash]);

		if (isCurrentlyExpanded) {
			expandedStoryHashes = { ...expandedStoryHashes, [storyHash]: false };
			return;
		}

		const topBefore = button?.getBoundingClientRect().top;
		const needsFetch = !hasFetchedStoryVillains(storyHash) && !loadingStoryHashes[storyHash];
		// Mark the fetch as started before the list renders, so an opened list
		// never shows "no villains" while they are still being fetched.
		if (needsFetch) {
			loadingStoryHashes = { ...loadingStoryHashes, [storyHash]: true };
			errorByStoryHash = { ...errorByStoryHash, [storyHash]: '' };
		}
		expandedStoryHashes = { [storyHash]: true };
		if (button && topBefore !== undefined) {
			await tick();
			const shift = button.getBoundingClientRect().top - topBefore;
			if (shift !== 0) window.scrollBy(0, shift);
		}
		if (!needsFetch) return;

		try {
			const response = await fetch(`/api/tarinat/${encodeURIComponent(storyHash)}/roistot`);
			if (!response.ok) throw new Error(`Roistojen haku epäonnistui (${response.status})`);
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
		return buildPageHref('/tarinat', {
			publication: filters.publication,
			sort: filters.sort,
			seed,
			page,
			pageSize: meta.pageSize,
			q: filters.q || undefined,
			year: filters.year > 0 ? filters.year : undefined
		});
	}

	function letterHref(letter: string): string {
		return buildPageHref('/tarinat', {
			publication: filters.publication,
			sort: filters.sort,
			pageSize: meta.pageSize,
			q: filters.q || undefined,
			year: filters.year > 0 ? filters.year : undefined,
			letter
		});
	}

	// The year row filters the list to one year (shown whole, without page
	// numbers); a search word is kept unless it was itself a year.
	function yearHref(year: number | null): string {
		const q = filters.q?.trim() ?? '';
		return buildPageHref('/tarinat', {
			publication: filters.publication,
			sort: filters.sort,
			seed,
			q: q && !/^\d{4}$/.test(q) ? q : undefined,
			year: year ?? undefined
		});
	}

	$: selectedYear = filters.year > 0 ? filters.year : null;
</script>

<svelte:head>
	<title>Tarinat – Texin roistot</title>
</svelte:head>

<section class="tarinat-page">
	<h1>Tarinat</h1>

	<FilterForm
		isLoading={isFilterLoading}
		showPagination={!selectedYear}
		resetHref="/tarinat"
		totalLabel="Tarinoita yhteensä {meta.total}"
		{meta}
		{pageTokens}
		{hasPrev}
		{hasNext}
		{pageHref}
		filterColumns="minmax(150px, 220px) minmax(200px, 320px) minmax(260px, 1fr) auto"
		filterColumns1500="minmax(150px, 1fr) minmax(200px, 1.35fr) minmax(230px, 1.4fr)"
		filterColumns1200="minmax(150px, 1fr) minmax(200px, 1fr)"
	>
		<label class="field">
			<span>Julkaisu</span>
			<!-- Changing the publication or the sort runs the search right away. -->
			<select
				name="publication"
				disabled={isFilterLoading}
				on:change={(e) => e.currentTarget.form?.requestSubmit()}
			>
				{#each publicationOptions as option (option.value)}
					<option value={option.value} selected={filters.publication === option.value}
						>{option.label}</option
					>
				{/each}
			</select>
		</label>

		<label class="field">
			<span>Järjestys</span>
			<select
				name="sort"
				disabled={isFilterLoading}
				on:change={(e) => e.currentTarget.form?.requestSubmit()}
			>
				{#each sortOptions as option (option.value)}
					<option value={option.value} selected={filters.sort === option.value}
						>{option.label}</option
					>
				{/each}
			</select>
		</label>

		<!-- A year as the search word (e.g. 1988) lists that year's stories. -->
		<label class="field search">
			<span>Hae hakusanalla</span>
			<input
				name="q"
				type="text"
				value={filters.q ?? ''}
				placeholder="esim. väijytys tai 1988"
				disabled={isFilterLoading}
			/>
		</label>

		<input type="hidden" name="page" value="1" />
		<input type="hidden" name="pageSize" value={meta.pageSize} />
	</FilterForm>

	{#if meta.letters?.length}
		<LetterBar letters={meta.letters} {letterHref} isLoading={isFilterLoading} />
	{/if}
	{#if meta.years?.length}
		<YearBar
			years={meta.years}
			{yearHref}
			allYearsHref={yearHref(null)}
			currentYear={selectedYear}
			isLoading={isFilterLoading}
		/>
	{/if}

	{#if stories.length === 0}
		<p class="empty">Ei tuloksia valituilla hakuehdoilla.</p>
	{:else}
		<div class="story-list" bind:this={storyList}>
			{#each stories as story (story.hash)}
				{@const storyHash = story.hash.trim()}
				{@const italianOriginal = italianOriginalPublication(story)}
				<article class="story-card">
					<h3><StoryTitle title={storyTitle(story)} /></h3>
					<p>
						<strong>Kertoi:</strong>
						<AuthorLinks
							authors={story.writtenBy}
							type="writer"
							on:selectStory={(e) => openStory(e.detail)}
						/>
					</p>
					<p>
						<strong>Piirsi:</strong>
						<AuthorLinks
							authors={story.drawnBy}
							type="drawer"
							on:selectStory={(e) => openStory(e.detail)}
						/>
					</p>
					<p>
						<strong>Suomensi:</strong>
						<AuthorLinks
							authors={story.translatedBy}
							type="translator"
							on:selectStory={(e) => openStory(e.detail)}
						/>
					</p>
					{#if italianOriginal}
						<p>
							<strong>Alkuperäisjulkaisu (Italia):</strong>
							{#if italianOriginal.title}<em>{italianOriginal.title}</em>{italianOriginal.details
									? ','
									: ''}{/if}
							{italianOriginal.details}
						</p>
					{/if}
					<p><strong>Ilmestynyt Suomessa:</strong> {publicationSummary(story)}</p>

					<button
						type="button"
						class="toggle-villains"
						on:click={(event) => toggleStoryVillains(storyHash, event.currentTarget)}
						disabled={!storyHash}
						aria-expanded={Boolean(expandedStoryHashes[storyHash])}
					>
						{#if expandedStoryHashes[storyHash]}
							Piilota tarinan roistot
						{:else}
							Näytä tarinan roistot
						{/if}
					</button>

					{#if expandedStoryHashes[storyHash]}
						<section class="story-villains">
							{#if errorByStoryHash[storyHash]}
								<p class="villain-error">{errorByStoryHash[storyHash]}</p>
							{:else if loadingStoryHashes[storyHash] || !(storyHash in villainsByStoryHash)}
								<p>Haetaan roistoja...</p>
							{:else if (villainsByStoryHash[storyHash] ?? []).length === 0}
								<p>Tarinalle ei löytynyt roistoja.</p>
							{:else}
								<!-- Names of the story's villains in roles and destinies link to their cards. -->
								{@const mentionTargets = storyMentionTargets(storyVillains(storyHash), storyHash)}
								<StoryWikiLink {story} />
								<div class="story-villains-list" data-story-villains={storyHash}>
									{#each storyVillains(storyHash) as villain (villain.hash)}
										{@const appearance = storyVillainForStory(villain, storyHash)}
										{@const cardTitle = storyVillainTitle(villain, storyHash)}
										<VillainCard
											title={cardTitle}
											codeNameTitle={storyVillainCodeNameTitle(villain, storyHash)}
											otherStoriesVillainHash={otherStoriesVillainHash(villain)}
											roles={appearance?.roles}
											destiny={appearance?.destiny}
											codeNames={appearance?.codeNames}
											villainKey={villain.hash}
											{villain}
											{mentionTargets}
											on:mention={(e) =>
												showMentionedVillain(
													document.querySelector<HTMLElement>(
														`[data-story-villains="${CSS.escape(storyHash)}"]`
													) ?? undefined,
													e.detail
												)}
											on:selectStory={(e) => openStory(e.detail)}
										/>
									{/each}
								</div>
							{/if}
						</section>
					{/if}
				</article>
			{/each}
		</div>
	{/if}

	{#if !selectedYear}
		<Pagination {meta} {pageTokens} {hasPrev} {hasNext} isLoading={isFilterLoading} {pageHref} />
	{/if}

	{#if selectedStory}
		<StoryPopup
			story={selectedStory}
			title={storyTitle(selectedStory)}
			publicationSummary={publicationSummary(selectedStory)}
			fromList={selectedFromList}
			on:close={() => (selectedStory = null)}
			on:backToList={(e) => {
				selectedStory = null;
				reopenedList = e.detail;
			}}
			on:selectStory={(e) => (selectedStory = e.detail)}
		/>
	{/if}

	<!-- The list a story was chosen from, reopened by the story popup's "Edellinen". -->
	{#if reopenedList}
		{@const list = reopenedList}
		<StoryListPopup
			{...list}
			on:close={() => (reopenedList = null)}
			on:selectStory={(e) => {
				// remember the list before closing it ({@const} follows reopenedList)
				chosenFromList = list;
				reopenedList = null;
				openStory(e.detail);
			}}
		/>
	{/if}
</section>

<style>
	.tarinat-page {
		margin: 1rem 1.5rem 0;
	}

	.field {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
		min-width: 0;
	}

	.field span {
		font-size: 0.95rem;
	}

	select,
	input {
		font-size: 1rem;
		padding: 0.45rem 0.5rem;
		border: 1px solid var(--border);
		background: var(--panel-bg);
		width: 100%;
		box-sizing: border-box;
	}

	.story-list {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	/* the paper background comes from app.css */
	.story-card {
		border: 1px solid var(--border);
		padding: 1rem;
		box-shadow: 0 4px 10px var(--shadow);
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
		border: 1px solid var(--border);
		background: var(--panel-bg);
		color: var(--ink);
		cursor: pointer;
	}

	.toggle-villains:disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}

	.story-villains {
		margin-top: 0.8rem;
		padding-top: 0.75rem;
		border-top: 1px solid var(--border);
	}

	.story-villains-list {
		display: flex;
		flex-direction: column;
		gap: 0.6rem;
	}

	.villain-error {
		color: var(--color-error);
	}

	.empty {
		border: 1px solid var(--border);
		padding: 1rem;
		background-color: var(--card-bg);
	}

	@media (max-width: 1200px) {
		.field.search {
			grid-column: 1 / -1;
		}
	}

	@media (max-width: 640px) {
		.tarinat-page {
			margin: 0.75rem 0.75rem 0;
		}
	}
</style>
