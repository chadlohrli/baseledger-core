package common

import (
	"os"

	logger "github.com/kthomas/go-logger"
)

const defaultLoggerName = "baseledger-consensus"
const defaultLogLevel = "INFO"

var (
	// Log is the configured logger
	Log      *logger.Logger
	LogLevel string
)

func init() {
	requireLogger()
}

func requireLogger() {
	LogLevel = os.Getenv("LOG_LEVEL")
	if LogLevel == "" {
		LogLevel = defaultLogLevel
	}

	var endpoint *string
	if os.Getenv("SYSLOG_ENDPOINT") != "" {
		endpt := os.Getenv("SYSLOG_ENDPOINT")
		endpoint = &endpt
	}

	Log = logger.NewLogger(defaultLoggerName, LogLevel, endpoint)
}
