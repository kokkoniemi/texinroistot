package main

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	_ "github.com/joho/godotenv/autoload"
	"github.com/kokkoniemi/texinroistot/internal/crypt"
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
	var stories []*db.Story
	var publications []*db.Publication

	for i, row := range rows[1:] {
		// Todo: remove after the function is ready
		if i > 100 {
			break
		}

		story, err := createStory(row)
		if err != nil {
			return err
		}

		// create authors:
		authors, story = handleAuthors(authors, story, row, "story_written_by")
		authors, story = handleAuthors(authors, story, row, "story_drawn_by")
		authors, story = handleAuthors(authors, story, row, "story_invented_by")

		// create villains & attach them to stories

		// create publications
		publications, story, err = handleBasePublications(publications, story, row, "pub_year", "pub_from", "pub_to")
		if err != nil {
			return err
		}
		publications, story, err = handleBasePublications(publications, story, row, "repub_year", "repub_from", "repub_to")
		if err != nil {
			return err
		}
		publications, story, err = handleBasePublications(publications, story, row, "italy_year", "italy_pub_from", "italy_pub_to")
		if err != nil {
			return err
		}

		storyIdx := slices.IndexFunc(stories, func(s *db.Story) bool {
			return s.Hash == story.Hash
		})
		if storyIdx == -1 {
			stories = append(stories, story)
		}
	}

	authorRepo := db.NewAuthorRepository()
	authors, err = authorRepo.BulkCreate(authors, *version)
	if err != nil {
		return err
	}

	// attach id's for authors
	for _, story := range stories {
		if story.WrittenBy != nil {
			story.WrittenBy = findMatchingAuthor(authors, story.WrittenBy)
		}
		if story.DrawnBy != nil {
			story.DrawnBy = findMatchingAuthor(authors, story.DrawnBy)
		}
		if story.InventedBy != nil {
			story.InventedBy = findMatchingAuthor(authors, story.InventedBy)
		}
	}

	storyRepo := db.NewStoryRepository()
	_, err = storyRepo.BulkCreate(stories, *version)
	if err != nil {
		return err
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

func getPublicationType(key string) string {
	if key == "pub_year" || key == "repub_year" {
		return "perus"
	}
	if key == "italy_year" {
		return "italia_perus"
	}
	return ""
}

func parseIssueNum(key string) (int, error) {
	parts := strings.Split(key, "(")
	key = parts[0]
	parts = strings.Split(key, "/")
	key = parts[len(parts)-1]
	return strconv.Atoi(key)
}

func handleBasePublications(publications []*db.Publication, story *db.Story, row []string, ykey string, fkey string, tokey string) ([]*db.Publication, *db.Story, error) {
	year, err := strconv.Atoi(getValue(row, ykey))
	if err != nil {
		return nil, nil, err
	}
	from, err := parseIssueNum(getValue(row, fkey))
	if err != nil {
		return nil, nil, err
	}
	to, err := parseIssueNum(getValue(row, tokey))
	if err != nil {
		return nil, nil, err
	}

	if year == 0 {
		return publications, story, nil
	}
	if from == 0 {
		from = to
	}
	if to == 0 {
		to = from
	}

	type issue struct {
		year int
		num  int
	}
	var issues []issue

	upTo := to
	if to < from {
		upTo = getPublishedAnnualCount(year)
	}

	// fill issues
	for i := from; i <= upTo; i++ {
		issues = append(issues, issue{year: year, num: i})
	}

	if to < from {
		for i := 1; i <= to; i++ {
			issues = append(issues, issue{year: year + 1, num: i})
		}
	}

	for _, iss := range issues {
		pubType := getPublicationType(ykey)
		pub := &db.Publication{
			Hash:  crypt.Hash(pubType + string(iss.year) + string(iss.num)),
			Type:  pubType,
			Year:  year,
			Issue: string(iss.num),
		}
		pubIndex := slices.IndexFunc(publications, func(p *db.Publication) bool {
			return p.Hash == pub.Hash
		})
		if pubIndex == -1 {
			publications = append(publications, pub)
		}
	}

	// attach story to publication

	return publications, story, nil
}

func findMatchingAuthor(authors []*db.Author, author *db.Author) *db.Author {
	authorIdx := slices.IndexFunc(authors, func(a *db.Author) bool {
		return a.FirstName == author.FirstName && a.LastName == author.LastName
	})
	return authors[authorIdx]
}

func createStory(row []string) (*db.Story, error) {
	orderNum, err := strconv.Atoi(getValue(row, "story_order_num"))
	if err != nil {
		return nil, err
	}

	hash := ""
	if orderNum != 0 {
		hash = crypt.Hash(string(orderNum))
	} else {
		hash = crypt.Hash(getValue(row, "story_title"))
	}

	return &db.Story{
		Hash:        hash,
		OrderNumber: orderNum,
	}, nil
}

func handleAuthors(authors []*db.Author, story *db.Story, row []string, key string) ([]*db.Author, *db.Story) {
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
				Hash:      crypt.Hash(firstName + lastName),
				FirstName: firstName,
				LastName:  lastName,
			}
			authors = append(authors, author)
		} else {
			author = authors[authorIdx]
		}

		if key == "story_written_by" {
			author.IsWriter = true
			story.WrittenBy = author
		} else if key == "story_drawn_by" {
			author.IsDrawer = true
			story.DrawnBy = author
		} else if key == "story_invented_by" {
			author.IsInventor = true
			story.InventedBy = author
		}
	}

	return authors, story
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
