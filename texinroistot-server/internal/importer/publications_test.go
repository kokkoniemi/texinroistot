package importer

import (
	"testing"

	"github.com/kokkoniemi/texinroistot/internal/db"
)

func newPublicationTestImporter() *importer {
	return &importer{
		columnIndexes: map[string]int{
			"story_title":       0,
			"pub_year":          1,
			"pub_from":          2,
			"pub_to":            3,
			"repub_year":        4,
			"repub_from":        5,
			"repub_to":          6,
			"pub_special":       7,
			"pub_kronikka":      8,
			"pub_kirjasto":      9,
			"italy_year":        10,
			"italy_pub_from":    11,
			"italy_pub_to":      12,
			"italy_pub_special": 13,
			"italy_story_title": 14,
		},
	}
}

func newPubRow(i *importer, rowIndex int, values map[string]string) row {
	maxIndex := 0
	for _, index := range i.columnIndexes {
		if index > maxIndex {
			maxIndex = index
		}
	}
	cells := make([]string, maxIndex+1)
	for key, index := range i.columnIndexes {
		cells[index] = values[key]
	}
	return row{importer: i, index: rowIndex, cells: cells}
}

func findPublication(i *importer, pubType string, year int, issue string) *importerPublication {
	for _, p := range i.publications {
		if p.item.Type == pubType && p.item.Year == year && p.item.Issue == issue {
			return p
		}
	}
	return nil
}

func findPublicationByIssue(i *importer, pubType string, issue string) *importerPublication {
	for _, p := range i.publications {
		if p.item.Type == pubType && p.item.Issue == issue {
			return p
		}
	}
	return nil
}

func storyPublicationsFor(i *importer, pubID id) []*importerStoryPublication {
	var out []*importerStoryPublication
	for _, sp := range i.storyPublications {
		if sp.publication == pubID {
			out = append(out, sp)
		}
	}
	return out
}

// 1. Finnish wrap-around year storage
// Vuosi=1974, Alkaen=12, Päättyen=1/75 → publications include (perus, 1974, 12)
// and (perus, 1975, 1), not (perus, 1974, 1).
func TestLoadBasePublication_FinnishWrapAroundUsesWrappedYear(t *testing.T) {
	i := newPublicationTestImporter()
	story := i.addStory(&db.Story{Hash: "story-coloradon"})

	r := newPubRow(i, 0, map[string]string{
		"story_title": "Coloradon rajoilla",
		"pub_year":    "1974",
		"pub_from":    "12",
		"pub_to":      "1/75",
	})

	if err := i.loadBasePublication(story.ID, r); err != nil {
		t.Fatalf("loadBasePublication failed: %v", err)
	}

	if findPublication(i, PUB_PERUS, 1974, "12") == nil {
		t.Errorf("expected publication (perus, 1974, 12)")
	}
	if findPublication(i, PUB_PERUS, 1975, "1") == nil {
		t.Errorf("expected publication (perus, 1975, 1)")
	}
	if findPublication(i, PUB_PERUS, 1974, "1") != nil {
		t.Errorf("did not expect publication (perus, 1974, 1) — outer year used incorrectly")
	}
}

