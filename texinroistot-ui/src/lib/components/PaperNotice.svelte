<script lang="ts">
	// A notice on rough, torn paper, like a poster on a saloon wall: the
	// front page notice and the Info page use it.
	export let maxWidth = '40rem';
</script>

<div class="paper-notice" style:max-width={maxWidth}>
	<!-- The paper: a clipped sheet (torn corners) frayed by the #ragged-edge filter. -->
	<div class="notice-paper" aria-hidden="true"><div class="notice-sheet"></div></div>

	<div class="notice-content">
		<slot />
	</div>
</div>

<style>
	.paper-notice {
		position: relative;
		isolation: isolate;
		margin: 2.5rem auto 0;
		/* the shadow follows the torn, frayed outline of the paper */
		filter: drop-shadow(0 5px 8px var(--shadow));
	}

	.notice-paper {
		position: absolute;
		inset: 0;
		z-index: -1;
		filter: url(#ragged-edge);
	}

	/* Rough, gritty paper with browned edges; chunks torn off two corners and two tears. */
	.notice-sheet {
		position: absolute;
		inset: 0;
		background:
			url('/paper-grit.svg') 0 0 / 420px 420px,
			var(--card-bg);
		box-shadow:
			inset 0 0 40px rgba(120, 85, 35, 0.14),
			inset 0 0 6px rgba(100, 70, 30, 0.25);
		clip-path: polygon(
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

	.notice-content {
		padding: 2rem 2rem 1.8rem;
	}

	@media (max-width: 640px) {
		.paper-notice {
			margin: 1.25rem 0.75rem 0;
		}

		.notice-content {
			padding: 1.5rem 1rem 1.4rem;
		}
	}
</style>
