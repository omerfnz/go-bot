package proxy

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// ProxyValidator validates proxy connectivity
type ProxyValidator struct {
	testURL    string
	timeout    time.Duration
	httpClient *http.Client
}

// NewProxyValidator creates a new proxy validator with the given test URL and timeout.
// If testURL is empty, it defaults to "https://www.google.com".
//
// Example:
//
//	validator := NewProxyValidator("https://www.google.com", 10*time.Second)
func NewProxyValidator(testURL string, timeout time.Duration) *ProxyValidator {
	if testURL == "" {
		testURL = "https://www.google.com"
	}
	if timeout == 0 {
		timeout = 10 * time.Second
	}

	return &ProxyValidator{
		testURL: testURL,
		timeout: timeout,
	}
}

// Validate tests if a proxy is working by making an HTTP GET request through it.
// It returns nil if the proxy is working, otherwise returns an error.
//
// Example:
//
//	validator := NewProxyValidator("https://www.google.com", 10*time.Second)
//	proxy, _ := ParseProxy("http://proxy.example.com:8080")
//	err := validator.Validate(proxy)
//	if err != nil {
//	    log.Printf("Proxy validation failed: %v", err)
//	}
func (pv *ProxyValidator) Validate(proxy *Proxy) error {
	if proxy == nil {
		return fmt.Errorf("proxy cannot be nil")
	}

	if proxy.IsBlacklisted {
		return fmt.Errorf("proxy is blacklisted")
	}

	// Parse proxy URL
	proxyURL, err := url.Parse(proxy.URL)
	if err != nil {
		return fmt.Errorf("failed to parse proxy URL: %w", err)
	}

	// Create HTTP client with proxy
	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   pv.timeout,
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), pv.timeout)
	defer cancel()

	// Create request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pv.testURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set realistic headers
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("proxy validation failed: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return fmt.Errorf("proxy returned status code %d", resp.StatusCode)
	}

	return nil
}

// ValidateAll validates all proxies in a list concurrently and returns a map of results.
// The map key is the proxy URL and the value is the validation error (nil if valid).
//
// Example:
//
//	validator := NewProxyValidator("https://www.google.com", 10*time.Second)
//	proxies := []*Proxy{proxy1, proxy2, proxy3}
//	results := validator.ValidateAll(context.Background(), proxies)
//	for proxyURL, err := range results {
//	    if err != nil {
//	        log.Printf("Proxy %s failed: %v", proxyURL, err)
//	    }
//	}
func (pv *ProxyValidator) ValidateAll(ctx context.Context, proxies []*Proxy) map[string]error {
	results := make(map[string]error)
	resultsChan := make(chan struct {
		url string
		err error
	}, len(proxies))

	// Validate proxies concurrently
	for _, proxy := range proxies {
		go func(p *Proxy) {
			err := pv.Validate(p)
			resultsChan <- struct {
				url string
				err error
			}{url: p.URL, err: err}
		}(proxy)
	}

	// Collect results
	for i := 0; i < len(proxies); i++ {
		select {
		case result := <-resultsChan:
			results[result.url] = result.err
		case <-ctx.Done():
			results["timeout"] = ctx.Err()
			return results
		}
	}

	return results
}

// ValidateWithRetry validates a proxy with retry logic.
// It attempts to validate the proxy up to maxRetries times with exponential backoff.
//
// Example:
//
//	validator := NewProxyValidator("https://www.google.com", 10*time.Second)
//	proxy, _ := ParseProxy("http://proxy.example.com:8080")
//	err := validator.ValidateWithRetry(context.Background(), proxy, 3)
func (pv *ProxyValidator) ValidateWithRetry(ctx context.Context, proxy *Proxy, maxRetries int) error {
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		// Check context
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Try validation
		err := pv.Validate(proxy)
		if err == nil {
			return nil
		}

		lastErr = err

		// Wait before retry (exponential backoff)
		if attempt < maxRetries-1 {
			backoff := time.Duration(1<<uint(attempt)) * time.Second
			timer := time.NewTimer(backoff)
			select {
			case <-timer.C:
			case <-ctx.Done():
				timer.Stop()
				return ctx.Err()
			}
		}
	}

	return fmt.Errorf("proxy validation failed after %d attempts: %w", maxRetries, lastErr)
}

// QuickValidate performs a quick validation check without full HTTP request.
// It only checks if the proxy URL is well-formed and not blacklisted.
func (pv *ProxyValidator) QuickValidate(proxy *Proxy) error {
	if proxy == nil {
		return fmt.Errorf("proxy cannot be nil")
	}

	if proxy.IsBlacklisted {
		return fmt.Errorf("proxy is blacklisted")
	}

	if proxy.Host == "" {
		return fmt.Errorf("proxy host is empty")
	}

	if proxy.Port < 1 || proxy.Port > 65535 {
		return fmt.Errorf("invalid proxy port: %d", proxy.Port)
	}

	return nil
}
