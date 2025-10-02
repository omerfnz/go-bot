package health

import (
	"context"
	"testing"
	"time"

	"github.com/omer/go-bot/internal/config"
	"github.com/omer/go-bot/internal/logger"
)

func TestNewHealthChecker(t *testing.T) {
	cfg := &config.Config{}
	log := logger.NewDefault()

	checker := NewHealthChecker(cfg, log)
	if checker == nil {
		t.Fatal("Expected health checker to be created")
	}
	if checker.config != cfg {
		t.Error("Expected config to be set")
	}
	if checker.logger != log {
		t.Error("Expected logger to be set")
	}
}

func TestHealthChecker_CheckAll(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping health check test in short mode")
	}

	cfg := &config.Config{
		Keywords: []config.Keyword{
			{Term: "test", TargetURL: "example.com"},
		},
		Proxies: []string{"http://proxy.example.com:8080"},
		Workers: 5,
	}
	log := logger.NewDefault()

	checker := NewHealthChecker(cfg, log)
	ctx := context.Background()

	results := checker.CheckAll(ctx)

	if len(results) == 0 {
		t.Error("Expected health check results")
	}

	// Verify all checks have names
	for _, result := range results {
		if result.Name == "" {
			t.Error("Expected all checks to have names")
		}
	}
}

func TestHealthChecker_CheckChrome(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping chrome check in short mode")
	}

	cfg := &config.Config{}
	log := logger.NewDefault()
	checker := NewHealthChecker(cfg, log)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result := checker.checkChrome(ctx)

	if result.Name != "Chrome/Chromium" {
		t.Errorf("Expected check name 'Chrome/Chromium', got '%s'", result.Name)
	}

	// Chrome might not be installed in CI, so just check structure
	if result.Details == nil {
		t.Error("Expected details to be initialized")
	}

	t.Logf("Chrome check result: %v - %s", result.Passed, result.Message)
}

func TestHealthChecker_CheckConfig(t *testing.T) {
	tests := []struct {
		name       string
		cfg        *config.Config
		shouldPass bool
	}{
		{
			name:       "nil config",
			cfg:        nil,
			shouldPass: false,
		},
		{
			name: "valid config",
			cfg: &config.Config{
				Keywords: []config.Keyword{
					{Term: "test", TargetURL: "example.com"},
				},
				Proxies:       []string{"http://proxy.example.com:8080"},
				Workers:       5,
				PageTimeout:   30,
				SearchTimeout: 20,
				MaxRetries:    3,
				RetryDelay:    5,
				Selectors: config.SelectorConfig{
					SearchBox:    "input",
					SearchButton: "button",
					ResultItem:   "div",
					ResultLink:   "a",
					NextButton:   "next",
				},
			},
			shouldPass: true,
		},
		{
			name: "invalid config - no keywords",
			cfg: &config.Config{
				Keywords: []config.Keyword{},
				Proxies:  []string{"http://proxy.example.com:8080"},
				Workers:  5,
			},
			shouldPass: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := logger.NewDefault()
			checker := NewHealthChecker(tt.cfg, log)
			ctx := context.Background()

			result := checker.checkConfig(ctx)

			if result.Passed != tt.shouldPass {
				t.Errorf("Expected passed=%v, got %v - %s", tt.shouldPass, result.Passed, result.Message)
			}
		})
	}
}

func TestHealthChecker_CheckProxies(t *testing.T) {
	tests := []struct {
		name       string
		proxies    []string
		shouldPass bool
	}{
		{
			name:       "no proxies",
			proxies:    []string{},
			shouldPass: false,
		},
		{
			name:       "valid proxies",
			proxies:    []string{"http://proxy1.com:8080", "http://proxy2.com:8080"},
			shouldPass: true,
		},
		{
			name:       "invalid proxies",
			proxies:    []string{"invalid", "also invalid"},
			shouldPass: false,
		},
		{
			name:       "mixed proxies",
			proxies:    []string{"http://valid.com:8080", "invalid"},
			shouldPass: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				Proxies: tt.proxies,
			}
			log := logger.NewDefault()
			checker := NewHealthChecker(cfg, log)
			ctx := context.Background()

			result := checker.checkProxies(ctx)

			if result.Passed != tt.shouldPass {
				t.Errorf("Expected passed=%v, got %v - %s", tt.shouldPass, result.Passed, result.Message)
			}
		})
	}
}

func TestHealthChecker_CheckDiskSpace(t *testing.T) {
	cfg := &config.Config{}
	log := logger.NewDefault()
	checker := NewHealthChecker(cfg, log)
	ctx := context.Background()

	result := checker.checkDiskSpace(ctx)

	if result.Name != "Disk Space" {
		t.Errorf("Expected check name 'Disk Space', got '%s'", result.Name)
	}

	// Should pass on most systems
	if !result.Passed {
		t.Logf("Disk space check failed: %s", result.Message)
	}
}

func TestHealthChecker_CheckMemory(t *testing.T) {
	cfg := &config.Config{}
	log := logger.NewDefault()
	checker := NewHealthChecker(cfg, log)
	ctx := context.Background()

	result := checker.checkMemory(ctx)

	if result.Name != "Memory" {
		t.Errorf("Expected check name 'Memory', got '%s'", result.Name)
	}

	// Check that memory details are populated
	if result.Details["alloc_mb"] == nil {
		t.Error("Expected alloc_mb to be set")
	}

	t.Logf("Memory check: %s", result.Message)
}

func TestAllPassed(t *testing.T) {
	tests := []struct {
		name     string
		results  []CheckResult
		expected bool
	}{
		{
			name:     "all passed",
			results:  []CheckResult{{Passed: true}, {Passed: true}},
			expected: true,
		},
		{
			name:     "one failed",
			results:  []CheckResult{{Passed: true}, {Passed: false}},
			expected: false,
		},
		{
			name:     "all failed",
			results:  []CheckResult{{Passed: false}, {Passed: false}},
			expected: false,
		},
		{
			name:     "empty results",
			results:  []CheckResult{},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AllPassed(tt.results)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestPrintResults(t *testing.T) {
	// Just verify it doesn't panic
	results := []CheckResult{
		{Name: "Test 1", Passed: true, Message: "OK", Details: map[string]interface{}{"key": "value"}},
		{Name: "Test 2", Passed: false, Message: "Failed", Details: map[string]interface{}{}},
	}

	// Should not panic
	PrintResults(results)
}
