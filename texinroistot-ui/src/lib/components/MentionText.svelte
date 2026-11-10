<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import { mentionSegments } from '$lib/listing/shared';
	import type { MentionTarget } from '$lib/listing/shared';

	// A role or destiny text with italic markup, where names of the story's
	// other villains are links that open (scroll to) that villain's card.
	export let text: string;
	export let targets: MentionTarget[];
	// The villain whose text this is; its own names are not linked.
	export let selfKey: string;

	const dispatch = createEventDispatcher<{ mention: string }>();

	$: segments = mentionSegments(text, targets, selfKey);
</script>

{#each segments as segment, index (index)}{#if segment.mention}{@const key = segment.mention}<button
			type="button"
			class="mention-link"
			on:click={() => dispatch('mention', key)}
			>{#if segment.italic}<em>{segment.text}</em>{:else}{segment.text}{/if}</button
		>{:else if segment.italic}<em>{segment.text}</em>{:else}{segment.text}{/if}{/each}

<style>
	.mention-link {
		border: 0;
		padding: 0;
		background: transparent;
		color: inherit;
		font: inherit;
		text-decoration: underline;
		text-decoration-style: dotted;
		text-underline-offset: 2px;
		cursor: pointer;
	}

	.mention-link:hover {
		color: var(--accent);
		text-decoration-style: solid;
	}
</style>
