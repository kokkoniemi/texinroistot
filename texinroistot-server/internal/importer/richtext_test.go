package importer

import (
	"testing"

	"github.com/xuri/excelize/v2"
)

func italicRun(text string) excelize.RichTextRun {
	return excelize.RichTextRun{Text: text, Font: &excelize.Font{Italic: true}}
}

func plainRun(text string) excelize.RichTextRun {
	return excelize.RichTextRun{Text: text}
}

func TestItalicMarkup(t *testing.T) {
	tests := []struct {
		name string
		runs []excelize.RichTextRun
		want string
	}{
		{
			name: "whitespace stays outside the tags",
			runs: []excelize.RichTextRun{
				plainRun("Aavevuorien kultaa varjelevien"),
				italicRun(" Ukkosen Poikien"),
				plainRun(" johtaja; kuului ennen Victorion sotureihin"),
			},
			want: "Aavevuorien kultaa varjelevien <i>Ukkosen Poikien</i> johtaja; kuului ennen Victorion sotureihin",
		},
		{
			name: "tags do not span the value separator",
			runs: []excelize.RichTextRun{
				plainRun("alku "),
				italicRun("yksi; kaksi"),
				plainRun(" loppu"),
			},
			want: "alku <i>yksi</i>; <i>kaksi</i> loppu",
		},
		{
			name: "consecutive italic runs are merged",
			runs: []excelize.RichTextRun{
				italicRun("Ukkosen "),
				italicRun("Poikien"),
				plainRun(" johtaja"),
			},
			want: "<i>Ukkosen Poikien</i> johtaja",
		},
		{
			name: "plain runs only",
			runs: []excelize.RichTextRun{plainRun("roisto; "), plainRun("johtaja")},
			want: "roisto; johtaja",
		},
		{
			name: "whitespace-only italic run gets no tags",
			runs: []excelize.RichTextRun{plainRun("a"), italicRun(" "), plainRun("b")},
			want: "a b",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := italicMarkup(tt.runs)
			if got != tt.want {
				t.Fatalf("italicMarkup() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestStripItalicMarkupRestoresPlainText(t *testing.T) {
	runs := []excelize.RichTextRun{
		plainRun("alku "),
		italicRun("yksi; kaksi"),
		plainRun(" loppu"),
	}
	want := "alku yksi; kaksi loppu"
	if got := stripItalicMarkup(italicMarkup(runs)); got != want {
		t.Fatalf("stripItalicMarkup() = %q, want %q", got, want)
	}
}

func TestApplyItalicMarkup_ReplacesOnlyItalicCells(t *testing.T) {
	file := excelize.NewFile()
	defer file.Close()
	sheet := "Sheet1"

	if err := file.SetCellRichText(sheet, "B2", []excelize.RichTextRun{
		plainRun("varjelevien"),
		italicRun(" Ukkosen Poikien"),
		plainRun(" johtaja"),
	}); err != nil {
		t.Fatalf("SetCellRichText failed: %v", err)
	}
	if err := file.SetCellStr(sheet, "B3", "tavallinen rooli"); err != nil {
		t.Fatalf("SetCellStr failed: %v", err)
	}

	i := &importer{columnIndexes: map[string]int{"roles": 1}}
	rows := [][]string{
		{"Nimi", "Rooli"},
		{"A", "varjelevien Ukkosen Poikien johtaja"},
		{"B", "tavallinen rooli"},
		{"C"},
	}

	if err := i.applyItalicMarkup(file, sheet, rows); err != nil {
		t.Fatalf("applyItalicMarkup failed: %v", err)
	}

	if got, want := rows[1][1], "varjelevien <i>Ukkosen Poikien</i> johtaja"; got != want {
		t.Errorf("row 2 = %q, want %q", got, want)
	}
	if got, want := rows[2][1], "tavallinen rooli"; got != want {
		t.Errorf("row 3 = %q, want %q", got, want)
	}
	if rows[0][1] != "Rooli" {
		t.Errorf("header row must not change, got %q", rows[0][1])
	}
}
