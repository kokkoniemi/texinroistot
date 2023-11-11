package db

import (
	"fmt"
	"strings"
)

type storyRepo struct{}

const bulkCreateStorySQL = `
INSERT INTO stories(title, original_title, order_num, written_by, drawn_by, invented_by, version)
VALUES
	%s;
`

// BulkCreate implements StoryRepository.
func (*storyRepo) BulkCreate(stories []*Story, version Version) ([]*Story, error) {
	if len(stories) > 100 {
		return nil, fmt.Errorf("too many stories")
	}

	var values []string

	for _, s := range stories {
		values = append(values, fmt.Sprintf(
			"('%s', '%s', %v, %v, %v, %v, %v)",
			s.Title,
			s.GetOriginalTitleForDB(),
			s.GetOrderNumberForDB(),
			s.GetWriterIDForDB(),
			s.GetDrawerIDForDB(),
			s.GetInventorIDForDB(),
			version.ID))
	}

	createString := fmt.Sprintf(bulkCreateStorySQL, strings.Join(values, ","))
	res, err := Execute(createString)
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
	s.written_by,
	s.drawn_by,
	s.invented_by,
	a.id,
	a.first_name,
	a.last_name,
	a.is_writer,
	a.is_drawer,
	a.is_inventor
FROM stories AS s
INNER JOIN authors AS a ON s.written_by = a.id
	OR s.drawn_by = a.id
	OR s.invented_by = a.id 
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
		var s Story
		if err = rows.Scan( /*TODO: add scanning*/ ); err != nil {
			return nil, err
		}
		stories = append(stories, &s)
	}

	return stories, nil
}

func NewStoryRepository() StoryRepository {
	return &storyRepo{}
}
