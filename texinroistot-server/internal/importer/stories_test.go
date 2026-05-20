package importer

import (
	"testing"
)

func newStoryTestImporter() *importer {
	return &importer{
		columnIndexes: map[string]int{
			"story_order_num": 0,
			"story_title":     1,
		},
	}
}

func newStoryRow(i *importer, rowIndex int, values map[string]string) row {
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

// Bug E: rows with empty Järjestysluku used to all collide on Hash("-1").
// Two rows with empty order and different titles must become two stories.
func TestLoadStory_EmptyOrderNumDoesNotCollideAcrossDifferentTitles(t *testing.T) {
	i := newStoryTestImporter()

	idA, err := i.loadStory(newStoryRow(i, 0, map[string]string{
		"story_order_num": "",
		"story_title":     "Tarina A",
	}))
	if err != nil {
		t.Fatalf("loadStory A failed: %v", err)
	}
	idB, err := i.loadStory(newStoryRow(i, 1, map[string]string{
		"story_order_num": "",
		"story_title":     "Tarina B",
	}))
	if err != nil {
		t.Fatalf("loadStory B failed: %v", err)
	}

	if idA == idB {
		t.Fatalf("expected separate story IDs for different titles, both got %d", idA)
	}
	if len(i.stories) != 2 {
		t.Fatalf("expected 2 stories, got %d", len(i.stories))
	}
}

// Bug E: rows with the same empty Järjestysluku and the same title must still
// be the same story (multi-villain rows for one untracked story).
func TestLoadStory_EmptyOrderNumSameTitleSharesStory(t *testing.T) {
	i := newStoryTestImporter()

	idA, err := i.loadStory(newStoryRow(i, 0, map[string]string{
		"story_order_num": "",
		"story_title":     "Tarina X",
	}))
	if err != nil {
		t.Fatalf("loadStory A failed: %v", err)
	}
	idB, err := i.loadStory(newStoryRow(i, 1, map[string]string{
		"story_order_num": "",
		"story_title":     "Tarina X",
	}))
	if err != nil {
		t.Fatalf("loadStory B failed: %v", err)
	}

	if idA != idB {
		t.Fatalf("expected shared story for same title, got %d and %d", idA, idB)
	}
	if len(i.stories) != 1 {
		t.Fatalf("expected 1 story, got %d", len(i.stories))
	}
}

// Existing behaviour: positive order numbers continue to deduplicate by order
// number regardless of title.
func TestLoadStory_PositiveOrderNumDedupesAcrossVillainRows(t *testing.T) {
	i := newStoryTestImporter()

	idA, err := i.loadStory(newStoryRow(i, 0, map[string]string{
		"story_order_num": "42",
		"story_title":     "Tarina A",
	}))
	if err != nil {
		t.Fatalf("loadStory A failed: %v", err)
	}
	idB, err := i.loadStory(newStoryRow(i, 1, map[string]string{
		"story_order_num": "42",
		"story_title":     "Tarina A",
	}))
	if err != nil {
		t.Fatalf("loadStory B failed: %v", err)
	}

	if idA != idB {
		t.Fatalf("expected same story id for same order num, got %d and %d", idA, idB)
	}
	if len(i.stories) != 1 {
		t.Fatalf("expected 1 story, got %d", len(i.stories))
	}
}