// 2. Two Finnish stories share an issue
// Story A wraps 12/74 → 1/75; Story B starts at 1/75. Both stories appear
// linked to the same publication ID for 1/75 with their own titles.
func TestLoadBasePublication_TwoStoriesShareWrappedIssue(t *testing.T) {
	i := newPublicationTestImporter()
	storyA := i.addStory(&db.Story{Hash: "story-coloradon"})
	storyB := i.addStory(&db.Story{Hash: "story-kuolema"})

	rowA := newPubRow(i, 0, map[string]string{
		"story_title": "Coloradon rajoilla",
		"pub_year":    "1974",
		"pub_from":    "11",
		"pub_to":      "1/75",
	})
	rowB := newPubRow(i, 1, map[string]string{
		"story_title": "Kuolema tulee taivaalta",
		"pub_year":    "1975",
		"pub_from":    "1",
		"pub_to":      "2",
	})

	if err := i.loadBasePublication(storyA.ID, rowA); err != nil {
		t.Fatalf("loadBasePublication A failed: %v", err)
	}
	if err := i.loadBasePublication(storyB.ID, rowB); err != nil {
		t.Fatalf("loadBasePublication B failed: %v", err)
	}

	pub1975_1 := findPublication(i, PUB_PERUS, 1975, "1")
	if pub1975_1 == nil {
		t.Fatalf("expected publication (perus, 1975, 1)")
	}

	links := storyPublicationsFor(i, pub1975_1.ID)
	if len(links) != 2 {
		t.Fatalf("expected 2 story-publication links for 1975/1, got %d", len(links))
	}

	gotTitles := map[id]string{}
	for _, l := range links {
		gotTitles[l.story] = l.title
	}
	if gotTitles[storyA.ID] != "Coloradon rajoilla" {
		t.Errorf("story A link missing or wrong title: %q", gotTitles[storyA.ID])
	}
	if gotTitles[storyB.ID] != "Kuolema tulee taivaalta" {
		t.Errorf("story B link missing or wrong title: %q", gotTitles[storyB.ID])
	}
}

// 3. Idempotence within one story
// Multiple villain rows for the same story still result in exactly one
// stories_in_publications row per publication.
func TestLoadBasePublication_IdempotentForSameStoryRow(t *testing.T) {
	i := newPublicationTestImporter()
	story := i.addStory(&db.Story{Hash: "story-x"})

	r := newPubRow(i, 0, map[string]string{
		"story_title": "Tarina X",
		"pub_year":    "1975",
		"pub_from":    "1",
		"pub_to":      "2",
	})

	if err := i.loadBasePublication(story.ID, r); err != nil {
		t.Fatalf("first loadBasePublication failed: %v", err)
	}
	if err := i.loadBasePublication(story.ID, r); err != nil {
		t.Fatalf("second loadBasePublication failed: %v", err)
	}

	if len(i.publications) != 2 {
		t.Fatalf("expected 2 publications, got %d", len(i.publications))
	}
	if len(i.storyPublications) != 2 {
		t.Fatalf("expected 2 story-publication links, got %d", len(i.storyPublications))
	}
}

// 4. Italian numbers do not wrap
// Italian vuosi=1974, alkunumero=160, päättymisnumero=162 → exactly three
// publications (italia_perus, 1974, "160"|"161"|"162").
func TestLoadItalianBasePublication_DoesNotWrap(t *testing.T) {
	i := newPublicationTestImporter()
	story := i.addStory(&db.Story{Hash: "story-italia"})

	r := newPubRow(i, 0, map[string]string{
		"italy_story_title": "La morte viene dal cielo",
		"italy_year":        "1974",
		"italy_pub_from":    "160",
		"italy_pub_to":      "162",
	})

	if err := i.loadItalianBasePublication(story.ID, r); err != nil {
		t.Fatalf("loadItalianBasePublication failed: %v", err)
	}

	for _, num := range []string{"160", "161", "162"} {
		if findPublication(i, PUB_IT_PERUS, 1974, num) == nil {
			t.Errorf("expected publication (italia_perus, 1974, %q)", num)
		}
	}

	if len(i.publications) != 3 {
		t.Fatalf("expected exactly 3 publications, got %d", len(i.publications))
	}
}

// 5. Italian high-number boundary
// Italian vuosi=1975, alkunumero=170, päättymisnumero=172 stays as
// 170/1975, 171/1975, 172/1975 (would have wrapped to 2/1975, 3/1975, 4/1975
// under the Finnish logic where annualCount=12 for 1971–78).
func TestLoadItalianBasePublication_HighNumbersDoNotModulo(t *testing.T) {
	i := newPublicationTestImporter()
	story := i.addStory(&db.Story{Hash: "story-italia-2"})

	r := newPubRow(i, 0, map[string]string{
		"italy_story_title": "Storia alta",
		"italy_year":        "1975",
		"italy_pub_from":    "170",
		"italy_pub_to":      "172",
	})

	if err := i.loadItalianBasePublication(story.ID, r); err != nil {
		t.Fatalf("loadItalianBasePublication failed: %v", err)
	}

	for _, num := range []string{"170", "171", "172"} {
		if findPublication(i, PUB_IT_PERUS, 1975, num) == nil {
			t.Errorf("expected publication (italia_perus, 1975, %q)", num)
		}
	}
	for _, num := range []string{"2", "3", "4"} {
		if findPublication(i, PUB_IT_PERUS, 1975, num) != nil {
			t.Errorf("did not expect publication (italia_perus, 1975, %q) — Italian must not wrap modulo annualCount", num)
		}
	}
}

