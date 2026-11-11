import type { Story } from '$lib/types';

// Wikipedia articles about the real people or gangs behind a whole story,
// shown at the top of the story's villain list. A story is found by its
// Finnish title (without the printing, e.g. "(1. p)"), since story hashes are
// salted per environment.
export type StoryWikipediaLink = { url: string; label: string };

type Entry = StoryWikipediaLink & { titles: string[] };

const entries: Entry[] = [
	{
		titles: ['Carsonin menneisyys', 'Kit Carsonin menneisyys'],
		url: 'https://en.wikipedia.org/wiki/Innocents_(gang)',
		label: 'Ketkä olivat oikeat Viattomat?'
	},
	{
		titles: ['Doc!'],
		url: 'https://en.wikipedia.org/wiki/Cochise_County_Cowboys',
		label: 'Mikä oli oikea Cowboys-kopla?'
	},
	{
		titles: ['Daltonin kopla'],
		url: 'https://fi.wikipedia.org/wiki/Daltonin_veljekset',
		label: 'Mikä oli oikea Daltonin kopla?'
	},
	{
		titles: ['Hurja kopla'],
		url: 'https://en.wikipedia.org/wiki/Butch_Cassidy%27s_Wild_Bunch',
		label: 'Mikä oli oikea Hurja kopla?'
	},
	{
		titles: ['Vegasin lainvalvojat'],
		url: 'https://en.wikipedia.org/wiki/Dodge_City_Gang',
		label: 'Mikä oli oikea Dodge Cityn kopla?'
	},
	{
		titles: ['Missouri!'],
		url: 'https://en.wikipedia.org/wiki/Jayhawker',
		label: 'Keitä olivat oikeat Jayhawkerit?'
	},
	{
		titles: ['Voittamattomat'],
		url: 'https://fi.wikipedia.org/wiki/Feenil%C3%A4isyys',
		label: 'Keitä olivat Irlannin vapaustaistelijat?'
	}
];

function normalizeTitle(title: string): string {
	return title
		.replace(/<\/?i>/g, '')
		.replace(/\s*\([^)]*\)\s*$/, '')
		.replace(/\s+/g, ' ')
		.trim()
		.toLowerCase();
}

const entriesByTitle = new Map<string, Entry>();
for (const entry of entries) {
	for (const title of entry.titles) entriesByTitle.set(normalizeTitle(title), entry);
}

export function storyWikipediaLink(
	story: Story | null | undefined
): StoryWikipediaLink | undefined {
	for (const publication of story?.publications ?? []) {
		const entry = entriesByTitle.get(normalizeTitle(publication.title ?? ''));
		if (entry) return { url: entry.url, label: entry.label };
	}
	return undefined;
}
