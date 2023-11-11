<script lang="ts">
    import { onMount } from "svelte";
    import { api } from "./api";

    let users: any[];

    onMount(async () => {
        const data = await api.admin.users.list();
        if (data) {
            users = data.users;
        }
    });
</script>

<ul>
    {#if users && !!users.length}
    {#each users as user}
        <li>{user.hash}, created at: {user.createdAt}</li>
    {/each}
    {/if}
</ul>
