package db

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// pagingLetters are the letters accepted for letter paging (Finnish alphabet).
var pagingLetters = "abcdefghijklmnopqrstuvwxyzåäö"

// NormalizePagingLetter returns the lower-case paging letter, or "" when the
// value is not a single letter of the Finnish alphabet.
func NormalizePagingLetter(value string) string {
	letter := strings.ToLower(strings.TrimSpace(value))
	if utf8.RuneCountInString(letter) != 1 || !strings.Contains(pagingLetters, letter) {
		return ""
	}
	return letter
}

// letterPagingQuery holds a list query's FROM clause, filters and the primary
// sort key (as SQL) that letter paging is based on.
type letterPagingQuery struct {
	from        string
	whereClause string
	whereArgs   []interface{}
	sortKey     string
}

// LetterPosition locates the first row whose sort key starts with a letter
// (or would be, if none does): the page it is on and its 0-based index there.
type LetterPosition struct {
	Page  int
	Index int
}

// positionForLetter counts the rows sorting before letter, with the same key
// and collation as the list itself, and turns the count into a page position.
func positionForLetter(q letterPagingQuery, letter string, pageSize int) (*LetterPosition, error) {
	if pageSize <= 0 {
		return nil, fmt.Errorf("invalid page size")
	}
	sql := fmt.Sprintf(`
SELECT COUNT(*)
FROM %s
WHERE %s
AND (%s) < $%d;
`, q.from, q.whereClause, q.sortKey, len(q.whereArgs)+1)

	rows, err := Query(sql, append(append([]interface{}{}, q.whereArgs...), letter)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	before := 0
	if rows.Next() {
		if err = rows.Scan(&before); err != nil {
			return nil, err
		}
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return &LetterPosition{Page: before/pageSize + 1, Index: before % pageSize}, nil
}

// initialLetters returns the paging letters that some row's sort key starts with.
func initialLetters(q letterPagingQuery) ([]string, error) {
	sql := fmt.Sprintf(`
SELECT DISTINCT LEFT((%s), 1)
FROM %s
WHERE %s
AND (%s) IS NOT NULL;
`, q.sortKey, q.from, q.whereClause, q.sortKey)

	rows, err := Query(sql, q.whereArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	found := map[string]bool{}
	for rows.Next() {
		var initial string
		if err = rows.Scan(&initial); err != nil {
			return nil, err
		}
		if letter := NormalizePagingLetter(initial); letter != "" {
			found[letter] = true
		}
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	letters := []string{}
	for _, letter := range pagingLetters {
		if found[string(letter)] {
			letters = append(letters, string(letter))
		}
	}
	return letters, nil
}

// LetterPaging returns the initial letters available for an alphabetical
// villain sort and, when letter is set, where that letter starts. For the
// publication date sorts it returns no letters and no position.
func (v *villainRepo) LetterPaging(version *Version, params VillainListParams, letter string) ([]string, *LetterPosition, error) {
	sortKey := villainSortKeyExpr(params.Sort)
	if version == nil || version.ID == 0 || sortKey == "" {
		return []string{}, nil, nil
	}
	whereClause, whereArgs, err := buildVillainListWhere(version.ID, params)
	if err != nil {
		return nil, nil, err
	}
	return letterPaging(letterPagingQuery{"villains AS v", whereClause, whereArgs, sortKey}, letter, params.PageSize)
}

// LetterPaging returns the initial letters available for the alphabetical
// story sort and, when letter is set, where that letter starts. For the
// publication date sorts it returns no letters and no position.
func (s *storyRepo) LetterPaging(version *Version, params StoryListParams, letter string) ([]string, *LetterPosition, error) {
	sortKey := storySortKeyExpr(params.Sort, params.Publication)
	if version == nil || version.ID == 0 || sortKey == "" {
		return []string{}, nil, nil
	}
	whereClause, whereArgs, err := buildStoryListWhere(version.ID, params)
	if err != nil {
		return nil, nil, err
	}
	return letterPaging(letterPagingQuery{"stories AS s", whereClause, whereArgs, sortKey}, letter, params.PageSize)
}

func letterPaging(q letterPagingQuery, letter string, pageSize int) ([]string, *LetterPosition, error) {
	letters, err := initialLetters(q)
	if err != nil {
		return nil, nil, err
	}
	if letter == "" {
		return letters, nil, nil
	}
	position, err := positionForLetter(q, letter, pageSize)
	if err != nil {
		return nil, nil, err
	}
	return letters, position, nil
}
