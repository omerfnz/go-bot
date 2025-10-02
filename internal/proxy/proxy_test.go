package proxy

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ===== proxy.go tests =====

func TestParseProxy_HTTP(t *testing.T) {
	proxy, err := ParseProxy("http://proxy.example.com:8080")
	require.NoError(t, err)
	require.NotNil(t, proxy)

	assert.Equal(t, "http://proxy.example.com:8080", proxy.URL)
	assert.Equal(t, "proxy.example.com", proxy.Host)
	assert.Equal(t, 8080, proxy.Port)
	assert.Equal(t, ProxyTypeHTTP, proxy.Type)
	assert.Equal(t, "", proxy.Username)
	assert.Equal(t, "", proxy.Password)
	assert.False(t, proxy.IsBlacklisted)
	assert.Equal(t, 0, proxy.FailCount)
	assert.Equal(t, 0, proxy.SuccessCount)
}

func TestParseProxy_HTTPS(t *testing.T) {
	proxy, err := ParseProxy("https://secure-proxy.com:443")
	require.NoError(t, err)

	assert.Equal(t, "secure-proxy.com", proxy.Host)
	assert.Equal(t, 443, proxy.Port)
	assert.Equal(t, ProxyTypeHTTPS, proxy.Type)
}

func TestParseProxy_SOCKS5(t *testing.T) {
	proxy, err := ParseProxy("socks5://socks-proxy.com:1080")
	require.NoError(t, err)

	assert.Equal(t, "socks-proxy.com", proxy.Host)
	assert.Equal(t, 1080, proxy.Port)
	assert.Equal(t, ProxyTypeSOCKS5, proxy.Type)
}

func TestParseProxy_WithAuth(t *testing.T) {
	proxy, err := ParseProxy("http://user:pass@proxy.com:8080")
	require.NoError(t, err)

	assert.Equal(t, "proxy.com", proxy.Host)
	assert.Equal(t, 8080, proxy.Port)
	assert.Equal(t, "user", proxy.Username)
	assert.Equal(t, "pass", proxy.Password)
}

func TestParseProxy_DefaultPort(t *testing.T) {
	tests := []struct {
		url          string
		expectedPort int
	}{
		{"http://proxy.com", 8080},
		{"https://proxy.com", 8080},
		{"socks5://proxy.com", 1080},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			proxy, err := ParseProxy(tt.url)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedPort, proxy.Port)
		})
	}
}

func TestParseProxy_EmptyURL(t *testing.T) {
	proxy, err := ParseProxy("")
	assert.Error(t, err)
	assert.Nil(t, proxy)
	assert.Contains(t, err.Error(), "proxy URL cannot be empty")
}

func TestParseProxy_InvalidURL(t *testing.T) {
	proxy, err := ParseProxy("not a valid url")
	assert.Error(t, err)
	assert.Nil(t, proxy)
}

func TestParseProxy_UnsupportedScheme(t *testing.T) {
	proxy, err := ParseProxy("ftp://proxy.com:8080")
	assert.Error(t, err)
	assert.Nil(t, proxy)
	assert.Contains(t, err.Error(), "unsupported proxy scheme")
}

func TestParseProxy_EmptyHost(t *testing.T) {
	proxy, err := ParseProxy("http://:8080")
	assert.Error(t, err)
	assert.Nil(t, proxy)
	assert.Contains(t, err.Error(), "proxy host cannot be empty")
}

func TestParseProxy_InvalidPort(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{"invalid_port", "http://proxy.com:invalid"},
		{"port_too_low", "http://proxy.com:0"},
		{"port_too_high", "http://proxy.com:99999"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proxy, err := ParseProxy(tt.url)
			assert.Error(t, err)
			assert.Nil(t, proxy)
		})
	}
}

