export type Author = {
	hash?: string;
	firstName: string;
	lastName: string;
	details?: string | null;
};

export type Publication = {
	type: string;
	year: number;
	issue: string;
};

export type StoryPublication = {
	title: string;
	isRepublication?: boolean;
	in?: Publication;
};

export type StoryBase = {
	orderNumber: number;
	writtenBy?: Author[] | null;
	drawnBy?: Author[] | null;
	translatedBy?: Author[] | null;
	publications?: StoryPublication[] | null;
};

export type Meta = {
	total: number;
	page: number;
	pageSize: number;
	totalPages: number;
	// Initial letters available for letter paging; empty for non-alphabetical sorts.
	letters?: string[];
	// Years available for year paging in date sorts; empty otherwise.
	years?: number[];
	// After a letter jump: 0-based index of the letter's first item on the page.
	letterIndex?: number | null;
};

// Scrolls the item at index (a child of list) to the top of the view; used
// after a letter or year jump so its first item starts the visible list. When
// the item is near the end of the page, space is added below the list so it
// can still reach the top; the space is removed on the next call.
export function scrollToListItem(list: HTMLElement | undefined, index?: number | null): void {
	if (!list) return;
	list.style.paddingBottom = '';
	if (index === null || index === undefined) return;
	const item = list.children[index];
	if (!item) return;

	const itemTop = item.getBoundingClientRect().top + window.scrollY;
	const missing = itemTop + window.innerHeight - document.documentElement.scrollHeight;
	if (missing > 0) list.style.paddingBottom = `${Math.ceil(missing)}px`;
	item.scrollIntoView({ block: 'start' });
}

export type PaginationToken = number | 'ellipsis';

type BaseSeriesIssue = {
	year: number;
	issue: string;
	issueNumber: number;
};

type PublicationOrderKey = {
	year: number;
	issueNumber: number;
	issue: string;
	index: number;
};

export function joinValues(values?: string[] | null, fallback = '-', separator = ', '): string {
	if (!values || values.length === 0) return fallback;
	return values.filter(Boolean).join(separator);
}

export function hasValues(values?: string[] | null): boolean {
	return Boolean(values && values.some((value) => Boolean(value)));
}

export type StoryTitleSegment = {
	text: string;
	isAnnotation: boolean;
};

const storyTitleAnnotationPattern = /\(\d+\.(?:\s*[–-]\s*\d+\.)?\s*p\)|;/g;

export type TextSegment = {
	text: string;
	italic: boolean;
};

// Splits text with the importer's <i>…</i> italic markup (used in Rooli and
// Kohtalo) into plain and italic segments. Nothing else is treated as markup.
export function italicSegments(text: string): TextSegment[] {
	const segments: TextSegment[] = [];
	const pattern = /<i>(.*?)<\/i>/g;
	let lastIndex = 0;
	for (const match of text.matchAll(pattern)) {
		const index = match.index ?? 0;
		if (index > lastIndex) segments.push({ text: text.slice(lastIndex, index), italic: false });
		if (match[1]) segments.push({ text: match[1], italic: true });
		lastIndex = index + match[0].length;
	}
	if (lastIndex < text.length) segments.push({ text: text.slice(lastIndex), italic: false });
	return segments;
}

const storyTitlePartNotePattern = /\s*\(\d+\.(?:\s*[–-]\s*\d+\.)?\s*p\)/g;

// The title of the story's first Finnish edition without part notes, e.g.
// "Taistelu erämaassa" instead of "Taistelu erämaassa (1. – 2. p); ...".
export function firstEditionTitle(publications?: StoryPublication[] | null): string {
	const [firstTitle] = nonItalianTitlesByFirstPublication(publications);
	return (firstTitle ?? '').replace(storyTitlePartNotePattern, '').trim();
}

