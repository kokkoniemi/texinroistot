package db

import (
	"database/sql/driver"
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
	if !strings.Contains(whereClause, "AND "+publicationYearSQL+" = $3") {
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
		Sort:        "alpha",
		Year:        1967,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if strings.Contains(whereClause, "p.type::text = ANY(") {
		t.Fatalf("did not expect publication-type clause for all publications, got %s", whereClause)
	}
	if !strings.Contains(whereClause, "AND "+publicationYearSQL+" = $2") {
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

// With the Finnish date sort, a year on all publications only matches Finnish
// publications; with the Italian date sort, only Italian ones.
func TestBuildStoryListWhere_YearOnAllFollowsDateSort(t *testing.T) {
	for sort, wantTypes := range map[string][]string{
		"fi_pub_date": finnishPublicationTypes,
		"":            finnishPublicationTypes,
		"it_pub_date": italianPublicationTypes,
	} {
		whereClause, args, err := buildStoryListWhere(9, StoryListParams{
			Publication: "all",
			Sort:        sort,
			Year:        1999,
		})
		if err != nil {
			t.Fatalf("%q: expected no error, got %v", sort, err)
		}
		if !strings.Contains(whereClause, "AND "+publicationYearSQL+" = $2") ||
			!strings.Contains(whereClause, "AND p.type::text = ANY($3)") {
			t.Fatalf("%q: expected year $2 and type $3 clauses, got %s", sort, whereClause)
		}
		if len(args) != 3 || args[1] != 1999 {
			t.Fatalf("%q: expected args [version 1999 types], got %v", sort, args)
		}
		types, ok := args[2].(driver.Valuer)
		if !ok {
			t.Fatalf("%q: expected publication types as an array param, got %T", sort, args[2])
		}
		gotValue, _ := types.Value()
		wantValue, _ := ArrayParam(wantTypes).Value()
		if gotValue != wantValue {
			t.Fatalf("%q: expected types %v, got %v", sort, wantValue, gotValue)
		}
	}
}

// The Finnish date sort with a year orders by that year's base series issue,
// then that year's Finnish special publications; without a year it is unchanged.
func TestBuildStoryOrderClause_YearOrdersByThatYear(t *testing.T) {
	clause := buildStoryOrderClause(StoryListParams{Sort: "fi_pub_date", Publication: "all", Year: 1999})
	perus := strings.Index(clause, "p.type = 'perus' AND p.year = 1999")
	special := strings.Index(clause, "'suur', 'maxi', 'muu_erikois'")
	if perus < 0 || special < 0 || special < perus {
		t.Fatalf("expected 1999 base issues before 1999 specials, got %s", clause)
	}

	withoutYear := StoryListParams{Sort: "fi_pub_date", Publication: "all"}
	if buildStoryOrderClause(withoutYear) != buildSortClause("fi_pub_date", "all") {
		t.Fatalf("without a year the order must not change")
	}
	alpha := StoryListParams{Sort: "alpha", Publication: "all", Year: 1999}
	if buildStoryOrderClause(alpha) != buildSortClause("alpha", "all") {
		t.Fatalf("the alphabetical sort must not change with a year")
	}
}

// The Italian date sort orders by the year of the first Italian publication
// (any Italian type) and then by the story order number; with a year it orders
// by the story order number only.
func TestItalianDateSort_YearThenOrderNumber(t *testing.T) {
	clause := buildSortClause("it_pub_date", "all")
	year := strings.Index(clause, publicationYearSQL)
	orderNum := strings.Index(clause, "s.order_num ASC")
	if year < 0 || orderNum < 0 || orderNum < year {
		t.Fatalf("expected Italian year before order number, got %s", clause)
	}
	for _, italianType := range italianPublicationTypes {
		if !strings.Contains(clause, "'"+italianType+"'") {
			t.Fatalf("expected %s in the Italian date sort, got %s", italianType, clause)
		}
	}

	withYear := buildStoryOrderClause(StoryListParams{Sort: "it_pub_date", Publication: "all", Year: 1999})
	if !strings.HasPrefix(withYear, "s.order_num ASC") {
		t.Fatalf("expected order number order with a year, got %s", withYear)
	}
}
