// pkg/logger/interface.go

package logger

import "time"

// LogLevel defines log severity levels
type LogLevel string

const (
	LevelDebug LogLevel = "DEBUG"
	LevelInfo  LogLevel = "INFO"
	LevelWarn  LogLevel = "WARN"
	LevelError LogLevel = "ERROR"
)

// LogEntry represents a single log record
type LogEntry struct {
	Level     LogLevel
	Message   string
	Timestamp time.Time
	Metadata  map[string]interface{}
}

// LogFilter provides filtering criteria for reading logs
type LogFilter struct {
	StartTime *time.Time
	EndTime   *time.Time
	Contains  string
}

// LogHandler allows extension (e.g., sending logs to external systems)
type LogHandler interface {
	Handle(entry LogEntry) error
}

// LogBackend is the unified interface that all backends must implement
type LogBackend interface {
	// Init initializes the backend with given configuration
	Init(config interface{}) error

	// Write writes a log entry
	Write(entry LogEntry) error

	// Read retrieves logs based on filter
	Read(level LogLevel, filter LogFilter) ([]LogEntry, error)

	// ClearLogs removes logs older than specified time
	ClearLogs(before time.Time) error

	// Close releases resources
	Close() error
}

// LogManager is the core interface for managing logs
type LogManager interface {
	WriteLog(level LogLevel, message string) error
	ReadLogs(level LogLevel, filter LogFilter) ([]LogEntry, error)
	ClearLogs(before time.Time) error
	RegisterLogHandler(handler LogHandler)
	Close() error
}