// 6. Shared special publication across stories
// Two stories listing the same Kronikka value both end up linked to a single
// publication row.
func TestLoadKronikka_SharedAcrossStories(t *testing.T) {
	i := newPublicationTestImporter()
	storyA := i.addStory(&db.Story{Hash: "story-a"})
	storyB := i.addStory(&db.Story{Hash: "story-b"})

	rowA := newPubRow(i, 0, map[string]string{
		"story_title":  "Tarina A;Kronikka A",
		"pub_kronikka": "Tex Kronikka 12",
	})
	rowB := newPubRow(i, 1, map[string]string{
		"story_title":  "Tarina B;Kronikka B",
		"pub_kronikka": "Tex Kronikka 12",
	})

	if err := i.loadKronikka(storyA.ID, rowA); err != nil {
		t.Fatalf("loadKronikka A failed: %v", err)
	}
	if err := i.loadKronikka(storyB.ID, rowB); err != nil {
		t.Fatalf("loadKronikka B failed: %v", err)
	}

	kronikkaPubs := 0
	var pubID id
	for _, p := range i.publications {
		if p.item.Type == PUB_KRONIKKA {
			kronikkaPubs++
			pubID = p.ID
		}
	}
	if kronikkaPubs != 1 {
		t.Fatalf("expected exactly 1 kronikka publication, got %d", kronikkaPubs)
	}

	links := storyPublicationsFor(i, pubID)
	if len(links) != 2 {
		t.Fatalf("expected 2 story-publication links for shared kronikka, got %d", len(links))
	}

	gotStories := map[id]bool{}
	for _, l := range links {
		gotStories[l.story] = true
	}
	if !gotStories[storyA.ID] || !gotStories[storyB.ID] {
		t.Fatalf("both stories should link to shared kronikka, got %#v", gotStories)
	}
}

// Sanity: italian range with to < from must error (no silent empty range).
func TestLoadItalianBasePublication_InvalidRangeErrors(t *testing.T) {
	i := newPublicationTestImporter()
	story := i.addStory(&db.Story{Hash: "story-bad"})

	r := newPubRow(i, 0, map[string]string{
		"italy_story_title": "Bad",
		"italy_year":        "1974",
		"italy_pub_from":    "5",
		"italy_pub_to":      "3",
	})

	err := i.loadItalianBasePublication(story.ID, r)
	if err == nil {
		t.Fatalf("expected error for italian range to<from")
	}
}

// Bug A: 1979 wrap-around must use annualCount=13, not 12.
// Row Vuosi=1979, Alkaen=13, Päättyen=2/80 must emit issue 13/1979,
// 1/1980, 2/1980 — previously issue 13/1979 was silently dropped.
func TestLoadBasePublication_1979WrapUsesCorrectCadence(t *testing.T) {
	i := newPublicationTestImporter()
	story := i.addStory(&db.Story{Hash: "story-1979"})

	r := newPubRow(i, 0, map[string]string{
		"story_title": "Tarina 1979",
		"pub_year":    "1979",
		"pub_from":    "13",
		"pub_to":      "2/80",
	})

	if err := i.loadBasePublication(story.ID, r); err != nil {
		t.Fatalf("loadBasePublication failed: %v", err)
	}

	for _, want := range []struct {
		year  int
		issue string
	}{{1979, "13"}, {1980, "1"}, {1980, "2"}} {
		if findPublication(i, PUB_PERUS, want.year, want.issue) == nil {
			t.Errorf("expected publication (perus, %d, %q)", want.year, want.issue)
		}
	}
}

