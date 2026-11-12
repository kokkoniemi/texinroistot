package db

import (
	"strings"
	"testing"
)

// The search looks in the villain's names and in their roles and destinies,
// but not in story or publication fields.
func TestBuildVillainListWhere_SearchUsesVillainNamesRolesAndDestinies(t *testing.T) {
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
		"array_to_string(vis.roles, ' ') ILIKE $2",
		"array_to_string(vis.destiny, ' ') ILIKE $2",
	}
	for _, part := range expectedParts {
		if !strings.Contains(whereClause, part) {
			t.Fatalf("expected where clause to include %q, got %s", part, whereClause)
		}
	}

	disallowedParts := []string{
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

// The name search looks in names only; the keyword search after it also in
// roles and destinies, each word with its own parameter.
func TestBuildVillainListWhere_NameSearchUsesNamesOnly(t *testing.T) {
	whereClause, args, err := buildVillainListWhere(5, VillainListParams{
		Publication: "all",
		Name:        "Hoodoo",
		Search:      "Vegas",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if strings.Contains(whereClause, "array_to_string(vis.roles, ' ') ILIKE $2") ||
		strings.Contains(whereClause, "array_to_string(vis.destiny, ' ') ILIKE $2") {
		t.Fatalf("expected the name search to skip roles and destinies, got %s", whereClause)
	}
	for _, part := range []string{
		"v.last_name ILIKE $2",
		"vis.last_name ILIKE $3",
		"array_to_string(vis.roles, ' ') ILIKE $3",
		"array_to_string(vis.destiny, ' ') ILIKE $3",
	} {
		if !strings.Contains(whereClause, part) {
			t.Fatalf("expected where clause to include %q, got %s", part, whereClause)
		}
	}
	if len(args) != 3 || args[1] != "%Hoodoo%" || args[2] != "%Vegas%" {
		t.Fatalf("expected args [5 %%Hoodoo%% %%Vegas%%], got %v", args)
	}
}

// A search of several words requires each word in some name field, so a whole
// name such as "Hoodoo Brown" (first name + last name) is found.
func TestBuildVillainListWhere_SearchWordsEachMatchAName(t *testing.T) {
	whereClause, args, err := buildVillainListWhere(5, VillainListParams{
		Publication: "all",
		Search:      "  Hoodoo   Brown ",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	for _, part := range []string{
		"array_to_string(v.first_names, ' ') ILIKE $2",
		"v.last_name ILIKE $2",
		"array_to_string(v.first_names, ' ') ILIKE $3",
		"v.last_name ILIKE $3",
	} {
		if !strings.Contains(whereClause, part) {
			t.Fatalf("expected where clause to include %q, got %s", part, whereClause)
		}
	}
	if len(args) != 3 || args[1] != "%Hoodoo%" || args[2] != "%Brown%" {
		t.Fatalf("expected args [5 %%Hoodoo%% %%Brown%%], got %v", args)
	}
}

func TestBuildVillainListWhere_VillainHashFollowsOtherFilters(t *testing.T) {
	whereClause, args, err := buildVillainListWhere(5, VillainListParams{
		Publication: "fi",
		Search:      "kit",
		VillainHash: " abc123 ",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !strings.Contains(whereClause, "v.hash = $4") {
		t.Fatalf("expected villain hash filter to use $4, got %s", whereClause)
	}
	if len(args) != 4 {
		t.Fatalf("expected 4 args (version + publication + search + hash), got %d", len(args))
	}
	if args[3] != "abc123" {
		t.Fatalf("expected trimmed villain hash arg, got %v", args[3])
	}
}

func TestBuildVillainListWhere_VillainHashAlone(t *testing.T) {
	whereClause, args, err := buildVillainListWhere(5, VillainListParams{
		Publication: "all",
		VillainHash: "abc123",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !strings.Contains(whereClause, "v.hash = $2") {
		t.Fatalf("expected villain hash filter to use $2, got %s", whereClause)
	}
	if len(args) != 2 || args[1] != "abc123" {
		t.Fatalf("expected args [5 abc123], got %v", args)
	}
}
