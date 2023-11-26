package importer

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/kokkoniemi/texinroistot/internal/crypt"
	"github.com/kokkoniemi/texinroistot/internal/db"
)

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

func (i *importer) addPublication(pub *db.Publication) *importerPublication {
	i.totalEntities++

	importerPublication := &importerPublication{
		ID:   id(i.totalEntities),
		item: pub,
	}
	i.publications = append(i.publications, importerPublication)

	return importerPublication
}

func (i *importer) addStoryPublication(storyID id, pubID id, title string) *importerStoryPublication {
	i.totalEntities++

	importerStoryPublication := &importerStoryPublication{
		ID:          id(i.totalEntities),
		story:       storyID,
		publication: pubID,
		title:       title,
	}
	i.storyPublications = append(i.storyPublications, importerStoryPublication)

	return importerStoryPublication
}

func (i *importer) hasStoryPublication(storyID id, pubID id) bool {
	return slices.IndexFunc(i.storyPublications, func(sp *importerStoryPublication) bool {
		return sp.story == storyID && sp.publication == pubID
	}) != -1
}

func (i *importer) importBasePublication(storyID id, r row) error {
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

	titles := strings.Split(r.getValue("story_title"), ";")
	if len(titles) == 0 {
		return fmt.Errorf("title is missing")
	}

	for _, issue := range getIssuesBetween(from, to, year) {
		pubType := "perus"
		pub := &db.Publication{
			Hash:  crypt.Hash(fmt.Sprintf("%s%v%v", pubType, issue["year"], issue["num"])),
			Type:  pubType,
			Year:  year,
			Issue: fmt.Sprintf("%v", issue["num"]),
		}
		if !i.hasPublicationWithHash(pub.Hash) {
			importerPublication := i.addPublication(pub)
			if !i.hasStoryPublication(storyID, importerPublication.ID) {
				i.addStoryPublication(storyID, importerPublication.ID, titles[0])
			}
		}
	}

	return nil
}

func (i *importer) importBaseRePublication(storyID id, r row) {}

func (i *importer) importItalianBasePublication(storyID id, r row) {}

func (i *importer) importSpecialPublication(storyID id, r row) {}

func (i *importer) importItalianSpecialPublication(storyID id, r row) {}

func (i *importer) importKronikka(storyID id, r row) {}

func (i *importer) importKirjasto(storyID id, r row) {}

func (i *importer) parseIssueNum(val string) (int, error) {
	parts := strings.Split(val, "(")
	val = parts[0]
	parts = strings.Split(val, "/")
	val = parts[0]
	return strconv.Atoi(val)
}

func (i *importer) hasPublicationWithHash(hash string) bool {
	return slices.IndexFunc(i.publications, func(p *importerPublication) bool {
		return p.item.Hash == hash
	}) != -1
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
