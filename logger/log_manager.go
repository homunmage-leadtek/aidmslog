// /logger/log_manager.go

package logger

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

type logManagerImpl struct {
	mu       sync.Mutex
	config   Config
	backend  LogBackend
	handlers []LogHandler

	// Async support
	logChannel chan LogEntry
	done       chan struct{}
	wg         sync.WaitGroup
	isAsync    bool
}

// NewLogManager creates a new LogManager with the given configuration
func NewLogManager(config Config) (LogManager, error) {
	lm := &logManagerImpl{
		config:  config,
		isAsync: config.Async,
	}

	// Create backend based on type
	var backend LogBackend
	var err error

	switch config.Backend {
	case BackendFile:
		backend = &FileBackend{}
	case BackendSQL:
		backend = &SQLBackend{}
	default:
		return nil, fmt.Errorf("unsupported backend type: %s", config.Backend)
	}

	// Initialize backend with its specific config
	if err = backend.Init(config.BackendConfig); err != nil {
		return nil, fmt.Errorf("failed to initialize backend: %w", err)
	}

	lm.backend = backend

	// Start async worker if enabled
	if lm.isAsync {
		lm.logChannel = make(chan LogEntry, 1000) // Buffer size of 1000
		lm.done = make(chan struct{})
		lm.startAsyncWorker()
	}

	return lm, nil
}

// startAsyncWorker starts the background goroutine for async logging
func (lm *logManagerImpl) startAsyncWorker() {
	lm.wg.Add(1)
	go func() {
		defer lm.wg.Done()
		for {
			select {
			case entry := <-lm.logChannel:
				// Write to backend
				if err := lm.backend.Write(entry); err != nil {
					// In production, you might want to handle this error better
					// For now, we'll just continue to avoid blocking
					fmt.Printf("async log write error: %v\n", err)
				}

				// Notify handlers
				lm.mu.Lock()
				handlers := make([]LogHandler, len(lm.handlers))
				copy(handlers, lm.handlers)
				lm.mu.Unlock()

				for _, h := range handlers {
					_ = h.Handle(entry)
				}

			case <-lm.done:
				// Drain remaining logs before exiting
				for {
					select {
					case entry := <-lm.logChannel:
						_ = lm.backend.Write(entry)

						lm.mu.Lock()
						handlers := make([]LogHandler, len(lm.handlers))
						copy(handlers, lm.handlers)
						lm.mu.Unlock()

						for _, h := range handlers {
							_ = h.Handle(entry)
						}
					default:
						return
					}
				}
			}
		}
	}()
}

func (lm *logManagerImpl) WriteLog(level LogLevel, message string) error {
	if lm.backend == nil {
		return errors.New("backend not initialized")
	}

	entry := LogEntry{
		Level:     level,
		Message:   message,
		Timestamp: time.Now(),
	}

	if lm.isAsync {
		// Async mode: send to channel
		select {
		case lm.logChannel <- entry:
			return nil
		case <-time.After(100 * time.Millisecond):
			return fmt.Errorf("log channel is full, log may be dropped")
		}
	} else {
		// Sync mode: write immediately
		if err := lm.backend.Write(entry); err != nil {
			return fmt.Errorf("failed to write log: %w", err)
		}

		// Notify handlers
		lm.mu.Lock()
		handlers := lm.handlers
		lm.mu.Unlock()

		for _, h := range handlers {
			_ = h.Handle(entry)
		}

		return nil
	}
}

func (lm *logManagerImpl) ReadLogs(level LogLevel, filter LogFilter) ([]LogEntry, error) {
	if lm.backend == nil {
		return nil, errors.New("backend not initialized")
	}
	return lm.backend.Read(level, filter)
}

func (lm *logManagerImpl) ClearLogs(before time.Time) error {
	if lm.backend == nil {
		return errors.New("backend not initialized")
	}
	return lm.backend.ClearLogs(before)
}

func (lm *logManagerImpl) RegisterLogHandler(handler LogHandler) {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	lm.handlers = append(lm.handlers, handler)
}

func (lm *logManagerImpl) Close() error {
	// Stop async worker if running
	if lm.isAsync && lm.done != nil {
		close(lm.done)
		lm.wg.Wait() // Wait for worker to finish processing remaining logs
		close(lm.logChannel)
	}

	if lm.backend == nil {
		return nil
	}
	return lm.backend.Close()
}
