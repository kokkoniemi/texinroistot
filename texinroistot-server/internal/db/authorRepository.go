package db

import (
	"fmt"
	"slices"
)

type authorRepo struct{}

// BulkCreate implements AuthorRepository.
func (a *authorRepo) BulkCreate(authors []*Author, version *Version) ([]*Author, error) {
	if len(authors) > 100 {
		return nil, fmt.Errorf("too many authors")
	}

	var values [][]interface{}
	for _, a := range authors {
		values = append(values, []interface{}{
			a.Hash, a.FirstName, a.LastName, a.IsWriter, a.IsDrawer, a.IsInventor, version.ID,
		})
	}

	rows, err := BulkInsertTxn(bulkInsertParams{
		Table: "authors",
		Columns: []string{
			"hash", "first_name", "last_name", "is_writer", "is_drawer", "is_inventor", "version",
		},
		Values: values,
	})

	if err != nil {
		return nil, err
	}

	createdAuthors, err := a.list(version, true, int(rows))
	if err != nil {
		return nil, err
	}
	slices.Reverse(createdAuthors)

	return createdAuthors, nil
}

// List implements AuthorRepository.
func (a *authorRepo) List(version *Version) ([]*Author, error) {
	return a.list(version, false, 0)
}

const listAuthorsSQL = `
SELECT
	id,
	hash,
	first_name,
	last_name,
	is_writer,
	is_drawer,
	is_inventor
FROM authors
WHERE
	version = $1
%v;
`

func (*authorRepo) list(version *Version, descending bool, limit int) ([]*Author, error) {
	var queryString string
	if descending {
		queryString = fmt.Sprintf(listAuthorsSQL, "ORDER BY id DESC %v")
	}
	if limit > 0 {
		queryString = fmt.Sprintf(queryString, fmt.Sprintf("LIMIT %v", limit))
	}
	rows, err := Query(queryString, version.ID)
	if err != nil {
		return nil, err
	}
	var authors []*Author

	for rows.Next() {
		var aBp AuthorBlueprint
		if err = rows.Scan(
			&aBp.ID,
			&aBp.Hash,
			&aBp.FirstName,
			&aBp.LastName,
			&aBp.IsWriter,
			&aBp.IsDrawer,
			&aBp.IsInventor,
		); err != nil {
			return nil, err
		}
		if aBp.AuthorExists() {
			authors = append(authors, aBp.ToAuthor())
		}
	}

	return authors, nil
}

const readAuthorSQL = `
SELECT
	id,
	hash,
	first_name,
	last_name,
	is_writer,
	is_drawer,
	is_inventor
FROM authors
WHERE id = $1;
`

// Read implements AuthorRepository.
func (*authorRepo) Read(authorID int) (*Author, error) {
	rows, err := Query(readAuthorSQL, authorID)
	if err != nil {
		return nil, err
	}
	var aBp AuthorBlueprint
	for rows.Next() {
		if err = rows.Scan(
			&aBp.ID,
			&aBp.Hash,
			&aBp.FirstName,
			&aBp.LastName,
			&aBp.IsWriter,
			&aBp.IsDrawer,
			&aBp.IsInventor,
		); err != nil {
			return nil, err
		}
	}
	if aBp.AuthorExists() {
		return aBp.ToAuthor(), nil
	}
	return nil, fmt.Errorf("corrupted author data")
}

func NewAuthorRepository() AuthorRepository {
	return &authorRepo{}
}
