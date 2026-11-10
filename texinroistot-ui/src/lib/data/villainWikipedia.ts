// Wikipedia articles of villains based on real people, by villain hash (the
// hash comes from the villain's id in the source spreadsheet, so it stays the
// same across imports). The Finnish article where there is one, else the
// English one. The link asks who the real person was.
export type VillainWikipediaLink = { url: string; label: string };

export const villainWikipedia: Record<string, VillainWikipediaLink> = {
	// Kenraali Stonewell
	f53bf06bed2ac8abb6e21e957bed5f106b60da381d1caaf34610b2299c114e22: {
		url: 'https://fi.wikipedia.org/wiki/Stonewall_Jackson',
		label: 'Kuka oli oikea kenraali Stonewall?'
	},
	// Kenraali Jackson
	'4079e3fc52b02e3bb118ce7c7fcd8616f44f301c87e57e4c1b9099e2a45c0d13': {
		url: 'https://fi.wikipedia.org/wiki/Stonewall_Jackson',
		label: 'Kuka oli oikea kenraali Jackson?'
	},
	// Belle Starr
	'8e419bec08835547d19943fc770a30204d4df454e9da833b6bd63f73e04a5992': {
		url: 'https://fi.wikipedia.org/wiki/Belle_Starr',
		label: 'Kuka oli oikea Belle Starr?'
	},
	// Soapy Smith
	'595b725d422a8706d6f2d033b6aaa55e5bdf87ab25ed2c8675c22261d025af9d': {
		url: 'https://en.wikipedia.org/wiki/Soapy_Smith',
		label: 'Kuka oli oikea Soapy Smith?'
	},
	// Jack McCall
	a7fb91a2cca6f648991db3523f26cf568d32731fd6bbb6ffb50958640d13a39c: {
		url: 'https://en.wikipedia.org/wiki/Jack_McCall',
		label: 'Kuka oli oikea Jack McCall?'
	},
	// Tom Pickett
	fea35376167c06127775d8d439a9eb52e152b5f08885fcfb139104a9491ac62d: {
		url: 'https://en.wikipedia.org/wiki/Tom_Pickett',
		label: 'Kuka oli oikea Tom Pickett?'
	},
	// Joshua Webb
	'408be6e239a50b15f3f3a3e31cf597b7aca712f163610067b269ba39ab6a8be9': {
		url: 'https://en.wikipedia.org/wiki/John_Joshua_Webb',
		label: 'Kuka oli oikea John Joshua Webb?'
	},
	// Dan Rudabaugh
	aea2ed958de82abb163a4144288db4d2fcab4f70cfd64c4d1b5a2ecae57bebd7: {
		url: 'https://fi.wikipedia.org/wiki/Dave_Rudabaugh',
		label: 'Kuka oli oikea Dave Rudabaugh?'
	},
	// Dave Mather
	ebf2a8b8380a2f7ba98508d1e78f00f94767e5987e24bacea05ccd585c612cce: {
		url: 'https://en.wikipedia.org/wiki/Mysterious_Dave_Mather',
		label: 'Kuka oli oikea Mysterious Dave Mather?'
	},
	// Jean Lafitte
	'0c818e94eace40e03a43d37472f3da6799879e172eef6ab83ec79d72da7a42a4': {
		url: 'https://fi.wikipedia.org/wiki/Jean_Lafitte',
		label: 'Kuka oli oikea Jean Lafitte?'
	},
	// Harry Tracy
	'7607d6c8467e9dce5be7acaa0dae4bbea1f3d9d4c74fd1563d941c5c24397f9c': {
		url: 'https://en.wikipedia.org/wiki/Harry_Tracy',
		label: 'Kuka oli oikea Harry Tracy?'
	},
	// Kid Curry
	ebb3c0c1f3cb0361222de54b39539224073dac33fbbd6bcbddf9d2594758b42f: {
		url: 'https://en.wikipedia.org/wiki/Harvey_Logan',
		label: 'Kuka oli oikea Harvey Logan?'
	},
	// Ben Kilpatrick
	'96e7d3b47bbb9a54277adb216c7fc775cc209c538e0ae71868e218ac37bfd732': {
		url: 'https://en.wikipedia.org/wiki/Ben_Kilpatrick',
		label: 'Kuka oli oikea Ben Kilpatrick?'
	},
	// Jack Ketchum
	aea2f216759858f7650e1db8e100b43506dbf91095a9935d0522b292cde7b347: {
		url: 'https://en.wikipedia.org/wiki/Tom_Ketchum',
		label: 'Kuka oli oikea Tom Ketchum?'
	},
	// Sundance Kid
	a7d1726883573b02b877772c3929b382aba29e05db0f44126b4d706b46440cf7: {
		url: 'https://fi.wikipedia.org/wiki/Sundance_Kid',
		label: 'Kuka oli oikea Sundance Kid?'
	},
	// Flatnose George
	a8f1384a1f04a220ab8df71cdd04fb450a3bcbec486259d7c42dc7f784849ecd: {
		url: 'https://en.wikipedia.org/wiki/George_Curry_(outlaw)',
		label: 'Kuka oli oikea George Curry?'
	},
	// Elzy Lay
	'6118a01a732e144e66dcc700c4cf05e0dabe53ac8d80f4b8ccedbdc099afd847': {
		url: 'https://en.wikipedia.org/wiki/Elzy_Lay',
		label: 'Kuka oli oikea Elzy Lay?'
	},
	// Charlie Pierce
	'24061d4992ee6a621cdaf8137d09707c203068816166626fccc7f3bc6f464e28': {
		url: 'https://en.wikipedia.org/wiki/Charley_Pierce',
		label: 'Kuka oli oikea Charley Pierce?'
	},
	// George Newcomb
	bc60fb08d5089105d74f06532378b63967b9ea08345443f92f902cc2fd5fab0a: {
		url: 'https://en.wikipedia.org/wiki/George_Newcomb',
		label: 'Kuka oli oikea George Newcomb?'
	},
	// Shanghai Kelly
	'5651982389a6f2b723c1af82ad5d46276b3c8ce37c0de6c2a9f3049d8070b7ad': {
		url: 'https://en.wikipedia.org/wiki/James_Kelly_(crimper)',
		label: 'Kuka oli oikea James Kelly?'
	},
	// Bill Doolin
	'6355cb4d55bba1eccbe00768cdbf7edd9c8337b894881c059b36d3ff30c36a19': {
		url: 'https://en.wikipedia.org/wiki/Bill_Doolin',
		label: 'Kuka oli oikea Bill Doolin?'
	},
	// Kenraali Custer
	d79419f4115881b5957470ed6da1d219d041f8862b77162db87ab4cb79d8352a: {
		url: 'https://fi.wikipedia.org/wiki/George_Armstrong_Custer',
		label: 'Kuka oli oikea George Armstrong Custer?'
	},
	// Franklin Buchanan
	'9180bb59fa3d332e9e957f3cf3f77fcf1aae79109e0fcb17603d3c013228cbe2': {
		url: 'https://en.wikipedia.org/wiki/Franklin_Buchanan',
		label: 'Kuka oli oikea Franklin Buchanan?'
	},
	// Willy Claiborne
	a1231c58e8ebdb19ac72280d08f42ff7a6e4975c26de9a03b4aa5aead14b59c7: {
		url: 'https://en.wikipedia.org/wiki/Billy_Claiborne',
		label: 'Kuka oli oikea Billy Claiborne?'
	},
	// Tom McLaury
	b7342d81a5bba5d7eb45ca0b6f9ba396d97046d17b75a7c96bae214f707d7c33: {
		url: 'https://en.wikipedia.org/wiki/Tom_McLaury',
		label: 'Kuka oli oikea Tom McLaury?'
	},
	// Frank Stilwell
	a3dcdc87699ad28514a7ae90543b7ad9bd6b6935ac4486f9232e74b3c6cd0b93: {
		url: 'https://en.wikipedia.org/wiki/Frank_Stilwell',
		label: 'Kuka oli oikea Frank Stilwell?'
	},
	// Frank McLaury
	f5265a5edc0ff96b09987b7f82203075bdea9ca1972c6b7a10ac82e2a5702881: {
		url: 'https://en.wikipedia.org/wiki/Frank_McLaury',
		label: 'Kuka oli oikea Frank McLaury?'
	},
	// Curly Bill
	'16bbd2430368ec3a78d89cc14ca53eed7ad4629d8f82bad13c0441bf907f47f2': {
		url: 'https://en.wikipedia.org/wiki/William_Brocius',
		label: 'Kuka oli oikea William Brocius?'
	},
	// Johnny Behan
	'1eed39ad2b01bde85e36d92c1f33195d9a710d524da9c73304340b9282bd48a4': {
		url: 'https://en.wikipedia.org/wiki/Johnny_Behan',
		label: 'Kuka oli oikea Johnny Behan?'
	},
	// Roy Bean
	cef722c7deee854aa533300373fe89ffcd5fddf96557209b7edea244b2dc84fb: {
		url: 'https://fi.wikipedia.org/wiki/Roy_Bean',
		label: 'Kuka oli oikea Roy Bean?'
	},
	// Samuel Quantrill
	f9e1398e37468fbe3a693174565f43f9a1f12a714454f4cf6f54a087bf754f27: {
		url: 'https://fi.wikipedia.org/wiki/William_Quantrill',
		label: 'Kuka oli oikea William Quantrill?'
	},
	// Bloody Bill Anderson
	f3192dba24fad689fb9d88af46b9aea0ffcc73b3e62f0c00cd2f752a44d6b0db: {
		url: 'https://fi.wikipedia.org/wiki/William_T._Anderson',
		label: 'Kuka oli oikea William T. Anderson?'
	},
	// Victorio
	'34c41882d8305cd8c77669b42fb124dc4ad7f3737cf7942500a7298bf9c9d8bc': {
		url: 'https://fi.wikipedia.org/wiki/Victorio',
		label: 'Kuka oli oikea Victorio?'
	},
	// Scarface Charlie
	'0fe3c9b39111707f264e2ced7810da7f190f4c15abc3010bd44569c2a623e5c7': {
		url: 'https://fi.wikipedia.org/wiki/Arpinaama_Charley',
		label: 'Kuka oli oikea Arpinaama Charley?'
	},
	// Nana
	'2fbab108b540d0a79d1095e57b3bdccee7caa4d036d452527b4c527a51b022bb': {
		url: 'https://fi.wikipedia.org/wiki/Nana_(apassi)',
		label: 'Kuka oli oikea Nana?'
	},
	// Moni-Hevonen
	'5de4a3c68112e223c1d8f8b11114124c672a537b50cdfdbd7064da0f5fe57fa5': {
		url: 'https://en.wikipedia.org/wiki/Plenty_Horses',
		label: 'Kuka oli oikea Plenty Horses?'
	},
	// Mangas Colorados
	d0f990badd5399347b679593d74c921c688caf883570a308165c1d07ad93ea21: {
		url: 'https://fi.wikipedia.org/wiki/Mangas_Coloradas',
		label: 'Kuka oli oikea Mangas Coloradas?'
	},
	// Lone Wolf
	b11f3c828737f58b44501ad4d4f4d376645672bf652f71ba20c7e70b94451275: {
		url: 'https://fi.wikipedia.org/wiki/Lone_Wolf_(intiaanijohtaja)',
		label: 'Kuka oli oikea Lone Wolf?'
	},
	// Kicking Bird
	f50ecb7ef3d872a60881d185c2400e904f9321e86ac6dcfa785e7d652a8b3fd2: {
		url: 'https://fi.wikipedia.org/wiki/Kicking_Bird',
		label: 'Kuka oli oikea Kicking Bird?'
	},
	// Iron Shirt
	'9b2f1deb9773517af6dfe9824cfa35b901fffd0391aaf8209138980a8b4cbd04': {
		url: 'https://en.wikipedia.org/wiki/Iron_Jacket',
		label: 'Kuka oli oikea Iron Jacket?'
	},
	// Hooker Jim
	'562be08e69e59d0c07ce4e7083e1e7ad3d1f3e88e3338daca3cc94d078b076d3': {
		url: 'https://en.wikipedia.org/wiki/Hooker_Jim',
		label: 'Kuka oli oikea Hooker Jim?'
	},
	// Cuchillo Negro
	'9d58d06ffd7a4bae055b1e7c58f7b2edfe410aaf946797e369b64ae0bb518415': {
		url: 'https://fi.wikipedia.org/wiki/Baishan',
		label: 'Kuka oli oikea Baishan?'
	},
	// Geronimo
	f54d1315675f146783cd06493f40ada87216e48b8e1a7b0146c7a857907d81ec: {
		url: 'https://fi.wikipedia.org/wiki/Geronimo',
		label: 'Kuka oli oikea Geronimo?'
	},
	// Dull Knife
	e609665833190543497036ce81b4d7ae729d1f9ff5428d216a60c7777c4c764a: {
		url: 'https://en.wikipedia.org/wiki/Morning_Star_(chief)',
		label: 'Kuka oli oikea Morning Star?'
	},
	// Curly Headed Doctor
	'0339ef4c5b4c80d5ccbb41ac3d1576a588289f0e56f49cd386acbf6fe6bb3524': {
		url: 'https://en.wikipedia.org/wiki/Curley_Headed_Doctor',
		label: 'Kuka oli oikea Curley Headed Doctor?'
	},
	// Ike Clanton
	'07dae3ca3a73343d6003b54dc276d2b8e6d87ef3811a8fd71c9a04f488b1b680': {
		url: 'https://fi.wikipedia.org/wiki/Ike_Clanton',
		label: 'Kuka oli oikea Ike Clanton?'
	},
	// Johnny Ringo
	'1d6eef147705fa18f66834623ed334bc1b814b08c541f8ab0894eef26e9ec2cc': {
		url: 'https://fi.wikipedia.org/wiki/Johnny_Ringo',
		label: 'Kuka oli oikea Johnny Ringo?'
	},
	// Billy Clanton
	'8e8b097899267df7c57106253bc406405abbc863d77854fb7cc8cedd93978fcd': {
		url: 'https://en.wikipedia.org/wiki/Billy_Clanton',
		label: 'Kuka oli oikea Billy Clanton?'
	},
	// Delgadito
	b44ab24b44dba8bbcd022748add63bd154daea12576adf13a9684cff2bb30f5a: {
		url: 'https://en.wikipedia.org/wiki/Delgadito_(Apache_chief)',
		label: 'Kuka oli oikea Delgadito?'
	},
	// Buffalo Chief
	a45d553b31c0f3105390dc4d09517a7e7a8584b68e52ce04047826f47f41159e: {
		url: 'https://en.wikipedia.org/wiki/Kechewaishke',
		label: 'Kuka oli oikea Kechewaishke?'
	},
	// Crazy Horse
	'605b5ff27085bdf8be8049b4c04194cd0a1f165edcc608998b4651f50aad2a59': {
		url: 'https://fi.wikipedia.org/wiki/Hullu_Hevonen',
		label: 'Kuka oli oikea Hullu Hevonen?'
	},
	// Jesse James
	'13a6bb02ef093307ae5ec8426ab636a8c2d1c579c233a09fd9332d3343f58e79': {
		url: 'https://fi.wikipedia.org/wiki/Jesse_James',
		label: 'Kuka oli oikea Jesse James?'
	},
	// Emmett Dalton
	d0bfcc2d254b87756cbdeeffda0d5b3cd6979d8b2a58dd763347b58c9d3311fd: {
		url: 'https://fi.wikipedia.org/wiki/Emmett_Dalton',
		label: 'Kuka oli oikea Emmett Dalton?'
	},
	// Bob Dalton
	'00a53c60c2fa82bf715a1aeb6f64026ae2ea78e32d350e50b5fe84831a4829d1': {
		url: 'https://fi.wikipedia.org/wiki/Bob_Dalton',
		label: 'Kuka oli oikea Bob Dalton?'
	},
	// George Leroy Parker
	'025551608eaa9de93197ea5064ceebd068890ae1435ada00f325f07000056547': {
		url: 'https://fi.wikipedia.org/wiki/Butch_Cassidy',
		label: 'Kuka oli oikea Butch Cassidy?'
	},
	// Hoodoo Brown
	'5aaec00a73068f6b7bce424a220c1ec5b72967f2f9258667794cb6af6ef6757c': {
		url: 'https://en.wikipedia.org/wiki/Hoodoo_Brown',
		label: 'Kuka oli oikea Hoodoo Brown?'
	}
};

export function villainWikipediaLink(
	hash: string | null | undefined
): VillainWikipediaLink | undefined {
	return hash ? villainWikipedia[hash] : undefined;
}
