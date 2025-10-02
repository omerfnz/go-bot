package serp

import (
	"testing"
	"time"

	"github.com/omer/go-bot/internal/browser"
	"github.com/omer/go-bot/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ===== Helper functions =====

func createTestSearcher(t *testing.T) (*Searcher, *browser.Browser) {
	if testing.Short() {
		t.Skip("Skipping searcher test in short mode")
	}

	log, err := logger.New(logger.Config{
		Level:      logger.DebugLevel,
		EnableFile: false,
	})
	require.NoError(t, err)

	b, err := browser.NewBrowser(browser.BrowserOptions{
		Headless: true,
		Timeout:  10 * time.Second,
	})
	require.NoError(t, err)

	searcher := NewSearcher(b, log)
	return searcher, b
}

// ===== Searcher creation tests =====

func TestNewSearcher(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping searcher test in short mode")
	}

	log, err := logger.New(logger.Config{
		Level:      logger.InfoLevel,
		EnableFile: false,
	})
	require.NoError(t, err)

	b, err := browser.NewBrowser(browser.BrowserOptions{Headless: true})
	require.NoError(t, err)
	defer b.Close()

	searcher := NewSearcher(b, log)
	assert.NotNil(t, searcher)
	assert.NotNil(t, searcher.browser)
	assert.NotNil(t, searcher.logger)
	assert.NotNil(t, searcher.selectors)
}

func TestNewSearcherWithSelectors(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping searcher test in short mode")
	}

	log, err := logger.New(logger.Config{
		Level:      logger.InfoLevel,
		EnableFile: false,
	})
	require.NoError(t, err)

	b, err := browser.NewBrowser(browser.BrowserOptions{Headless: true})
	require.NoError(t, err)
	defer b.Close()

	customSelectors := Selectors{
		SearchBox:  "input.custom",
		ResultItem: "div.custom-result",
	}

	searcher := NewSearcherWithSelectors(b, log, customSelectors)
	assert.NotNil(t, searcher)
	assert.Equal(t, "input.custom", searcher.selectors.SearchBox)
	assert.Equal(t, "div.custom-result", searcher.selectors.ResultItem)
}

func TestDefaultSelectors(t *testing.T) {
	selectors := DefaultSelectors()
	assert.NotEmpty(t, selectors.SearchBox)
	assert.NotEmpty(t, selectors.SearchButton)
	assert.NotEmpty(t, selectors.ResultItem)
	assert.NotEmpty(t, selectors.ResultLink)
	assert.NotEmpty(t, selectors.ResultTitle)
	assert.NotEmpty(t, selectors.NextButton)
	assert.NotEmpty(t, selectors.CaptchaFrame)
}

// ===== Search tests =====

func TestSearch_EmptyKeyword(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping searcher test in short mode")
	}

	searcher, b := createTestSearcher(t)
	defer b.Close()

	err := searcher.Search("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "keyword cannot be empty")
}

func TestSearch_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	searcher, b := createTestSearcher(t)
	defer b.Close()

	// Perform a search
	err := searcher.Search("golang")
	if err != nil {
		// Network or Google-related issues are acceptable in tests
		t.Logf("Search failed (network issue expected): %v", err)
		t.Skip("Skipping - network required")
	}

	// Verify we're on a Google results page
	url, err := b.GetCurrentURL()
	if err == nil {
		assert.Contains(t, url, "google.com")
	}
}

// ===== GetResults tests =====

func TestGetResults_NoNavigation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping searcher test in short mode")
	}

	searcher, b := createTestSearcher(t)
	defer b.Close()

	// Try to get results without navigating
	_, err := searcher.GetResults()
	assert.Error(t, err)
}

// ===== FindTarget tests =====

func TestFindTarget_EmptyURL(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping searcher test in short mode")
	}

	searcher, b := createTestSearcher(t)
	defer b.Close()

	_, err := searcher.FindTarget("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "target URL cannot be empty")
}

func TestFindTarget_NoResults(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping searcher test in short mode")
	}

	searcher, b := createTestSearcher(t)
	defer b.Close()

	// Try to find target without any search
	_, err := searcher.FindTarget("example.com")
	assert.Error(t, err)
}

// ===== HasCaptcha tests =====

func TestHasCaptcha(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping searcher test in short mode")
	}

	log, err := logger.New(logger.Config{
		Level:      logger.DebugLevel,
		EnableFile: false,
	})
	require.NoError(t, err)

	b, err := browser.NewBrowser(browser.BrowserOptions{
		Headless: true,
		Timeout:  10 * time.Second,
	})
	require.NoError(t, err)
	defer b.Close()

	// Use a simpler selector for testing
	customSelectors := DefaultSelectors()
	customSelectors.CaptchaFrame = "iframe" // Simple selector

	searcher := NewSearcherWithSelectors(b, log, customSelectors)

	// Create a mock page with iframe
	mockHTML := `data:text/html;charset=utf-8,<!DOCTYPE html><html><body><iframe src="test"></iframe></body></html>`

	err = b.Navigate(mockHTML)
	require.NoError(t, err)

	// Small delay to ensure page is loaded
	time.Sleep(500 * time.Millisecond)

	// Check for CAPTCHA (which is just checking for iframe)
	hasCaptcha := searcher.HasCaptcha()
	assert.True(t, hasCaptcha)
}

