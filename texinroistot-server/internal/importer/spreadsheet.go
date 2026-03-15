package importer

import (
	"bytes"
	"fmt"

	"github.com/kokkoniemi/texinroistot/internal/db"
	"github.com/xuri/excelize/v2"
)

const inputSheetName = "Taul1"

func ImportSpreadsheetFromFile(path string) (*db.Version, error) {
	file, err := excelize.OpenFile(path)
	if err != nil {
		return nil, err
	}
	defer closeSpreadsheet(file)

	return importSpreadsheet(file)
}

func ImportSpreadsheetFromBytes(content []byte) (*db.Version, error) {
	file, err := excelize.OpenReader(bytes.NewReader(content))
	if err != nil {
		return nil, err
	}
	defer closeSpreadsheet(file)

	return importSpreadsheet(file)
}

func importSpreadsheet(file *excelize.File) (*db.Version, error) {
	rows, err := file.GetRows(inputSheetName)
	if err != nil {
		return nil, err
	}

	if len(rows) <= 1 {
		return nil, fmt.Errorf("no content")
	}

	spreadsheetImporter, err := NewSpreadsheetImporter(rows[0])
	if err != nil {
		return nil, err
	}
	if err := spreadsheetImporter.LoadData(rows[1:]); err != nil {
		return nil, err
	}

	return spreadsheetImporter.PersistDataWithVersion()
}

func closeSpreadsheet(file *excelize.File) error {
	return file.Close()
}
