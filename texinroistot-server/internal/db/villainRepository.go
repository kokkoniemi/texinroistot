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
	"it": {
		"italia_perus",
		"italia_erikois",
		"italia_serie_extra",
		"italia_texone",
		"italia_mini_texone_maxi_tex",
		"italia_almanacco_del_west",
		"italia_color_tex",
		"italia_tex_romanzi_a_fumetti",
		"italia_tex_magazine",
	},
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
				ArrayParam(sv.OtherNames),
				ArrayParam(sv.CodeNames),
				ArrayParam(sv.Destiny),
				ArrayParam(sv.Roles),
				ArrayParam(sv.FirstNames),
				sv.LastName,
			})
		}
	}

	_, err = BulkInsertTxn(bulkInsertParams{
		Table:   "villains_in_stories",
		Columns: []string{"villain", "story", "hash", "nicknames", "other_names", "code_names", "destiny", "roles", "first_names", "last_name"},
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

func normalizeSortSQL(expr string) string {
	return fmt.Sprintf(`NULLIF(regexp_replace(lower(COALESCE(%s, '')), '[[:punct:][:space:]]+', '', 'g'), '')`, expr)
}

func storyVillainSortSQL(column string) string {
	return fmt.Sprintf(`(
	SELECT MIN(%s)
	FROM villains_in_stories AS vis
	WHERE vis.villain = v.id
)`, normalizeSortSQL(fmt.Sprintf("array_to_string(vis.%s, ' ')", column)))
}

// storyVillainNameColumns are the per-story name lists that have their own
// alphabetical sort. In those sorts a villain is listed once under each
// distinct name, so a villain with the nicknames "Ice" and "Blue" is found
// under both B and I.
var storyVillainNameColumns = map[string]string{
	"nickname":   "nicknames",
	"other_name": "other_names",
	"code_name":  "code_names",
}

// villainListFrom returns the FROM clause of a villain list. For the per-story
// name sorts each villain row is repeated for each of its distinct names,
// exposed as sn.key (the normalized sort key) and sn.name.
func villainListFrom(sort string) string {
	column, expanded := storyVillainNameColumns[sort]
	if !expanded {
		return "villains AS v"
	}
	key := normalizeSortSQL("btrim(n.name)")
	return fmt.Sprintf(`villains AS v
CROSS JOIN LATERAL (
	SELECT DISTINCT ON (%s) %s AS key, btrim(n.name) AS name
	FROM villains_in_stories AS vis
	CROSS JOIN LATERAL unnest(vis.%s) AS n(name)
	JOIN stories AS ns ON ns.id = vis.story
	WHERE vis.villain = v.id
	AND %s IS NOT NULL
	ORDER BY %s, ns.order_num ASC NULLS LAST
) AS sn`, key, key, column, key, key)
}

// villainListSortKeyExpr is the primary sort key that a villain list is
// ordered and letter-paged by: sn.key for the per-story name sorts, otherwise
// villainSortKeyExpr.
func villainListSortKeyExpr(sort string) string {
	if _, expanded := storyVillainNameColumns[sort]; expanded {
		return "sn.key"
	}
	return villainSortKeyExpr(sort)
}

// villainSortKeyExpr returns the primary sort key of an alphabetical villain
// sort, or "" for the publication date sorts and the grouped default sort,
// which have no single key. Letter paging uses the same key.
func villainSortKeyExpr(sort string) string {
	switch sort {
	case "fi_pub_date", "it_pub_date", "default":
		return ""
	case "last_name":
		return normalizeSortSQL("v.last_name")
	case "nickname":
		return storyVillainSortSQL("nicknames")
	case "other_name":
		return storyVillainSortSQL("other_names")
	case "code_name":
		return storyVillainSortSQL("code_names")
	case "rank":
		return normalizeSortSQL("array_to_string(v.ranks, ' ')")
	default:
		return normalizeSortSQL("array_to_string(v.first_names, ' ')")
	}
}

// firstAppearanceSQL is the order number of the first story a villain
// appears in.
const firstAppearanceSQL = `(
	SELECT MIN(fs.order_num)
	FROM villains_in_stories AS fvis
	JOIN stories AS fs ON fs.id = fvis.story
	WHERE fvis.villain = v.id
)`

func buildVillainSortClause(sort string) string {
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

	firstNameExpr := villainSortKeyExpr("first_name")
	lastNameExpr := villainSortKeyExpr("last_name")

	switch sort {
	case "fi_pub_date":
		return fmt.Sprintf(`%s ASC NULLS LAST, %s ASC NULLS LAST, %s ASC NULLS LAST, v.id ASC`, publicationSortExpr("perus"), lastNameExpr, firstNameExpr)
	case "it_pub_date":
		return fmt.Sprintf(`%s ASC NULLS LAST, %s ASC NULLS LAST, %s ASC NULLS LAST, v.id ASC`, publicationSortExpr("italia_perus"), lastNameExpr, firstNameExpr)
	case "last_name":
		return fmt.Sprintf(`%s ASC NULLS LAST, %s ASC NULLS LAST, v.id ASC`, lastNameExpr, firstNameExpr)
	case "rank":
		// Within a rank: villains with a last name by last name (Sheriffi
		// Abbott ... Whitaker), then those without one by the name they do
		// have, first name, other name, nickname or code name (Sheriffi Clem,
		// Harry, Mike ...), and bare ranks last in order of the story they
		// first appear in, which also settles villains with the same name.
		return fmt.Sprintf(`%s ASC NULLS LAST, %s ASC NULLS LAST, COALESCE(%s, %s, %s, %s) ASC NULLS LAST, %s ASC NULLS LAST, v.id ASC`,
			villainListSortKeyExpr(sort), lastNameExpr, firstNameExpr,
			villainSortKeyExpr("other_name"), villainSortKeyExpr("nickname"), villainSortKeyExpr("code_name"),
			firstAppearanceSQL)
	case "nickname", "other_name", "code_name":
		return fmt.Sprintf(`%s ASC NULLS LAST, %s ASC NULLS LAST, v.id ASC`, villainListSortKeyExpr(sort), lastNameExpr)
	case "first_name":
		return fmt.Sprintf(`%s ASC NULLS LAST, %s ASC NULLS LAST, v.id ASC`, firstNameExpr, lastNameExpr)
	default:
		return defaultVillainSortClause()
	}
}

// defaultVillainSortKeys are the name fields of the default sort in group
// order: a villain belongs to the first group whose name it has.
var defaultVillainSortKeys = []string{"first_name", "last_name", "other_name", "nickname", "rank", "code_name"}

// defaultVillainSortClause orders villains group by group (first name, then
// last name, other name, nickname, rank, code name, then villains with none of
// these), each group alphabetically by its own name. Villains with the same
// name (e.g. several bare "Sheriffi") are in order of the story they first
// appear in. It has no letter paging.
func defaultVillainSortClause() string {
	keys := make([]string, len(defaultVillainSortKeys))
	groupCases := make([]string, len(defaultVillainSortKeys))
	for i, sort := range defaultVillainSortKeys {
		keys[i] = villainSortKeyExpr(sort)
		groupCases[i] = fmt.Sprintf("WHEN %s IS NOT NULL THEN %d", keys[i], i+1)
	}
	return fmt.Sprintf(
		`CASE %s ELSE %d END ASC, COALESCE(%s) ASC NULLS LAST, %s ASC NULLS LAST, %s ASC NULLS LAST, v.id ASC`,
		strings.Join(groupCases, " "),
		len(defaultVillainSortKeys)+1,
		strings.Join(keys, ", "),
		villainSortKeyExpr("last_name"),
		firstAppearanceSQL,
	)
}

func buildVillainSortRequirementClause(sort string) string {
	nonEmptyArray := func(expr string) string {
		return fmt.Sprintf(`COALESCE(btrim(array_to_string(%s, '')), '') <> ''`, expr)
	}
	nonEmptyStoryVillainArray := func(column string) string {
		return fmt.Sprintf(`EXISTS (
	SELECT 1 FROM villains_in_stories AS vis
	WHERE vis.villain = v.id
	AND COALESCE(btrim(array_to_string(vis.%s, '')), '') <> ''
)`, column)
	}

	switch sort {
	case "first_name":
		return nonEmptyArray("v.first_names")
	case "last_name":
		return `COALESCE(btrim(v.last_name), '') <> ''`
	case "rank":
		return nonEmptyArray("v.ranks")
	case "nickname":
		return nonEmptyStoryVillainArray("nicknames")
	case "other_name":
		return nonEmptyStoryVillainArray("other_names")
	case "code_name":
		return nonEmptyStoryVillainArray("code_names")
	default:
		return ""
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

	if sortRequirement := buildVillainSortRequirementClause(params.Sort); sortRequirement != "" {
		clauses = append(clauses, sortRequirement)
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
		WHERE vis.villain = v.id
		AND (
			array_to_string(vis.nicknames, ' ') ILIKE $%d
			OR array_to_string(vis.other_names, ' ') ILIKE $%d
			OR array_to_string(vis.code_names, ' ') ILIKE $%d
			OR array_to_string(vis.first_names, ' ') ILIKE $%d
			OR vis.last_name ILIKE $%d
		)
	)
)`, argPos, argPos, argPos, argPos, argPos, argPos, argPos, argPos))
		args = append(args, "%"+search+"%")
		argPos++
	}

	villainHash := strings.TrimSpace(params.VillainHash)
	if len(villainHash) > 0 {
		clauses = append(clauses, fmt.Sprintf("v.hash = $%d", argPos))
		args = append(args, villainHash)
	}

	return strings.Join(clauses, " AND "), args, nil
}

// searchMatchesRank reports whether a search term matches the rank of a listed
// villain ("sheriffi", "kapteeni"). Such a search in the default order is
// listed like the rank order: by rank, then last name, then the other names,
// so all sheriffs are in last name order instead of grouped by first name.
func searchMatchesRank(whereClause string, whereArgs []interface{}, search string) (bool, error) {
	search = strings.TrimSpace(search)
	if search == "" {
		return false, nil
	}
	rows, err := Query(fmt.Sprintf(`
SELECT EXISTS (
	SELECT 1 FROM villains AS v
	WHERE %s
	AND array_to_string(v.ranks, ' ') ILIKE $%d
);
`, whereClause, len(whereArgs)+1), append(append([]interface{}{}, whereArgs...), "%"+search+"%")...)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	matches := false
	if rows.Next() {
		if err = rows.Scan(&matches); err != nil {
			return false, err
		}
	}
	return matches, rows.Err()
}

func (v *villainRepo) selectVillainRowsFiltered(version *Version, params VillainListParams) ([]*Villain, []int, int, error) {
	if version.ID == 0 || params.Page <= 0 || params.PageSize <= 0 {
		return nil, nil, 0, fmt.Errorf("invalid parameters")
	}

	whereClause, whereArgs, err := buildVillainListWhere(version.ID, params)
	if err != nil {
		return nil, nil, 0, err
	}

	from := villainListFrom(params.Sort)
	countSQL := fmt.Sprintf(`
SELECT COUNT(*)
FROM %s
WHERE %s;
`, from, whereClause)

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
	if params.Sort == "default" {
		byRank, err := searchMatchesRank(whereClause, whereArgs, params.Search)
		if err != nil {
			return nil, nil, 0, err
		}
		if byRank {
			orderClause = buildVillainSortClause("rank")
		}
	}
	limitArgPos := len(whereArgs) + 1
	offsetArgPos := len(whereArgs) + 2
	offset := (params.Page - 1) * params.PageSize

	sortNameExpr := "''"
	if _, expanded := storyVillainNameColumns[params.Sort]; expanded {
		sortNameExpr = "sn.name"
	}

	querySQL := fmt.Sprintf(`
SELECT
	v.id,
	v.hash,
	COALESCE(v.ranks, ARRAY[]::varchar[]),
	COALESCE(v.first_names, ARRAY[]::varchar[]),
	COALESCE(v.last_name, ''),
	%s
FROM %s
WHERE %s
ORDER BY %s
LIMIT $%d OFFSET $%d;
`, sortNameExpr, from, whereClause, orderClause, limitArgPos, offsetArgPos)

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
			SortName   string
		}
		if err = rows.Scan(
			&row.ID,
			&row.Hash,
			ArrayParam(&row.Ranks),
			ArrayParam(&row.FirstNames),
			&row.LastName,
			&row.SortName,
		); err != nil {
			return nil, nil, 0, err
		}

		villains = append(villains, &Villain{
			ID:         row.ID,
			Hash:       row.Hash,
			Ranks:      row.Ranks,
			FirstNames: row.FirstNames,
			LastName:   row.LastName,
			SortName:   row.SortName,
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
	COALESCE(vis.other_names, ARRAY[]::varchar[]),
	COALESCE(vis.code_names, ARRAY[]::varchar[]),
	COALESCE(vis.roles, ARRAY[]::varchar[]),
	COALESCE(vis.destiny, ARRAY[]::varchar[]),
	COALESCE(vis.first_names, ARRAY[]::varchar[]),
	COALESCE(vis.last_name, ''),
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
			VillainID  int
			ID         int
			Hash       string
			Nicknames  []string
			OtherNames []string
			CodeNames  []string
			Roles      []string
			Destiny    []string
			FirstNames []string
			LastName   string
			StoryID    int
			StoryHash  string
			OrderNum   sql.NullInt64
		}

		if err = rows.Scan(
			&row.VillainID,
			&row.ID,
			&row.Hash,
			ArrayParam(&row.Nicknames),
			ArrayParam(&row.OtherNames),
			ArrayParam(&row.CodeNames),
			ArrayParam(&row.Roles),
			ArrayParam(&row.Destiny),
			ArrayParam(&row.FirstNames),
			&row.LastName,
			&row.StoryID,
			&row.StoryHash,
			&row.OrderNum,
		); err != nil {
			return nil, nil, nil, err
		}

		story := storyByID[row.StoryID]
		if story == nil {
			story = &Story{
				ID:   row.StoryID,
				Hash: row.StoryHash,
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
			ID:         row.ID,
			Hash:       row.Hash,
			FirstNames: row.FirstNames,
			LastName:   row.LastName,
			Nicknames:  row.Nicknames,
			OtherNames: row.OtherNames,
			CodeNames:  row.CodeNames,
			Roles:      row.Roles,
			Destiny:    row.Destiny,
			Story:      story,
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

const selectVillainsByStoryHashSQL = `
SELECT
	COALESCE(v.id, 0),
	COALESCE(v.hash, ''),
	COALESCE(v.ranks, ARRAY[]::varchar[]),
	COALESCE(v.first_names, ARRAY[]::varchar[]),
	COALESCE(v.last_name, ''),
	COALESCE(vis.id, 0),
	COALESCE(vis.hash, ''),
	COALESCE(vis.nicknames, ARRAY[]::varchar[]),
	COALESCE(vis.other_names, ARRAY[]::varchar[]),
	COALESCE(vis.code_names, ARRAY[]::varchar[]),
	COALESCE(vis.roles, ARRAY[]::varchar[]),
	COALESCE(vis.destiny, ARRAY[]::varchar[]),
	COALESCE(vis.first_names, ARRAY[]::varchar[]),
	COALESCE(vis.last_name, ''),
	COALESCE(s.id, 0),
	COALESCE(s.hash, ''),
	(
		SELECT COUNT(DISTINCT all_vis.story)
		FROM villains_in_stories AS all_vis
		WHERE all_vis.villain = v.id
	)
FROM stories AS s
LEFT JOIN villains_in_stories AS vis ON vis.story = s.id
LEFT JOIN villains AS v ON v.id = vis.villain AND v.version = s.version
WHERE s.hash = $1
AND s.version = $2
ORDER BY
	NULLIF(regexp_replace(lower(
		CASE WHEN ` + storyVillainHasNameSQL + `
			THEN COALESCE(array_to_string(vis.first_names, ' '), '') || COALESCE(vis.last_name, '')
			ELSE COALESCE(array_to_string(v.first_names, ' '), '') || COALESCE(v.last_name, '')
		END
	), '[[:punct:][:space:]]+', '', 'g'), '') ASC NULLS LAST,
	v.id ASC,
	vis.id ASC;
`

// storyVillainHasNameSQL is true when a story gives the villain a name of its
// own, which then replaces the villain's name in that story.
const storyVillainHasNameSQL = `(COALESCE(btrim(array_to_string(vis.first_names, '')), '') <> '' OR COALESCE(btrim(vis.last_name), '') <> '')`

// hasStoryName reports whether a story gives the villain a name of its own.
func hasStoryName(firstNames []string, lastName string) bool {
	if strings.TrimSpace(lastName) != "" {
		return true
	}
	for _, name := range firstNames {
		if strings.TrimSpace(name) != "" {
			return true
		}
	}
	return false
}

// ListByStoryHash returns villains that appear in a single story in the active version.
func (*villainRepo) ListByStoryHash(version *Version, storyHash string) ([]*Villain, bool, error) {
	if version == nil || version.ID == 0 {
		return nil, false, fmt.Errorf("invalid version")
	}

	storyHash = strings.TrimSpace(storyHash)
	if storyHash == "" {
		return nil, false, fmt.Errorf("story hash is required")
	}

	rows, err := Query(selectVillainsByStoryHashSQL, storyHash, version.ID)
	if err != nil {
		return nil, false, err
	}
	defer rows.Close()

	storyFound := false
	var villains []*Villain

	for rows.Next() {
		storyFound = true

		var row struct {
			VillainID        int
			VillainHash      string
			Ranks            []string
			FirstNames       []string
			LastName         string
			StoryVillainID   int
			StoryVillainHash string
			Nicknames        []string
			OtherNames       []string
			CodeNames        []string
			Roles            []string
			Destiny          []string
			StoryFirstNames  []string
			StoryLastName    string
			StoryID          int
			StoryHash        string
			StoryCount       int
		}

		if err = rows.Scan(
			&row.VillainID,
			&row.VillainHash,
			ArrayParam(&row.Ranks),
			ArrayParam(&row.FirstNames),
			&row.LastName,
			&row.StoryVillainID,
			&row.StoryVillainHash,
			ArrayParam(&row.Nicknames),
			ArrayParam(&row.OtherNames),
			ArrayParam(&row.CodeNames),
			ArrayParam(&row.Roles),
			ArrayParam(&row.Destiny),
			ArrayParam(&row.StoryFirstNames),
			&row.StoryLastName,
			&row.StoryID,
			&row.StoryHash,
			&row.StoryCount,
		); err != nil {
			return nil, false, err
		}

		if row.VillainID == 0 || row.StoryVillainID == 0 {
			continue
		}

		// In a single story the villain goes by the name that story gives,
		// e.g. "Kid Rodelo" before he is revealed to be Philip Rainey.
		firstNames, lastName := row.FirstNames, row.LastName
		if hasStoryName(row.StoryFirstNames, row.StoryLastName) {
			firstNames, lastName = row.StoryFirstNames, row.StoryLastName
		}

		villains = append(villains, &Villain{
			ID:         row.VillainID,
			Hash:       row.VillainHash,
			Ranks:      row.Ranks,
			FirstNames: firstNames,
			LastName:   lastName,
			StoryCount: row.StoryCount,
			As: []*StoryVillain{
				{
					ID:         row.StoryVillainID,
					Hash:       row.StoryVillainHash,
					FirstNames: row.StoryFirstNames,
					LastName:   row.StoryLastName,
					Nicknames:  row.Nicknames,
					OtherNames: row.OtherNames,
					CodeNames:  row.CodeNames,
					Roles:      row.Roles,
					Destiny:    row.Destiny,
					Story: &Story{
						ID:   row.StoryID,
						Hash: row.StoryHash,
					},
				},
			},
		})
	}

	if err = rows.Err(); err != nil {
		return nil, false, err
	}

	return villains, storyFound, nil
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
