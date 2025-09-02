<script lang="ts">
	import { writable } from 'svelte/store'
	import { setContext } from 'svelte'; 
	import type { PageData } from './$types';
	export let data: PageData;

	type TStory = {
		orderNumber: number;
	};

	const stories = writable<TStory[]>();
	$: stories.set(data.stories);
	setContext('stories', stories);
</script>

<h1>Tarinat</h1>

{#each $stories as story}
<hr />
	<h3>#{story.orderNumber} {story.title}</h3>
	<span>Kirjoitti:</span> {#each story.writtenBy as writer}<span>{writer.firstName} {writer.lastName}</span>{/each}
	<br />
	<span>Piirsi:</span> {#each story.drawnBy as drawer}<span>{drawer.firstName} {drawer.lastName}</span>{/each}

<hr />
{/each}
