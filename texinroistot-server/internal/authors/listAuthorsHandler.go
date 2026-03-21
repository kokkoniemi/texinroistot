package authors

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/kokkoniemi/texinroistot/internal/db"
)

const (
	defaultType     = "writer"
	defaultSort     = "last_name"
	defaultPage     = 1
	defaultPageSize = 25
	maxPageSize     = 100
)

var allowedTypes = map[string]bool{
	"writer": true,
	"drawer": true,
}

var allowedSorts = map[string]bool{
	"first_name": true,
	"last_name":  true,
}

type authorListParams struct {
	Type     string
	Sort     string
	Search   string
	Page     int
	PageSize int
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

func parseAuthorListParams(c *fiber.Ctx) (authorListParams, error) {
	page, err := parsePositiveInt(c.Query("page"), defaultPage)
	if err != nil {
		return authorListParams{}, fmt.Errorf("page must be a positive integer")
	}

	pageSize, err := parsePositiveInt(c.Query("pageSize"), defaultPageSize)
	if err != nil {
		return authorListParams{}, fmt.Errorf("pageSize must be a positive integer")
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	authorType, err := parseAllowedValue(c.Query("type"), defaultType, allowedTypes)
	if err != nil {
		return authorListParams{}, fmt.Errorf("type is invalid")
	}

	sortValue, err := parseAllowedValue(c.Query("sort"), defaultSort, allowedSorts)
	if err != nil {
		return authorListParams{}, fmt.Errorf("sort is invalid")
	}

	return authorListParams{
		Type:     authorType,
		Sort:     sortValue,
		Search:   strings.TrimSpace(c.Query("q", "")),
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func authorMatchesType(author *db.Author, authorType string) bool {
	if author == nil {
		return false
	}
	if authorType == "writer" {
		return author.IsWriter
	}
	return author.IsDrawer
}

func normalizedText(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func authorNameMatches(author *db.Author, search string) bool {
	if search == "" {
		return true
	}
	search = normalizedText(search)
	firstName := normalizedText(author.FirstName)
	lastName := normalizedText(author.LastName)
	fullName := strings.TrimSpace(firstName + " " + lastName)
	return strings.Contains(firstName, search) ||
		strings.Contains(lastName, search) ||
		strings.Contains(fullName, search)
}

func compareAuthors(sortValue string, a *db.Author, b *db.Author) bool {
	firstA := normalizedText(a.FirstName)
	lastA := normalizedText(a.LastName)
	firstB := normalizedText(b.FirstName)
	lastB := normalizedText(b.LastName)

	if sortValue == "first_name" {
		if firstA != firstB {
			return firstA < firstB
		}
		if lastA != lastB {
			return lastA < lastB
		}
		return normalizedText(a.Hash) < normalizedText(b.Hash)
	}

	if lastA != lastB {
		return lastA < lastB
	}
	if firstA != firstB {
		return firstA < firstB
	}
	return normalizedText(a.Hash) < normalizedText(b.Hash)
}

func paginateAuthors(authors []*db.Author, page int, pageSize int) []*db.Author {
	if len(authors) == 0 {
		return []*db.Author{}
	}

	offset := (page - 1) * pageSize
	if offset >= len(authors) {
		return []*db.Author{}
	}

	end := offset + pageSize
	if end > len(authors) {
		end = len(authors)
	}

	return authors[offset:end]
}

func ListAuthorsHandler(c *fiber.Ctx) error {
	versionRepo := db.NewVersionRepository() // TODO: move active version to fiber context
	version, err := versionRepo.GetActive()
	if err != nil {
		return c.SendStatus(500)
	}

	params, err := parseAuthorListParams(c)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	authorRepo := db.NewAuthorRepository()
	allAuthors, err := authorRepo.List(version)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to list authors"})
	}

	var filteredAuthors []*db.Author
	for _, author := range allAuthors {
		if !authorMatchesType(author, params.Type) {
			continue
		}
		if !authorNameMatches(author, params.Search) {
			continue
		}
		filteredAuthors = append(filteredAuthors, author)
	}

	sort.SliceStable(filteredAuthors, func(i, j int) bool {
		return compareAuthors(params.Sort, filteredAuthors[i], filteredAuthors[j])
	})

	total := len(filteredAuthors)
	totalPages := 0
	if total > 0 {
		totalPages = (total + params.PageSize - 1) / params.PageSize
	}

	return c.JSON(fiber.Map{
		"authors": paginateAuthors(filteredAuthors, params.Page, params.PageSize),
		"meta": fiber.Map{
			"total":      total,
			"page":       params.Page,
			"pageSize":   params.PageSize,
			"totalPages": totalPages,
		},
		"filters": fiber.Map{
			"type": params.Type,
			"sort": params.Sort,
			"q":    params.Search,
		},
	})
}
