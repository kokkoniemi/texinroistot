<script lang="ts">
	import type { Meta, PaginationToken } from '$lib/listing/shared';
	import Pagination from './Pagination.svelte';

	export let isLoading: boolean = false;
	export let resetHref: string;
	export let totalLabel: string;
	export let meta: Meta;
	export let pageTokens: PaginationToken[];
	export let hasPrev: boolean;
	export let hasNext: boolean;
	export let pageHref: (page: number) => string;
	export let filterColumns: string = '1fr';
	export let filterColumns1500: string = '';
	export let filterColumns1200: string = '';

	$: formStyle = [
		`--filters-columns: ${filterColumns}`,
		filterColumns1500 && `--filters-columns-1500: ${filterColumns1500}`,
		filterColumns1200 && `--filters-columns-1200: ${filterColumns1200}`
	]
		.filter(Boolean)
		.join('; ');
</script>

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
	<p class="result-total">{totalLabel}</p>
	<Pagination top {meta} {pageTokens} {hasPrev} {hasNext} {isLoading} {pageHref} />
	<p class="result-page">
		Sivu {meta.totalPages === 0 ? 0 : meta.page} / {meta.totalPages === 0 ? 0 : meta.totalPages}
	</p>
</div>

<style>
	.filters {
		display: grid;
		grid-template-columns: var(--filters-columns, 1fr);
		gap: 0.75rem;
		align-items: end;
		padding: 0.75rem;
		border: 1px solid black;
		background-color: #f7f7f7;
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
		border: 1px solid black;
		background: black;
		color: white;
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
