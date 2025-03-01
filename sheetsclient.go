package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// GoogleSheetsClient implements SheetsClient using the Google Sheets API
type GoogleSheetsClient struct {
	service *sheets.Service
	logger  *log.Logger
}

// NewGoogleSheetsClient creates a new GoogleSheetsClient instance
func NewGoogleSheetsClient(logger *log.Logger) *GoogleSheetsClient {
	return &GoogleSheetsClient{
		logger: logger,
	}
}

// Connect establishes a connection to Google Sheets API
func (gsc *GoogleSheetsClient) Connect(credentialsPath string) error {
	gsc.logger.Printf("Connecting to Google Sheets API using credentials: %s", credentialsPath)

	ctx := context.Background()
	b, err := os.ReadFile(credentialsPath)
	if err != nil {
		return fmt.Errorf("unable to read client secret file: %w", err)
	}

	// If you're using a service account
	config, err := google.JWTConfigFromJSON(b, sheets.SpreadsheetsScope)
	if err != nil {
		return fmt.Errorf("unable to parse client secret file to config: %w", err)
	}
	client := config.Client(ctx)

	// Create the sheets service
	srv, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return fmt.Errorf("unable to retrieve Sheets client: %w", err)
	}

	gsc.service = srv
	gsc.logger.Println("Successfully connected to Google Sheets API")
	return nil
}

// UpdateSheet updates the specified sheet with the provided data
func (gsc *GoogleSheetsClient) UpdateSheet(sheetID, sheetName string, data [][]string) error {
	if gsc.service == nil {
		return fmt.Errorf("not connected to Google Sheets API")
	}

	gsc.logger.Printf("Updating sheet '%s' in spreadsheet '%s'", sheetName, sheetID)

	// First, we need to find the sheet by name to get its ID
	resp, err := gsc.service.Spreadsheets.Get(sheetID).Do()
	if err != nil {
		return fmt.Errorf("unable to retrieve spreadsheet: %w", err)
	}

	var sheetFound bool
	for _, sheet := range resp.Sheets {
		if sheet.Properties.Title == sheetName {
			sheetFound = true
			break
		}
	}

	if !sheetFound {
		return fmt.Errorf("sheet '%s' not found in spreadsheet", sheetName)
	}

	// Convert data to ValueRange
	valueRange := &sheets.ValueRange{
		Values: make([][]interface{}, len(data)),
	}
	for i, row := range data {
		valueRange.Values[i] = make([]interface{}, len(row))
		for j, cell := range row {
			valueRange.Values[i][j] = cell
		}
	}

	// Clear the sheet first (to remove any old data)
	clearRequest := gsc.service.Spreadsheets.Values.Clear(sheetID, sheetName, &sheets.ClearValuesRequest{})
	_, err = clearRequest.Do()
	if err != nil {
		return fmt.Errorf("unable to clear sheet: %w", err)
	}

	// Update the sheet with new data
	updateRequest := gsc.service.Spreadsheets.Values.Update(sheetID, sheetName, valueRange)
	updateRequest.ValueInputOption("USER_ENTERED")
	_, err = updateRequest.Do()
	if err != nil {
		return fmt.Errorf("unable to update sheet: %w", err)
	}

	gsc.logger.Printf("Successfully updated sheet '%s' with %d rows", sheetName, len(data))
	return nil
}
