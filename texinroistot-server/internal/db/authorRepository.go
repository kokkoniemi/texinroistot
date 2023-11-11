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
func (a *authorRepo) BulkCreate(authors []*Author, version Version) ([]*Author, error) {
	if len(authors) > 100 {
		return nil, fmt.Errorf("too many authors")
	}

	var values []string

	for _, author := range authors {
		values = append(values, fmt.Sprintf(
			"('%s', '%s', %v, %v, %v, %v)",
			author.FirstName,
			author.LastName,
			author.IsWriter,
			author.IsDrawer,
			author.IsInventor,
			version.ID))
	}

	createString := fmt.Sprintf(bulkCreateAuthorsSQL, strings.Join(values, ","))
	res, err := Execute(createString)
	if err != nil {
		return nil, err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}

	if int(rows) != len(authors) {
		return nil, fmt.Errorf("something went wrong creating authors")
	}

	return a.list(version, true, int(rows))
}

// List implements AuthorRepository.
func (a *authorRepo) List(version Version) ([]*Author, error) {
	return a.list(version, false, 0)
}

const listAuthorsSQL = `
SELECT
	id,
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

func (*authorRepo) list(version Version, descending bool, limit int) ([]*Author, error) {
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
