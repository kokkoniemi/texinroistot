<script lang="ts">
	import { createEventDispatcher, getContext } from 'svelte';
	import type { Author } from '$lib/listing/shared';
	import { REMEMBER_STORY_LIST, authorStoryList } from '$lib/listing/storyLists';
	import type { AuthorRole, RememberStoryList } from '$lib/listing/storyLists';
	import type { Story } from '$lib/types';
	import StoryListPopup from './StoryListPopup.svelte';

	export let authors: Author[] | null | undefined;
	// Which of the author's stories the popup lists: as writer, drawer or translator.
	export let type: AuthorRole;

	// selectStory: a story was chosen in the author popup; the parent shows it in
	// its story popup.
	const dispatch = createEventDispatcher<{ selectStory: Story }>();
	// In a story popup: lets its "Edellinen" come back to the author's list.
	const rememberStoryList = getContext<RememberStoryList | undefined>(REMEMBER_STORY_LIST);

	type AuthorLink = { hash: string; name: string };
	// Authors sharing the same details are combined with "&" as in the sheet, e.g.
	// "Birger Terni & Pia Ravenna (1. – 2. p)" or, without details,
	// "Mauro Boselli & Giovanni Luigi Bonelli". ";" separates different details.
	type AuthorGroup = { links: AuthorLink[]; details: string };

	let selected: AuthorLink | null = null;
	$: selectedList = selected ? authorStoryList(selected.name, selected.hash, type) : null;

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

{#if selectedList}
	{@const list = selectedList}
	<StoryListPopup
		{...list}
		on:close={() => (selected = null)}
		on:selectStory={(e) => {
			rememberStoryList?.(list);
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
