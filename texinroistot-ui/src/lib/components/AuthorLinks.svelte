<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import { translatedPartTitles } from '$lib/listing/shared';
	import { authorWikipediaUrl } from '$lib/data/authorWikipedia';
	import type { Author } from '$lib/listing/shared';
	import type { AuthorStoriesResponse, Story } from '$lib/types';
	import StoryListPopup from './StoryListPopup.svelte';

	export let authors: Author[] | null | undefined;
	// Which of the author's stories the popup lists: as writer, drawer or translator.
	export let type: 'writer' | 'drawer' | 'translator';

	// selectStory: a story was chosen in the author popup; the parent shows it in
	// its story popup.
	const dispatch = createEventDispatcher<{ selectStory: Story }>();

	const roleLabels = { writer: 'Kertoja', drawer: 'Piirtäjä', translator: 'Suomentaja' };

	type AuthorLink = { hash: string; name: string };
	// Authors sharing the same details are combined with "&" as in the sheet, e.g.
	// "Birger Terni & Pia Ravenna (1. – 2. p)" or, without details,
	// "Mauro Boselli & Giovanni Luigi Bonelli". ";" separates different details.
	type AuthorGroup = { links: AuthorLink[]; details: string };

	let selected: AuthorLink | null = null;

	async function loadAuthorStories(authorHash: string): Promise<Story[]> {
		const response = await fetch(
			`/api/tekijat/${encodeURIComponent(authorHash)}/tarinat?type=${encodeURIComponent(type)}`
		);
		if (!response.ok) throw new Error(`Tarinoiden haku epäonnistui (${response.status})`);
		const payload = (await response.json()) as AuthorStoriesResponse;
		return payload.stories ?? [];
	}

	function groupAuthors(list: Author[] | null | undefined): AuthorGroup[] {
		const groups: AuthorGroup[] = [];
		for (const author of list ?? []) {
			const name = `${author.firstName} ${author.lastName}`.trim();
			if (!name) continue;
			const link = { hash: (author.hash ?? '').trim(), name };
			const details = (author.details ?? '').trim();
			const group = groups.find((existing) => existing.details === details);
			if (group) {
				if (!group.links.some((existing) => existing.name === name)) group.links.push(link);
			} else {
				groups.push({ links: [link], details });
			}
		}
		return groups;
	}

	$: groups = groupAuthors(authors);
</script>

{#if groups.length === 0}-{:else}{#each groups as group, groupIndex (groupIndex)}{groupIndex > 0
			? '; '
			: ''}{#each group.links as link, linkIndex (linkIndex)}{linkIndex > 0
				? ' & '
				: ''}{#if link.hash}<button
					type="button"
					class="author-link"
					on:click={() => (selected = link)}>{link.name}</button
				>{:else}{link.name}{/if}{/each}{group.details ? ` (${group.details})` : ''}{/each}{/if}

{#if selected}
	{@const authorHash = selected.hash}
	<StoryListPopup
		title={selected.name}
		subtitle={roleLabels[type]}
		wikipediaUrl={authorWikipediaUrl(selected.name)}
		loadStories={() => loadAuthorStories(authorHash)}
		storyTitles={type === 'translator'
			? (story) => translatedPartTitles(story, authorHash)
			: undefined}
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
