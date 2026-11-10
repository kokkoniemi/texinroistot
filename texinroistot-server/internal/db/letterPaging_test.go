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
		listKey := villainListSortKeyExpr(sort)
		if !strings.HasPrefix(buildVillainSortClause(sort), listKey) {
			t.Errorf("sort clause for %q must start from its letter paging key %s", sort, listKey)
		}
	}
}

func TestBuildVillainSortClause_RankOrdersVillainsWithoutLastName(t *testing.T) {
	clause := buildVillainSortClause("rank")
	// last name first, then the other names of villains without one, so
	// "Sheriffi Clem" and "Sheriffi Harry" are not left in insertion order
	lastName := villainSortKeyExpr("last_name")
	fallback := "COALESCE(" + villainSortKeyExpr("first_name")
	if !strings.Contains(clause, lastName) || !strings.Contains(clause, fallback) {
		t.Fatalf("rank sort should order by last name, then by the other names: %s", clause)
	}
	if strings.Index(clause, lastName) > strings.Index(clause, fallback) {
		t.Fatalf("last name must come before the other names: %s", clause)
	}
	// a bare rank is ordered by the story it first appears in
	if strings.Index(clause, firstAppearanceSQL) < strings.Index(clause, fallback) {
		t.Fatalf("first appearance must settle what the names do not: %s", clause)
	}
}

func TestDefaultVillainSortClause_SameNameByFirstAppearance(t *testing.T) {
	clause := defaultVillainSortClause()
	if !strings.Contains(clause, firstAppearanceSQL) ||
		strings.Index(clause, firstAppearanceSQL) > strings.Index(clause, "v.id ASC") {
		t.Fatalf("villains with the same name should be ordered by first appearance: %s", clause)
	}
}

func TestSearchMatchesRank_NoSearchKeepsDefaultOrder(t *testing.T) {
	// without a search term no query is made and the default order stays
	matches, err := searchMatchesRank("v.version = $1", []interface{}{1}, "  ")
	if err != nil || matches {
		t.Fatalf("expected no rank order without a search, got %v, %v", matches, err)
	}
}

func TestVillainListFrom_PerStoryNameSortsListEachName(t *testing.T) {
	for sort, column := range map[string]string{
		"nickname":   "vis.nicknames",
		"other_name": "vis.other_names",
		"code_name":  "vis.code_names",
	} {
		from := villainListFrom(sort)
		// one row per villain and distinct name, so "Ice" and "Blue" are
		// both listed (under I and under B)
		if !strings.Contains(from, "unnest("+column+")") || !strings.Contains(from, "DISTINCT ON") {
			t.Errorf("FROM for %q should list each distinct name, got %s", sort, from)
		}
		if villainListSortKeyExpr(sort) != "sn.key" {
			t.Errorf("sort %q should be keyed by the listed name", sort)
		}
	}
	for _, sort := range []string{"default", "first_name", "last_name", "rank", "fi_pub_date"} {
		if from := villainListFrom(sort); from != "villains AS v" {
			t.Errorf("sort %q should list each villain once, got %s", sort, from)
		}
	}
}

func TestStorySortKeyExpr_OnlyAlphaSort(t *testing.T) {
	for _, sort := range []string{"fi_pub_date", "it_pub_date"} {
		if key := storySortKeyExpr(sort, "all"); key != "" {
			t.Errorf("expected no letter paging key for %q, got %s", sort, key)
		}
	}
	key := storySortKeyExpr("alpha_it", "texone")
	if !strings.Contains(key, "'italia_texone'") {
		t.Errorf("Italian alpha key for texone should use texone titles, got %s", key)
	}
	for _, sort := range []string{"alpha", "alpha_it"} {
		if !strings.HasPrefix(buildSortClause(sort, "texone"), storySortKeyExpr(sort, "texone")) {
			t.Errorf("%s sort clause must start with its letter paging key", sort)
		}
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

// Year paging uses the same key as the date sort it jumps in, with that key's
// year span, and is off for alphabetical sorts and single-year lists.
func TestStoryDateKey(t *testing.T) {
	cases := []struct {
		params StoryListParams
		span   int
		order  string
	}{
		{StoryListParams{Publication: "all", Sort: "fi_pub_date"}, finnishDateKeySpan, buildSortClause("fi_pub_date", "all")},
		{StoryListParams{Publication: "texone", Sort: "it_pub_date"}, italianYearKeySpan, buildSortClause("it_pub_date", "texone")},
		{StoryListParams{Publication: "suur", Sort: "fi_pub_date"}, finnishDateKeySpan, finnishSpecialOrderClause(StoryListParams{Publication: "suur", Sort: "fi_pub_date"})},
	}
	for _, c := range cases {
		key, span := storyDateKey(c.params)
		if key == "" || span != c.span {
			t.Fatalf("%+v: expected a date key with span %d, got %q span %d", c.params, c.span, key, span)
		}
		if !strings.HasPrefix(c.order, key) {
			t.Fatalf("%+v: the sort must start with the year paging key", c.params)
		}
	}
	for _, params := range []StoryListParams{
		{Publication: "all", Sort: "alpha"},
		{Publication: "texone", Sort: "alpha_it"},
		{Publication: "all", Sort: "fi_pub_date", Year: 1999},
	} {
		if key, _ := storyDateKey(params); key != "" {
			t.Fatalf("%+v: expected no year paging", params)
		}
	}
}
