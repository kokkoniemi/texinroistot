package db

import (
	"database/sql"
	"fmt"
	"slices"
	"strings"
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
			var details interface{}
			if strings.TrimSpace(a.Details) != "" {
				details = strings.TrimSpace(a.Details)
			}
			storyAuthorValues = append(storyAuthorValues, []interface{}{
				story.ID,
				a.ID,
				atype,
				details,
			})
		}

	}

	for _, s := range stories {
		appendAuthorValue(s, s.WrittenBy, "writer")
		appendAuthorValue(s, s.DrawnBy, "drawer")
		appendAuthorValue(s, s.TranslatedBy, "translator")
	}

	_, err = BulkInsertTxn(bulkInsertParams{
		Table:   "authors_in_stories",
		Columns: []string{"story", "author", "type", "details"},
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

var publicationTypesByFilter = map[string][]string{
	"all":      {},
	"perus_fi": {"perus"},
	"perus_it": {"italia_perus"},
	"suur":     {"suur"},
	"maxi":     {"maxi"},
	"kirjasto": {"kirjasto"},
	"kronikka": {"kronikka"},
	"special":  {"muu_erikois", "italia_erikois"},
}

const authorExistsByHashSQL = `
SELECT 1
FROM authors AS a
WHERE
	a.version = $1
	AND a.hash = $2
LIMIT 1;
`

const selectStoriesByAuthorHashSQL = `
SELECT DISTINCT
	s.id,
	s.hash,
	s.order_num
FROM stories AS s
JOIN authors_in_stories AS sa ON sa.story = s.id
JOIN authors AS a ON a.id = sa.author
WHERE
	s.version = $1
	AND a.hash = $2
%v
ORDER BY s.order_num ASC NULLS LAST, s.id ASC;
`

func normalizeAuthorStoryType(raw string) (string, error) {
	authorType := strings.TrimSpace(strings.ToLower(raw))
	if authorType == "" {
		return "", nil
	}
	switch authorType {
	case "writer", "drawer", "translator":
		return authorType, nil
	default:
		return "", fmt.Errorf("invalid author type")
	}
}

// ListByAuthorHash implements StoryRepository.
func (s *storyRepo) ListByAuthorHash(version *Version, authorHash string, authorType string) ([]*Story, bool, error) {
	if version.ID == 0 {
		return nil, false, fmt.Errorf("invalid version")
	}

	authorHash = strings.TrimSpace(authorHash)
	if authorHash == "" {
		return nil, false, fmt.Errorf("author hash is required")
	}

	normalizedType, err := normalizeAuthorStoryType(authorType)
	if err != nil {
		return nil, false, err
	}

	existsRows, err := Query(authorExistsByHashSQL, version.ID, authorHash)
	if err != nil {
		return nil, false, err
	}
	defer existsRows.Close()

	if !existsRows.Next() {
		return []*Story{}, false, nil
	}

	querySQL := fmt.Sprintf(selectStoriesByAuthorHashSQL, "")
	args := []interface{}{version.ID, authorHash}
	if normalizedType != "" {
		querySQL = fmt.Sprintf(selectStoriesByAuthorHashSQL, "AND sa.type = $3")
		args = append(args, normalizedType)
	}

	rows, err := Query(querySQL, args...)
	if err != nil {
		return nil, true, err
	}
	defer rows.Close()

	var stories []*Story
	var storyIDs []int

	for rows.Next() {
		var story Story
		if err = rows.Scan(&story.ID, &story.Hash, &story.OrderNumber); err != nil {
			return nil, true, err
		}
		stories = append(stories, &story)
		storyIDs = append(storyIDs, story.ID)
	}

	if err = rows.Err(); err != nil {
		return nil, true, err
	}

	if len(stories) == 0 {
		return []*Story{}, true, nil
	}

	if err = s.hydrateStories(stories, storyIDs); err != nil {
		return nil, true, err
	}

	return stories, true, nil
}

// ListFiltered implements StoryRepository.
func (s *storyRepo) ListFiltered(version *Version, params StoryListParams) ([]*Story, int, error) {
	stories, storyIDs, total, err := s.selectStoryRowsFiltered(version, params)
	if err != nil {
		return nil, 0, err
	}
	if len(stories) == 0 {
		return stories, total, nil
	}

	if err = s.hydrateStories(stories, storyIDs); err != nil {
		return nil, 0, err
	}
	return stories, total, nil
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

func mapPublicationFilterToTypes(filter string) ([]string, error) {
	types, found := publicationTypesByFilter[filter]
	if !found {
		return nil, fmt.Errorf("invalid publication filter")
	}
	return types, nil
}

func buildSortClause(sort string, publication string) string {
	publicationSortExpr := func(pubType string) string {
		return fmt.Sprintf(`(
	SELECT MIN(
		(COALESCE(p.year, 0) * 1000) + COALESCE(NULLIF(regexp_replace(p.issue, '[^0-9]', '', 'g'), '')::int, 0)
	)
	FROM stories_in_publications AS sip
	JOIN publications AS p ON p.id = sip.publication
	WHERE sip.story = s.id AND p.type = '%s'
)`, pubType)
	}
	alphaTitleExpr := func(pubType string) string {
		return fmt.Sprintf(`(
	SELECT MIN(
		NULLIF(
			regexp_replace(
				lower(BTRIM(sip.title)),
				'[[:punct:][:space:]]+',
				'',
				'g'
			),
			''
		)
	)
	FROM stories_in_publications AS sip
	JOIN publications AS p ON p.id = sip.publication
	WHERE sip.story = s.id
	AND p.type = '%s'
)`, pubType)
	}

	switch sort {
	case "alpha":
		alphaTitleType := "perus"
		if publication == "perus_it" {
			alphaTitleType = "italia_perus"
		}
		return fmt.Sprintf(`LOWER(%s) ASC NULLS LAST, s.order_num ASC NULLS LAST, s.id ASC`, alphaTitleExpr(alphaTitleType))
	case "it_pub_date":
		return fmt.Sprintf(`%s ASC NULLS LAST, s.order_num ASC NULLS LAST, s.id ASC`, publicationSortExpr("italia_perus"))
	default:
		return fmt.Sprintf(`%s ASC NULLS LAST, s.order_num ASC NULLS LAST, s.id ASC`, publicationSortExpr("perus"))
	}
}

func buildStoryListWhere(versionID int, params StoryListParams) (string, []interface{}, error) {
	clauses := []string{"s.version = $1"}
	args := []interface{}{versionID}
	argPos := 2

	publicationTypes, err := mapPublicationFilterToTypes(params.Publication)
	if err != nil {
		return "", nil, err
	}
	if len(publicationTypes) > 0 {
		if params.Year > 0 {
			clauses = append(clauses, fmt.Sprintf(`
EXISTS (
	SELECT 1
	FROM stories_in_publications AS sip
	JOIN publications AS p ON p.id = sip.publication
	WHERE sip.story = s.id
	AND p.type::text = ANY($%d)
	AND p.year = $%d
)`, argPos, argPos+1))
			args = append(args, ArrayParam(publicationTypes), params.Year)
			argPos += 2
		} else {
			clauses = append(clauses, fmt.Sprintf(`
EXISTS (
	SELECT 1
	FROM stories_in_publications AS sip
	JOIN publications AS p ON p.id = sip.publication
	WHERE sip.story = s.id
	AND p.type::text = ANY($%d)
)`, argPos))
			args = append(args, ArrayParam(publicationTypes))
			argPos++
		}
	} else if params.Year > 0 {
		clauses = append(clauses, fmt.Sprintf(`
EXISTS (
	SELECT 1
	FROM stories_in_publications AS sip
	JOIN publications AS p ON p.id = sip.publication
	WHERE sip.story = s.id
	AND p.year = $%d
)`, argPos))
		args = append(args, params.Year)
		argPos++
	}

	search := strings.TrimSpace(params.Search)
	if len(search) > 0 {
		clauses = append(clauses, fmt.Sprintf(`
(
	EXISTS (
		SELECT 1
		FROM stories_in_publications AS sip
		JOIN publications AS p ON p.id = sip.publication
		WHERE sip.story = s.id
		AND (
			sip.title ILIKE $%d
			OR p.issue ILIKE $%d
			OR CAST(p.year AS TEXT) ILIKE $%d
			OR p.type::text ILIKE $%d
		)
	)
	OR EXISTS (
		SELECT 1
		FROM authors_in_stories AS sa
		JOIN authors AS a ON a.id = sa.author
		WHERE sa.story = s.id
		AND (
			a.first_name ILIKE $%d
			OR a.last_name ILIKE $%d
			OR (COALESCE(a.first_name, '') || ' ' || COALESCE(a.last_name, '')) ILIKE $%d
		)
	)
)`, argPos, argPos, argPos, argPos, argPos, argPos, argPos))
		args = append(args, "%"+search+"%")
	}

	return strings.Join(clauses, " AND "), args, nil
}

func (s *storyRepo) selectStoryRowsFiltered(version *Version, params StoryListParams) ([]*Story, []int, int, error) {
	if version.ID == 0 || params.PageSize <= 0 || params.Page <= 0 {
		return nil, nil, 0, fmt.Errorf("invalid parameters")
	}

	whereClause, whereArgs, err := buildStoryListWhere(version.ID, params)
	if err != nil {
		return nil, nil, 0, err
	}

	countSQL := fmt.Sprintf(`
SELECT COUNT(*)
FROM stories AS s
WHERE %s;
`, whereClause)

	countRows, err := Query(countSQL, whereArgs...)
	if err != nil {
		return nil, nil, 0, err
	}
	defer countRows.Close()

	total := 0
	if countRows.Next() {
		if err = countRows.Scan(&total); err != nil {
			return nil, nil, 0, err
		}
	}
	if total == 0 {
		return []*Story{}, []int{}, 0, nil
	}

	orderClause := buildSortClause(params.Sort, params.Publication)
	limitArgPos := len(whereArgs) + 1
	offsetArgPos := len(whereArgs) + 2
	offset := (params.Page - 1) * params.PageSize

	querySQL := fmt.Sprintf(`
SELECT
	s.id,
	s.hash,
	s.order_num
FROM stories AS s
WHERE %s
ORDER BY %s
LIMIT $%d OFFSET $%d;
`, whereClause, orderClause, limitArgPos, offsetArgPos)

	args := append(whereArgs, params.PageSize, offset)
	rows, err := Query(querySQL, args...)
	if err != nil {
		return nil, nil, 0, err
	}
	defer rows.Close()

	var stories []*Story
	var storyIDs []int

	for rows.Next() {
		var story Story
		if err = rows.Scan(&story.ID, &story.Hash, &story.OrderNumber); err != nil {
			return nil, nil, 0, err
		}
		stories = append(stories, &story)
		storyIDs = append(storyIDs, story.ID)
	}

	if err = rows.Err(); err != nil {
		return nil, nil, 0, err
	}

	return stories, storyIDs, total, nil
}

const selectAuthorsInStoriesSQL = `
SELECT
	sa.story,
	sa.author,
	sa.type,
	sa.details
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
	a.is_translator
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
	Details sql.NullString
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
			&info.Details,
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
			&a.IsTranslator,
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
	if len(stories) == 0 {
		return stories, nil
	}

	if err = s.hydrateStories(stories, storyIDs); err != nil {
		return nil, err
	}

	return stories, nil
}

func (s *storyRepo) hydrateStories(stories []*Story, storyIDs []int) error {
	// TODO: use channels to not block storyPublication request
	authorInfos, authors, err := s.selectStoryAuthorRows(storyIDs)
	if err != nil {
		return err
	}

	storyPublications, err := s.selectStoryPublicationRows(storyIDs)
	if err != nil {
		return err
	}

	getAuthorsByInfo := func(infos []*ainfo) ([]*Author, []*Author, []*Author) {
		var writers []*Author
		var drawers []*Author
		var translators []*Author

		for _, info := range infos {
			for _, a := range authors {
				if info.Author == a.ID {
					if info.Type == "writer" {
						withDetails := *a
						if info.Details.Valid {
							withDetails.Details = strings.TrimSpace(info.Details.String)
						}
						writers = append(writers, &withDetails)
					} else if info.Type == "drawer" {
						withDetails := *a
						if info.Details.Valid {
							withDetails.Details = strings.TrimSpace(info.Details.String)
						}
						drawers = append(drawers, &withDetails)
					} else if info.Type == "translator" {
						withDetails := *a
						if info.Details.Valid {
							withDetails.Details = strings.TrimSpace(info.Details.String)
						}
						translators = append(translators, &withDetails)
					}
					break
				}
			}
		}

		return writers, drawers, translators
	}

	for idx := range stories {
		infos := authorInfos[stories[idx].ID]
		stories[idx].WrittenBy, stories[idx].DrawnBy, stories[idx].TranslatedBy = getAuthorsByInfo(infos)

		stories[idx].Publications = storyPublications[stories[idx].ID]
	}

	return nil
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
