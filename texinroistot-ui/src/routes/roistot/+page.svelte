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
	import type { Meta, PaginationToken, StoryBase } from '$lib/listing/shared';
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
		publication: string;
		sort: string;
		q: string;
	};

	const sortOptions = [
		{ value: 'first_name', label: 'Etunimen mukaan' },
		{ value: 'last_name', label: 'Sukunimen mukaan' },
		{ value: 'nickname', label: 'Lempinimen mukaan' },
		{ value: 'other_name', label: 'Etnisen nimen mukaan' },
		{ value: 'rank', label: 'Arvon mukaan' },
		{ value: 'fi_pub_date', label: 'Suomen julkaisupäivän mukaan' },
		{ value: 'it_pub_date', label: 'Alkuperäisessä ilmestymisjärjestyksessä (Italia)' }
	];

	let villains: Villain[] = [];
	let meta: Meta = { total: 0, page: 1, pageSize: 25, totalPages: 0 };
	let filters: Filters = { publication: 'fi', sort: 'fi_pub_date', q: '' };
	let hasPrev = false;
	let hasNext = false;
	let isFilterLoading = false;
	let pageTokens: PaginationToken[] = [];
	let selectedStory: Story | null = null;
	let popupStoryVillainsExpanded = false;
	let popupStoryVillainsLoading = false;
	let popupStoryVillainsError = '';
	let popupStoryVillains: Villain[] = [];
	let popupStoryVillainsLoadedHash = '';

	$: villains = data.villains ?? [];
	$: meta = data.meta ?? { total: 0, page: 1, pageSize: 25, totalPages: 0 };
	$: filters = data.filters ?? { publication: 'fi', sort: 'fi_pub_date', q: '' };
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

	function normalizeStoryHash(raw?: string | null): string {
		return (raw ?? '').trim();
	}

	function storyTitle(story?: Story | null): string {
		if (!story) return '-';
		const publications = story.publications ?? [];
		const baseTitles = nonItalianTitlesByFirstPublication(
			publications.filter((publication) => publication.in?.type === 'perus')
		);
		if (baseTitles.length > 0) return baseTitles[0];

		const nonItalianTitles = nonItalianTitlesByFirstPublication(publications);
		if (nonItalianTitles.length > 0) return nonItalianTitles[0];

		const anyTitle = publications.find((publication) => Boolean(publication.title?.trim()))?.title?.trim();
		return anyTitle || 'Nimetön tarina';
	}

	function cardTitle(story: Story): string {
		const uniqueTitles = nonItalianTitlesByFirstPublication(story.publications);
		return uniqueTitles.length > 0 ? uniqueTitles.join('; ') : 'Nimetön tarina';
	}

	function publicationSummary(story?: Story | null): string {
		if (!story) return '-';
		return publicationSummaryFromPublications(story.publications, 'Ei julkaisutietoja');
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

	function popupVillainNicknames(villain: Villain, storyHash: string): string[] {
		return (storyVillainForStory(villain, storyHash)?.nicknames ?? [])
			.map((nickname) => nickname.trim())
			.filter((nickname, index, values) => Boolean(nickname) && values.indexOf(nickname) === index);
	}

	function popupVillainTitle(villain: Villain, storyHash: string): string {
		const realName = villainRealName(villain);
		const nicknames = popupVillainNicknames(villain, storyHash);
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

	function openStoryPopup(story?: Story | null): void {
		if (!story) return;
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
			popupStoryVillains = payload.villains ?? [];
			popupStoryVillainsLoadedHash = storyHash;
		} catch (error) {
			popupStoryVillainsError =
				error instanceof Error ? error.message : 'Roistojen haku epäonnistui';
		} finally {
			popupStoryVillainsLoading = false;
		}
	}

	function pageHref(page: number): string {
		return buildPageHref('/roistot', {
			publication: filters.publication,
			sort: filters.sort,
			page,
			pageSize: meta.pageSize,
			q: filters.q || undefined
		});
	}
</script>

<svelte:window on:keydown={handleWindowKeydown} />

<section class="roistot-page">
	<h1>Roistot</h1>

	<form method="GET" class="filters">
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

		<input type="hidden" name="publication" value={filters.publication} />
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

	<div class="result-row">
		<p class="result-total">Roistoja yhteensä {meta.total}</p>

		<nav class="pagination pagination-top">
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
		</nav>

		<p class="result-page">
			Sivu {meta.totalPages === 0 ? 0 : meta.page} / {meta.totalPages === 0 ? 0 : meta.totalPages}
		</p>
	</div>

	{#if villains.length === 0}
		<p class="empty">Ei tuloksia valituilla hakuehdoilla.</p>
	{:else}
		<div class="villain-list">
				{#each villains as villain}
					{@const appearance = primaryAppearance(villain)}
					{@const title = villainTitle(villain)}
					{@const displayName = joinValues(appearance?.otherNames, '').trim()}
					{@const codeNames = appearanceCodeNames(appearance)}
					{@const nicknameTitle = title.nicknames.map((nickname) => `"${nickname}"`).join(', ')}
					{@const cardTitle = [title.baseName, nicknameTitle, displayName].filter(Boolean).join(', ')}
					{@const hasCodeNames = codeNames.length > 0}
					<article class="villain-card">
						<h3>{#if cardTitle}{cardTitle}{/if}{#if hasCodeNames}{cardTitle ? ', ' : ''}{#each codeNames as codeName, index}{#if index > 0}, {/if}<em>{codeName}</em>{/each}{:else if !cardTitle}Nimetön roisto{/if}</h3>
						<p><strong>Rooli:</strong> {joinValues(appearance?.roles, '-', '; ')}</p>
						{#if hasValues(appearance?.destiny)}
							<p><strong>Kohtalo:</strong> {joinValues(appearance?.destiny, '-', '; ')}</p>
						{/if}
						{#if hasValues(appearance?.codeNames)}
							<p><strong>Salanimi:</strong> {joinValues(appearance?.codeNames)}</p>
						{/if}
						<p>
							<strong>Tarina:</strong>
							{#if appearance?.story}
								<button type="button" class="story-link" on:click={() => openStoryPopup(appearance.story)}
									>{storyTitle(appearance.story)}</button
								>
							{:else}
								-
							{/if}
						</p>
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
					<h3 id="story-popup-title">{cardTitle(selectedStory)}</h3>
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
									{#each popupStoryVillains as villain}
										{@const appearance = storyVillainForStory(villain, selectedStoryHash)}
										{@const baseTitle = popupVillainTitle(villain, selectedStoryHash)}
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
	</nav>
</section>

<style>
	.roistot-page {
		margin: 1rem 1.5rem 0;
	}

	.filters {
		display: grid;
		grid-template-columns: minmax(300px, 360px) minmax(320px, 1fr) auto;
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
		border: 1px solid black;
		background: #fff;
		color: #111;
		cursor: pointer;
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
		.roistot-page {
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
