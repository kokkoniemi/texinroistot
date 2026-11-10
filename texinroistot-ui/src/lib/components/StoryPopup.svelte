<script lang="ts">
	import { createEventDispatcher, onMount, setContext, tick } from 'svelte';
	import {
		italianOriginalPublication,
		storyVillainForStory,
		otherStoriesVillainHash,
		storyVillainTitle,
		storyMentionTargets,
		showMentionedVillain
	} from '$lib/listing/shared';
	import type { Story, Villain, StoryVillainsResponse } from '$lib/types';
	import AuthorLinks from './AuthorLinks.svelte';
	import StoryTitle from './StoryTitle.svelte';
	import VillainCard from './VillainCard.svelte';

	export let story: Story;
	export let title: string;
	export let publicationSummary: string;
	// Opened from a villain's name in a role or destiny: the villain list opens
	// with that villain's card brought into view.
	export let focusVillainKey = '';

	// selectStory: another story was chosen from an author popup; the parent
	// shows that story in this popup instead.
	const dispatch = createEventDispatcher<{ close: void; selectStory: Story }>();

	let closeButton: HTMLButtonElement;

	// Story lists opened from this card (author or villain links) show a back
	// button that returns here.
	setContext('inStoryPopup', true);

	onMount(async () => {
		closeButton.focus();
		if (!focusVillainKey) return;
		await toggleVillains();
		await tick();
		showMentionedVillain(villainList, focusVillainKey);
	});

	let villainsExpanded = false;
	let villainsLoading = false;
	let villainsError = '';
	let villains: Villain[] = [];
	let villainsLoadedHash = '';
	let prevHash = '';
	let prevStory: Story | null = null;

	// Cards visited before this one in the popup, so "Edellinen" can step back
	// through them. Each remembers whether its villain list was open.
	type HistoryEntry = { story: Story; villainsOpen: boolean };
	let history: HistoryEntry[] = [];
	let goingBack = false;
	let reopenVillains = false;

	$: storyHash = story.hash.trim();
	$: italianOriginal = italianOriginalPublication(story);
	// Names of the story's villains in roles and destinies link to their cards.
	$: mentionTargets = storyMentionTargets(villains, storyHash);
	let villainList: HTMLElement | undefined;
	$: {
		if (storyHash !== prevHash) {
			if (prevStory && !goingBack) {
				history = [...history, { story: prevStory, villainsOpen: villainsExpanded }];
			}
			prevHash = storyHash;
			prevStory = story;
			villainsExpanded = false;
			villainsLoading = false;
			villainsError = '';
			villains = [];
			villainsLoadedHash = '';
			if (goingBack) {
				goingBack = false;
				if (reopenVillains) tick().then(toggleVillains);
			}
		}
	}

	function close(): void {
		dispatch('close');
	}

	function back(): void {
		const previous = history[history.length - 1];
		if (!previous) return;
		history = history.slice(0, -1);
		goingBack = true;
		reopenVillains = previous.villainsOpen;
		dispatch('selectStory', previous.story);
	}

	async function toggleVillains(): Promise<void> {
		if (!storyHash) return;

		if (villainsExpanded) {
			villainsExpanded = false;
			return;
		}

		villainsExpanded = true;
		if (villainsLoadedHash === storyHash || villainsLoading) return;

		villainsLoading = true;
		villainsError = '';

		try {
			const response = await fetch(`/api/tarinat/${encodeURIComponent(storyHash)}/roistot`);
			if (!response.ok) throw new Error(`Roistojen haku epäonnistui (${response.status})`);
			const payload = (await response.json()) as StoryVillainsResponse;
			if (story.hash.trim() === storyHash) {
				villains = payload.villains ?? [];
				villainsLoadedHash = storyHash;
			}
		} catch (error) {
			if (story.hash.trim() === storyHash) {
				villainsError = error instanceof Error ? error.message : 'Roistojen haku epäonnistui';
			}
		} finally {
			if (story.hash.trim() === storyHash) {
				villainsLoading = false;
			}
		}
	}
</script>

<svelte:window
	on:keydown={(e) => {
		if (e.key === 'Escape') close();
	}}
/>

