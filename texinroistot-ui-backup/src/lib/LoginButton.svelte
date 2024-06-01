<script context="module" lang="ts">
    declare var window: {
        google: any;
    };
</script>

<script lang="ts">
    import { onMount } from "svelte";

    import { appStatus } from "../lib/store";
    let readyToLogin: boolean;

    appStatus.subscribe((value) => {
        readyToLogin = value.loginInitialized;
    });

    onMount(async () => {
        while (!readyToLogin) {
            await new Promise(r => setTimeout(r, 50));
        }
        
        window.google.accounts.id.renderButton(
            document.getElementById("loginbutton"),
            {
                type: "standard",
                width: 300,
            }
        );
    })
</script>


<div id="loginbutton" />
