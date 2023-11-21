package admin

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/google/uuid"

	"github.com/kokkoniemi/texinroistot/internal/crypt"
	"github.com/kokkoniemi/texinroistot/internal/db"
)

type id string
type row struct {
	importer *sprdImporter
	columns  []string
}

func (r row) getValue(key string) string {
	index, ok := r.importer.columnIndexes[key]
	if !ok {
		return ""
	}
	return r.columns[index]
}

type importerAuthor struct {
	ID   id
	item *db.Author
}

type importerStory struct {
	ID        id
	item      *db.Story
	writers   []id
	drawers   []id
	inventors []id
}

type importerPublication struct {
	ID   id
	item *db.Publication
}

type importerStoryPublication struct {
	ID          id
	story       id
	publication id
	title       string
}

type importerVillain struct {
	ID   id
	item *db.Villain
}

type importerStoryVillain struct {
	ID      id
	story   id
	villain id
}

var defaultColumns = map[string]string{
	"Arvo":                     "rank",
	"Etunimi":                  "first_name",
	"Sukunimi":                 "last_name",
	"Lempinimi/Intiaaninimi":   "nickname",
	"Salanimi/Alias":           "alias",
	"Rooli":                    "role",
	"Kohtalo":                  "destiny",
	"Tarina":                   "story_title",
	"Kertoi":                   "story_written_by",
	"Piirsi":                   "story_drawn_by",
	"Käsikirjoitti/Ideoi":      "story_invented_by",
	"Vuosi":                    "pub_year",
	"Alkaen":                   "pub_from",
	"Päättyen":                 "pub_to",
	"UVuosi":                   "repub_year",
	"Ualkaen":                  "repub_from",
	"Upäättyen":                "repub_to",
	"Erikoisjulkaisu":          "pub_special",
	"Kronikka":                 "pub_kronikka",
	"Kirjasto":                 "pub_kirjasto",
	"Italian vuosi":            "italy_year",
	"Italian alkunumero":       "italy_pub_from",
	"Italian päättymisnumero":  "italy_pub_to",
	"Italian erikoisjulkaisu":  "italy_pub_special",
	"Italian tarina":           "italy_story_title",
	"Järjestysluku":            "story_order_num",
	"Sama numero, sama roisto": "villain_id",
}

type sprdImporter struct {
	authors           []*importerAuthor
	stories           []*importerStory
	publications      []*importerPublication
	storyPublications []*importerStoryPublication
	villains          []*importerVillain
	storyVillains     []*importerStoryVillain
	columnIndexes     map[string]int
	columnNames       map[string]string
}

func NewSpreadsheetImporter(titleRow []string) *sprdImporter {
	columnNames := defaultColumns
	columnIndexes := map[string]int{}

	for index, title := range titleRow {
		key := columnNames[title]
		columnIndexes[key] = index
	}

	return &sprdImporter{
		columnNames:   defaultColumns,
		columnIndexes: columnIndexes,
	}
}

func (i *sprdImporter) LoadData(dataRows [][]string) error {
	for index, dataRow := range dataRows {
		// FIXME: remove after importer is ready
		if index > 200 {
			break
		}
		row := row{importer: i, columns: dataRow}

		storyID, err := i.loadStory(row)
		if err != nil {
			return err
		}
		i.loadWriters(storyID, row)
		i.loadDrawer(storyID, row)
		i.loadInventor(storyID, row)
		err = i.loadBasePublication(storyID, row)
		if err != nil {
			return err
		}
		i.loadBaseRePublication(storyID, row)
		i.loadItalianBasePublication(storyID, row)
		i.loadSpecialPublication(storyID, row)
		i.loadItalianSpecialPublication(storyID, row)
		i.loadKronikka(storyID, row)
		i.loadKirjasto(storyID, row)
		i.loadVillain(storyID, row)
	}

	return nil
}

func (i *sprdImporter) getAuthorIndexWithName(firstName string, lastName string) int {
	return slices.IndexFunc(i.authors, func(a *importerAuthor) bool {
		return a.item.FirstName == firstName && a.item.LastName == lastName
	})
}

func (i *sprdImporter) hasAuthorWithName(firstName string, lastName string) bool {
	return i.getAuthorIndexWithName(firstName, lastName) != -1
}

func (i *sprdImporter) getAuthorWithName(firstName string, lastName string) *importerAuthor {
	idx := i.getAuthorIndexWithName(firstName, lastName)
	if idx != -1 {
		return i.authors[idx]
	}
	return nil
}

func (i *sprdImporter) getStory(storyID id) *importerStory {
	storyIdx := slices.IndexFunc(i.stories, func(s *importerStory) bool {
		return s.ID == storyID
	})
	if storyIdx != -1 {
		return i.stories[storyIdx]
	}
	return nil
}

func (i *sprdImporter) setWriterForStory(storyID id, writerID id) {
	story := i.getStory(storyID)
	if story != nil && !slices.Contains(story.writers, writerID) {
		story.writers = append(story.writers, writerID)
	}
}

func (i *sprdImporter) setDrawerForStory(storyID id, drawerID id) {
	story := i.getStory(storyID)
	if story != nil && !slices.Contains(story.drawers, drawerID) {
		story.drawers = append(story.drawers, drawerID)
	}
}

func (i *sprdImporter) setInventorForStory(storyID id, inventorID id) {
	story := i.getStory(storyID)
	if story != nil && !slices.Contains(story.inventors, inventorID) {
		story.inventors = append(story.inventors, inventorID)
	}
}

