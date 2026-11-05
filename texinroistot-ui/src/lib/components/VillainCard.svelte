<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import { hasValues, joinValues } from '$lib/listing/shared';
	import type { Story, Villain } from '$lib/types';
	import ItalicText from './ItalicText.svelte';
	import StoryListPopup from './StoryListPopup.svelte';

	export let title: string;
	// Set only for villains in several stories; opens a popup with all of them.
	export let otherStoriesVillainHash: string | undefined = undefined;
	export let roles: string[] | null | undefined = undefined;
	export let destiny: string[] | null | undefined = undefined;
	export let codeNames: string[] | null | undefined = undefined;

	// selectStory: a story was chosen in the popup; the parent shows it in its
	// story popup.
	const dispatch = createEventDispatcher<{ selectStory: Story }>();

	let otherStoriesOpen = false;

	async function loadVillainStories(villainHash: string): Promise<Story[]> {
		const response = await fetch(
			`/api/roistot?publication=all&villain=${encodeURIComponent(villainHash)}`
		);
		if (!response.ok) throw new Error(`Tarinoiden haku epäonnistui (${response.status})`);
		const payload = (await response.json()) as { villains?: Villain[] };

		const stories: Story[] = [];
		const seenHashes = new Set<string>();
		for (const appearance of payload.villains?.[0]?.as ?? []) {
			const story = appearance.story;
			if (!story || seenHashes.has(story.hash)) continue;
			seenHashes.add(story.hash);
			stories.push(story);
		}
		return stories;
	}
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
	{#if otherStoriesVillainHash}
		{@const villainHash = otherStoriesVillainHash}
		<p>
			<button type="button" class="other-stories-link" on:click={() => (otherStoriesOpen = true)}
				>Näytä roiston muut tarinat</button
			>
		</p>
		{#if otherStoriesOpen}
			<StoryListPopup
				{title}
				subtitle="Roisto"
				loadStories={() => loadVillainStories(villainHash)}
				on:close={() => (otherStoriesOpen = false)}
				on:selectStory={(e) => {
					otherStoriesOpen = false;
					dispatch('selectStory', e.detail);
				}}
			/>
		{/if}
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

	.other-stories-link {
		border: 0;
		padding: 0;
		background: transparent;
		color: inherit;
		font: inherit;
		text-decoration: underline;
		cursor: pointer;
	}
</style>
