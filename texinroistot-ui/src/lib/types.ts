import type { Author, StoryBase } from '$lib/listing/shared';

export type Story = StoryBase & {
	hash: string;
};

export type StoryVillain = {
	hash: string;
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
