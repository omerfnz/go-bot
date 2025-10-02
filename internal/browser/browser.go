// Package browser provides browser automation functionality using chromedp.
// It wraps chromedp for easier usage and lifecycle management.
package browser

import (
	"context"
	"fmt"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/omer/go-bot/internal/proxy"
)

// Browser represents a browser instance with chromedp
type Browser struct {
	ctx         context.Context
	cancel      context.CancelFunc
	ctxCancel   context.CancelFunc
	allocCtx    context.Context
	allocCancel context.CancelFunc
	proxy       *proxy.Proxy
	userAgent   string
	headless    bool
}

// BrowserOptions holds configuration options for creating a browser instance
type BrowserOptions struct {
	Headless  bool          // Run browser in headless mode
	Proxy     *proxy.Proxy  // Proxy to use (optional)
	UserAgent string        // Custom user agent (optional)
	Timeout   time.Duration // Context timeout (default: 30s)
}

// NewBrowser creates a new browser instance with the given options.
// It initializes chromedp context and returns a Browser ready to use.
// Remember to call Close() when done.
//
// Example:
//
//	browser, err := NewBrowser(BrowserOptions{
//	    Headless: true,
//	    Timeout: 30 * time.Second,
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer browser.Close()
func NewBrowser(opts BrowserOptions) (*Browser, error) {
	// Set default timeout
	if opts.Timeout == 0 {
		opts.Timeout = 30 * time.Second
	}

	// Set default user agent
	if opts.UserAgent == "" {
		opts.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
	}

	// Prepare chromedp options
	allocOpts := []chromedp.ExecAllocatorOption{
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.UserAgent(opts.UserAgent),
	}

	// Add headless option
	if opts.Headless {
		allocOpts = append(allocOpts, chromedp.Headless)
	}

	// Add proxy if provided
	if opts.Proxy != nil {
		proxyURL := fmt.Sprintf("%s://%s:%d", opts.Proxy.Type, opts.Proxy.Host, opts.Proxy.Port)
		allocOpts = append(allocOpts, chromedp.ProxyServer(proxyURL))
	}

	// Create allocator context
	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), allocOpts...)

	// Create browser context
	ctx, ctxCancel := chromedp.NewContext(allocCtx)

	// Add timeout
	ctx, cancel := context.WithTimeout(ctx, opts.Timeout)

	browser := &Browser{
		ctx:         ctx,
		cancel:      cancel,
		ctxCancel:   ctxCancel,
		allocCtx:    allocCtx,
		allocCancel: allocCancel,
		proxy:       opts.Proxy,
		userAgent:   opts.UserAgent,
		headless:    opts.Headless,
	}

	return browser, nil
}

// Navigate navigates to the specified URL.
// It waits for the page to be ready before returning.
//
// Example:
//
//	err := browser.Navigate("https://www.google.com")
//	if err != nil {
//	    log.Fatal(err)
//	}
func (b *Browser) Navigate(url string) error {
	if url == "" {
		return fmt.Errorf("URL cannot be empty")
	}

	return chromedp.Run(b.ctx,
		chromedp.Navigate(url),
		chromedp.WaitReady("body"),
	)
}

// GetContext returns the browser's context.
// This can be used for advanced chromedp operations.
func (b *Browser) GetContext() context.Context {
	return b.ctx
}

// Close closes the browser and cleans up all resources.
// It's safe to call Close multiple times.
func (b *Browser) Close() error {
	if b.cancel != nil {
		b.cancel()
		b.cancel = nil
	}
	if b.ctxCancel != nil {
		b.ctxCancel()
		b.ctxCancel = nil
	}
	if b.allocCancel != nil {
		b.allocCancel()
		b.allocCancel = nil
	}
	return nil
}

// IsHeadless returns whether the browser is running in headless mode
func (b *Browser) IsHeadless() bool {
	return b.headless
}

// GetUserAgent returns the browser's user agent string
func (b *Browser) GetUserAgent() string {
	return b.userAgent
}

// GetProxy returns the browser's proxy configuration
func (b *Browser) GetProxy() *proxy.Proxy {
	return b.proxy
}
