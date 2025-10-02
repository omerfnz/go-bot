package browser

import (
	"context"
	"testing"
	"time"

	"github.com/omer/go-bot/internal/proxy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ===== browser.go tests =====

func TestNewBrowser_DefaultOptions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping browser test in short mode")
	}

	browser, err := NewBrowser(BrowserOptions{
		Headless: true,
	})
	require.NoError(t, err)
	require.NotNil(t, browser)
	defer browser.Close()

	assert.True(t, browser.IsHeadless())
	assert.NotEmpty(t, browser.GetUserAgent())
	assert.Nil(t, browser.GetProxy())
}

func TestNewBrowser_WithCustomUserAgent(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping browser test in short mode")
	}

	customUA := "Custom User Agent"
	browser, err := NewBrowser(BrowserOptions{
		Headless:  true,
		UserAgent: customUA,
	})
	require.NoError(t, err)
	defer browser.Close()

	assert.Equal(t, customUA, browser.GetUserAgent())
}

func TestNewBrowser_WithProxy(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping browser test in short mode")
	}

	testProxy, err := proxy.ParseProxy("http://proxy.example.com:8080")
	require.NoError(t, err)

	browser, err := NewBrowser(BrowserOptions{
		Headless: true,
		Proxy:    testProxy,
	})
	require.NoError(t, err)
	defer browser.Close()

	assert.NotNil(t, browser.GetProxy())
	assert.Equal(t, testProxy, browser.GetProxy())
}

func TestNewBrowser_WithTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping browser test in short mode")
	}

	timeout := 10 * time.Second
	browser, err := NewBrowser(BrowserOptions{
		Headless: true,
		Timeout:  timeout,
	})
	require.NoError(t, err)
	defer browser.Close()

	assert.NotNil(t, browser.GetContext())
}

func TestBrowser_Navigate_EmptyURL(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping browser test in short mode")
	}

	browser, err := NewBrowser(BrowserOptions{Headless: true})
	require.NoError(t, err)
	defer browser.Close()

	err = browser.Navigate("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "URL cannot be empty")
}

func TestBrowser_Close_MultipleCalls(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping browser test in short mode")
	}

	browser, err := NewBrowser(BrowserOptions{Headless: true})
	require.NoError(t, err)

	// First close should work
	err = browser.Close()
	assert.NoError(t, err)

	// Second close should also work (idempotent)
	err = browser.Close()
	assert.NoError(t, err)
}

func TestBrowser_GetContext(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping browser test in short mode")
	}

	browser, err := NewBrowser(BrowserOptions{Headless: true})
	require.NoError(t, err)
	defer browser.Close()

	ctx := browser.GetContext()
	assert.NotNil(t, ctx)

	// Context should be valid
	select {
	case <-ctx.Done():
		t.Fatal("Context should not be done immediately")
	default:
		// OK
	}
}

func TestBrowser_IsHeadless(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping browser test in short mode")
	}

	tests := []struct {
		name     string
		headless bool
	}{
		{"headless_true", true},
		{"headless_false", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			browser, err := NewBrowser(BrowserOptions{
				Headless: tt.headless,
			})
			require.NoError(t, err)
			defer browser.Close()

			assert.Equal(t, tt.headless, browser.IsHeadless())
		})
	}
}

// ===== actions.go tests =====

func TestBrowser_Type_EmptySelector(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping browser test in short mode")
	}

	browser, err := NewBrowser(BrowserOptions{Headless: true})
	require.NoError(t, err)
	defer browser.Close()

	err = browser.Type("", "test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "selector cannot be empty")
}

func TestBrowser_Click_EmptySelector(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping browser test in short mode")
	}

	browser, err := NewBrowser(BrowserOptions{Headless: true})
	require.NoError(t, err)
	defer browser.Close()

	err = browser.Click("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "selector cannot be empty")
}

func TestBrowser_WaitVisible_EmptySelector(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping browser test in short mode")
	}

	browser, err := NewBrowser(BrowserOptions{Headless: true})
	require.NoError(t, err)
	defer browser.Close()

	err = browser.WaitVisible("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "selector cannot be empty")
}

