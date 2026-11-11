<script lang="ts">
	import { createEventDispatcher, getContext, onMount } from 'svelte';
	import {
		nonItalianTitlesByFirstPublication,
		openInOwnWindow,
		publicationSummaryFromPublications
	} from '$lib/listing/shared';
	import { handOverPopupStep, popupBack, pushPopupStep } from '$lib/popupHistory';
	import type { Story } from '$lib/types';
	import StoryTitle from './StoryTitle.svelte';

	// Popup listing an author's or a villain's stories in the Tekijät page format.
	export let title: string;
	export let subtitle = '';
	export let loadStories: () => Promise<Story[]>;
	export let emptyText = 'Tarinoita ei löytynyt.';
	// Optional: the titles to show for a story, e.g. only a translator's parts.
	// Defaults to all of the story's titles.
	export let storyTitles: ((story: Story) => string[]) | undefined = undefined;
	// Optional: a link to the author's Wikipedia article.
	export let wikipediaUrl: string | undefined = undefined;

	// selectStory: a story title was clicked; the parent replaces this popup with
	// the story popup.
	const dispatch = createEventDispatcher<{ close: void; selectStory: Story }>();

	// Opened from a story card (StoryPopup): "Edellinen" returns to that card.
	const openedFromStoryCard = getContext<boolean>('inStoryPopup') ?? false;

	let closeButton: HTMLButtonElement;
	let stories: Story[] = [];
	let loading = true;
	let errorMessage = '';

	// The list is a step in the browser history: back (the buttons, Escape, or
	// the browser's back / a swipe on a phone) closes it.
	let historyDepth = 0;

	function close(): void {
		popupBack();
	}

	function selectStory(story: Story): void {
		// the story card opened next takes over this list's history entry
		handOverPopupStep(historyDepth);
		dispatch('selectStory', story);
	}

	function storyTitle(story: Story): string {
		const titles = storyTitles
			? storyTitles(story)
			: nonItalianTitlesByFirstPublication(story.publications);
		return titles.length > 0 ? titles.join('; ') : 'Nimetön tarina';
	}

	onMount(async () => {
		historyDepth = pushPopupStep(() => dispatch('close'));
		closeButton.focus();
		try {
			stories = await loadStories();
		} catch (error) {
			errorMessage = error instanceof Error ? error.message : 'Tarinoiden haku epäonnistui';
		} finally {
			loading = false;
		}
	});
</script>

<!-- Capture phase: Escape closes only this popup, not a story popup underneath. -->
<svelte:window
	on:keydown|capture={(e) => {
		if (e.key !== 'Escape') return;
		e.stopImmediatePropagation();
		close();
	}}
/>

<div class="list-popup-backdrop" role="presentation" on:click|self={close}>
	<div class="list-popup" role="dialog" aria-modal="true" aria-labelledby="list-popup-title">
		<div class="list-popup-actions">
			{#if openedFromStoryCard}
				<button type="button" class="list-popup-back" on:click={close}>&larr; Edellinen</button>
			{/if}
			<button type="button" class="list-popup-close" on:click={close} bind:this={closeButton}
				>Sulje</button
			>
		</div>

		<article class="list-card">
			<h3 id="list-popup-title">{title}</h3>
			{#if subtitle || wikipediaUrl}
				<p class="list-subtitle">
					{subtitle}{#if wikipediaUrl}{subtitle ? ' · ' : ''}<a
							href={wikipediaUrl}
							target="_blank"
							rel="noopener noreferrer"
							on:click={openInOwnWindow}>Wikipedia</a
						>{/if}
				</p>
			{/if}

			<section class="list-stories">
				{#if loading}
					<p>Haetaan tarinoita...</p>
				{:else if errorMessage}
					<p class="list-stories-error">{errorMessage}</p>
				{:else if stories.length === 0}
					<p>{emptyText}</p>
				{:else}
					<ul class="list-story-list">
						{#each stories as story (story.hash)}
							<li>
								<button type="button" class="list-story-link" on:click={() => selectStory(story)}>
									<StoryTitle title={storyTitle(story)} />
								</button>
								<br />
								<span
									>{publicationSummaryFromPublications(
										story.publications,
										'Ei julkaisutietoja'
									)}</span
								>
							</li>
						{/each}
					</ul>
				{/if}
			</section>
		</article>
	</div>
</div>

<style>
	.list-popup-backdrop {
		position: fixed;
		inset: 0;
		z-index: 1100;
		background: rgba(43, 27, 16, 0.55);
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 1rem;
	}

	.list-popup {
		width: min(920px, 100%);
		max-height: calc(100vh - 2rem);
		overflow: auto;
		background: var(--page-bg);
		border: 1px solid var(--border);
		box-shadow: 0 10px 30px var(--shadow);
		padding: 0.9rem;
		font-weight: normal;
		text-align: left;
	}

	.list-popup-actions {
		display: flex;
		justify-content: flex-end;
		margin-bottom: 0.6rem;
	}

	/* the back button sits on the left, Sulje stays on the right */
	.list-popup-back {
		margin-right: auto;
		padding: 0.35rem 0.7rem;
		font-size: 0.95rem;
		border: 1px solid var(--border);
		background: var(--panel-bg);
		color: var(--ink);
		cursor: pointer;
	}

	.list-popup-close {
		padding: 0.35rem 0.7rem;
		font-size: 0.95rem;
		border: 1px solid var(--border);
		background: var(--panel-bg);
		color: var(--ink);
		cursor: pointer;
	}

	.list-card {
		border: 1px solid var(--border);
		background-color: var(--card-bg);
		padding: 1rem;
		box-shadow: 0 4px 10px var(--shadow);
	}

	.list-card h3 {
		margin: 0 0 0.25rem;
	}

	.list-subtitle {
		margin: 0 0 0.5rem;
	}

	.list-stories {
		border: 1px solid var(--border);
		background: var(--panel-bg);
		padding: 0.65rem 0.8rem;
	}

	.list-story-list {
		margin: 0;
		padding-left: 1.1rem;
		display: grid;
		gap: 0.4rem;
	}

	.list-story-link {
		border: 0;
		padding: 0;
		background: transparent;
		color: inherit;
		font: inherit;
		font-weight: 700;
		text-align: left;
		text-decoration: underline;
		cursor: pointer;
	}

	.list-stories-error {
		margin: 0;
		color: var(--color-error);
	}

	@media (max-width: 640px) {
		.list-popup-backdrop {
			padding: 0.5rem;
		}

		.list-popup {
			max-height: calc(100vh - 1rem);
			padding: 0.55rem;
		}
	}
</style>
