<script lang="ts">
	import type { PageData } from './$types';

	export let data: PageData;

	let isLoggingOut = false;
	let logoutError = '';

	async function logout(event: SubmitEvent): Promise<void> {
		event.preventDefault();
		if (isLoggingOut) return;

		isLoggingOut = true;
		logoutError = '';

		try {
			const response = await fetch('/api/logout', {
				method: 'POST'
			});

			if (!response.ok) {
				logoutError = 'Uloskirjautuminen epäonnistui.';
				return;
			}

			window.location.assign('/hallinta');
		} catch {
			logoutError = 'Uloskirjautuminen epäonnistui.';
		} finally {
			isLoggingOut = false;
		}
	}
</script>

<svelte:head>
	{#if !data.user.loggedIn && data.googleClientId}
		<script src="https://accounts.google.com/gsi/client" async defer></script>
	{/if}
</svelte:head>

<section class="hallinta-page">
	<h1>Hallinta</h1>

	{#if data.user.loggedIn}
		<p><strong>Kirjautunut käyttäjä:</strong> {data.user.email}</p>
		<p>Hallinnan toiminnot tulossa.</p>
		<form method="POST" action="/api/logout" on:submit={logout}>
			<button type="submit" disabled={isLoggingOut}>
				{isLoggingOut ? 'Kirjaudutaan ulos...' : 'Kirjaudu ulos'}
			</button>
		</form>
		{#if logoutError}
			<p class="config-error">{logoutError}</p>
		{/if}
	{:else}
		<p>Kirjaudu sisään Google-tilillä jatkaaksesi.</p>

		{#if data.googleClientId}
			<div
				id="g_id_onload"
				data-client_id={data.googleClientId}
				data-login_uri="/api/login"
				data-ux_mode="redirect"
				data-auto_prompt="false"
			></div>
			<div
				class="g_id_signin"
				data-type="standard"
				data-size="large"
				data-theme="outline"
				data-text="signin_with"
				data-shape="rectangular"
				data-logo_alignment="left"
			></div>
		{:else}
			<p class="config-error">
				Google-kirjautuminen ei ole käytössä: aseta `PUBLIC_GOOGLE_OAUTH2_CLIENT_ID` frontendille.
			</p>
		{/if}
	{/if}
</section>

<style>
	.hallinta-page {
		margin: 1rem 1.5rem 0;
	}

	button {
		font-size: 1rem;
		padding: 0.45rem 1rem;
		border: 1px solid black;
		background: black;
		color: white;
		cursor: pointer;
	}

	.config-error {
		color: #8a0000;
	}

	@media (max-width: 640px) {
		.hallinta-page {
			margin: 0.75rem 0.75rem 0;
		}
	}
</style>