func TestBrowser_WaitNotVisible_EmptySelector(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping browser test in short mode")
	}

	browser, err := NewBrowser(BrowserOptions{Headless: true})
	require.NoError(t, err)
	defer browser.Close()

	err = browser.WaitNotVisible("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "selector cannot be empty")
}

func TestBrowser_GetText_EmptySelector(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping browser test in short mode")
	}

	browser, err := NewBrowser(BrowserOptions{Headless: true})
	require.NoError(t, err)
	defer browser.Close()

	text, err := browser.GetText("")
	assert.Error(t, err)
	assert.Empty(t, text)
	assert.Contains(t, err.Error(), "selector cannot be empty")
}

func TestBrowser_GetAttribute_EmptySelector(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping browser test in short mode")
	}

	browser, err := NewBrowser(BrowserOptions{Headless: true})
	require.NoError(t, err)
	defer browser.Close()

	value, err := browser.GetAttribute("", "href")
	assert.Error(t, err)
	assert.Empty(t, value)
	assert.Contains(t, err.Error(), "selector cannot be empty")
}

func TestBrowser_GetAttribute_EmptyAttribute(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping browser test in short mode")
	}

	browser, err := NewBrowser(BrowserOptions{Headless: true})
	require.NoError(t, err)
	defer browser.Close()

	value, err := browser.GetAttribute("a", "")
	assert.Error(t, err)
	assert.Empty(t, value)
	assert.Contains(t, err.Error(), "attribute cannot be empty")
}

func TestBrowser_ElementExists_EmptySelector(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping browser test in short mode")
	}

	browser, err := NewBrowser(BrowserOptions{Headless: true})
	require.NoError(t, err)
	defer browser.Close()

	exists := browser.ElementExists("")
	assert.False(t, exists)
}

func TestBrowser_ScrollToElement_EmptySelector(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping browser test in short mode")
	}

	browser, err := NewBrowser(BrowserOptions{Headless: true})
	require.NoError(t, err)
	defer browser.Close()

	err = browser.ScrollToElement("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "selector cannot be empty")
}

func TestBrowser_Sleep(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping browser test in short mode")
	}

	browser, err := NewBrowser(BrowserOptions{Headless: true})
	require.NoError(t, err)
	defer browser.Close()

	start := time.Now()
	err = browser.Sleep(100 * time.Millisecond)
	elapsed := time.Since(start)

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, elapsed, 100*time.Millisecond)
}

func TestBrowser_Scroll(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping browser test in short mode")
	}

	browser, err := NewBrowser(BrowserOptions{Headless: true})
	require.NoError(t, err)
	defer browser.Close()

	// Scroll should not return error even without a page
	// (it just executes JavaScript)
	err = browser.Scroll(0, 100)
	assert.NoError(t, err)
}

// ===== Integration tests (require actual browser) =====

// These tests require a real browser and network access
// Run with: go test -v ./internal/browser/

func TestBrowser_Navigate_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	browser, err := NewBrowser(BrowserOptions{
		Headless: true,
		Timeout:  10 * time.Second,
	})
	require.NoError(t, err)
	defer browser.Close()

	// Navigate to example.com
	err = browser.Navigate("https://example.com")
	if err != nil {
		t.Logf("Navigation failed (expected if no internet): %v", err)
		t.Skip("Skipping navigation test - network required")
	}
}

func TestBrowser_GetTitle_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	browser, err := NewBrowser(BrowserOptions{
		Headless: true,
		Timeout:  10 * time.Second,
	})
	require.NoError(t, err)
	defer browser.Close()

	err = browser.Navigate("https://example.com")
	if err != nil {
		t.Skip("Skipping test - network required")
	}

	title, err := browser.GetTitle()
	assert.NoError(t, err)
	assert.Contains(t, title, "Example")
}

