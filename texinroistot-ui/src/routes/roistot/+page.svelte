<script lang="ts">
	import { afterNavigate } from '$app/navigation';
	import { navigating } from '$app/stores';
	import {
		buildPageHref,
		firstEditionTitle,
		hasValues,
		joinValues,
		nonItalianTitlesByFirstPublication,
		paginationTokens,
		publicationSummaryFromPublications,
		scrollToListItem,
		storyMentionTargets
	} from '$lib/listing/shared';
	import type { MentionTarget, Meta, PaginationToken } from '$lib/listing/shared';
	import type { Story, StoryVillain, StoryVillainsResponse, Villain } from '$lib/types';
	import FilterForm from '$lib/components/FilterForm.svelte';
	import Pagination from '$lib/components/Pagination.svelte';
	import ItalicText from '$lib/components/ItalicText.svelte';
	import MentionText from '$lib/components/MentionText.svelte';
	import LetterBar from '$lib/components/LetterBar.svelte';
	import StoryPopup from '$lib/components/StoryPopup.svelte';
	import type { PageData } from './$types';

	export let data: PageData;

	type Filters = {
		publication: string;
		sort: string;
		q: string;
	};

	const defaultSortOption = 'default';

	const sortOptions = [
		{ value: 'default', label: 'Oletus' },
		{ value: 'first_name', label: 'Etunimen mukaan' },
		{ value: 'last_name', label: 'Sukunimen mukaan' },
		{ value: 'nickname', label: 'Lempinimen mukaan' },
		{ value: 'other_name', label: 'Etnisen nimen mukaan' },
		{ value: 'code_name', label: 'Salanimen mukaan' },
		{ value: 'rank', label: 'Arvon mukaan' }
	];

	const sortRequirementLabels: Record<string, string> = {
		first_name: 'etunimi',
		last_name: 'sukunimi',
		nickname: 'lempinimi',
		other_name: 'etninen nimi',
		code_name: 'salanimi',
		rank: 'arvo'
	};

	let villains: Villain[] = [];
	let meta: Meta = { total: 0, page: 1, pageSize: 25, totalPages: 0 };
	let filters: Filters = { publication: 'fi', sort: defaultSortOption, q: '' };
	let hasPrev = false;
	let hasNext = false;
	let isFilterLoading = false;
	let pageTokens: PaginationToken[] = [];
	let selectedSort = defaultSortOption;
	let selectedStory: Story | null = null;
	// A villain named in a role or destiny: the story card opens with that
	// villain's card brought into view.
	let focusVillainKey = '';
	let villainList: HTMLElement | undefined;

	// The villains of each story shown with a role or destiny, fetched once per
	// story, so the names of other villains in those texts can be links.
	let storyVillainsByHash: Record<string, Villain[]> = {};
	const requestedStories = new Set<string>();

	async function loadStoryVillains(storyHash: string): Promise<void> {
		if (!storyHash || requestedStories.has(storyHash)) return;
		requestedStories.add(storyHash);
		try {
			const response = await fetch(`/api/tarinat/${encodeURIComponent(storyHash)}/roistot`);
			if (!response.ok) return;
			const payload = (await response.json()) as StoryVillainsResponse;
			storyVillainsByHash = { ...storyVillainsByHash, [storyHash]: payload.villains ?? [] };
		} catch {
			// without the story's villains the texts stay plain
		}
	}

	function mentionTargetsFor(
		byHash: Record<string, Villain[]>,
		storyHash: string
	): MentionTarget[] {
		const storyVillains = byHash[storyHash];
		return storyVillains ? storyMentionTargets(storyVillains, storyHash) : [];
	}

	function openMention(story: Story | undefined, villainKey: string): void {
		if (!story) return;
		focusVillainKey = villainKey;
		selectedStory = story;
	}

	$: for (const villain of villains) {
		const stories = villainStories(villain);
		const appearance = primaryAppearance(villain);
		if (stories.length <= 1 && (hasValues(appearance?.roles) || hasValues(appearance?.destiny))) {
			loadStoryVillains((appearance?.story?.hash ?? '').trim());
		}
	}

	afterNavigate(() => {
		selectedStory = null;
		// After a letter jump, bring the letter's first villain to the top.
		scrollToListItem(villainList, meta.letterIndex);
	});

	$: villains = data.villains ?? [];
	$: meta = data.meta ?? { total: 0, page: 1, pageSize: 25, totalPages: 0 };
	$: filters = data.filters ?? { publication: 'fi', sort: defaultSortOption, q: '' };
	$: selectedSort = sortOptions.some((option) => option.value === filters.sort)
		? filters.sort
		: defaultSortOption;
	$: sortRequirementLabel = sortRequirementLabels[selectedSort] ?? '';
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

	// Distinct code names over all of a villain's stories.
	function villainCodeNames(villain: Villain): string[] {
		return appearanceCodeNames({
			hash: villain.hash,
			codeNames: (villain.as ?? []).flatMap((appearance) => appearance.codeNames ?? [])
		});
	}

	function appearanceCodeNames(appearance?: StoryVillain | null): string[] {
		return (appearance?.codeNames ?? [])
			.map((codeName) => codeName.trim())
			.filter((codeName, index, values) => Boolean(codeName) && values.indexOf(codeName) === index);
	}

	type VillainTitle = {
		rank: string;
		realName: string;
		nicknames: string[];
	};

	// Under the nickname and code name sorts the name the villain is listed
	// under comes first, so "Blue" leads under B and "Ice" under I.
	function listedNameFirst(names: string[], listedName?: string): string[] {
		const listed = (listedName ?? '').trim().toLowerCase();
		const index = names.findIndex((name) => name.trim().toLowerCase() === listed);
		if (!listed || index <= 0) return names;
		return [names[index], ...names.slice(0, index), ...names.slice(index + 1)];
	}

	function villainTitle(villain: Villain): VillainTitle {
		const rank = joinValues(villain.ranks, '').trim();
		const realName = villainRealName(villain);
		const nicknames = villainNicknames(villain);
		return { rank, realName, nicknames };
	}

	// The rank goes in front of the first name shown, whichever it is: "Don
	// Esteban de Queretaro", and "Vanha Juan" for a villain known only by the
	// other name Juan (not "Vanha, Juan").
	function withRank(rank: string, names: string[]): string {
		if (names.length === 0) return rank;
		return [`${rank} ${names[0]}`.trim(), ...names.slice(1)].join(', ');
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
			const storyHash = story.hash.trim();
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

	function storyListTitle(story: Story): string {
		return firstEditionTitle(story.publications) || 'Nimetön tarina';
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

	function letterHref(letter: string): string {
		return buildPageHref('/roistot', {
			publication: filters.publication,
			sort: selectedSort,
			pageSize: meta.pageSize,
			q: filters.q || undefined,
			letter
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
		resetHref="/roistot?sort=default&publication=fi"
		totalLabel="Roistoja yhteensä {meta.total}"
		totalInfo={sortRequirementLabel
			? `Vain roistot, joilla on ${sortRequirementLabel}, näytetään.`
			: ''}
		{meta}
		{pageTokens}
		{hasPrev}
		{hasNext}
		{pageHref}
		filterColumns="minmax(300px, 360px) minmax(320px, 1fr) auto"
	>
		<label class="field">
			<span>Järjestys</span>
			<!-- Changing the sort runs the search right away. -->
			<select
				name="sort"
				disabled={isFilterLoading}
				on:change={(e) => e.currentTarget.form?.requestSubmit()}
			>
				{#each sortOptions as option (option.value)}
					<option value={option.value} selected={selectedSort === option.value}
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
	</FilterForm>

	<details class="sort-info">
		<summary>Tietoa järjestyksistä</summary>
		<ul>
			<li>
				Etunimen mukaan järjestäessä näytetään kaikki ne, joiden etunimi on mainittu tarinassa tai
				joita kutsutaan lempinimellä, jota käytetään kuin etunimeä.
			</li>
			<li>
				Sukunimen mukaan järjestäessä näytetään kaikki ne, joiden sukunimi tai oletettu sukunimi on
				mainittu tarinassa.
			</li>
			<li>
				Lempinimen mukaan järjestäessä näytetään kaikki ne, joiden lempi- tai liikanimi on mainittu
				tarinassa.
			</li>
			<li>
				Etnisen nimen mukaan järjestäessä näytetään ne roistot, joiden nimi ei taivu perinteiseen
				etunimi-sukunimi-muotoon. Tällaisia nimiä ovat esim. intiaanien ja kiinalaisten nimet.
			</li>
			<li>
				Arvo tarkoittaa tässä listauksessa ennen varsinaista nimeä käytettyä ”titteliä”, kuten
				sheriffi, kapteeni ym.
			</li>
		</ul>
	</details>

	{#if meta.letters?.length}
		<LetterBar letters={meta.letters} {letterHref} isLoading={isFilterLoading} />
	{/if}

	{#if villains.length === 0}
		<p class="empty">Ei tuloksia valituilla hakuehdoilla.</p>
	{:else}
		<div class="villain-list" bind:this={villainList}>
			<!-- In the name sorts a villain may be listed once under each of their names. -->
			{#each villains as villain (`${villain.hash}:${villain.sortName ?? ''}`)}
				{@const appearance = primaryAppearance(villain)}
				{@const stories = villainStories(villain)}
				{@const title = villainTitle(villain)}
				{@const displayName = joinValues(appearance?.otherNames, '').trim()}
				{@const codeNames = listedNameFirst(appearanceCodeNames(appearance), villain.sortName)}
				{@const shownCodeNames =
					stories.length <= 1
						? codeNames
						: selectedSort === 'code_name'
							? listedNameFirst(villainCodeNames(villain), villain.sortName)
							: []}
				{@const nicknameTitle = listedNameFirst(title.nicknames, villain.sortName)
					.map((nickname) => `"${nickname}"`)
					.join(', ')}
				{@const cardNames = [title.realName, displayName, nicknameTitle].filter(Boolean)}
				{@const villainCardTitle = cardNames.length > 0 ? withRank(title.rank, cardNames) : ''}
				<article class="villain-card">
					<!-- Code names are shown in the heading only when the villain has no other name. -->
					<h3>
						{#if villainCardTitle}{villainCardTitle}{:else if codeNames.length > 0}{title.rank
								? `${title.rank} `
								: ''}{#each codeNames as codeName, index (codeName)}{index > 0 ? '; ' : ''}<em
									>{codeName}</em
								>{/each}{:else if title.rank}{title.rank}{:else}Nimetön roisto{/if}
					</h3>
					<!-- Role and destiny are story-specific, so they are only shown for villains
					     appearing in a single story. -->
					{#if stories.length <= 1}
						<!-- Names of the story's other villains link to the story card, focused on them. -->
						{@const storyHash = (appearance?.story?.hash ?? '').trim()}
						{@const mentionTargets = mentionTargetsFor(storyVillainsByHash, storyHash)}
						{#if hasValues(appearance?.roles)}
							<p>
								<strong>Rooli:</strong>
								{#if mentionTargets.length > 0}<MentionText
										text={joinValues(appearance?.roles, '-', '; ')}
										targets={mentionTargets}
										selfKey={villain.hash}
										on:mention={(e) => openMention(stories[0], e.detail)}
									/>{:else}<ItalicText text={joinValues(appearance?.roles, '-', '; ')} />{/if}
							</p>
						{/if}
						{#if hasValues(appearance?.destiny)}
							<p>
								<strong>Kohtalo:</strong>
								{#if mentionTargets.length > 0}<MentionText
										text={joinValues(appearance?.destiny, '-', '; ')}
										targets={mentionTargets}
										selfKey={villain.hash}
										on:mention={(e) => openMention(stories[0], e.detail)}
									/>{:else}<ItalicText text={joinValues(appearance?.destiny, '-', '; ')} />{/if}
							</p>
						{/if}
					{/if}
					<!-- In code name order the code names are bold, and shown for villains in
					     several stories too, so it is clear what the villain is listed under. -->
					{#if shownCodeNames.length > 0}
						<p>
							<strong>Salanimi:</strong>
							{#if selectedSort === 'code_name'}<strong>{shownCodeNames.join('; ')}</strong
								>{:else}{shownCodeNames.join('; ')}{/if}
						</p>
					{/if}
					<p>
						<strong>{stories.length === 1 ? 'Tarina' : 'Tarinat'}:</strong>
						{#if stories.length > 0}
							{#each stories as story, index (story.hash)}
								{#if index > 0},
								{/if}
								<button type="button" class="story-link" on:click={() => (selectedStory = story)}
									>{storyListTitle(story)}</button
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
			{focusVillainKey}
			on:close={() => {
				selectedStory = null;
				focusVillainKey = '';
			}}
			on:selectStory={(e) => (selectedStory = e.detail)}
		/>
	{/if}

	<Pagination {meta} {pageTokens} {hasPrev} {hasNext} isLoading={isFilterLoading} {pageHref} />
</section>

<style>
	.roistot-page {
		margin: 1rem 1.5rem 0;
	}

	.sort-info {
		margin: 0.75rem 0;
		border: 1px solid var(--border);
		background: var(--panel-bg);
		padding: 0.5rem 0.75rem;
	}

	.sort-info summary {
		cursor: pointer;
		font-weight: 700;
	}

	.sort-info ul {
		margin: 0.5rem 0 0.25rem;
		padding-left: 1.2rem;
		display: grid;
		gap: 0.35rem;
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

	.villain-list {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	.villain-card {
		border: 1px solid var(--border);
		background-color: var(--card-bg);
		padding: 1rem;
		box-shadow: 0 4px 10px var(--shadow);
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
		border: 1px solid var(--border);
		padding: 1rem;
		background-color: var(--card-bg);
	}

	@media (max-width: 640px) {
		.roistot-page {
			margin: 0.75rem 0.75rem 0;
		}
	}
</style>
