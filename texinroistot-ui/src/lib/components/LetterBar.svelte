<script lang="ts">
	// A–Ö row for alphabetical lists: a letter jumps to the page where it starts.
	// Letters no item starts with are shown greyed out.
	export let letters: string[];
	export let letterHref: (letter: string) => string;
	export let isLoading = false;

	const alphabet = [...'abcdefghijklmnopqrstuvwxyzåäö'];

	$: available = new Set(letters);
</script>

<nav class="letter-bar" aria-label="Hyppää kirjaimeen">
	{#each alphabet as letter (letter)}
		{#if available.has(letter) && !isLoading}
			<a href={letterHref(letter)}>{letter.toUpperCase()}</a>
		{:else}
			<span class="disabled">{letter.toUpperCase()}</span>
		{/if}
	{/each}
</nav>

<style>
	.letter-bar {
		margin: 0.75rem 0;
		display: flex;
		flex-wrap: wrap;
		justify-content: center;
		gap: 0.35rem 0.6rem;
	}

	.letter-bar a,
	.letter-bar span {
		min-width: 1.1rem;
		text-align: center;
	}

	.disabled {
		color: #999;
	}
</style>
