<script lang="ts">
	// Year row for date-ordered lists: a year shows only that year's stories.
	// With many years, decades are shown first ("1980-luku") and open into their
	// years; years without stories are greyed out.
	export let years: number[];
	export let yearHref: (year: number) => string;
	// Link back to the whole list, shown while a year is selected.
	export let allYearsHref: string;
	export let currentYear: number | null = null;
	export let isLoading = false;

	const maxYearsWithoutDecades = 12;

	let openDecade: number | null = null;

	$: available = new Set(years);
	$: decades = [...new Set(years.map((year) => Math.floor(year / 10) * 10))];
	$: useDecades = years.length > maxYearsWithoutDecades;
	// Keep the decade of the selected year open.
	$: if (currentYear !== null) openDecade = Math.floor(currentYear / 10) * 10;
	$: shownYears = useDecades
		? openDecade === null
			? []
			: Array.from({ length: 10 }, (_, i) => (openDecade ?? 0) + i)
		: years;

	function toggleDecade(decade: number): void {
		openDecade = openDecade === decade ? null : decade;
	}
</script>

<nav class="year-bar" aria-label="Valitse vuosi">
	{#if currentYear !== null}
		<div class="row">
			<a href={allYearsHref}>Kaikki</a>
		</div>
	{/if}
	{#if useDecades}
		<div class="row">
			{#each decades as decade (decade)}
				<button
					type="button"
					class="decade"
					class:open={openDecade === decade}
					aria-expanded={openDecade === decade}
					on:click={() => toggleDecade(decade)}>{decade}-luku</button
				>
			{/each}
		</div>
	{/if}
	{#if shownYears.length > 0}
		<div class="row">
			{#each shownYears as year (year)}
				{#if available.has(year) && !isLoading}
					<a href={yearHref(year)} class:current={year === currentYear}>{year}</a>
				{:else}
					<span class="disabled">{year}</span>
				{/if}
			{/each}
		</div>
	{/if}
</nav>

<style>
	.year-bar {
		margin: 0.75rem 0;
		display: grid;
		gap: 0.4rem;
	}

	.row {
		display: flex;
		flex-wrap: wrap;
		justify-content: center;
		gap: 0.35rem 0.75rem;
	}

	.decade {
		border: 0;
		padding: 0;
		background: transparent;
		color: inherit;
		font: inherit;
		text-decoration: underline;
		cursor: pointer;
	}

	.decade.open,
	.current {
		font-weight: 700;
	}

	.disabled {
		color: #999;
	}
</style>
