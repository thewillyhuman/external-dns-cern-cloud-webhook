// Package log provides a zerolog implementation of the Logger interface.
package log

import (
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
)

// ZeroLogger is an adapter that wraps a zerolog.Logger to satisfy the Logger interface.
//
// This struct is the bridge between the application's logging interface and the zerolog library.
// It holds an instance of a zerolog.Logger and delegates the logging calls to it.
// This approach was chosen to keep the rest of the application completely decoupled from the
// specifics of the zerolog library.
type ZeroLogger struct {
	logger zerolog.Logger
}

// NewLogger creates a new Logger implementation that uses zerolog as the backend.
//
// This function initializes a new zerolog.Logger with a specific log level.
// It is configured to write to the console with a human-readable format and includes timestamps.
// The choice of zerolog was based on its performance and structured logging capabilities,
// which are well-suited for a production environment.
func NewLogger(level Level) Logger {
	// Parse the application's log level into a zerolog-compatible level.
	loggerLevel, err := zerolog.ParseLevel(LevelNames[level])
	if err != nil {
		// If the log level is invalid, print an error and continue.
		// This is a rare case that should only happen if the LevelNames map is out of sync.
		fmt.Printf("Error creating logger with level %s\n", LevelNames[level])
	}

	// Create a new zerolog.Logger instance.
	// We use a ConsoleWriter for human-readable output during development.
	// For a production environment, this could be switched to a JSON writer for easier parsing by log aggregators.
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}).
		Level(loggerLevel).
		With().
		Timestamp().
		Logger()

	// Return a new ZeroLogger instance that wraps the configured zerolog.Logger.
	return &ZeroLogger{logger: logger}
}

// Debug logs a formatted message at the Debug level.
// It uses the Msgf method of the underlying zerolog.Logger to format the message.
func (z *ZeroLogger) Debug(format string, args ...any) {
	z.logger.Debug().Msgf(format, args...)
}

// Info logs a formatted message at the Info level.
// It uses the Msgf method of the underlying zerolog.Logger to format the message.
func (z *ZeroLogger) Info(format string, args ...any) {
	z.logger.Info().Msgf(format, args...)
}

// Warn logs a formatted message at the Warn level.
// It uses the Msgf method of the underlying zerolog.Logger to format the message.
func (z *ZeroLogger) Warn(format string, args ...any) {
	z.logger.Warn().Msgf(format, args...)
}

// Error logs a formatted message at the Error level.
// It uses the Msgf method of the underlying zerolog.Logger to format the message.
func (z *ZeroLogger) Error(format string, args ...any) {
	z.logger.Error().Msgf(format, args...)
}
