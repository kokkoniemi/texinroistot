package villains

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/kokkoniemi/texinroistot/internal/db"
)

const (
	defaultPublicationFilter = "fi"
	defaultSort              = "fi_pub_date"
	defaultPage              = 1
	defaultPageSize          = 25
	maxPageSize              = 100
)

var allowedPublicationFilters = map[string]bool{
	"all": true,
	"fi":  true,
	"it":  true,
}

var allowedSorts = map[string]bool{
	"first_name":  true,
	"last_name":   true,
	"nickname":    true,
	"other_name":  true,
	"code_name":   true,
	"rank":        true,
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

func parseVillainListParams(c *fiber.Ctx) (db.VillainListParams, error) {
	page, err := parsePositiveInt(c.Query("page"), defaultPage)
	if err != nil {
		return db.VillainListParams{}, fmt.Errorf("page must be a positive integer")
	}

	pageSize, err := parsePositiveInt(c.Query("pageSize"), defaultPageSize)
	if err != nil {
		return db.VillainListParams{}, fmt.Errorf("pageSize must be a positive integer")
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	publication, err := parseAllowedValue(c.Query("publication"), defaultPublicationFilter, allowedPublicationFilters)
	if err != nil {
		return db.VillainListParams{}, fmt.Errorf("publication filter is invalid")
	}

	sort, err := parseAllowedValue(c.Query("sort"), defaultSort, allowedSorts)
	if err != nil {
		return db.VillainListParams{}, fmt.Errorf("sort is invalid")
	}

	return db.VillainListParams{
		Publication: publication,
		Sort:        sort,
		Search:      strings.TrimSpace(c.Query("q", "")),
		Page:        page,
		PageSize:    pageSize,
	}, nil
}

func ListVillainsHandler(c *fiber.Ctx) error {
	versionRepo := db.NewVersionRepository() // TODO: move active version to fiber context
	version, err := versionRepo.GetActive()
	if err != nil {
		return c.SendStatus(500)
	}

	params, err := parseVillainListParams(c)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	villainRepo := db.NewVillainRepository()
	villains, total, err := villainRepo.ListFiltered(version, params)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to list villains"})
	}

	totalPages := 0
	if total > 0 {
		totalPages = (total + params.PageSize - 1) / params.PageSize
	}

	return c.JSON(fiber.Map{
		"villains": villains,
		"meta": fiber.Map{
			"total":      total,
			"page":       params.Page,
			"pageSize":   params.PageSize,
			"totalPages": totalPages,
		},
		"filters": fiber.Map{
			"publication": params.Publication,
			"sort":        params.Sort,
			"q":           params.Search,
		},
	})
}
