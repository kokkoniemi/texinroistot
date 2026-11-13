<script lang="ts">
	import { createEventDispatcher, getContext } from 'svelte';
	import { hasValues, joinValues } from '$lib/listing/shared';
	import type { MentionTarget } from '$lib/listing/shared';
	import { REMEMBER_STORY_LIST, villainStoryList } from '$lib/listing/storyLists';
	import type { RememberStoryList } from '$lib/listing/storyLists';
	import type { Story, Villain } from '$lib/types';
	import ItalicText from './ItalicText.svelte';
	import MentionText from './MentionText.svelte';
	import StoryListPopup from './StoryListPopup.svelte';
	import VillainWikiLink from './VillainWikiLink.svelte';

	export let title: string;
	// Set when the villain is headed by code names only: they are shown in
	// italics (after the rank), as on the Roistot page.
	export let codeNameTitle: { rank: string; codeNames: string[] } | null = null;
	// Set only for villains in several stories; opens a popup with all of them.
	export let otherStoriesVillainHash: string | undefined = undefined;
	export let roles: string[] | null | undefined = undefined;
	export let destiny: string[] | null | undefined = undefined;
	export let codeNames: string[] | null | undefined = undefined;
	// In a story's villain list: this villain's key (hash) and the story's
	// villains, so names of other villains in the role and destiny are links.
	export let villainKey = '';
	export let mentionTargets: MentionTarget[] = [];
	// The villain itself, for the Wikipedia link of villains based on real people.
	export let villain: Villain | undefined = undefined;

	// selectStory: a story was chosen in the popup; the parent shows it in its
	// story popup. mention: the name of another villain of the story was
	// clicked (its key); the parent brings that villain's card into view.
	const dispatch = createEventDispatcher<{ selectStory: Story; mention: string }>();

	// In a story popup: lets its "Edellinen" come back to the villain's list.
	const rememberStoryList = getContext<RememberStoryList | undefined>(REMEMBER_STORY_LIST);

	let otherStoriesOpen = false;
</script>

<article class="story-villain-card" data-villain-key={villainKey || undefined}>
	<h4>
		{#if codeNameTitle}{codeNameTitle.rank
				? `${codeNameTitle.rank} `
				: ''}{#each codeNameTitle.codeNames as codeName, index (codeName)}{index > 0 ? '; ' : ''}<em
					>{codeName}</em
				>{/each}{:else}{title}{/if}
	</h4>
	{#if hasValues(roles)}
		<p>
			<strong>Rooli:</strong>
			{#if mentionTargets.length > 0}<MentionText
					text={joinValues(roles, '-', '; ')}
					targets={mentionTargets}
					selfKey={villainKey}
					on:mention
				/>{:else}<ItalicText text={joinValues(roles, '-', '; ')} />{/if}
		</p>
	{/if}
	{#if hasValues(destiny)}
		<p>
			<strong>Kohtalo:</strong>
			{#if mentionTargets.length > 0}<MentionText
					text={joinValues(destiny, '-', '; ')}
					targets={mentionTargets}
					selfKey={villainKey}
					on:mention
				/>{:else}<ItalicText text={joinValues(destiny, '-', '; ')} />{/if}
		</p>
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
			{@const list = villainStoryList(title, villainHash)}
			<StoryListPopup
				{...list}
				on:close={() => (otherStoriesOpen = false)}
				on:selectStory={(e) => {
					rememberStoryList?.(list);
					otherStoriesOpen = false;
					dispatch('selectStory', e.detail);
				}}
			/>
		{/if}
	{/if}
	<!-- "Kuka oli oikea …?" last on the card -->
	<VillainWikiLink {villain} />
</article>

<style>
	/* the paper background comes from app.css */
	.story-villain-card {
		border: 1px solid var(--border);
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
