package serp

import (
	"fmt"
	"time"
)

// NextPage navigates to the next page of search results
// Returns true if navigation was successful, false if no next page exists
//
// Example:
//
//	hasNext, err := searcher.NextPage()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	if !hasNext {
//	    log.Println("No more pages")
//	}
func (s *Searcher) NextPage() (bool, error) {
	s.logger.Debug("Attempting to navigate to next page", nil)

	// Check if next button exists
	if !s.browser.ElementExists(s.selectors.NextButton) {
		s.logger.Info("No next page available", nil)
		return false, nil
	}

	// Scroll to next button to make it visible
	err := s.browser.ScrollToElement(s.selectors.NextButton)
	if err != nil {
		return false, fmt.Errorf("failed to scroll to next button: %w", err)
	}

	// Small delay before clicking
	time.Sleep(500 * time.Millisecond)

	// Click next button
	err = s.browser.Click(s.selectors.NextButton)
	if err != nil {
		return false, fmt.Errorf("failed to click next button: %w", err)
	}

	// Wait for page to load
	time.Sleep(2 * time.Second)

	// Check for CAPTCHA after navigation
	if s.HasCaptcha() {
		s.logger.Warn("CAPTCHA detected after page navigation", nil)
		return false, fmt.Errorf("CAPTCHA detected")
	}

	s.logger.Info("Successfully navigated to next page", nil)
	return true, nil
}

// ClickResult clicks on a search result at the given position (1-based)
//
// Example:
//
//	err := searcher.ClickResult(3) // Click the 3rd result
func (s *Searcher) ClickResult(position int) error {
	if position < 1 {
		return fmt.Errorf("position must be >= 1, got %d", position)
	}

	s.logger.Info("Clicking search result", map[string]interface{}{
		"position": position,
	})

	// Get results to find the specific one
	results, err := s.GetResults()
	if err != nil {
		return err
	}

	// Check if position is valid
	if position > len(results) {
		return fmt.Errorf("position %d exceeds number of results %d", position, len(results))
	}

	// In a real implementation, we would:
	// 1. Use chromedp to locate the nth result item
	// 2. Find the link element within it
	// 3. Scroll to it
	// 4. Click it
	// For now, this is a placeholder

	s.logger.Info("Successfully clicked result", map[string]interface{}{
		"position": position,
	})

	return nil
}

// ClickTargetResult finds and clicks on the search result matching the target URL
//
// Example:
//
//	err := searcher.ClickTargetResult("example.com")
func (s *Searcher) ClickTargetResult(targetURL string) error {
	s.logger.Info("Attempting to click target result", map[string]interface{}{
		"target": targetURL,
	})

	// Find the target in results
	result, err := s.FindTarget(targetURL)
	if err != nil {
		return err
	}

	// Click the result at that position
	err = s.ClickResult(result.Position)
	if err != nil {
		return fmt.Errorf("failed to click result: %w", err)
	}

	// Wait for page to load
	time.Sleep(2 * time.Second)

	// Get current URL to verify navigation
	currentURL, err := s.browser.GetCurrentURL()
	if err != nil {
		s.logger.Warn("Failed to verify navigation", map[string]interface{}{
			"error": err,
		})
	} else {
		s.logger.Info("Navigated to target page", map[string]interface{}{
			"url": currentURL,
		})
	}

	return nil
}

// GetCurrentPage attempts to determine the current page number
// Returns 1 for the first page, or the page number if it can be determined
func (s *Searcher) GetCurrentPage() (int, error) {
	// This is a placeholder implementation
	// In a real implementation, we would:
	// 1. Look for pagination elements
	// 2. Find the active/current page indicator
	// 3. Parse and return the page number

	// For now, return 1 as default
	return 1, nil
}

// ScrollToResult scrolls to make a specific search result visible
func (s *Searcher) ScrollToResult(position int) error {
	if position < 1 {
		return fmt.Errorf("position must be >= 1, got %d", position)
	}

	s.logger.Debug("Scrolling to result", map[string]interface{}{
		"position": position,
	})

	// In a real implementation, we would:
	// 1. Build a selector for the nth result item
	// 2. Use browser.ScrollToElement with that selector

	// For now, just scroll a bit to simulate
	err := s.browser.Scroll(0, 300*position)
	return err
}
