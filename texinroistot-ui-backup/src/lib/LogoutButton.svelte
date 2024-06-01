<script context="module" lang="ts">
    declare var window: {
        google: any;
    };
</script>

<script lang="ts">
    import { api } from "./api";

    export let email: string;

    function logout(e: Event) {
        e.preventDefault();
        window.google.accounts.id.revoke(email, async () => {
            await api.auth.logout.post()
            document.location.reload();
        });
    }
</script>

<button on:click={logout}>Logout</button>
