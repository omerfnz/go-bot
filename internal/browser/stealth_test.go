package browser

import (
	"context"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
)

func TestApplyStealthMode(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping browser test in short mode")
	}

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	err := ApplyStealthMode(ctx)
	if err != nil {
		t.Errorf("ApplyStealthMode() error = %v", err)
	}
}

func TestDisableWebDriver(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping browser test in short mode")
	}

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	err := disableWebDriver(ctx)
	if err != nil {
		t.Errorf("disableWebDriver() error = %v", err)
	}

	// Verify webdriver is undefined
	var result bool
	err = chromedp.Run(ctx,
		chromedp.Evaluate(`navigator.webdriver === undefined`, &result),
	)
	if err != nil {
		t.Errorf("Failed to evaluate webdriver: %v", err)
	}
	if !result {
		t.Error("navigator.webdriver should be undefined")
	}
}

func TestEnableChromeRuntime(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping browser test in short mode")
	}

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	err := enableChromeRuntime(ctx)
	if err != nil {
		t.Errorf("enableChromeRuntime() error = %v", err)
	}

	// Verify chrome object exists
	var result bool
	err = chromedp.Run(ctx,
		chromedp.Evaluate(`typeof window.chrome !== 'undefined'`, &result),
	)
	if err != nil {
		t.Errorf("Failed to evaluate chrome: %v", err)
	}
	if !result {
		t.Error("window.chrome should be defined")
	}
}

func TestFixPermissions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping browser test in short mode")
	}

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	err := fixPermissions(ctx)
	if err != nil {
		t.Errorf("fixPermissions() error = %v", err)
	}
}

func TestFixPlugins(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping browser test in short mode")
	}

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	err := fixPlugins(ctx)
	if err != nil {
		t.Errorf("fixPlugins() error = %v", err)
	}

	// Verify plugins exist
	var length int64
	err = chromedp.Run(ctx,
		chromedp.Evaluate(`navigator.plugins.length`, &length),
	)
	if err != nil {
		t.Errorf("Failed to evaluate plugins: %v", err)
	}
	if length == 0 {
		t.Error("navigator.plugins should not be empty")
	}
}

func TestFixLanguages(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping browser test in short mode")
	}

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	err := fixLanguages(ctx)
	if err != nil {
		t.Errorf("fixLanguages() error = %v", err)
	}

	// Verify languages are set
	var length int64
	err = chromedp.Run(ctx,
		chromedp.Evaluate(`navigator.languages.length`, &length),
	)
	if err != nil {
		t.Errorf("Failed to evaluate languages: %v", err)
	}
	if length == 0 {
		t.Error("navigator.languages should not be empty")
	}
}

func TestRandomizeFingerprint(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping browser test in short mode")
	}

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	err := RandomizeFingerprint(ctx)
	if err != nil {
		t.Errorf("RandomizeFingerprint() error = %v", err)
	}
}

func TestGenerateRandomFingerprint(t *testing.T) {
	// This test doesn't need a browser
	fingerprint1 := generateRandomFingerprint()
	fingerprint2 := generateRandomFingerprint()

	if fingerprint1.UserAgent == "" {
		t.Error("UserAgent should not be empty")
	}
	if fingerprint1.Platform == "" {
		t.Error("Platform should not be empty")
	}
	if fingerprint1.Vendor == "" {
		t.Error("Vendor should not be empty")
	}
	if fingerprint1.WebGL == "" {
		t.Error("WebGL should not be empty")
	}
	if fingerprint1.Resolution[0] == 0 || fingerprint1.Resolution[1] == 0 {
		t.Error("Resolution should not be zero")
	}

	// Verify some randomness (might occasionally fail due to randomness)
	allSame := fingerprint1.UserAgent == fingerprint2.UserAgent &&
		fingerprint1.Platform == fingerprint2.Platform &&
		fingerprint1.Vendor == fingerprint2.Vendor &&
		fingerprint1.WebGL == fingerprint2.WebGL

	if allSame {
		t.Log("Warning: Two consecutive fingerprints were identical (might be random)")
	}
}

func TestSetUserAgent(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping browser test in short mode")
	}

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	testUA := "TestUserAgent/1.0"
	err := setUserAgent(ctx, testUA)
	if err != nil {
		t.Errorf("setUserAgent() error = %v", err)
	}

	// Note: In real chromedp, navigator.userAgent might not be overridable
	// This test just verifies the script runs without error
}

func TestSetPlatform(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping browser test in short mode")
	}

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	err := setPlatform(ctx, "TestPlatform")
	if err != nil {
		t.Errorf("setPlatform() error = %v", err)
	}
}

func TestSetVendor(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping browser test in short mode")
	}

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	err := setVendor(ctx, "Test Vendor")
	if err != nil {
		t.Errorf("setVendor() error = %v", err)
	}
}

func TestSetWebGLVendor(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping browser test in short mode")
	}

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	err := setWebGLVendor(ctx, "Test WebGL Vendor")
	if err != nil {
		t.Errorf("setWebGLVendor() error = %v", err)
	}
}

func TestDisableAutomationFlags(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping browser test in short mode")
	}

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	err := DisableAutomationFlags(ctx)
	if err != nil {
		t.Errorf("DisableAutomationFlags() error = %v", err)
	}
}

// Benchmark tests
func BenchmarkGenerateRandomFingerprint(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = generateRandomFingerprint()
	}
}