// Bug A: 1980+ wrap-around must use annualCount=16, not 12.
// Row Vuosi=1980, Alkaen=16, Päättyen=2/81 previously produced NO publications
// (loop range was 16..14 — empty). After the fix it must emit 16/1980, 1/1981, 2/1981.
func TestLoadBasePublication_1980WrapUsesCorrectCadence(t *testing.T) {
	i := newPublicationTestImporter()
	story := i.addStory(&db.Story{Hash: "story-1980"})

	r := newPubRow(i, 0, map[string]string{
		"story_title": "Tarina 1980",
		"pub_year":    "1980",
		"pub_from":    "16",
		"pub_to":      "2/81",
	})

	if err := i.loadBasePublication(story.ID, r); err != nil {
		t.Fatalf("loadBasePublication failed: %v", err)
	}

	for _, want := range []struct {
		year  int
		issue string
	}{{1980, "16"}, {1981, "1"}, {1981, "2"}} {
		if findPublication(i, PUB_PERUS, want.year, want.issue) == nil {
			t.Errorf("expected publication (perus, %d, %q)", want.year, want.issue)
		}
	}
}

// Bug A: 1985 mid-year-to-next-year must emit every issue in between,
// not just the wrap endpoints. Cadence is 16, so Alkaen=12 to Päättyen=1/86
// must produce issues 12..16/1985 + 1/1986.
func TestLoadBasePublication_1985WrapEmitsAllIntermediateIssues(t *testing.T) {
	i := newPublicationTestImporter()
	story := i.addStory(&db.Story{Hash: "story-1985"})

	r := newPubRow(i, 0, map[string]string{
		"story_title": "Tarina 1985",
		"pub_year":    "1985",
		"pub_from":    "12",
		"pub_to":      "1/86",
	})

	if err := i.loadBasePublication(story.ID, r); err != nil {
		t.Fatalf("loadBasePublication failed: %v", err)
	}

	for _, want := range []struct {
		year  int
		issue string
	}{{1985, "12"}, {1985, "13"}, {1985, "14"}, {1985, "15"}, {1985, "16"}, {1986, "1"}} {
		if findPublication(i, PUB_PERUS, want.year, want.issue) == nil {
			t.Errorf("expected publication (perus, %d, %q)", want.year, want.issue)
		}
	}
}

// Bug C: pub_special values separated by ';' must produce one publication per value
// with the correct type for each, and both must link to the story.
// "Suuralbumi 6 (2002); Pulp-pokkari 6 (2011)" — first classifies as suur, the
// second is muu_erikois. Before the fix the whole string became one suur publication.
func TestLoadSpecialPublication_SplitsSemicolonValues(t *testing.T) {
	i := newPublicationTestImporter()
	story := i.addStory(&db.Story{Hash: "story-colorado"})

	r := newPubRow(i, 0, map[string]string{
		"story_title": "Coloradon kaivoskahakka",
		"pub_special": "Suuralbumi 6 (2002); Pulp-pokkari 6 (2011)",
	})

	if err := i.loadSpecialPublication(story.ID, r); err != nil {
		t.Fatalf("loadSpecialPublication failed: %v", err)
	}

	if findPublicationByIssue(i, PUB_SUUR, "Suuralbumi 6 (2002)") == nil {
		t.Errorf("expected publication (suur, %q)", "Suuralbumi 6 (2002)")
	}
	if findPublicationByIssue(i, PUB_MUU, "Pulp-pokkari 6 (2011)") == nil {
		t.Errorf("expected publication (muu_erikois, %q)", "Pulp-pokkari 6 (2011)")
	}

	if len(i.publications) != 2 {
		t.Fatalf("expected 2 publications, got %d", len(i.publications))
	}
	if len(i.storyPublications) != 2 {
		t.Fatalf("expected 2 story-publication links, got %d", len(i.storyPublications))
	}
}

