package importer

import (
	"fmt"
	"strings"

	"github.com/kokkoniemi/texinroistot/internal/db"
)

type importer struct {
	authors           []*importerAuthor
	stories           []*importerStory
	publications      []*importerPublication
	storyPublications []*importerStoryPublication
	villains          []*importerVillain
	storyVillains     []*importerStoryVillain
	columnIndexes     map[string]int
	columnNames       map[string]string
	totalEntities     uint64
}

func NewSpreadsheetImporter(titleRow []string) *importer {
	columnNames := defaultColumns
	columnIndexes := map[string]int{}

	for index, title := range titleRow {
		key := columnNames[title]
		columnIndexes[key] = index
	}

	return &importer{
		columnNames:   defaultColumns,
		columnIndexes: columnIndexes,
		totalEntities: 0,
	}
}

func (i *importer) LoadData(dataRows [][]string) error {
	for index, dataRow := range dataRows {
		if index > 1000 { // FIXME: remove after importer is ready
			break
		}
		row := row{importer: i, cells: dataRow, index: index}

		storyID, err := i.loadStory(row)
		if err != nil {
			return err
		}
		i.loadWriters(storyID, row)
		i.loadDrawers(storyID, row)
		i.loadInventors(storyID, row)
		err = i.loadBasePublication(storyID, row)
		if err != nil {
			return err
		}
		err = i.loadBaseRePublication(storyID, row)
		if err != nil {
			return err
		}
		err = i.loadItalianBasePublication(storyID, row)
		if err != nil {
			return err
		}
		err = i.loadSpecialPublication(storyID, row)
		if err != nil {
			return err
		}
		err = i.loadItalianSpecialPublication(storyID, row)
		if err != nil {
			return err
		}
		err = i.loadKronikka(storyID, row)
		if err != nil {
			return err
		}
		err = i.loadKirjasto(storyID, row)
		if err != nil {
			return err
		}
		i.importVillain(storyID, row)
	}

	return nil
}

func (i *importer) PersistData() error {
	fmt.Println(i)

	versionRepo := db.NewVersionRepository()
	version, err := versionRepo.Create(db.Version{IsActive: false})
	if err != nil {
		return err
	}

	// Save other models in following order:
	//  1. Author -> 2. Publication -> 3. Story -> 4. StoryPublication -> 5. Villain -> 6. StoryVillain
	//
	// Notes for step 3. and 4. Story:
	//      - Stories must be created in the db before StoryPublications
	// 	- Attach Author to Story (db column authors_in_stories)
	// Notes for step 5.
	//      - Attach villain to story

	err = i.persistAuthors(version)
	if err != nil {
		return err
	}

	err = i.persistPublications(version)
	if err != nil {
		return err
	}
	err = i.persistStories(version)
	if err != nil {
		return err
	}
	// i.persistStoryPublications(verion)
	// i.persistVillains(version)

	//fmt.Println(version)
	return nil
}

func (i *importer) TrimmedSplit(str string, delimiter string) []string {
	values := strings.Split(str, delimiter)
	for index := range values {
		values[index] = strings.TrimSpace(values[index])
	}
	return values
}

// helper function to chunk slices for appropriate size to be imported
func ChunkSlice[T any](items []T, size int) [][]T {
	var chunks [][]T
	for i := 0; i < len(items); i += size {
		end := i + size
		if end > len(items) {
			end = len(items)
		}
		chunks = append(chunks, items[i:end])
	}
	return chunks
}

type id uint64
type row struct {
	importer *importer
	index    int
	cells    []string
}

func (r row) getValue(key string) string {
	index, ok := r.importer.columnIndexes[key]
	if !ok {
		return ""
	}
	if len(r.cells) <= index {
		fmt.Printf("no value for key '%s' in row '%v'\n", key, r.index)
		return ""
	}

	return r.cells[index]
}

var defaultColumns = map[string]string{
	"Arvo":                     "ranks",
	"Etunimi":                  "first_names",
	"Sukunimi":                 "last_name",
	"Lempinimi/Intiaaninimi":   "nicknames",
	"Salanimi/Alias":           "aliases",
	"Rooli":                    "roles",
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
