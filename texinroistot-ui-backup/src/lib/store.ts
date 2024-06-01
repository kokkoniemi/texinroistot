import { writable } from 'svelte/store';

export const user = writable({
    loggedIn: false,
    email: ''
});

export const appStatus = writable({
    loginInitialized: false
});
