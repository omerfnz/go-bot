package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create a temporary config file
func createTempConfig(t *testing.T, config *Config) string {
	t.Helper()
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")

	data, err := json.Marshal(config)
	require.NoError(t, err)

	err = os.WriteFile(configPath, data, 0644)
	require.NoError(t, err)

	return configPath
}

// Helper function to create a valid test config
func createValidConfig() *Config {
	return &Config{
		Headless: true,
		Workers:  5,
		Interval: 300,
		Keywords: []Keyword{
			{Term: "golang tutorial", TargetURL: "example.com"},
		},
		Proxies:       []string{"http://proxy1.com:8080"},
		PageTimeout:   30,
		SearchTimeout: 15,
		MaxRetries:    3,
		RetryDelay:    5,
		Selectors: SelectorConfig{
			SearchBox:    "input[name='q']",
			SearchButton: "button[type='submit']",
			ResultItem:   "div.result",
			ResultLink:   "a",
			NextButton:   "a.next",
		},
	}
}

func TestLoad_ValidConfig(t *testing.T) {
	validConfig := createValidConfig()
	configPath := createTempConfig(t, validConfig)

	config, err := Load(configPath)
	require.NoError(t, err)
	require.NotNil(t, config)

	assert.Equal(t, validConfig.Headless, config.Headless)
	assert.Equal(t, validConfig.Workers, config.Workers)
	assert.Equal(t, validConfig.Interval, config.Interval)
	assert.Len(t, config.Keywords, 1)
	assert.Equal(t, "golang tutorial", config.Keywords[0].Term)
	assert.Equal(t, "example.com", config.Keywords[0].TargetURL)
}

func TestLoad_FileNotFound(t *testing.T) {
	config, err := Load("nonexistent.json")
	assert.Error(t, err)
	assert.Nil(t, config)
	assert.Contains(t, err.Error(), "failed to read config file")
}

func TestLoad_InvalidJSON(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "invalid.json")

	err := os.WriteFile(configPath, []byte("{ invalid json }"), 0644)
	require.NoError(t, err)

	config, err := Load(configPath)
	assert.Error(t, err)
	assert.Nil(t, config)
	assert.Contains(t, err.Error(), "failed to parse config JSON")
}

func TestLoad_InvalidConfig(t *testing.T) {
	invalidConfig := &Config{
		Workers: -1, // Invalid
	}
	configPath := createTempConfig(t, invalidConfig)

	config, err := Load(configPath)
	assert.Error(t, err)
	assert.Nil(t, config)
	assert.Contains(t, err.Error(), "invalid configuration")
}

func TestValidate_ValidConfig(t *testing.T) {
	config := createValidConfig()
	err := config.Validate()
	assert.NoError(t, err)
}

func TestValidate_WorkersTooLow(t *testing.T) {
	config := createValidConfig()
	config.Workers = 0
	err := config.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "workers must be at least 1")
}

func TestValidate_WorkersTooHigh(t *testing.T) {
	config := createValidConfig()
	config.Workers = 101
	err := config.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "workers cannot exceed 100")
}

func TestValidate_NegativeInterval(t *testing.T) {
	config := createValidConfig()
	config.Interval = -1
	err := config.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "interval must be non-negative")
}

func TestValidate_NoKeywords(t *testing.T) {
	config := createValidConfig()
	config.Keywords = []Keyword{}
	err := config.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least one keyword is required")
}

func TestValidate_EmptyKeywordTerm(t *testing.T) {
	config := createValidConfig()
	config.Keywords = []Keyword{
		{Term: "", TargetURL: "example.com"},
	}
	err := config.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "term cannot be empty")
}

func TestValidate_EmptyKeywordTargetURL(t *testing.T) {
	config := createValidConfig()
	config.Keywords = []Keyword{
		{Term: "golang", TargetURL: ""},
	}
	err := config.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "target_url cannot be empty")
}

func TestValidate_EmptyProxy(t *testing.T) {
	config := createValidConfig()
	config.Proxies = []string{""}
	err := config.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "proxy URL cannot be empty")
}

func TestValidate_PageTimeoutTooLow(t *testing.T) {
	config := createValidConfig()
	config.PageTimeout = 0
	err := config.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "page_timeout must be at least 1 second")
}

func TestValidate_SearchTimeoutTooLow(t *testing.T) {
	config := createValidConfig()
	config.SearchTimeout = 0
	err := config.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "search_timeout must be at least 1 second")
}

func TestValidate_NegativeMaxRetries(t *testing.T) {
	config := createValidConfig()
	config.MaxRetries = -1
	err := config.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "max_retries must be non-negative")
}

func TestValidate_NegativeRetryDelay(t *testing.T) {
	config := createValidConfig()
	config.RetryDelay = -1
	err := config.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "retry_delay must be non-negative")
}

func TestValidate_EmptySearchBoxSelector(t *testing.T) {
	config := createValidConfig()
	config.Selectors.SearchBox = ""
	err := config.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "selectors.search_box cannot be empty")
}

func TestValidate_EmptyResultItemSelector(t *testing.T) {
	config := createValidConfig()
	config.Selectors.ResultItem = ""
	err := config.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "selectors.result_item cannot be empty")
}

