<script lang="ts">
	import FeedbackPopup from '$lib/components/FeedbackPopup.svelte';
	import PaperNotice from '$lib/components/PaperNotice.svelte';

	let feedbackOpen = false;
</script>

<svelte:head>
	<title>Info – Texin roistot</title>
</svelte:head>

<section class="info-page">
	<h1 class="on-wall">Info</h1>

	<!-- The same torn paper notice as on the front page. -->
	<PaperNotice maxWidth="46rem">
		<article class="info-text">
			<h2>Texin roistot -tietokanta</h2>
			<p>
				Tietokanta sisältää kaikki Tex Willer -sarjakuvassa ennen vuotta 2020 Suomessa julkaistut
				tarinat. Roistoista on huomioitu kaikki ne, joiden nimi tai kutsumanimi on mainittu
				tarinassa. Jokaiselle roistolle on määritelty rooli, eli millä agendalla kelmi on tarinassa
				liikkunut. Kohtalo-kenttään on kirjoitettu mitä lurjukselle on ko. tarinassa tapahtunut. Jos
				kohtalo ei selviä tarinassa, sitä ei ole määritelty. Roiston nimi tarinassa on se, millä
				nimellä juuri siinä tarinassa roistoa on kutsuttu.
			</p>
			<p>
				Tämä tilasto on täysin subjektiivinen, yhden henkilön näkemys. Roistoksi on määritelty
				kaikki ne, jotka ovat rikollisia tai vastustavat ystäviämme tarinassa. Osa rikollisista on
				jo parantanut tapansa ja toimii tarinassa ystäviemme puolella, mutta on silti päätynyt
				tilastoon. Jos konnan nimi on mainittu tarinassa, mutta maininnasta ei selviä, onko ko.
				henkilö rikollinen, häntä ei ole merkitty tilastoon. Tällaisia henkilöitä ovat esimerkiksi
				monet historialliset intiaanipäälliköt, kuten Geronimo, Victorio ym.
			</p>
			<p>
				Miksi tietokanta ei sisällä 2020-luvun tarinoita? Tex Willer -tarinoiden taso on
				valitettavasti laskenut 1990-luvulta saakka. Isä ja poika Bonellin kuolema, Nizzin
				tarinoiden tason lasku ja Bosellin tulo pääkirjoittajaksi ovat muuttaneet sarjakuvaa liikaa.
				Viime vuosina tarinoita on julkaistu eri formaateissa niin paljon, että johonkin piti vetää
				raja. Monet uudet erikoisjulkaisut eivät mielestäni muistuta enää vähääkään Tex Willer
				-sarjakuvaa – jos haluan lukea Aku Ankkaa tai Blueberryä, luen niitä. Siksi rajasin tilaston
				vuoteen 2020.
			</p>
			<p>
				Tarinoiden tietoihin olen saanut lähteeksi Janne Viitalan ylläpitämän tietokannan. Sivuston
				teknisestä toteutuksesta on vastannut Mikko K. Kiitos!
			</p>
			<p class="signature">Jouni K</p>
			<p class="feedback-line">
				Tietokannasta ja sivustosta voi antaa palautetta <button
					type="button"
					class="feedback-link"
					on:click={() => (feedbackOpen = true)}>tästä</button
				>.
			</p>
		</article>
	</PaperNotice>
</section>

{#if feedbackOpen}
	<FeedbackPopup on:close={() => (feedbackOpen = false)} />
{/if}

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

	.info-text .signature {
		font-style: italic;
	}

	/* the feedback line in the display font, set apart from the text above */
	.info-text .feedback-line {
		margin-top: 1.25rem;
		font-family: var(--font-display);
		font-size: 1.05rem;
		letter-spacing: 0.02em;
	}

	.feedback-link {
		border: 0;
		padding: 0;
		background: transparent;
		color: var(--accent);
		font: inherit;
		letter-spacing: inherit;
		text-decoration: underline;
		cursor: pointer;
	}

	.feedback-link:hover {
		color: var(--accent-hover);
	}

	@media (max-width: 640px) {
		.info-page {
			margin: 0.75rem 0.75rem 0;
		}
	}
</style>
