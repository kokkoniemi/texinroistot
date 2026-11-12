<script lang="ts">
	import { createEventDispatcher, onMount } from 'svelte';
	import { popupBack, pushPopupStep } from '$lib/popupHistory';

	// Popup with a form for feedback on the database and the site.
	const dispatch = createEventDispatcher<{ close: void }>();

	const MAX_MESSAGE_LENGTH = 5000;

	let messageField: HTMLTextAreaElement;
	let message = '';
	let contact = '';
	// Hidden from people; a bot filling every field fills it too.
	let website = '';
	let sending = false;
	let sent = false;
	let errorMessage = '';

	// The popup is a step in the browser history: back (the buttons, Escape, or
	// the browser's back / a swipe on a phone) closes it.
	onMount(() => {
		pushPopupStep(() => dispatch('close'));
		messageField.focus();
	});

	function close(): void {
		popupBack();
	}

	async function send(): Promise<void> {
		if (!message.trim() || sending) return;
		sending = true;
		errorMessage = '';
		try {
			const res = await fetch('/api/palaute', {
				method: 'POST',
				headers: { 'content-type': 'application/json' },
				body: JSON.stringify({ message, contact, website })
			});
			if (res.status === 429) {
				errorMessage = 'Olet lähettänyt jo useamman palautteen. Yritä hetken päästä uudelleen.';
			} else if (!res.ok) {
				errorMessage = 'Palautteen lähetys epäonnistui. Yritä myöhemmin uudelleen.';
			} else {
				sent = true;
			}
		} catch {
			errorMessage = 'Palautteen lähetys epäonnistui. Yritä myöhemmin uudelleen.';
		} finally {
			sending = false;
		}
	}
</script>

<svelte:window
	on:keydown|capture={(e) => {
		if (e.key !== 'Escape') return;
		e.stopImmediatePropagation();
		close();
	}}
/>

<div class="feedback-backdrop" role="presentation" on:click|self={close}>
	<div class="feedback-popup" role="dialog" aria-modal="true" aria-labelledby="feedback-title">
		<div class="feedback-actions">
			<button type="button" class="feedback-button" on:click={close}>Sulje</button>
		</div>

		<article class="feedback-card">
			<h3 id="feedback-title">Palaute</h3>
			{#if sent}
				<p class="feedback-thanks">Kiitos palautteesta!</p>
				<button type="button" class="feedback-button" on:click={close}>Sulje</button>
			{:else}
				<form on:submit|preventDefault={send}>
					<label for="feedback-message">Palaute tietokannasta tai sivustosta</label>
					<textarea
						id="feedback-message"
						rows="7"
						maxlength={MAX_MESSAGE_LENGTH}
						required
						bind:value={message}
						bind:this={messageField}
					></textarea>

					<label for="feedback-contact">Nimi tai sähköposti (vapaaehtoinen)</label>
					<input
						id="feedback-contact"
						type="text"
						maxlength="200"
						autocomplete="email"
						bind:value={contact}
					/>
					<p class="feedback-hint">Jätä yhteystieto, jos haluat vastauksen.</p>

					<div class="feedback-website" aria-hidden="true">
						<label for="feedback-website">Kotisivu</label>
						<input
							id="feedback-website"
							type="text"
							tabindex="-1"
							autocomplete="off"
							bind:value={website}
						/>
					</div>

					{#if errorMessage}
						<p class="feedback-error" role="alert">{errorMessage}</p>
					{/if}

					<button type="submit" class="feedback-send" disabled={sending || !message.trim()}>
						{sending ? 'Lähetetään...' : 'Lähetä palaute'}
					</button>
				</form>
			{/if}
		</article>
	</div>
</div>

<style>
	.feedback-backdrop {
		position: fixed;
		inset: 0;
		z-index: 1100;
		background: rgba(43, 27, 16, 0.55);
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 1rem;
	}

	.feedback-popup {
		width: min(640px, 100%);
		max-height: calc(100vh - 2rem);
		overflow: auto;
		background: var(--page-bg);
		border: 1px solid var(--border);
		box-shadow: 0 10px 30px var(--shadow);
		padding: 0.9rem;
		text-align: left;
	}

	.feedback-actions {
		display: flex;
		justify-content: flex-end;
		margin-bottom: 0.6rem;
	}

	.feedback-button {
		padding: 0.35rem 0.7rem;
		font-size: 0.95rem;
		border: 1px solid var(--border);
		background: var(--panel-bg);
		color: var(--ink);
		cursor: pointer;
	}

	.feedback-card {
		border: 1px solid var(--border);
		background-color: var(--card-bg);
		padding: 1rem;
		box-shadow: 0 4px 10px var(--shadow);
	}

	.feedback-card h3 {
		margin: 0 0 0.75rem;
	}

	form {
		display: grid;
		gap: 0.35rem;
	}

	label {
		font-weight: 700;
	}

	textarea,
	input {
		width: 100%;
		box-sizing: border-box;
		padding: 0.45rem 0.55rem;
		border: 1px solid var(--border);
		background: var(--panel-bg);
		color: var(--ink);
		font: inherit;
	}

	textarea {
		resize: vertical;
		margin-bottom: 0.5rem;
	}

	.feedback-hint {
		margin: 0 0 0.5rem;
		font-size: 0.9rem;
	}

	/* off screen rather than display: none, which some bots skip */
	.feedback-website {
		position: absolute;
		left: -10000px;
		width: 1px;
		height: 1px;
		overflow: hidden;
	}

	.feedback-error {
		margin: 0 0 0.5rem;
		color: var(--color-error);
	}

	.feedback-send {
		justify-self: start;
		padding: 0.5rem 1rem;
		border: 1px solid var(--border);
		background: var(--accent);
		color: var(--on-accent);
		font: inherit;
		font-weight: 700;
		cursor: pointer;
	}

	.feedback-send:disabled {
		opacity: 0.6;
		cursor: default;
	}

	.feedback-thanks {
		margin: 0 0 0.75rem;
	}

	@media (max-width: 640px) {
		.feedback-backdrop {
			padding: 0.5rem;
		}

		.feedback-popup {
			max-height: calc(100vh - 1rem);
			padding: 0.55rem;
		}
	}
</style>
