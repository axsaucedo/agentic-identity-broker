// Package config contains domain-specific configuration types and errors.
package config

import "fmt"

// LogLevel is an enumeration of valid log levels.
type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
)

// Validate checks if the log level is valid.
// Returns an error if the log level is not one of the defined constants.
func (l LogLevel) Validate() error {
	switch l {
	case LogLevelDebug, LogLevelInfo, LogLevelWarn, LogLevelError:
		return nil
	default:
		return fmt.Errorf("invalid log level: %s, must be one of [debug, info, warn, error]", l)
	}
}

// String returns the string representation of the log level.
func (l LogLevel) String() string {
	return string(l)
}

// LogFormat is an enumeration of valid log output formats.
type LogFormat string

const (
	LogFormatText LogFormat = "text"
	LogFormatJSON LogFormat = "json"
)

// Validate checks if the log format is valid.
// Returns an error if the log format is not one of the defined constants.
func (f LogFormat) Validate() error {
	switch f {
	case LogFormatText, LogFormatJSON:
		return nil
	default:
		return fmt.Errorf("invalid log format: %s, must be one of [text, json]", f)
	}
}

// String returns the string representation of the log format.
func (f LogFormat) String() string {
	return string(f)
}
