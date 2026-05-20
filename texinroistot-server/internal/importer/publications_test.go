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
