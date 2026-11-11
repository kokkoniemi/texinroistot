<script lang="ts">
	// An old weathered board wall behind the pages, in place of the paper
	// background: the paper layers are hidden while the wall shows (the styles
	// below key on body:has(.wood-wall), so they apply from the first paint).
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
			radial-gradient(ellipse at 50% 40%, transparent 55%, rgba(20, 12, 6, 0.35)),
			/* the warm leather brown of the header bar: hue and saturation only
			   (color blend), so the grain keeps its own light and dark */
			linear-gradient(rgba(118, 66, 34, 0.4), rgba(118, 66, 34, 0.4)),
			/* weathered planks: Poly Haven "weathered_planks" (CC0), turned to lie horizontally */
			url('/wood-planks.jpg') 0 0 / 1024px 1024px repeat,
			#3b2416;
		background-blend-mode: normal, color, normal, normal;
	}

	/* the paper background's layers are not shown over the wall */
	:global(body:has(.wood-wall)) {
		background-color: #3b2416;
	}

	:global(body:has(.wood-wall)::before),
	:global(body:has(.wood-wall)::after),
	:global(body:has(.wood-wall) .paper-folds) {
		display: none;
	}

	/* text straight on the wall is light, so it reads on the grey boards */
	:global(body:has(.wood-wall) .site-footer),
	:global(body:has(.wood-wall) .site-footer a),
	:global(body:has(.wood-wall) .on-wall),
	:global(body:has(.wood-wall) .site-main > section > h1) {
		color: #f1ebdd;
		text-shadow: 0 1px 2px rgba(0, 0, 0, 0.6);
	}

	:global(body:has(.wood-wall) .site-footer) {
		border-top-color: rgba(241, 235, 221, 0.35);
	}

	/* Lists' totals, page numbers and letter and year rows sit straight on the
	   wall: light text, links in gold. */
	:global(body:has(.wood-wall) .result-row),
	:global(body:has(.wood-wall) .pagination),
	:global(body:has(.wood-wall) .letter-bar),
	:global(body:has(.wood-wall) .year-bar) {
		color: #f1ebdd;
		text-shadow: 0 1px 2px rgba(0, 0, 0, 0.7);
	}

	:global(body:has(.wood-wall) .result-row a),
	:global(body:has(.wood-wall) .pagination a),
	:global(body:has(.wood-wall) .letter-bar a),
	:global(body:has(.wood-wall) .year-bar a),
	:global(body:has(.wood-wall) .year-bar button) {
		color: #f0c66e;
	}

	:global(body:has(.wood-wall) .result-row .disabled),
	:global(body:has(.wood-wall) .pagination .disabled),
	:global(body:has(.wood-wall) .pagination .ellipsis),
	:global(body:has(.wood-wall) .letter-bar .disabled),
	:global(body:has(.wood-wall) .year-bar .disabled) {
		color: rgba(241, 235, 221, 0.45);
	}

	:global(body:has(.wood-wall) .pagination .current-page) {
		color: #fff8e6;
		font-weight: 700;
	}
</style>
