package main

import (
	"fmt"
	"slices"
	"strings"

	_ "github.com/joho/godotenv/autoload"
	"github.com/kokkoniemi/texinroistot/internal/db"
	"github.com/xuri/excelize/v2"
)

func main() {
	err := parseExcel()
	if err != nil {
		panic(err)
	}
}

func parseExcel() error {
	f, err := excelize.OpenFile("Texinroistot.xlsx")

	if err != nil {
		return err
	}

	defer closeExcel(f)

	rows, err := f.GetRows("Taul1")
	if err != nil {
		return err
	}

	if len(rows) <= 0 {
		return fmt.Errorf("no content")
	}

	for index, title := range rows[0] {
		key := colTypes[title]
		colIndexes[key] = index
	}

	// create version
	versionRepo := db.NewVersionRepository()
	version, err := versionRepo.Create(db.Version{IsActive: false})

	if err != nil {
		return err
	}

	var authors []*db.Author

	for i, row := range rows[1:] {
		// Todo: remove after the function is ready
		if i > 10 {
			break
		}

		// create authors:
		authors = addAuthors(authors, row, "story_written_by")
		authors = addAuthors(authors, row, "story_drawn_by")
		authors = addAuthors(authors, row, "story_invented_by")

		// create stories:
		//

		// create villains & attach them to stories

		// create publications & attach stories to them

	}

	authorRepo := db.NewAuthorRepository()
	_, err = authorRepo.BulkCreate(authors, *version)
	if err != nil {
		return err
	}

	return nil
}

func addAuthors(authors []*db.Author, row []string, key string) []*db.Author {
	names := strings.Split(getValue(row, key), ";")

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
		var author *db.Author
		authorIdx := slices.IndexFunc(authors, func(a *db.Author) bool {
			return a.FirstName == firstName && a.LastName == lastName
		})
		if authorIdx == -1 {
			author = &db.Author{
				FirstName: firstName,
				LastName:  lastName,
			}
			authors = append(authors, author)
		} else {
			author = authors[authorIdx]
		}

		if key == "story_written_by" {
			author.IsWriter = true
		} else if key == "story_drawn_by" {
			author.IsDrawer = true
		} else if key == "story_invented_by" {
			author.IsInventor = true
		}
	}
	return authors
}

func getValue(row []string, key string) string {
	return row[colIndexes[key]]
}

var colIndexes = map[string]int{}
var colTypes = map[string]string{
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

func closeExcel(f *excelize.File) error {
	err := f.Close()
	if err != nil {
		return err
	}
	return nil
}
