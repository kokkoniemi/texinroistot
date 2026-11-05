package db

import (
	"strings"
	"testing"
)

func TestNormalizePagingLetter(t *testing.T) {
	tests := map[string]string{
		"k":   "k",
		"K":   "k",
		" Ä ": "ä",
		"ö":   "ö",
		"Å":   "å",
		"":    "",
		"ab":  "",
		"1":   "",
		"é":   "",
		"%":   "",
	}
	for input, want := range tests {
		if got := NormalizePagingLetter(input); got != want {
			t.Errorf("NormalizePagingLetter(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestVillainSortKeyExpr_OnlyAlphabeticalSorts(t *testing.T) {
	for _, sort := range []string{"fi_pub_date", "it_pub_date", "default"} {
		if key := villainSortKeyExpr(sort); key != "" {
			t.Errorf("expected no letter paging key for %q, got %s", sort, key)
		}
	}
	for sort, column := range map[string]string{
		"first_name": "v.first_names",
		"last_name":  "v.last_name",
		"rank":       "v.ranks",
		"nickname":   "vis.nicknames",
		"other_name": "vis.other_names",
		"code_name":  "vis.code_names",
	} {
		key := villainSortKeyExpr(sort)
		if !strings.Contains(key, column) {
			t.Errorf("sort key for %q should use %s, got %s", sort, column, key)
		}
		if !strings.Contains(buildVillainSortClause(sort), key) {
			t.Errorf("sort clause for %q must start from its letter paging key", sort)
		}
	}
}

func TestStorySortKeyExpr_OnlyAlphaSort(t *testing.T) {
	for _, sort := range []string{"fi_pub_date", "it_pub_date"} {
		if key := storySortKeyExpr(sort, "all"); key != "" {
			t.Errorf("expected no letter paging key for %q, got %s", sort, key)
		}
	}
	key := storySortKeyExpr("alpha", "texone")
	if !strings.Contains(key, "'italia_texone'") {
		t.Errorf("alpha key for texone should use texone titles, got %s", key)
	}
	if !strings.HasPrefix(buildSortClause("alpha", "texone"), key) {
		t.Errorf("alpha sort clause must start with its letter paging key")
	}
}

// The default villain sort groups villains by the first name field they have,
// in this order: first name, last name, other name, nickname, rank, code name.
func TestDefaultVillainSortClause_GroupOrder(t *testing.T) {
	clause := buildVillainSortClause("default")
	columns := []string{
		"v.first_names",
		"v.last_name",
		"vis.other_names",
		"vis.nicknames",
		"v.ranks",
		"vis.code_names",
	}
	previous := -1
	for i, column := range columns {
		position := strings.Index(clause, column)
		if position < 0 {
			t.Fatalf("default sort clause does not use %s: %s", column, clause)
		}
		if position < previous {
			t.Fatalf("%s must come after %s in the default sort", column, columns[i-1])
		}
		previous = position
	}
	if !strings.HasPrefix(clause, "CASE WHEN") || !strings.Contains(clause, "ELSE 7 END") {
		t.Fatalf("default sort must start with the 7 groups: %s", clause)
	}
	if buildVillainSortClause("first_name") == clause {
		t.Fatalf("first_name sort must stay a plain first name sort")
	}
}