func TestProxy_String(t *testing.T) {
	tests := []struct {
		name     string
		proxyURL string
		expected string
	}{
		{
			name:     "without_auth",
			proxyURL: "http://proxy.com:8080",
			expected: "http://proxy.com:8080",
		},
		{
			name:     "with_auth",
			proxyURL: "http://user:password@proxy.com:8080",
			expected: "http://user:***@proxy.com:8080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proxy, err := ParseProxy(tt.proxyURL)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, proxy.String())
		})
	}
}

func TestProxy_IsHealthy(t *testing.T) {
	proxy, _ := ParseProxy("http://proxy.com:8080")

	// Initially healthy
	assert.True(t, proxy.IsHealthy())

	// Still healthy with few failures
	proxy.FailCount = 4
	assert.True(t, proxy.IsHealthy())

	// Unhealthy with many failures
	proxy.FailCount = 5
	assert.False(t, proxy.IsHealthy())

	// Unhealthy if blacklisted
	proxy.FailCount = 0
	proxy.IsBlacklisted = true
	assert.False(t, proxy.IsHealthy())
}

func TestProxy_RecordSuccess(t *testing.T) {
	proxy, _ := ParseProxy("http://proxy.com:8080")
	proxy.FailCount = 3

	before := time.Now()
	proxy.RecordSuccess()
	after := time.Now()

	assert.Equal(t, 1, proxy.SuccessCount)
	assert.Equal(t, 0, proxy.FailCount) // Should reset
	assert.True(t, proxy.LastUsed.After(before) || proxy.LastUsed.Equal(before))
	assert.True(t, proxy.LastUsed.Before(after) || proxy.LastUsed.Equal(after))
}

func TestProxy_RecordFailure(t *testing.T) {
	proxy, _ := ParseProxy("http://proxy.com:8080")

	before := time.Now()
	proxy.RecordFailure()
	after := time.Now()

	assert.Equal(t, 1, proxy.FailCount)
	assert.True(t, proxy.LastUsed.After(before) || proxy.LastUsed.Equal(before))
	assert.True(t, proxy.LastUsed.Before(after) || proxy.LastUsed.Equal(after))
}

func TestProxy_GetSuccessRate(t *testing.T) {
	proxy, _ := ParseProxy("http://proxy.com:8080")

	// No requests yet
	assert.Equal(t, 0.0, proxy.GetSuccessRate())

	// 3 successes, 1 failure = 75%
	proxy.SuccessCount = 3
	proxy.FailCount = 1
	assert.Equal(t, 0.75, proxy.GetSuccessRate())

	// All successes
	proxy.SuccessCount = 10
	proxy.FailCount = 0
	assert.Equal(t, 1.0, proxy.GetSuccessRate())

	// All failures
	proxy.SuccessCount = 0
	proxy.FailCount = 10
	assert.Equal(t, 0.0, proxy.GetSuccessRate())
}

// ===== pool.go tests =====

func TestNewProxyPool_Success(t *testing.T) {
	urls := []string{
		"http://proxy1.com:8080",
		"http://proxy2.com:8080",
	}

	pool, err := NewProxyPool(urls, RotationStrategyRoundRobin)
	require.NoError(t, err)
	require.NotNil(t, pool)

	assert.Equal(t, 2, pool.Size())
	assert.Equal(t, RotationStrategyRoundRobin, pool.strategy)
}

func TestNewProxyPool_EmptyList(t *testing.T) {
	pool, err := NewProxyPool([]string{}, RotationStrategyRoundRobin)
	assert.Error(t, err)
	assert.Nil(t, pool)
	assert.Contains(t, err.Error(), "at least one proxy URL is required")
}

func TestNewProxyPool_InvalidStrategy(t *testing.T) {
	urls := []string{"http://proxy.com:8080"}
	pool, err := NewProxyPool(urls, RotationStrategy("invalid"))
	assert.Error(t, err)
	assert.Nil(t, pool)
	assert.Contains(t, err.Error(), "invalid rotation strategy")
}

