package stories

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/kokkoniemi/texinroistot/internal/db"
)

const (
	defaultPublicationFilter = "perus_fi"
	defaultSort              = "fi_pub_date"
	defaultYear              = 0
	defaultPage              = 1
	defaultPageSize          = 25
	maxPageSize              = 100
)

var allowedPublicationFilters = map[string]bool{
	"all":                   true,
	"perus_fi":              true,
	"perus_it":              true,
	"suur":                  true,
	"maxi":                  true,
	"kirjasto":              true,
	"kronikka":              true,
	"serie_extra":           true,
	"texone":                true,
	"mini_texone_maxi_tex":  true,
	"almanacco_del_west":    true,
	"color_tex":             true,
	"tex_romanzi_a_fumetti": true,
	"tex_magazine":          true,
	"special":               true,
}

var allowedSorts = map[string]bool{
	"alpha":       true,
	"alpha_it":    true,
	"fi_pub_date": true,
	"it_pub_date": true,
}

func parsePositiveInt(raw string, fallback int) (int, error) {
	raw = strings.TrimSpace(raw)
	if len(raw) == 0 {
		return fallback, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return 0, fmt.Errorf("invalid integer value")
	}
	return value, nil
}

func parseAllowedValue(raw string, fallback string, allowed map[string]bool) (string, error) {
	raw = strings.TrimSpace(strings.ToLower(raw))
	if len(raw) == 0 {
		return fallback, nil
	}
	if !allowed[raw] {
		return "", fmt.Errorf("invalid value")
	}
	return raw, nil
}

// parseJumpYear returns the year of a jumpYear query value, or 0 when it is
// missing or not a plausible publication year.
func parseJumpYear(raw string) int {
	year, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || year < 1900 || year > 2100 {
		return 0
	}
	return year
}

func parseStoryListParams(c *fiber.Ctx) (db.StoryListParams, error) {
	page, err := parsePositiveInt(c.Query("page"), defaultPage)
	if err != nil {
		return db.StoryListParams{}, fmt.Errorf("page must be a positive integer")
	}

	pageSize, err := parsePositiveInt(c.Query("pageSize"), defaultPageSize)
	if err != nil {
		return db.StoryListParams{}, fmt.Errorf("pageSize must be a positive integer")
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	publication, err := parseAllowedValue(c.Query("publication"), defaultPublicationFilter, allowedPublicationFilters)
	if err != nil {
		return db.StoryListParams{}, fmt.Errorf("publication filter is invalid")
	}

	sort, err := parseAllowedValue(c.Query("sort"), defaultSort, allowedSorts)
	if err != nil {
		return db.StoryListParams{}, fmt.Errorf("sort is invalid")
	}

	year, err := parsePositiveInt(c.Query("year"), defaultYear)
	if err != nil {
		return db.StoryListParams{}, fmt.Errorf("year must be a positive integer")
	}

	// A search word that is a year (e.g. "1988") works as the year filter.
	search := strings.TrimSpace(c.Query("q", ""))
	if searchYear := parseJumpYear(search); searchYear > 0 {
		year = searchYear
		search = ""
	}
	// A single year is listed whole on one page (a year has at most a few dozen stories).
	if year > 0 {
		page = 1
		pageSize = maxPageSize
	}

	return db.StoryListParams{
		Publication: publication,
		Sort:        sort,
		Search:      search,
		Year:        year,
		Page:        page,
		PageSize:    pageSize,
	}, nil
}

func ListStoriesHandler(c *fiber.Ctx) error {
	versionRepo := db.NewVersionRepository() // TODO: move active version to fiber context
	version, err := versionRepo.GetActive()
	if err != nil {
		return c.SendStatus(500)
	}

	params, err := parseStoryListParams(c)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	storyRepo := db.NewStoryRepository()

	// letter=<a-ö> jumps to the page where that letter starts (alphabetical sort only).
	letters, letterPosition, err := storyRepo.LetterPaging(version, params, db.NormalizePagingLetter(c.Query("letter")))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to list stories"})
	}
	// jumpYear=<year> jumps to the page where that year starts (date sorts only).
	years, yearPosition, err := storyRepo.YearPaging(version, params, parseJumpYear(c.Query("jumpYear")))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to list stories"})
	}
	if letterPosition == nil {
		letterPosition = yearPosition
	}
	// letterIndex: 0-based position of the letter's or year's first story on the page.
	var letterIndex interface{}
	if letterPosition != nil {
		params.Page = letterPosition.Page
		letterIndex = letterPosition.Index
	}

	stories, total, err := storyRepo.ListFiltered(version, params)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to list stories"})
	}

	totalPages := 0
	if total > 0 {
		totalPages = (total + params.PageSize - 1) / params.PageSize
	}

	return c.JSON(fiber.Map{
		"stories": stories,
		"meta": fiber.Map{
			"total":       total,
			"page":        params.Page,
			"pageSize":    params.PageSize,
			"totalPages":  totalPages,
			"letters":     letters,
			"years":       years,
			"letterIndex": letterIndex,
		},
		"filters": fiber.Map{
			"publication": params.Publication,
			"sort":        params.Sort,
			// the search word as given, also when it was a year
			"q":           strings.TrimSpace(c.Query("q", "")),
			"year":        params.Year,
		},
	})
}
