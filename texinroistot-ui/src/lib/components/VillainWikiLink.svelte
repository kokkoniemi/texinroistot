<script lang="ts">
	import { villainWikipediaLink } from '$lib/data/villainWikipedia';
	import { openInOwnWindow } from '$lib/listing/shared';
	import type { Villain } from '$lib/types';

	// Villains based on real people link to their Wikipedia article ("Kuka oli
	// oikea …?", opened in a window of its own); found by the villain's names.
	export let villain: Villain | null | undefined;

	$: link = villainWikipediaLink(villain);
</script>

{#if link?.more?.length}
	<!-- several real people behind the villain: the question, then each by name -->
	{@const people = [{ name: link.name ?? link.url, url: link.url }, ...link.more]}
	<p class="villain-wiki">
		{link.label}
		{#each people as person, index (person.url)}{index > 0 ? ', ' : ''}<a
				href={person.url}
				target="_blank"
				rel="noopener noreferrer"
				on:click={openInOwnWindow}>{person.name}</a
			>{/each}
	</p>
{:else if link}
	<p class="villain-wiki">
		<a href={link.url} target="_blank" rel="noopener noreferrer" on:click={openInOwnWindow}
			>{link.label}</a
		>
	</p>
{/if}

<style>
	.villain-wiki {
		margin: 0.2rem 0;
	}
</style>
