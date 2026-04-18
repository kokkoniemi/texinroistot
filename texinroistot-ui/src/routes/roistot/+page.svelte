<script lang="ts">
	import { navigating } from '$app/stores';
	import {
		buildPageHref,
		hasValues,
		joinValues,
		nonItalianTitlesByFirstPublication,
		paginationTokens,
		publicationSummaryFromPublications
	} from '$lib/listing/shared';
	import type { Meta, PaginationToken, StoryBase } from '$lib/listing/shared';
	import FilterForm from '$lib/components/FilterForm.svelte';
	import Pagination from '$lib/components/Pagination.svelte';
	import StoryPopup from '$lib/components/StoryPopup.svelte';
	import type { PageData } from './$types';

	export let data: PageData;

	type Story = StoryBase & {
		hash?: string;
	};

	type StoryVillain = {
		hash?: string;
		nicknames?: string[] | null;
		otherNames?: string[] | null;
		codeNames?: string[] | null;
		roles?: string[] | null;
		destiny?: string[] | null;
		story?: Story | null;
	};

	type Villain = {
		hash?: string;
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

	const defaultSortOption = 'first_name';

	const sortOptions = [
		{ value: 'first_name', label: 'Etunimen mukaan' },
		{ value: 'last_name', label: 'Sukunimen mukaan' },
		{ value: 'nickname', label: 'Lempinimen mukaan' },
		{ value: 'other_name', label: 'Etnisen nimen mukaan' },
		{ value: 'rank', label: 'Arvon mukaan' }
	];

	let villains: Villain[] = [];
	let meta: Meta = { total: 0, page: 1, pageSize: 25, totalPages: 0 };
	let filters: Filters = { publication: 'fi', sort: defaultSortOption, q: '' };
	let hasPrev = false;
	let hasNext = false;
	let isFilterLoading = false;
	let pageTokens: PaginationToken[] = [];
	let selectedSort = defaultSortOption;
	let selectedStory: Story | null = null;

	$: villains = data.villains ?? [];
	$: meta = data.meta ?? { total: 0, page: 1, pageSize: 25, totalPages: 0 };
	$: filters = data.filters ?? { publication: 'fi', sort: defaultSortOption, q: '' };
	$: selectedSort = sortOptions.some((option) => option.value === filters.sort)
		? filters.sort
		: defaultSortOption;
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

	function villainStories(villain: Villain): Story[] {
		const stories: Story[] = [];
		const seenHashes = new Set<string>();
		for (const appearance of villain.as ?? []) {
			const story = appearance.story;
			if (!story) continue;
			const storyHash = (story.hash ?? '').trim();
			if (storyHash && seenHashes.has(storyHash)) continue;
			if (storyHash) seenHashes.add(storyHash);
			stories.push(story);
		}
		return stories;
	}

	function storyCardTitle(story: Story): string {
		const uniqueTitles = nonItalianTitlesByFirstPublication(story.publications);
		return uniqueTitles.length > 0 ? uniqueTitles.join('; ') : 'Nimetön tarina';
	}

	function publicationSummary(story?: Story | null): string {
		if (!story) return '-';
		return publicationSummaryFromPublications(story.publications, 'Ei julkaisutietoja');
	}

	function pageHref(page: number): string {
		return buildPageHref('/roistot', {
			publication: filters.publication,
			sort: selectedSort,
			page,
			pageSize: meta.pageSize,
			q: filters.q || undefined
		});
	}
</script>

<svelte:head>
	<title>Roistot – Texin roistot</title>
</svelte:head>

<section class="roistot-page">
	<h1>Roistot</h1>

	<FilterForm
		isLoading={isFilterLoading}
		resetHref="/roistot?sort=first_name&publication=fi"
		totalLabel="Roistoja yhteensä {meta.total}"
		{meta}
		{pageTokens}
		{hasPrev}
		{hasNext}
		{pageHref}
		filterColumns="minmax(300px, 360px) minmax(320px, 1fr) auto"
	>
		<label class="field">
			<span>Järjestys</span>
			<select name="sort" disabled={isFilterLoading}>
				{#each sortOptions as option (option.value)}
					<option value={option.value} selected={selectedSort === option.value}>{option.label}</option>
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
	</FilterForm>

	{#if villains.length === 0}
		<p class="empty">Ei tuloksia valituilla hakuehdoilla.</p>
	{:else}
		<div class="villain-list">
			{#each villains as villain, i (villain.hash ?? i)}
				{@const appearance = primaryAppearance(villain)}
				{@const stories = villainStories(villain)}
				{@const title = villainTitle(villain)}
				{@const displayName = joinValues(appearance?.otherNames, '').trim()}
				{@const codeNames = appearanceCodeNames(appearance)}
				{@const nicknameTitle = title.nicknames.map((nickname) => `"${nickname}"`).join(', ')}
				{@const villainCardTitle = [title.baseName, nicknameTitle, displayName].filter(Boolean).join(', ')}
				{@const hasCodeNames = codeNames.length > 0}
				<article class="villain-card">
					<h3>{#if villainCardTitle}{villainCardTitle}{/if}{#if hasCodeNames}{villainCardTitle ? ', ' : ''}{#each codeNames as codeName, index (codeName)}{#if index > 0}, {/if}<em>{codeName}</em>{/each}{:else if !villainCardTitle}Nimetön roisto{/if}</h3>
					{#if hasValues(appearance?.roles)}
						<p><strong>Rooli:</strong> {joinValues(appearance?.roles, '-', '; ')}</p>
					{/if}
					{#if hasValues(appearance?.destiny)}
						<p><strong>Kohtalo:</strong> {joinValues(appearance?.destiny, '-', '; ')}</p>
					{/if}
					{#if hasValues(appearance?.codeNames)}
						<p><strong>Salanimi:</strong> {joinValues(appearance?.codeNames)}</p>
					{/if}
					<p>
						<strong>Tarinat:</strong>
						{#if stories.length > 0}
							{#each stories as story, index (story.hash ?? index)}
								{#if index > 0}, {/if}
								<button type="button" class="story-link" on:click={() => (selectedStory = story)}
									>{storyCardTitle(story)}</button
								>
							{/each}
						{:else}
							-
						{/if}
					</p>
				</article>
			{/each}
		</div>
	{/if}

	{#if selectedStory}
		<StoryPopup
			story={selectedStory}
			title={storyCardTitle(selectedStory)}
			publicationSummary={publicationSummary(selectedStory)}
			on:close={() => (selectedStory = null)}
		/>
	{/if}

	<Pagination {meta} {pageTokens} {hasPrev} {hasNext} isLoading={isFilterLoading} {pageHref} />
</section>

<style>
	.roistot-page {
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

	.story-link {
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

	.empty {
		border: 1px solid black;
		padding: 1rem;
		background-color: #f7f7f7;
	}

	@media (max-width: 640px) {
		.roistot-page {
			margin: 0.75rem 0.75rem 0;
		}
	}
</style>
