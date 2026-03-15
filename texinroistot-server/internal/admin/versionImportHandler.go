package admin

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/kokkoniemi/texinroistot/internal/config"
	"github.com/kokkoniemi/texinroistot/internal/db"
	"github.com/kokkoniemi/texinroistot/internal/importer"
)

const (
	importRequestTimeout = 2 * time.Minute
	maxImportFileBytes   = 20 * 1024 * 1024
	importUserAgent      = "Mozilla/5.0 (Windows NT 10.0; Win64; x64)"
)

var (
	errInvalidImportURL         = errors.New("invalid import url")
	errImportDownloadFailed     = errors.New("failed to download import file")
	errImportInvalidSpreadsheet = errors.New("downloaded file is not a valid xlsx")

	importStateMu    sync.Mutex
	importRunning    bool
	runVersionImport = importVersionFromURL
)

func ImportVersionHandler(c *fiber.Ctx) error {
	if !startImport() {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "import already running"})
	}
	defer finishImport()

	version, err := runVersionImport(config.ImportExcelURL)
	if err != nil {
		if errors.Is(err, errInvalidImportURL) ||
			errors.Is(err, errImportDownloadFailed) ||
			errors.Is(err, errImportInvalidSpreadsheet) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to import version"})
	}

	return c.JSON(fiber.Map{
		"imported": true,
		"version":  version,
	})
}

func startImport() bool {
	importStateMu.Lock()
	defer importStateMu.Unlock()

	if importRunning {
		return false
	}
	importRunning = true
	return true
}

func finishImport() {
	importStateMu.Lock()
	defer importStateMu.Unlock()
	importRunning = false
}

func importVersionFromURL(rawURL string) (*db.Version, error) {
	fileURL, err := buildImportURL(rawURL)
	if err != nil {
		return nil, err
	}

	content, err := downloadSpreadsheet(fileURL)
	if err != nil {
		return nil, err
	}

	version, err := importer.ImportSpreadsheetFromBytes(content)
	if err != nil {
		return nil, err
	}
	return version, nil
}

func buildImportURL(rawURL string) (string, error) {
	trimmed := strings.TrimSpace(rawURL)
	if trimmed == "" {
		return "", fmt.Errorf("%w: empty url", errInvalidImportURL)
	}

	parsed, err := url.Parse(trimmed)
	if err != nil {
		return "", fmt.Errorf("%w: %v", errInvalidImportURL, err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("%w: unsupported scheme", errInvalidImportURL)
	}
	if parsed.Hostname() == "" {
		return "", fmt.Errorf("%w: host is required", errInvalidImportURL)
	}

	if isOneDriveHost(parsed.Hostname()) {
		query := parsed.Query()
		if strings.TrimSpace(query.Get("download")) == "" {
			query.Set("download", "1")
			parsed.RawQuery = query.Encode()
		}
	}

	return parsed.String(), nil
}

func isOneDriveHost(host string) bool {
	normalized := strings.ToLower(strings.TrimSpace(host))
	return normalized == "1drv.ms" ||
		normalized == "onedrive.live.com" ||
		strings.HasSuffix(normalized, ".onedrive.live.com")
}

func downloadSpreadsheet(sourceURL string) ([]byte, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to initialize cookie jar", errImportDownloadFailed)
	}

	client := &http.Client{
		Timeout: importRequestTimeout,
		Jar:     jar,
	}

	content, contentType, err := downloadSpreadsheetOnce(client, sourceURL)
	if err != nil {
		return nil, err
	}
	if err := validateSpreadsheet(content, contentType); err != nil {
		return nil, err
	}
	return content, nil
}

func downloadSpreadsheetOnce(client *http.Client, sourceURL string) ([]byte, string, error) {
	req, err := http.NewRequest(http.MethodGet, sourceURL, nil)
	if err != nil {
		return nil, "", fmt.Errorf("%w: %v", errImportDownloadFailed, err)
	}
	req.Header.Set("User-Agent", importUserAgent)

	response, err := client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("%w: %v", errImportDownloadFailed, err)
	}
	defer response.Body.Close()

	limited := io.LimitReader(response.Body, maxImportFileBytes+1)
	content, err := io.ReadAll(limited)
	if err != nil {
		return nil, "", fmt.Errorf("%w: %v", errImportDownloadFailed, err)
	}
	if len(content) > maxImportFileBytes {
		return nil, "", fmt.Errorf("%w: file too large", errImportDownloadFailed)
	}

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, "", fmt.Errorf("%w: got status %d", errImportDownloadFailed, response.StatusCode)
	}

	return content, response.Header.Get("Content-Type"), nil
}

func isLikelyHTMLResponse(content []byte, contentType string) bool {
	if strings.Contains(strings.ToLower(contentType), "text/html") {
		return true
	}

	if len(content) == 0 {
		return false
	}

	maxSample := len(content)
	if maxSample > 1024 {
		maxSample = 1024
	}
	sample := strings.ToLower(string(content[:maxSample]))
	return strings.Contains(sample, "<html") ||
		strings.Contains(sample, "<!doctype") ||
		strings.Contains(sample, "login.live.com")
}

func validateSpreadsheet(content []byte, contentType string) error {
	if len(content) == 0 {
		return fmt.Errorf("%w: empty response", errImportInvalidSpreadsheet)
	}
	if bytes.HasPrefix(content, []byte("PK\x03\x04")) {
		return nil
	}

	if isLikelyHTMLResponse(content, contentType) {
		return fmt.Errorf(
			"%w: received html/login page. ensure OneDrive sharing link is publicly downloadable",
			errImportInvalidSpreadsheet,
		)
	}

	return fmt.Errorf("%w: missing xlsx zip signature", errImportInvalidSpreadsheet)
}
