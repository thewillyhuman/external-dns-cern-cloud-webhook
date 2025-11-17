// Package log defines a flexible, level-based logging interface.
//
// This package provides the core types and contracts for creating and using structured logging.
// The design is centered around the Logger interface, which decouples the application from any specific
// logging implementation. This allows for easy swapping of logging backends (e.g., zerolog, logrus)
// without changing the application code.
//
// The package also defines a set of standard log levels and provides a global logger instance
// for convenience.
package log

import "strings"

// Level defines the severity of a log message.
// Using a custom type for log levels provides type safety and allows for easy extension.
type Level int

// Defines the available log levels for the logger.
// The iota keyword is used to create a set of incrementing integer constants,
// which is a common and efficient way to define enums in Go.
const (
	// LevelDebug is for detailed, diagnostic information, typically only
	// useful during development and debugging.
	LevelDebug Level = iota

	// LevelInfo is for general, informational messages that highlight
	// the progress or state of the application.
	LevelInfo

	// LevelWarn indicates a potential problem or a non-critical event
	// that may require attention.
	LevelWarn

	// LevelError designates a significant error or failure that has occurred
	// and likely requires investigation.
	LevelError

	// DefaultLogLevel represents the fallback log level for all the application.
	// This is used when a log level cannot be determined from the configuration.
	DefaultLogLevel = LevelInfo
)

var (
	// LevelNames is a map of log levels to their string representations.
	// This is useful for parsing log levels from configuration and for printing log levels in a human-readable format.
	LevelNames = map[Level]string{
		LevelDebug: "debug",
		LevelInfo:  "info",
		LevelWarn:  "warn",
		LevelError: "error",
	}

	// GlobalLogger is a global instance of the Logger interface.
	// This is provided for convenience, allowing any part of the application to log messages
	// without needing to explicitly pass a logger instance.
	// It is the responsibility of the main application to initialize this logger.
	GlobalLogger Logger
)

// levelValues is a reverse map of log level names to their Level values.
// This is used for efficient parsing of log levels from strings.
var levelValues = make(map[string]Level)

// init is a special Go function that is executed when the package is initialized.
// This is used to populate the levelValues map, ensuring that it is ready for use
// as soon as the package is loaded.
func init() {
	for level, name := range LevelNames {
		levelValues[name] = level
	}
}

// Logger defines the standard contract for a level-based logger.
//
// Implementations are responsible for handling log message filtering based
// on the logger's configured level and formatting the output.
// The interface-based design was chosen to decouple the application from a specific
// logging library. This allows for greater flexibility and makes it easier to
// switch to a different logging implementation in the future if needed.
type Logger interface {
	// Debug logs a formatted message at the Debug level.
	// Arguments are handled in the manner of fmt.Sprintf.
	Debug(format string, args ...any)

	// Info logs a formatted message at the Info level.
	// Arguments are handled in the manner of fmt.Sprintf.
	Info(format string, args ...any)

	// Warn logs a formatted message at the Warn level.
	// Arguments are handled in the manner of fmt.Sprintf.
	Warn(format string, args ...any)

	// Error logs a formatted message at the Error level.
	// Arguments are handled in the manner of fmt.Sprintf.
	Error(format string, args ...any)
}

// LevelFromString parses a string and returns the corresponding log level.
//
// This function is case-insensitive. If the string does not match any known
// log level, it returns false. This is the preferred way to convert user-provided
// strings (e.g., from configuration) into a log.Level.
func LevelFromString(name string) (Level, bool) {
	level, ok := levelValues[strings.ToLower(name)]
	return level, ok
}
