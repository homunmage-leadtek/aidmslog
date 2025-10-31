// pkg/logger/log_manager.go

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
}

// NewLogManager creates a new LogManager with the given configuration
func NewLogManager(config Config) (LogManager, error) {
	lm := &logManagerImpl{
		config: config,
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
	return lm, nil
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
	if lm.backend == nil {
		return nil
	}
	return lm.backend.Close()
}
