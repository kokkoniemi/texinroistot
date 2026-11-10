<script lang="ts">
	import PaperNotice from '$lib/components/PaperNotice.svelte';
</script>

<svelte:head>
	<title>Info – Texin roistot</title>
</svelte:head>

<section class="info-page">
	<h1>Info</h1>

	<!-- The same torn paper notice as on the front page. -->
	<PaperNotice maxWidth="46rem">
		<article class="info-text">
			<h2>Texin roistot -tietokanta</h2>
			<p>
				Texin roistot -tietokanta kokoaa yhteen Tex Willer -sarjakuvan roistot, tarinat ja tekijät.
				Tietokannasta voi hakea roistoja nimen, lempinimen, etnisen nimen, salanimen tai arvonimen
				mukaan, selata tarinoita ilmestymisjärjestyksessä ja nähdä kunkin kertojan, piirtäjän ja
				suomentajan tarinat.
			</p>
			<p>
				Tilasto sisältää kaikki Suomessa ennen vuotta 2020 julkaistut Tex Willer -tarinat
				roistoineen.
			</p>
		</article>
	</PaperNotice>
</section>

<style>
	.info-page {
		margin: 1rem 1.5rem 0;
	}

	.info-text h2 {
		margin: 0 0 0.6rem;
		font-size: 1.4rem;
	}

	.info-text p {
		margin: 0 0 0.75rem;
	}

	.info-text p:last-child {
		margin-bottom: 0;
	}

	@media (max-width: 640px) {
		.info-page {
			margin: 0.75rem 0.75rem 0;
		}
	}
</style>
