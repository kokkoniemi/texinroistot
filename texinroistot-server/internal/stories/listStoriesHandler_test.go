package stories

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/kokkoniemi/texinroistot/internal/db"
)

func testParseParamsRoute() *fiber.App {
	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		params, err := parseStoryListParams(c)
		if err != nil {
			return c.Status(400).SendString(err.Error())
		}
		return c.JSON(params)
	})
	return app
}

func decodeParamsFromResponse(t *testing.T, app *fiber.App, target string) (db.StoryListParams, int, string) {
	t.Helper()

	req := httptest.NewRequest("GET", target, nil)
	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}
	body := string(bodyBytes)

	if res.StatusCode != 200 {
		return db.StoryListParams{}, res.StatusCode, body
	}

	var params db.StoryListParams
	if err = json.Unmarshal(bodyBytes, &params); err != nil {
		t.Fatalf("failed to decode params json: %v", err)
	}
	return params, res.StatusCode, body
}

func TestParseStoryListParamsDefaults(t *testing.T) {
	app := testParseParamsRoute()

	params, status, body := decodeParamsFromResponse(t, app, "/")
	if status != 200 {
		t.Fatalf("expected 200, got %d (%s)", status, body)
	}

	if params.Publication != defaultPublicationFilter {
		t.Fatalf("expected default publication %q, got %q", defaultPublicationFilter, params.Publication)
	}
	if params.Sort != defaultSort {
		t.Fatalf("expected default sort %q, got %q", defaultSort, params.Sort)
	}
	if params.Page != defaultPage {
		t.Fatalf("expected default page %d, got %d", defaultPage, params.Page)
	}
	if params.PageSize != defaultPageSize {
		t.Fatalf("expected default pageSize %d, got %d", defaultPageSize, params.PageSize)
	}
	if params.Search != "" {
		t.Fatalf("expected empty default search, got %q", params.Search)
	}
	if params.Year != defaultYear {
		t.Fatalf("expected default year %d, got %d", defaultYear, params.Year)
	}
}

func TestParseStoryListParamsRejectsInvalidEnum(t *testing.T) {
	app := testParseParamsRoute()

	_, status, body := decodeParamsFromResponse(t, app, "/?sort=unknown")
	if status != 400 {
		t.Fatalf("expected 400 for invalid sort, got %d (%s)", status, body)
	}

	_, status, body = decodeParamsFromResponse(t, app, "/?publication=invalid")
	if status != 400 {
		t.Fatalf("expected 400 for invalid publication filter, got %d (%s)", status, body)
	}
}

func TestParseStoryListParamsRejectsInvalidPage(t *testing.T) {
	app := testParseParamsRoute()

	_, status, body := decodeParamsFromResponse(t, app, "/?page=0")
	if status != 400 {
		t.Fatalf("expected 400 for invalid page, got %d (%s)", status, body)
	}
}

func TestParseStoryListParamsClampsPageSize(t *testing.T) {
	app := testParseParamsRoute()

	params, status, body := decodeParamsFromResponse(t, app, "/?pageSize=1000")
	if status != 200 {
		t.Fatalf("expected 200, got %d (%s)", status, body)
	}

	if params.PageSize != maxPageSize {
		t.Fatalf("expected pageSize to be clamped to %d, got %d", maxPageSize, params.PageSize)
	}
}

func TestParseStoryListParamsParsesYear(t *testing.T) {
	app := testParseParamsRoute()

	params, status, body := decodeParamsFromResponse(t, app, "/?year=1980")
	if status != 200 {
		t.Fatalf("expected 200, got %d (%s)", status, body)
	}

	if params.Year != 1980 {
		t.Fatalf("expected parsed year to be 1980, got %d", params.Year)
	}
}

func TestParseStoryListParamsRejectsInvalidYear(t *testing.T) {
	app := testParseParamsRoute()

	_, status, body := decodeParamsFromResponse(t, app, "/?year=0")
	if status != 400 {
		t.Fatalf("expected 400 for invalid year, got %d (%s)", status, body)
	}

	_, status, body = decodeParamsFromResponse(t, app, "/?year=year1980")
	if status != 400 {
		t.Fatalf("expected 400 for invalid year format, got %d (%s)", status, body)
	}
}
