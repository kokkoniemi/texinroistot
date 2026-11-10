<script lang="ts">
	import { villainWikipediaLink } from '$lib/data/villainWikipedia';
	import { openInOwnWindow } from '$lib/listing/shared';

	// The villain's hash: villains based on real people link to their
	// Wikipedia article ("Kuka oli oikea …?", opened in a window of its own).
	export let hash: string | null | undefined;

	$: link = villainWikipediaLink(hash);
</script>

{#if link}
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
