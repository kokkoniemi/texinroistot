<script lang="ts">
  import router from "page";

  import Home from "./routes/Home.svelte";
  import Manage from "./routes/Manage.svelte";
  import { onMount } from "svelte";
    import { user } from "./lib/store";
    import { api } from "./lib/api";

  let page: ConstructorOfATypedSvelteComponent;
  let params: Record<string, unknown>;

  router("/", () => (page = Home));
  router("/manage", () => (page = Manage));

  router.start();

  type MeResponse = {
    loggedIn: boolean;
    email: string;
  }

  onMount(async () => {
    const data: MeResponse = await api.auth.me.read()
    if (data) {
      user.update(() => data)
    }
  });
</script>

<nav>
  <a href="/">Home</a>
  <a href="/manage">Manage</a>
</nav>

<main>
  <svelte:component this={page} {params} />
</main>
