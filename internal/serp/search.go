// Package serp provides search engine result page (SERP) automation functionality.
// It handles Google search operations, result parsing, and target URL finding.
package serp

import (
	"fmt"
	"strings"
	"time"

	"github.com/omer/go-bot/internal/browser"
	"github.com/omer/go-bot/internal/logger"
)

// SearchResult represents a single search result from Google
type SearchResult struct {
	Title       string // Result title (h3 text)
	URL         string // Result URL (href attribute)
	Description string // Result description/snippet
	Position    int    // Position in search results (1-based)
}

// Selectors holds CSS selectors for Google search elements
type Selectors struct {
	SearchBox    string // Search input box selector
	SearchButton string // Search button selector
	ResultItem   string // Individual result container selector
	ResultLink   string // Result link selector (relative to result item)
	ResultTitle  string // Result title selector (relative to result item)
	NextButton   string // Next page button selector
	CaptchaFrame string // CAPTCHA iframe selector
}

// DefaultSelectors returns the default Google search selectors
func DefaultSelectors() Selectors {
	return Selectors{
		SearchBox:    "textarea[name='q']",
		SearchButton: "input[name='btnK']",
		ResultItem:   "div.g",
		ResultLink:   "a[href]",
		ResultTitle:  "h3",
		NextButton:   "a#pnnext",
		CaptchaFrame: "iframe[src*='recaptcha']",
	}
}

// Searcher handles Google search operations
type Searcher struct {
	browser   *browser.Browser
	selectors Selectors
	logger    *logger.Logger
}

// SearchOptions holds configuration for search operations
type SearchOptions struct {
	Keyword   string        // Search keyword
	TargetURL string        // Target URL to find
	MaxPages  int           // Maximum pages to search (default: 5)
	Timeout   time.Duration // Timeout for search operation (default: 30s)
}

// NewSearcher creates a new Searcher instance
//
// Example:
//
//	searcher := serp.NewSearcher(browser, logger)
func NewSearcher(b *browser.Browser, log *logger.Logger) *Searcher {
	return &Searcher{
		browser:   b,
		selectors: DefaultSelectors(),
		logger:    log,
	}
}

// NewSearcherWithSelectors creates a new Searcher with custom selectors
func NewSearcherWithSelectors(b *browser.Browser, log *logger.Logger, selectors Selectors) *Searcher {
	return &Searcher{
		browser:   b,
		selectors: selectors,
		logger:    log,
	}
}

// Search performs a Google search with the given keyword
//
// Example:
//
//	err := searcher.Search("golang tutorial")
func (s *Searcher) Search(keyword string) error {
	if keyword == "" {
		return fmt.Errorf("keyword cannot be empty")
	}

	s.logger.Info("Starting search", map[string]interface{}{
		"keyword": keyword,
	})

	// Navigate to Google
	err := s.browser.Navigate("https://www.google.com")
	if err != nil {
		return fmt.Errorf("failed to navigate to Google: %w", err)
	}

	// Wait for search box to be visible
	err = s.browser.WaitVisible(s.selectors.SearchBox)
	if err != nil {
		return fmt.Errorf("search box not found: %w", err)
	}

	// Type the keyword
	err = s.browser.Type(s.selectors.SearchBox, keyword)
	if err != nil {
		return fmt.Errorf("failed to type keyword: %w", err)
	}

	// Small delay before submitting
	time.Sleep(500 * time.Millisecond)

	// Submit the search (press Enter)
	err = s.browser.Type(s.selectors.SearchBox, "\n")
	if err != nil {
		return fmt.Errorf("failed to submit search: %w", err)
	}

	// Wait for results to load
	time.Sleep(2 * time.Second)

	// Check for CAPTCHA
	if s.browser.ElementExists(s.selectors.CaptchaFrame) {
		s.logger.Warn("CAPTCHA detected", nil)
		return fmt.Errorf("CAPTCHA detected - please solve manually or use a different proxy")
	}

	s.logger.Info("Search completed successfully", nil)
	return nil
}

// GetResults parses and returns all search results from the current page
//
// Example:
//
//	results, err := searcher.GetResults()
func (s *Searcher) GetResults() ([]SearchResult, error) {
	s.logger.Debug("Parsing search results", nil)

	// Wait for results to be visible
	err := s.browser.WaitVisible(s.selectors.ResultItem)
	if err != nil {
		return nil, fmt.Errorf("no results found: %w", err)
	}

	// In a real implementation, we would use chromedp to extract multiple elements
	// For now, we'll return a placeholder implementation
	// This will be properly implemented with chromedp.Nodes and iteration

	results := []SearchResult{}

	// Note: This is a simplified version. Full implementation would:
	// 1. Use chromedp.Nodes to get all result items
	// 2. Iterate through each item
	// 3. Extract title, URL, description for each
	// 4. Return the complete list

	s.logger.Info("Found search results", map[string]interface{}{
		"count": len(results),
	})

	return results, nil
}

// FindTarget searches for a specific target URL in the search results
// Returns the SearchResult and its position if found, error otherwise
//
// Example:
//
//	result, err := searcher.FindTarget("example.com")
func (s *Searcher) FindTarget(targetURL string) (*SearchResult, error) {
	if targetURL == "" {
		return nil, fmt.Errorf("target URL cannot be empty")
	}

	s.logger.Info("Searching for target URL", map[string]interface{}{
		"target": targetURL,
	})

	// Normalize target URL (remove protocol, www, trailing slash)
	normalizedTarget := normalizeURL(targetURL)

	results, err := s.GetResults()
	if err != nil {
		return nil, err
	}

	// Search through results
	for _, result := range results {
		normalizedResultURL := normalizeURL(result.URL)

		// Check if result URL contains target
		if strings.Contains(normalizedResultURL, normalizedTarget) {
			s.logger.Info("Target found", map[string]interface{}{
				"position": result.Position,
				"url":      result.URL,
			})
			return &result, nil
		}
	}

	s.logger.Warn("Target not found in current page", map[string]interface{}{
		"target": targetURL,
	})
	return nil, fmt.Errorf("target URL not found: %s", targetURL)
}

// HasCaptcha checks if a CAPTCHA is present on the page
func (s *Searcher) HasCaptcha() bool {
	return s.browser.ElementExists(s.selectors.CaptchaFrame)
}

// normalizeURL removes protocol, www prefix, and trailing slashes
func normalizeURL(url string) string {
	// Convert to lowercase first
	url = strings.ToLower(url)

	// Remove protocol
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")

	// Remove www
	url = strings.TrimPrefix(url, "www.")

	// Remove trailing slash
	url = strings.TrimSuffix(url, "/")

	return url
}
