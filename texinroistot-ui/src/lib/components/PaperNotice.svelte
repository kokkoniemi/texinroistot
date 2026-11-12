<script lang="ts">
	// A notice on rough, torn paper, like a poster on a saloon wall: the
	// front page notice and the Info page use it.
	import { onMount } from 'svelte';

	export let maxWidth = '40rem';

	// Browsers draw a filtered element into one image of limited size, and a
	// phone's pixels are small: a tall notice (the Info page on a phone) went
	// over the limit and its paper was cut off. A notice taller than this many
	// device pixels goes without the frayed edge and the soft shadow (filters);
	// a plain shadow of the torn outline takes the shadow's place.
	const MAX_FILTERED_HEIGHT = 3000;

	let noticeHeight = 0;
	let pixelRatio = 1;
	onMount(() => {
		pixelRatio = window.devicePixelRatio || 1;
	});
	$: plain = noticeHeight * pixelRatio > MAX_FILTERED_HEIGHT;
</script>

<div class="paper-notice" class:plain style:max-width={maxWidth} bind:clientHeight={noticeHeight}>
	<!-- The paper: a clipped sheet (torn corners) frayed by the #ragged-edge filter. -->
	<div class="notice-paper" aria-hidden="true">
		<div class="notice-plain-shadow"></div>
		<div class="notice-sheet"></div>
	</div>

	<div class="notice-content">
		<slot />
	</div>

	<!-- Nailed to the wall: nail heads clear of the torn corners. -->
	<span class="nail nail-top-left" aria-hidden="true"></span>
	<span class="nail nail-top-right" aria-hidden="true"></span>
	<span class="nail nail-bottom-left" aria-hidden="true"></span>
	<span class="nail nail-bottom-right" aria-hidden="true"></span>
</div>

