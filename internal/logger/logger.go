// Package logger provides structured logging functionality for the application.
// It supports both console and file logging with configurable log levels.
package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

// LogLevel represents the logging level
type LogLevel string

const (
	// DebugLevel logs are typically voluminous, and are usually disabled in production.
	DebugLevel LogLevel = "debug"
	// InfoLevel is the default logging priority.
	InfoLevel LogLevel = "info"
	// WarnLevel logs are more important than Info, but don't need individual human review.
	WarnLevel LogLevel = "warn"
	// ErrorLevel logs are high-priority. If an application is running smoothly,
	// it shouldn't generate any error-level logs.
	ErrorLevel LogLevel = "error"
)

// Config holds the configuration for the logger
type Config struct {
	Level      LogLevel // Log level (debug, info, warn, error)
	LogFile    string   // Path to log file (empty for console only)
	EnableFile bool     // Enable file logging
}

// Logger wraps logrus.Logger and provides cleanup functionality
type Logger struct {
	*logrus.Logger
	logFile *os.File // Log file handle for cleanup
}

// New creates a new logger with the given configuration.
// It returns a configured Logger instance that wraps logrus.Logger.
// Remember to call Close() when done to properly close the log file.
//
// Example:
//
//	logger, err := New(Config{
//	    Level: InfoLevel,
//	    LogFile: "logs/app.log",
//	    EnableFile: true,
//	})
//	if err != nil {
//	    panic(err)
//	}
//	defer logger.Close()
func New(config Config) (*Logger, error) {
	logrusLogger := logrus.New()

	// Set log level
	level, err := parseLogLevel(config.Level)
	if err != nil {
		return nil, fmt.Errorf("invalid log level: %w", err)
	}
	logrusLogger.SetLevel(level)

	// Set formatter
	logrusLogger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
		ForceColors:     true,
	})

	wrappedLogger := &Logger{
		Logger:  logrusLogger,
		logFile: nil,
	}

	// Configure output
	if config.EnableFile && config.LogFile != "" {
		// Create log directory if it doesn't exist
		logDir := filepath.Dir(config.LogFile)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create log directory: %w", err)
		}

		// Open log file
		file, err := os.OpenFile(config.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}

		wrappedLogger.logFile = file

		// Write to both file and console
		multiWriter := io.MultiWriter(os.Stdout, file)
		logrusLogger.SetOutput(multiWriter)
	} else {
		// Console only
		logrusLogger.SetOutput(os.Stdout)
	}

	return wrappedLogger, nil
}

// Close closes the log file if it was opened.
// It's safe to call Close multiple times.
func (l *Logger) Close() error {
	if l.logFile != nil {
		err := l.logFile.Close()
		l.logFile = nil
		return err
	}
	return nil
}

// parseLogLevel converts string log level to logrus.Level
func parseLogLevel(level LogLevel) (logrus.Level, error) {
	switch level {
	case DebugLevel:
		return logrus.DebugLevel, nil
	case InfoLevel:
		return logrus.InfoLevel, nil
	case WarnLevel:
		return logrus.WarnLevel, nil
	case ErrorLevel:
		return logrus.ErrorLevel, nil
	default:
		return logrus.InfoLevel, fmt.Errorf("unknown log level: %s", level)
	}
}

// NewDefault creates a logger with default configuration (Info level, console only)
func NewDefault() *Logger {
	logger, _ := New(Config{
		Level:      InfoLevel,
		EnableFile: false,
	})
	return logger
}
