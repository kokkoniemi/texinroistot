<script lang="ts">
	import { hasValues, joinValues } from '$lib/listing/shared';
	import ItalicText from './ItalicText.svelte';

	export let title: string;
	// Link to the villain's other stories; only given for villains in several stories.
	export let otherStoriesHref: string | undefined = undefined;
	export let roles: string[] | null | undefined = undefined;
	export let destiny: string[] | null | undefined = undefined;
	export let codeNames: string[] | null | undefined = undefined;
</script>

<article class="story-villain-card">
	<h4>{title}</h4>
	{#if hasValues(roles)}
		<p><strong>Rooli:</strong> <ItalicText text={joinValues(roles, '-', '; ')} /></p>
	{/if}
	{#if hasValues(destiny)}
		<p><strong>Kohtalo:</strong> <ItalicText text={joinValues(destiny, '-', '; ')} /></p>
	{/if}
	{#if hasValues(codeNames)}
		<p><strong>Salanimi:</strong> {joinValues(codeNames, '-', '; ')}</p>
	{/if}
	{#if otherStoriesHref}
		<p><a href={otherStoriesHref}>Näytä roiston muut tarinat</a></p>
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
