package proxy

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// RotationStrategy defines the strategy for proxy rotation
type RotationStrategy string

const (
	// RotationStrategyRoundRobin rotates proxies in a round-robin fashion
	RotationStrategyRoundRobin RotationStrategy = "round-robin"
	// RotationStrategyRandom selects proxies randomly (v1.1)
	RotationStrategyRandom RotationStrategy = "random"
)

// ProxyPool manages a pool of proxies with rotation strategy
type ProxyPool struct {
	proxies   []*Proxy
	current   int
	strategy  RotationStrategy
	mu        sync.Mutex
	blacklist map[string]bool
}

// NewProxyPool creates a new proxy pool with the given proxy URLs and rotation strategy.
// It parses all proxy URLs and returns an error if any URL is invalid.
//
// Example:
//
//	pool, err := NewProxyPool(
//	    []string{"http://proxy1.com:8080", "http://proxy2.com:8080"},
//	    RotationStrategyRoundRobin,
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
func NewProxyPool(proxyURLs []string, strategy RotationStrategy) (*ProxyPool, error) {
	if len(proxyURLs) == 0 {
		return nil, fmt.Errorf("at least one proxy URL is required")
	}

	// Validate strategy
	if strategy != RotationStrategyRoundRobin && strategy != RotationStrategyRandom {
		return nil, fmt.Errorf("invalid rotation strategy: %s", strategy)
	}

	proxies := make([]*Proxy, 0, len(proxyURLs))
	for i, proxyURL := range proxyURLs {
		proxy, err := ParseProxy(proxyURL)
		if err != nil {
			return nil, fmt.Errorf("failed to parse proxy[%d]: %w", i, err)
		}
		proxies = append(proxies, proxy)
	}

	pool := &ProxyPool{
		proxies:   proxies,
		current:   0,
		strategy:  strategy,
		blacklist: make(map[string]bool),
	}

	return pool, nil
}

// Get returns the next available proxy according to the rotation strategy.
// It skips blacklisted proxies and returns an error if no healthy proxy is available.
func (pp *ProxyPool) Get() (*Proxy, error) {
	pp.mu.Lock()
	defer pp.mu.Unlock()

	if len(pp.proxies) == 0 {
		return nil, fmt.Errorf("proxy pool is empty")
	}

	// Try to find a healthy proxy
	attempts := 0
	maxAttempts := len(pp.proxies)

	for attempts < maxAttempts {
		var proxy *Proxy

		switch pp.strategy {
		case RotationStrategyRoundRobin:
			proxy = pp.proxies[pp.current]
			pp.current = (pp.current + 1) % len(pp.proxies)
		case RotationStrategyRandom:
			// Use a seeded random number generator
			r := rand.New(rand.NewSource(time.Now().UnixNano()))
			idx := r.Intn(len(pp.proxies))
			proxy = pp.proxies[idx]
		}

		// Check if proxy is healthy
		if proxy.IsHealthy() {
			return proxy, nil
		}

		attempts++
	}

	return nil, fmt.Errorf("no healthy proxy available in pool")
}

// Release returns a proxy to the pool and records the result.
// If success is true, it records a successful request; otherwise, it records a failure.
func (pp *ProxyPool) Release(proxy *Proxy, success bool) {
	pp.mu.Lock()
	defer pp.mu.Unlock()

	if success {
		proxy.RecordSuccess()
	} else {
		proxy.RecordFailure()
		// Auto-blacklist if too many failures
		if proxy.FailCount >= 5 {
			pp.blacklist[proxy.URL] = true
			proxy.IsBlacklisted = true
		}
	}
}

// Blacklist marks a proxy as blacklisted and removes it from rotation.
func (pp *ProxyPool) Blacklist(proxy *Proxy) {
	pp.mu.Lock()
	defer pp.mu.Unlock()

	pp.blacklist[proxy.URL] = true
	proxy.IsBlacklisted = true
}

// RemoveFromBlacklist removes a proxy from the blacklist.
func (pp *ProxyPool) RemoveFromBlacklist(proxy *Proxy) {
	pp.mu.Lock()
	defer pp.mu.Unlock()

	delete(pp.blacklist, proxy.URL)
	proxy.IsBlacklisted = false
	proxy.FailCount = 0
}

// IsBlacklisted checks if a proxy URL is blacklisted.
func (pp *ProxyPool) IsBlacklisted(proxyURL string) bool {
	pp.mu.Lock()
	defer pp.mu.Unlock()

	return pp.blacklist[proxyURL]
}

