<script lang="ts">
    import LoginButton from "../lib/LoginButton.svelte";
    import LogoutButton from "../lib/LogoutButton.svelte";
    import { user } from "../lib/store";
    let userVal: any;

    user.subscribe((value) => {
        userVal = value;
    });
</script>

<h1>Manage</h1>

{#if !userVal.loggedIn}
    <LoginButton />
{:else}
    <LogoutButton email="{userVal.email}"/>
{/if}
