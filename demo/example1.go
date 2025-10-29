package main

import (
	"fmt"
	"log"

	"github.com/homunmage-leadtek/aidmslog/pkg/logger"
)

func main() {
	// Example 1: File-based logger
	fmt.Println("=== File Backend Example ===")
	fileConfig := logger.Config{
		Backend: logger.BackendFile,
		BackendConfig: logger.FileConfig{
			FilePath:      "./logs/app.log",
			MaxFileSizeMB: 10,
		},
	}

	fileLm, err := logger.NewLogManager(fileConfig)
	if err != nil {
		log.Fatalf("Failed to create file log manager: %v", err)
	}
	defer fileLm.Close()

	fileLm.WriteLog(logger.LevelInfo, "File log writer initialized")
	fileLm.WriteLog(logger.LevelError, "Something went wrong!")
	fileLm.WriteLog(logger.LevelDebug, "Debugging message")
	fileLm.WriteLog(logger.LevelWarn, "Low disk space")

	fmt.Println("✅ Logs written to ./logs/app.log")

	// Example 2: SQL-based logger
	fmt.Println("\n=== SQL Backend Example ===")
	sqlConfig := logger.Config{
		Backend: logger.BackendSQL,
		BackendConfig: logger.SQLConfig{
			DSN:       "user:password@tcp(localhost:3306)/mydb",
			TableName: "application_logs",
			Driver:    "mysql",
		},
	}

	sqlLm, err := logger.NewLogManager(sqlConfig)
	if err != nil {
		log.Fatalf("Failed to create SQL log manager: %v", err)
	}
	defer sqlLm.Close()

	sqlLm.WriteLog(logger.LevelInfo, "SQL logger initialized")
	sqlLm.WriteLog(logger.LevelWarn, "Connection pool size exceeded")

	fmt.Println("✅ Logs written to SQL database")

	// Example 3: Using default configs
	fmt.Println("\n=== Using Default Config ===")
	defaultConfig := logger.Config{
		Backend:       logger.BackendFile,
		BackendConfig: logger.DefaultFileConfig(),
		DefaultLevel:  logger.LevelInfo,
	}

	defaultLm, err := logger.NewLogManager(defaultConfig)
	if err != nil {
		log.Fatalf("Failed to create default log manager: %v", err)
	}
	defer defaultLm.Close()

	defaultLm.WriteLog(logger.LevelInfo, "Using default configuration")
	fmt.Println("✅ Logs written with default config")
}
