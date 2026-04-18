<script lang="ts">
	import { browser } from '$app/environment';
	import { onMount } from 'svelte';
	import type { PageData } from './$types';

	export let data: PageData;

	type AdminUser = {
		hash: string;
		isAdmin: boolean;
		createdAt?: string;
	};

	type AdminVersion = {
		id: number;
		createdAt?: string;
		isActive: boolean;
	};

	let users: AdminUser[] = [...(data.users ?? [])];
	let usersError = data.usersError ?? '';
	let versions: AdminVersion[] = [...(data.versions ?? [])];
	let versionsError = data.versionsError ?? '';
	let importUrl = data.importUrl ?? '';
	let isLoggingOut = false;
	let logoutError = '';
	let isDeletingAccount = false;
	let deleteAccountError = '';
	let grantEmail = '';
	let isGrantingAdmin = false;
	let grantAdminError = '';
	let grantAdminSuccess = '';
	let isActivatingVersionID: number | null = null;
	let isDeletingVersionID: number | null = null;
	let isImportingVersion = false;
	let versionActionError = '';
	let versionActionSuccess = '';
	let isLoggingInWithGoogle = false;
	let loginError = '';
	let loginCsrfToken = '';
	let googleSignInContainer: HTMLDivElement | null = null;
	let googleButtonInitialized = false;

	type GoogleCredentialResponse = {
		credential?: string;
	};

	type GoogleIdApi = {
		initialize: (options: {
			client_id: string;
			callback: (response: GoogleCredentialResponse) => void | Promise<void>;
			auto_select?: boolean;
		}) => void;
		renderButton: (
			element: HTMLElement,
			options: {
				type: 'standard';
				size: 'large';
				theme: 'outline';
				text: 'signin_with';
				shape: 'rectangular';
				logo_alignment: 'left';
			}
		) => void;
	};

	type WindowWithGoogle = Window & {
		google?: {
			accounts?: {
				id?: GoogleIdApi;
			};
		};
	};

	function formatCreatedAt(createdAt?: string): string {
		if (!createdAt) return '-';
		const date = new Date(createdAt);
		if (Number.isNaN(date.getTime())) {
			return createdAt;
		}
		return new Intl.DateTimeFormat('fi-FI', {
			dateStyle: 'short',
			timeStyle: 'short'
		}).format(date);
	}

	function readCookie(name: string): string {
		if (!browser) return '';
		const encodedName = `${encodeURIComponent(name)}=`;
		for (const cookiePart of document.cookie.split(';')) {
			const cookie = cookiePart.trim();
			if (cookie.startsWith(encodedName)) {
				return decodeURIComponent(cookie.slice(encodedName.length));
			}
		}
		return '';
	}

	function ensureLoginCsrfToken(): string {
		if (!browser) return '';
		if (loginCsrfToken) return loginCsrfToken;

		const existing = readCookie('g_csrf_token');
		if (existing) {
			loginCsrfToken = existing;
			return loginCsrfToken;
		}

		const randomBytes = new Uint8Array(24);
		crypto.getRandomValues(randomBytes);
		loginCsrfToken = Array.from(randomBytes, (byte) => byte.toString(16).padStart(2, '0')).join('');
		const secureAttribute = window.location.protocol === 'https:' ? '; Secure' : '';
		document.cookie = `g_csrf_token=${encodeURIComponent(loginCsrfToken)}; Path=/; SameSite=Lax${secureAttribute}`;
		return loginCsrfToken;
	}

	async function loginWithGoogleCredential(credential: string): Promise<void> {
		if (isLoggingInWithGoogle) return;

		isLoggingInWithGoogle = true;
		loginError = '';

		try {
			const csrfToken = ensureLoginCsrfToken();
			const response = await fetch('/api/login', {
				method: 'POST',
				headers: { 'content-type': 'application/json' },
				body: JSON.stringify({
					credential,
					g_csrf_token: csrfToken
				})
			});

			if (!response.ok) {
				const payload = (await response.json().catch(() => null)) as { error?: string } | null;
				loginError = payload?.error ?? 'Kirjautuminen epäonnistui.';
				return;
			}

			window.location.assign('/hallinta');
		} catch {
			loginError = 'Kirjautuminen epäonnistui.';
		} finally {
			isLoggingInWithGoogle = false;
		}
	}

	function initializeGoogleButton(): boolean {
		if (!browser || data.user.loggedIn || !data.googleClientId || !googleSignInContainer) {
			return false;
		}

		const googleIdApi = (window as WindowWithGoogle).google?.accounts?.id;
		if (!googleIdApi) {
			return false;
		}
		if (googleButtonInitialized) {
			return true;
		}

		googleIdApi.initialize({
			client_id: data.googleClientId,
			callback: async (response: GoogleCredentialResponse) => {
				if (!response?.credential) {
					loginError = 'Kirjautuminen epäonnistui.';
					return;
				}
				await loginWithGoogleCredential(response.credential);
			}
		});

		googleIdApi.renderButton(googleSignInContainer, {
			type: 'standard',
			size: 'large',
			theme: 'outline',
			text: 'signin_with',
			shape: 'rectangular',
			logo_alignment: 'left'
		});

		googleButtonInitialized = true;
		return true;
	}

	onMount(() => {
		if (data.user.loggedIn || !data.googleClientId) {
			return;
		}

		ensureLoginCsrfToken();
		if (initializeGoogleButton()) {
			return;
		}

		let attempts = 0;
		const intervalId = window.setInterval(() => {
			attempts += 1;
			if (initializeGoogleButton()) {
				window.clearInterval(intervalId);
				return;
			}
			if (attempts >= 50) {
				window.clearInterval(intervalId);
				loginError = 'Google-kirjautumisen alustus epäonnistui.';
			}
		}, 200);

		return () => {
			window.clearInterval(intervalId);
		};
	});

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

	async function activateVersion(versionID: number): Promise<void> {
		if (isActivatingVersionID !== null || isDeletingVersionID !== null || isImportingVersion)
			return;

		isActivatingVersionID = versionID;
		versionActionError = '';
		versionActionSuccess = '';

		try {
			const response = await fetch(`/api/admin/versions/${versionID}/activate`, {
				method: 'POST'
			});
			const payload = (await response.json().catch(() => null)) as {
				error?: string;
				version?: AdminVersion;
			} | null;

			if (!response.ok) {
				versionActionError = payload?.error ?? 'Aktiivisen version asettaminen epäonnistui.';
				return;
			}

			const activeVersionID = payload?.version?.id ?? versionID;
			versions = versions.map((version) => ({
				...version,
				isActive: version.id === activeVersionID
			}));
			versionActionSuccess = `Versio ${activeVersionID} asetettu aktiiviseksi.`;
		} catch {
			versionActionError = 'Aktiivisen version asettaminen epäonnistui.';
		} finally {
			isActivatingVersionID = null;
		}
	}

	async function deleteVersion(version: AdminVersion): Promise<void> {
		if (
			version.isActive ||
			isActivatingVersionID !== null ||
			isDeletingVersionID !== null ||
			isImportingVersion
		)
			return;
		if (!window.confirm(`Poistetaanko versio ${version.id}? Tätä ei voi perua.`)) return;

		isDeletingVersionID = version.id;
		versionActionError = '';
		versionActionSuccess = '';

		try {
			const response = await fetch(`/api/admin/versions/${version.id}`, {
				method: 'DELETE'
			});
			const payload = (await response.json().catch(() => null)) as {
				error?: string;
			} | null;

			if (!response.ok) {
				versionActionError = payload?.error ?? 'Version poistaminen epäonnistui.';
				return;
			}

			versions = versions.filter((item) => item.id !== version.id);
			versionActionSuccess = `Versio ${version.id} poistettu.`;
		} catch {
			versionActionError = 'Version poistaminen epäonnistui.';
		} finally {
			isDeletingVersionID = null;
		}
	}

	async function refreshVersions(): Promise<boolean> {
		try {
			const response = await fetch('/api/admin/versions');
			const payload = (await response.json().catch(() => null)) as {
				error?: string;
				versions?: AdminVersion[];
				importUrl?: string;
			} | null;

			if (!response.ok) {
				versionsError = payload?.error ?? 'Versioiden haku epäonnistui.';
				return false;
			}

			versions = payload?.versions ?? [];
			importUrl = payload?.importUrl ?? '';
			versionsError = '';
			return true;
		} catch {
			versionsError = 'Versioiden haku epäonnistui.';
			return false;
		}
	}

	async function importVersionFromOneDrive(): Promise<void> {
		if (isImportingVersion || isActivatingVersionID !== null || isDeletingVersionID !== null)
			return;

		isImportingVersion = true;
		versionActionError = '';
		versionActionSuccess = '';

		try {
			const response = await fetch('/api/admin/versions/import', {
				method: 'POST'
			});
			const payload = (await response.json().catch(() => null)) as {
				error?: string;
				version?: AdminVersion;
			} | null;

			if (!response.ok) {
				versionActionError = payload?.error ?? 'Version tuonti epäonnistui.';
				return;
			}

			await refreshVersions();
			if (payload?.version?.id) {
				versionActionSuccess = `Uusi versio ${payload.version.id} tuotiin OneDrivesta.`;
			} else {
				versionActionSuccess = 'Uusi versio tuotiin OneDrivesta.';
			}
		} catch {
			versionActionError = 'Version tuonti epäonnistui.';
		} finally {
			isImportingVersion = false;
		}
	}