// Splits a story title into text and annotations (part notes such as "(1. p)"
// or "(1. – 2. p)", and the ";" between titles), so the annotations can be
// styled separately from the title text.
export function storyTitleSegments(title: string): StoryTitleSegment[] {
	const segments: StoryTitleSegment[] = [];
	let lastIndex = 0;
	for (const match of title.matchAll(storyTitleAnnotationPattern)) {
		const index = match.index ?? 0;
		if (index > lastIndex) {
			segments.push({ text: title.slice(lastIndex, index), isAnnotation: false });
		}
		segments.push({ text: match[0], isAnnotation: true });
		lastIndex = index + match[0].length;
	}
	if (lastIndex < title.length) {
		segments.push({ text: title.slice(lastIndex), isAnnotation: false });
	}
	return segments;
}

export function publicationItem(publication: StoryPublication): string {
	const pub = publication.in;
	if (!pub) return publication.title;

	if (pub.year && pub.issue) return `${pub.issue}/${pub.year}`;
	if (pub.issue) return pub.issue;
	if (pub.year) return `${pub.year}`;
	return publication.title;
}

function publicationOrderKey(publication: StoryPublication, index: number): PublicationOrderKey {
	const year = publication.in?.year ?? 0;
	const issue = (publication.in?.issue ?? '').trim();
	const parsedIssueNumber = Number.parseInt(issue.replace(/[^0-9]/g, ''), 10);
	return {
		year: year > 0 ? year : Number.MAX_SAFE_INTEGER,
		issueNumber: Number.isNaN(parsedIssueNumber) ? Number.MAX_SAFE_INTEGER : parsedIssueNumber,
		issue,
		index
	};
}

function comparePublicationOrderKey(a: PublicationOrderKey, b: PublicationOrderKey): number {
	if (a.year !== b.year) return a.year - b.year;
	if (a.issueNumber !== b.issueNumber) return a.issueNumber - b.issueNumber;
	const issueCompare = a.issue.localeCompare(b.issue);
	if (issueCompare !== 0) return issueCompare;
	return a.index - b.index;
}

export function nonItalianTitlesByFirstPublication(
	publications?: StoryPublication[] | null
): string[] {
	const earliestByTitle = new Map<string, { title: string; key: PublicationOrderKey }>();

	for (const [index, publication] of (publications ?? []).entries()) {
		if (publication.in?.type?.startsWith('italia_')) continue;

		const title = publication.title.trim();
		if (!title) continue;

		const key = publicationOrderKey(publication, index);
		const existing = earliestByTitle.get(title);
		if (!existing || comparePublicationOrderKey(key, existing.key) < 0) {
			earliestByTitle.set(title, { title, key });
		}
	}

	const titles = [...earliestByTitle.values()]
		.sort((a, b) => {
			const order = comparePublicationOrderKey(a.key, b.key);
			if (order !== 0) return order;
			return a.title.localeCompare(b.title);
		})
		.map((entry) => entry.title);

	// When every title has a part note, order by the first part number
	// ("(1. p)", "(2. – 3. p)", ...) instead of publication date.
	const partNumbers = titles.map(firstPartNumber);
	if (titles.length > 1 && partNumbers.every((part) => part !== null)) {
		return titles
			.map((title, index) => ({ title, part: partNumbers[index] as number }))
			.sort((a, b) => a.part - b.part)
			.map((entry) => entry.title);
	}
	return titles;
}

function firstPartNumber(title: string): number | null {
	const match = title.match(/\((\d+)\.(?:\s*[–-]\s*\d+\.)?\s*p\)/);
	return match ? Number(match[1]) : null;
}

// The part range of a part note such as "1. p" or "2. – 3. p".
function partRange(text: string): [number, number] | null {
	const match = text.match(/(\d+)\.(?:\s*[–-]\s*(\d+)\.)?\s*p\b/);
	if (!match) return null;
	const from = Number(match[1]);
	return [from, match[2] ? Number(match[2]) : from];
}

