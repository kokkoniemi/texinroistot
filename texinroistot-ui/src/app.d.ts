// See https://kit.svelte.dev/docs/types#app
// for information about these interfaces
declare global {
	namespace App {
		// interface Error {}
		// interface Locals {}
		// interface PageData {}
		// Shallow routing: how many popup steps (story cards, story lists) are
		// open, so the browser's back (a swipe on a phone) steps back through them.
		interface PageState {
			popupDepth?: number;
			// The page load the depth belongs to (entries from before a reload are stale).
			popupSession?: string;
		}
		// interface Platform {}
	}
}

export {};
