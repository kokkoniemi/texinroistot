package db

import (
	"fmt"

	"github.com/lib/pq"
)

type authorRepo struct{}

// BulkCreate implements AuthorRepository.
func (a *authorRepo) BulkCreate(authors []*Author, version Version) ([]*Author, error) {
	if len(authors) > 100 {
		return nil, fmt.Errorf("too many authors")
	}

	txn, err := StartTransaction()
	if err != nil {
		return nil, err
	}

	stmt, err := txn.Prepare(pq.CopyIn(
		"authors",
		"first_name",
		"last_name",
		"is_writer",
		"is_drawer",
		"is_inventor",
		"version",
	))
	if err != nil {
		return nil, err
	}

	for _, a := range authors {
		_, err = stmt.Exec(a.FirstName, a.LastName, a.IsWriter, a.IsDrawer, a.IsInventor, version.ID)
		if err != nil {
			return nil, err
		}
	}

	res, err := stmt.Exec()
	if err != nil {
		return nil, err
	}

	err = stmt.Close()
	if err != nil {
		return nil, err
	}

	err = txn.Commit()
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