func TestHasNoCaptcha(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping searcher test in short mode")
	}

	searcher, b := createTestSearcher(t)
	defer b.Close()

	// Create a simple page without CAPTCHA
	mockHTML := `data:text/html,
	<!DOCTYPE html>
	<html>
	<body>
		<h1>Test Page</h1>
	</body>
	</html>`

	err := b.Navigate(mockHTML)
	require.NoError(t, err)

	// Check for CAPTCHA
	hasCaptcha := searcher.HasCaptcha()
	assert.False(t, hasCaptcha)
}

// ===== URL normalization tests =====

func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "http_protocol",
			input:    "http://example.com",
			expected: "example.com",
		},
		{
			name:     "https_protocol",
			input:    "https://example.com",
			expected: "example.com",
		},
		{
			name:     "with_www",
			input:    "https://www.example.com",
			expected: "example.com",
		},
		{
			name:     "trailing_slash",
			input:    "https://example.com/",
			expected: "example.com",
		},
		{
			name:     "with_path",
			input:    "https://example.com/page",
			expected: "example.com/page",
		},
		{
			name:     "uppercase",
			input:    "https://Example.COM",
			expected: "example.com",
		},
		{
			name:     "complex",
			input:    "HTTPS://WWW.Example.COM/",
			expected: "example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeURL(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// ===== Navigation tests =====

func TestNextPage_NoButton(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping searcher test in short mode")
	}

	searcher, b := createTestSearcher(t)
	defer b.Close()

	// Navigate to a simple page without next button
	mockHTML := `data:text/html,
	<!DOCTYPE html>
	<html>
	<body>
		<h1>No Next Button</h1>
	</body>
	</html>`

	err := b.Navigate(mockHTML)
	require.NoError(t, err)

	// Try to go to next page
	hasNext, err := searcher.NextPage()
	assert.NoError(t, err)
	assert.False(t, hasNext)
}

func TestClickResult_InvalidPosition(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping searcher test in short mode")
	}

	searcher, b := createTestSearcher(t)
	defer b.Close()

	// Test with invalid position (0)
	err := searcher.ClickResult(0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "position must be >= 1")

	// Test with negative position
	err = searcher.ClickResult(-5)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "position must be >= 1")
}

func TestClickTargetResult_EmptyTarget(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping searcher test in short mode")
	}

	searcher, b := createTestSearcher(t)
	defer b.Close()

	err := searcher.ClickTargetResult("")
	assert.Error(t, err)
}

func TestGetCurrentPage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping searcher test in short mode")
	}

	searcher, b := createTestSearcher(t)
	defer b.Close()

	page, err := searcher.GetCurrentPage()
	assert.NoError(t, err)
	assert.Equal(t, 1, page)
}

func TestScrollToResult_InvalidPosition(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping searcher test in short mode")
	}

	searcher, b := createTestSearcher(t)
	defer b.Close()

	// Test with invalid position
	err := searcher.ScrollToResult(0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "position must be >= 1")
}

func TestScrollToResult_ValidPosition(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping searcher test in short mode")
	}

	searcher, b := createTestSearcher(t)
	defer b.Close()

	// Navigate to a simple page
	mockHTML := `data:text/html,
	<!DOCTYPE html>
	<html>
	<body style="height: 3000px;">
		<h1>Tall Page</h1>
	</body>
	</html>`

	err := b.Navigate(mockHTML)
	require.NoError(t, err)

	// Test scrolling to a result
	err = searcher.ScrollToResult(3)
	assert.NoError(t, err)
}

// ===== SearchResult struct test =====

func TestSearchResult_Struct(t *testing.T) {
	result := SearchResult{
		Title:       "Test Title",
		URL:         "https://example.com",
		Description: "Test description",
		Position:    1,
	}

	assert.Equal(t, "Test Title", result.Title)
	assert.Equal(t, "https://example.com", result.URL)
	assert.Equal(t, "Test description", result.Description)
	assert.Equal(t, 1, result.Position)
}

// ===== Selectors struct test =====

func TestSelectors_Struct(t *testing.T) {
	selectors := Selectors{
		SearchBox:    "input#search",
		SearchButton: "button#submit",
		ResultItem:   "div.result",
		ResultLink:   "a.link",
		ResultTitle:  "h3.title",
		NextButton:   "a.next",
		CaptchaFrame: "iframe.captcha",
	}

	assert.Equal(t, "input#search", selectors.SearchBox)
	assert.Equal(t, "button#submit", selectors.SearchButton)
	assert.Equal(t, "div.result", selectors.ResultItem)
	assert.Equal(t, "a.link", selectors.ResultLink)
	assert.Equal(t, "h3.title", selectors.ResultTitle)
	assert.Equal(t, "a.next", selectors.NextButton)
	assert.Equal(t, "iframe.captcha", selectors.CaptchaFrame)
}
