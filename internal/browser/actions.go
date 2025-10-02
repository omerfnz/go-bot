package browser

import (
	"fmt"
	"time"

	"github.com/chromedp/chromedp"
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