func TestBrowser_GetCurrentURL_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	browser, err := NewBrowser(BrowserOptions{
		Headless: true,
		Timeout:  10 * time.Second,
	})
	require.NoError(t, err)
	defer browser.Close()

	testURL := "https://example.com"
	err = browser.Navigate(testURL)
	if err != nil {
		t.Skip("Skipping test - network required")
	}

	url, err := browser.GetCurrentURL()
	assert.NoError(t, err)
	assert.Contains(t, url, "example.com")
}

// Test context timeout
func TestBrowser_ContextTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping browser test in short mode")
	}

	browser, err := NewBrowser(BrowserOptions{
		Headless: true,
		Timeout:  1 * time.Nanosecond, // Very short timeout
	})
	require.NoError(t, err)
	defer browser.Close()

	// Context should be expired
	select {
	case <-browser.GetContext().Done():
		assert.Error(t, browser.GetContext().Err())
		assert.Equal(t, context.DeadlineExceeded, browser.GetContext().Err())
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Context should have timed out")
	}
}

// ===== Additional integration tests for full coverage =====

func TestBrowser_TypeClick_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	browser, err := NewBrowser(BrowserOptions{
		Headless: true,
		Timeout:  10 * time.Second,
	})
	require.NoError(t, err)
	defer browser.Close()

	// Create a simple HTML with input and button
	htmlContent := `data:text/html,
	<!DOCTYPE html>
	<html>
	<head><title>Test Page</title></head>
	<body>
		<h1 class="title">Test Title</h1>
		<input type="text" id="testInput" />
		<button id="testButton">Click Me</button>
		<a href="#" id="testLink">Test Link</a>
	</body>
	</html>`

	err = browser.Navigate(htmlContent)
	assert.NoError(t, err)

	// Test Type - success case
	err = browser.Type("#testInput", "Hello World")
	assert.NoError(t, err)

	// Test Click - success case
	err = browser.Click("#testButton")
	assert.NoError(t, err)

	// Test GetText - already tested in earlier test but ensure it works
	text, err := browser.GetText("h1")
	assert.NoError(t, err)
	assert.Equal(t, "Test Title", text)

	// Test GetAttribute - success case
	attr, err := browser.GetAttribute("h1", "class")
	assert.NoError(t, err)
	assert.Equal(t, "title", attr)

	// Test WaitVisible - success case
	err = browser.WaitVisible("body")
	assert.NoError(t, err)

	// Test ElementExists - success cases
	exists := browser.ElementExists("h1")
	assert.True(t, exists)

	exists = browser.ElementExists("div#nonexistent123")
	assert.False(t, exists)

	// Test ScrollToElement - success case
	err = browser.ScrollToElement("h1")
	assert.NoError(t, err)
}

func TestBrowser_Screenshot_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	browser, err := NewBrowser(BrowserOptions{
		Headless: true,
		Timeout:  10 * time.Second,
	})
	require.NoError(t, err)
	defer browser.Close()

	err = browser.Navigate("https://example.com")
	if err != nil {
		t.Skip("Skipping test - network required")
	}

	buf, err := browser.Screenshot()
	assert.NoError(t, err)
	assert.NotEmpty(t, buf)
	assert.Greater(t, len(buf), 1000) // Should be a reasonable size
}

func TestBrowser_Reload_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	browser, err := NewBrowser(BrowserOptions{
		Headless: true,
		Timeout:  10 * time.Second,
	})
	require.NoError(t, err)
	defer browser.Close()

	err = browser.Navigate("https://example.com")
	if err != nil {
		t.Skip("Skipping test - network required")
	}

	// Test reload
	err = browser.Reload()
	assert.NoError(t, err)

	// Verify we're still on the same page
	url, err := browser.GetCurrentURL()
	assert.NoError(t, err)
	assert.Contains(t, url, "example.com")
}

