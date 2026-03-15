<script lang="ts">
	import type { PageData } from './$types';

	export let data: PageData;

	type AdminUser = {
		hash: string;
		isAdmin: boolean;
		createdAt?: string;
	};

	let users: AdminUser[] = [...(data.users ?? [])];
	let isLoggingOut = false;
	let logoutError = '';
	let isDeletingAccount = false;
	let deleteAccountError = '';
	let grantEmail = '';
	let isGrantingAdmin = false;
	let grantAdminError = '';
	let grantAdminSuccess = '';

	async function logout(event: SubmitEvent): Promise<void> {
		event.preventDefault();
		if (isLoggingOut) return;

		isLoggingOut = true;
		logoutError = '';

		try {
			const response = await fetch('/api/logout', { method: 'POST' });
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

	async function deleteAccount(event: SubmitEvent): Promise<void> {
		event.preventDefault();
		if (isDeletingAccount) return;

		isDeletingAccount = true;
		deleteAccountError = '';

		try {
			const response = await fetch('/api/me', { method: 'DELETE' });
			if (!response.ok) {
				deleteAccountError = 'Käyttäjätilin poistaminen epäonnistui.';
				return;
			}

			window.location.assign('/hallinta');
		} catch {
			deleteAccountError = 'Käyttäjätilin poistaminen epäonnistui.';
		} finally {
			isDeletingAccount = false;
		}
	}

	async function grantAdmin(event: SubmitEvent): Promise<void> {
		event.preventDefault();
		if (isGrantingAdmin) return;

		const trimmedEmail = grantEmail.trim();
		if (!trimmedEmail) {
			grantAdminError = 'Anna sähköpostiosoite.';
			grantAdminSuccess = '';
			return;
		}

		isGrantingAdmin = true;
		grantAdminError = '';
		grantAdminSuccess = '';

		try {
			const response = await fetch('/api/admin/users/grant-admin', {
				method: 'POST',
				headers: { 'content-type': 'application/json' },
				body: JSON.stringify({ email: trimmedEmail })
			});
			const payload = (await response.json().catch(() => null)) as {
				error?: string;
				user?: AdminUser;
			} | null;

			if (!response.ok) {
				grantAdminError = payload?.error ?? 'Admin-oikeuden myöntäminen epäonnistui.';
				return;
			}

			const updatedUser = payload?.user;
			if (updatedUser) {
				const existingIndex = users.findIndex((user) => user.hash === updatedUser.hash);
				if (existingIndex === -1) {
					users = [...users, updatedUser];
				} else {
					users = users.map((user, index) => (index === existingIndex ? updatedUser : user));
				}
			}

			grantAdminSuccess = `Admin-oikeus myönnetty käyttäjälle ${trimmedEmail}.`;
			grantEmail = '';
		} catch {
			grantAdminError = 'Admin-oikeuden myöntäminen epäonnistui.';
		} finally {
			isGrantingAdmin = false;
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

		{#if data.user.isAdmin}
			<p>Hallinnan toiminnot tulossa.</p>

			<section class="admin-section">
				<h2>Admin-oikeudet</h2>
				<form method="POST" on:submit={grantAdmin} class="grant-form">
					<label>
						<span>Myönnä admin-oikeus sähköpostilla</span>
						<input
							type="email"
							bind:value={grantEmail}
							placeholder="esim. käyttäjä@example.com"
							disabled={isGrantingAdmin}
						/>
					</label>
					<button type="submit" disabled={isGrantingAdmin}>
						{isGrantingAdmin ? 'Myönnetään...' : 'Myönnä admin-oikeus'}
					</button>
				</form>

				{#if grantAdminError}
					<p class="config-error">{grantAdminError}</p>
				{/if}
				{#if grantAdminSuccess}
					<p class="success-message">{grantAdminSuccess}</p>
				{/if}
				{#if data.usersError}
					<p class="config-error">{data.usersError}</p>
				{/if}

				<div class="users-list">
					<h3>Kirjautuneet käyttäjät</h3>
					{#if users.length === 0}
						<p>Käyttäjiä ei löytynyt.</p>
					{:else}
						<ul>
							{#each users as user}
								<li>
									<span class="user-hash">{user.hash}</span>
									<span>{user.isAdmin ? 'admin' : 'ei admin'}</span>
								</li>
							{/each}
						</ul>
					{/if}
				</div>
			</section>
		{:else}
			<p class="no-access">Sinulla ei ole oikeuksia hallintaan.</p>
		{/if}

		<div class="account-actions">
			<form method="POST" action="/api/logout" on:submit={logout}>
				<button type="submit" disabled={isLoggingOut}>
					{isLoggingOut ? 'Kirjaudutaan ulos...' : 'Kirjaudu ulos'}
				</button>
			</form>
			<form method="POST" on:submit={deleteAccount}>
				<button type="submit" class="danger" disabled={isDeletingAccount}>
					{isDeletingAccount ? 'Poistetaan...' : 'Poista käyttäjätilisi'}
				</button>
			</form>
		</div>

		{#if logoutError}
			<p class="config-error">{logoutError}</p>
		{/if}
		{#if deleteAccountError}
			<p class="config-error">{deleteAccountError}</p>
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

	.admin-section {
		margin: 1rem 0;
		padding: 0.8rem;
		border: 1px solid black;
		background: #f7f7f7;
	}

	.grant-form {
		display: flex;
		flex-wrap: wrap;
		gap: 0.6rem;
		align-items: end;
	}

	.grant-form label {
		display: flex;
		flex-direction: column;
		gap: 0.3rem;
		min-width: 280px;
	}

	.grant-form input {
		font-size: 1rem;
		padding: 0.45rem 0.5rem;
		border: 1px solid black;
	}

	.users-list ul {
		padding-left: 1.2rem;
	}

	.users-list li {
		display: flex;
		gap: 0.75rem;
		align-items: baseline;
	}

	.user-hash {
		font-family: monospace;
		font-size: 0.92rem;
	}

	.account-actions {
		display: flex;
		flex-wrap: wrap;
		gap: 0.6rem;
		margin-top: 1rem;
	}

	button {
		font-size: 1rem;
		padding: 0.45rem 1rem;
		border: 1px solid black;
		background: black;
		color: white;
		cursor: pointer;
	}

	button.danger {
		background: #111;
	}

	.no-access {
		font-weight: 700;
	}

	.config-error {
		color: #8a0000;
	}

	.success-message {
		color: #0d5e2b;
	}

	@media (max-width: 640px) {
		.hallinta-page {
			margin: 0.75rem 0.75rem 0;
		}

		.grant-form label {
			min-width: 100%;
		}
	}
</style>
