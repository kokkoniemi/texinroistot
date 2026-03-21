export type Author = {
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
};

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

export function authorList(authors?: Author[] | null, separator = ', '): string {
	if (!authors || authors.length === 0) return '-';
	return authors
		.map((author) => {
			const base = `${author.firstName} ${author.lastName}`.trim();
			const details = (author.details ?? '').trim();
			if (details) return `${base} (${details})`.trim();
			return base;
		})
		.filter(Boolean)
		.join(separator);
}

export function publicationItem(publication: StoryPublication): string {
	const pub = publication.in;
	if (!pub) return publication.title;

	if (pub.year && pub.issue) return `${pub.year}/${pub.issue}`;
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

	return [...earliestByTitle.values()]
		.sort((a, b) => {
			const order = comparePublicationOrderKey(a.key, b.key);
			if (order !== 0) return order;
			return a.title.localeCompare(b.title);
		})
		.map((entry) => entry.title);
}

function parseBaseSeriesIssue(publication: StoryPublication): BaseSeriesIssue | null {
	const issue = (publication.in?.issue ?? '').trim();
	const year = publication.in?.year ?? 0;
	if (!issue || year <= 0) return null;

	const issueNumber = Number.parseInt(issue.replace(/[^0-9]/g, ''), 10);
	if (Number.isNaN(issueNumber)) return null;

	return { year, issue, issueNumber };
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
			const isContinuous =
				current.year === previous.year && current.issueNumber === previous.issueNumber + 1;
			if (isContinuous) {
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
		kronikka: 'Kronikka (näköispainos)',
		kirjasto: 'Kirjasto'
	};

	const groups: Record<string, StoryPublication[]> = {};
	for (const publication of publications ?? []) {
		const pType = publication.in?.type ?? 'muu_erikois';
		if (!groups[pType]) groups[pType] = [];
		groups[pType].push(publication);
	}

	const order = [
		'perus',
		'kronikka',
		'suur',
		'maxi',
		'muu_erikois',
		'italia_erikois',
		'kirjasto'
	];
	const parts = order.flatMap((type) => {
		const groupedPublications = groups[type] ?? [];
		if (groupedPublications.length === 0) return [];

		const items =
			type === 'perus'
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
