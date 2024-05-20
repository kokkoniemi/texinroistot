package db

import (
	"fmt"
	"slices"
)

type storyRepo struct{}

// BulkCreate implements StoryRepository. TODO: add support for publications
func (s *storyRepo) BulkCreate(stories []*Story, version *Version) ([]*Story, error) {
	if len(stories) > MaxBulkCreateSize {
		return nil, fmt.Errorf("max number of %d stories exceeded", MaxBulkCreateSize)
	}

	var values [][]interface{}
	for _, s := range stories {
		values = append(values, []interface{}{
			s.Hash,
			s.GetOrderNumberForDB(),
			// TODO: FIX many2many issue with authors
			//s.GetWriterIDForDB(),
			//s.GetDrawerIDForDB(),
			//s.GetInventorIDForDB(),
			version.ID,
		})
	}

	rows, err := BulkInsertTxn(bulkInsertParams{
		Table: "stories",
		Columns: []string{
			"hash", "order_num", "written_by", "drawn_by", "invented_by", "version",
		},
		Values: values,
	})

	if err != nil {
		return nil, err
	}

	createdStories, err := s.list(version, true, int(rows))
	if err != nil {
		return nil, err
	}
	slices.Reverse(createdStories)

	return createdStories, nil
}

// List implements StoryRepository.
func (s *storyRepo) List(version *Version) ([]*Story, error) {
	return s.list(version, false, 0)
}

const listStoriesSQL = `
SELECT
	s.id,
	s.hash,
	s.order_num,
	w.id,
	w.hash,
	w.first_name,
	w.last_name,
	w.is_writer,
	w.is_drawer,
	w.is_inventor,
	d.id,
	d.hash,
	d.first_name,
	d.last_name,
	d.is_writer,
	d.is_drawer,
	d.is_inventor,
	i.id,
	i.hash,
	i.first_name,
	i.last_name,
	i.is_writer,
	i.is_drawer,
	i.is_inventor
FROM stories AS s
LEFT JOIN authors AS w ON s.written_by = w.id
LEFT JOIN authors AS d ON s.drawn_by = d.id
LEFT JOIN authors AS i ON s.invented_by = i.id
WHERE
	s.version = $1
%v;
`

func (*storyRepo) list(version *Version, descending bool, limit int) ([]*Story, error) {
	if version.ID == 0 || limit <= 0 {
		return nil, fmt.Errorf("invalid parameters")
	}

	var queryString string
	if descending {
		queryString = fmt.Sprintf(listStoriesSQL, "ORDER BY s.id DESC %v")
	}

	queryString = fmt.Sprintf(queryString, fmt.Sprintf("LIMIT %v", limit))

	rows, err := Query(queryString, version.ID)
	if err != nil {
		return nil, err
	}
	var stories []*Story

	for rows.Next() {
		var s Story
		var writerBp AuthorBlueprint
		var drawerBp AuthorBlueprint
		var inventorBp AuthorBlueprint
		if err = rows.Scan(
			&s.ID,
			&s.Hash,
			&s.OrderNumber,
			&writerBp.ID,
			&writerBp.Hash,
			&writerBp.FirstName,
			&writerBp.LastName,
			&writerBp.IsWriter,
			&writerBp.IsDrawer,
			&writerBp.IsInventor,
			&drawerBp.ID,
			&drawerBp.Hash,
			&drawerBp.FirstName,
			&drawerBp.LastName,
			&drawerBp.IsWriter,
			&drawerBp.IsDrawer,
			&drawerBp.IsInventor,
			&inventorBp.ID,
			&inventorBp.Hash,
			&inventorBp.FirstName,
			&inventorBp.LastName,
			&inventorBp.IsWriter,
			&inventorBp.IsDrawer,
			&inventorBp.IsInventor,
		); err != nil {
			return nil, err
		}
		// FIXME: fix many-2-many issue in db model
		// if writerBp.AuthorExists() {
		// 	s.WrittenBy = writerBp.ToAuthor()
		// }
		// if drawerBp.AuthorExists() {
		// 	s.DrawnBy = drawerBp.ToAuthor()
		// }
		// if inventorBp.AuthorExists() {
		// 	s.InventedBy = inventorBp.ToAuthor()
		// }
		stories = append(stories, &s)
	}

	return stories, nil
}

const listPublicationsSQL = `
SELECT
	p.id,
	p.hash,
	p.type,
	p.year,
	p.issue
FROM publications as p
WHERE
	p.version = $1
%v;
`

func (*storyRepo) listPublications(version *Version, descending bool, limit int) ([]*Publication, error) {
	if version.ID == 0 || limit <= 0 {
		return nil, fmt.Errorf("invalid parameters")
	}

	var queryString string
	if descending {
		queryString = fmt.Sprintf(listPublicationsSQL, "ORDER BY p.id DESC %v")
	}
	queryString = fmt.Sprintf(queryString, fmt.Sprintf("LIMIT %v", limit))

	rows, err := Query(queryString, version.ID)
	if err != nil {
		return nil, err
	}
	var publications []*Publication
	for rows.Next() {
		var p Publication
		if err = rows.Scan(
			&p.ID,
			&p.Hash,
			&p.Type,
			&p.Year,
			&p.Issue,
		); err != nil {
			return nil, err
		}
		publications = append(publications, &p)
	}

	return publications, nil
}

func (s *storyRepo) BulkCreatePublications(publications []*Publication, version *Version) ([]*Publication, error) {
	if len(publications) > MaxBulkCreateSize {
		return nil, fmt.Errorf("max number of %d publications exceeded", MaxBulkCreateSize)
	}

	var values [][]interface{}
	for _, p := range publications {
		values = append(values, []interface{}{
			p.Hash,
			p.Type,
			p.Year,
			p.Issue,
			version.ID,
		})
	}

	rows, err := BulkInsertTxn(bulkInsertParams{
		Table:   "publications",
		Columns: []string{"hash", "type", "year", "issue", "version"},
		Values:  values,
	})
	if err != nil {
		return nil, err
	}

	descending := true
	limit := int(rows)
	createdPublications, err := s.listPublications(version, descending, limit)
	if err != nil {
		return nil, err
	}
	slices.Reverse(createdPublications)

	return createdPublications, nil
}

func NewStoryRepository() StoryRepository {
	return &storyRepo{}
}
