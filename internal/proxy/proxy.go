// Package proxy provides proxy management and rotation functionality.
// It supports HTTP, HTTPS, and SOCKS5 proxies with connection pooling.
package proxy

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// ProxyType represents the type of proxy
type ProxyType string

const (
	// ProxyTypeHTTP represents an HTTP proxy
	ProxyTypeHTTP ProxyType = "http"
	// ProxyTypeHTTPS represents an HTTPS proxy
	ProxyTypeHTTPS ProxyType = "https"
	// ProxyTypeSOCKS5 represents a SOCKS5 proxy
	ProxyTypeSOCKS5 ProxyType = "socks5"
)

// Proxy represents a proxy server configuration
type Proxy struct {
	URL           string    // Full proxy URL (e.g., "http://host:port")
	Host          string    // Proxy host
	Port          int       // Proxy port
	Username      string    // Optional username for authentication (v1.3)
	Password      string    // Optional password for authentication (v1.3)
	Type          ProxyType // Proxy type (http, https, socks5)
	LastUsed      time.Time // Last time this proxy was used
	FailCount     int       // Number of consecutive failures
	SuccessCount  int       // Number of successful requests
	IsBlacklisted bool      // Whether this proxy is blacklisted
}

// ParseProxy parses a proxy URL string and returns a Proxy instance.
// Supported formats:
//   - http://host:port
//   - https://host:port
//   - socks5://host:port
//   - http://username:password@host:port (v1.3)
//
// Example:
//
//	proxy, err := ParseProxy("http://proxy.example.com:8080")
//	if err != nil {
//	    log.Fatal(err)
//	}
func ParseProxy(proxyURL string) (*Proxy, error) {
	if proxyURL == "" {
		return nil, fmt.Errorf("proxy URL cannot be empty")
	}

	// Parse URL
	u, err := url.Parse(proxyURL)
	if err != nil {
		return nil, fmt.Errorf("invalid proxy URL: %w", err)
	}

	// Validate scheme
	var proxyType ProxyType
	switch strings.ToLower(u.Scheme) {
	case "http":
		proxyType = ProxyTypeHTTP
	case "https":
		proxyType = ProxyTypeHTTPS
	case "socks5":
		proxyType = ProxyTypeSOCKS5
	default:
		return nil, fmt.Errorf("unsupported proxy scheme: %s (supported: http, https, socks5)", u.Scheme)
	}

	// Extract host
	host := u.Hostname()
	if host == "" {
		return nil, fmt.Errorf("proxy host cannot be empty")
	}

	// Extract port
	portStr := u.Port()
	var port int
	if portStr != "" {
		port, err = strconv.Atoi(portStr)
		if err != nil {
			return nil, fmt.Errorf("invalid proxy port: %w", err)
		}
		if port < 1 || port > 65535 {
			return nil, fmt.Errorf("proxy port must be between 1 and 65535, got %d", port)
		}
	} else {
		// Default ports
		switch proxyType {
		case ProxyTypeHTTP, ProxyTypeHTTPS:
			port = 8080
		case ProxyTypeSOCKS5:
			port = 1080
		}
	}

	// Extract authentication (v1.3)
	var username, password string
	if u.User != nil {
		username = u.User.Username()
		password, _ = u.User.Password()
	}

	proxy := &Proxy{
		URL:           proxyURL,
		Host:          host,
		Port:          port,
		Username:      username,
		Password:      password,
		Type:          proxyType,
		LastUsed:      time.Time{},
		FailCount:     0,
		SuccessCount:  0,
		IsBlacklisted: false,
	}

	return proxy, nil
}

// String returns a string representation of the proxy
func (p *Proxy) String() string {
	if p.Username != "" {
		return fmt.Sprintf("%s://%s:***@%s:%d", p.Type, p.Username, p.Host, p.Port)
	}
	return fmt.Sprintf("%s://%s:%d", p.Type, p.Host, p.Port)
}

// IsHealthy returns true if the proxy is considered healthy
// A proxy is healthy if it's not blacklisted and has fewer than 5 consecutive failures
func (p *Proxy) IsHealthy() bool {
	return !p.IsBlacklisted && p.FailCount < 5
}

// RecordSuccess records a successful request through this proxy
func (p *Proxy) RecordSuccess() {
	p.SuccessCount++
	p.FailCount = 0 // Reset fail count on success
	p.LastUsed = time.Now()
}

// RecordFailure records a failed request through this proxy
func (p *Proxy) RecordFailure() {
	p.FailCount++
	p.LastUsed = time.Now()
}

// GetSuccessRate returns the success rate of this proxy (0.0 to 1.0)
func (p *Proxy) GetSuccessRate() float64 {
	total := p.SuccessCount + p.FailCount
	if total == 0 {
		return 0.0
	}
	return float64(p.SuccessCount) / float64(total)
}
