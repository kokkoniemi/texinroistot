package main

import (
	"fmt"

	_ "github.com/joho/godotenv/autoload"
	"github.com/kokkoniemi/texinroistot/internal/admin"
	"github.com/xuri/excelize/v2"
)

func main() {
	err := parseExcel()
	if err != nil {
		panic(err)
	}
}

func parseExcel() error {
	f, err := excelize.OpenFile("Texinroistot.xlsx")

	if err != nil {
		return err
	}

	defer closeExcel(f)

	rows, err := f.GetRows("Taul1")
	if err != nil {
		return err
	}

	if len(rows) <= 1 {
		return fmt.Errorf("no content")
	}

	importer := admin.NewSpreadsheetImporter(rows[0])
	err = importer.LoadData(rows[1:])
	if err != nil {
		return err
	}
	importer.SaveData()

	return nil
}

func closeExcel(f *excelize.File) error {
	err := f.Close()
	if err != nil {
		return err
	}
	return nil
}
