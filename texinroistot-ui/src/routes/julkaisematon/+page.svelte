<script lang="ts">
	import type { ActionData, PageData } from './$types';

	export let data: PageData;
	export let form: ActionData;
</script>

<section class="gate-page">
	<h1>Sivustoa ei ole vielä julkaistu</h1>
	<p>Palvelu on tällä hetkellä suljetussa testikäytössä. Syötä salasana jatkaaksesi.</p>

	{#if !data.passwordConfigured}
		<p class="error">Testikäytön salasanaa ei ole konfiguroitu oikein.</p>
	{:else}
		<form method="POST" class="gate-form">
			<input type="hidden" name="next" value={form?.next ?? data.next} />
			<label for="password">Salasana</label>
			<input
				id="password"
				name="password"
				type="password"
				autocomplete="current-password"
				required
			/>
			<button type="submit">Avaa sivusto</button>
		</form>
	{/if}

	{#if form?.error}
		<p class="error">{form.error}</p>
	{/if}
</section>

<style>
	.gate-page {
		max-width: 560px;
		margin: 96px auto;
		padding: 24px;
		border: 1px solid #000;
		background: #ffffed;
	}

	h1 {
		margin-top: 0;
		margin-bottom: 12px;
	}

	p {
		margin: 0 0 16px;
	}

	.gate-form {
		display: grid;
		gap: 10px;
	}

	label {
		font-weight: 700;
	}

	input {
		font: inherit;
		padding: 10px;
		border: 1px solid #000;
		background: #fff;
	}

	button {
		font: inherit;
		padding: 10px 14px;
		border: 1px solid #000;
		background: #000;
		color: #fff;
		cursor: pointer;
	}

	.error {
		color: #8f0000;
		font-weight: 700;
	}

	@media (max-width: 700px) {
		.gate-page {
			margin: 40px 14px;
			padding: 16px;
		}
	}
</style>
