<script lang="ts">
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
		orderNumber: number;
		writtenBy?: Author[] | null;
		drawnBy?: Author[] | null;
		inventedBy?: Author[] | null;
		publications?: StoryPublication[] | null;
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

	const stories: Story[] = data.stories ?? [];
	const meta: Meta = data.meta ?? { total: 0, page: 1, pageSize: 25, totalPages: 0 };
	const filters: Filters = data.filters ?? { publication: 'perus_fi', sort: 'fi_pub_date', q: '' };

	const hasPrev = meta.page > 1;
	const hasNext = meta.page < meta.totalPages;

	function authorList(authors?: Author[] | null): string {
		if (!authors || authors.length === 0) return '-';
		return authors.map((author) => `${author.firstName} ${author.lastName}`.trim()).join(', ');
	}

	function storyTitle(story: Story): string {
		return story.publications?.[0]?.title ?? 'Nimetön tarina';
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

	function pageHref(page: number): string {
		const params = new URLSearchParams();
		params.set('publication', filters.publication);
		params.set('sort', filters.sort);
		params.set('page', String(page));
		params.set('pageSize', String(meta.pageSize));
		if (filters.q) params.set('q', filters.q);
		return `/tarinat?${params.toString()}`;
	}
</script>

<section class="tarinat-page">
	<h1>Tarinat</h1>

	<form method="GET" class="filters">
		<label class="field">
			<span>Julkaisu</span>
			<select name="publication">
				{#each publicationOptions as option}
					<option value={option.value} selected={filters.publication === option.value}
						>{option.label}</option
					>
				{/each}
			</select>
		</label>

		<label class="field">
			<span>Järjestys</span>
			<select name="sort">
				{#each sortOptions as option}
					<option value={option.value} selected={filters.sort === option.value}
						>{option.label}</option
					>
				{/each}
			</select>
		</label>

		<label class="field search">
			<span>Hae hakusanalla</span>
			<input name="q" type="text" value={filters.q ?? ''} placeholder="Kirjoita hakusana..." />
		</label>

		<input type="hidden" name="page" value="1" />
		<input type="hidden" name="pageSize" value={meta.pageSize} />

		<div class="actions">
			<button type="submit">Hae</button>
			<a href="/tarinat">Palauta oletukset</a>
		</div>
	</form>

	<section class="result-header">
		<p>Tarinoita yhteensä {meta.total}</p>
		<p>
			Sivu {meta.totalPages === 0 ? 0 : meta.page} / {meta.totalPages === 0 ? 0 : meta.totalPages}
		</p>
	</section>

	{#if stories.length === 0}
		<p class="empty">Ei tuloksia valituilla hakuehdoilla.</p>
	{:else}
		<div class="story-list">
			{#each stories as story}
				<article class="story-card">
					<h3>#{story.orderNumber} {storyTitle(story)}</h3>
					<p><strong>Kirjoitti:</strong> {authorList(story.writtenBy)}</p>
					<p><strong>Piirsi:</strong> {authorList(story.drawnBy)}</p>
					<p><strong>Ideoi:</strong> {authorList(story.inventedBy)}</p>
					<p><strong>Julkaisut:</strong> {publicationSummary(story)}</p>
				</article>
			{/each}
		</div>
	{/if}

	<nav class="pagination">
		{#if hasPrev}
			<a href={pageHref(meta.page - 1)}>Edellinen</a>
		{:else}
			<span class="disabled">Edellinen</span>
		{/if}

		{#if hasNext}
			<a href={pageHref(meta.page + 1)}>Seuraava</a>
		{:else}
			<span class="disabled">Seuraava</span>
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

	button {
		font-size: 1rem;
		padding: 0.45rem 1rem;
		border: 1px solid black;
		background: black;
		color: white;
		cursor: pointer;
	}

	.result-header {
		display: flex;
		justify-content: space-between;
		margin: 1rem 0;
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

	.pagination {
		margin: 1.25rem 0 0.5rem;
		display: flex;
		justify-content: center;
		gap: 1rem;
	}

	.disabled {
		color: #666;
		text-decoration: none;
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
</style>