// The story's titles for the parts the translator translated, from the
// translator's part note, e.g. "Taistelu erämaassa (1. – 2. p)" for a "1. – 2. p"
// translator. Titles without part notes get the translator's note appended.
// Without a translator note (or a matching title) all titles are returned.
export function translatedPartTitles(
	story: Pick<StoryBase, 'publications' | 'translatedBy'>,
	translatorHash: string
): string[] {
	const titles = nonItalianTitlesByFirstPublication(story.publications);
	const translatorDetails = (story.translatedBy ?? [])
		.filter((author) => (author.hash ?? '').trim() === translatorHash)
		.map((author) => (author.details ?? '').trim());
	const ranges = translatorDetails
		.map(partRange)
		.filter((range): range is [number, number] => range !== null);
	if (ranges.length === 0) return titles;

	// Titles without part notes (e.g. "Nogalesista etelään"): add the
	// translator's note instead, e.g. "Nogalesista etelään (1. p)".
	if (titles.every((title) => partRange(title) === null)) {
		const note = translatorDetails.filter((details) => partRange(details) !== null).join(', ');
		return titles.map((title) => `${title} (${note})`);
	}

	const matching = titles.filter((title) => {
		const range = partRange(title);
		return range !== null && ranges.some(([from, to]) => range[0] <= to && range[1] >= from);
	});
	return matching.length > 0 ? matching : titles;
}

function parseBaseSeriesIssue(publication: StoryPublication): BaseSeriesIssue | null {
	const issue = (publication.in?.issue ?? '').trim();
	const year = publication.in?.year ?? 0;
	if (!issue || year <= 0) return null;

	const issueNumber = Number.parseInt(issue.replace(/[^0-9]/g, ''), 10);
	if (Number.isNaN(issueNumber)) return null;

	return { year, issue, issueNumber };
}

// Issues per year in the Finnish base series; keep in sync with
// getPublishedAnnualCount in texinroistot-server/internal/importer/publications.go.
function baseSeriesAnnualCount(year: number): number | null {
	if (year === 1953) return 25;
	if (year >= 1954 && year <= 1964) return 26;
	if (year === 1965) return 27;
	if (year >= 1971 && year <= 1978) return 12;
	if (year === 1979) return 13;
	if (year >= 1980) return 16;
	return null;
}

// No base series issues were published in 1966–1970.
function nextBaseSeriesYear(year: number): number {
	return year === 1965 ? 1971 : year + 1;
}

function isNextBaseSeriesIssue(previous: BaseSeriesIssue, current: BaseSeriesIssue): boolean {
	if (current.year === previous.year) {
		return current.issueNumber === previous.issueNumber + 1;
	}
	return (
		current.year === nextBaseSeriesYear(previous.year) &&
		current.issueNumber === 1 &&
		previous.issueNumber === baseSeriesAnnualCount(previous.year)
	);
}

function formatBaseSeriesRange(start: BaseSeriesIssue, end: BaseSeriesIssue): string {
	if (start.year === end.year && start.issueNumber === end.issueNumber) {
		return `${start.issue}/${start.year}`;
	}
	return `${start.issue}/${start.year}–${end.issue}/${end.year}`;
}

function baseSeriesSummary(publications: StoryPublication[]): string {
	const parsedIssues: BaseSeriesIssue[] = [];
	const fallbackItems: string[] = [];

	for (const publication of publications) {
		const parsed = parseBaseSeriesIssue(publication);
		if (parsed) {
			parsedIssues.push(parsed);
			continue;
		}
		const item = publicationItem(publication).trim();
		if (item) fallbackItems.push(item);
	}

	const dedupedParsed = new Map<string, BaseSeriesIssue>();
	for (const issue of parsedIssues) {
		const key = `${issue.year}:${issue.issueNumber}`;
		if (!dedupedParsed.has(key)) dedupedParsed.set(key, issue);
	}
	const orderedIssues = [...dedupedParsed.values()].sort((a, b) => {
		if (a.year !== b.year) return a.year - b.year;
		if (a.issueNumber !== b.issueNumber) return a.issueNumber - b.issueNumber;
		return a.issue.localeCompare(b.issue);
	});

	const ranges: string[] = [];
	if (orderedIssues.length > 0) {
		let start = orderedIssues[0];
		let previous = orderedIssues[0];

		for (let i = 1; i < orderedIssues.length; i++) {
			const current = orderedIssues[i];
			if (isNextBaseSeriesIssue(previous, current)) {
				previous = current;
				continue;
			}
			ranges.push(formatBaseSeriesRange(start, previous));
			start = current;
			previous = current;
		}

		ranges.push(formatBaseSeriesRange(start, previous));
	}

	const uniqueFallbackItems = fallbackItems.filter(
		(item, index, values) => values.indexOf(item) === index
	);
	const items = [...ranges, ...uniqueFallbackItems];
	return items.join(', ');
}

