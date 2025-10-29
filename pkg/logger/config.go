// pkg/logger/config.go
package logger

// BackendType defines the storage backend for logs
type BackendType string

const (
	BackendFile BackendType = "file"
	BackendSQL  BackendType = "sql"
)

// Config is the main configuration for LogManager
type Config struct {
	Backend       BackendType
	BackendConfig interface{} // FileConfig or SQLConfig

	// Common settings
	Async        bool
	DefaultLevel LogLevel
}

// FileConfig contains file backend specific settings
type FileConfig struct {
	FilePath      string
	MaxFileSizeMB int
}

// SQLConfig contains SQL backend specific settings
type SQLConfig struct {
	DSN       string // e.g., "user:password@tcp(localhost:3306)/dbname"
	TableName string // default: "logs"
	Driver    string // "mysql", "postgres", "sqlite"
}

// DefaultFileConfig returns default file configuration
func DefaultFileConfig() FileConfig {
	return FileConfig{
		FilePath:      "./logs/app.log",
		MaxFileSizeMB: 10,
	}
}

// DefaultSQLConfig returns default SQL configuration
func DefaultSQLConfig(dsn string) SQLConfig {
	return SQLConfig{
		DSN:       dsn,
		TableName: "logs",
		Driver:    "mysql",
	}
}
