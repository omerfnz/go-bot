package browser

import (
	"fmt"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/omer/go-bot/pkg/utils"
)

// Type types text into an element identified by the CSS selector.
// It waits for the element to be visible before typing.
//
// Example:
//
//	err := browser.Type("input[name='q']", "golang tutorial")
func (b *Browser) Type(selector, text string) error {
	if selector == "" {
		return fmt.Errorf("selector cannot be empty")
	}

	return chromedp.Run(b.ctx,
		chromedp.WaitVisible(selector),
		chromedp.Clear(selector),
		chromedp.SendKeys(selector, text),
	)
}

// Click clicks on an element identified by the CSS selector.
// It waits for the element to be visible before clicking.
//
// Example:
//
//	err := browser.Click("button[type='submit']")
func (b *Browser) Click(selector string) error {
	if selector == "" {
		return fmt.Errorf("selector cannot be empty")
	}

	return chromedp.Run(b.ctx,
		chromedp.WaitVisible(selector),
		chromedp.Click(selector),
	)
}

// WaitVisible waits for an element to become visible.
// It returns an error if the element doesn't appear within the timeout.
//
// Example:
//
//	err := browser.WaitVisible("div.search-results")
func (b *Browser) WaitVisible(selector string) error {
	if selector == "" {
		return fmt.Errorf("selector cannot be empty")
	}

	return chromedp.Run(b.ctx,
		chromedp.WaitVisible(selector),
	)
}

// WaitNotVisible waits for an element to become invisible or be removed from the DOM.
//
// Example:
//
//	err := browser.WaitNotVisible("div.loading")
func (b *Browser) WaitNotVisible(selector string) error {
	if selector == "" {
		return fmt.Errorf("selector cannot be empty")
	}

	return chromedp.Run(b.ctx,
		chromedp.WaitNotVisible(selector),
	)
}

// GetText retrieves the text content of an element.
//
// Example:
//
//	text, err := browser.GetText("h1.title")
func (b *Browser) GetText(selector string) (string, error) {
	if selector == "" {
		return "", fmt.Errorf("selector cannot be empty")
	}

	var text string
	err := chromedp.Run(b.ctx,
		chromedp.WaitVisible(selector),
		chromedp.Text(selector, &text),
	)
	if err != nil {
		return "", err
	}

	return text, nil
}

// GetAttribute retrieves an attribute value from an element.
//
// Example:
//
//	href, err := browser.GetAttribute("a.link", "href")
func (b *Browser) GetAttribute(selector, attribute string) (string, error) {
	if selector == "" {
		return "", fmt.Errorf("selector cannot be empty")
	}
	if attribute == "" {
		return "", fmt.Errorf("attribute cannot be empty")
	}

	var value string
	err := chromedp.Run(b.ctx,
		chromedp.WaitVisible(selector),
		chromedp.AttributeValue(selector, attribute, &value, nil),
	)
	if err != nil {
		return "", err
	}

	return value, nil
}

// ElementExists checks if an element exists in the DOM (doesn't need to be visible).
//
// Example:
//
//	exists := browser.ElementExists("div#captcha")
func (b *Browser) ElementExists(selector string) bool {
	if selector == "" {
		return false
	}

	var exists bool
	err := chromedp.Run(b.ctx,
		chromedp.Evaluate(fmt.Sprintf("document.querySelector('%s') !== null", selector), &exists),
	)

	return err == nil && exists
}

// Sleep pauses execution for the specified duration.
//
// Example:
//
//	browser.Sleep(2 * time.Second)
func (b *Browser) Sleep(duration time.Duration) error {
	return chromedp.Run(b.ctx,
		chromedp.Sleep(duration),
	)
}

// Scroll scrolls the page by the specified pixel amounts.
//
// Example:
//
//	err := browser.Scroll(0, 500) // Scroll down 500px
func (b *Browser) Scroll(x, y int) error {
	return chromedp.Run(b.ctx,
		chromedp.Evaluate(fmt.Sprintf("window.scrollBy(%d, %d)", x, y), nil),
	)
}

// ScrollToElement scrolls to make an element visible in the viewport.
//
// Example:
//
//	err := browser.ScrollToElement("div.footer")
func (b *Browser) ScrollToElement(selector string) error {
	if selector == "" {
		return fmt.Errorf("selector cannot be empty")
	}

	return chromedp.Run(b.ctx,
		chromedp.ScrollIntoView(selector),
	)
}