<div class="story-popup-backdrop" role="presentation" on:click|self={close}>
	<div class="story-popup" role="dialog" aria-modal="true" aria-labelledby="story-popup-title">
		<div class="story-popup-actions">
			{#if history.length > 0}
				<button type="button" class="story-popup-back" on:click={back}>&larr; Edellinen</button>
			{/if}
			<button type="button" class="story-popup-close" on:click={close} bind:this={closeButton}
				>Sulje</button
			>
		</div>

		<article class="story-card popup-story-card">
			<h3 id="story-popup-title"><StoryTitle {title} /></h3>
			<p>
				<strong>Kertoi:</strong>
				<AuthorLinks
					authors={story.writtenBy}
					type="writer"
					on:selectStory={(e) => dispatch('selectStory', e.detail)}
				/>
			</p>
			<p>
				<strong>Piirsi:</strong>
				<AuthorLinks
					authors={story.drawnBy}
					type="drawer"
					on:selectStory={(e) => dispatch('selectStory', e.detail)}
				/>
			</p>
			<p>
				<strong>Suomensi:</strong>
				<AuthorLinks
					authors={story.translatedBy}
					type="translator"
					on:selectStory={(e) => dispatch('selectStory', e.detail)}
				/>
			</p>
			{#if italianOriginal}
				<p>
					<strong>Alkuperäisjulkaisu (Italia):</strong>
					{#if italianOriginal.title}<em>{italianOriginal.title}</em>{italianOriginal.details
							? ','
							: ''}{/if}
					{italianOriginal.details}
				</p>
			{/if}
			<p><strong>Ilmestynyt Suomessa:</strong> {publicationSummary}</p>

			<button
				type="button"
				class="toggle-villains"
				on:click={toggleVillains}
				disabled={!storyHash}
				aria-expanded={villainsExpanded}
			>
				{#if villainsExpanded}Piilota tarinan roistot{:else}Näytä tarinan roistot{/if}
			</button>

			{#if villainsExpanded}
				<section class="story-villains">
					{#if villainsLoading}
						<p>Haetaan roistoja...</p>
					{:else if villainsError}
						<p class="villain-error">{villainsError}</p>
					{:else if villains.length === 0}
						<p>Tarinalle ei löytynyt roistoja.</p>
					{:else}
						<div class="story-villains-list" bind:this={villainList}>
							{#each villains as villain (villain.hash)}
								{@const appearance = storyVillainForStory(villain, storyHash)}
								{@const cardTitle = storyVillainTitle(villain, storyHash)}
								<VillainCard
									title={cardTitle}
									otherStoriesVillainHash={otherStoriesVillainHash(villain)}
									roles={appearance?.roles}
									destiny={appearance?.destiny}
									codeNames={appearance?.codeNames}
									villainKey={villain.hash}
									{mentionTargets}
									on:mention={(e) => showMentionedVillain(villainList, e.detail)}
									on:selectStory={(e) => dispatch('selectStory', e.detail)}
								/>
							{/each}
						</div>
					{/if}
				</section>
			{/if}
		</article>
	</div>
</div>

<style>
	.story-popup-backdrop {
		position: fixed;
		inset: 0;
		z-index: 1000;
		background: rgba(43, 27, 16, 0.55);
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 1rem;
	}

	.story-popup {
		width: min(920px, 100%);
		max-height: calc(100vh - 2rem);
		overflow: auto;
		background: var(--page-bg);
		border: 1px solid var(--border);
		box-shadow: 0 10px 30px var(--shadow);
		padding: 0.9rem;
	}

	.story-popup-actions {
		display: flex;
		justify-content: flex-end;
		margin-bottom: 0.6rem;
	}

	/* the back button sits on the left, Sulje stays on the right */
	.story-popup-back {
		margin-right: auto;
		padding: 0.35rem 0.7rem;
		font-size: 0.95rem;
		border: 1px solid var(--border);
		background: var(--panel-bg);
		color: var(--ink);
		cursor: pointer;
	}

	.story-popup-close {
		padding: 0.35rem 0.7rem;
		font-size: 0.95rem;
		border: 1px solid var(--border);
		background: var(--panel-bg);
		color: var(--ink);
		cursor: pointer;
	}

	.popup-story-card {
		margin: 0;
	}

	.story-card {
		border: 1px solid var(--border);
		background-color: var(--card-bg);
		padding: 1rem;
		box-shadow: 0 4px 10px var(--shadow);
	}

	.story-card h3 {
		margin: 0 0 0.5rem;
	}

	.story-card p {
		margin: 0.25rem 0;
	}

	.toggle-villains {
		margin-top: 0.75rem;
		padding: 0.35rem 0.7rem;
		font-size: 0.95rem;
		border: 1px solid var(--border);
		background: var(--panel-bg);
		color: var(--ink);
		cursor: pointer;
	}

	.toggle-villains:disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}

	.story-villains {
		margin-top: 0.8rem;
		padding-top: 0.75rem;
		border-top: 1px solid var(--border);
	}

	.story-villains-list {
		display: flex;
		flex-direction: column;
		gap: 0.6rem;
	}

	.villain-error {
		color: var(--color-error);
	}

	@media (max-width: 640px) {
		.story-popup-backdrop {
			padding: 0.5rem;
		}

		.story-popup {
			max-height: calc(100vh - 1rem);
			padding: 0.55rem;
		}
	}
</style>
