package db

import (
	"database/sql"
	"fmt"
	"sort"
	"strings"
)

type villainRepo struct{}

var villainPublicationTypesByFilter = map[string][]string{
	"all": {},
	"fi":  {"perus", "maxi", "suur", "muu_erikois", "kronikka", "kirjasto"},
	"it":  {"italia_perus", "italia_erikois"},
}

func (v *villainRepo) BulkCreate(villains []*Villain, version *Version) ([]*Villain, error) {
	// save villains
	if len(villains) > MaxBulkCreateSize {
		return nil, fmt.Errorf("max number of %d villains exceeded", MaxBulkCreateSize)
	}

	var villainValues [][]interface{}

	for _, v := range villains {
		villainValues = append(villainValues, []interface{}{
			v.Hash,
			ArrayParam(v.Ranks),
			ArrayParam(v.FirstNames),
			v.LastName,
			version.ID,
		})
	}
	numRows, err := BulkInsertTxn(bulkInsertParams{
		Table:   "villains",
		Columns: []string{"hash", "ranks", "first_names", "last_name", "version"},
		Values:  villainValues,
	})
	if err != nil {
		return nil, err
	}

	villains, err = v.setIDsFromDB(villains, numRows)
	if err != nil {
		return nil, err
	}

	// save story villains
	var storyVillainValues [][]interface{}

	for _, v := range villains {
		for _, sv := range v.As {
			if v.ID == 0 {
				fmt.Println("v: ", v)
			}
			storyVillainValues = append(storyVillainValues, []interface{}{
				v.ID,
				sv.Story.ID, // TODO: make sure that this is found
				sv.Hash,
				ArrayParam(sv.Nicknames),
				ArrayParam(sv.Aliases),
				ArrayParam(sv.Destiny),
				ArrayParam(sv.Roles),
			})
		}
	}

	_, err = BulkInsertTxn(bulkInsertParams{
		Table:   "villains_in_stories",
		Columns: []string{"villain", "story", "hash", "nicknames", "aliases", "destiny", "roles"},
		Values:  storyVillainValues,
	})
	if err != nil {
		return nil, err
	}

	// list created villains

	return nil, nil
}

const setVillainIDsSQL = `
SELECT
	v.id,
	v.hash
FROM villains as v
ORDER BY v.id DESC
LIMIT %v;
`

func (v *villainRepo) setIDsFromDB(villains []*Villain, savedRows int64) ([]*Villain, error) {
	queryString := fmt.Sprintf(setVillainIDsSQL, savedRows)
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

		for idx := range villains {
			if villains[idx].Hash == row.Hash {
				villains[idx].ID = row.ID
				break
			}
			if idx == len(villains)-1 && villains[idx].Hash != row.Hash {
				return nil, fmt.Errorf("matching villain not found from db")
			}
		}
	}

	return villains, nil
}

func mapVillainPublicationFilterToTypes(filter string) ([]string, error) {
	types, found := villainPublicationTypesByFilter[filter]
	if !found {
		return nil, fmt.Errorf("invalid publication filter")
	}
	return types, nil
}

func buildVillainSortClause(sort string) string {
	normalizeSQL := func(expr string) string {
		return fmt.Sprintf(`NULLIF(regexp_replace(lower(COALESCE(%s, '')), '[[:punct:][:space:]]+', '', 'g'), '')`, expr)
	}
	publicationSortExpr := func(pubType string) string {
		return fmt.Sprintf(`(
	SELECT MIN(
		(COALESCE(p.year, 0) * 1000) + COALESCE(NULLIF(regexp_replace(p.issue, '[^0-9]', '', 'g'), '')::int, 0)
	)
	FROM villains_in_stories AS vis
	JOIN stories_in_publications AS sip ON sip.story = vis.story
	JOIN publications AS p ON p.id = sip.publication
	WHERE vis.villain = v.id
	AND p.type = '%s'
)`, pubType)
	}

	firstNameExpr := normalizeSQL("array_to_string(v.first_names, ' ')")
	lastNameExpr := normalizeSQL("v.last_name")
	nicknameExpr := fmt.Sprintf(`(
	SELECT MIN(%s)
	FROM villains_in_stories AS vis
	WHERE vis.villain = v.id
)`, normalizeSQL("array_to_string(vis.nicknames, ' ')"))
	rankExpr := normalizeSQL("array_to_string(v.ranks, ' ')")

	switch sort {
	case "fi_pub_date":
		return fmt.Sprintf(`%s ASC NULLS LAST, %s ASC NULLS LAST, %s ASC NULLS LAST, v.id ASC`, publicationSortExpr("perus"), lastNameExpr, firstNameExpr)
	case "it_pub_date":
		return fmt.Sprintf(`%s ASC NULLS LAST, %s ASC NULLS LAST, %s ASC NULLS LAST, v.id ASC`, publicationSortExpr("italia_perus"), lastNameExpr, firstNameExpr)
	case "last_name":
		return fmt.Sprintf(`%s ASC NULLS LAST, %s ASC NULLS LAST, v.id ASC`, lastNameExpr, firstNameExpr)
	case "nickname":
		return fmt.Sprintf(`%s ASC NULLS LAST, %s ASC NULLS LAST, v.id ASC`, nicknameExpr, lastNameExpr)
	case "rank":
		return fmt.Sprintf(`%s ASC NULLS LAST, %s ASC NULLS LAST, v.id ASC`, rankExpr, lastNameExpr)
	default:
		return fmt.Sprintf(`%s ASC NULLS LAST, %s ASC NULLS LAST, v.id ASC`, firstNameExpr, lastNameExpr)
	}
}

