package main

import (
	_ "github.com/joho/godotenv/autoload"
	"github.com/kokkoniemi/texinroistot/internal/importer"
)

func main() {
	err := parseExcel()
	if err != nil {
		panic(err)
	}
}

func parseExcel() error {
	_, err := importer.ImportSpreadsheetFromFile("Texinroistot.xlsx")
	return err
}
