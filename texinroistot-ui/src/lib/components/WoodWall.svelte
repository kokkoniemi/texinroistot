<script lang="ts">
	import { onMount } from 'svelte';

	// An old weathered board wall behind the page (front page and Info), in place of
	// the paper background: the paper layers are hidden while the wall shows.
	onMount(() => {
		document.body.classList.add('on-wood-wall');
		return () => document.body.classList.remove('on-wood-wall');
	});
</script>

<div class="wood-wall" aria-hidden="true"></div>

<style>
	.wood-wall {
		position: fixed;
		inset: 0;
		z-index: -1;
		pointer-events: none;
		background:
			/* darker towards the edges, as in the shade of a porch roof */
			radial-gradient(ellipse at 50% 40%, transparent 55%, rgba(20, 16, 12, 0.35)),
			/* weathered planks: Poly Haven "weathered_planks" (CC0), turned to lie horizontally */
			url('/wood-planks.jpg') 0 0 / 1024px 1024px repeat,
			#3b2f25;
	}

	/* the paper background's layers are not shown over the wall */
	:global(body.on-wood-wall) {
		background-color: #3b2f25;
	}

	:global(body.on-wood-wall::before),
	:global(body.on-wood-wall::after),
	:global(body.on-wood-wall .paper-folds) {
		display: none;
	}

	/* text straight on the wall is light, so it reads on the grey boards */
	:global(body.on-wood-wall .site-footer),
	:global(body.on-wood-wall .site-footer a),
	:global(body.on-wood-wall .on-wall) {
		color: #f1ebdd;
		text-shadow: 0 1px 2px rgba(0, 0, 0, 0.6);
	}

	:global(body.on-wood-wall .site-footer) {
		border-top-color: rgba(241, 235, 221, 0.35);
	}
</style>