</script>

<svelte:head>
	<title>Hallinta – Texin roistot</title>
	{#if !data.user.loggedIn && data.googleClientId}
		<script src="https://accounts.google.com/gsi/client" async defer></script>
	{/if}
</svelte:head>

<section class="hallinta-page">
	<h1>Hallinta</h1>

	{#if data.user.loggedIn}
		<p><strong>Kirjautunut käyttäjä:</strong> {data.user.email}</p>

		{#if data.user.isAdmin}
			<section class="admin-section">
				<h2>Versiot</h2>
				<div class="version-toolbar">
					<button
						type="button"
						on:click={importVersionFromOneDrive}
						disabled={isImportingVersion ||
							isActivatingVersionID !== null ||
							isDeletingVersionID !== null}
					>
						{isImportingVersion ? 'Tuodaan OneDrivesta...' : 'Tuo uusi versio OneDrivesta'}
					</button>
				</div>
				<p class="import-url-status">
					<strong>Käytössä oleva tuonti-URL:</strong>
					{#if importUrl}
						<a href={importUrl} target="_blank" rel="noreferrer noopener">{importUrl}</a>
					{:else}
						<em>Ei asetettu</em>
					{/if}
				</p>
				{#if versionsError}
					<p class="config-error">{versionsError}</p>
				{/if}
				{#if versionActionError}
					<p class="config-error">{versionActionError}</p>
				{/if}
				{#if versionActionSuccess}
					<p class="success-message">{versionActionSuccess}</p>
				{/if}

				<div class="versions-list">
					{#if versions.length === 0}
						<p>Versioita ei löytynyt.</p>
					{:else}
						<ul>
							{#each versions as version}
								<li class="version-item">
									<div class="version-meta">
										<span><strong>ID:</strong> {version.id}</span>
										<span><strong>Luotu:</strong> {formatCreatedAt(version.createdAt)}</span>
										<span class:active-status={version.isActive} class="version-status">
											{version.isActive ? 'Aktiivinen' : 'Ei aktiivinen'}
										</span>
									</div>

									<div class="version-actions">
										{#if version.isActive}
											<span class="active-note">Aktiivinen versio (ei poistettavissa)</span>
										{:else}
											<button
												type="button"
												on:click={() => activateVersion(version.id)}
												disabled={isImportingVersion ||
													isActivatingVersionID !== null ||
													isDeletingVersionID !== null}
											>
												{isActivatingVersionID === version.id
													? 'Asetetaan...'
													: 'Aseta aktiiviseksi'}
											</button>
											<button
												type="button"
												class="danger"
												on:click={() => deleteVersion(version)}
												disabled={isImportingVersion ||
													isActivatingVersionID !== null ||
													isDeletingVersionID !== null}
											>
												{isDeletingVersionID === version.id ? 'Poistetaan...' : 'Poista versio'}
											</button>
										{/if}
									</div>
								</li>
							{/each}
						</ul>
					{/if}
				</div>
			</section>

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
				{#if usersError}
					<p class="config-error">{usersError}</p>
				{/if}

				<div class="users-list">
					<h3>Käyttäjät</h3>
					{#if users.length === 0}
						<p>Käyttäjiä ei löytynyt.</p>
					{:else}
						<table class="users-table">
							<thead>
								<tr>
									<th>Sähköpostiosoitteen tiiviste (kuten tietokannassa)</th>
									<th>Pääsyoikeus</th>
									<th>Luotu</th>
								</tr>
							</thead>
							<tbody>
								{#each users as user}
									<tr>
										<td class="user-hash">{user.hash}</td>
										<td>{user.isAdmin ? 'admin' : 'ei admin'}</td>
										<td>{formatCreatedAt(user.createdAt)}</td>
									</tr>
								{/each}
							</tbody>
						</table>
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
			<div class="g_id_signin" bind:this={googleSignInContainer}></div>
			{#if isLoggingInWithGoogle}
				<p>Kirjaudutaan sisään...</p>
			{/if}
			{#if loginError}
				<p class="config-error">{loginError}</p>
			{/if}
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

	.users-table {
		width: 100%;
		border-collapse: collapse;
		background: white;
	}

	.users-table th,
	.users-table td {
		padding: 0.45rem 0.55rem;
		border: 1px solid black;
		text-align: left;
		vertical-align: top;
	}

	.users-table th {
		font-weight: 700;
	}

	.versions-list ul {
		padding-left: 0;
		list-style: none;
	}

	.version-toolbar {
		margin-bottom: 0.6rem;
	}

	.import-url-status {
		margin: 0 0 0.6rem;
		font-size: 0.9rem;
		color: #2f2f2f;
	}

	.import-url-status a {
		margin-left: 0.35rem;
		color: inherit;
		word-break: break-all;
	}

	.version-item {
		display: flex;
		flex-wrap: wrap;
		justify-content: space-between;
		gap: 0.75rem;
		padding: 0.6rem;
		border: 1px solid black;
		background: white;
	}

	.version-item + .version-item {
		margin-top: 0.6rem;
	}

	.version-meta {
		display: flex;
		flex-wrap: wrap;
		gap: 0.5rem 1rem;
		align-items: baseline;
	}

	.version-status {
		font-weight: 700;
	}

	.version-status.active-status {
		color: #0d5e2b;
	}

	.version-actions {
		display: flex;
		flex-wrap: wrap;
		gap: 0.5rem;
	}

	.active-note {
		font-weight: 700;
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