func TestNewProxyPool_InvalidProxyURL(t *testing.T) {
	urls := []string{
		"http://proxy1.com:8080",
		"invalid url",
	}

	pool, err := NewProxyPool(urls, RotationStrategyRoundRobin)
	assert.Error(t, err)
	assert.Nil(t, pool)
	assert.Contains(t, err.Error(), "failed to parse proxy[1]")
}

func TestProxyPool_Get_RoundRobin(t *testing.T) {
	urls := []string{
		"http://proxy1.com:8080",
		"http://proxy2.com:8080",
		"http://proxy3.com:8080",
	}

	pool, err := NewProxyPool(urls, RotationStrategyRoundRobin)
	require.NoError(t, err)

	// Should rotate in order
	proxy1, err := pool.Get()
	require.NoError(t, err)
	assert.Equal(t, "proxy1.com", proxy1.Host)

	proxy2, err := pool.Get()
	require.NoError(t, err)
	assert.Equal(t, "proxy2.com", proxy2.Host)

	proxy3, err := pool.Get()
	require.NoError(t, err)
	assert.Equal(t, "proxy3.com", proxy3.Host)

	// Should wrap around
	proxy4, err := pool.Get()
	require.NoError(t, err)
	assert.Equal(t, "proxy1.com", proxy4.Host)
}

func TestProxyPool_Get_SkipsBlacklisted(t *testing.T) {
	urls := []string{
		"http://proxy1.com:8080",
		"http://proxy2.com:8080",
	}

	pool, err := NewProxyPool(urls, RotationStrategyRoundRobin)
	require.NoError(t, err)

	// Blacklist first proxy
	proxy1, _ := pool.Get()
	pool.Blacklist(proxy1)

	// Should skip blacklisted and return proxy2
	proxy, err := pool.Get()
	require.NoError(t, err)
	assert.Equal(t, "proxy2.com", proxy.Host)
}

func TestProxyPool_Get_NoHealthyProxies(t *testing.T) {
	urls := []string{
		"http://proxy1.com:8080",
	}

	pool, err := NewProxyPool(urls, RotationStrategyRoundRobin)
	require.NoError(t, err)

	// Blacklist the only proxy
	proxy, _ := pool.Get()
	pool.Blacklist(proxy)

	// Should return error
	proxy, err = pool.Get()
	assert.Error(t, err)
	assert.Nil(t, proxy)
	assert.Contains(t, err.Error(), "no healthy proxy available")
}

func TestProxyPool_Release_Success(t *testing.T) {
	urls := []string{"http://proxy.com:8080"}
	pool, _ := NewProxyPool(urls, RotationStrategyRoundRobin)

	proxy, _ := pool.Get()
	pool.Release(proxy, true)

	assert.Equal(t, 1, proxy.SuccessCount)
	assert.Equal(t, 0, proxy.FailCount)
}

func TestProxyPool_Release_Failure(t *testing.T) {
	urls := []string{"http://proxy.com:8080"}
	pool, _ := NewProxyPool(urls, RotationStrategyRoundRobin)

	proxy, _ := pool.Get()
	pool.Release(proxy, false)

	assert.Equal(t, 0, proxy.SuccessCount)
	assert.Equal(t, 1, proxy.FailCount)
}

func TestProxyPool_Release_AutoBlacklist(t *testing.T) {
	urls := []string{"http://proxy.com:8080"}
	pool, _ := NewProxyPool(urls, RotationStrategyRoundRobin)

	proxy, _ := pool.Get()

	// Record 5 failures
	for i := 0; i < 5; i++ {
		pool.Release(proxy, false)
	}

	assert.True(t, proxy.IsBlacklisted)
	assert.True(t, pool.IsBlacklisted(proxy.URL))
}

func TestProxyPool_Blacklist(t *testing.T) {
	urls := []string{"http://proxy.com:8080"}
	pool, _ := NewProxyPool(urls, RotationStrategyRoundRobin)

	proxy, _ := pool.Get()
	pool.Blacklist(proxy)

	assert.True(t, proxy.IsBlacklisted)
	assert.True(t, pool.IsBlacklisted(proxy.URL))
}