func (i *sprdImporter) loadStory(r row) (id, error) {
	orderNum, err := strconv.Atoi(r.getValue("story_order_num"))
	if err != nil {
		return "", err
	}

	hash := ""
	if orderNum != 0 {
		hash = crypt.Hash(fmt.Sprintf("%v", orderNum))
	} else {
		hash = crypt.Hash(r.getValue("story_title"))
	}

	s := &importerStory{
		ID: id(uuid.NewString()),
		item: &db.Story{
			Hash:        hash,
			OrderNumber: orderNum,
		},
	}

	i.stories = append(i.stories, s)

	return s.ID, nil
}

func (i *sprdImporter) loadAuthorColumn(r row, columnName string) []*importerAuthor {
	names := strings.Split(r.getValue(columnName), ";")
	var authors []*importerAuthor
	for _, n := range names {
		if len(n) == 0 {
			continue
		}
		nameParts := strings.Split(n, ",")
		var firstName string
		var lastName string
		if len(nameParts) == 1 {
			firstName = nameParts[0]
		} else {
			lastName = nameParts[0]
			firstName = strings.TrimSpace(strings.Join(nameParts[1:], " "))
		}

		var author *importerAuthor

		if !i.hasAuthorWithName(firstName, lastName) {
			author = &importerAuthor{
				ID: id(uuid.NewString()),
				item: &db.Author{
					Hash:      crypt.Hash(firstName + lastName),
					FirstName: firstName,
					LastName:  lastName,
				},
			}
		} else {
			author = i.getAuthorWithName(firstName, lastName)
		}

		authors = append(authors, author)
	}

	return authors
}

func (i *sprdImporter) loadWriters(storyID id, r row) {
	writers := i.loadAuthorColumn(r, "story_written_by")
	for _, writer := range writers {
		writer.item.IsWriter = true
		i.setWriterForStory(storyID, writer.ID)
	}
}

func (i *sprdImporter) loadDrawer(storyID id, r row) {
	drawers := i.loadAuthorColumn(r, "story_drawn_by")
	for _, drawer := range drawers {
		drawer.item.IsDrawer = true
		i.setDrawerForStory(storyID, drawer.ID)
	}
}

func (i *sprdImporter) loadInventor(storyID id, r row) {
	inventors := i.loadAuthorColumn(r, "story_invented_by")
	for _, inventor := range inventors {
		inventor.item.IsInventor = true
		i.setDrawerForStory(storyID, inventor.ID)
	}
}

func (i *sprdImporter) parseIssueNum(val string) (int, error) {
	parts := strings.Split(val, "(")
	val = parts[0]
	parts = strings.Split(val, "/")
	val = parts[0]
	return strconv.Atoi(val)
}

func (i *sprdImporter) hasPublicationWithHash(hash string) bool {
	return slices.IndexFunc(i.publications, func(p *importerPublication) bool {
		return p.item.Hash == hash
	}) != -1
}

func (i *sprdImporter) loadBasePublication(storyID id, r row) error {
	year, err := strconv.Atoi(r.getValue("pub_year"))
	if err != nil {
		return err
	}
	from, err := i.parseIssueNum(r.getValue("pub_from"))
	if err != nil {
		return err
	}
	to, err := i.parseIssueNum(r.getValue("pub_to"))
	if err != nil {
		return err
	}

	if year == 0 {
		return nil
	}
	if from == 0 {
		from = to
	}
	if to == 0 {
		to = from
	}

	for _, issue := range getIssuesBetween(from, to, year) {
		pubType := "perus"
		pub := &db.Publication{
			Hash:  crypt.Hash(pubType + string(issue["year"]) + string(issue["num"])),
			Type:  pubType,
			Year:  year,
			Issue: string(issue["num"]),
		}
		if !i.hasPublicationWithHash(pub.Hash) {
			importerPublication := &importerPublication{
				ID:   id(uuid.NewString()),
				item: pub,
			}
			i.publications = append(i.publications, importerPublication)
			// TODO: add storypublication
		}
	}

	return nil
}

func getPublishedAnnualCount(year int) int {
	if year == 1953 {
		return 25
	}
	if year == 1954 || year == 1965 {
		return 27
	}
	if year >= 1955 && year <= 1964 {
		return 26
	}
	if year == 1971 || year == 1972 || (year >= 1974 && year <= 1978) {
		return 12
	}
	if year == 1973 {
		return 11
	}
	if year == 1979 {
		return 13
	}
	if year >= 1980 {
		return 16
	}
	return 0

}

func getIssuesBetween(from int, to int, year int) []map[string]int {
	issues := []map[string]int{}

	annualCount := getPublishedAnnualCount(year)
	upTo := to
	if to < from {
		upTo = annualCount + to
	}

	for i := from; i <= upTo; i++ {
		y := year
		num := i
		if i > annualCount {
			y = year + 1
			num = i % annualCount
		}
		issues = append(issues, map[string]int{
			"year": y,
			"num":  num,
		})
	}

	return issues
}

func (i *sprdImporter) loadBaseRePublication(storyID id, r row) {}

func (i *sprdImporter) loadItalianBasePublication(storyID id, r row) {}

func (i *sprdImporter) loadSpecialPublication(storyID id, r row) {}

func (i *sprdImporter) loadItalianSpecialPublication(storyID id, r row) {}

func (i *sprdImporter) loadKronikka(storyID id, r row) {}

func (i *sprdImporter) loadKirjasto(storyID id, r row) {}

func (i *sprdImporter) loadVillain(storyID id, r row) {}

func (i *sprdImporter) SaveData() error {
	fmt.Println(i)

	versionRepo := db.NewVersionRepository()
	version, err := versionRepo.Create(db.Version{IsActive: false})
	if err != nil {
		return err
	}

	// Save authors

	// Save stories

	fmt.Println(version)
	return nil
}
