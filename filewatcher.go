package main

import (
	"context"
	"log"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// FSNotifyWatcher implements FileWatcher using fsnotify
type FSNotifyWatcher struct {
	filePath      string
	checkInterval time.Duration
	watcher       *fsnotify.Watcher
	events        chan string
	done          chan struct{}
	logger        *log.Logger
	mu            sync.Mutex
	isRunning     bool
}

// NewFSNotifyWatcher creates a new FSNotifyWatcher instance
func NewFSNotifyWatcher(filePath string, checkInterval time.Duration, logger *log.Logger) *FSNotifyWatcher {
	return &FSNotifyWatcher{
		filePath:      filePath,
		checkInterval: checkInterval,
		events:        make(chan string),
		done:          make(chan struct{}),
		logger:        logger,
	}
}

// Start begins watching for file changes and sends notifications on the returned channel
func (fw *FSNotifyWatcher) Start(ctx context.Context) (<-chan string, error) {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	if fw.isRunning {
		return fw.events, nil
	}

	// Create a new watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	fw.watcher = watcher

	// Get the directory to watch
	dir := filepath.Dir(fw.filePath)
	filename := filepath.Base(fw.filePath)

	// Start watching the directory
	err = fw.watcher.Add(dir)
	if err != nil {
		fw.watcher.Close()
		return nil, err
	}

	fw.isRunning = true

	// Start the goroutine that processes events
	go func() {
		defer close(fw.events)
		defer fw.watcher.Close()

		for {
			select {
			case <-ctx.Done():
				fw.logger.Println("Context cancelled, stopping file watcher")
				return
			case <-fw.done:
				fw.logger.Println("File watcher stopped")
				return
			case event, ok := <-fw.watcher.Events:
				if !ok {
					return
				}

				// Only process events for the target file
				if filepath.Base(event.Name) != filename {
					continue
				}

				// Check if this is a modification event
				if event.Op&fsnotify.Write == fsnotify.Write ||
					event.Op&fsnotify.Create == fsnotify.Create {
					fw.logger.Printf("Modified file: %s", event.Name)
					fw.events <- event.Name
				}
			case err, ok := <-fw.watcher.Errors:
				if !ok {
					return
				}
				fw.logger.Printf("Error watching file: %v", err)
			}
		}
	}()

	return fw.events, nil
}

// Stop halts the file watching process
func (fw *FSNotifyWatcher) Stop() error {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	if !fw.isRunning {
		return nil
	}

	close(fw.done)
	fw.isRunning = false
	return nil
}
