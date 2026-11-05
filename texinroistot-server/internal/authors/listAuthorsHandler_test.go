package authors

import (
	"sort"
	"testing"

	"github.com/kokkoniemi/texinroistot/internal/db"
)

func sortedNames(sortValue string, authors []*db.Author) []string {
	sort.SliceStable(authors, func(i, j int) bool {
		return compareAuthors(sortValue, authors[i], authors[j])
	})
	names := make([]string, len(authors))
	for i, author := range authors {
		names[i] = author.FirstName + " " + author.LastName
	}
	return names
}

func testAuthors() []*db.Author {
	return []*db.Author{
		{Hash: "1", FirstName: "Leomacs"},
		{Hash: "2", FirstName: "Andrea", LastName: "Accardi"},
		{Hash: "3", FirstName: "Magnus"},
		{Hash: "4", FirstName: "Mario", LastName: "Majo"},
		{Hash: "5", FirstName: "Giovanni", LastName: "Ticci"},
	}
}

func TestCompareAuthors_SingleNameSortsAsLastName(t *testing.T) {
	got := sortedNames("last_name", testAuthors())
	want := []string{"Andrea Accardi", "Leomacs ", "Magnus ", "Mario Majo", "Giovanni Ticci"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("last_name order = %q, want %q", got, want)
		}
	}
}

func TestCompareAuthors_SingleNameSortsAsFirstName(t *testing.T) {
	got := sortedNames("first_name", testAuthors())
	want := []string{"Andrea Accardi", "Giovanni Ticci", "Leomacs ", "Magnus ", "Mario Majo"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("first_name order = %q, want %q", got, want)
		}
	}
}