// Bug C: italian special publication must also split on ';'.
func TestLoadItalianSpecialPublication_SplitsSemicolonValues(t *testing.T) {
	i := newPublicationTestImporter()
	story := i.addStory(&db.Story{Hash: "story-it-special"})

	r := newPubRow(i, 0, map[string]string{
		"italy_story_title": "Storia",
		"italy_pub_special": "Texone 1; Color Tex 2",
	})

	if err := i.loadItalianSpecialPublication(story.ID, r); err != nil {
		t.Fatalf("loadItalianSpecialPublication failed: %v", err)
	}

	if findPublicationByIssue(i, PUB_IT_TEXONE, "Texone 1") == nil {
		t.Errorf("expected publication (italia_texone, %q)", "Texone 1")
	}
	if findPublicationByIssue(i, PUB_IT_COLOR_TEX, "Color Tex 2") == nil {
		t.Errorf("expected publication (italia_color_tex, %q)", "Color Tex 2")
	}

	if len(i.publications) != 2 {
		t.Fatalf("expected 2 publications, got %d", len(i.publications))
	}
}

// Bug B: loadKirjasto must use the PUB_KIRJASTO title index, not PUB_KRONIKKA's.
// With pub_from + repub_from + pub_kronikka all set and 3 titles, Kronikka claims
// titles[2] and Kirjasto's index becomes 3 — out of range, so Kirjasto must fall
// back to titles[0]. Before the fix Kirjasto was reusing titles[2] (Kronikka's title).
func TestLoadKirjasto_UsesKirjastoTitleIndex(t *testing.T) {
	i := newPublicationTestImporter()
	story := i.addStory(&db.Story{Hash: "story-laredo"})

	r := newPubRow(i, 0, map[string]string{
		"story_title":  "Rajua peliä Laredossa (1. p); Ystävä pulassa (2. p); Laredon tiellä (3. p)",
		"pub_year":     "1965",
		"pub_from":     "1",
		"pub_to":       "2",
		"repub_year":   "1980",
		"repub_from":   "1",
		"repub_to":     "2",
		"pub_kronikka": "5",
		"pub_kirjasto": "24-25",
	})

	if err := i.loadKronikka(story.ID, r); err != nil {
		t.Fatalf("loadKronikka failed: %v", err)
	}
	if err := i.loadKirjasto(story.ID, r); err != nil {
		t.Fatalf("loadKirjasto failed: %v", err)
	}

	var kronikkaTitle, kirjastoTitle string
	for _, sp := range i.storyPublications {
		pub := i.getPublicationWithID(sp.publication)
		switch pub.item.Type {
		case PUB_KRONIKKA:
			kronikkaTitle = sp.title
		case PUB_KIRJASTO:
			kirjastoTitle = sp.title
		}
	}

	if kronikkaTitle != "Laredon tiellä (3. p)" {
		t.Errorf("kronikka title: got %q, want %q", kronikkaTitle, "Laredon tiellä (3. p)")
	}
	if kirjastoTitle != "Rajua peliä Laredossa (1. p)" {
		t.Errorf("kirjasto title: got %q, want %q (titles[0] fallback)",
			kirjastoTitle, "Rajua peliä Laredossa (1. p)")
	}
}

// Bug F: a Päättyen value with /yy must match outerYear+1. A multi-year wrap
// would silently produce the wrong year because parseIssueNum strips /yy.
func TestLoadBasePublication_MultiYearWrapErrors(t *testing.T) {
	i := newPublicationTestImporter()
	story := i.addStory(&db.Story{Hash: "story-multiyear"})

	r := newPubRow(i, 0, map[string]string{
		"story_title": "Tarina monivuotinen",
		"pub_year":    "1974",
		"pub_from":    "12",
		"pub_to":      "1/76",
	})

	err := i.loadBasePublication(story.ID, r)
	if err == nil {
		t.Fatalf("expected error for multi-year wrap (1974 → 1/76)")
	}
}

