package importer

import (
	"fmt"
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

		storyID, err := i.importStory(row)
		if err != nil {
			return err
		}
		i.importWriters(storyID, row)
		i.importDrawer(storyID, row)
		i.importInventor(storyID, row)
		err = i.importBasePublication(storyID, row)
		if err != nil {
			return err
		}
		i.importBaseRePublication(storyID, row)
		i.importItalianBasePublication(storyID, row)
		i.importSpecialPublication(storyID, row)
		i.importItalianSpecialPublication(storyID, row)
		i.importKronikka(storyID, row)
		i.importKirjasto(storyID, row)
		i.importVillain(storyID, row)
	}

	return nil
}

func (i *importer) SaveData() error {
	fmt.Println(i)

	// versionRepo := db.NewVersionRepository()
	// version, err := versionRepo.Create(db.Version{IsActive: false})
	// if err != nil {
	// 	return err
	// }

	// Save authors

	// Save stories

	//fmt.Println(version)
	return nil
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
	if len(r.cells) < index+1 {
		fmt.Sprintln("no value for key '%s' in row '%v'", key, r.index)
		return ""
	}

	return r.cells[index]
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
