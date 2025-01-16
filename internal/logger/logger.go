package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

type Logger struct {
	zerolog.Logger
}

func NewLogger() Logger {
	// Set Global Log Level
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	// Create a console writer
	consoleWriter := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}

	// Create a logger with the console writer
	zl := zerolog.New(consoleWriter).With().Timestamp().Logger()
	return Logger{zl}
}

// Info logs an Info level message.
func (l Logger) Info(msg string, fields ...interface{}) {
	l.Logger.Info().Msgf(msg, fields...)
}

// Error logs an Error level message.
func (l Logger) Error(msg string, err error, fields ...interface{}) {
	l.Logger.Error().Err(err).Msgf(msg, fields...)
}

// Debug logs a Debug level message.
func (l Logger) Debug(msg string, fields ...interface{}) {
	l.Logger.Debug().Msgf(msg, fields...)
}

// Fatal logs a Fatal level message and exits the program.
func (l Logger) Fatal(msg string, err error, fields ...interface{}) {
	l.Logger.Fatal().Err(err).Msgf(msg, fields...)
	os.Exit(1) // Exit the program with a non-zero exit code
}

func (l Logger) Trace(msg string, fields ...interface{}) {
	l.Logger.Trace().Msgf(msg, fields...)
}

func (l Logger) Warn(msg string, fields ...interface{}) {
	l.Logger.Warn().Msgf(msg, fields...)
}
