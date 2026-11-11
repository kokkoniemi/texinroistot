import { pushState } from '$app/navigation';

// Popup steps in the browser history, so that the browser's back (a swipe or
// the back key on a phone) works like the popups' "Edellinen" button.
//
// Each step (a story popup opened, another story shown in it, a story list
// opened) adds a history entry with its depth in page.state, and a handler
// that undoes it. Going back in the browser lowers the depth; the layout
// passes the new depth to syncPopupDepth, which runs the handlers of the steps
// above it, last first. The popups' own back and close buttons go back in the
// browser history too, so all ways back take the same path.

type Step = { depth: number; onBack: () => void };

// History entries left from before a reload carry a depth with no popups open;
// only this page load's entries count.
const session = Math.random().toString(36).slice(2);

const steps: Step[] = [];
let currentDepth = 0;
// A story list closing because a story was chosen from it hands its history
// entry over to the story card opened next, so one back undoes the choice.
let handedOver: Step | null = null;

// Adds a step and returns its depth.
export function pushPopupStep(onBack: () => void): number {
	if (typeof window === 'undefined') return 0;
	if (handedOver && steps[steps.length - 1] === handedOver) {
		handedOver.onBack = onBack;
		const depth = handedOver.depth;
		handedOver = null;
		return depth;
	}
	handedOver = null;
	const depth = currentDepth + 1;
	steps.push({ depth, onBack });
	currentDepth = depth;
	pushState('', { popupDepth: depth, popupSession: session });
	return depth;
}

// The step at this depth is ending because something replaces it (a story
// chosen from a list); the next pushPopupStep reuses its history entry.
export function handOverPopupStep(depth: number): void {
	const step = steps.find((candidate) => candidate.depth === depth);
	if (step && steps[steps.length - 1] === step) {
		step.onBack = () => {};
		handedOver = step;
	}
}

// One step back, as the browser's back would.
export function popupBack(): void {
	if (currentDepth > 0) history.back();
}

// Back to just below the given depth: closes that step and everything above it.
export function popupBackBelow(depth: number): void {
	const count = currentDepth - (depth - 1);
	if (count > 0) history.go(-count);
}

// Steps are undone only when the browser really went back (popstate): page.state
// can lag behind, and a late value from an earlier push is not a step back.
let wentBack = false;
if (typeof window !== 'undefined') {
	window.addEventListener('popstate', () => (wentBack = true));
}

// Called by the layout whenever page.state changes.
export function syncPopupDepth(state: App.PageState): void {
	const depth = state.popupSession === session ? (state.popupDepth ?? 0) : 0;
	if (depth >= currentDepth) {
		currentDepth = depth;
		return;
	}
	if (!wentBack) {
		// A navigation to another page clears the state; its popups are gone.
		if (depth === 0) {
			steps.length = 0;
			handedOver = null;
			currentDepth = 0;
		}
		return;
	}
	wentBack = false;
	currentDepth = depth;
	const undone: Step[] = [];
	while (steps.length > 0 && steps[steps.length - 1].depth > depth) {
		const step = steps.pop() as Step;
		if (step === handedOver) handedOver = null;
		undone.push(step);
	}
	// Undo after the router has finished with the popstate: a step that adds a
	// new entry (a story list reopened) would otherwise replace the current one.
	setTimeout(() => {
		for (const step of undone) step.onBack();
	});
}
