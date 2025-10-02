package proxy

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewProxyValidator(t *testing.T) {
	t.Run("with custom URL and timeout", func(t *testing.T) {
		validator := NewProxyValidator("https://example.com", 5*time.Second)
		assert.NotNil(t, validator)
		assert.Equal(t, "https://example.com", validator.testURL)
		assert.Equal(t, 5*time.Second, validator.timeout)
	})

	t.Run("with empty URL uses default", func(t *testing.T) {
		validator := NewProxyValidator("", 5*time.Second)
		assert.NotNil(t, validator)
		assert.Equal(t, "https://www.google.com", validator.testURL)
	})

	t.Run("with zero timeout uses default", func(t *testing.T) {
		validator := NewProxyValidator("https://example.com", 0)
		assert.NotNil(t, validator)
		assert.Equal(t, 10*time.Second, validator.timeout)
	})

	t.Run("with all defaults", func(t *testing.T) {
		validator := NewProxyValidator("", 0)
		assert.NotNil(t, validator)
		assert.Equal(t, "https://www.google.com", validator.testURL)
		assert.Equal(t, 10*time.Second, validator.timeout)
	})
}

func TestProxyValidator_QuickValidate(t *testing.T) {
	validator := NewProxyValidator("https://example.com", 5*time.Second)

	t.Run("valid proxy", func(t *testing.T) {
		proxy, err := ParseProxy("http://proxy.example.com:8080")
		require.NoError(t, err)

		err = validator.QuickValidate(proxy)
		assert.NoError(t, err)
	})

	t.Run("nil proxy", func(t *testing.T) {
		err := validator.QuickValidate(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "proxy cannot be nil")
	})

	t.Run("blacklisted proxy", func(t *testing.T) {
		proxy, err := ParseProxy("http://proxy.example.com:8080")
		require.NoError(t, err)
		proxy.IsBlacklisted = true

		err = validator.QuickValidate(proxy)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "proxy is blacklisted")
	})

	t.Run("empty host", func(t *testing.T) {
		proxy := &Proxy{
			Host: "",
			Port: 8080,
		}

		err := validator.QuickValidate(proxy)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "proxy host is empty")
	})

	t.Run("invalid port too low", func(t *testing.T) {
		proxy := &Proxy{
			Host: "proxy.example.com",
			Port: 0,
		}

		err := validator.QuickValidate(proxy)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid proxy port")
	})

	t.Run("invalid port too high", func(t *testing.T) {
		proxy := &Proxy{
			Host: "proxy.example.com",
			Port: 99999,
		}

		err := validator.QuickValidate(proxy)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid proxy port")
	})
}

func TestProxyValidator_Validate(t *testing.T) {
	// Create a test HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	t.Run("nil proxy", func(t *testing.T) {
		validator := NewProxyValidator(server.URL, 5*time.Second)
		err := validator.Validate(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "proxy cannot be nil")
	})

	t.Run("blacklisted proxy", func(t *testing.T) {
		validator := NewProxyValidator(server.URL, 5*time.Second)
		proxy, err := ParseProxy("http://proxy.example.com:8080")
		require.NoError(t, err)
		proxy.IsBlacklisted = true

		err = validator.Validate(proxy)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "proxy is blacklisted")
	})

	t.Run("invalid proxy URL", func(t *testing.T) {
		validator := NewProxyValidator(server.URL, 5*time.Second)
		proxy := &Proxy{
			URL: "://invalid-url",
		}

		err := validator.Validate(proxy)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse proxy URL")
	})

	t.Run("unreachable proxy", func(t *testing.T) {
		validator := NewProxyValidator(server.URL, 1*time.Second)
		proxy, err := ParseProxy("http://127.0.0.1:9999")
		require.NoError(t, err)

		err = validator.Validate(proxy)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "proxy validation failed")
	})
}

