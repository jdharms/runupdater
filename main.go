package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"golang.org/x/sys/windows/svc"
)

// Config holds all configuration information for the application
type Config struct {
	// File watching configuration
	LiveSplitExportPath string
	FileCheckInterval   time.Duration

	// Google Sheets configuration
	GoogleSheetsCredentialsPath string
	GoogleSheetID               string
	RawDataSheet1Name           string
	RawDataSheet2Name           string

	// Logging configuration
	LogPath  string
	LogLevel string
}

// FileWatcher defines an interface for watching changes to a file
type FileWatcher interface {
	// Start begins watching for file changes and sends notifications on the returned channel
	Start(ctx context.Context) (<-chan string, error)
	// Stop halts the file watching process
	Stop() error
}

// ExcelReader defines an interface for reading data from Excel files
type ExcelReader interface {
	// ReadSheet reads the specified sheet from the Excel file at the given path
	// and returns the data as a 2D slice of strings
	ReadSheet(filePath, sheetName string) ([][]string, error)
}

// SheetsClient defines an interface for interacting with Google Sheets
type SheetsClient interface {
	// Connect establishes a connection to Google Sheets API
	Connect(credentialsPath string) error
	// UpdateSheet updates the specified sheet with the provided data
	UpdateSheet(sheetID, sheetName string, data [][]string) error
}

// RunUpdater coordinates the file watching and data syncing processes
type RunUpdater struct {
	config       Config
	fileWatcher  FileWatcher
	excelReader  ExcelReader
	sheetsClient SheetsClient
	logger       *log.Logger
}

// NewRunUpdater creates a new instance of RunUpdater
func NewRunUpdater(config Config) *RunUpdater {
	logger := log.New(os.Stdout, "RunUpdater: ", log.LstdFlags)

	if config.LogPath != "" {
		logFile, err := os.OpenFile(config.LogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err == nil {
			logger = log.New(logFile, "RunUpdater: ", log.LstdFlags)
		} else {
			logger.Printf("Failed to open log file: %v", err)
		}
	}

	// Create concrete implementations of each interface
	fileWatcher := NewFSNotifyWatcher(config.LiveSplitExportPath, config.FileCheckInterval, logger)
	excelReader := NewExcelizeReader(logger)
	sheetsClient := NewGoogleSheetsClient(logger)

	return &RunUpdater{
		config:       config,
		fileWatcher:  fileWatcher,
		excelReader:  excelReader,
		sheetsClient: sheetsClient,
		logger:       logger,
	}
}

// Start begins the run updater service
func (ru *RunUpdater) Start(ctx context.Context) error {
	ru.logger.Println("Starting RunUpdater service")

	// Connect to Google Sheets
	err := ru.sheetsClient.Connect(ru.config.GoogleSheetsCredentialsPath)
	if err != nil {
		return fmt.Errorf("failed to connect to Google Sheets: %w", err)
	}

	// Start the file watcher
	changesCh, err := ru.fileWatcher.Start(ctx)
	if err != nil {
		return fmt.Errorf("failed to start file watcher: %w", err)
	}

	// Process file changes
	go func() {
		for {
			select {
			case <-ctx.Done():
				ru.logger.Println("Context cancelled, stopping processing")
				return
			case changedFile, ok := <-changesCh:
				if !ok {
					ru.logger.Println("File watcher channel closed")
					return
				}
				ru.logger.Printf("Detected change in file: %s", changedFile)
				err := ru.processFileChange(changedFile)
				if err != nil {
					ru.logger.Printf("Error processing file change: %v", err)
				}
			}
		}
	}()

	return nil
}

// Stop gracefully shuts down the run updater
func (ru *RunUpdater) Stop() error {
	ru.logger.Println("Stopping RunUpdater service")
	return ru.fileWatcher.Stop()
}

// processFileChange handles changes to the LiveSplit export file
func (ru *RunUpdater) processFileChange(filePath string) error {
	// Only process if the changed file is the one we're watching
	if filepath.Clean(filePath) != filepath.Clean(ru.config.LiveSplitExportPath) {
		return nil
	}

	ru.logger.Println("Processing changes to LiveSplit export file")

	// Read the first sheet
	data1, err := ru.excelReader.ReadSheet(filePath, ru.config.RawDataSheet1Name)
	if err != nil {
		return fmt.Errorf("failed to read first sheet: %w", err)
	}

	// Read the second sheet
	data2, err := ru.excelReader.ReadSheet(filePath, ru.config.RawDataSheet2Name)
	if err != nil {
		return fmt.Errorf("failed to read second sheet: %w", err)
	}

	// Update the first Google Sheet
	err = ru.sheetsClient.UpdateSheet(ru.config.GoogleSheetID, ru.config.RawDataSheet1Name, data1)
	if err != nil {
		return fmt.Errorf("failed to update first Google Sheet: %w", err)
	}

	// Update the second Google Sheet
	err = ru.sheetsClient.UpdateSheet(ru.config.GoogleSheetID, ru.config.RawDataSheet2Name, data2)
	if err != nil {
		return fmt.Errorf("failed to update second Google Sheet: %w", err)
	}

	ru.logger.Println("Successfully updated Google Sheets")
	return nil
}

func main() {
	// Load configuration from file
	configPath := "config.json"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	config, err := LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Check if running as a service
	isService, err := svc.IsWindowsService()
	if err != nil {
		log.Fatalf("Failed to determine if running as service: %v", err)
	}

	if isService {
		// Run as a Windows service
		RunAsService("RunUpdater", false, config)
	} else {
		// Run as a regular program for testing/debugging
		updater := NewRunUpdater(config)

		// Create a context for the service
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Start the service
		err := updater.Start(ctx)
		if err != nil {
			log.Fatalf("Failed to start RunUpdater: %v", err)
		}

		fmt.Println("RunUpdater is running. Press Ctrl+C to stop.")

		// Wait for Ctrl+C
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c

		fmt.Println("Shutting down...")
		cancel()
		updater.Stop()
	}
}