func TestProxyPool_RemoveFromBlacklist(t *testing.T) {
	urls := []string{"http://proxy.com:8080"}
	pool, _ := NewProxyPool(urls, RotationStrategyRoundRobin)

	proxy, _ := pool.Get()
	proxy.FailCount = 3
	pool.Blacklist(proxy)

	pool.RemoveFromBlacklist(proxy)

	assert.False(t, proxy.IsBlacklisted)
	assert.False(t, pool.IsBlacklisted(proxy.URL))
	assert.Equal(t, 0, proxy.FailCount)
}

func TestProxyPool_GetProxies(t *testing.T) {
	urls := []string{
		"http://proxy1.com:8080",
		"http://proxy2.com:8080",
	}

	pool, _ := NewProxyPool(urls, RotationStrategyRoundRobin)
	proxies := pool.GetProxies()

	assert.Len(t, proxies, 2)
	assert.Equal(t, "proxy1.com", proxies[0].Host)
	assert.Equal(t, "proxy2.com", proxies[1].Host)
}

func TestProxyPool_Size(t *testing.T) {
	urls := []string{
		"http://proxy1.com:8080",
		"http://proxy2.com:8080",
		"http://proxy3.com:8080",
	}

	pool, _ := NewProxyPool(urls, RotationStrategyRoundRobin)
	assert.Equal(t, 3, pool.Size())
}

func TestProxyPool_HealthyCount(t *testing.T) {
	urls := []string{
		"http://proxy1.com:8080",
		"http://proxy2.com:8080",
		"http://proxy3.com:8080",
	}

	pool, _ := NewProxyPool(urls, RotationStrategyRoundRobin)

	// All healthy initially
	assert.Equal(t, 3, pool.HealthyCount())

	// Blacklist one
	proxy, _ := pool.Get()
	pool.Blacklist(proxy)
	assert.Equal(t, 2, pool.HealthyCount())
}

func TestProxyPool_GetStats(t *testing.T) {
	urls := []string{
		"http://proxy1.com:8080",
		"http://proxy2.com:8080",
	}

	pool, _ := NewProxyPool(urls, RotationStrategyRoundRobin)

	// Record some successes and failures
	proxy1, _ := pool.Get()
	pool.Release(proxy1, true)
	pool.Release(proxy1, true)
	pool.Release(proxy1, false)

	proxy2, _ := pool.Get()
	pool.Release(proxy2, false)
	pool.Release(proxy2, false)

	stats := pool.GetStats()

	assert.Equal(t, 2, stats["total_proxies"])
	assert.Equal(t, 2, stats["healthy_proxies"])
	assert.Equal(t, 0, stats["blacklisted_proxies"])
	assert.Equal(t, 2, stats["total_success"])
	assert.Equal(t, 3, stats["total_failures"])
	assert.Equal(t, 0.4, stats["success_rate"])
	assert.Equal(t, RotationStrategyRoundRobin, stats["strategy"])
}

func TestProxyPool_GetStats_NoRequests(t *testing.T) {
	urls := []string{"http://proxy.com:8080"}
	pool, _ := NewProxyPool(urls, RotationStrategyRoundRobin)

	stats := pool.GetStats()
	assert.Equal(t, 0.0, stats["success_rate"])
}

func TestProxyPool_Concurrency(t *testing.T) {
	urls := []string{
		"http://proxy1.com:8080",
		"http://proxy2.com:8080",
		"http://proxy3.com:8080",
	}

	pool, _ := NewProxyPool(urls, RotationStrategyRoundRobin)

	var wg sync.WaitGroup
	iterations := 100

	// Concurrent Get operations
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			proxy, err := pool.Get()
			assert.NoError(t, err)
			assert.NotNil(t, proxy)
		}()
	}

	wg.Wait()
}

