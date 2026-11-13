<script lang="ts">
	// A notice on rough, torn paper, like a poster on a saloon wall: the
	// front page notice and the Info page use it.
	import { onMount } from 'svelte';

	export let maxWidth = '40rem';
	// Degrees the notice hangs askew on the wall (several notices one under
	// another, e.g. the Info page on a phone, each a little differently).
	export let tilt = 0;
	// Where the paper is torn, in percent: the tear down from the top edge
	// (from the left) and the tear in from the left edge (from the top), so
	// notices one under another are torn in different places.
	export let tearX = 31;
	export let tearY = 58;
	// The edge the side tear is torn in from.
	export let tearSide: 'left' | 'right' = 'left';
	// Chunks torn off two corners (top right and bottom left); without them
	// the corners are whole and sharp.
	export let tornCorners = true;

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

<div
	class="paper-notice"
	class:plain
	style:max-width={maxWidth}
	style:rotate={tilt ? `${tilt}deg` : undefined}
	style:--tear-x="{tearX}%"
	style:--tear-y="{tearY}%"
	style:--left-tear={tearSide === 'left' ? 1 : 0}
	style:--right-tear={tearSide === 'right' ? 1 : 0}
	style:--corner={tornCorners ? 1 : 0}
	bind:clientHeight={noticeHeight}
>
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

	/* The same paper as the cards (app.css): the paper colour with a lit
	   relief and the grit, here with fold creases. Browned edges; chunks torn
	   off two corners and two tears. */
	.notice-sheet {
		position: absolute;
		inset: 0;
		/* the folds as a light map over the paper (hard-light: grey leaves the
		   paper as it is), on the paper image of the cards; ready images rather
		   than SVG filters, which a phone computed slowly (app.css) */
		background:
			url('/paper-folds-light.webp') center / cover no-repeat,
			url('/paper-sheet.webp') 0 0 / 840px 840px repeat var(--page-bg);
		background-blend-mode: hard-light, normal;
		box-shadow:
			inset 0 0 50px rgba(120, 75, 25, 0.3),
			inset 0 0 8px rgba(90, 55, 20, 0.45);
		clip-path: var(--torn);
	}

	/* the torn outline of the sheet, shared by its plain shadow */
	.notice-paper {
		--torn: polygon(
			0 0,
			/* a slanting tear down from the top edge */ var(--tear-x) 0,
			calc(var(--tear-x) + 9px) 16px,
			calc(var(--tear-x) + 5px) 24px,
			calc(var(--tear-x) + 14px) 38px,
			calc(var(--tear-x) + 13px) 22px,
			calc(var(--tear-x) + 17px) 12px,
			calc(var(--tear-x) + 12px) 0,
			calc(100% - 70px * var(--corner)) 0,
			calc(100% - 52px * var(--corner)) calc(10px * var(--corner)),
			calc(100% - 38px * var(--corner)) calc(7px * var(--corner)),
			calc(100% - 24px * var(--corner)) calc(22px * var(--corner)),
			calc(100% - 10px * var(--corner)) calc(26px * var(--corner)),
			100% calc(48px * var(--corner)),
			/* or the tear in from the right edge */ 100% calc(var(--tear-y) - 1px),
			calc(100% - 13px * var(--right-tear)) var(--tear-y),
			calc(100% - 24px * var(--right-tear)) calc(var(--tear-y) + 3px),
			calc(100% - 40px * var(--right-tear)) calc(var(--tear-y) + 2px),
			calc(100% - 22px * var(--right-tear)) calc(var(--tear-y) + 9px),
			calc(100% - 14px * var(--right-tear)) calc(var(--tear-y) + 7px),
			100% calc(var(--tear-y) + 10px),
			100% 100%,
			calc(58px * var(--corner)) 100%,
			calc(44px * var(--corner)) calc(100% - 12px * var(--corner)),
			calc(26px * var(--corner)) calc(100% - 8px * var(--corner)),
			calc(12px * var(--corner)) calc(100% - 22px * var(--corner)),
			0 calc(100% - 30px * var(--corner)),
			/* a tear in from the left edge */ 0 calc(var(--tear-y) + 10px),
			calc(14px * var(--left-tear)) calc(var(--tear-y) + 7px),
			calc(22px * var(--left-tear)) calc(var(--tear-y) + 9px),
			calc(40px * var(--left-tear)) calc(var(--tear-y) + 2px),
			calc(24px * var(--left-tear)) calc(var(--tear-y) + 3px),
			calc(13px * var(--left-tear)) var(--tear-y),
			0 calc(var(--tear-y) - 1px)
		);
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
			/* room above and below the text for the nails, and on the sides for
			   the tear */
			padding: 2.1rem 1.8rem 3rem;
		}

		/* the top right nail up by the edge, like the top left one, above the text */
		.nail-top-right {
			top: 14px;
		}

		/* On a narrow paper the tears are smaller, the same shape scaled down to
		   0.7, clear of the text (the top tear above the first line, the side
		   tear beside the text). */
		.notice-paper {
			--torn: polygon(
				0 0,
				var(--tear-x) 0,
				calc(var(--tear-x) + 6px) 11px,
				calc(var(--tear-x) + 4px) 17px,
				calc(var(--tear-x) + 10px) 27px,
				calc(var(--tear-x) + 9px) 15px,
				calc(var(--tear-x) + 12px) 8px,
				calc(var(--tear-x) + 8px) 0,
				calc(100% - 70px * var(--corner)) 0,
				calc(100% - 52px * var(--corner)) calc(10px * var(--corner)),
				calc(100% - 38px * var(--corner)) calc(7px * var(--corner)),
				calc(100% - 24px * var(--corner)) calc(22px * var(--corner)),
				calc(100% - 10px * var(--corner)) calc(26px * var(--corner)),
				100% calc(48px * var(--corner)),
				100% calc(var(--tear-y) - 1px),
				calc(100% - 9px * var(--right-tear)) var(--tear-y),
				calc(100% - 17px * var(--right-tear)) calc(var(--tear-y) + 2px),
				calc(100% - 28px * var(--right-tear)) calc(var(--tear-y) + 1px),
				calc(100% - 15px * var(--right-tear)) calc(var(--tear-y) + 6px),
				calc(100% - 10px * var(--right-tear)) calc(var(--tear-y) + 5px),
				100% calc(var(--tear-y) + 7px),
				100% 100%,
				calc(58px * var(--corner)) 100%,
				calc(44px * var(--corner)) calc(100% - 12px * var(--corner)),
				calc(26px * var(--corner)) calc(100% - 8px * var(--corner)),
				calc(12px * var(--corner)) calc(100% - 22px * var(--corner)),
				0 calc(100% - 30px * var(--corner)),
				0 calc(var(--tear-y) + 7px),
				calc(10px * var(--left-tear)) calc(var(--tear-y) + 5px),
				calc(15px * var(--left-tear)) calc(var(--tear-y) + 6px),
				calc(28px * var(--left-tear)) calc(var(--tear-y) + 1px),
				calc(17px * var(--left-tear)) calc(var(--tear-y) + 2px),
				calc(9px * var(--left-tear)) var(--tear-y),
				0 calc(var(--tear-y) - 1px)
			);
		}
	}
</style>
