package serp

import (
	"fmt"
	"time"

	"github.com/omer/go-bot/pkg/utils"
)

// BrowseTarget simulates human browsing behavior on the target website.
// It performs random scrolling, optional link clicking, and realistic wait times.
//
// Parameters:
//   - minDuration: minimum time to spend on the site
//   - maxDuration: maximum time to spend on the site
//   - clickLinks: whether to randomly click internal links
func (s *Searcher) BrowseTarget(minDuration, maxDuration time.Duration, clickLinks bool) error {
	s.logger.Info("Starting human-like browsing on target site")

	// Calculate total browsing time
	browseTime := utils.RandomDuration(minDuration, maxDuration)
	s.logger.WithField("duration", browseTime).Debug("Browse duration selected")

	endTime := time.Now().Add(browseTime)

	// Perform initial scroll down
	if err := s.browser.ScrollRandom(utils.RandomInt(2, 4), 300, 600); err != nil {
		return fmt.Errorf("initial scroll failed: %w", err)
	}

	// Wait and read
	if err := s.browser.WaitRandom(2*time.Second, 5*time.Second); err != nil {
		return fmt.Errorf("wait failed: %w", err)
	}

	// Continue browsing until time is up
	iterations := 0
	maxIterations := 20 // Prevent infinite loop

	for time.Now().Before(endTime) && iterations < maxIterations {
		iterations++

		// Random action: scroll or click link
		action := utils.RandomInt(1, 100)

		switch {
		case action <= 60: // 60% chance to scroll
			if err := s.randomScroll(); err != nil {
				s.logger.WithError(err).Warn("Random scroll failed")
			}

		case action <= 90 && clickLinks: // 30% chance to click a link (if enabled)
			if err := s.randomLinkClick(); err != nil {
				s.logger.WithError(err).Debug("Random link click failed (might be normal)")
			}

		default: // 10% chance to just wait
			if err := s.browser.WaitRandom(3*time.Second, 7*time.Second); err != nil {
				return fmt.Errorf("wait failed: %w", err)
			}
		}

		// Random wait between actions
		if err := s.browser.WaitRandom(1*time.Second, 3*time.Second); err != nil {
			return fmt.Errorf("wait between actions failed: %w", err)
		}

		// Check if we should stop
		if time.Now().After(endTime) {
			break
		}
	}

	// Final scroll up a bit (human behavior)
	if utils.RandomBool() {
		if err := s.browser.Scroll(0, -utils.RandomInt(100, 300)); err != nil {
			s.logger.WithError(err).Debug("Final scroll up failed")
		}
	}

	s.logger.WithField("iterations", iterations).Info("Finished browsing target site")
	return nil
}

// randomScroll performs a random scroll action
func (s *Searcher) randomScroll() error {
	// Decide scroll direction (mostly down, sometimes up)
	scrollDown := utils.RandomInt(1, 100) <= 80 // 80% down, 20% up

	if scrollDown {
		return s.browser.ScrollRandom(utils.RandomInt(1, 3), 200, 500)
	}

	// Scroll up
	pixels := utils.RandomInt(100, 300)
	return s.browser.Scroll(0, -pixels)
}

// randomLinkClick attempts to click a random internal link
func (s *Searcher) randomLinkClick() error {
	// Get all links on the page
	links, err := s.getInternalLinks()
	if err != nil {
		return fmt.Errorf("failed to get links: %w", err)
	}

	if len(links) == 0 {
		return fmt.Errorf("no internal links found")
	}

	// Select a random link
	linkIndex := utils.RandomInt(0, len(links)-1)
	selector := fmt.Sprintf("a:nth-of-type(%d)", linkIndex+1)

	s.logger.WithField("selector", selector).Debug("Attempting to click random link")

	// Scroll to link first
	if err := s.browser.ScrollToElementSmoothly(selector); err != nil {
		return fmt.Errorf("failed to scroll to link: %w", err)
	}

	// Hover over link
	if err := s.browser.HoverElement(selector); err != nil {
		s.logger.WithError(err).Debug("Hover failed")
	}

	// Click with delay
	if err := s.browser.ClickWithDelay(selector, 500*time.Millisecond, 2*time.Second); err != nil {
		return fmt.Errorf("failed to click link: %w", err)
	}

	// Wait for page load
	if err := s.browser.WaitRandom(2*time.Second, 4*time.Second); err != nil {
		return fmt.Errorf("wait after click failed: %w", err)
	}

	return nil
}

// getInternalLinks returns a slice of internal link selectors
func (s *Searcher) getInternalLinks() ([]string, error) {
	// Get current URL to determine internal links
	currentURL, err := s.browser.GetCurrentURL()
	if err != nil {
		return nil, fmt.Errorf("failed to get current URL: %w", err)
	}

	// For now, just return a simple selector
	// In a more advanced version, we would filter links by domain
	s.logger.WithField("currentURL", currentURL).Debug("Getting internal links")

	// Simple approach: return indices for up to 10 links
	// This is a placeholder - real implementation would need to evaluate JS
	var links []string
	for i := 0; i < 10; i++ {
		links = append(links, fmt.Sprintf("a:nth-of-type(%d)", i+1))
	}

	return links, nil
}

// SimulateReading waits for a random "reading" time
func (s *Searcher) SimulateReading(minSeconds, maxSeconds int) error {
	duration := utils.RandomDuration(
		time.Duration(minSeconds)*time.Second,
		time.Duration(maxSeconds)*time.Second,
	)

	s.logger.WithField("duration", duration).Debug("Simulating reading time")
	return s.browser.Sleep(duration)
}

// ScrollToBottom scrolls to the bottom of the page smoothly
func (s *Searcher) ScrollToBottom() error {
	s.logger.Debug("Scrolling to page bottom")

	// Fallback: scroll in increments (page height detection is complex)
	return s.browser.ScrollRandom(10, 400, 600)
}
