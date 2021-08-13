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

// PanicIfEmpty panics if the given string is empty
func PanicIfEmpty(val string, msg string) {
	if val == "" {
		panic(msg)
	}
}

// StringOrNil returns the given string or nil when empty
func StringOrNil(str string) *string {
	if str == "" {
		return nil
	}
	return &str
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
