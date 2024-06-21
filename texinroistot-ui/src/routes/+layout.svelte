<script>
	import '../app.css';
	import { page } from '$app/stores';
</script>

<ul class="top-menu">
	<li class:active={$page.url.pathname === '/'}>
		<a href="/">Etusivu</a>
	</li>
	<li class:active={$page.url.pathname === '/roistot'}>
		<a href="/roistot">Roistot</a>
	</li>
	<li class:active={$page.url.pathname === '/tarinat'}>
		<a href="/tarinat">Tarinat</a>
	</li>
</ul>

<slot />

<hr />

<p>Viimeisin päivitys: ....</p>
<p>Kaikki oikeudet pidätetään | <a href="/hallinta">Hallinta</a></p>

<style>
	.top-menu {
		list-style-type: none;
		margin: 0;
		padding: 0 0 0 20px;
		display: block;
		border-bottom: 1px solid black;
	}
	.top-menu li {
		display: inline-block;
		text-align: center;
		width: 100px;
		padding: 10px 0;
		border-width: 1px 1px 0 1px;
		border-color: black;
		border-style: solid;
		margin-bottom: -1px;
	}
	.top-menu li.active {
		border-bottom: 1px solid #ffffed;
		font-weight: bold;
	}
	.top-menu li a,
	.top-menu li a:active,
	.top-menu li a:visited {
		color: black;
		text-decoration: none;
	}
</style>
