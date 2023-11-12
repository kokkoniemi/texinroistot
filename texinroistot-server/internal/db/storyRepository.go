package db

import (
	"fmt"

	"github.com/lib/pq"
)

type storyRepo struct{}

// BulkCreate implements StoryRepository.
func (*storyRepo) BulkCreate(stories []*Story, version Version) ([]*Story, error) {
	if len(stories) > 100 {
		return nil, fmt.Errorf("too many stories")
	}

	txn, err := StartTransaction()
	if err != nil {
		return nil, err
	}

	stmt, err := txn.Prepare(pq.CopyIn(
		"stories",
		"title",
		"original_title",
		"order_num",
		"written_by",
		"drawn_by",
		"invented_by",
		"version",
	))
	if err != nil {
		return nil, err
	}

	for _, s := range stories {
		_, err = stmt.Exec(
			s.Title,
			s.GetOriginalTitleForDB(),
			s.GetOrderNumberForDB(),
			s.GetWriterIDForDB(),
			s.GetDrawerIDForDB(),
			s.GetInventorIDForDB(),
			version.ID)
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

	if int(rows) != len(stories) {
		return nil, fmt.Errorf("something went wrong when creating stories")
	}

	return stories, nil
}

// List implements StoryRepository.
func (s *storyRepo) List(version Version) ([]*Story, error) {
	return s.list(version, false, 0)
}

const listStoriesSQL = `
SELECT
	s.id,
	s.title,
	s.original_title,
	s.order_num,
	w.id,
	w.first_name,
	w.last_name,
	w.is_writer,
	w.is_drawer,
	w.is_inventor,
	d.id,
	d.first_name,
	d.last_name,
	d.is_writer,
	d.is_drawer,
	d.is_inventor,
	i.id,
	i.first_name,
	i.last_name,
	i.is_writer,
	i.is_drawer,
	i.is_inventor,
FROM stories AS s
INNER JOIN authors AS w ON s.written_by = a.id
INNER JOIN authors AS d ON s.drawn_by = d.id
INNER JOIN authors AS i ON s.invented_by = i.id
WHERE
	version = $1
%v;
`

func (*storyRepo) list(version Version, descending bool, limit int) ([]*Story, error) {
	var queryString string
	if descending {
		queryString = fmt.Sprintf(listStoriesSQL, "ORDER BY id DESC %v")
	}
	if limit > 0 {
		queryString = fmt.Sprintf(queryString, fmt.Sprintf("LIMIT %v", limit))
	}
	rows, err := Query(queryString, version.ID)
	if err != nil {
		return nil, err
	}
	var stories []*Story

	for rows.Next() {
		var s *Story
		var writer *Author
		var drawer *Author
		var inventor *Author
		if err = rows.Scan(
			&s.ID,
			&s.Title,
			&s.OriginalTitle,
			&s.OrderNumber,
			&writer.ID,
			&writer.FirstName,
			&writer.LastName,
			&writer.IsWriter,
			&writer.IsDrawer,
			&writer.IsInventor,
			&drawer.ID,
			&drawer.FirstName,
			&drawer.LastName,
			&drawer.IsWriter,
			&drawer.IsDrawer,
			&drawer.IsInventor,
			&inventor.ID,
			&inventor.FirstName,
			&inventor.LastName,
			&inventor.IsWriter,
			&inventor.IsDrawer,
			&inventor.IsInventor,
		); err != nil {
			return nil, err
		}
		if writer.Exists() {
			s.WrittenBy = writer
		}
		if drawer.Exists() {
			s.DrawnBy = drawer
		}
		if inventor.Exists() {
			s.InventedBy = inventor
		}
		stories = append(stories, s)
	}

	return stories, nil
}

func NewStoryRepository() StoryRepository {
	return &storyRepo{}
}
