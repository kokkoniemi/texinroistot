package db

import (
	"fmt"
	"strings"
)

type authorRepo struct{}

const bulkCreateAuthorsSQL = `
INSERT INTO authors(first_name, last_name, is_writer, is_drawer, is_inventor, version)
VALUES
	%s;
`

// BulkCreate implements AuthorRepository.
func (*authorRepo) BulkCreate(authors []*Author, version Version) (int, error) {
	if len(authors) > 100 {
		return 0, fmt.Errorf("too many authors")
	}

	var values []string

	for _, a := range authors {
		values = append(values, fmt.Sprintf(
			"('%s', '%s', %v, %v, %v, %v)",
			a.FirstName,
			a.LastName,
			a.IsWriter,
			a.IsDrawer,
			a.IsInventor,
			version.ID))
	}

	createString := fmt.Sprintf(bulkCreateAuthorsSQL, strings.Join(values, ","))
	res, err := Execute(createString)
	if err != nil {
		return 0, err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}

	return int(rows), nil
}

const listAuthorsSQL = `
SELECT
	id,
	first_name,
	last_name,
	is_writer,
	is_drawer,
	is_inventor
FROM authors;
`

// List implements AuthorRepository.
func (*authorRepo) List() ([]*Author, error) {
	rows, err := Query(listAuthorsSQL)
	if err != nil {
		return nil, err
	}
	var authors []*Author

	for rows.Next() {
		var a Author
		if err = rows.Scan(&a.ID,
			&a.FirstName,
			&a.LastName,
			&a.IsWriter,
			&a.IsDrawer,
			&a.IsInventor,
		); err != nil {
			return nil, err
		}
		authors = append(authors, &a)
	}

	return authors, nil
}

const readAuthorSQL = `
SELECT
	id,
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
	var a Author
	for rows.Next() {
		if err = rows.Scan(
			&a.ID,
			&a.FirstName,
			&a.LastName,
			&a.IsWriter,
			&a.IsDrawer,
			&a.IsInventor,
		); err != nil {
			return nil, err
		}
	}
	return &a, nil
}

func NewAuthorRepository() AuthorRepository {
	return &authorRepo{}
}