func buildVillainListWhere(versionID int, params VillainListParams) (string, []interface{}, error) {
	clauses := []string{"v.version = $1"}
	args := []interface{}{versionID}
	argPos := 2

	publicationTypes, err := mapVillainPublicationFilterToTypes(params.Publication)
	if err != nil {
		return "", nil, err
	}
	if len(publicationTypes) > 0 {
		clauses = append(clauses, fmt.Sprintf(`
EXISTS (
	SELECT 1
	FROM villains_in_stories AS vis
	JOIN stories AS s ON s.id = vis.story
	JOIN stories_in_publications AS sip ON sip.story = s.id
	JOIN publications AS p ON p.id = sip.publication
	WHERE vis.villain = v.id
	AND p.type::text = ANY($%d)
)`, argPos))
		args = append(args, ArrayParam(publicationTypes))
		argPos++
	}

	search := strings.TrimSpace(params.Search)
	if len(search) > 0 {
		clauses = append(clauses, fmt.Sprintf(`
(
	array_to_string(v.first_names, ' ') ILIKE $%d
	OR v.last_name ILIKE $%d
	OR array_to_string(v.ranks, ' ') ILIKE $%d
	OR EXISTS (
		SELECT 1
		FROM villains_in_stories AS vis
		LEFT JOIN stories AS s ON s.id = vis.story
		LEFT JOIN stories_in_publications AS sip ON sip.story = s.id
		LEFT JOIN publications AS p ON p.id = sip.publication
		WHERE vis.villain = v.id
		AND (
			array_to_string(vis.nicknames, ' ') ILIKE $%d
			OR array_to_string(vis.aliases, ' ') ILIKE $%d
			OR array_to_string(vis.roles, ' ') ILIKE $%d
			OR array_to_string(vis.destiny, ' ') ILIKE $%d
			OR sip.title ILIKE $%d
			OR p.issue ILIKE $%d
			OR p.type::text ILIKE $%d
		)
	)
)`, argPos, argPos, argPos, argPos, argPos, argPos, argPos, argPos, argPos, argPos))
		args = append(args, "%"+search+"%")
	}

	return strings.Join(clauses, " AND "), args, nil
}