func TestLoadEnv_OverrideValues(t *testing.T) {
	config := createValidConfig()

	// Set environment variables
	os.Setenv("HEADLESS", "false")
	os.Setenv("WORKERS", "10")
	os.Setenv("INTERVAL", "600")
	os.Setenv("PAGE_TIMEOUT", "60")
	os.Setenv("SEARCH_TIMEOUT", "30")
	os.Setenv("MAX_RETRIES", "5")
	os.Setenv("RETRY_DELAY", "10")
	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("LOG_FILE", "test.log")

	defer func() {
		os.Unsetenv("HEADLESS")
		os.Unsetenv("WORKERS")
		os.Unsetenv("INTERVAL")
		os.Unsetenv("PAGE_TIMEOUT")
		os.Unsetenv("SEARCH_TIMEOUT")
		os.Unsetenv("MAX_RETRIES")
		os.Unsetenv("RETRY_DELAY")
		os.Unsetenv("LOG_LEVEL")
		os.Unsetenv("LOG_FILE")
	}()

	err := config.LoadEnv()
	require.NoError(t, err)

	assert.False(t, config.Headless)
	assert.Equal(t, 10, config.Workers)
	assert.Equal(t, 600, config.Interval)
	assert.Equal(t, 60, config.PageTimeout)
	assert.Equal(t, 30, config.SearchTimeout)
	assert.Equal(t, 5, config.MaxRetries)
	assert.Equal(t, 10, config.RetryDelay)
	assert.Equal(t, "debug", config.LogLevel)
	assert.Equal(t, "test.log", config.LogFile)
}

func TestLoadEnv_InvalidBoolValue(t *testing.T) {
	config := createValidConfig()
	originalHeadless := config.Headless

	os.Setenv("HEADLESS", "invalid")
	defer os.Unsetenv("HEADLESS")

	err := config.LoadEnv()
	require.NoError(t, err)

	// Should keep original value when parsing fails
	assert.Equal(t, originalHeadless, config.Headless)
}

func TestLoadEnv_InvalidIntValue(t *testing.T) {
	config := createValidConfig()
	originalWorkers := config.Workers

	os.Setenv("WORKERS", "invalid")
	defer os.Unsetenv("WORKERS")

	err := config.LoadEnv()
	require.NoError(t, err)

	// Should keep original value when parsing fails
	assert.Equal(t, originalWorkers, config.Workers)
}

func TestLoadEnv_InvalidAfterOverride(t *testing.T) {
	config := createValidConfig()

	// Set invalid value
	os.Setenv("WORKERS", "0")
	defer os.Unsetenv("WORKERS")

	err := config.LoadEnv()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid configuration after env override")
}

func TestLoadEnv_NoEnvFile(t *testing.T) {
	config := createValidConfig()
	err := config.LoadEnv()
	assert.NoError(t, err) // Should not fail if .env doesn't exist
}

func TestLoadWithEnv_Success(t *testing.T) {
	validConfig := createValidConfig()
	configPath := createTempConfig(t, validConfig)

	os.Setenv("WORKERS", "20")
	defer os.Unsetenv("WORKERS")

	config, err := LoadWithEnv(configPath)
	require.NoError(t, err)
	require.NotNil(t, config)

	assert.Equal(t, 20, config.Workers)
}

func TestLoadWithEnv_ConfigFileError(t *testing.T) {
	os.Setenv("WORKERS", "20")
	defer os.Unsetenv("WORKERS")

	config, err := LoadWithEnv("nonexistent.json")
	assert.Error(t, err)
	assert.Nil(t, config)
}

func TestSetDefaults(t *testing.T) {
	config := &Config{
		Keywords: []Keyword{
			{Term: "test", TargetURL: "example.com"},
		},
		Selectors: SelectorConfig{
			SearchBox:  "input",
			ResultItem: "div",
		},
	}

	config.SetDefaults()

	assert.Equal(t, 30, config.PageTimeout)
	assert.Equal(t, 15, config.SearchTimeout)
	assert.Equal(t, 3, config.MaxRetries)
	assert.Equal(t, 5, config.RetryDelay)
	assert.Equal(t, 5, config.Workers)
	assert.Equal(t, 300, config.Interval)
	assert.Equal(t, "info", config.LogLevel)
}

func TestSetDefaults_DoesNotOverrideExisting(t *testing.T) {
	config := &Config{
		PageTimeout:   60,
		SearchTimeout: 30,
		MaxRetries:    10,
		RetryDelay:    15,
		Workers:       20,
		Interval:      600,
		LogLevel:      "debug",
		Keywords: []Keyword{
			{Term: "test", TargetURL: "example.com"},
		},
		Selectors: SelectorConfig{
			SearchBox:  "input",
			ResultItem: "div",
		},
	}

	config.SetDefaults()

	// Should not override existing non-zero values
	assert.Equal(t, 60, config.PageTimeout)
	assert.Equal(t, 30, config.SearchTimeout)
	assert.Equal(t, 10, config.MaxRetries)
	assert.Equal(t, 15, config.RetryDelay)
	assert.Equal(t, 20, config.Workers)
	assert.Equal(t, 600, config.Interval)
	assert.Equal(t, "debug", config.LogLevel)
}

func TestConfig_MultipleKeywords(t *testing.T) {
	config := createValidConfig()
	config.Keywords = []Keyword{
		{Term: "golang", TargetURL: "example1.com"},
		{Term: "go programming", TargetURL: "example2.com"},
		{Term: "go tutorial", TargetURL: "example3.com"},
	}

	err := config.Validate()
	assert.NoError(t, err)
}

func TestConfig_NoProxies(t *testing.T) {
	config := createValidConfig()
	config.Proxies = []string{} // Empty proxy list should be valid

	err := config.Validate()
	assert.NoError(t, err)
}

func TestConfig_MultipleProxies(t *testing.T) {
	config := createValidConfig()
	config.Proxies = []string{
		"http://proxy1.com:8080",
		"http://proxy2.com:8080",
		"socks5://proxy3.com:1080",
	}

	err := config.Validate()
	assert.NoError(t, err)
}
