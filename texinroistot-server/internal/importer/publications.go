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

const (
	PUB_PERUS    = "perus"
	PUB_MAXI     = "maxi"
	PUB_SUUR     = "suur"
	PUB_MUU      = "muu_erikois"
	PUB_KRONIKKA = "kronikka"
	PUB_KIRJASTO = "kirjasto"
	PUB_IT_PERUS = "italia_perus"
	PUB_IT_ERIK  = "italia_erikois"
)

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

func (i *importer) handleBasePublications(
	storyID id,
	r row,
	pubType string,
	titleCol string,
	titleIndex int,
	yearCol string,
	fromCol string,
	toCol string,
) error {
	yearVal := r.getValue(yearCol)
	fromVal := r.getValue(fromCol)
	toVal := r.getValue(toCol)

	if len(fromVal) == 0 || len(toVal) == 0 || len(yearVal) == 0 {
		return nil
	}

	year, err := strconv.Atoi(yearVal)
	if err != nil {
		return err
	}
	from, err := i.parseIssueNum(fromVal)
	if err != nil {
		return err
	}
	to, err := i.parseIssueNum(toVal)
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

	titles := strings.Split(r.getValue(titleCol), ";")
	if len(titles) == 0 {
		return fmt.Errorf("title is missing")
	}
	title := titles[0]
	if len(titles) >= titleIndex+1 {
		title = titles[titleIndex]
	}

	for _, issue := range getIssuesBetween(from, to, year) {
		pub := &db.Publication{
			Hash:  crypt.Hash(fmt.Sprintf("%s%v%v", pubType, issue["year"], issue["num"])),
			Type:  pubType,
			Year:  year,
			Issue: fmt.Sprintf("%v", issue["num"]),
		}
		if !i.hasPublicationWithHash(pub.Hash) {
			importerPublication := i.addPublication(pub)
			if !i.hasStoryPublication(storyID, importerPublication.ID) {
				i.addStoryPublication(storyID, importerPublication.ID, title)
			}
		}
	}

	return nil
}

func (i *importer) importBasePublication(storyID id, r row) error {
	return i.handleBasePublications(
		storyID, r, PUB_PERUS, "story_title", 0, "pub_year", "pub_from",
		"pub_to")
}

func (i *importer) importBaseRePublication(storyID id, r row) error {
	return i.handleBasePublications(
		storyID, r, PUB_PERUS, "story_title", 1, "repub_year", "repub_from",
		"repub_to")
}

func (i *importer) importItalianBasePublication(storyID id, r row) error {
	return i.handleBasePublications(
		storyID, r, PUB_IT_PERUS, "italy_story_title", 0, "italy_year",
		"italy_pub_from", "italy_pub_to")
}

func (i *importer) importSpecialPublication(storyID id, r row) error {
	val := strings.TrimSpace(r.getValue("pub_special"))
	if len(val) == 0 {
		return nil
	}

	pubType := PUB_MUU
	if strings.Contains(strings.ToLower(val), "suuralbumi") {
		pubType = PUB_SUUR
	} else if strings.Contains(strings.ToLower(val), "maxi-tex") {
		pubType = PUB_MAXI
	}

	pub := &db.Publication{
		Hash:  crypt.Hash(fmt.Sprintf("%s%s", pubType, val)),
		Type:  pubType,
		Issue: val,
	}

	titles := strings.Split(r.getValue("story_title"), ";")
	if len(titles) == 0 {
		return fmt.Errorf("title is missing")
	}
	title := titles[0]
	if len(titles) >= 3 {
		title = titles[2]
	}

	if !i.hasPublicationWithHash(pub.Hash) {
		importerPublication := i.addPublication(pub)
		if !i.hasStoryPublication(storyID, importerPublication.ID) {
			i.addStoryPublication(storyID, importerPublication.ID, title)
		}
	}

	return nil
}

func (i *importer) importItalianSpecialPublication(storyID id, r row) error {
	val := strings.TrimSpace(r.getValue("italy_pub_special"))
	if len(val) == 0 {
		return nil
	}

	pub := &db.Publication{
		Hash:  crypt.Hash(fmt.Sprintf("%s%s", PUB_IT_ERIK, val)),
		Type:  PUB_IT_ERIK,
		Issue: val,
	}

	titles := strings.Split(r.getValue("italy_story_title"), ";")
	if len(titles) == 0 {
		return fmt.Errorf("title is missing")
	}
	title := titles[0]
	if len(titles) >= 2 {
		title = titles[1]
	}

	if !i.hasPublicationWithHash(pub.Hash) {
		importerPublication := i.addPublication(pub)
		if !i.hasStoryPublication(storyID, importerPublication.ID) {
			i.addStoryPublication(storyID, importerPublication.ID, title)
		}
	}

	return nil
}

func (i *importer) importKronikka(storyID id, r row) error {
	val := strings.TrimSpace(r.getValue("pub_kronikka"))
	if len(val) == 0 {
		return nil
	}

	pub := &db.Publication{
		Hash:  crypt.Hash(fmt.Sprintf("%s%s", PUB_KRONIKKA, val)),
		Type:  PUB_KRONIKKA,
		Issue: val,
	}

	titles := strings.Split(r.getValue("story_title"), ";")
	if len(titles) == 0 {
		return fmt.Errorf("title is missing")
	}
	title := titles[0]
	if len(titles) >= 4 {
		title = titles[3]
	}

	if !i.hasPublicationWithHash(pub.Hash) {
		importerPublication := i.addPublication(pub)
		if !i.hasStoryPublication(storyID, importerPublication.ID) {
			i.addStoryPublication(storyID, importerPublication.ID, title)
		}
	}

	return nil
}

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
	return 99999

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
