<script lang="ts">
	import { createEventDispatcher, onMount } from 'svelte';
	import {
		authorList,
		italianOriginalPublication,
		storyVillainForStory,
		otherStoriesVillainHash,
		storyVillainTitle
	} from '$lib/listing/shared';
	import type { Story, Villain, StoryVillainsResponse } from '$lib/types';
	import AuthorLinks from './AuthorLinks.svelte';
	import StoryTitle from './StoryTitle.svelte';
	import VillainCard from './VillainCard.svelte';

	export let story: Story;
	export let title: string;
	export let publicationSummary: string;

	// selectStory: another story was chosen from an author popup; the parent
	// shows that story in this popup instead.
	const dispatch = createEventDispatcher<{ close: void; selectStory: Story }>();

	let closeButton: HTMLButtonElement;

	onMount(() => closeButton.focus());

	let villainsExpanded = false;
	let villainsLoading = false;
	let villainsError = '';
	let villains: Villain[] = [];
	let villainsLoadedHash = '';
	let prevHash = '';

	$: storyHash = story.hash.trim();
	$: italianOriginal = italianOriginalPublication(story);
	$: {
		if (storyHash !== prevHash) {
			prevHash = storyHash;
			villainsExpanded = false;
			villainsLoading = false;
			villainsError = '';
			villains = [];
			villainsLoadedHash = '';
		}
	}

	function close(): void {
		dispatch('close');
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
			<p><strong>Suomensi:</strong> {authorList(story.translatedBy, '; ')}</p>
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
						<div class="story-villains-list">
							{#each villains as villain (villain.hash)}
								{@const appearance = storyVillainForStory(villain, storyHash)}
								{@const cardTitle = storyVillainTitle(villain, storyHash)}
								<VillainCard
									title={cardTitle}
									otherStoriesVillainHash={otherStoriesVillainHash(villain)}
									roles={appearance?.roles}
									destiny={appearance?.destiny}
									codeNames={appearance?.codeNames}
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
		background: rgba(0, 0, 0, 0.5);
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 1rem;
	}

	.story-popup {
		width: min(920px, 100%);
		max-height: calc(100vh - 2rem);
		overflow: auto;
		background: #ffffed;
		border: 1px solid black;
		box-shadow: 0 10px 30px rgba(0, 0, 0, 0.35);
		padding: 0.9rem;
	}

	.story-popup-actions {
		display: flex;
		justify-content: flex-end;
		margin-bottom: 0.6rem;
	}

	.story-popup-close {
		padding: 0.35rem 0.7rem;
		font-size: 0.95rem;
		border: 1px solid black;
		background: #fff;
		color: #111;
		cursor: pointer;
	}

	.popup-story-card {
		margin: 0;
	}

	.story-card {
		border: 1px solid black;
		background-color: #f7f7f7;
		padding: 1rem;
		box-shadow: 0 4px 10px rgba(0, 0, 0, 0.15);
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
		border: 1px solid black;
		background: #fff;
		color: #111;
		cursor: pointer;
	}

	.toggle-villains:disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}

	.story-villains {
		margin-top: 0.8rem;
		padding-top: 0.75rem;
		border-top: 1px solid black;
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