func TestProxyPool_ConcurrentRelease(t *testing.T) {
	urls := []string{"http://proxy.com:8080"}
	pool, _ := NewProxyPool(urls, RotationStrategyRoundRobin)

	var wg sync.WaitGroup
	iterations := 50

	// Concurrent Get and Release operations
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func(success bool) {
			defer wg.Done()
			proxy, err := pool.Get()
			if err == nil {
				pool.Release(proxy, success)
			}
		}(i%2 == 0)
	}

	wg.Wait()

	// Check that operations completed without panic
	proxies := pool.GetProxies()
	assert.Len(t, proxies, 1)
	// At least some operations should have been recorded
	assert.Greater(t, proxies[0].SuccessCount+proxies[0].FailCount, 0)
}

// ===== New Faz 2 tests =====

func TestProxyPool_RandomStrategy(t *testing.T) {
	urls := []string{
		"http://proxy1.com:8080",
		"http://proxy2.com:8080",
		"http://proxy3.com:8080",
		"http://proxy4.com:8080",
		"http://proxy5.com:8080",
	}

	pool, err := NewProxyPool(urls, RotationStrategyRandom)
	require.NoError(t, err)

	// Get multiple proxies and track which ones we get
	seen := make(map[string]int)
	for i := 0; i < 50; i++ {
		proxy, err := pool.Get()
		require.NoError(t, err)
		seen[proxy.Host]++
	}

	// With random strategy, we should see multiple different proxies
	// (probability of seeing only one proxy in 50 attempts is extremely low)
	assert.Greater(t, len(seen), 1, "Random strategy should select different proxies")
}

func TestProxyPool_ResetBlacklist(t *testing.T) {
	urls := []string{
		"http://proxy1.com:8080",
		"http://proxy2.com:8080",
	}

	pool, err := NewProxyPool(urls, RotationStrategyRoundRobin)
	require.NoError(t, err)

	// Blacklist all proxies
	proxy1, _ := pool.Get()
	pool.Blacklist(proxy1)
	proxy2, _ := pool.Get()
	pool.Blacklist(proxy2)

	assert.Equal(t, 0, pool.HealthyCount())

	// Reset blacklist
	pool.ResetBlacklist()

	// All should be healthy again
	assert.Equal(t, 2, pool.HealthyCount())
	assert.Equal(t, 0, proxy1.FailCount)
	assert.Equal(t, 0, proxy2.FailCount)
	assert.False(t, proxy1.IsBlacklisted)
	assert.False(t, proxy2.IsBlacklisted)
}

func TestProxyPool_AddProxy(t *testing.T) {
	urls := []string{"http://proxy1.com:8080"}
	pool, err := NewProxyPool(urls, RotationStrategyRoundRobin)
	require.NoError(t, err)

	assert.Equal(t, 1, pool.Size())

	// Add a new proxy
	err = pool.AddProxy("http://proxy2.com:8080")
	require.NoError(t, err)

	assert.Equal(t, 2, pool.Size())

	// Get proxies and verify both are present
	proxies := pool.GetProxies()
	hosts := []string{proxies[0].Host, proxies[1].Host}
	assert.Contains(t, hosts, "proxy1.com")
	assert.Contains(t, hosts, "proxy2.com")
}

func TestProxyPool_AddProxy_InvalidURL(t *testing.T) {
	urls := []string{"http://proxy1.com:8080"}
	pool, err := NewProxyPool(urls, RotationStrategyRoundRobin)
	require.NoError(t, err)

	err = pool.AddProxy("invalid-url")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to add proxy")

	// Pool size should remain unchanged
	assert.Equal(t, 1, pool.Size())
}

func TestProxyPool_RemoveProxy(t *testing.T) {
	urls := []string{
		"http://proxy1.com:8080",
		"http://proxy2.com:8080",
		"http://proxy3.com:8080",
	}
	pool, err := NewProxyPool(urls, RotationStrategyRoundRobin)
	require.NoError(t, err)

	assert.Equal(t, 3, pool.Size())

	// Remove a proxy
	removed := pool.RemoveProxy("http://proxy2.com:8080")
	assert.True(t, removed)
	assert.Equal(t, 2, pool.Size())

	// Verify proxy is gone
	proxies := pool.GetProxies()
	for _, proxy := range proxies {
		assert.NotEqual(t, "proxy2.com", proxy.Host)
	}
}

