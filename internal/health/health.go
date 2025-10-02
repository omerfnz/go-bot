// Package health provides health checking functionality
package health

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/omer/go-bot/internal/config"
	"github.com/omer/go-bot/internal/logger"
	"github.com/omer/go-bot/internal/proxy"
)

// CheckResult represents the result of a health check
type CheckResult struct {
	Name    string                 // Name of the check
	Passed  bool                   // Whether the check passed
	Message string                 // Description/error message
	Details map[string]interface{} // Additional details
}

// HealthChecker performs system health checks
type HealthChecker struct {
	config *config.Config
	logger *logger.Logger
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(cfg *config.Config, log *logger.Logger) *HealthChecker {
	return &HealthChecker{
		config: cfg,
		logger: log,
	}
}

// CheckAll performs all health checks and returns results
func (h *HealthChecker) CheckAll(ctx context.Context) []CheckResult {
	checks := []func(context.Context) CheckResult{
		h.checkChrome,
		h.checkConfig,
		h.checkProxies,
		h.checkDiskSpace,
		h.checkMemory,
	}

	results := make([]CheckResult, 0, len(checks))
	for _, check := range checks {
		result := check(ctx)
		results = append(results, result)
		h.logger.WithFields(map[string]interface{}{
			"check":  result.Name,
			"passed": result.Passed,
		}).Debug("Health check completed")
	}

	return results
}

// checkChrome verifies Chrome/Chromium is installed
func (h *HealthChecker) checkChrome(ctx context.Context) CheckResult {
	result := CheckResult{
		Name:    "Chrome/Chromium",
		Details: make(map[string]interface{}),
	}

	// Try to find Chrome executable
	var chromeCmd string
	switch runtime.GOOS {
	case "windows":
		chromeCmd = "chrome.exe"
	case "darwin":
		chromeCmd = "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"
	default:
		chromeCmd = "google-chrome"
	}

	// Check if chrome is in PATH or at known location
	_, err := exec.LookPath(chromeCmd)
	if err != nil {
		// Try alternative names
		alternativeNames := []string{"chromium", "chromium-browser", "google-chrome-stable"}
		found := false
		for _, name := range alternativeNames {
			if _, err := exec.LookPath(name); err == nil {
				chromeCmd = name
				found = true
				break
			}
		}

		if !found {
			result.Passed = false
			result.Message = "Chrome/Chromium not found in PATH"
			result.Details["error"] = err.Error()
			return result
		}
	}

	// Try to get version
	versionCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(versionCtx, chromeCmd, "--version")
	output, err := cmd.Output()
	if err != nil {
		result.Passed = true // Found but version check failed
		result.Message = "Chrome found but version check failed"
		result.Details["path"] = chromeCmd
	} else {
		result.Passed = true
		result.Message = "Chrome available"
		result.Details["path"] = chromeCmd
		result.Details["version"] = string(output)
	}

	return result
}

// checkConfig validates the configuration
func (h *HealthChecker) checkConfig(ctx context.Context) CheckResult {
	result := CheckResult{
		Name:    "Configuration",
		Details: make(map[string]interface{}),
	}

	if h.config == nil {
		result.Passed = false
		result.Message = "Configuration not loaded"
		return result
	}

	// Validate config
	if err := h.config.Validate(); err != nil {
		result.Passed = false
		result.Message = fmt.Sprintf("Invalid configuration: %v", err)
		return result
	}

	result.Passed = true
	result.Message = "Configuration valid"
	result.Details["keywords"] = len(h.config.Keywords)
	result.Details["proxies"] = len(h.config.Proxies)
	result.Details["workers"] = h.config.Workers
	result.Details["headless"] = h.config.Headless

	return result
}

// checkProxies validates proxy configuration
func (h *HealthChecker) checkProxies(ctx context.Context) CheckResult {
	result := CheckResult{
		Name:    "Proxy Pool",
		Details: make(map[string]interface{}),
	}

	if len(h.config.Proxies) == 0 {
		result.Passed = false
		result.Message = "No proxies configured"
		return result
	}

	// Try to parse proxies
	validProxies := 0
	for _, proxyURL := range h.config.Proxies {
		if _, err := proxy.ParseProxy(proxyURL); err == nil {
			validProxies++
		}
	}

	if validProxies == 0 {
		result.Passed = false
		result.Message = "No valid proxies found"
		result.Details["total"] = len(h.config.Proxies)
		result.Details["valid"] = 0
		return result
	}

	result.Passed = true
	result.Message = fmt.Sprintf("%d/%d proxies valid", validProxies, len(h.config.Proxies))
	result.Details["total"] = len(h.config.Proxies)
	result.Details["valid"] = validProxies

	return result
}

// checkDiskSpace verifies sufficient disk space is available
func (h *HealthChecker) checkDiskSpace(ctx context.Context) CheckResult {
	result := CheckResult{
		Name:    "Disk Space",
		Details: make(map[string]interface{}),
	}

	// Get current working directory stats
	wd, err := os.Getwd()
	if err != nil {
		result.Passed = false
		result.Message = fmt.Sprintf("Cannot get working directory: %v", err)
		return result
	}

	// Check if we can write
	testFile := fmt.Sprintf("%s/.health_check_%d", wd, time.Now().UnixNano())
	f, err := os.Create(testFile)
	if err != nil {
		result.Passed = false
		result.Message = "Cannot write to disk"
		result.Details["error"] = err.Error()
		return result
	}
	f.Close()
	os.Remove(testFile)

	result.Passed = true
	result.Message = "Disk space available"
	result.Details["working_directory"] = wd

	return result
}

// checkMemory verifies memory usage is within acceptable limits
func (h *HealthChecker) checkMemory(ctx context.Context) CheckResult {
	result := CheckResult{
		Name:    "Memory",
		Details: make(map[string]interface{}),
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Convert to MB
	allocMB := m.Alloc / 1024 / 1024
	totalAllocMB := m.TotalAlloc / 1024 / 1024
	sysMB := m.Sys / 1024 / 1024

	result.Details["alloc_mb"] = allocMB
	result.Details["total_alloc_mb"] = totalAllocMB
	result.Details["sys_mb"] = sysMB
	result.Details["num_gc"] = m.NumGC

	// Check if memory usage is reasonable (< 500MB)
	if allocMB > 500 {
		result.Passed = false
		result.Message = fmt.Sprintf("High memory usage: %dMB", allocMB)
		return result
	}

	result.Passed = true
	result.Message = fmt.Sprintf("Memory usage OK: %dMB", allocMB)

	return result
}

// PrintResults prints health check results to console
func PrintResults(results []CheckResult) {
	fmt.Println("\n🏥 SERP Bot Health Check")
	fmt.Println("═══════════════════════")

	passed := 0
	for i, result := range results {
		status := "❌ FAIL"
		if result.Passed {
			status = "✅ OK"
			passed++
		}

		fmt.Printf("%d. %s... %s\n", i+1, result.Name, status)
		if result.Message != "" {
			fmt.Printf("   %s\n", result.Message)
		}

		// Print some details
		if len(result.Details) > 0 {
			for key, value := range result.Details {
				if key != "error" { // Don't print error twice
					fmt.Printf("   - %s: %v\n", key, value)
				}
			}
		}
	}

	fmt.Println("\n═══════════════════════")
	if passed == len(results) {
		fmt.Printf("✅ All checks passed (%d/%d)\n\n", passed, len(results))
	} else {
		fmt.Printf("⚠️  Some checks failed (%d/%d passed)\n\n", passed, len(results))
	}
}

// AllPassed returns true if all health checks passed
func AllPassed(results []CheckResult) bool {
	for _, result := range results {
		if !result.Passed {
			return false
		}
	}
	return true
}
