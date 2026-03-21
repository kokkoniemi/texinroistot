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

func NewSpreadsheetImporter(titleRow []string) (*importer, error) {
	columnNames := defaultColumns
	columnIndexes := map[string]int{}

	for index, title := range titleRow {
		key, ok := columnNames[strings.TrimSpace(title)]
		if !ok {
			continue
		}
		columnIndexes[key] = index
	}

	var missingColumns []string
	for _, key := range requiredColumnKeys {
		if _, ok := columnIndexes[key]; !ok {
			missingColumns = append(missingColumns, key)
		}
	}
	if len(missingColumns) > 0 {
		return nil, fmt.Errorf("missing required columns: %s", strings.Join(missingColumns, ", "))
	}

	return &importer{
		columnNames:   defaultColumns,
		columnIndexes: columnIndexes,
		totalEntities: 0,
	}, nil
}

func (i *importer) LoadData(dataRows [][]string) error {
	for index, dataRow := range dataRows {
		row := row{importer: i, cells: dataRow, index: index}

		storyID, err := i.loadStory(row)
		if err != nil {
			return err
		}
		i.loadWriters(storyID, row)
		i.loadDrawers(storyID, row)
		i.loadTranslators(storyID, row)
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
		err = i.loadVillain(storyID, row)
		if err != nil {
			return err
		}
	}

	return nil
}

func (i *importer) PersistData() error {
	_, err := i.PersistDataWithVersion()
	return err
}

func (i *importer) PersistDataWithVersion() (*db.Version, error) {
	versionRepo := db.NewVersionRepository()
	version, err := versionRepo.Create(db.Version{IsActive: false})
	if err != nil {
		return nil, err
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
		return nil, err
	}

	err = i.persistPublications(version)
	if err != nil {
		return nil, err
	}
	err = i.persistStories(version)
	if err != nil {
		return nil, err
	}
	err = i.persistVillains(version)
	if err != nil {
		return nil, err
	}

	return version, nil
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
	"Arvo":                                                     "ranks",
	"Etunimi (sisältää nimet, joita käytetään kuin etunimeä)": "first_names",
	"Sukunimi (sisältää nimet, joita käytetään kuin sukunimeä)": "last_name",
	"Nimi (muut kuin etunimi-sukunimi-tyyppiset nimet)":         "other_names",
	"Lempinimi":                                                  "nicknames",
	"Salanimi":                                                   "code_names",
	"Rooli":                                                      "roles",
	"Kohtalo":                                                    "destiny",
	"Tarina":                                                     "story_title",
	"Kertoi":                                                     "story_written_by",
	"Piirsi":                                                     "story_drawn_by",
	"Suomensi":                                                   "story_translated_by",
	"Vuosi":                                                      "pub_year",
	"Alkaen":                                                     "pub_from",
	"Päättyen":                                                   "pub_to",
	"UVuosi":                                                     "repub_year",
	"Ualkaen":                                                    "repub_from",
	"Upäättyen":                                                  "repub_to",
	"Erikoisjulkaisu":                                            "pub_special",
	"Kronikka":                                                   "pub_kronikka",
	"Kirjasto":                                                   "pub_kirjasto",
	"Italian vuosi":                                              "italy_year",
	"Italian alkunumero":                                         "italy_pub_from",
	"Italian päättymisnumero":                                    "italy_pub_to",
	"Italian erikoisjulkaisu":                                    "italy_pub_special",
	"Italian tarina":                                             "italy_story_title",
	"Järjestysluku":                                              "story_order_num",
	"Sama numero, sama roisto":                                   "villain_id",
}

var requiredColumnKeys = []string{
	"ranks",
	"first_names",
	"last_name",
	"other_names",
	"nicknames",
	"code_names",
	"roles",
	"destiny",
	"story_title",
	"story_written_by",
	"story_drawn_by",
	"story_translated_by",
	"pub_year",
	"pub_from",
	"pub_to",
	"repub_year",
	"repub_from",
	"repub_to",
	"pub_special",
	"pub_kronikka",
	"pub_kirjasto",
	"italy_year",
	"italy_pub_from",
	"italy_pub_to",
	"italy_pub_special",
	"italy_story_title",
	"story_order_num",
	"villain_id",
}