func TestBrowser_NavigationHistory_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	browser, err := NewBrowser(BrowserOptions{
		Headless: true,
		Timeout:  10 * time.Second,
	})
	require.NoError(t, err)
	defer browser.Close()

	// Navigate to first page
	err = browser.Navigate("https://example.com")
	if err != nil {
		t.Skip("Skipping test - network required")
	}

	// Get initial URL
	_, err = browser.GetCurrentURL()
	assert.NoError(t, err)

	// Try to go back (should not error even though there's no history)
	err = browser.GoBack()
	assert.NoError(t, err)

	// Try to go forward
	err = browser.GoForward()
	assert.NoError(t, err)
}

func TestBrowser_WaitNotVisible_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	browser, err := NewBrowser(BrowserOptions{
		Headless: true,
		Timeout:  10 * time.Second,
	})
	require.NoError(t, err)
	defer browser.Close()

	// Create a simple HTML with a disappearing element
	htmlContent := `data:text/html,
	<!DOCTYPE html>
	<html>
	<head><title>Test</title></head>
	<body>
		<div id="visible">Visible</div>
		<div id="hidden" style="display:none;">Hidden</div>
	</body>
	</html>`

	err = browser.Navigate(htmlContent)
	assert.NoError(t, err)

	// Test WaitNotVisible on hidden element
	err = browser.WaitNotVisible("#hidden")
	assert.NoError(t, err)
}

// Human-like behavior tests

func TestBrowser_TypeHumanLike_EmptySelector(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping browser test in short mode")
	}

	browser, err := NewBrowser(BrowserOptions{
		Headless: true,
		Timeout:  5 * time.Second,
	})
	assert.NoError(t, err)
	defer browser.Close()

	err = browser.TypeHumanLike("", "test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "selector cannot be empty")
}

func TestBrowser_ClickWithDelay_EmptySelector(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping browser test in short mode")
	}

	browser, err := NewBrowser(BrowserOptions{
		Headless: true,
		Timeout:  5 * time.Second,
	})
	assert.NoError(t, err)
	defer browser.Close()

	err = browser.ClickWithDelay("", 1*time.Second, 2*time.Second)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "selector cannot be empty")
}

func TestBrowser_ScrollRandom(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping browser test in short mode")
	}

	browser, err := NewBrowser(BrowserOptions{
		Headless: true,
		Timeout:  5 * time.Second,
	})
	assert.NoError(t, err)
	defer browser.Close()

	// Navigate to a page first
	err = browser.Navigate("https://example.com")
	assert.NoError(t, err)

	// ScrollRandom should work
	err = browser.ScrollRandom(2, 100, 200)
	assert.NoError(t, err)
}

func TestBrowser_WaitRandom(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping browser test in short mode")
	}

	browser, err := NewBrowser(BrowserOptions{
		Headless: true,
		Timeout:  5 * time.Second,
	})
	assert.NoError(t, err)
	defer browser.Close()

	start := time.Now()
	err = browser.WaitRandom(100*time.Millisecond, 200*time.Millisecond)
	elapsed := time.Since(start)

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, elapsed, 100*time.Millisecond)
	assert.LessOrEqual(t, elapsed, 300*time.Millisecond) // Allow some overhead
}

func TestBrowser_MouseMoveToElement_EmptySelector(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping browser test in short mode")
	}

	browser, err := NewBrowser(BrowserOptions{
		Headless: true,
		Timeout:  5 * time.Second,
	})
	assert.NoError(t, err)
	defer browser.Close()

	err = browser.MouseMoveToElement("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "selector cannot be empty")
}

func TestBrowser_ScrollToElementSmoothly_EmptySelector(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping browser test in short mode")
	}

	browser, err := NewBrowser(BrowserOptions{
		Headless: true,
		Timeout:  5 * time.Second,
	})
	assert.NoError(t, err)
	defer browser.Close()

	err = browser.ScrollToElementSmoothly("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "selector cannot be empty")
}

func TestBrowser_HoverElement_EmptySelector(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping browser test in short mode")
	}

	browser, err := NewBrowser(BrowserOptions{
		Headless: true,
		Timeout:  5 * time.Second,
	})
	assert.NoError(t, err)
	defer browser.Close()

	err = browser.HoverElement("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "selector cannot be empty")
}