<style>
	.paper-notice {
		position: relative;
		isolation: isolate;
		margin: 2.5rem auto 0;
		/* the shadow on the wall follows the torn, frayed outline of the paper */
		filter: drop-shadow(4px 10px 12px rgba(15, 8, 3, 0.6));
	}

	/* An old square cut nail, hammered in a little crooked: rust has run
	   down from it (::before), the rusty, uneven
	   iron head sits on top (::after). */
	.nail {
		position: absolute;
		width: 12px;
		height: 11px;
		--tilt: 0deg;
	}

	.nail::before {
		content: '';
		position: absolute;
		left: -18px;
		top: -17px;
		width: 48px;
		height: 66px;
		background:
			/* rust run down the paper */ linear-gradient(
				to bottom,
				rgba(115, 55, 18, 0.5),
				rgba(115, 55, 18, 0)
			)
			50% 24px / 5px 34px no-repeat;
	}

	.nail::after {
		content: '';
		position: absolute;
		inset: 0;
		border-radius: 2px 3px 1px 3px;
		background:
			radial-gradient(circle at 34% 30%, rgba(205, 170, 125, 0.55), transparent 45%),
			radial-gradient(circle at 70% 75%, rgba(140, 65, 20, 0.6), transparent 55%),
			linear-gradient(140deg, #6e5038, #3a271a 55%, #1e140c);
		/* The head stands a little off the paper, lit from the top left: a sharp
		   contact shadow right at its edge and a longer, fainter cast shadow to
		   the bottom right. The offsets are turned back against the nail's tilt
		   so every shadow falls the same way on the page. */
		--sx: calc(cos(var(--tilt)) * 1px + sin(var(--tilt)) * 1.5px);
		--sy: calc(cos(var(--tilt)) * 1.5px - sin(var(--tilt)) * 1px);
		box-shadow:
			calc(var(--sx) * 1) calc(var(--sy) * 1) 0.5px rgba(25, 12, 2, 0.8),
			calc(var(--sx) * 3) calc(var(--sy) * 3) 2px rgba(45, 25, 8, 0.45),
			calc(var(--sx) * 5) calc(var(--sy) * 5) 5px rgba(45, 25, 8, 0.2),
			inset -1px -1px 1px rgba(0, 0, 0, 0.5),
			inset 1px 1px 0 rgba(220, 190, 150, 0.35);
		transform: rotate(var(--tilt));
	}

	.nail-top-left {
		--tilt: 9deg;
		top: 16px;
		left: 22px;
	}

	.nail-top-right {
		--tilt: -14deg;
		top: 40px;
		right: 30px;
	}

	.nail-bottom-left {
		--tilt: 22deg;
		bottom: 32px;
		left: 28px;
	}

	.nail-bottom-right {
		--tilt: -6deg;
		bottom: 18px;
		right: 22px;
	}

	.notice-paper {
		position: absolute;
		inset: 0;
		z-index: -1;
		filter: url(#ragged-edge);
	}

	/* The same paper as the page background on the other pages (app.css): the
	   paper colour with its fold creases, a lit relief (::before) and the grit
	   (::after). Browned edges; chunks torn off two corners and two tears. */
	.notice-sheet {
		position: absolute;
		inset: 0;
		background:
			url('/paper-folds.svg') center / cover no-repeat,
			var(--page-bg);
		box-shadow:
			inset 0 0 50px rgba(120, 75, 25, 0.3),
			inset 0 0 8px rgba(90, 55, 20, 0.45);
		clip-path: var(--torn);
	}

	/* the torn outline of the sheet, shared by its plain shadow */
	.notice-paper {
		--torn: polygon(
			0 0,
			/* a slanting tear down from the top edge */ 31% 0,
			calc(31% + 9px) 16px,
			calc(31% + 5px) 24px,
			calc(31% + 14px) 38px,
			calc(31% + 13px) 22px,
			calc(31% + 17px) 12px,
			calc(31% + 12px) 0,
			calc(100% - 70px) 0,
			calc(100% - 52px) 10px,
			calc(100% - 38px) 7px,
			calc(100% - 24px) 22px,
			calc(100% - 10px) 26px,
			100% 48px,
			100% 100%,
			58px 100%,
			44px calc(100% - 12px),
			26px calc(100% - 8px),
			12px calc(100% - 22px),
			0 calc(100% - 30px),
			/* a tear in from the left edge */ 0 calc(58% + 10px),
			14px calc(58% + 7px),
			22px calc(58% + 9px),
			40px calc(58% + 2px),
			24px calc(58% + 3px),
			13px 58%,
			0 calc(58% - 1px)
		);
	}
	.notice-sheet::before,
	.notice-sheet::after {
		content: '';
		position: absolute;
		inset: 0;
		pointer-events: none;
	}

	/* the lit relief of body::before */
	.notice-sheet::before {
		opacity: 0.35;
		mix-blend-mode: multiply;
		background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='700' height='700'%3E%3Cfilter id='w'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='0.006 0.011' numOctaves='5' seed='11' stitchTiles='stitch'/%3E%3CfeDiffuseLighting lighting-color='%23fff3dc' surfaceScale='4'%3E%3CfeDistantLight azimuth='225' elevation='48'/%3E%3C/feDiffuseLighting%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23w)'/%3E%3C/svg%3E");
		background-size: 700px 700px;
	}

	/* the grit of body::after */
	.notice-sheet::after {
		opacity: 0.7;
		background-image: url('/paper-grit.svg');
		background-size: 420px 420px;
	}

	.notice-content {
		padding: 2rem 2rem 2.4rem;
	}

	/* A notice too tall for filters: no frayed edge or soft shadow, but a
	   plain, slightly offset shadow of the torn outline. */
	.notice-plain-shadow {
		display: none;
	}

	.paper-notice.plain {
		filter: none;
	}

	.paper-notice.plain .notice-paper {
		filter: none;
	}

	.paper-notice.plain .notice-plain-shadow {
		display: block;
		position: absolute;
		inset: 0;
		transform: translate(3px, 6px);
		background: rgba(15, 8, 3, 0.45);
		clip-path: var(--torn);
	}
	@media (max-width: 640px) {
		.paper-notice {
			margin: 1.25rem 0.75rem 0;
		}

		.notice-content {
			/* room below the text for the bottom nails */
			padding: 1.5rem 1rem 2.4rem;
		}
	}
</style>
