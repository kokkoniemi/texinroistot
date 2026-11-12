import type { Author, StoryBase } from '$lib/listing/shared';

export type Story = StoryBase & {
	hash: string;
};

export type StoryVillain = {
	hash: string;
	// The name the villain goes by in this story; empty when the story gives
	// none. In a story's villain list the villain's own name is already this.
	firstNames?: string[] | null;
	lastName?: string;
	nicknames?: string[] | null;
	otherNames?: string[] | null;
	codeNames?: string[] | null;
	roles?: string[] | null;
	destiny?: string[] | null;
	story?: Story | null;
};

export type Villain = {
	hash: string;
	ranks?: string[] | null;
	firstNames?: string[] | null;
	lastName?: string | null;
	as?: StoryVillain[] | null;
	// Only set by the story villains endpoint.
	storyCount?: number;
	// The name the villain is listed under in a nickname, other name or code
	// name sort; a villain with several such names is listed under each.
	sortName?: string;
	// Only set by the story villains endpoint, where ranks and names are the
	// ones the story gives: the villain's own name and all of their ranks.
	identity?: {
		ranks?: string[] | null;
		firstNames?: string[] | null;
		lastName?: string | null;
	};
};

export type StoryVillainsResponse = {
	storyHash: string;
	villains: Villain[];
	meta?: {
		total: number;
	};
};

export type ListedAuthor = Author & {
	hash?: string;
	isWriter?: boolean;
	isDrawer?: boolean;
};

export type AuthorStoriesResponse = {
	authorHash: string;
	stories: Story[];
	meta?: {
		total: number;
	};
};

export type AdminUser = {
	hash: string;
	isAdmin: boolean;
	createdAt?: string;
};

export type AdminVersion = {
	id: number;
	createdAt?: string;
	isActive: boolean;
};

// Feedback sent from the Info page.
export type AdminFeedback = {
	id: number;
	createdAt: string;
	message: string;
	// the sender's name or email address ('' when not given)
	contact: string;
};
