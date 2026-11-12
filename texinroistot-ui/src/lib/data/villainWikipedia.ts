import type { Villain } from '$lib/types';

// Wikipedia articles of villains based on real people: the Finnish article
// where there is one, else the English one, linked as "Kuka oli oikea ...?".
//
// A villain is found by name (first and last name, a nickname or an other
// name, in any of their stories), since villain hashes are salted per
// environment. Where several villains share the name, a rank or a phrase in
// one of the villain's roles picks the right one.
// A villain modelled on several people (more) shows the label followed by a
// link to each, by name; otherwise the label itself is the link.
export type WikipediaPerson = { name: string; url: string };
export type VillainWikipediaLink = {
	url: string;
	label: string;
	name?: string;
	more?: WikipediaPerson[];
};

type Entry = VillainWikipediaLink & { names: string[]; rank?: string; role?: string };

const entries: Entry[] = [
	{
		names: ['Stonewell'],
		rank: 'Kenraali',
		url: 'https://fi.wikipedia.org/wiki/Stonewall_Jackson',
		label: 'Kuka oli oikea kenraali Stonewall?'
	},
	{
		names: ['Jackson'],
		rank: 'Kenraali',
		url: 'https://fi.wikipedia.org/wiki/Stonewall_Jackson',
		label: 'Kuka oli oikea kenraali Jackson?'
	},
	{
		names: ['Belle Starr'],
		url: 'https://fi.wikipedia.org/wiki/Belle_Starr',
		label: 'Kuka oli oikea Belle Starr?'
	},
	{
		names: ['Soapy Smith'],
		url: 'https://en.wikipedia.org/wiki/Soapy_Smith',
		label: 'Kuka oli oikea Soapy Smith?'
	},
	{
		names: ['Jack McCall'],
		url: 'https://en.wikipedia.org/wiki/Jack_McCall',
		label: 'Kuka oli oikea Jack McCall?'
	},
	{
		names: ['Cyrus Skinner'],
		url: 'https://en.wikipedia.org/wiki/Cyrus_Skinner',
		label: 'Kuka oli oikea Cyrus Skinner?'
	},
	{
		names: ["Pagan O'Leary"],
		url: 'https://www.dib.ie/biography/oleary-patrick-pagan-a6865',
		label: "Kuka oli oikea Pagan O'Leary?"
	},
	{
		names: ['David McCanles'],
		url: 'https://en.wikipedia.org/wiki/McCanles_Gang',
		label: 'Kuka oli oikea David McCanles?'
	},
	{
		names: ['Charles Jennison'],
		url: 'https://en.wikipedia.org/wiki/Charles_R._Jennison',
		label: 'Kuka oli oikea eversti Jennison?'
	},
	{
		names: ['John Fremont'],
		url: 'https://fi.wikipedia.org/wiki/John_C._Fr%C3%A9mont',
		label: 'Kuka oli oikea John Fremont?'
	},
	{
		names: ['Diamond Jim', 'Timantti-Jim'],
		url: 'https://en.wikipedia.org/wiki/Big_Jim_Colosimo',
		label: 'Kuka oli oikea Diamond Jim?'
	},
	{
		names: ['Johnny Devine'],
		url: 'https://www.foundsf.org/Irish_1851',
		label: 'Kuka oli oikea Johnny Devine?'
	},
	{
		names: ["Tom O'Day"],
		url: 'https://tomodayproductions.com/tom-oday-1',
		label: "Kuka oli oikea Tom O'Day?"
	},
	{
		names: ['Bedford'],
		rank: 'Kenraali',
		url: 'https://fi.wikipedia.org/wiki/Nathan_Bedford_Forrest',
		label: 'Kuka oli oikea kenraali Bedford?'
	},
	{
		names: ['Potkiva Karhu'],
		url: 'https://en.wikipedia.org/wiki/Kicking_Bear',
		label: 'Kuka oli oikea Potkiva Karhu?'
	},
	{
		names: ['Musta Hirvi'],
		role: 'Wovoka',
		url: 'https://fi.wikipedia.org/wiki/Musta_Hirvi',
		label: 'Kuka oli oikea Musta Hirvi?'
	},
	{
		names: ['Wovoka', 'Wowoka'],
		role: 'paiute|piute',
		url: 'https://fi.wikipedia.org/wiki/Wovoka',
		label: 'Kuka oli oikea Wovoka?'
	},
	{
		names: ['Satanta'],
		role: 'kiowa',
		url: 'https://fi.wikipedia.org/wiki/Satanta',
		label: 'Kuka oli oikea Satanta?'
	},
	{ names: ['Ouray'], url: 'https://fi.wikipedia.org/wiki/Ouray', label: 'Kuka oli oikea Ouray?' },
	{
		names: ['Johnny Lame'],
		url: 'https://en.wikipedia.org/wiki/Cornelius_Donahue',
		label: 'Kuka oli oikea Johnny Lame?',
		name: 'Cornelius Donahue',
		more: [{ name: 'Clubfoot George', url: 'https://en.wikipedia.org/wiki/Clubfoot_George' }]
	},
	{
		names: ['Ihmissyöjä-Boone'],
		url: 'https://en.wikipedia.org/wiki/Boone_Helm',
		label: 'Kuka oli oikea Ihmissyöjä-Boone?'
	},
	{
		names: ['Tom Pickett'],
		url: 'https://en.wikipedia.org/wiki/Tom_Pickett',
		label: 'Kuka oli oikea Tom Pickett?'
	},
	{
		names: ['Joshua Webb'],
		url: 'https://en.wikipedia.org/wiki/John_Joshua_Webb',
		label: 'Kuka oli oikea John Joshua Webb?'
	},
	{
		names: ['Dan Rudabaugh'],
		url: 'https://fi.wikipedia.org/wiki/Dave_Rudabaugh',
		label: 'Kuka oli oikea Dave Rudabaugh?'
	},
	{
		names: ['Dave Mather'],
		url: 'https://en.wikipedia.org/wiki/Mysterious_Dave_Mather',
		label: 'Kuka oli oikea Mysterious Dave Mather?'
	},
	{
		names: ['Jean Lafitte'],
		url: 'https://fi.wikipedia.org/wiki/Jean_Lafitte',
		label: 'Kuka oli oikea Jean Lafitte?'
	},
	{
		names: ['Harry Tracy'],
		url: 'https://en.wikipedia.org/wiki/Harry_Tracy',
		label: 'Kuka oli oikea Harry Tracy?'
	},
	{
		names: ['Kid Curry'],
		url: 'https://en.wikipedia.org/wiki/Harvey_Logan',
		label: 'Kuka oli oikea Kid Curry?'
	},
	{
		names: ['Ben Kilpatrick'],
		url: 'https://en.wikipedia.org/wiki/Ben_Kilpatrick',
		label: 'Kuka oli oikea Ben Kilpatrick?'
	},
	{
		names: ['Jack Ketchum'],
		url: 'https://en.wikipedia.org/wiki/Tom_Ketchum',
		label: 'Kuka oli oikea Black Jack?'
	},
	{
		names: ['Sundance Kid'],
		url: 'https://fi.wikipedia.org/wiki/Sundance_Kid',
		label: 'Kuka oli oikea Sundance Kid?'
	},
	{
		names: ['Flatnose George'],
		url: 'https://en.wikipedia.org/wiki/George_Curry_(outlaw)',
		label: 'Kuka oli oikea Flatnose George?'
	},
	{
		names: ['Elzy Lay'],
		url: 'https://en.wikipedia.org/wiki/Elzy_Lay',
		label: 'Kuka oli oikea Elzy Lay?'
	},
	{
		names: ['Charlie Pierce'],
		url: 'https://en.wikipedia.org/wiki/Charley_Pierce',
		label: 'Kuka oli oikea Charley Pierce?'
	},
	{
		names: ['George Newcomb'],
		url: 'https://en.wikipedia.org/wiki/George_Newcomb',
		label: 'Kuka oli oikea George Newcomb?'
	},
	{
		names: ['Shanghai-Kelly', 'Shanghain Kelly'],
		url: 'https://en.wikipedia.org/wiki/James_Kelly_(crimper)',
		label: 'Kuka oli oikea "Shanghai" Kelly?'
	},
	{
		names: ['Bill Doolin'],
		url: 'https://en.wikipedia.org/wiki/Bill_Doolin',
		label: 'Kuka oli oikea Bill Doolin?'
	},
	{
		names: ['Cherokee Bill'],
		url: 'https://en.wikipedia.org/wiki/Crawford_Goldsby',
		label: 'Kuka oli oikea Cherokee Bill?'
	},
	{
		names: ['Rube Burrows'],
		url: 'https://en.wikipedia.org/wiki/Rube_Burrow',
		label: 'Kuka oli oikea Rube Burrows?'
	},
	{
		names: ['Jim Burrows'],
		url: 'https://en.wikipedia.org/wiki/Rube_Burrow',
		label: 'Kuka oli oikea Jim Burrows?'
	},
	{
		names: ['Lucky Davis'],
		url: 'https://en.wikipedia.org/wiki/Rufus_Buck_Gang',
		label: 'Kuka oli oikea Lucky Davis?'
	},
	{
		names: ['Rufus Buck'],
		url: 'https://en.wikipedia.org/wiki/Rufus_Buck_Gang',
		label: 'Kuka oli oikea Rufus Buck?'
	},
	{
		names: ['Custer'],
		rank: 'Kenraali',
		url: 'https://fi.wikipedia.org/wiki/George_Armstrong_Custer',
		label: 'Kuka oli oikea kenraali Custer?'
	},
	{
		names: ['Franklin Buchanan'],
		url: 'https://en.wikipedia.org/wiki/Franklin_Buchanan',
		label: 'Kuka oli oikea Franklin Buchanan?'
	},
	{
		names: ['Willy Claiborne'],
		url: 'https://en.wikipedia.org/wiki/Billy_Claiborne',
		label: 'Kuka oli oikea William Claiborne?'
	},
	{
		names: ['Tom McLaury'],
		url: 'https://en.wikipedia.org/wiki/Tom_McLaury',
		label: 'Kuka oli oikea Tom McLaury?'
	},
	{
		names: ['Frank Stilwell'],
		url: 'https://en.wikipedia.org/wiki/Frank_Stilwell',
		label: 'Kuka oli oikea Frank Stilwell?'
	},
	{
		names: ['Frank McLaury'],
		url: 'https://en.wikipedia.org/wiki/Frank_McLaury',
		label: 'Kuka oli oikea Frank McLaury?'
	},
	{
		names: ['Curly Bill'],
		url: 'https://en.wikipedia.org/wiki/William_Brocius',
		label: 'Kuka oli oikea Curly Bill?'
	},
	{
		names: ['Johnny Behan'],
		url: 'https://en.wikipedia.org/wiki/Johnny_Behan',
		label: 'Kuka oli oikea Johnny Behan?'
	},
	{
		names: ['Juan Cortina'],
		url: 'https://en.wikipedia.org/wiki/Juan_Cortina',
		label: 'Kuka oli oikea Juan Cortina?'
	},
	{
		names: ['Roy Bean'],
		url: 'https://fi.wikipedia.org/wiki/Roy_Bean',
		label: 'Kuka oli oikea Roy Bean?'
	},
	{
		names: ['Samuel Quantrill (Quantrell)', 'Samuel Quantrell'],
		url: 'https://fi.wikipedia.org/wiki/William_Quantrill',
		label: 'Kuka oli oikea Quantrill?'
	},
	{
		names: ['Bloody Bill'],
		url: 'https://fi.wikipedia.org/wiki/William_T._Anderson',
		label: 'Kuka oli oikea Bloody Bill Anderson?'
	},
	{
		names: ['Victorio'],
		role: 'apachepäällikkö',
		url: 'https://fi.wikipedia.org/wiki/Victorio',
		label: 'Kuka oli oikea Victorio?'
	},
	{
		names: ['Scarface Charlie'],
		url: 'https://fi.wikipedia.org/wiki/Arpinaama_Charley',
		label: 'Kuka oli oikea Scarface Charley?'
	},
	{
		names: ['Nana'],
		rank: 'Poppamies',
		url: 'https://fi.wikipedia.org/wiki/Nana_(apassi)',
		label: 'Kuka oli oikea Nana?'
	},
	{
		names: ['Plenty Horse', 'Moni-Hevonen'],
		url: 'https://en.wikipedia.org/wiki/Plenty_Horses',
		label: 'Kuka oli oikea Plenty Horses?'
	},
	{
		names: ['Mangas Colorados'],
		url: 'https://fi.wikipedia.org/wiki/Mangas_Coloradas',
		label: 'Kuka oli oikea Mangas Coloradas?'
	},
	{
		names: ['Lone Wolf'],
		url: 'https://fi.wikipedia.org/wiki/Lone_Wolf_(intiaanijohtaja)',
		label: 'Kuka oli oikea Lone Wolf?'
	},
	{
		names: ['Kicking Bird'],
		url: 'https://fi.wikipedia.org/wiki/Kicking_Bird',
		label: 'Kuka oli oikea Kicking Bird?'
	},
	{
		names: ['Quanah Parker'],
		url: 'https://fi.wikipedia.org/wiki/Quanah_Parker',
		label: 'Kuka oli oikea Quanah Parker?'
	},
	{
		names: ['Iron Shirt'],
		url: 'https://en.wikipedia.org/wiki/Iron_Jacket',
		label: 'Kuka oli oikea Iron Jacket?'
	},
	{
		names: ['Captain Jack'],
		url: 'https://fi.wikipedia.org/wiki/Kapteeni_Jack',
		label: 'Kuka oli oikea Captain Jack?'
	},
	{
		names: ['Hooker Jim'],
		url: 'https://en.wikipedia.org/wiki/Hooker_Jim',
		label: 'Kuka oli oikea Hooker Jim?'
	},
	{
		names: ['Cuchillo Negro'],
		url: 'https://en.wikipedia.org/wiki/Baishan_(Apache)',
		label: 'Kuka oli oikea Cuchillo Negro?'
	},
	{
		names: ['Geronimo'],
		url: 'https://fi.wikipedia.org/wiki/Geronimo',
		label: 'Kuka oli oikea Geronimo?'
	},
	{
		names: ['Dull Knife'],
		url: 'https://en.wikipedia.org/wiki/Morning_Star_(chief)',
		label: 'Kuka oli oikea Dull Knife?'
	},
	{
		names: ['Curly Headed Doctor'],
		url: 'https://en.wikipedia.org/wiki/Curley_Headed_Doctor',
		label: 'Kuka oli oikea Curley Headed Doctor?'
	},
	{
		names: ['Ike Clanton'],
		url: 'https://fi.wikipedia.org/wiki/Ike_Clanton',
		label: 'Kuka oli oikea Ike Clanton?'
	},
	{
		names: ['Johnny Ringo'],
		url: 'https://fi.wikipedia.org/wiki/Johnny_Ringo',
		label: 'Kuka oli oikea Johnny Ringo?'
	},
	{
		names: ['Billy Clanton'],
		url: 'https://en.wikipedia.org/wiki/Billy_Clanton',
		label: 'Kuka oli oikea Billy Clanton?'
	},
	{
		names: ['Delgadito'],
		url: 'https://en.wikipedia.org/wiki/Delgadito_(Apache_chief)',
		label: 'Kuka oli oikea Delgadito?'
	},
	{
		names: ['Buffalo Chief'],
		url: 'https://en.wikipedia.org/wiki/Kechewaishke',
		label: 'Kuka oli oikea Chief Buffalo?'
	},
	{
		names: ['Crazy Horse'],
		url: 'https://fi.wikipedia.org/wiki/Hullu_Hevonen',
		label: 'Kuka oli oikea Hullu Hevonen?'
	},
	{
		names: ['Kid Billy'],
		url: 'https://fi.wikipedia.org/wiki/Billy_the_Kid',
		label: 'Kuka oli oikea Billy the Kid?'
	},
	{
		names: ['Cole Younger'],
		url: 'https://fi.wikipedia.org/wiki/Cole_Younger',
		label: 'Kuka oli oikea Cole Younger?'
	},
	{
		names: ['Jesse James'],
		url: 'https://fi.wikipedia.org/wiki/Jesse_James',
		label: 'Kuka oli oikea Jesse James?'
	},
	{
		names: ['Emmett Dalton'],
		url: 'https://fi.wikipedia.org/wiki/Emmett_Dalton',
		label: 'Kuka oli oikea Emmett Dalton?'
	},
	{
		names: ['Bob Dalton'],
		url: 'https://fi.wikipedia.org/wiki/Bob_Dalton',
		label: 'Kuka oli oikea Bob Dalton?'
	},
	{
		names: ['Grat Dalton'],
		url: 'https://fi.wikipedia.org/wiki/Grat_Dalton',
		label: 'Kuka oli oikea Grat Dalton?'
	},
	{
		names: ['George Leroy Parker'],
		url: 'https://fi.wikipedia.org/wiki/Butch_Cassidy',
		label: 'Kuka oli oikea Butch Cassidy?'
	},
	{
		names: ['Ah-Toy'],
		url: 'https://en.wikipedia.org/wiki/Ah_Toy',
		label: 'Kuka oli oikea Ah-Toy?'
	},
	{
		names: ['Hoodoo Brown'],
		url: 'https://en.wikipedia.org/wiki/Hoodoo_Brown',
		label: 'Kuka oli oikea Hoodoo Brown?'
	}
];

