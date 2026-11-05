package importer

import (
	"strings"

	"github.com/xuri/excelize/v2"
)

// Italic parts of Rooli and Kohtalo cells are kept as <i>…</i> markup so the
// UI can show them in italics. Tags never span the ";" value separator, so
// every value produced by TrimmedSplit is balanced on its own.
const (
	italicOpen  = "<i>"
	italicClose = "</i>"
)

// italicMarkupColumns lists the column keys whose italic formatting is kept.
var italicMarkupColumns = []string{"roles", "destiny"}

// italicMarkup renders rich text runs as plain text with italic runs wrapped
// in <i>…</i>. Whitespace and ";" separators are kept outside the tags.
func italicMarkup(runs []excelize.RichTextRun) string {
	var out strings.Builder
	var italic strings.Builder

	flushItalic := func() {
		text := italic.String()
		italic.Reset()
		for index, part := range strings.Split(text, ";") {
			if index > 0 {
				out.WriteString(";")
			}
			trimmed := strings.TrimSpace(part)
			if trimmed == "" {
				out.WriteString(part)
				continue
			}
			start := strings.Index(part, trimmed)
			out.WriteString(part[:start])
			out.WriteString(italicOpen)
			out.WriteString(trimmed)
			out.WriteString(italicClose)
			out.WriteString(part[start+len(trimmed):])
		}
	}

	for _, run := range runs {
		if run.Font != nil && run.Font.Italic {
			italic.WriteString(run.Text)
			continue
		}
		flushItalic()
		out.WriteString(run.Text)
	}
	flushItalic()

	return out.String()
}

// stripItalicMarkup removes the <i>…</i> tags added by italicMarkup.
func stripItalicMarkup(value string) string {
	return strings.NewReplacer(italicOpen, "", italicClose, "").Replace(value)
}

func stripItalicMarkupAll(values []string) []string {
	stripped := make([]string, len(values))
	for index, value := range values {
		stripped[index] = stripItalicMarkup(value)
	}
	return stripped
}

func hasItalicRun(runs []excelize.RichTextRun) bool {
	for _, run := range runs {
		if run.Font != nil && run.Font.Italic && strings.TrimSpace(run.Text) != "" {
			return true
		}
	}
	return false
}

// applyItalicMarkup replaces the plain values of italicMarkupColumns in rows
// (header row first, as returned by GetRows) with italic markup where the
// cell contains italic rich text.
func (i *importer) applyItalicMarkup(file *excelize.File, sheet string, rows [][]string) error {
	for _, key := range italicMarkupColumns {
		columnIndex, ok := i.columnIndexes[key]
		if !ok {
			continue
		}
		for rowIndex := 1; rowIndex < len(rows); rowIndex++ {
			cells := rows[rowIndex]
			if columnIndex >= len(cells) || strings.TrimSpace(cells[columnIndex]) == "" {
				continue
			}

			cellName, err := excelize.CoordinatesToCellName(columnIndex+1, rowIndex+1)
			if err != nil {
				return err
			}
			runs, err := file.GetCellRichText(sheet, cellName)
			if err != nil {
				return err
			}
			if !hasItalicRun(runs) {
				continue
			}

			markup := italicMarkup(runs)
			// Only replace when the text itself is unchanged.
			if stripItalicMarkup(markup) != cells[columnIndex] {
				continue
			}
			cells[columnIndex] = markup
		}
	}
	return nil
}
