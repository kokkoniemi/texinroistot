<script lang="ts">
	import { hasValues, joinValues } from '$lib/listing/shared';

	export let title: string;
	export let ranks: string[] | null | undefined = undefined;
	export let roles: string[] | null | undefined = undefined;
	export let destiny: string[] | null | undefined = undefined;
	export let codeNames: string[] | null | undefined = undefined;
</script>

<article class="story-villain-card">
	<h4>{title}</h4>
	{#if hasValues(ranks)}
		<p><strong>Arvo:</strong> {joinValues(ranks)}</p>
	{/if}
	{#if hasValues(roles)}
		<p><strong>Rooli:</strong> {joinValues(roles, '-', '; ')}</p>
	{/if}
	{#if hasValues(destiny)}
		<p><strong>Kohtalo:</strong> {joinValues(destiny, '-', '; ')}</p>
	{/if}
	{#if hasValues(codeNames)}
		<p><strong>Salanimi:</strong> {joinValues(codeNames)}</p>
	{/if}
</article>

<style>
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
</style>
