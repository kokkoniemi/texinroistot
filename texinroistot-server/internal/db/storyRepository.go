package db

import (
	"database/sql"
	"fmt"
	"slices"
)

type storyRepo struct{}

// BulkCreate implements StoryRepository.
func (s *storyRepo) BulkCreate(stories []*Story, version *Version) ([]*Story, error) {
	if len(stories) > MaxBulkCreateSize {
		return nil, fmt.Errorf("max number of %d stories exceeded", MaxBulkCreateSize)
	}

	var storyValues [][]interface{}
	// create stories
	for _, s := range stories {
		storyValues = append(storyValues, []interface{}{
			s.GetOrderNumberForDB(),
			s.Hash,
			version.ID,
		})
	}
	numRows, err := BulkInsertTxn(bulkInsertParams{
		Table: "stories",
		Columns: []string{
			"order_num", "hash", "version",
		},
		Values: storyValues,
	})

	if err != nil {
		return nil, err
	}

	// update stories with their ids
	stories, err = s.setIDsFromDB(stories, numRows)
	if err != nil {
		return nil, err
	}

	// create story authors
	var storyAuthorValues [][]interface{}

	appendAuthorValue := func(story *Story, authors []*Author, atype string) {
		for _, a := range authors {
			storyAuthorValues = append(storyAuthorValues, []interface{}{
				story.ID,
				a.ID,
				atype,
			})
		}

	}

	for _, s := range stories {
		appendAuthorValue(s, s.WrittenBy, "writer")
		appendAuthorValue(s, s.DrawnBy, "drawer")
		appendAuthorValue(s, s.InventedBy, "inventor")
	}

	_, err = BulkInsertTxn(bulkInsertParams{
		Table:   "authors_in_stories",
		Columns: []string{"story", "author", "type"},
		Values:  storyAuthorValues,
	})
	if err != nil {
		return nil, err
	}

	// create storyPublications
	var storyPubValues [][]interface{}

	for _, s := range stories {
		for _, p := range s.Publications {
			// fmt.Println("----", p.ID, p.In.ID, s.ID)
			storyPubValues = append(storyPubValues, []interface{}{
				s.ID,
				p.In.ID,
				p.Title,
			})
		}
	}

	_, err = BulkInsertTxn(bulkInsertParams{
		Table:   "stories_in_publications",
		Columns: []string{"story", "publication", "title"},
		Values:  storyPubValues,
	})
	if err != nil {
		return nil, err
	}

	// return a list of created stories
	descending := true // to get latest rows, order by descending
	limit := numRows
	createdStories, err := s.list(version, descending, int(limit), 0)
	if err != nil {
		return nil, err
	}
	slices.Reverse(createdStories)

	return createdStories, nil
}

const setIDsSQL = `
SELECT
	s.id,
	s.hash
FROM stories AS s
ORDER BY s.id DESC
LIMIT %v;
`

func (s *storyRepo) setIDsFromDB(stories []*Story, savedRows int64) ([]*Story, error) {
	queryString := fmt.Sprintf(setIDsSQL, savedRows)
	rows, err := Query(queryString)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var row struct {
			ID   int
			Hash string
		}

		if err = rows.Scan(&row.ID, &row.Hash); err != nil {
			return nil, err
		}

		for idx := range stories {
			if stories[idx].Hash == row.Hash {
				stories[idx].ID = row.ID
				break
			}
			if idx == len(stories)-1 && stories[idx].Hash != row.Hash {
				return nil, fmt.Errorf("matching story not found from db")
			}
		}
	}

	return stories, nil
}

// List implements StoryRepository.
func (s *storyRepo) List(version *Version, limit int, offset int) ([]*Story, error) {
	return s.list(version, false, limit, offset)
}

const selectStoriesSQL = `
SELECT
	s.id,
	s.hash,
	s.order_num
FROM stories AS s
WHERE
	s.version = $1
%v;
`
const selectAuthorsInStoriesSQL = `
SELECT
	sa.story,
	sa.author,
	sa.type
FROM authors_in_stories AS sa
WHERE sa.story = ANY($1);
`
const selectAuthorsByIDsSQL = `
SELECT
	a.id,
	a.hash,
	a.first_name,
	a.last_name,
	a.is_writer,
	a.is_drawer,
	a.is_inventor
FROM authors AS a
WHERE a.id = ANY($1);
`
const selectStoryPublicationsSQL = `
SELECT
	p.id,
	p.hash,
	p.type,
	p.year,
	p.issue,
	sip.id,
	sip.title,
	sip.story
FROM publications as p
JOIN stories_in_publications AS sip ON p.id = sip.publication
WHERE p.id IN (
	SELECT sip.publication
	FROM stories_in_publications AS sip
	WHERE sip.story = ANY($1)
);
`

