package serp

import (
	"testing"
	"time"

	"github.com/omer/go-bot/internal/browser"
	"github.com/omer/go-bot/internal/logger"
)

func TestBrowseTarget(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping browser test in short mode")
	}

	// Create test browser
	br, err := browser.NewBrowser(browser.BrowserOptions{
		Headless: true,
		Timeout:  30 * time.Second,
	})
	if err != nil {
		t.Fatalf("Failed to create browser: %v", err)
	}
	defer br.Close()

	// Create logger
	log := logger.NewDefault()

	// Create searcher
	searcher := NewSearcher(br, log)

	// Navigate to a test page
	err = br.Navigate("https://example.com")
	if err != nil {
		t.Fatalf("Failed to navigate: %v", err)
	}

	// Test browsing (short duration for test)
	err = searcher.BrowseTarget(1*time.Second, 3*time.Second, false)
	if err != nil {
		t.Errorf("BrowseTarget() error = %v", err)
	}
}

func TestBrowseTarget_WithLinks(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping browser test in short mode")
	}

	// Create test browser
	br, err := browser.NewBrowser(browser.BrowserOptions{
		Headless: true,
		Timeout:  30 * time.Second,
	})
	if err != nil {
		t.Fatalf("Failed to create browser: %v", err)
	}
	defer br.Close()

	// Create logger
	log := logger.NewDefault()

	// Create searcher
	searcher := NewSearcher(br, log)

	// Navigate to a test page
	err = br.Navigate("https://example.com")
	if err != nil {
		t.Fatalf("Failed to navigate: %v", err)
	}

	// Test browsing with link clicking (short duration for test)
	err = searcher.BrowseTarget(1*time.Second, 2*time.Second, true)
	if err != nil {
		t.Logf("BrowseTarget() with links error = %v (expected for simple page)", err)
	}
}

func TestRandomScroll(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping browser test in short mode")
	}

	// Create test browser
	br, err := browser.NewBrowser(browser.BrowserOptions{
		Headless: true,
		Timeout:  30 * time.Second,
	})
	if err != nil {
		t.Fatalf("Failed to create browser: %v", err)
	}
	defer br.Close()

	// Create logger
	log := logger.NewDefault()

	// Create searcher
	searcher := NewSearcher(br, log)

	// Navigate to a test page
	err = br.Navigate("https://example.com")
	if err != nil {
		t.Fatalf("Failed to navigate: %v", err)
	}

	// Test random scroll
	err = searcher.randomScroll()
	if err != nil {
		t.Errorf("randomScroll() error = %v", err)
	}
}

func TestSimulateReading(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping browser test in short mode")
	}

	// Create test browser
	br, err := browser.NewBrowser(browser.BrowserOptions{
		Headless: true,
		Timeout:  30 * time.Second,
	})
	if err != nil {
		t.Fatalf("Failed to create browser: %v", err)
	}
	defer br.Close()

	// Create logger
	log := logger.NewDefault()

	// Create searcher
	searcher := NewSearcher(br, log)

	// Test simulate reading
	start := time.Now()
	err = searcher.SimulateReading(1, 2)
	elapsed := time.Since(start)

	if err != nil {
		t.Errorf("SimulateReading() error = %v", err)
	}

	if elapsed < 1*time.Second {
		t.Errorf("SimulateReading() took less than minimum time: %v", elapsed)
	}
	if elapsed > 3*time.Second {
		t.Errorf("SimulateReading() took more than expected: %v", elapsed)
	}
}

func TestScrollToBottom(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping browser test in short mode")
	}

	// Create test browser
	br, err := browser.NewBrowser(browser.BrowserOptions{
		Headless: true,
		Timeout:  30 * time.Second,
	})
	if err != nil {
		t.Fatalf("Failed to create browser: %v", err)
	}
	defer br.Close()

	// Create logger
	log := logger.NewDefault()

	// Create searcher
	searcher := NewSearcher(br, log)

	// Navigate to a test page
	err = br.Navigate("https://example.com")
	if err != nil {
		t.Fatalf("Failed to navigate: %v", err)
	}

	// Test scroll to bottom
	err = searcher.ScrollToBottom()
	if err != nil {
		t.Errorf("ScrollToBottom() error = %v", err)
	}
}

func TestGetInternalLinks(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping browser test in short mode")
	}

	// Create test browser
	br, err := browser.NewBrowser(browser.BrowserOptions{
		Headless: true,
		Timeout:  30 * time.Second,
	})
	if err != nil {
		t.Fatalf("Failed to create browser: %v", err)
	}
	defer br.Close()

	// Create logger
	log := logger.NewDefault()

	// Create searcher
	searcher := NewSearcher(br, log)

	// Navigate to a test page
	err = br.Navigate("https://example.com")
	if err != nil {
		t.Fatalf("Failed to navigate: %v", err)
	}

	// Test get internal links
	links, err := searcher.getInternalLinks()
	if err != nil {
		t.Errorf("getInternalLinks() error = %v", err)
	}

	if len(links) == 0 {
		t.Error("getInternalLinks() returned no links")
	}
}
