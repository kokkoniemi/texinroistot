package importer

import "testing"

func TestParseAuthorName_KeepsSurnameParticlesWithLastName(t *testing.T) {
	cases := []struct {
		name, first, last string
	}{
		{"Pasquale Del Vecchio", "Pasquale", "Del Vecchio"},
		{"Victor de la Fuente", "Victor", "de la Fuente"},
		{"Giulio De Vita", "Giulio", "De Vita"},
		{"Roberto De Angelis", "Roberto", "De Angelis"},
		{"Raffaele Della Monica", "Raffaele", "Della Monica"},
		{"Giovanni Luigi Bonelli", "Giovanni Luigi", "Bonelli"},
		{"Leomacs", "Leomacs", ""},
		// a particle alone before the last name is not taken from the first name
		{"De Vita", "De", "Vita"},
	}
	for _, c := range cases {
		first, last := parseAuthorName(c.name)
		if first != c.first || last != c.last {
			t.Fatalf("parseAuthorName(%q) = (%q, %q), want (%q, %q)", c.name, first, last, c.first, c.last)
		}
	}
}
