<script lang="ts">
	import { storyWikipediaLink } from '$lib/data/storyWikipedia';
	import { openInOwnWindow } from '$lib/listing/shared';
	import type { Story } from '$lib/types';

	// A story based on real people or gangs links to their Wikipedia article
	// ("Ketkä olivat oikeat …?"), at the top of the story's villain list.
	export let story: Story | null | undefined;

	$: link = storyWikipediaLink(story);
</script>

{#if link}
	<p class="story-wiki">
		<a href={link.url} target="_blank" rel="noopener noreferrer" on:click={openInOwnWindow}
			>{link.label}</a
		>
	</p>
{/if}

<style>
	.story-wiki {
		margin: 0 0 0.6rem;
	}
</style>
