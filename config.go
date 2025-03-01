package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// LoadConfig loads configuration from a JSON file
func LoadConfig(path string) (Config, error) {
	// Default configuration
	config := Config{
		FileCheckInterval: 5 * time.Second,
		LogLevel:          "info",
	}

	// Read configuration file
	file, err := os.Open(path)
	if err != nil {
		return config, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	// Parse JSON
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return config, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate required fields
	if config.LiveSplitExportPath == "" {
		return config, fmt.Errorf("LiveSplitExportPath is required")
	}
	if config.GoogleSheetsCredentialsPath == "" {
		return config, fmt.Errorf("GoogleSheetsCredentialsPath is required")
	}
	if config.GoogleSheetID == "" {
		return config, fmt.Errorf("GoogleSheetID is required")
	}
	if config.RawDataSheet1Name == "" {
		return config, fmt.Errorf("RawDataSheet1Name is required")
	}
	if config.RawDataSheet2Name == "" {
		return config, fmt.Errorf("RawDataSheet2Name is required")
	}

	return config, nil
}