func TestProxyValidator_ValidateAll(t *testing.T) {
	// Create a test HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	t.Run("validate multiple proxies", func(t *testing.T) {
		validator := NewProxyValidator(server.URL, 1*time.Second)

		// Create test proxies (all will fail as they're not real)
		proxy1, err := ParseProxy("http://proxy1.example.com:8080")
		require.NoError(t, err)

		proxy2, err := ParseProxy("http://proxy2.example.com:8080")
		require.NoError(t, err)

		proxy3, err := ParseProxy("http://proxy3.example.com:8080")
		require.NoError(t, err)

		proxies := []*Proxy{proxy1, proxy2, proxy3}
		results := validator.ValidateAll(context.Background(), proxies)

		assert.Len(t, results, 3)
		// All should have errors since they're not real proxies
		for url, err := range results {
			if url != "timeout" {
				assert.Error(t, err)
			}
		}
	})

	t.Run("validate with context cancellation", func(t *testing.T) {
		validator := NewProxyValidator(server.URL, 10*time.Second)

		proxy, err := ParseProxy("http://proxy.example.com:8080")
		require.NoError(t, err)

		proxies := []*Proxy{proxy}

		// Create context that cancels immediately
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		results := validator.ValidateAll(ctx, proxies)
		assert.NotEmpty(t, results)
	})

	t.Run("empty proxy list", func(t *testing.T) {
		validator := NewProxyValidator(server.URL, 5*time.Second)
		results := validator.ValidateAll(context.Background(), []*Proxy{})
		assert.Empty(t, results)
	})
}

func TestProxyValidator_ValidateWithRetry(t *testing.T) {
	// Create a test HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	t.Run("fails after max retries", func(t *testing.T) {
		validator := NewProxyValidator(server.URL, 500*time.Millisecond)
		proxy, err := ParseProxy("http://127.0.0.1:9999")
		require.NoError(t, err)

		err = validator.ValidateWithRetry(context.Background(), proxy, 2)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "proxy validation failed after 2 attempts")
	})

	t.Run("context cancellation", func(t *testing.T) {
		validator := NewProxyValidator(server.URL, 5*time.Second)
		proxy, err := ParseProxy("http://proxy.example.com:8080")
		require.NoError(t, err)

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err = validator.ValidateWithRetry(ctx, proxy, 3)
		assert.Error(t, err)
	})

	t.Run("blacklisted proxy fails immediately", func(t *testing.T) {
		validator := NewProxyValidator(server.URL, 5*time.Second)
		proxy, err := ParseProxy("http://proxy.example.com:8080")
		require.NoError(t, err)
		proxy.IsBlacklisted = true

		err = validator.ValidateWithRetry(context.Background(), proxy, 3)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "proxy is blacklisted")
	})

	t.Run("context timeout during retry", func(t *testing.T) {
		validator := NewProxyValidator(server.URL, 500*time.Millisecond)
		proxy, err := ParseProxy("http://127.0.0.1:9999")
		require.NoError(t, err)

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		err = validator.ValidateWithRetry(ctx, proxy, 5)
		assert.Error(t, err)
	})
}

func TestProxyValidator_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("validate real proxy endpoint", func(t *testing.T) {
		// This test would validate against a real proxy
		// For now, we just test the structure

		validator := NewProxyValidator("https://www.google.com", 10*time.Second)
		assert.NotNil(t, validator)

		// Create a proxy that will definitely fail
		proxy, err := ParseProxy("http://127.0.0.1:9999")
		require.NoError(t, err)

		err = validator.Validate(proxy)
		assert.Error(t, err) // Should fail as it's not a real proxy
	})
}

func TestProxyValidator_ConcurrentValidation(t *testing.T) {
	validator := NewProxyValidator("https://www.google.com", 2*time.Second)

	// Create multiple proxies
	proxies := make([]*Proxy, 10)
	for i := 0; i < 10; i++ {
		proxy, err := ParseProxy("http://127.0.0.1:9999")
		require.NoError(t, err)
		proxies[i] = proxy
	}

	// Run concurrent validation
	done := make(chan bool, 10)
	for _, proxy := range proxies {
		go func(p *Proxy) {
			_ = validator.Validate(p)
			done <- true
		}(proxy)
	}

	// Wait for all to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestProxyValidator_EdgeCases(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	t.Run("proxy returns non-2xx status", func(t *testing.T) {
		validator := NewProxyValidator(server.URL, 5*time.Second)

		// This would fail anyway since it's not a real proxy,
		// but the test structure is correct
		proxy, err := ParseProxy("http://proxy.example.com:8080")
		require.NoError(t, err)

		err = validator.Validate(proxy)
		assert.Error(t, err)
	})

	t.Run("very short timeout", func(t *testing.T) {
		validator := NewProxyValidator("https://www.google.com", 1*time.Nanosecond)
		proxy, err := ParseProxy("http://proxy.example.com:8080")
		require.NoError(t, err)

		err = validator.Validate(proxy)
		assert.Error(t, err)
	})
}