// Bug F: 4-digit explicit wrap year is accepted.
func TestLoadBasePublication_FourDigitWrapYearAccepted(t *testing.T) {
	i := newPublicationTestImporter()
	story := i.addStory(&db.Story{Hash: "story-4digit"})

	r := newPubRow(i, 0, map[string]string{
		"story_title": "Tarina nelidigit",
		"pub_year":    "1974",
		"pub_from":    "12",
		"pub_to":      "1/1975",
	})

	if err := i.loadBasePublication(story.ID, r); err != nil {
		t.Fatalf("expected 4-digit /1975 to be accepted, got: %v", err)
	}
}

// Bug F: Päättyen with parenthetical annotation ("1/99 (ei 13-16/98)")
// must still validate cleanly because the annotation is stripped first.
func TestLoadBaseRePublication_ParentheticalAnnotationAccepted(t *testing.T) {
	i := newPublicationTestImporter()
	story := i.addStory(&db.Story{Hash: "story-annotated"})

	r := newPubRow(i, 0, map[string]string{
		"story_title": "Tarina A; Tarina A repub",
		"repub_year":  "1998",
		"repub_from":  "12",
		"repub_to":    "1/99 (ei 13-16/98)",
	})

	if err := i.loadBaseRePublication(story.ID, r); err != nil {
		t.Fatalf("expected annotated wrap to validate, got: %v", err)
	}
}

// Bug G: parseNonBaseTitle must only count a base/repub slot when ALL three
// (year, from, to) columns are set — otherwise no base/repub publication is
// created, so the title slot is not occupied.
func TestParseNonBaseTitle_BumpsOnlyWhenFullTriplePresent(t *testing.T) {
	i := newPublicationTestImporter()

	// pub_from set but pub_year missing → no base publication created, so the
	// title slot is not occupied. Special publication should use titles[0].
	r := newPubRow(i, 0, map[string]string{
		"story_title": "Only special",
		"pub_year":    "",
		"pub_from":    "5",
		"pub_to":      "5",
		"pub_special": "Suuralbumi 1",
	})

	title, err := i.parseNonBaseTitle(PUB_SUUR, r)
	if err != nil {
		t.Fatalf("parseNonBaseTitle failed: %v", err)
	}
	if title != "Only special" {
		t.Errorf("special title: got %q, want %q", title, "Only special")
	}
}

// Bug H: "(ei X-Y/yy)" annotation in a wrap-around must exclude the named
// range. UVuosi=1998, Ualkaen=12, Upäättyen='1/99 (ei 13-16/98)' should
// emit (perus, 1998, 12) and (perus, 1999, 1) only.
func TestLoadBaseRePublication_AnnotationExcludesWrapRange(t *testing.T) {
	i := newPublicationTestImporter()
	story := i.addStory(&db.Story{Hash: "story-billy"})

	r := newPubRow(i, 0, map[string]string{
		"story_title": "Billy Kidin konnakopla (1. – 2. p); Billy Kidin kopla (3. – 4. p)",
		"repub_year":  "1998",
		"repub_from":  "12",
		"repub_to":    "1/99 (ei 13-16/98)",
	})

	if err := i.loadBaseRePublication(story.ID, r); err != nil {
		t.Fatalf("loadBaseRePublication failed: %v", err)
	}

	for _, want := range []struct {
		year  int
		issue string
	}{{1998, "12"}, {1999, "1"}} {
		if findPublication(i, PUB_PERUS, want.year, want.issue) == nil {
			t.Errorf("expected publication (perus, %d, %q)", want.year, want.issue)
		}
	}
	for _, no := range []struct {
		year  int
		issue string
	}{{1998, "13"}, {1998, "14"}, {1998, "15"}, {1998, "16"}} {
		if findPublication(i, PUB_PERUS, no.year, no.issue) != nil {
			t.Errorf("did not expect publication (perus, %d, %q) — should be excluded by (ei …)", no.year, no.issue)
		}
	}
}

