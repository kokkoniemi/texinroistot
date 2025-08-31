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
	<h3>#{story.orderNumber}</h3>
	{#each story.publications as publication}
	<span>{publication.title}</span>;
	{/each}
<hr />
{/each}
