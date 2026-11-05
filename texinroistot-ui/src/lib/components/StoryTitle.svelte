<script lang="ts">
	import { storyTitleSegments } from '$lib/listing/shared';

	export let title: string;

	$: segments = storyTitleSegments(title);
</script>

{#each segments as segment, index (index)}{#if segment.isAnnotation}<span class="annotation"
			>{segment.text}</span
		>{:else}{segment.text}{/if}{/each}

<style>
	.annotation {
		font-weight: normal;
	}
</style>
