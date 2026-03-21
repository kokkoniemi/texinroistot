package db

import (
	"strings"
	"testing"
)

func TestBuildStoryListWhere_YearUsesSamePublicationExistsWhenPublicationFilterSet(t *testing.T) {
	whereClause, args, err := buildStoryListWhere(7, StoryListParams{
		Publication: "perus_it",
		Year:        1980,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !strings.Contains(whereClause, "AND p.type::text = ANY($2)") {
		t.Fatalf("expected publication filter to use $2, got %s", whereClause)
	}
	if !strings.Contains(whereClause, "AND p.year = $3") {
		t.Fatalf("expected year filter to use $3 in publication EXISTS, got %s", whereClause)
	}

	if len(args) != 3 {
		t.Fatalf("expected 3 args (version + publication + year), got %d", len(args))
	}
	if args[0] != 7 {
		t.Fatalf("expected version id arg to be 7, got %v", args[0])
	}
	if args[2] != 1980 {
		t.Fatalf("expected year arg to be 1980, got %v", args[2])
	}
}

func TestBuildStoryListWhere_YearWithoutPublicationFilter(t *testing.T) {
	whereClause, args, err := buildStoryListWhere(9, StoryListParams{
		Publication: "all",
		Year:        1967,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if strings.Contains(whereClause, "p.type::text = ANY(") {
		t.Fatalf("did not expect publication-type clause for all publications, got %s", whereClause)
	}
	if !strings.Contains(whereClause, "AND p.year = $2") {
		t.Fatalf("expected year filter to use $2 when publication filter is all, got %s", whereClause)
	}

	if len(args) != 2 {
		t.Fatalf("expected 2 args (version + year), got %d", len(args))
	}
	if args[0] != 9 {
		t.Fatalf("expected version id arg to be 9, got %v", args[0])
	}
	if args[1] != 1967 {
		t.Fatalf("expected year arg to be 1967, got %v", args[1])
	}
}

func TestMapPublicationFilterToTypes_ItalianSpecialSeries(t *testing.T) {
	types, err := mapPublicationFilterToTypes("texone")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(types) != 1 {
		t.Fatalf("expected 1 publication type for texone filter, got %d (%v)", len(types), types)
	}
	if types[0] != "italia_texone" {
		t.Fatalf("expected italia_texone publication type, got %q", types[0])
	}
}

func TestBuildSortClause_AlphaUsesSelectedItalianSpecialSeries(t *testing.T) {
	clause := buildSortClause("alpha", "texone")
	if !strings.Contains(clause, "p.type = 'italia_texone'") {
		t.Fatalf("expected alpha sort to use italia_texone publication type, got %s", clause)
	}
}
