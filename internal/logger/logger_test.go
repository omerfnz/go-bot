package logger

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew_ConsoleOnly(t *testing.T) {
	config := Config{
		Level:      InfoLevel,
		EnableFile: false,
	}

	logger, err := New(config)
	require.NoError(t, err)
	require.NotNil(t, logger)

	assert.Equal(t, logrus.InfoLevel, logger.GetLevel())
}

func TestNew_WithFile(t *testing.T) {
	// Create temp directory for test
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "test.log")

	config := Config{
		Level:      DebugLevel,
		LogFile:    logFile,
		EnableFile: true,
	}

	logger, err := New(config)
	require.NoError(t, err)
	require.NotNil(t, logger)
	defer logger.Close()

	assert.Equal(t, logrus.DebugLevel, logger.GetLevel())

	// Write a log message
	logger.Info("test message")

	// Check if log file was created
	_, err = os.Stat(logFile)
	assert.NoError(t, err, "log file should be created")
}

func TestNew_CreateLogDirectory(t *testing.T) {
	// Create temp directory
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "nested", "dir", "test.log")

	config := Config{
		Level:      InfoLevel,
		LogFile:    logFile,
		EnableFile: true,
	}

	logger, err := New(config)
	require.NoError(t, err)
	require.NotNil(t, logger)
	defer logger.Close()

	// Check if nested directories were created
	logDir := filepath.Dir(logFile)
	info, err := os.Stat(logDir)
	assert.NoError(t, err, "nested directories should be created")
	assert.True(t, info.IsDir())
}

func TestNew_InvalidLogLevel(t *testing.T) {
	config := Config{
		Level:      LogLevel("invalid"),
		EnableFile: false,
	}

	logger, err := New(config)
	assert.Error(t, err)
	assert.Nil(t, logger)
	assert.Contains(t, err.Error(), "invalid log level")
}

func TestNew_InvalidLogFilePath(t *testing.T) {
	// Use an invalid path (directory without write permissions on Windows is tricky,
	// so we'll use a path that can't be created)
	config := Config{
		Level:      InfoLevel,
		LogFile:    string([]byte{0}), // invalid path
		EnableFile: true,
	}

	logger, err := New(config)
	assert.Error(t, err)
	assert.Nil(t, logger)
}

func TestParseLogLevel_Debug(t *testing.T) {
	level, err := parseLogLevel(DebugLevel)
	assert.NoError(t, err)
	assert.Equal(t, logrus.DebugLevel, level)
}

func TestParseLogLevel_Info(t *testing.T) {
	level, err := parseLogLevel(InfoLevel)
	assert.NoError(t, err)
	assert.Equal(t, logrus.InfoLevel, level)
}

func TestParseLogLevel_Warn(t *testing.T) {
	level, err := parseLogLevel(WarnLevel)
	assert.NoError(t, err)
	assert.Equal(t, logrus.WarnLevel, level)
}

func TestParseLogLevel_Error(t *testing.T) {
	level, err := parseLogLevel(ErrorLevel)
	assert.NoError(t, err)
	assert.Equal(t, logrus.ErrorLevel, level)
}

func TestParseLogLevel_Invalid(t *testing.T) {
	level, err := parseLogLevel(LogLevel("invalid"))
	assert.Error(t, err)
	assert.Equal(t, logrus.InfoLevel, level) // Should return default
	assert.Contains(t, err.Error(), "unknown log level")
}

func TestNewDefault(t *testing.T) {
	logger := NewDefault()
	require.NotNil(t, logger)
	assert.Equal(t, logrus.InfoLevel, logger.GetLevel())
}

func TestLogLevels_AllLevels(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name  string
		level LogLevel
		want  logrus.Level
	}{
		{"debug", DebugLevel, logrus.DebugLevel},
		{"info", InfoLevel, logrus.InfoLevel},
		{"warn", WarnLevel, logrus.WarnLevel},
		{"error", ErrorLevel, logrus.ErrorLevel},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logFile := filepath.Join(tempDir, tt.name+".log")
			config := Config{
				Level:      tt.level,
				LogFile:    logFile,
				EnableFile: true,
			}

			logger, err := New(config)
			require.NoError(t, err)
			defer logger.Close()
			assert.Equal(t, tt.want, logger.GetLevel())
		})
	}
}

func TestLogger_LogOutput(t *testing.T) {
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "output.log")

	config := Config{
		Level:      InfoLevel,
		LogFile:    logFile,
		EnableFile: true,
	}

	logger, err := New(config)
	require.NoError(t, err)
	defer logger.Close()

	// Write various log levels
	logger.Debug("debug message")  // Should not appear (level is Info)
	logger.Info("info message")    // Should appear
	logger.Warn("warning message") // Should appear
	logger.Error("error message")  // Should appear

	// Read log file
	content, err := os.ReadFile(logFile)
	require.NoError(t, err)

	logContent := string(content)

	// Debug should not be in the log (level is Info)
	assert.NotContains(t, logContent, "debug message")

	// Others should be present
	assert.Contains(t, logContent, "info message")
	assert.Contains(t, logContent, "warning message")
	assert.Contains(t, logContent, "error message")
}

// TestLogger_Close tests the Close functionality
func TestLogger_Close(t *testing.T) {
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "close_test.log")

	config := Config{
		Level:      InfoLevel,
		LogFile:    logFile,
		EnableFile: true,
	}

	logger, err := New(config)
	require.NoError(t, err)

	// Close should work
	err = logger.Close()
	assert.NoError(t, err)

	// Multiple closes should be safe
	err = logger.Close()
	assert.NoError(t, err)
}

// TestLogger_CloseConsoleOnly tests Close with console-only logger
func TestLogger_CloseConsoleOnly(t *testing.T) {
	config := Config{
		Level:      InfoLevel,
		EnableFile: false,
	}

	logger, err := New(config)
	require.NoError(t, err)

	// Close should work even without file
	err = logger.Close()
	assert.NoError(t, err)
}
