// Package config provides configuration management for the application.
// It supports loading configuration from JSON files and environment variables.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds the application configuration
type Config struct {
	// General settings
	Headless bool `json:"headless" env:"HEADLESS"`
	Workers  int  `json:"workers" env:"WORKERS"`
	Interval int  `json:"interval" env:"INTERVAL"` // in seconds

	// Task settings
	Keywords []Keyword `json:"keywords"`
	Proxies  []string  `json:"proxies"`

	// Timeout settings (in seconds)
	PageTimeout   int `json:"page_timeout" env:"PAGE_TIMEOUT"`
	SearchTimeout int `json:"search_timeout" env:"SEARCH_TIMEOUT"`

	// Retry settings
	MaxRetries int `json:"max_retries" env:"MAX_RETRIES"`
	RetryDelay int `json:"retry_delay" env:"RETRY_DELAY"` // in seconds

	// Selectors
	Selectors SelectorConfig `json:"selectors"`

	// Logging (from env only)
	LogLevel string `env:"LOG_LEVEL"`
	LogFile  string `env:"LOG_FILE"`
}

// Keyword represents a search keyword and its target URL
type Keyword struct {
	Term      string `json:"term"`
	TargetURL string `json:"target_url"`
}

// SelectorConfig holds CSS selectors for web scraping
type SelectorConfig struct {
	SearchBox    string `json:"search_box"`
	SearchButton string `json:"search_button"`
	ResultItem   string `json:"result_item"`
	ResultLink   string `json:"result_link"`
	NextButton   string `json:"next_button"`
}

// Load reads configuration from a JSON file and returns a Config instance.
// It does NOT load environment variables - call LoadEnv() separately if needed.
//
// Example:
//
//	config, err := Load("configs/config.json")
//	if err != nil {
//	    log.Fatal(err)
//	}
func Load(configPath string) (*Config, error) {
	// Read file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse JSON
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config JSON: %w", err)
	}

	// Validate
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// LoadEnv loads environment variables and overrides config values.
// It first tries to load from .env file, then reads from environment.
// Returns error only if .env file exists but cannot be loaded.
func (c *Config) LoadEnv() error {
	// Try to load .env file (ignore error if file doesn't exist)
	_ = godotenv.Load()

	// Override config values from environment variables
	if val := os.Getenv("HEADLESS"); val != "" {
		if b, err := strconv.ParseBool(val); err == nil {
			c.Headless = b
		}
	}

	if val := os.Getenv("WORKERS"); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			c.Workers = i
		}
	}

	if val := os.Getenv("INTERVAL"); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			c.Interval = i
		}
	}

	if val := os.Getenv("PAGE_TIMEOUT"); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			c.PageTimeout = i
		}
	}

	if val := os.Getenv("SEARCH_TIMEOUT"); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			c.SearchTimeout = i
		}
	}

	if val := os.Getenv("MAX_RETRIES"); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			c.MaxRetries = i
		}
	}

	if val := os.Getenv("RETRY_DELAY"); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			c.RetryDelay = i
		}
	}

	// Load env-only settings
	if val := os.Getenv("LOG_LEVEL"); val != "" {
		c.LogLevel = val
	}

	if val := os.Getenv("LOG_FILE"); val != "" {
		c.LogFile = val
	}

	// Validate after env override
	if err := c.Validate(); err != nil {
		return fmt.Errorf("invalid configuration after env override: %w", err)
	}

	return nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// Validate Workers
	if c.Workers < 1 {
		return fmt.Errorf("workers must be at least 1, got %d", c.Workers)
	}
	if c.Workers > 100 {
		return fmt.Errorf("workers cannot exceed 100, got %d", c.Workers)
	}

	// Validate Interval
	if c.Interval < 0 {
		return fmt.Errorf("interval must be non-negative, got %d", c.Interval)
	}

	// Validate Keywords
	if len(c.Keywords) == 0 {
		return fmt.Errorf("at least one keyword is required")
	}

	for i, kw := range c.Keywords {
		if kw.Term == "" {
			return fmt.Errorf("keyword[%d]: term cannot be empty", i)
		}
		if kw.TargetURL == "" {
			return fmt.Errorf("keyword[%d]: target_url cannot be empty", i)
		}
	}

	// Validate Proxies (optional, but if provided must not be empty strings)
	for i, proxy := range c.Proxies {
		if proxy == "" {
			return fmt.Errorf("proxy[%d]: proxy URL cannot be empty", i)
		}
	}

	// Validate Timeouts
	if c.PageTimeout < 1 {
		return fmt.Errorf("page_timeout must be at least 1 second, got %d", c.PageTimeout)
	}
	if c.SearchTimeout < 1 {
		return fmt.Errorf("search_timeout must be at least 1 second, got %d", c.SearchTimeout)
	}

	// Validate Retry settings
	if c.MaxRetries < 0 {
		return fmt.Errorf("max_retries must be non-negative, got %d", c.MaxRetries)
	}
	if c.RetryDelay < 0 {
		return fmt.Errorf("retry_delay must be non-negative, got %d", c.RetryDelay)
	}

	// Validate Selectors
	if c.Selectors.SearchBox == "" {
		return fmt.Errorf("selectors.search_box cannot be empty")
	}
	if c.Selectors.ResultItem == "" {
		return fmt.Errorf("selectors.result_item cannot be empty")
	}

	return nil
}

// LoadWithEnv is a convenience function that loads config from file
// and then applies environment variable overrides.
//
// Example:
//
//	config, err := LoadWithEnv("configs/config.json")
//	if err != nil {
//	    log.Fatal(err)
//	}
func LoadWithEnv(configPath string) (*Config, error) {
	config, err := Load(configPath)
	if err != nil {
		return nil, err
	}

	if err := config.LoadEnv(); err != nil {
		return nil, err
	}

	return config, nil
}

// SetDefaults sets default values for optional configuration fields
func (c *Config) SetDefaults() {
	if c.PageTimeout == 0 {
		c.PageTimeout = 30
	}
	if c.SearchTimeout == 0 {
		c.SearchTimeout = 15
	}
	if c.MaxRetries == 0 {
		c.MaxRetries = 3
	}
	if c.RetryDelay == 0 {
		c.RetryDelay = 5
	}
	if c.Workers == 0 {
		c.Workers = 5
	}
	if c.Interval == 0 {
		c.Interval = 300
	}
	if c.LogLevel == "" {
		c.LogLevel = "info"
	}
}