function normalize(text: string): string {
	return text.replace(/\s+/g, ' ').trim().toLowerCase();
}

const entriesByName = new Map<string, Entry[]>();
for (const entry of entries) {
	for (const name of entry.names) {
		const key = normalize(name);
		entriesByName.set(key, [...(entriesByName.get(key) ?? []), entry]);
	}
}

// Every name the villain goes by: first and last name (all first names, and
// each on its own), the last name alone, nicknames and other names.
function villainNames(villain: Villain): Set<string> {
	const names = new Set<string>();
	// the villain's own name too, when the story gives another (story villains)
	const sources = [villain, ...(villain.identity ? [villain.identity] : []), ...(villain.as ?? [])];
	for (const source of sources) {
		const firstNames = (source.firstNames ?? []).map((name) => name.trim()).filter(Boolean);
		const lastName = (source.lastName ?? '').trim();
		names.add(normalize(`${firstNames.join(' ')} ${lastName}`));
		for (const firstName of firstNames) names.add(normalize(`${firstName} ${lastName}`));
		if (lastName) names.add(normalize(lastName));
	}
	for (const appearance of villain.as ?? []) {
		for (const name of [...(appearance.nicknames ?? []), ...(appearance.otherNames ?? [])]) {
			names.add(normalize(name));
		}
	}
	names.delete('');
	return names;
}

function entryMatches(entry: Entry, villain: Villain): boolean {
	if (entry.rank) {
		const rank = normalize(entry.rank);
		// all of the villain's ranks, also those other stories give
		const ranks = [...(villain.ranks ?? []), ...(villain.identity?.ranks ?? [])];
		if (!ranks.some((villainRank) => normalize(villainRank) === rank)) return false;
	}
	if (entry.role) {
		// alternatives separated by | (a role may be worded differently per story)
		const wanted = entry.role.split('|').map(normalize);
		const roles = (villain.as ?? []).flatMap((appearance) => appearance.roles ?? []);
		if (!roles.some((villainRole) => wanted.some((role) => normalize(villainRole).includes(role))))
			return false;
	}
	return true;
}

export function villainWikipediaLink(
	villain: Villain | null | undefined
): VillainWikipediaLink | undefined {
	if (!villain) return undefined;
	for (const name of villainNames(villain)) {
		const entry = entriesByName.get(name)?.find((candidate) => entryMatches(candidate, villain));
		if (entry) return { url: entry.url, label: entry.label, name: entry.name, more: entry.more };
	}
	return undefined;
}
