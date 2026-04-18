<script lang="ts">
	import { navigating } from '$app/stores';
	import {
		authorList,
		buildPageHref,
		hasValues,
		joinValues,
		nonItalianTitlesByFirstPublication,
		paginationTokens,
		publicationSummaryFromPublications,
		italianOriginalPublication,
		storyVillainForStory,
		storyVillainTitle
	} from '$lib/listing/shared';
	import type { Meta, PaginationToken } from '$lib/listing/shared';
	import type { Story, Villain, StoryVillainsResponse } from '$lib/types';
	import FilterForm from '$lib/components/FilterForm.svelte';
	import Pagination from '$lib/components/Pagination.svelte';
	import VillainCard from '$lib/components/VillainCard.svelte';
	import type { PageData } from './$types';

	export let data: PageData;

	type Filters = {
		publication: string;
		sort: string;
		q: string;
		year: number;
	};

	const publicationOptions = [
		{ value: 'all', label: 'Näytä kaikki' },
		{ value: 'perus_fi', label: 'Suomen perussarja' },
		{ value: 'perus_it', label: 'Italian perussarja' },
		{ value: 'suur', label: 'Suuralbumit' },
		{ value: 'maxi', label: 'Maxi-Tex' },
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

	const sortOptions = [
		{ value: 'fi_pub_date', label: 'Suomen julkaisupäivän mukaan' },
		{ value: 'it_pub_date', label: 'Alkuperäisessä ilmestymisjärjestyksessä (Italia)' },
		{ value: 'alpha', label: 'Aakkosjärjestyksessä' }
	];

	let stories: Story[] = [];
	let meta: Meta = { total: 0, page: 1, pageSize: 25, totalPages: 0 };
	let filters: Filters = { publication: 'perus_fi', sort: 'fi_pub_date', q: '', year: 0 };
	let hasPrev = false;
	let hasNext = false;
	let expandedStoryHashes: Record<string, boolean> = {};
	let loadingStoryHashes: Record<string, boolean> = {};
	let errorByStoryHash: Record<string, string> = {};
	let villainsByStoryHash: Record<string, Villain[]> = {};
	let storyListSignature = '';
	let isFilterLoading = false;
	let pageTokens: PaginationToken[] = [];

	$: stories = data.stories ?? [];
	$: meta = data.meta ?? { total: 0, page: 1, pageSize: 25, totalPages: 0 };
	$: filters = data.filters ?? { publication: 'perus_fi', sort: 'fi_pub_date', q: '', year: 0 };
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

	async function toggleStoryVillains(storyHash: string): Promise<void> {
		if (!storyHash) return;
		const isCurrentlyExpanded = Boolean(expandedStoryHashes[storyHash]);

		if (isCurrentlyExpanded) {
			expandedStoryHashes = { ...expandedStoryHashes, [storyHash]: false };
			return;
		}

		expandedStoryHashes = { ...expandedStoryHashes, [storyHash]: true };
		if (hasFetchedStoryVillains(storyHash) || loadingStoryHashes[storyHash]) return;

		loadingStoryHashes = { ...loadingStoryHashes, [storyHash]: true };
		errorByStoryHash = { ...errorByStoryHash, [storyHash]: '' };

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
			page,
			pageSize: meta.pageSize,
			q: filters.q || undefined,
			year: filters.year > 0 ? filters.year : undefined
		});
	}
</script>

<svelte:head>
	<title>Tarinat – Texin roistot</title>
</svelte:head>

<section class="tarinat-page">
	<h1>Tarinat</h1>

	<FilterForm
		isLoading={isFilterLoading}
		resetHref="/tarinat"
		totalLabel="Tarinoita yhteensä {meta.total}"
		{meta}
		{pageTokens}
		{hasPrev}
		{hasNext}
		{pageHref}
		filterColumns="minmax(150px, 220px) minmax(200px, 320px) minmax(120px, 150px) minmax(260px, 1fr) auto"
		filterColumns1500="minmax(150px, 1fr) minmax(200px, 1.35fr) minmax(120px, 0.7fr) minmax(230px, 1.4fr)"
		filterColumns1200="minmax(150px, 1fr) minmax(200px, 1fr)"
	>
		<label class="field">
			<span>Julkaisu</span>
			<select name="publication" disabled={isFilterLoading}>
				{#each publicationOptions as option (option.value)}
					<option value={option.value} selected={filters.publication === option.value}
						>{option.label}</option
					>
				{/each}
			</select>
		</label>

		<label class="field">
			<span>Järjestys</span>
			<select name="sort" disabled={isFilterLoading}>
				{#each sortOptions as option (option.value)}
					<option value={option.value} selected={filters.sort === option.value}>{option.label}</option>
				{/each}
			</select>
		</label>

		<label class="field year">
			<span>Vuosi</span>
			<input
				name="year"
				type="number"
				min="1"
				step="1"
				value={filters.year > 0 ? String(filters.year) : ''}
				placeholder="esim. 1980"
				disabled={isFilterLoading}
			/>
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
	</FilterForm>

	{#if stories.length === 0}
		<p class="empty">Ei tuloksia valituilla hakuehdoilla.</p>
	{:else}
		<div class="story-list">
			{#each stories as story (story.hash)}
				{@const storyHash = story.hash.trim()}
				{@const italianOriginal = italianOriginalPublication(story)}
				<article class="story-card">
					<h3>{storyTitle(story)}</h3>
					<p><strong>Kertoi:</strong> {authorList(story.writtenBy, '; ')}</p>
					<p><strong>Piirsi:</strong> {authorList(story.drawnBy, '; ')}</p>
					<p><strong>Suomensi:</strong> {authorList(story.translatedBy, '; ')}</p>
					{#if italianOriginal}
						<p>
							<strong>Alkuperäisjulkaisu (Italia):</strong>
							{#if italianOriginal.title}<em>{italianOriginal.title}</em>{/if}
							{#if italianOriginal.details}{italianOriginal.title ? ', ' : ''}{italianOriginal.details}{/if}
						</p>
					{/if}
					<p><strong>Ilmestynyt Suomessa:</strong> {publicationSummary(story)}</p>

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
									{#each storyVillains(storyHash) as villain (villain.hash)}
										{@const appearance = storyVillainForStory(villain, storyHash)}
										{@const baseTitle = storyVillainTitle(villain, storyHash)}
										{@const displayName = joinValues(appearance?.otherNames, '').trim()}
										{@const cardTitle =
											displayName && baseTitle === 'Nimetön roisto'
												? displayName
												: displayName
													? `${baseTitle}, ${displayName}`
													: baseTitle}
										<VillainCard
											title={cardTitle}
											ranks={villain.ranks}
											roles={appearance?.roles}
											destiny={appearance?.destiny}
											codeNames={appearance?.codeNames}
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

	<Pagination {meta} {pageTokens} {hasPrev} {hasNext} isLoading={isFilterLoading} {pageHref} />
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
		border: 1px solid black;
		background: #fff;
		width: 100%;
		box-sizing: border-box;
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

	.villain-error {
		color: var(--color-error);
	}

	.empty {
		border: 1px solid black;
		padding: 1rem;
		background-color: #f7f7f7;
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
