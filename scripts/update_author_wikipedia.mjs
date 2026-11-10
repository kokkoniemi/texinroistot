#!/usr/bin/env node
// Builds texinroistot-ui/src/lib/data/authorWikipedia.ts: a Wikipedia link for
// each author (writer, drawer, translator) whose article can be found and is
// about comics or translation, so a namesake's article is not linked.
//
// Author names come from a running backend (default http://localhost:6969).
// Articles are looked up by exact title (following redirects) in Finnish,
// Italian and English Wikipedia, preferring them in that order, and kept only
// when the article intro mentions comics or translation work.
//
// Usage (from the repository root, with the local stack running), then format
// the generated file with the UI's Prettier settings:
//   node scripts/update_author_wikipedia.mjs [backend-url]
//   docker compose exec frontend npx prettier --write src/lib/data/authorWikipedia.ts

import { writeFileSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';

const backend = process.argv[2] ?? 'http://localhost:6969';
const root = join(dirname(fileURLToPath(import.meta.url)), '..');
const output = join(root, 'texinroistot-ui/src/lib/data/authorWikipedia.ts');
const languages = ['fi', 'it', 'en'];
const userAgent = 'texinroistot-author-links/1.0 (https://github.com/kokkoniemi/texinroistot)';
const aboutComics =
	/fumett|tex willer|bonelli|sarjakuv|comic|cartoon|disegnat|sceneggiat|illustrat|kääntäj|suomentaj/i;

// Articles confirmed by hand whose intro the check above does not recognize.
const confirmed = {
	'Dante Spada': 'https://es.wikipedia.org/wiki/Dante_Spada',
	'Ernesto Seijas': 'https://en.wikipedia.org/wiki/Ernesto_Garc%C3%ADa_Seijas',
	'Giuseppe Prisco': 'https://es.wikipedia.org/wiki/Giuseppe_Prisco',
	'Michele Rubini': 'https://es.wikipedia.org/wiki/Michele_Rubini',
	'Pia Ravenna': 'https://fi.wikipedia.org/wiki/Pia_Ravenna'
};

const sleep = (ms) => new Promise((resolve) => setTimeout(resolve, ms));

const articleUrl = (language, title) =>
	`https://${language}.wikipedia.org/wiki/${encodeURIComponent(title.replaceAll(' ', '_'))}`;

// Wikipedia rate-limits bursts; wait between requests and retry politely.
async function wikiQuery(language, params) {
	const url = `https://${language}.wikipedia.org/w/api.php?format=json&${new URLSearchParams(params)}`;
	for (let attempt = 1; attempt <= 5; attempt++) {
		await sleep(1500 * attempt);
		const response = await fetch(url, { headers: { 'User-Agent': userAgent } });
		const text = await response.text();
		try {
			return JSON.parse(text);
		} catch {
			// rate limited: try again after a longer pause
		}
	}
	throw new Error(`Wikipedia (${language}) did not answer`);
}

async function authorNames() {
	const names = new Set();
	for (const type of ['writer', 'drawer', 'translator']) {
		const response = await fetch(`${backend}/api/authors?type=${type}&pageSize=100`);
		if (!response.ok) throw new Error(`authors (${type}): ${response.status}`);
		const { authors = [], meta } = await response.json();
		if (meta && meta.totalPages > 1) throw new Error(`more than 100 ${type}s; add paging`);
		for (const author of authors) {
			const name = `${author.firstName} ${author.lastName}`.trim();
			if (name) names.add(name);
		}
	}
	return [...names].sort((a, b) => a.localeCompare(b, 'fi'));
}

// Article titles for the names that have one (not a disambiguation page).
async function articleTitles(language, names) {
	const titles = new Map();
	for (let i = 0; i < names.length; i += 50) {
		const batch = names.slice(i, i + 50);
		const { query = {} } = await wikiQuery(language, {
			action: 'query',
			redirects: '1',
			prop: 'pageprops',
			titles: batch.join('|')
		});
		const normalized = new Map((query.normalized ?? []).map((r) => [r.from, r.to]));
		const redirects = new Map((query.redirects ?? []).map((r) => [r.from, r.to]));
		const pages = Object.values(query.pages ?? {});
		for (const name of batch) {
			let title = normalized.get(name) ?? name;
			title = redirects.get(title) ?? title;
			const page = pages.find((p) => p.title === title);
			if (page && page.missing === undefined && !('disambiguation' in (page.pageprops ?? {}))) {
				titles.set(name, title);
			}
		}
	}
	return titles;
}

async function intros(language, titles) {
	const texts = new Map();
	for (let i = 0; i < titles.length; i += 20) {
		const { query = {} } = await wikiQuery(language, {
			action: 'query',
			prop: 'extracts',
			exintro: '1',
			explaintext: '1',
			exlimit: '20',
			titles: titles.slice(i, i + 20).join('|')
		});
		for (const page of Object.values(query.pages ?? {})) texts.set(page.title, page.extract ?? '');
	}
	return texts;
}

// When a namesake has the plain title (Claudio Villa the singer), the author's
// article has a qualifier: "Claudio Villa (fumettista)". These are tried after
// the plain name.
const qualifiers = {
	fi: ['sarjakuvataiteilija', 'sarjakuvapiirtäjä', 'piirtäjä', 'kirjailija', 'suomentaja'],
	it: ['fumettista', 'disegnatore', 'sceneggiatore', 'illustratore'],
	en: ['comics', 'cartoonist', 'artist', 'illustrator', 'writer']
};

const names = await authorNames();
const links = Object.fromEntries(
	Object.entries(confirmed).filter(([name]) => names.includes(name))
);
for (const language of languages) {
	const candidates = names.flatMap((name) => [
		[name, name],
		...qualifiers[language].map((qualifier) => [name, `${name} (${qualifier})`])
	]);
	const titles = await articleTitles(
		language,
		candidates.map(([, candidate]) => candidate)
	);
	const texts = await intros(language, [...new Set(titles.values())]);
	for (const [name, candidate] of candidates) {
		const title = titles.get(candidate);
		if (links[name] || !title || !aboutComics.test(texts.get(title) ?? '')) continue;
		links[name] = articleUrl(language, title);
	}
}

// Last, search for the authors still without a link by name and "Tex", and
// take a result whose title contains the name and whose intro is about comics.
for (const language of languages) {
	for (const name of names.filter((n) => !links[n])) {
		const { query = {} } = await wikiQuery(language, {
			action: 'query',
			list: 'search',
			srsearch: `"${name}" Tex`,
			srlimit: '3'
		});
		const found = (query.search ?? [])
			.map((result) => result.title)
			.filter((title) => title.toLowerCase().includes(name.toLowerCase()));
		if (found.length === 0) continue;
		const texts = await intros(language, found);
		const title = found.find((t) => aboutComics.test(texts.get(t) ?? ''));
		if (title) links[name] = articleUrl(language, title);
	}
}

// single quotes, as the UI's Prettier settings format the file
const quote = (value) => `'${value.replaceAll('\\', '\\\\').replaceAll("'", "\\'")}'`;
const entries = names
	.filter((name) => links[name])
	.map((name) => `\t${quote(name)}: ${quote(links[name])}`);
writeFileSync(
	output,
	`// Generated by scripts/update_author_wikipedia.mjs; do not edit by hand.
// Wikipedia articles of authors, by name: Finnish preferred, then Italian,
// then English, and only articles whose intro is about comics or translation;
// plus a few confirmed by hand in the script.
export const authorWikipedia: Record<string, string> = {
${entries.join(',\n')}
};

export function authorWikipediaUrl(name: string): string | undefined {
	return authorWikipedia[name.trim()];
}
`
);
console.log(`${entries.length} of ${names.length} authors linked; wrote ${output}`);
