<script lang="ts">
	import type { Meta, PaginationToken } from '$lib/listing/shared';
	import Pagination from './Pagination.svelte';

	export let isLoading: boolean = false;
	export let resetHref: string;
	export let totalLabel: string;
	export let totalInfo: string = '';
	export let meta: Meta;
	export let pageTokens: PaginationToken[];
	export let hasPrev: boolean;
	export let hasNext: boolean;
	export let pageHref: (page: number) => string;
	export let filterColumns: string = '1fr';
	export let filterColumns1500: string = '';
	export let filterColumns1200: string = '';
	// Hide the page numbers when the whole result is on one page (e.g. a single year).
	export let showPagination = true;

	let showInfo = false;
	let infoWrapper: HTMLSpanElement | null = null;

	$: formStyle = [
		`--filters-columns: ${filterColumns}`,
		filterColumns1500 && `--filters-columns-1500: ${filterColumns1500}`,
		filterColumns1200 && `--filters-columns-1200: ${filterColumns1200}`
	]
		.filter(Boolean)
		.join('; ');

	function handleDocumentClick(event: MouseEvent) {
		if (!showInfo) return;
		const target = event.target as Node | null;
		if (infoWrapper && target && !infoWrapper.contains(target)) {
			showInfo = false;
		}
	}

	function handleDocumentKey(event: KeyboardEvent) {
		if (event.key === 'Escape' && showInfo) {
			showInfo = false;
		}
	}
</script>

<svelte:window on:click={handleDocumentClick} on:keydown={handleDocumentKey} />

<form method="GET" class="filters" style={formStyle}>
	<slot />
	<div class="actions">
		<button type="submit" disabled={isLoading}>{isLoading ? 'Ladataan...' : 'Hae'}</button>
		<a
			href={resetHref}
			class:loading-link-disabled={isLoading}
			aria-disabled={isLoading || undefined}
			tabindex={isLoading ? -1 : undefined}
		>
			Palauta oletukset
		</a>
	</div>
</form>

<div class="result-row">
	<p class="result-total">
		{totalLabel}{#if totalInfo}<span class="info-wrapper" bind:this={infoWrapper}
				><button
					type="button"
					class="info-trigger"
					aria-label="Lisätietoa"
					aria-expanded={showInfo}
					on:click={() => (showInfo = !showInfo)}>i</button
				>{#if showInfo}<span class="info-popover" role="tooltip">{totalInfo}</span>{/if}</span
			>{/if}
	</p>
	{#if showPagination}
		<Pagination top {meta} {pageTokens} {hasPrev} {hasNext} {isLoading} {pageHref} />
		<p class="result-page">
			Sivu {meta.totalPages === 0 ? 0 : meta.page} / {meta.totalPages === 0 ? 0 : meta.totalPages}
		</p>
	{/if}
</div>

<style>
	.filters {
		display: grid;
		grid-template-columns: var(--filters-columns, 1fr);
		gap: 0.75rem;
		align-items: end;
		padding: 0.75rem;
		border: 1px solid var(--border);
		background-color: var(--card-bg);
	}

	@media (max-width: 1500px) {
		.filters {
			grid-template-columns: var(--filters-columns-1500, var(--filters-columns, 1fr));
		}
	}

	@media (max-width: 1200px) {
		.filters {
			grid-template-columns: var(--filters-columns-1200, 1fr);
		}
	}

	@media (max-width: 900px) {
		.filters {
			grid-template-columns: 1fr;
		}
	}

	.actions {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		flex-wrap: wrap;
		justify-content: flex-start;
		grid-column: 1 / -1;
		justify-self: start;
	}

	.actions a {
		white-space: nowrap;
	}

	.actions button {
		font-size: 1rem;
		padding: 0.45rem 1rem;
		border: 1px solid var(--border);
		background: var(--accent);
		color: var(--on-accent);
		cursor: pointer;
	}

	.actions button:disabled {
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

	/* the total stands out on the board wall */
	.result-total {
		font-size: 1.05rem;
		font-weight: 700;
	}

	.info-wrapper {
		position: relative;
		display: inline-block;
		margin-left: 0.35rem;
	}

	/* a clear round "i" on the wall: light paper with a gold rim */
	.info-trigger {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: 1.45rem;
		height: 1.45rem;
		padding: 0;
		border: 2px solid var(--gold);
		border-radius: 50%;
		background: var(--panel-bg);
		color: var(--ink);
		font: inherit;
		font-size: 0.95rem;
		font-weight: 700;
		font-style: italic;
		line-height: 1;
		font-family: Georgia, 'Times New Roman', serif;
		text-shadow: none;
		box-shadow: 0 1px 3px rgba(0, 0, 0, 0.5);
		vertical-align: middle;
		cursor: pointer;
	}

	.info-trigger:hover,
	.info-trigger:focus-visible {
		background: var(--accent);
		color: var(--page-bg);
		outline: none;
	}

	.info-popover {
		position: absolute;
		top: calc(100% + 0.35rem);
		left: 0;
		z-index: 10;
		min-width: 220px;
		max-width: 320px;
		padding: 0.5rem 0.65rem;
		border: 1px solid var(--border);
		background: var(--panel-bg);
		box-shadow: 0 4px 10px var(--shadow);
		/* dark text on the light box, not the wall's light text */
		color: var(--ink);
		text-shadow: none;
		font-size: 0.9rem;
		font-weight: 400;
		line-height: 1.35;
		white-space: normal;
	}

	.result-page {
		text-align: right;
		justify-self: end;
	}

	@media (max-width: 640px) {
		.result-row {
			display: flex;
			flex-wrap: wrap;
			align-items: baseline;
			margin: 0.85rem 0;
			/* the info box opens below the row, across its width */
			position: relative;
		}

		.info-wrapper {
			position: static;
		}

		.info-popover {
			left: 0;
			right: 0;
			min-width: 0;
			max-width: none;
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
	}
</style>
