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

// ensureStoryPublication ensures pub exists and that storyID is linked to it.
// If a publication with the same hash already exists, it is reused; either way
// the story-publication link is created when missing.
func (i *importer) ensureStoryPublication(storyID id, pub *db.Publication, title string) {
	idx := i.getPublicationIndexWithHash(pub.Hash)
	var pubID id
	if idx == -1 {
		pubID = i.addPublication(pub).ID
	} else {
		pubID = i.publications[idx].ID
	}
	if !i.hasStoryPublication(storyID, pubID) {
		i.addStoryPublication(storyID, pubID, title)
	}
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
	wrapAnnually bool,
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
	title := strings.TrimSpace(titles[0])
	if len(titles) >= titleIndex+1 {
		title = strings.TrimSpace(titles[titleIndex])
	}

	var issues []map[string]int
	if wrapAnnually {
		wrap, err := detectWrap(toVal, from, to, year)
		if err != nil {
			return err
		}
		issues, err = getIssuesBetween(from, to, year, wrap)
		if err != nil {
			return err
		}
		issues, err = applyExclusion(issues, toVal, year)
		if err != nil {
			return err
		}
	} else {
		if to < from {
			return fmt.Errorf("italian publication range invalid: from=%d > to=%d", from, to)
		}
		issues = italianStyleIssues(from, to, year)
	}

	for _, issue := range issues {
		pub := &db.Publication{
			Hash:  crypt.Hash(fmt.Sprintf("%s%v%v", pubType, issue["year"], issue["num"])),
			Type:  pubType,
			Year:  issue["year"],
			Issue: fmt.Sprintf("%v", issue["num"]),
		}
		i.ensureStoryPublication(storyID, pub, title)
	}

	return nil
}

func (i *importer) loadBasePublication(storyID id, r row) error {
	return i.handleBasePublications(
		storyID, r, PUB_PERUS, "story_title", 0, "pub_year", "pub_from",
		"pub_to", true)
}

func (i *importer) loadBaseRePublication(storyID id, r row) error {
	return i.handleBasePublications(
		storyID, r, PUB_PERUS, "story_title", 1, "repub_year", "repub_from",
		"repub_to", true)
}

func (i *importer) loadItalianBasePublication(storyID id, r row) error {
	return i.handleBasePublications(
		storyID, r, PUB_IT_PERUS, "italy_story_title", 0, "italy_year",
		"italy_pub_from", "italy_pub_to", false)
}

// parseNonBaseTitle parses the title for publications other than PUB_PERUS, PUB_IT_PERUS, PUB_IT_ERIK
func (i *importer) parseNonBaseTitle(pubType string, r row) (string, error) {
	titles := strings.Split(r.getValue("story_title"), ";")
	if len(titles) == 0 {
		return "", fmt.Errorf("Could not find title")
	}

	index := 0
	hasNonEmpty := func(field string) bool {
		return len(strings.TrimSpace(r.getValue(field))) > 0
	}
	// A base/repub publication only occupies a title slot when the importer
	// would actually create it — that requires the full (year, from, to) trio.
	if hasNonEmpty("pub_year") && hasNonEmpty("pub_from") && hasNonEmpty("pub_to") {
		index++
	}
	if hasNonEmpty("repub_year") && hasNonEmpty("repub_from") && hasNonEmpty("repub_to") {
		index++
	}

	if pubType == PUB_KRONIKKA || pubType == PUB_KIRJASTO {
		if hasNonEmpty("pub_special") {
			index++
		}
	}
	if pubType == PUB_KIRJASTO {
		if hasNonEmpty("pub_kronikka") {
			index++
		}
	}

	if index < len(titles) {
		return strings.TrimSpace(titles[index]), nil
	}
	return strings.TrimSpace(titles[0]), nil
}