export function publicationSummaryFromPublications(
	publications?: StoryPublication[] | null,
	emptyText = '-'
): string {
	const sectionLabels: Record<string, string> = {
		kronikka: 'Kronikka (näk. p.)',
		kirjasto: 'Kirjasto'
	};

	// Base series re-publications (UVuosi/Ualkaen/Upäättyen) are listed after
	// special publications instead of next to the original base series issues.
	const republicationGroup = 'perus_uusinta';

	const groups: Record<string, StoryPublication[]> = {};
	for (const publication of publications ?? []) {
		const pType = publication.in?.type ?? 'muu_erikois';
		const group = pType === 'perus' && publication.isRepublication ? republicationGroup : pType;
		if (!groups[group]) groups[group] = [];
		groups[group].push(publication);
	}

	// Finnish publications only; Italian ones (italia_*) are shown separately as
	// the Italian original publication.
	const order = [
		'perus',
		'kronikka',
		'suur',
		'maxi',
		'muu_erikois',
		republicationGroup,
		'kirjasto'
	];
	const parts = order.flatMap((type) => {
		const groupedPublications = groups[type] ?? [];
		if (groupedPublications.length === 0) return [];

		const items =
			type === 'perus' || type === republicationGroup
				? baseSeriesSummary(groupedPublications)
				: groupedPublications
						.map((publication) => publicationItem(publication).trim())
						.filter((item, index, values) => Boolean(item) && values.indexOf(item) === index)
						.join(', ');

		if (!items) return [];
		const sectionLabel = sectionLabels[type];
		return [sectionLabel ? `${sectionLabel} ${items}` : items];
	});

	return parts.length > 0 ? parts.join('; ') : emptyText;
}

export function paginationTokens(currentPage: number, totalPages: number): PaginationToken[] {
	if (totalPages <= 0) return [];

	const visiblePages = new Set<number>([1, totalPages]);
	for (let page = currentPage - 1; page <= currentPage + 1; page++) {
		if (page >= 1 && page <= totalPages) visiblePages.add(page);
	}
	if (currentPage <= 3) {
		visiblePages.add(2);
		visiblePages.add(3);
	}
	if (currentPage >= totalPages - 2) {
		visiblePages.add(totalPages - 1);
		visiblePages.add(totalPages - 2);
	}

	const orderedPages = [...visiblePages]
		.filter((page) => page >= 1 && page <= totalPages)
		.sort((a, b) => a - b);
	const tokens: PaginationToken[] = [];
	let previousPage = 0;
	for (const page of orderedPages) {
		if (previousPage > 0 && page - previousPage > 1) {
			tokens.push('ellipsis');
		}
		tokens.push(page);
		previousPage = page;
	}
	return tokens;
}

export type ItalianOriginalPublication = {
	title: string;
	details: string;
};

