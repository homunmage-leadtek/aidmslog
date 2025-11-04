// /logger/backend_file.go

package logger

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

	// Create directory if needed
	dir := filepath.Dir(fileConfig.FilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Open file in append mode
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

// ✅ Helper: apply LogFilter (struct-based filter)
func applyFilter(entry LogEntry, filter LogFilter) bool {
	// Filter by keyword (Contains)
	if filter.Contains != "" && !strings.Contains(entry.Message, filter.Contains) {
		return false
	}

	// Filter: StartTime (entry must be AFTER start)
	if filter.StartTime != nil && entry.Timestamp.Before(*filter.StartTime) {
		return false
	}

	// Filter: EndTime (entry must be BEFORE end)
	if filter.EndTime != nil && entry.Timestamp.After(*filter.EndTime) {
		return false
	}

	return true
}

// ✅ Full implementation of Read()
func (fb *FileBackend) Read(level LogLevel, filter LogFilter) ([]LogEntry, error) {
	fb.mu.Lock()
	defer fb.mu.Unlock()

	if fb.file == nil {
		return nil, fmt.Errorf("file backend not initialized")
	}

	// Must open a NEW reader (fb.file is write-only)
	rf, err := os.Open(fb.config.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file for reading: %w", err)
	}
	defer rf.Close()

	scanner := bufio.NewScanner(rf)
	var results []LogEntry

	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			continue
		}

		// Expected format:
		// [2025-01-01T12:00:00Z] INFO : message
		if !strings.HasPrefix(line, "[") {
			continue
		}

		end := strings.Index(line, "]")
		if end == -1 {
			continue
		}

		tsStr := line[1:end]
		ts, err := time.Parse(time.RFC3339, tsStr)
		if err != nil {
			continue
		}

		rest := strings.TrimSpace(line[end+1:])
		parts := strings.SplitN(rest, ":", 2)
		if len(parts) != 2 {
			continue
		}

		levelStr := strings.TrimSpace(parts[0])
		msg := strings.TrimSpace(parts[1])

		entry := LogEntry{
			Timestamp: ts,
			Level:     LogLevel(levelStr),
			Message:   msg,
		}

		// Filter by level
		if level != "" && entry.Level != level {
			continue
		}

		// Apply struct-based filter
		if !applyFilter(entry, filter) {
			continue
		}

		results = append(results, entry)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

// ✅ Full implementation of ClearLogs(before)
func (fb *FileBackend) ClearLogs(before time.Time) error {
	fb.mu.Lock()
	defer fb.mu.Unlock()

	if fb.file == nil {
		return fmt.Errorf("file backend not initialized")
	}

	// Read all logs (no filter, no level)
	logs, err := fb.Read("", LogFilter{})
	if err != nil {
		return fmt.Errorf("clear logs failed: %w", err)
	}

	// Keep only logs newer than `before`
	var kept []LogEntry
	for _, e := range logs {
		if e.Timestamp.After(before) {
			kept = append(kept, e)
		}
	}

	// Truncate file
	fb.file.Close()
	f, err := os.OpenFile(fb.config.FilePath, os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to truncate log file: %w", err)
	}
	fb.file = f

	// Rewrite logs
	for _, e := range kept {
		if err := fb.Write(e); err != nil {
			return err
		}
	}

	return nil
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