func (*storyRepo) selectStoryRows(version *Version, descending bool, limit int) ([]*Story, []int, error) {
	if version.ID == 0 || limit <= 0 {
		return nil, nil, fmt.Errorf("invalid parameters")
	}

	var queryString string
	if descending {
		queryString = fmt.Sprintf(selectStoriesSQL, "ORDER BY s.id DESC %v")
	} else {
		queryString = fmt.Sprintf(selectStoriesSQL, "ORDER BY s.id ASC %v")
	}

	queryString = fmt.Sprintf(queryString, fmt.Sprintf("LIMIT %v", limit))

	rows, err := Query(queryString, version.ID)
	if err != nil {
		return nil, nil, err
	}
	var stories []*Story
	var storyIDs []int

	for rows.Next() {
		var s Story
		if err = rows.Scan(
			&s.ID,
			&s.Hash,
			&s.OrderNumber,
		); err != nil {
			return nil, nil, err
		}
		stories = append(stories, &s)
		storyIDs = append(storyIDs, s.ID)
	}

	return stories, storyIDs, nil
}

type ainfo struct {
	Story  int
	Author int
	Type   string
}

func (*storyRepo) selectStoryAuthorRows(storyIDs []int) (map[int][]*ainfo, []*Author, error) {
	rows, err := Query(selectAuthorsInStoriesSQL, ArrayParam(storyIDs))
	if err != nil {
		return nil, nil, err
	}

	var authorInfos = make(map[int][]*ainfo)
	var authorIDs []int

	for rows.Next() {
		var info ainfo
		if err = rows.Scan(
			&info.Story,
			&info.Author,
			&info.Type,
		); err != nil {
			return nil, nil, err
		}

		authorInfos[info.Story] = append(authorInfos[info.Story], &info)
		authorIDs = append(authorIDs, info.Author)
	}

	rows, err = Query(selectAuthorsByIDsSQL, ArrayParam(authorIDs))
	if err != nil {
		return nil, nil, err
	}
	var authors []*Author

	for rows.Next() {
		var a Author
		if err = rows.Scan(
			&a.ID,
			&a.Hash,
			&a.FirstName,
			&a.LastName,
			&a.IsWriter,
			&a.IsDrawer,
			&a.IsInventor,
		); err != nil {
			return nil, nil, err
		}
		authors = append(authors, &a)
	}
	return authorInfos, authors, nil
}

func (*storyRepo) selectStoryPublicationRows(storyIDs []int) (map[int][]*StoryPublication, error) {
	rows, err := Query(selectStoryPublicationsSQL, ArrayParam(storyIDs))
	if err != nil {
		return nil, err
	}

	var storyPublications = make(map[int][]*StoryPublication)

	for rows.Next() {
		var p Publication
		var sp StoryPublication
		var storyID sql.NullInt64

		if err = rows.Scan(
			&p.ID,
			&p.Hash,
			&p.Type,
			&p.Year,
			&p.Issue,
			&sp.ID,
			&sp.Title,
			&storyID,
		); err != nil {
			return nil, err
		}

		sp.In = &p

		if storyID.Int64 == 0 {
			return nil, fmt.Errorf("something went wrong")
		}
		k := int(storyID.Int64)
		storyPublications[k] = append(storyPublications[k], &sp)
	}
	return storyPublications, nil
}

// here I have experimented combining the list both with and without JOINS. TODO: do offset logic
func (s *storyRepo) list(version *Version, descending bool, limit int, offset int) ([]*Story, error) {
	stories, storyIDs, err := s.selectStoryRows(version, descending, limit)
	if err != nil {
		return nil, err
	}

	// TODO: use channels to not block storyPublication request
	authorInfos, authors, err := s.selectStoryAuthorRows(storyIDs)
	if err != nil {
		return nil, err
	}

	storyPublications, err := s.selectStoryPublicationRows(storyIDs)
	if err != nil {
		return nil, err
	}

	getAuthorsByInfo := func(infos []*ainfo) ([]*Author, []*Author, []*Author) {
		var writers []*Author
		var drawers []*Author
		var inventors []*Author

		for _, info := range infos {
			for _, a := range authors {
				if info.Author == a.ID {
					if info.Type == "writer" {
						writers = append(writers, a)
					} else if info.Type == "drawer" {
						drawers = append(drawers, a)
					} else if info.Type == "inventor" {
						inventors = append(inventors, a)
					}
					break
				}
			}
		}

		return writers, drawers, inventors
	}

	for idx := range stories {
		infos := authorInfos[stories[idx].ID]
		stories[idx].WrittenBy, stories[idx].DrawnBy, stories[idx].InventedBy = getAuthorsByInfo(infos)

		stories[idx].Publications = storyPublications[stories[idx].ID]
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