func (v *villainRepo) selectVillainRowsFiltered(version *Version, params VillainListParams) ([]*Villain, []int, int, error) {
	if version.ID == 0 || params.Page <= 0 || params.PageSize <= 0 {
		return nil, nil, 0, fmt.Errorf("invalid parameters")
	}

	whereClause, whereArgs, err := buildVillainListWhere(version.ID, params)
	if err != nil {
		return nil, nil, 0, err
	}

	countSQL := fmt.Sprintf(`
SELECT COUNT(*)
FROM villains AS v
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
		return []*Villain{}, []int{}, 0, nil
	}

	orderClause := buildVillainSortClause(params.Sort)
	limitArgPos := len(whereArgs) + 1
	offsetArgPos := len(whereArgs) + 2
	offset := (params.Page - 1) * params.PageSize

	querySQL := fmt.Sprintf(`
SELECT
	v.id,
	v.hash,
	COALESCE(v.ranks, ARRAY[]::varchar[]),
	COALESCE(v.first_names, ARRAY[]::varchar[]),
	COALESCE(v.last_name, '')
FROM villains AS v
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

	var villains []*Villain
	var villainIDs []int

	for rows.Next() {
		var row struct {
			ID         int
			Hash       string
			Ranks      []string
			FirstNames []string
			LastName   string
		}
		if err = rows.Scan(
			&row.ID,
			&row.Hash,
			ArrayParam(&row.Ranks),
			ArrayParam(&row.FirstNames),
			&row.LastName,
		); err != nil {
			return nil, nil, 0, err
		}

		villains = append(villains, &Villain{
			ID:         row.ID,
			Hash:       row.Hash,
			Ranks:      row.Ranks,
			FirstNames: row.FirstNames,
			LastName:   row.LastName,
		})
		villainIDs = append(villainIDs, row.ID)
	}

	if err = rows.Err(); err != nil {
		return nil, nil, 0, err
	}

	return villains, villainIDs, total, nil
}

const selectStoryVillainsByVillainIDsSQL = `
SELECT
	vis.villain,
	vis.id,
	vis.hash,
	COALESCE(vis.nicknames, ARRAY[]::varchar[]),
	COALESCE(vis.aliases, ARRAY[]::varchar[]),
	COALESCE(vis.roles, ARRAY[]::varchar[]),
	COALESCE(vis.destiny, ARRAY[]::varchar[]),
	s.id,
	s.hash,
	s.order_num
FROM villains_in_stories AS vis
JOIN stories AS s ON s.id = vis.story
WHERE vis.villain = ANY($1)
ORDER BY
	vis.villain ASC,
	s.order_num ASC NULLS LAST,
	s.id ASC;
`

func (*villainRepo) selectStoryVillainRows(villainIDs []int) (map[int][]*StoryVillain, []*Story, []int, error) {
	rows, err := Query(selectStoryVillainsByVillainIDsSQL, ArrayParam(villainIDs))
	if err != nil {
		return nil, nil, nil, err
	}
	defer rows.Close()

	asByVillain := make(map[int][]*StoryVillain)
	storyByID := make(map[int]*Story)

	for rows.Next() {
		var row struct {
			VillainID int
			ID        int
			Hash      string
			Nicknames []string
			Aliases   []string
			Roles     []string
			Destiny   []string
			StoryID   int
			StoryHash string
			OrderNum  sql.NullInt64
		}

		if err = rows.Scan(
			&row.VillainID,
			&row.ID,
			&row.Hash,
			ArrayParam(&row.Nicknames),
			ArrayParam(&row.Aliases),
			ArrayParam(&row.Roles),
			ArrayParam(&row.Destiny),
			&row.StoryID,
			&row.StoryHash,
			&row.OrderNum,
		); err != nil {
			return nil, nil, nil, err
		}

		story := storyByID[row.StoryID]
		if story == nil {
			story = &Story{
				ID:    row.StoryID,
				Hash:  row.StoryHash,
				OrderNumber: func() int {
					if row.OrderNum.Valid {
						return int(row.OrderNum.Int64)
					}
					return 0
				}(),
			}
			storyByID[row.StoryID] = story
		}

		storyVillain := &StoryVillain{
			ID:        row.ID,
			Hash:      row.Hash,
			Nicknames: row.Nicknames,
			Aliases:   row.Aliases,
			Roles:     row.Roles,
			Destiny:   row.Destiny,
			Story:     story,
		}
		asByVillain[row.VillainID] = append(asByVillain[row.VillainID], storyVillain)
	}

	if err = rows.Err(); err != nil {
		return nil, nil, nil, err
	}

	var storyIDs []int
	for id := range storyByID {
		storyIDs = append(storyIDs, id)
	}
	sort.Ints(storyIDs)

	var stories []*Story
	for _, id := range storyIDs {
		stories = append(stories, storyByID[id])
	}

	return asByVillain, stories, storyIDs, nil
}

func (v *villainRepo) hydrateVillains(villains []*Villain, villainIDs []int) error {
	asByVillain, stories, storyIDs, err := v.selectStoryVillainRows(villainIDs)
	if err != nil {
		return err
	}

	if len(stories) > 0 {
		storyRepo := &storyRepo{}
		if err = storyRepo.hydrateStories(stories, storyIDs); err != nil {
			return err
		}
	}

	for _, villain := range villains {
		villain.As = asByVillain[villain.ID]
	}

	return nil
}

// ListFiltered implements VillainRepository.
func (v *villainRepo) ListFiltered(version *Version, params VillainListParams) ([]*Villain, int, error) {
	villains, villainIDs, total, err := v.selectVillainRowsFiltered(version, params)
	if err != nil {
		return nil, 0, err
	}
	if len(villains) == 0 {
		return villains, total, nil
	}

	if err = v.hydrateVillains(villains, villainIDs); err != nil {
		return nil, 0, err
	}
	return villains, total, nil
}

func NewVillainRepository() VillainRepository {
	return &villainRepo{}
}