export function italianOriginalPublication(story: StoryBase): ItalianOriginalPublication | null {
	const italianPublications = (story.publications ?? []).filter((publication) =>
		publication.in?.type?.startsWith('italia_')
	);
	if (italianPublications.length === 0) return null;

	const titles = italianPublications
		.map((publication) => publication.title.trim())
		.filter((title, index, values) => Boolean(title) && values.indexOf(title) === index);
	const titlePart = titles.join('; ');

	const issues = italianPublications
		.filter((publication) => publication.in?.type === 'italia_perus')
		.map((publication) => {
			const issue = (publication.in?.issue ?? '').trim();
			const year = publication.in?.year ?? 0;
			const sortIssue = Number.parseInt(issue.replace(/\D/g, ''), 10);
			return {
				issue,
				year,
				sortIssue: Number.isNaN(sortIssue) ? Number.MAX_SAFE_INTEGER : sortIssue
			};
		})
		.filter((entry) => entry.issue && entry.year > 0)
		.filter(
			(entry, index, values) =>
				values.findIndex((other) => other.issue === entry.issue && other.year === entry.year) ===
				index
		)
		.sort((a, b) => {
			if (a.year !== b.year) return a.year - b.year;
			if (a.sortIssue !== b.sortIssue) return a.sortIssue - b.sortIssue;
			return a.issue.localeCompare(b.issue);
		});

	// Italian issue numbers form one continuous series, so the year is shown
	// once after the issues, e.g. "1-2 (1948)" or "25-27 (1949-1950)".
	let issuePart = '';
	if (issues.length === 1) {
		issuePart = `${issues[0].issue} (${issues[0].year})`;
	} else if (issues.length > 1) {
		const first = issues[0];
		const last = issues[issues.length - 1];
		const yearPart = first.year === last.year ? `${first.year}` : `${first.year}-${last.year}`;
		issuePart = `${first.issue}-${last.issue} (${yearPart})`;
	}

	// Italian special publications (Maxi Tex, Texone, Color Tex, ...): the issue
	// text already names the series, so only the year is added, e.g.
	// "Maxi Tex 16 (2013)".
	const specialIssues = italianPublications
		.filter((publication) => publication.in?.type !== 'italia_perus')
		.map((publication) => {
			const issue = (publication.in?.issue ?? '').trim();
			const year = publication.in?.year ?? 0;
			if (!issue) return publicationItem(publication).trim();
			return year > 0 ? `${issue} (${year})` : issue;
		})
		.filter((item, index, values) => Boolean(item) && values.indexOf(item) === index);

	let details = [issuePart, ...specialIssues].filter(Boolean).join('; ');
	if (story.orderNumber > 0) {
		const storyNumberPart = `tarina nro ${story.orderNumber}`;
		details = details ? `${details}, ${storyNumberPart}` : storyNumberPart;
	}

	if (!titlePart && !details) return null;
	return { title: titlePart, details };
}

export function storyVillainForStory<T extends { story?: { hash?: string | null } | null }>(
	villain: { as?: T[] | null },
	storyHash: string
): T | null {
	const appearances = villain.as ?? [];
	const matchingStory = appearances.find(
		(appearance) => (appearance.story?.hash ?? '').trim() === storyHash
	);
	if (matchingStory) return matchingStory;
	return appearances.length > 0 ? appearances[0] : null;
}

type VillainForTitle = {
	ranks?: string[] | null;
	firstNames?: string[] | null;
	lastName?: string | null;
	as?: Array<{
		nicknames?: string[] | null;
		otherNames?: string[] | null;
		codeNames?: string[] | null;
		story?: { hash?: string | null } | null;
	}> | null;
};

export function storyVillainTitle(villain: VillainForTitle, storyHash: string): string {
	const firstNames = joinValues(villain.firstNames, '').trim();
	const lastName = (villain.lastName ?? '').trim();
	const realName = `${firstNames} ${lastName}`.trim();

	const appearance = storyVillainForStory(villain, storyHash);
	const uniqueTrimmed = (values?: string[] | null) =>
		(values ?? []).map((v) => v.trim()).filter((v, i, arr) => Boolean(v) && arr.indexOf(v) === i);
	const otherNames = uniqueTrimmed(appearance?.otherNames);
	const nicknames = uniqueTrimmed(appearance?.nicknames);
	const codeNames = uniqueTrimmed(appearance?.codeNames);
	const quotedNicknames = nicknames.map((n) => `"${n}"`);

	// The rank goes in front of the name, e.g. "Don Esteban de Queretaro".
	const rank = joinValues(villain.ranks, '').trim();
	const withRank = (name: string) => (rank ? `${rank} ${name}` : name);

	// Real name first, then other names (Excel column D), then nicknames (column E).
	const names = [realName, ...otherNames, ...quotedNicknames].filter(Boolean);
	if (names.length > 0) return [withRank(names[0]), ...names.slice(1)].join(', ');
	if (codeNames.length > 0) return withRank(codeNames.join('; '));
	if (rank) return rank;
	return 'Nimetön roisto';
}

