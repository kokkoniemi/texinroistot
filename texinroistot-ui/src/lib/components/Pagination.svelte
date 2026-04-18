<script lang="ts">
	import type { Meta, PaginationToken } from '$lib/listing/shared';

	export let meta: Meta;
	export let pageTokens: PaginationToken[];
	export let hasPrev: boolean;
	export let hasNext: boolean;
	export let isLoading: boolean = false;
	export let pageHref: (page: number) => string;
	export let top: boolean = false;
</script>

<nav class="pagination" class:pagination-top={top}>
	{#if hasPrev && !isLoading}
		<a href={pageHref(meta.page - 1)}>Edellinen</a>
	{:else}
		<span class="disabled">Edellinen</span>
	{/if}

	{#each pageTokens as token, i (i)}
		{#if token === 'ellipsis'}
			<span class="ellipsis">...</span>
		{:else if token === meta.page}
			<span class="current-page">{token}</span>
		{:else if !isLoading}
			<a href={pageHref(token)}>{token}</a>
		{:else}
			<span class="disabled">{token}</span>
		{/if}
	{/each}

	{#if hasNext && !isLoading}
		<a href={pageHref(meta.page + 1)}>Seuraava</a>
	{:else}
		<span class="disabled">Seuraava</span>
	{/if}
</nav>

<style>
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

	@media (max-width: 640px) {
		.pagination {
			justify-content: center;
			gap: 0.55rem;
		}

		.pagination-top {
			order: 3;
			width: 100%;
			justify-content: center;
		}
	}
</style>
