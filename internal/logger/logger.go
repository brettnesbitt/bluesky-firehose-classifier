package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

var log zerolog.Logger

func init() {
	// Set Global Log Level
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	// Create a console writer
	consoleWriter := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}

	// Create a logger with the console writer
	log = zerolog.New(consoleWriter).With().Timestamp().Logger()
}

// Info logs an Info level message.
func Info(msg string, fields ...interface{}) {
	log.Info().Msg(msg)
}

// Error logs an Error level message.
func Error(msg string, err error, fields ...interface{}) {
	log.Error().Err(err).Msg(msg)
}

// Debug logs a Debug level message.
func Debug(msg string, fields ...interface{}) {
	log.Debug().Msg(msg)
}

// Fatal logs a Fatal level message and exits the program.
func Fatal(msg string, err error, fields ...interface{}) {
	log.Fatal().Err(err).Msg(msg)
	os.Exit(1) // Exit the program with a non-zero exit code
}