// A villain that a role or destiny text can mention: key identifies the villain
// (its hash), names are the names it can be mentioned by.
export type MentionTarget = { key: string; names: string[] };
export type MentionSegment = TextSegment & { mention?: string };

const MIN_MENTION_NAME_LENGTH = 3;

// The names a villain can be mentioned by in a story: each first name with the
// last name, the last name alone, and the story's nicknames, other names and
// code names. Quoted parts of first names (Raymond "Ray") are left out.
export function villainMentionNames(villain: VillainForTitle, storyHash: string): string[] {
	const appearance = storyVillainForStory(villain, storyHash);
	const trimmed = (values?: string[] | null) =>
		(values ?? []).map((value) => value.replace(/<\/?i>/g, '').trim()).filter(Boolean);
	// "Mickey (Miky)" and "Quantrill (Quantrell)" can be mentioned by either spelling.
	const spellings = (name: string): string[] => {
		const match = name.match(/^(.*?)\s*\(([^)]*)\)\s*$/);
		if (!match) return [name];
		return [match[1], ...match[2].split(',')].map((part) => part.trim()).filter(Boolean);
	};
	const firstNames = trimmed(villain.firstNames)
		.map((name) =>
			name
				.replace(/"[^"]*"/g, '')
				.replace(/\s+/g, ' ')
				.trim()
		)
		.filter(Boolean)
		.flatMap(spellings);
	const lastNames = spellings((villain.lastName ?? '').trim()).filter(Boolean);
	// The rank is part of the name in a text: "Don Felipen", "Kapteeni Garridon";
	// it may also be written in lower case ("don Felipen").
	const ranks = trimmed(villain.ranks).flatMap((rank) => {
		const lower = rank.charAt(0).toLowerCase() + rank.slice(1);
		return lower === rank ? [rank] : [rank, lower];
	});

	const names = new Set<string>();
	const addWithRanks = (name: string) => {
		names.add(name);
		for (const rank of ranks) names.add(`${rank} ${name}`);
	};
	for (const firstName of firstNames) {
		for (const lastName of lastNames) addWithRanks(`${firstName} ${lastName}`);
		// The first name alone, also with a rank ("Don Felipen" for Don Felipe
		// Fuente Madariaga y Castillo). mentionSegments links it only when no
		// other villain of the story has the same name.
		addWithRanks(firstName);
	}
	for (const lastName of lastNames) addWithRanks(lastName);
	for (const name of trimmed(appearance?.nicknames)) names.add(name);
	for (const name of trimmed(appearance?.otherNames)) names.add(name);
	for (const name of trimmed(appearance?.codeNames)) names.add(name);
	return [...names].filter((name) => name.length >= MIN_MENTION_NAME_LENGTH);
}

export function storyMentionTargets(
	villains: Array<VillainForTitle & { hash: string }>,
	storyHash: string
): MentionTarget[] {
	return villains.map((villain) => ({
		key: villain.hash,
		names: villainMentionNames(villain, storyHash)
	}));
}

function escapeRegExp(value: string): string {
	return value.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
}

