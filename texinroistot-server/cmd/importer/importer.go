package main

import (
	"fmt"

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

	// map[string][string]

	// create version

	// create authors

	// create stories

	// create villains & attach them to stories

	// create publications & attach stories to them

	for index, row := range rows {
		fmt.Println(index)
		fmt.Println(row)
	}

	return nil
}

func closeExcel(f *excelize.File) {
	err := f.Close()
	if err != nil {
		fmt.Println(err)
	}
}