// Bug H: in-year "(ei X-Y)" annotation. UVuosi=1999, Ualkaen=1,
// Upäättyen='16 (ei 3-15)' should emit issues 1, 2, 16 of 1999.
func TestLoadBaseRePublication_AnnotationExcludesInYearRange(t *testing.T) {
	i := newPublicationTestImporter()
	story := i.addStory(&db.Story{Hash: "story-santafe"})

	r := newPubRow(i, 0, map[string]string{
		"story_title": "Santa Fen tuho (1. p); Santa Fen hyökkäys (2. – 3. p)",
		"repub_year":  "1999",
		"repub_from":  "1",
		"repub_to":    "16 (ei 3-15)",
	})

	if err := i.loadBaseRePublication(story.ID, r); err != nil {
		t.Fatalf("loadBaseRePublication failed: %v", err)
	}

	for _, want := range []string{"1", "2", "16"} {
		if findPublication(i, PUB_PERUS, 1999, want) == nil {
			t.Errorf("expected publication (perus, 1999, %q)", want)
		}
	}
	if len(i.publications) != 3 {
		t.Fatalf("expected exactly 3 publications, got %d", len(i.publications))
	}
}

// Bug H: explicit /yy on Päättyen detects wrap even when to==from. UVuosi=2008,
// Ualkaen=1, Upäättyen='1/09 (ei 2-15/08)' should emit (perus, 2008, 1),
// (perus, 2008, 16), (perus, 2009, 1).
func TestLoadBaseRePublication_ExplicitYearTriggersWrap(t *testing.T) {
	i := newPublicationTestImporter()
	story := i.addStory(&db.Story{Hash: "story-kuolemanpolku"})

	r := newPubRow(i, 0, map[string]string{
		"story_title": "Kuoleman polku (1. p); Kuoleman polulla (2. – 3. p)",
		"repub_year":  "2008",
		"repub_from":  "1",
		"repub_to":    "1/09 (ei 2-15/08)",
	})

	if err := i.loadBaseRePublication(story.ID, r); err != nil {
		t.Fatalf("loadBaseRePublication failed: %v", err)
	}

	for _, want := range []struct {
		year  int
		issue string
	}{{2008, "1"}, {2008, "16"}, {2009, "1"}} {
		if findPublication(i, PUB_PERUS, want.year, want.issue) == nil {
			t.Errorf("expected publication (perus, %d, %q)", want.year, want.issue)
		}
	}
	if len(i.publications) != 3 {
		t.Fatalf("expected exactly 3 publications, got %d", len(i.publications))
	}
}

// Bug H: in-year exclusion on base publication. Vuosi=2009, Alkaen=2,
// Päättyen='6 (ei 3-4)' should emit (perus, 2009, 2/5/6).
func TestLoadBasePublication_AnnotationExcludesInYearBase(t *testing.T) {
	i := newPublicationTestImporter()
	story := i.addStory(&db.Story{Hash: "story-bourbon"})

	r := newPubRow(i, 0, map[string]string{
		"story_title": "Bourbon Streetin murhat",
		"pub_year":    "2009",
		"pub_from":    "2",
		"pub_to":      "6 (ei 3-4)",
	})

	if err := i.loadBasePublication(story.ID, r); err != nil {
		t.Fatalf("loadBasePublication failed: %v", err)
	}

	for _, want := range []string{"2", "5", "6"} {
		if findPublication(i, PUB_PERUS, 2009, want) == nil {
			t.Errorf("expected publication (perus, 2009, %q)", want)
		}
	}
	for _, no := range []string{"3", "4"} {
		if findPublication(i, PUB_PERUS, 2009, no) != nil {
			t.Errorf("did not expect publication (perus, 2009, %q)", no)
		}
	}
}

// Bug A: unknown cadence (no branch in getPublishedAnnualCount) must surface
// as a clear error rather than silently producing issue 0 via i % -1.
func TestLoadBasePublication_UnknownCadenceErrors(t *testing.T) {
	i := newPublicationTestImporter()
	story := i.addStory(&db.Story{Hash: "story-unknown"})

	// Year 1968 has no cadence branch; importer must refuse rather than
	// silently produce garbage.
	r := newPubRow(i, 0, map[string]string{
		"story_title": "Tarina 1968",
		"pub_year":    "1968",
		"pub_from":    "5",
		"pub_to":      "6",
	})

	err := i.loadBasePublication(story.ID, r)
	if err == nil {
		t.Fatalf("expected error for unknown annual cadence")
	}
}
