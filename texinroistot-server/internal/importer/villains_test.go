package importer

import (
	"reflect"
	"strings"
	"testing"

	"github.com/kokkoniemi/texinroistot/internal/db"
)

func newVillainTestImporter() *importer {
	return &importer{
		columnIndexes: map[string]int{
			"villain_id":  0,
			"ranks":       1,
			"first_names": 2,
			"last_name":   3,
			"nicknames":   4,
			"other_names": 5,
			"code_names":  6,
			"roles":       7,
			"destiny":     8,
		},
	}
}

func newVillainRow(i *importer, rowIndex int, values map[string]string) row {
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

	return row{
		importer: i,
		index:    rowIndex,
		cells:    cells,
	}
}

func TestLoadVillain_AllowsSameVillainIDAcrossMultipleStories(t *testing.T) {
	i := newVillainTestImporter()
	storyA := i.addStory(&db.Story{Hash: "story-a"})
	storyB := i.addStory(&db.Story{Hash: "story-b"})

	rowA := newVillainRow(i, 0, map[string]string{
		"villain_id":  "101",
		"ranks":       "Kapteeni",
		"first_names": "John",
		"last_name":   "Doe",
		"nicknames":   "The Fox",
		"other_names": "Kettu",
		"code_names":  "Ghost",
		"roles":       "johtaja",
		"destiny":     "pakeni",
	})
	rowB := newVillainRow(i, 1, map[string]string{
		"villain_id":  "101",
		"ranks":       "Kapteeni",
		"first_names": "John",
		"last_name":   "Doe",
		"nicknames":   "The Fox",
		"other_names": "Kettu",
		"code_names":  "Ghost",
		"roles":       "vakooja",
		"destiny":     "vangittiin",
	})

	if err := i.loadVillain(storyA.ID, rowA); err != nil {
		t.Fatalf("first loadVillain failed: %v", err)
	}
	if err := i.loadVillain(storyB.ID, rowB); err != nil {
		t.Fatalf("second loadVillain failed: %v", err)
	}

	if len(i.villains) != 1 {
		t.Fatalf("expected one villain, got %d", len(i.villains))
	}
	if len(i.storyVillains) != 2 {
		t.Fatalf("expected two story-villain entries, got %d", len(i.storyVillains))
	}

	villain := i.villains[0]
	storyVillainA := i.getStoryVillain(villain.ID, storyA.ID)
	storyVillainB := i.getStoryVillain(villain.ID, storyB.ID)
	if storyVillainA == nil || storyVillainB == nil {
		t.Fatalf("expected story villain entries for both stories")
	}

	if !reflect.DeepEqual(storyVillainA.item.Roles, []string{"johtaja"}) {
		t.Fatalf("unexpected roles for story A: %#v", storyVillainA.item.Roles)
	}
	if !reflect.DeepEqual(storyVillainB.item.Roles, []string{"vakooja"}) {
		t.Fatalf("unexpected roles for story B: %#v", storyVillainB.item.Roles)
	}
	if !reflect.DeepEqual(storyVillainA.item.Destiny, []string{"pakeni"}) {
		t.Fatalf("unexpected destiny for story A: %#v", storyVillainA.item.Destiny)
	}
	if !reflect.DeepEqual(storyVillainB.item.Destiny, []string{"vangittiin"}) {
		t.Fatalf("unexpected destiny for story B: %#v", storyVillainB.item.Destiny)
	}
}

func TestLoadVillain_FailsWhenSameVillainIDAppearsTwiceInSameStory(t *testing.T) {
	i := newVillainTestImporter()
	storyA := i.addStory(&db.Story{Hash: "story-a"})

	firstRow := newVillainRow(i, 0, map[string]string{
		"villain_id":  "101",
		"ranks":       "Kapteeni",
		"first_names": "John",
		"last_name":   "Doe",
		"nicknames":   "The Fox",
		"other_names": "Kettu",
		"code_names":  "Ghost",
		"roles":       "johtaja",
		"destiny":     "pakeni",
	})
	secondRow := newVillainRow(i, 1, map[string]string{
		"villain_id":  "101",
		"ranks":       "Kapteeni",
		"first_names": "John",
		"last_name":   "Doe",
		"nicknames":   "The Fox",
		"other_names": "Kettu",
		"code_names":  "Ghost",
		"roles":       "vakooja",
		"destiny":     "vangittiin",
	})

	if err := i.loadVillain(storyA.ID, firstRow); err != nil {
		t.Fatalf("first loadVillain failed: %v", err)
	}
	err := i.loadVillain(storyA.ID, secondRow)
	if err == nil {
		t.Fatalf("expected duplicate same story to fail")
	}
	if !strings.Contains(err.Error(), "duplicate villain_id 101") {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(i.storyVillains) != 1 {
		t.Fatalf("expected only one story-villain entry after duplicate error, got %d", len(i.storyVillains))
	}
	if !reflect.DeepEqual(i.storyVillains[0].item.Roles, []string{"johtaja"}) {
		t.Fatalf("unexpected roles in stored entry: %#v", i.storyVillains[0].item.Roles)
	}
}