func TestProxyPool_RemoveProxy_NotFound(t *testing.T) {
	urls := []string{"http://proxy1.com:8080"}
	pool, err := NewProxyPool(urls, RotationStrategyRoundRobin)
	require.NoError(t, err)

	removed := pool.RemoveProxy("http://nonexistent.com:8080")
	assert.False(t, removed)
	assert.Equal(t, 1, pool.Size())
}

func TestProxyPool_RemoveProxy_AdjustsCurrentIndex(t *testing.T) {
	urls := []string{
		"http://proxy1.com:8080",
		"http://proxy2.com:8080",
	}
	pool, err := NewProxyPool(urls, RotationStrategyRoundRobin)
	require.NoError(t, err)

	// Get both proxies to advance current index
	pool.Get()
	pool.Get()

	// Current index should be 0 now (wrapped around)
	// Remove both proxies
	pool.RemoveProxy("http://proxy1.com:8080")
	pool.RemoveProxy("http://proxy2.com:8080")

	// Current index should be adjusted
	assert.Equal(t, 0, pool.current)
}

func TestProxyPool_RemoveProxy_RemovesFromBlacklist(t *testing.T) {
	urls := []string{
		"http://proxy1.com:8080",
		"http://proxy2.com:8080",
	}
	pool, err := NewProxyPool(urls, RotationStrategyRoundRobin)
	require.NoError(t, err)

	// Blacklist a proxy
	proxy, _ := pool.Get()
	pool.Blacklist(proxy)
	proxyURL := proxy.URL

	assert.True(t, pool.IsBlacklisted(proxyURL))

	// Remove the proxy
	removed := pool.RemoveProxy(proxyURL)
	assert.True(t, removed)

	// Should no longer be in blacklist
	assert.False(t, pool.IsBlacklisted(proxyURL))
}

func TestProxyPool_ValidateAll(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	urls := []string{
		"http://proxy1.example.com:8080",
		"http://proxy2.example.com:8080",
	}
	pool, err := NewProxyPool(urls, RotationStrategyRoundRobin)
	require.NoError(t, err)

	validator := NewProxyValidator("https://www.google.com", 2*time.Second)
	results := pool.ValidateAll(context.Background(), validator)

	// All should fail as they're not real proxies
	assert.Len(t, results, 2)
	for url, err := range results {
		if url != "timeout" {
			assert.Error(t, err)
		}
	}

	// Failed proxies should be blacklisted
	assert.Equal(t, 0, pool.HealthyCount())
}

func TestProxyPool_ValidateAll_WithTimeout(t *testing.T) {
	urls := []string{"http://proxy.example.com:8080"}
	pool, err := NewProxyPool(urls, RotationStrategyRoundRobin)
	require.NoError(t, err)

	validator := NewProxyValidator("https://www.google.com", 1*time.Second)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	results := pool.ValidateAll(ctx, validator)
	assert.NotEmpty(t, results)
}

func TestProxyPool_ConcurrentAddRemove(t *testing.T) {
	urls := []string{"http://proxy1.com:8080"}
	pool, err := NewProxyPool(urls, RotationStrategyRoundRobin)
	require.NoError(t, err)

	var wg sync.WaitGroup
	operations := 20

	// Concurrent add operations
	for i := 0; i < operations; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			proxyURL := "http://proxy" + string(rune('a'+idx)) + ".com:8080"
			_ = pool.AddProxy(proxyURL)
		}(i)
	}

	// Concurrent get operations
	for i := 0; i < operations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = pool.Get()
		}()
	}

	wg.Wait()

	// Pool should have grown (at least some adds should have succeeded)
	assert.Greater(t, pool.Size(), 1)
}
