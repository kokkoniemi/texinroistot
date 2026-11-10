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

func TestLoadVillain_KeepsNamePerStoryAndUsesLatestForVillain(t *testing.T) {
	i := newVillainTestImporter()
	// added out of story order on purpose: the latest story decides the name
	later := i.addStory(&db.Story{Hash: "yuman-vanki", OrderNumber: 446})
	earlier := i.addStory(&db.Story{Hash: "seitseman-tappajaa", OrderNumber: 307})

	laterRow := newVillainRow(i, 0, map[string]string{
		"villain_id":  "202",
		"first_names": "Philip",
		"last_name":   "Rainey",
		"nicknames":   "Kid Rodelo",
	})
	earlierRow := newVillainRow(i, 1, map[string]string{
		"villain_id":  "202",
		"first_names": "Kid",
		"last_name":   "Rodelo",
	})

	if err := i.loadVillain(later.ID, laterRow); err != nil {
		t.Fatalf("loadVillain (later story) failed: %v", err)
	}
	if err := i.loadVillain(earlier.ID, earlierRow); err != nil {
		t.Fatalf("loadVillain (earlier story) failed: %v", err)
	}

	villain := i.villains[0]
	earlierSV := i.getStoryVillain(villain.ID, earlier.ID)
	laterSV := i.getStoryVillain(villain.ID, later.ID)
	if !reflect.DeepEqual(earlierSV.item.FirstNames, []string{"Kid"}) || earlierSV.item.LastName != "Rodelo" {
		t.Fatalf("earlier story name: %#v %q", earlierSV.item.FirstNames, earlierSV.item.LastName)
	}
	if !reflect.DeepEqual(laterSV.item.FirstNames, []string{"Philip"}) || laterSV.item.LastName != "Rainey" {
		t.Fatalf("later story name: %#v %q", laterSV.item.FirstNames, laterSV.item.LastName)
	}

	villain.item.As = i.getStoryVillainItems(villain.ID)
	applyLatestStoryNames(villain.item)
	if !reflect.DeepEqual(villain.item.FirstNames, []string{"Philip"}) || villain.item.LastName != "Rainey" {
		t.Fatalf("villain name: %#v %q", villain.item.FirstNames, villain.item.LastName)
	}
}

func TestApplyLatestStoryNames_CombinesFirstNamesUnderTheSameLastName(t *testing.T) {
	cases := []struct {
		name  string
		as    []*db.StoryVillain
		first []string
		last  string
	}{
		{
			name: "shorter name contained in the longer one",
			as: []*db.StoryVillain{
				{FirstNames: []string{`Raymond "Ray"`}, LastName: "Clemmons", Story: &db.Story{OrderNumber: 268}},
				{FirstNames: []string{"Ray"}, LastName: "Clemmons", Story: &db.Story{OrderNumber: 307}},
			},
			first: []string{`Raymond "Ray"`},
			last:  "Clemmons",
		},
		{
			name: "story order decides, not row order",
			as: []*db.StoryVillain{
				{FirstNames: []string{"Kenneth"}, LastName: "Bowen", Story: &db.Story{OrderNumber: 541}},
				{FirstNames: []string{`Kenneth "Ken"`}, LastName: "Bowen", Story: &db.Story{OrderNumber: 495}},
			},
			first: []string{`Kenneth "Ken"`},
			last:  "Bowen",
		},
		{
			name: "a first name spelling variant goes in parentheses",
			as: []*db.StoryVillain{
				{FirstNames: []string{"Mickey"}, LastName: "Welt", Story: &db.Story{OrderNumber: 98}},
				{FirstNames: []string{"Miky"}, LastName: "Welt", Story: &db.Story{OrderNumber: 518}},
			},
			first: []string{"Mickey (Miky)"},
			last:  "Welt",
		},
		{
			name: "a last name spelling variant goes in parentheses and shares first names",
			as: []*db.StoryVillain{
				{FirstNames: []string{""}, LastName: "Quantrell", Story: &db.Story{OrderNumber: 33}},
				{FirstNames: []string{"Samuel"}, LastName: "Quantrell", Story: &db.Story{OrderNumber: 34}},
				{FirstNames: []string{""}, LastName: "Quantrill", Story: &db.Story{OrderNumber: 145}},
				{FirstNames: []string{""}, LastName: "Quantrill", Story: &db.Story{OrderNumber: 393}},
			},
			first: []string{"Samuel"},
			last:  "Quantrill (Quantrell)",
		},
		{
			name: "different first names are both kept",
			as: []*db.StoryVillain{
				{FirstNames: []string{"John"}, LastName: "Doe", Story: &db.Story{OrderNumber: 1}},
				{FirstNames: []string{"Peter"}, LastName: "Doe", Story: &db.Story{OrderNumber: 2}},
			},
			first: []string{"John", "Peter"},
			last:  "Doe",
		},
		{
			name: "a story without a last name shares it; one without names changes nothing",
			as: []*db.StoryVillain{
				{FirstNames: []string{"John"}, LastName: "Doe", Story: &db.Story{OrderNumber: 10}},
				{FirstNames: []string{"John"}, LastName: "", Story: &db.Story{OrderNumber: 20}},
				{FirstNames: []string{""}, LastName: "", Story: &db.Story{OrderNumber: 30}},
			},
			first: []string{"John"},
			last:  "Doe",
		},
		{
			name: "a new last name leaves the earlier first names out",
			as: []*db.StoryVillain{
				{FirstNames: []string{"Kid"}, LastName: "Rodelo", Story: &db.Story{OrderNumber: 307}},
				{FirstNames: []string{"Philip"}, LastName: "Rainey", Story: &db.Story{OrderNumber: 446}},
			},
			first: []string{"Philip"},
			last:  "Rainey",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			villain := &db.Villain{As: tc.as}
			applyLatestStoryNames(villain)
			if !reflect.DeepEqual(villain.FirstNames, tc.first) {
				t.Fatalf("first names: %#v, want %#v", villain.FirstNames, tc.first)
			}
			if villain.LastName != tc.last {
				t.Fatalf("last name: %q, want %q", villain.LastName, tc.last)
			}
		})
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
