<script lang="ts">
	import { italicSegments } from '$lib/listing/shared';

	export let text: string;

	$: segments = italicSegments(text);
</script>

{#each segments as segment, index (index)}{#if segment.italic}<em>{segment.text}</em
		>{:else}{segment.text}{/if}{/each}
