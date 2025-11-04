// /logger/log_manager_test.go

package logger

import (
	"fmt"
	"os"
	"sync"
	"testing"
	"time"
)

func TestAsyncLogging(t *testing.T) {
	// Create temp log file
	tmpFile := "./test_async.log"
	defer os.Remove(tmpFile)

	config := Config{
		Backend: BackendFile,
		BackendConfig: FileConfig{
			FilePath:      tmpFile,
			MaxFileSizeMB: 10,
		},
		Async: true,
	}

	lm, err := NewLogManager(config)
	if err != nil {
		t.Fatalf("Failed to create log manager: %v", err)
	}
	defer lm.Close()

	// Write logs concurrently
	var wg sync.WaitGroup
	numGoroutines := 10
	logsPerGoroutine := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < logsPerGoroutine; j++ {
				err := lm.WriteLog(LevelInfo, fmt.Sprintf("Goroutine %d - Log %d", id, j))
				if err != nil {
					t.Errorf("Failed to write log: %v", err)
				}
			}
		}(i)
	}

	wg.Wait()

	// Allow time for async writes to complete
	time.Sleep(200 * time.Millisecond)

	// Verify logs were written
	if _, err := os.Stat(tmpFile); os.IsNotExist(err) {
		t.Error("Log file was not created")
	}
}

func TestSyncLogging(t *testing.T) {
	tmpFile := "./test_sync.log"
	defer os.Remove(tmpFile)

	config := Config{
		Backend: BackendFile,
		BackendConfig: FileConfig{
			FilePath:      tmpFile,
			MaxFileSizeMB: 10,
		},
		Async: false,
	}

	lm, err := NewLogManager(config)
	if err != nil {
		t.Fatalf("Failed to create log manager: %v", err)
	}
	defer lm.Close()

	// Write logs
	for i := 0; i < 100; i++ {
		err := lm.WriteLog(LevelInfo, fmt.Sprintf("Sync log %d", i))
		if err != nil {
			t.Errorf("Failed to write log: %v", err)
		}
	}

	// Verify logs were written immediately
	if _, err := os.Stat(tmpFile); os.IsNotExist(err) {
		t.Error("Log file was not created")
	}
}

func TestAsyncVsSyncPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	numLogs := 1000

	// Test async
	asyncFile := "./test_async_perf.log"
	defer os.Remove(asyncFile)

	asyncConfig := Config{
		Backend: BackendFile,
		BackendConfig: FileConfig{
			FilePath:      asyncFile,
			MaxFileSizeMB: 10,
		},
		Async: true,
	}

	asyncLm, _ := NewLogManager(asyncConfig)
	defer asyncLm.Close()

	asyncStart := time.Now()
	for i := 0; i < numLogs; i++ {
		asyncLm.WriteLog(LevelInfo, fmt.Sprintf("Async log %d", i))
	}
	asyncDuration := time.Since(asyncStart)

	// Wait for async writes
	time.Sleep(200 * time.Millisecond)

	// Test sync
	syncFile := "./test_sync_perf.log"
	defer os.Remove(syncFile)

	syncConfig := Config{
		Backend: BackendFile,
		BackendConfig: FileConfig{
			FilePath:      syncFile,
			MaxFileSizeMB: 10,
		},
		Async: false,
	}

	syncLm, _ := NewLogManager(syncConfig)
	defer syncLm.Close()

	syncStart := time.Now()
	for i := 0; i < numLogs; i++ {
		syncLm.WriteLog(LevelInfo, fmt.Sprintf("Sync log %d", i))
	}
	syncDuration := time.Since(syncStart)

	t.Logf("Async duration: %v", asyncDuration)
	t.Logf("Sync duration: %v", syncDuration)
	t.Logf("Speedup: %.2fx", float64(syncDuration)/float64(asyncDuration))

	if asyncDuration >= syncDuration {
		t.Log("Warning: Async logging was not faster than sync (this can happen with small datasets)")
	}
}

func TestAsyncLogHandlers(t *testing.T) {
	tmpFile := "./test_handler.log"
	defer os.Remove(tmpFile)

	config := Config{
		Backend: BackendFile,
		BackendConfig: FileConfig{
			FilePath:      tmpFile,
			MaxFileSizeMB: 10,
		},
		Async: true,
	}

	lm, err := NewLogManager(config)
	if err != nil {
		t.Fatalf("Failed to create log manager: %v", err)
	}
	defer lm.Close()

	// Register custom handler
	handler := &TestLogHandler{
		handledLogs: make([]LogEntry, 0),
	}
	lm.RegisterLogHandler(handler)

	// Write logs
	lm.WriteLog(LevelInfo, "Test message 1")
	lm.WriteLog(LevelError, "Test error 1")
	lm.WriteLog(LevelWarn, "Test warning 1")

	// Wait for async processing
	time.Sleep(200 * time.Millisecond)

	if len(handler.handledLogs) != 3 {
		t.Errorf("Expected 3 handled logs, got %d", len(handler.handledLogs))
	}
}

func TestAsyncGracefulShutdown(t *testing.T) {
	tmpFile := "./test_shutdown.log"
	defer os.Remove(tmpFile)

	config := Config{
		Backend: BackendFile,
		BackendConfig: FileConfig{
			FilePath:      tmpFile,
			MaxFileSizeMB: 10,
		},
		Async: true,
	}

	lm, err := NewLogManager(config)
	if err != nil {
		t.Fatalf("Failed to create log manager: %v", err)
	}

	// Write many logs
	for i := 0; i < 1000; i++ {
		lm.WriteLog(LevelInfo, fmt.Sprintf("Log %d", i))
	}

	// Close should wait for all logs to be written
	err = lm.Close()
	if err != nil {
		t.Errorf("Failed to close log manager: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(tmpFile); os.IsNotExist(err) {
		t.Error("Log file was not created after close")
	}
}

// TestLogHandler for testing
type TestLogHandler struct {
	mu          sync.Mutex
	handledLogs []LogEntry
}

func (h *TestLogHandler) Handle(entry LogEntry) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.handledLogs = append(h.handledLogs, entry)
	return nil
}