// Splits a role or destiny text into italic/plain segments and marks the
// mentions of other villains of the story. A name may carry a Finnish case
// ending ("Pattermanille", "Ducaux'n"); the whole inflected word is the
// mention. Longer names win over shorter ones they contain, and a name shared
// by several villains is not linked, since it could mean either.
export function mentionSegments(
	text: string,
	targets: MentionTarget[],
	selfKey: string
): MentionSegment[] {
	const owners = new Map<string, Set<string>>();
	for (const target of targets) {
		if (target.key === selfKey) continue;
		for (const name of target.names) {
			const owner = owners.get(name) ?? new Set<string>();
			owner.add(target.key);
			owners.set(name, owner);
		}
	}
	const names = [...owners.entries()]
		.filter(([, keys]) => keys.size === 1)
		.map(([name, keys]) => ({ name, key: [...keys][0] }))
		.sort((a, b) => b.name.length - a.name.length);

	const segments: MentionSegment[] = [];
	for (const segment of italicSegments(text)) {
		const taken: Array<{ start: number; end: number; key: string }> = [];
		for (const { name, key } of names) {
			const pattern = new RegExp(
				`(^|[^\\p{L}])(${escapeRegExp(name)}(?:'?\\p{Ll}{0,4})?)(?![\\p{L}])`,
				'gu'
			);
			for (const match of segment.text.matchAll(pattern)) {
				const start = (match.index ?? 0) + match[1].length;
				const end = start + match[2].length;
				if (taken.some((range) => start < range.end && end > range.start)) continue;
				taken.push({ start, end, key });
			}
		}
		taken.sort((a, b) => a.start - b.start);
		let position = 0;
		for (const range of taken) {
			if (range.start > position) {
				segments.push({ text: segment.text.slice(position, range.start), italic: segment.italic });
			}
			segments.push({
				text: segment.text.slice(range.start, range.end),
				italic: segment.italic,
				mention: range.key
			});
			position = range.end;
		}
		if (position < segment.text.length) {
			segments.push({ text: segment.text.slice(position), italic: segment.italic });
		}
	}
	return segments;
}

// Opens an external page (e.g. an author's Wikipedia article) in a window of
// its own beside the site. Used as a link's click handler; the link keeps
// href and target="_blank", so where a separate window cannot be opened (or
// on phones) the page opens in a new tab instead.
export function openInOwnWindow(event: MouseEvent): void {
	const link = event.currentTarget as HTMLAnchorElement | null;
	if (!link || event.ctrlKey || event.metaKey || event.shiftKey || event.button !== 0) return;
	// phones and tablets have no separate windows: keep the new tab
	if (window.matchMedia('(pointer: coarse)').matches) return;
	const width = Math.min(1100, window.screen.availWidth - 80);
	const height = Math.min(900, window.screen.availHeight - 80);
	const opened = window.open(
		link.href,
		'texinroistot-wikipedia',
		`popup,width=${width},height=${height}`
	);
	if (!opened) return; // blocked: the link opens a new tab instead
	opened.opener = null;
	opened.focus();
	event.preventDefault();
}

// Brings a mentioned villain's card (VillainCard with villainKey) into view in
// a story's villain list and highlights it for a moment.
export function showMentionedVillain(container: HTMLElement | undefined, key: string): void {
	const card = container?.querySelector<HTMLElement>(`[data-villain-key="${CSS.escape(key)}"]`);
	if (!card) return;
	card.scrollIntoView({ behavior: 'smooth', block: 'center' });
	card.classList.remove('mention-highlight');
	// restart the highlight animation when the same card is shown again
	void card.offsetWidth;
	card.classList.add('mention-highlight');
	window.setTimeout(() => card.classList.remove('mention-highlight'), 2000);
}

// Villains appearing in several stories get a "Näytä roiston muut tarinat"
// popup; returns the villain hash for those and undefined for the rest.
export function otherStoriesVillainHash(villain: {
	hash: string;
	storyCount?: number;
}): string | undefined {
	return (villain.storyCount ?? 0) > 1 ? villain.hash : undefined;
}

export function buildPageHref(
	pathname: string,
	params: Record<string, string | number | undefined | null>
): string {
	const searchParams = new URLSearchParams();
	for (const [key, value] of Object.entries(params)) {
		if (value === undefined || value === null || value === '') continue;
		searchParams.set(key, String(value));
	}

	const query = searchParams.toString();
	return query ? `${pathname}?${query}` : pathname;
}
