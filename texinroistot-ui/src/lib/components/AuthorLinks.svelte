<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import type { Author } from '$lib/listing/shared';
	import type { AuthorStoriesResponse, Story } from '$lib/types';
	import StoryListPopup from './StoryListPopup.svelte';

	export let authors: Author[] | null | undefined;
	// Which of the author's stories the popup lists: as writer or as drawer.
	export let type: 'writer' | 'drawer';

	// selectStory: a story was chosen in the author popup; the parent shows it in
	// its story popup.
	const dispatch = createEventDispatcher<{ selectStory: Story }>();

	type AuthorLink = { hash: string; name: string; details: string };

	let selected: AuthorLink | null = null;

	async function loadAuthorStories(authorHash: string): Promise<Story[]> {
		const response = await fetch(
			`/api/tekijat/${encodeURIComponent(authorHash)}/tarinat?type=${encodeURIComponent(type)}`
		);
		if (!response.ok) throw new Error(`Tarinoiden haku epäonnistui (${response.status})`);
		const payload = (await response.json()) as AuthorStoriesResponse;
		return payload.stories ?? [];
	}

	$: links = (authors ?? [])
		.map((author) => ({
			hash: (author.hash ?? '').trim(),
			name: `${author.firstName} ${author.lastName}`.trim(),
			details: (author.details ?? '').trim()
		}))
		.filter((link) => Boolean(link.name));
</script>

{#if links.length === 0}-{:else}{#each links as link, index (index)}{index > 0
			? '; '
			: ''}{#if link.hash}<button
				type="button"
				class="author-link"
				on:click={() => (selected = link)}>{link.name}</button
			>{:else}{link.name}{/if}{link.details ? ` (${link.details})` : ''}{/each}{/if}

{#if selected}
	{@const authorHash = selected.hash}
	<StoryListPopup
		title={selected.name}
		subtitle={type === 'drawer' ? 'Piirtäjä' : 'Kertoja'}
		loadStories={() => loadAuthorStories(authorHash)}
		emptyText="Tekijälle ei löytynyt tarinoita."
		on:close={() => (selected = null)}
		on:selectStory={(e) => {
			selected = null;
			dispatch('selectStory', e.detail);
		}}
	/>
{/if}

<style>
	.author-link {
		border: 0;
		padding: 0;
		background: transparent;
		color: inherit;
		font: inherit;
		text-decoration: underline;
		cursor: pointer;
	}
</style>
