package main

import (
	"fmt"
	"log"

	"github.com/xuri/excelize/v2"
)

// ExcelizeReader implements ExcelReader using the excelize library
type ExcelizeReader struct {
	logger *log.Logger
}

// NewExcelizeReader creates a new ExcelizeReader instance
func NewExcelizeReader(logger *log.Logger) *ExcelizeReader {
	return &ExcelizeReader{
		logger: logger,
	}
}

// ReadSheet reads the specified sheet from the Excel file at the given path
// and returns the data as a 2D slice of strings
func (er *ExcelizeReader) ReadSheet(filePath, sheetName string) ([][]string, error) {
	er.logger.Printf("Reading sheet '%s' from file '%s'", sheetName, filePath)

	// Open the Excel file
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open Excel file: %w", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			er.logger.Printf("Error closing Excel file: %v", err)
		}
	}()

	// Check if the sheet exists
	index, err := f.GetSheetIndex(sheetName)
	if err != nil || index < 0 {
		return nil, fmt.Errorf("sheet '%s' not found: %w", sheetName, err)
	}

	// Get all rows from the sheet
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to read rows from sheet: %w", err)
	}

	er.logger.Printf("Successfully read %d rows from sheet '%s'", len(rows), sheetName)
	return rows, nil
}
