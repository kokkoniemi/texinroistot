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
	PUB_IT_SERIE_EXTRA          = "italia_serie_extra"
	PUB_IT_TEXONE               = "italia_texone"
	PUB_IT_MINI_TEXONE_MAXI_TEX = "italia_mini_texone_maxi_tex"
	PUB_IT_ALMANACCO_DEL_WEST   = "italia_almanacco_del_west"
	PUB_IT_COLOR_TEX            = "italia_color_tex"
	PUB_IT_TEX_ROMANZI          = "italia_tex_romanzi_a_fumetti"
	PUB_IT_TEX_MAGAZINE         = "italia_tex_magazine"
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

func (i *importer) getPublicationIndexWithHash(hash string) int {
	return slices.IndexFunc(i.publications, func(p *importerPublication) bool {
		return p.item.Hash == hash
	})
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

func (i *importer) getStoryPublications(storyID id) []*importerStoryPublication {
	var filtered []*importerStoryPublication
	for idx := range i.storyPublications {
		if i.storyPublications[idx].story == storyID {
			filtered = append(filtered, i.storyPublications[idx])
		}
	}
	return filtered
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

	year, err := strconv.Atoi(strings.TrimSpace(yearVal))
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

func (i *importer) loadBasePublication(storyID id, r row) error {
	return i.handleBasePublications(
		storyID, r, PUB_PERUS, "story_title", 0, "pub_year", "pub_from",
		"pub_to")
}

func (i *importer) loadBaseRePublication(storyID id, r row) error {
	return i.handleBasePublications(
		storyID, r, PUB_PERUS, "story_title", 1, "repub_year", "repub_from",
		"repub_to")
}

func (i *importer) loadItalianBasePublication(storyID id, r row) error {
	return i.handleBasePublications(
		storyID, r, PUB_IT_PERUS, "italy_story_title", 0, "italy_year",
		"italy_pub_from", "italy_pub_to")
}

// parseNonBaseTitle parses the title for publications other than PUB_PERUS, PUB_IT_PERUS, PUB_IT_ERIK
func (i *importer) parseNonBaseTitle(pubType string, r row) (string, error) {
	titles := strings.Split(r.getValue("story_title"), ";")
	if len(titles) == 0 {
		return "", fmt.Errorf("Could not find title")
	}

	index := 0
	incrementIndex := func(fields ...string) {
		for _, field := range fields {
			if len(strings.TrimSpace(r.getValue(field))) > 0 {
				index++
			}
		}
	}

	if pubType == PUB_MAXI || pubType == PUB_SUUR || pubType == PUB_MUU {
		incrementIndex("pub_from", "repub_from")

	} else if pubType == PUB_KRONIKKA {
		incrementIndex("pub_from", "repub_from", "pub_special")
	} else if pubType == PUB_KIRJASTO {
		incrementIndex("pub_from", "repub_from", "pub_special", "pub_kronikka")
	}

	if index < len(titles) {
		return titles[index], nil
	}
	return titles[0], nil
}

func (i *importer) loadSpecialPublication(storyID id, r row) error {
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

	title, err := i.parseNonBaseTitle(pubType, r)
	if err != nil {
		return err
	}

	if !i.hasPublicationWithHash(pub.Hash) {
		importerPublication := i.addPublication(pub)
		if !i.hasStoryPublication(storyID, importerPublication.ID) {
			i.addStoryPublication(storyID, importerPublication.ID, title)
		}
	}

	return nil
}

func (i *importer) loadItalianSpecialPublication(storyID id, r row) error {
	val := strings.TrimSpace(r.getValue("italy_pub_special"))
	if len(val) == 0 {
		return nil
	}

	pubType := italianSpecialPublicationType(val)
	pub := &db.Publication{
		Hash:  crypt.Hash(fmt.Sprintf("%s%s", pubType, val)),
		Type:  pubType,
		Issue: val,
	}

	titles := strings.Split(r.getValue("italy_story_title"), ";")
	if len(titles) == 0 {
		return fmt.Errorf("title is missing")
	}
	title := strings.TrimSpace(titles[0])
	if len(titles) >= 2 {
		title = strings.TrimSpace(titles[1])
	}

	if !i.hasPublicationWithHash(pub.Hash) {
		importerPublication := i.addPublication(pub)
		if !i.hasStoryPublication(storyID, importerPublication.ID) {
			i.addStoryPublication(storyID, importerPublication.ID, title)
		}
	}

	return nil
}

func italianSpecialPublicationType(value string) string {
	normalize := strings.NewReplacer("&", " ", "-", " ", "_", " ", "/", " ")
	normalized := strings.Join(strings.Fields(normalize.Replace(strings.ToLower(strings.TrimSpace(value)))), " ")

	switch {
	case strings.Contains(normalized, "mini texone"), strings.Contains(normalized, "maxi tex"):
		return PUB_IT_MINI_TEXONE_MAXI_TEX
	case strings.Contains(normalized, "serie extra"):
		return PUB_IT_SERIE_EXTRA
	case strings.Contains(normalized, "texone"):
		return PUB_IT_TEXONE
	case strings.Contains(normalized, "almanacco del west"):
		return PUB_IT_ALMANACCO_DEL_WEST
	case strings.Contains(normalized, "color tex"):
		return PUB_IT_COLOR_TEX
	case strings.Contains(normalized, "romanzi a fumetti"):
		return PUB_IT_TEX_ROMANZI
	case strings.Contains(normalized, "tex magazine"):
		return PUB_IT_TEX_MAGAZINE
	default:
		return PUB_IT_ERIK
	}
}

func (i *importer) loadKronikka(storyID id, r row) error {
	val := strings.TrimSpace(r.getValue("pub_kronikka"))
	if len(val) == 0 {
		return nil
	}

	pub := &db.Publication{
		Hash:  crypt.Hash(fmt.Sprintf("%s%s", PUB_KRONIKKA, val)),
		Type:  PUB_KRONIKKA,
		Issue: val,
	}
	title, err := i.parseNonBaseTitle(PUB_KRONIKKA, r)
	if err != nil {
		return err
	}

	if !i.hasPublicationWithHash(pub.Hash) {
		importerPublication := i.addPublication(pub)
		if !i.hasStoryPublication(storyID, importerPublication.ID) {
			i.addStoryPublication(storyID, importerPublication.ID, title)
		}
	}

	return nil
}

func (i *importer) loadKirjasto(storyID id, r row) error {
	val := strings.TrimSpace(r.getValue("pub_kirjasto"))
	if len(val) == 0 {
		return nil
	}

	pub := &db.Publication{
		Hash:  crypt.Hash(fmt.Sprintf("%s%s", PUB_KIRJASTO, val)),
		Type:  PUB_KIRJASTO,
		Issue: val,
	}
	title, err := i.parseNonBaseTitle(PUB_KRONIKKA, r)
	if err != nil {
		return err
	}

	if !i.hasPublicationWithHash(pub.Hash) {
		importerPublication := i.addPublication(pub)
		if !i.hasStoryPublication(storyID, importerPublication.ID) {
			i.addStoryPublication(storyID, importerPublication.ID, title)
		}
	}

	return nil
}

func (i *importer) parseIssueNum(val string) (int, error) {
	parts := strings.Split(val, "(")
	val = parts[0]
	parts = strings.Split(val, "/")
	val = parts[0]
	return strconv.Atoi(strings.TrimSpace(val))
}

func (i *importer) hasPublicationWithHash(hash string) bool {
	return slices.IndexFunc(i.publications, func(p *importerPublication) bool {
		return p.item.Hash == hash
	}) != -1
}

func (i *importer) getPublicationWithID(pubID id) *importerPublication {
	for _, p := range i.publications {
		if p.ID == pubID {
			return p
		}
	}
	return nil
}

func (i *importer) getPublicationItems() []*db.Publication {
	var items []*db.Publication

	for index := range i.publications {
		items = append(items, i.publications[index].item)
	}

	return items
}

// setPublicationItems sets persisted Publications to importer after save to db
func (i *importer) setPublicationItems(items []*db.Publication) error {
	if len(items) != len(i.publications)%db.MaxBulkCreateSize && len(items) != db.MaxBulkCreateSize {
		fmt.Println(len(items), db.MaxBulkCreateSize, len(i.publications))
		return fmt.Errorf("Mismatch in the number of Publications")
	}

	for index := range items {
		importerIndex := i.getPublicationIndexWithHash(items[index].Hash)
		if importerIndex == -1 {
			return fmt.Errorf("Tried to set unknown Publication")
		}
		i.publications[importerIndex].item = items[index]
	}

	return nil
}

// persistPublications writes Publications loaded in importer to db
func (i *importer) persistPublications(version *db.Version) error {
	var err error
	storyRepo := db.NewStoryRepository()
	chunks := ChunkSlice(i.getPublicationItems(), db.MaxBulkCreateSize)
	for _, chunk := range chunks {
		publications, err := storyRepo.BulkCreatePublications(chunk, version)
		if err != nil {
			return err
		}
		err = i.setPublicationItems(publications)
		if err != nil {
			return err
		}
	}
	return err
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
	if year >= 1971 || year <= 1978 {
		return 12
	}
	if year == 1979 {
		return 13
	}
	if year >= 1980 {
		return 16
	}
	return -1

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
