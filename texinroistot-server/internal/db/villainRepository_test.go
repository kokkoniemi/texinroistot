package db

import (
	"strings"
	"testing"
)

func TestBuildVillainListWhere_SearchUsesOnlyVillainIdentityFields(t *testing.T) {
	whereClause, args, err := buildVillainListWhere(17, VillainListParams{
		Publication: "all",
		Search:      "Don",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expectedParts := []string{
		"array_to_string(v.first_names, ' ') ILIKE $2",
		"v.last_name ILIKE $2",
		"array_to_string(v.ranks, ' ') ILIKE $2",
		"array_to_string(vis.nicknames, ' ') ILIKE $2",
		"array_to_string(vis.other_names, ' ') ILIKE $2",
		"array_to_string(vis.code_names, ' ') ILIKE $2",
	}
	for _, part := range expectedParts {
		if !strings.Contains(whereClause, part) {
			t.Fatalf("expected where clause to include %q, got %s", part, whereClause)
		}
	}

	disallowedParts := []string{
		"array_to_string(vis.roles, ' ') ILIKE",
		"array_to_string(vis.destiny, ' ') ILIKE",
		"sip.title ILIKE",
		"p.issue ILIKE",
		"p.type::text ILIKE",
		"JOIN stories AS s ON s.id = vis.story",
		"JOIN stories_in_publications AS sip ON sip.story = s.id",
		"JOIN publications AS p ON p.id = sip.publication",
	}
	for _, part := range disallowedParts {
		if strings.Contains(whereClause, part) {
			t.Fatalf("expected where clause to exclude %q, got %s", part, whereClause)
		}
	}

	if len(args) != 2 {
		t.Fatalf("expected 2 args (version + search), got %d", len(args))
	}
	if args[0] != 17 {
		t.Fatalf("expected version id arg to be 17, got %v", args[0])
	}
	if args[1] != "%Don%" {
		t.Fatalf("expected search arg to be %%Don%%, got %v", args[1])
	}
}

func TestBuildVillainListWhere_SearchArgPositionWithPublicationFilter(t *testing.T) {
	whereClause, args, err := buildVillainListWhere(5, VillainListParams{
		Publication: "fi",
		Search:      "kit",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !strings.Contains(whereClause, "AND p.type::text = ANY($2)") {
		t.Fatalf("expected publication filter to use $2, got %s", whereClause)
	}
	if !strings.Contains(whereClause, "array_to_string(v.first_names, ' ') ILIKE $3") {
		t.Fatalf("expected search to use $3 with publication filter, got %s", whereClause)
	}

	if len(args) != 3 {
		t.Fatalf("expected 3 args (version + publication + search), got %d", len(args))
	}
	if args[0] != 5 {
		t.Fatalf("expected version id arg to be 5, got %v", args[0])
	}
	if args[2] != "%kit%" {
		t.Fatalf("expected search arg to be %%kit%%, got %v", args[2])
	}
}
