// pkg/logger/file_backend.go

package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type FileBackend struct {
	mu     sync.Mutex
	config FileConfig
	file   *os.File
}

func (fb *FileBackend) Init(config interface{}) error {
	fileConfig, ok := config.(FileConfig)
	if !ok {
		return fmt.Errorf("invalid config type for file backend")
	}

	fb.config = fileConfig

	// Create directory if not exists
	dir := filepath.Dir(fileConfig.FilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Open file
	f, err := os.OpenFile(fileConfig.FilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	fb.file = f
	return nil
}

func (fb *FileBackend) Write(entry LogEntry) error {
	fb.mu.Lock()
	defer fb.mu.Unlock()

	if fb.file == nil {
		return fmt.Errorf("file backend not initialized")
	}

	timestamp := entry.Timestamp.Format(time.RFC3339)
	line := fmt.Sprintf("[%s] %-5s: %s\n", timestamp, entry.Level, entry.Message)

	if _, err := fb.file.WriteString(line); err != nil {
		return fmt.Errorf("failed to write log: %w", err)
	}

	return nil
}

func (fb *FileBackend) Read(level LogLevel, filter LogFilter) ([]LogEntry, error) {
	// TODO: Implement file reading with filtering
	return []LogEntry{}, fmt.Errorf("read not implemented for file backend")
}

func (fb *FileBackend) ClearLogs(before time.Time) error {
	// TODO: Implement log rotation/clearing
	return fmt.Errorf("clear logs not implemented for file backend")
}

func (fb *FileBackend) Close() error {
	fb.mu.Lock()
	defer fb.mu.Unlock()

	if fb.file != nil {
		err := fb.file.Close()
		fb.file = nil
		return err
	}
	return nil
}