func (i *importer) loadSpecialPublication(storyID id, r row) error {
	for _, val := range strings.Split(r.getValue("pub_special"), ";") {
		val = strings.TrimSpace(val)
		if len(val) == 0 {
			continue
		}

		pubType := PUB_MUU
		lower := strings.ToLower(val)
		if strings.Contains(lower, "suuralbumi") {
			pubType = PUB_SUUR
		} else if strings.Contains(lower, "maxi-tex") {
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

		i.ensureStoryPublication(storyID, pub, title)
	}

	return nil
}

func (i *importer) loadItalianSpecialPublication(storyID id, r row) error {
	titles := strings.Split(r.getValue("italy_story_title"), ";")
	title := ""
	if len(titles) > 0 {
		title = strings.TrimSpace(titles[0])
	}
	if len(titles) >= 2 {
		title = strings.TrimSpace(titles[1])
	}

	for _, val := range strings.Split(r.getValue("italy_pub_special"), ";") {
		val = strings.TrimSpace(val)
		if len(val) == 0 {
			continue
		}

		pubType := italianSpecialPublicationType(val)
		pub := &db.Publication{
			Hash:  crypt.Hash(fmt.Sprintf("%s%s", pubType, val)),
			Type:  pubType,
			Issue: val,
		}

		i.ensureStoryPublication(storyID, pub, title)
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

	i.ensureStoryPublication(storyID, pub, title)

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
	title, err := i.parseNonBaseTitle(PUB_KIRJASTO, r)
	if err != nil {
		return err
	}

	i.ensureStoryPublication(storyID, pub, title)

	return nil
}

func (i *importer) parseIssueNum(val string) (int, error) {
	parts := strings.Split(val, "(")
	val = parts[0]
	parts = strings.Split(val, "/")
	val = parts[0]
	return strconv.Atoi(strings.TrimSpace(val))
}

// detectWrap returns whether toVal indicates a year-boundary wrap. Wrap is
// detected either from `to < from` (legacy: "1" after "12" implies 1 of next
// year) or from an explicit "/yy" suffix on toVal (handles the case where the
// integer range alone would look in-year, e.g. "1" after "1" with /09).
// A "/yy" that resolves to a year other than outerYear or outerYear+1 is
// rejected — the importer does not support multi-year wraps.
func detectWrap(toVal string, from, to, outerYear int) (bool, error) {
	raw := strings.TrimSpace(strings.Split(toVal, "(")[0])
	parts := strings.Split(raw, "/")
	if len(parts) < 2 || strings.TrimSpace(parts[1]) == "" {
		return to < from, nil
	}
	yearStr := strings.TrimSpace(parts[1])
	parsed, err := strconv.Atoi(yearStr)
	if err != nil {
		return false, fmt.Errorf("invalid wrap year %q in %q: %w", yearStr, toVal, err)
	}
	matches := func(target int) bool {
		if parsed == target {
			return true
		}
		return len(yearStr) <= 2 && parsed == target%100
	}
	if matches(outerYear + 1) {
		return true, nil
	}
	if matches(outerYear) {
		return false, nil
	}
	return false, fmt.Errorf("publication wrap year mismatch: %q implies year %d, but outer year is %d", toVal, parsed, outerYear)
}

// applyExclusion removes issues matching an "(ei X-Y[/yy])" annotation in
// toVal. If no annotation is present the input slice is returned unchanged.
func applyExclusion(issues []map[string]int, toVal string, outerYear int) ([]map[string]int, error) {
	exFrom, exTo, exYear, has, err := parseExclusion(toVal, outerYear)
	if err != nil {
		return nil, err
	}
	if !has {
		return issues, nil
	}
	filtered := make([]map[string]int, 0, len(issues))
	for _, iss := range issues {
		if iss["year"] == exYear && iss["num"] >= exFrom && iss["num"] <= exTo {
			continue
		}
		filtered = append(filtered, iss)
	}
	return filtered, nil
}

// parseExclusion extracts an "(ei X-Y)" or "(ei X-Y/yy)" annotation. Returns
// has=false when no such annotation exists. When the annotation has no /yy
// suffix the exclusion year defaults to outerYear.
func parseExclusion(toVal string, outerYear int) (exFrom, exTo, exYear int, has bool, err error) {
	open := strings.Index(toVal, "(")
	if open == -1 {
		return 0, 0, 0, false, nil
	}
	rest := toVal[open+1:]
	close := strings.Index(rest, ")")
	if close == -1 {
		return 0, 0, 0, false, nil
	}
	annot := strings.TrimSpace(rest[:close])
	if !strings.HasPrefix(strings.ToLower(annot), "ei ") {
		return 0, 0, 0, false, nil
	}
	body := strings.TrimSpace(annot[3:])

	rangePart := body
	exYear = outerYear
	if slash := strings.Index(body, "/"); slash != -1 {
		rangePart = strings.TrimSpace(body[:slash])
		yStr := strings.TrimSpace(body[slash+1:])
		if yStr != "" {
			y, err := strconv.Atoi(yStr)
			if err != nil {
				return 0, 0, 0, false, fmt.Errorf("invalid ei year %q in %q: %w", yStr, toVal, err)
			}
			if len(yStr) <= 2 {
				// Anchor a 2-digit year near outerYear (±50yr window).
				century := (outerYear / 100) * 100
				y = century + y
				if y < outerYear-50 {
					y += 100
				} else if y > outerYear+50 {
					y -= 100
				}
			}
			exYear = y
		}
	}

	bounds := strings.Split(rangePart, "-")
	if len(bounds) != 2 {
		return 0, 0, 0, false, fmt.Errorf("invalid ei range %q in %q", rangePart, toVal)
	}
	exFrom, err = strconv.Atoi(strings.TrimSpace(bounds[0]))
	if err != nil {
		return 0, 0, 0, false, fmt.Errorf("invalid ei range start %q in %q: %w", bounds[0], toVal, err)
	}
	exTo, err = strconv.Atoi(strings.TrimSpace(bounds[1]))
	if err != nil {
		return 0, 0, 0, false, fmt.Errorf("invalid ei range end %q in %q: %w", bounds[1], toVal, err)
	}
	return exFrom, exTo, exYear, true, nil
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
	if year >= 1971 && year <= 1978 {
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

// italianStyleIssues iterates from..to inclusive without wrapping across years.
// Italian Tex numbers are a single ever-growing series; the year only records
// when the issue was printed.
func italianStyleIssues(from, to, year int) []map[string]int {
	out := make([]map[string]int, 0, to-from+1)
	for n := from; n <= to; n++ {
		out = append(out, map[string]int{"year": year, "num": n})
	}
	return out
}

func getIssuesBetween(from int, to int, year int, wrap bool) ([]map[string]int, error) {
	annualCount := getPublishedAnnualCount(year)
	if annualCount <= 0 {
		return nil, fmt.Errorf("no known annual publication count for year %d", year)
	}

	upTo := to
	if wrap {
		upTo = annualCount + to
	}

	issues := []map[string]int{}
	for i := from; i <= upTo; i++ {
		y := year
		num := i
		if i > annualCount {
			y = year + 1
			num = ((i - 1) % annualCount) + 1
		}
		issues = append(issues, map[string]int{
			"year": y,
			"num":  num,
		})
	}

	return issues, nil
}