// Screenshot takes a screenshot of the entire page and returns the image data.
//
// Example:
//
//	imageData, err := browser.Screenshot()
//	if err == nil {
//	    os.WriteFile("screenshot.png", imageData, 0644)
//	}
func (b *Browser) Screenshot() ([]byte, error) {
	var buf []byte
	err := chromedp.Run(b.ctx,
		chromedp.FullScreenshot(&buf, 90),
	)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

// GetCurrentURL returns the current page URL.
//
// Example:
//
//	url, err := browser.GetCurrentURL()
func (b *Browser) GetCurrentURL() (string, error) {
	var url string
	err := chromedp.Run(b.ctx,
		chromedp.Evaluate("window.location.href", &url),
	)
	if err != nil {
		return "", err
	}

	return url, nil
}

// GetTitle returns the current page title.
//
// Example:
//
//	title, err := browser.GetTitle()
func (b *Browser) GetTitle() (string, error) {
	var title string
	err := chromedp.Run(b.ctx,
		chromedp.Title(&title),
	)
	if err != nil {
		return "", err
	}

	return title, nil
}

// Reload reloads the current page.
//
// Example:
//
//	err := browser.Reload()
func (b *Browser) Reload() error {
	return chromedp.Run(b.ctx,
		chromedp.Reload(),
	)
}

// GoBack navigates back in browser history.
//
// Example:
//
//	err := browser.GoBack()
func (b *Browser) GoBack() error {
	return chromedp.Run(b.ctx,
		chromedp.NavigateBack(),
	)
}

// GoForward navigates forward in browser history.
//
// Example:
//
//	err := browser.GoForward()
func (b *Browser) GoForward() error {
	return chromedp.Run(b.ctx,
		chromedp.NavigateForward(),
	)
}

// TypeHumanLike types text character by character with random delays to simulate human typing.
// This helps avoid bot detection by mimicking natural typing patterns.
//
// Example:
//
//	err := browser.TypeHumanLike("input[name='q']", "golang tutorial")
func (b *Browser) TypeHumanLike(selector, text string) error {
	if selector == "" {
		return fmt.Errorf("selector cannot be empty")
	}

	// Wait for element and clear it
	if err := chromedp.Run(b.ctx,
		chromedp.WaitVisible(selector),
		chromedp.Clear(selector),
		chromedp.Click(selector),
	); err != nil {
		return err
	}

	// Type each character with random delay
	for _, char := range text {
		delay := utils.RandomDuration(50*time.Millisecond, 200*time.Millisecond)

		if err := chromedp.Run(b.ctx,
			chromedp.SendKeys(selector, string(char)),
			chromedp.Sleep(delay),
		); err != nil {
			return err
		}
	}

	return nil
}

// ClickWithDelay clicks an element after waiting for a random human-like delay.
// This helps simulate natural user behavior.
//
// Example:
//
//	err := browser.ClickWithDelay("button[type='submit']", 1*time.Second, 3*time.Second)
func (b *Browser) ClickWithDelay(selector string, minDelay, maxDelay time.Duration) error {
	if selector == "" {
		return fmt.Errorf("selector cannot be empty")
	}

	delay := utils.RandomDuration(minDelay, maxDelay)

	return chromedp.Run(b.ctx,
		chromedp.WaitVisible(selector),
		chromedp.Sleep(delay),
		chromedp.Click(selector),
	)
}

// ScrollRandom scrolls the page by random amounts to simulate human browsing.
// It scrolls down with some variation in scroll distance.
//
// Example:
//
//	err := browser.ScrollRandom(3, 500, 1000)
func (b *Browser) ScrollRandom(times int, minPixels, maxPixels int) error {
	for i := 0; i < times; i++ {
		pixels := utils.RandomInt(minPixels, maxPixels)
		delay := utils.RandomDuration(500*time.Millisecond, 2*time.Second)

		if err := chromedp.Run(b.ctx,
			chromedp.Evaluate(fmt.Sprintf("window.scrollBy(0, %d)", pixels), nil),
			chromedp.Sleep(delay),
		); err != nil {
			return err
		}
	}

	return nil
}

// WaitRandom waits for a random duration between min and max.
// Useful for simulating human reading/thinking time.
//
// Example:
//
//	err := browser.WaitRandom(2*time.Second, 5*time.Second)
func (b *Browser) WaitRandom(min, max time.Duration) error {
	delay := utils.RandomDuration(min, max)
	return b.Sleep(delay)
}

// MouseMoveToElement moves the mouse to an element before interacting with it.
// This can help bypass some bot detection systems.
//
// Example:
//
//	err := browser.MouseMoveToElement("button#submit")
func (b *Browser) MouseMoveToElement(selector string) error {
	if selector == "" {
		return fmt.Errorf("selector cannot be empty")
	}

	// Move mouse to element center
	script := fmt.Sprintf(`
		const element = document.querySelector('%s');
		if (element) {
			const rect = element.getBoundingClientRect();
			const x = rect.left + rect.width / 2;
			const y = rect.top + rect.height / 2;
			const event = new MouseEvent('mousemove', {
				view: window,
				bubbles: true,
				cancelable: true,
				clientX: x,
				clientY: y
			});
			element.dispatchEvent(event);
		}
	`, selector)

	return chromedp.Run(b.ctx,
		chromedp.WaitVisible(selector),
		chromedp.Evaluate(script, nil),
	)
}

// ScrollToElementSmoothly scrolls to an element smoothly with a delay.
// This is more human-like than instant scrolling.
//
// Example:
//
//	err := browser.ScrollToElementSmoothly("div.footer")
func (b *Browser) ScrollToElementSmoothly(selector string) error {
	if selector == "" {
		return fmt.Errorf("selector cannot be empty")
	}

	script := fmt.Sprintf(`
		const element = document.querySelector('%s');
		if (element) {
			element.scrollIntoView({ behavior: 'smooth', block: 'center' });
		}
	`, selector)

	delay := utils.RandomDuration(500*time.Millisecond, 1500*time.Millisecond)

	return chromedp.Run(b.ctx,
		chromedp.Evaluate(script, nil),
		chromedp.Sleep(delay),
	)
}

// HoverElement simulates hovering over an element.
// This can trigger hover effects and appear more human-like.
//
// Example:
//
//	err := browser.HoverElement("a.menu-item")
func (b *Browser) HoverElement(selector string) error {
	if selector == "" {
		return fmt.Errorf("selector cannot be empty")
	}

	script := fmt.Sprintf(`
		const element = document.querySelector('%s');
		if (element) {
			const event = new MouseEvent('mouseover', {
				view: window,
				bubbles: true,
				cancelable: true
			});
			element.dispatchEvent(event);
		}
	`, selector)

	return chromedp.Run(b.ctx,
		chromedp.WaitVisible(selector),
		chromedp.Evaluate(script, nil),
		chromedp.Sleep(utils.RandomDuration(100*time.Millisecond, 500*time.Millisecond)),
	)
}
