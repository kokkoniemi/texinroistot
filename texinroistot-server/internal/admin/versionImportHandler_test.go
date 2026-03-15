package admin

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/kokkoniemi/texinroistot/internal/db"
	"github.com/xuri/excelize/v2"
)

func TestBuildImportURLAddsDownloadParam(t *testing.T) {
	url, err := buildImportURL("https://1drv.ms/x/s!abc")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !bytes.Contains([]byte(url), []byte("download=1")) {
		t.Fatalf("expected download=1 in url, got %q", url)
	}
}

func TestBuildImportURLRejectsInvalidScheme(t *testing.T) {
	_, err := buildImportURL("ftp://1drv.ms/x/s!abc")
	if !errors.Is(err, errInvalidImportURL) {
		t.Fatalf("expected errInvalidImportURL, got %v", err)
	}
}

func TestDownloadSpreadsheetSupportsRedirect(t *testing.T) {
	xlsxBytes := mustBuildXLSX(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/redirect" {
			http.Redirect(w, r, "/file", http.StatusFound)
			return
		}
		if r.URL.Path == "/file" {
			w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
			_, _ = w.Write(xlsxBytes)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	content, err := downloadSpreadsheet(server.URL + "/redirect")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(content) == 0 {
		t.Fatalf("expected non-empty content")
	}
}

func TestDownloadSpreadsheetRejectsHTML(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte("<html><body>login.live.com</body></html>"))
	}))
	defer server.Close()

	_, err := downloadSpreadsheet(server.URL)
	if !errors.Is(err, errImportInvalidSpreadsheet) {
		t.Fatalf("expected errImportInvalidSpreadsheet, got %v", err)
	}
}

func TestImportVersionHandlerReturnsConflictWhenImportIsRunning(t *testing.T) {
	resetImportState(t)
	importRunning = true

	app := fiber.New()
	app.Post("/api/admin/versions/import", ImportVersionHandler)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/versions/import", nil)
	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if res.StatusCode != fiber.StatusConflict {
		t.Fatalf("expected %d, got %d", fiber.StatusConflict, res.StatusCode)
	}
}

func TestImportVersionHandlerSuccess(t *testing.T) {
	resetImportState(t)

	runVersionImport = func(_ string) (*db.Version, error) {
		return &db.Version{ID: 123, IsActive: false}, nil
	}

	app := fiber.New()
	app.Post("/api/admin/versions/import", ImportVersionHandler)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/versions/import", nil)
	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if res.StatusCode != fiber.StatusOK {
		t.Fatalf("expected %d, got %d", fiber.StatusOK, res.StatusCode)
	}
}

func TestImportVersionHandlerMapsSpreadsheetValidationErrorToBadRequest(t *testing.T) {
	resetImportState(t)

	runVersionImport = func(_ string) (*db.Version, error) {
		return nil, errImportInvalidSpreadsheet
	}

	app := fiber.New()
	app.Post("/api/admin/versions/import", ImportVersionHandler)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/versions/import", nil)
	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if res.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected %d, got %d", fiber.StatusBadRequest, res.StatusCode)
	}
}

func resetImportState(t *testing.T) {
	t.Helper()

	runVersionImport = importVersionFromURL
	importRunning = false
	t.Cleanup(func() {
		runVersionImport = importVersionFromURL
		importRunning = false
	})
}

func mustBuildXLSX(t *testing.T) []byte {
	t.Helper()

	file := excelize.NewFile()
	file.SetCellValue("Sheet1", "A1", "header")

	buffer, err := file.WriteToBuffer()
	if err != nil {
		t.Fatalf("failed to build xlsx bytes: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("failed to close xlsx: %v", err)
	}

	return buffer.Bytes()
}
