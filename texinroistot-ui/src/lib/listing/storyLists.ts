import { authorWikipediaUrl } from '$lib/data/authorWikipedia';
import type { AuthorStoriesResponse, Story, Villain } from '$lib/types';
import { translatedPartTitles } from './shared';

// What a StoryListPopup shows: an author's or a villain's stories. Kept as a
// value so a story popup can reopen the list when "Edellinen" steps back to it.
export type StoryListSpec = {
	title: string;
	subtitle: string;
	loadStories: () => Promise<Story[]>;
	emptyText?: string;
	storyTitles?: (story: Story) => string[];
	wikipediaUrl?: string;
};

// Context key of the story popup's callback: a story is about to be opened
// from this list, so the popup's history remembers the list.
export const REMEMBER_STORY_LIST = 'rememberStoryList';
export type RememberStoryList = (list: StoryListSpec) => void;

export type AuthorRole = 'writer' | 'drawer' | 'translator';

const roleLabels: Record<AuthorRole, string> = {
	writer: 'Kertoja',
	drawer: 'Piirtäjä',
	translator: 'Suomentaja'
};

export function authorStoryList(name: string, authorHash: string, type: AuthorRole): StoryListSpec {
	return {
		title: name,
		subtitle: roleLabels[type],
		wikipediaUrl: authorWikipediaUrl(name),
		emptyText: 'Tekijälle ei löytynyt tarinoita.',
		storyTitles:
			type === 'translator' ? (story) => translatedPartTitles(story, authorHash) : undefined,
		loadStories: async () => {
			const response = await fetch(
				`/api/tekijat/${encodeURIComponent(authorHash)}/tarinat?type=${encodeURIComponent(type)}`
			);
			if (!response.ok) throw new Error(`Tarinoiden haku epäonnistui (${response.status})`);
			const payload = (await response.json()) as AuthorStoriesResponse;
			return payload.stories ?? [];
		}
	};
}

export function villainStoryList(title: string, villainHash: string): StoryListSpec {
	return {
		title,
		subtitle: 'Roisto',
		loadStories: async () => {
			const response = await fetch(
				`/api/roistot?publication=all&villain=${encodeURIComponent(villainHash)}`
			);
			if (!response.ok) throw new Error(`Tarinoiden haku epäonnistui (${response.status})`);
			const payload = (await response.json()) as { villains?: Villain[] };

			const stories: Story[] = [];
			const seenHashes = new Set<string>();
			for (const appearance of payload.villains?.[0]?.as ?? []) {
				const story = appearance.story;
				if (!story || seenHashes.has(story.hash)) continue;
				seenHashes.add(story.hash);
				stories.push(story);
			}
			return stories;
		}
	};
}