// GetProxies returns a copy of all proxies in the pool.
func (pp *ProxyPool) GetProxies() []*Proxy {
	pp.mu.Lock()
	defer pp.mu.Unlock()

	// Return a copy to prevent external modification
	proxies := make([]*Proxy, len(pp.proxies))
	copy(proxies, pp.proxies)
	return proxies
}

// Size returns the number of proxies in the pool.
func (pp *ProxyPool) Size() int {
	pp.mu.Lock()
	defer pp.mu.Unlock()

	return len(pp.proxies)
}

// HealthyCount returns the number of healthy (non-blacklisted) proxies.
func (pp *ProxyPool) HealthyCount() int {
	pp.mu.Lock()
	defer pp.mu.Unlock()

	count := 0
	for _, proxy := range pp.proxies {
		if proxy.IsHealthy() {
			count++
		}
	}
	return count
}

// GetStats returns statistics about the proxy pool.
func (pp *ProxyPool) GetStats() map[string]interface{} {
	pp.mu.Lock()
	defer pp.mu.Unlock()

	totalSuccess := 0
	totalFail := 0
	blacklistedCount := 0

	for _, proxy := range pp.proxies {
		totalSuccess += proxy.SuccessCount
		totalFail += proxy.FailCount
		if proxy.IsBlacklisted {
			blacklistedCount++
		}
	}

	stats := map[string]interface{}{
		"total_proxies":       len(pp.proxies),
		"healthy_proxies":     len(pp.proxies) - blacklistedCount,
		"blacklisted_proxies": blacklistedCount,
		"total_success":       totalSuccess,
		"total_failures":      totalFail,
		"strategy":            pp.strategy,
	}

	if totalSuccess+totalFail > 0 {
		stats["success_rate"] = float64(totalSuccess) / float64(totalSuccess+totalFail)
	} else {
		stats["success_rate"] = 0.0
	}

	return stats
}

// ValidateAll validates all proxies in the pool using a ProxyValidator.
// It returns a map of proxy URLs to validation errors.
// Proxies that fail validation are automatically blacklisted.
//
// Example:
//
//	validator := NewProxyValidator("https://www.google.com", 10*time.Second)
//	results := pool.ValidateAll(context.Background(), validator)
//	for proxyURL, err := range results {
//	    if err != nil {
//	        log.Printf("Proxy %s failed validation: %v", proxyURL, err)
//	    }
//	}
func (pp *ProxyPool) ValidateAll(ctx context.Context, validator *ProxyValidator) map[string]error {
	pp.mu.Lock()
	proxies := make([]*Proxy, len(pp.proxies))
	copy(proxies, pp.proxies)
	pp.mu.Unlock()

	// Validate all proxies concurrently
	results := validator.ValidateAll(ctx, proxies)

	// Blacklist failed proxies
	pp.mu.Lock()
	defer pp.mu.Unlock()

	for proxyURL, err := range results {
		if err != nil && proxyURL != "timeout" {
			// Find and blacklist the proxy
			for _, proxy := range pp.proxies {
				if proxy.URL == proxyURL {
					pp.blacklist[proxy.URL] = true
					proxy.IsBlacklisted = true
					break
				}
			}
		}
	}

	return results
}

// ResetBlacklist removes all proxies from the blacklist and resets their fail counts.
func (pp *ProxyPool) ResetBlacklist() {
	pp.mu.Lock()
	defer pp.mu.Unlock()

	pp.blacklist = make(map[string]bool)
	for _, proxy := range pp.proxies {
		proxy.IsBlacklisted = false
		proxy.FailCount = 0
	}
}

// AddProxy adds a new proxy to the pool.
// It parses the proxy URL and returns an error if the URL is invalid.
func (pp *ProxyPool) AddProxy(proxyURL string) error {
	proxy, err := ParseProxy(proxyURL)
	if err != nil {
		return fmt.Errorf("failed to add proxy: %w", err)
	}

	pp.mu.Lock()
	defer pp.mu.Unlock()

	pp.proxies = append(pp.proxies, proxy)
	return nil
}

// RemoveProxy removes a proxy from the pool by URL.
// It returns true if the proxy was found and removed, false otherwise.
func (pp *ProxyPool) RemoveProxy(proxyURL string) bool {
	pp.mu.Lock()
	defer pp.mu.Unlock()

	for i, proxy := range pp.proxies {
		if proxy.URL == proxyURL {
			// Remove proxy from slice
			pp.proxies = append(pp.proxies[:i], pp.proxies[i+1:]...)
			// Remove from blacklist if present
			delete(pp.blacklist, proxyURL)
			// Adjust current index if needed
			if pp.current >= len(pp.proxies) && len(pp.proxies) > 0 {
				pp.current = 0
			}
			return true
		}
	}

	return false
}
