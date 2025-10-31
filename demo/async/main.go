package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/homunmage-leadtek/aidmslog/logger"
)

func main() {
	// Example 1: Async File-based logger
	fmt.Println("=== Async File Backend Example ===")
	asyncFileConfig := logger.Config{
		Backend: logger.BackendFile,
		BackendConfig: logger.FileConfig{
			FilePath:      "./logs/async_app.log",
			MaxFileSizeMB: 10,
		},
		Async:        true, // Enable async mode
		DefaultLevel: logger.LevelInfo,
	}

	asyncFileLm, err := logger.NewLogManager(asyncFileConfig)
	if err != nil {
		log.Fatalf("Failed to create async file log manager: %v", err)
	}
	defer asyncFileLm.Close()

	// High-volume logging test
	fmt.Println("Starting high-volume async logging test...")
	start := time.Now()

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				asyncFileLm.WriteLog(logger.LevelInfo, fmt.Sprintf("Goroutine %d - Message %d", id, j))
			}
		}(i)
	}

	wg.Wait()
	elapsed := time.Since(start)
	fmt.Printf("✅ Async logging completed 1000 messages in %v\n", elapsed)

	// Give time for async writes to complete
	time.Sleep(100 * time.Millisecond)

	// Example 2: Sync vs Async comparison
	fmt.Println("\n=== Sync vs Async Comparison ===")

	// Sync logger
	syncConfig := logger.Config{
		Backend: logger.BackendFile,
		BackendConfig: logger.FileConfig{
			FilePath:      "./logs/sync_app.log",
			MaxFileSizeMB: 10,
		},
		Async: false, // Sync mode
	}

	syncLm, err := logger.NewLogManager(syncConfig)
	if err != nil {
		log.Fatalf("Failed to create sync log manager: %v", err)
	}
	defer syncLm.Close()

	// Benchmark sync
	start = time.Now()
	for i := 0; i < 1000; i++ {
		syncLm.WriteLog(logger.LevelInfo, fmt.Sprintf("Sync message %d", i))
	}
	syncElapsed := time.Since(start)
	fmt.Printf("Sync logging: 1000 messages in %v\n", syncElapsed)

	// Benchmark async
	start = time.Now()
	for i := 0; i < 1000; i++ {
		asyncFileLm.WriteLog(logger.LevelInfo, fmt.Sprintf("Async message %d", i))
	}
	asyncElapsed := time.Since(start)
	fmt.Printf("Async logging: 1000 messages in %v\n", asyncElapsed)
	fmt.Printf("Speedup: %.2fx faster\n", float64(syncElapsed)/float64(asyncElapsed))

	// Example 3: Using with custom handler
	fmt.Println("\n=== Async Logger with Custom Handler ===")

	// Create custom handler
	customHandler := &CustomLogHandler{}
	asyncFileLm.RegisterLogHandler(customHandler)

	asyncFileLm.WriteLog(logger.LevelInfo, "Message with custom handler")
	asyncFileLm.WriteLog(logger.LevelError, "Error with custom handler")
	asyncFileLm.WriteLog(logger.LevelWarn, "Warning with custom handler")

	// Wait for async processing
	time.Sleep(100 * time.Millisecond)

	fmt.Println("✅ All examples completed")
}

// CustomLogHandler is an example log handler
type CustomLogHandler struct{}

func (h *CustomLogHandler) Handle(entry logger.LogEntry) error {
	if entry.Level == logger.LevelError {
		fmt.Printf("🚨 ALERT: Error logged at %s: %s\n",
			entry.Timestamp.Format(time.RFC3339), entry.Message)
	}
	return nil
}
