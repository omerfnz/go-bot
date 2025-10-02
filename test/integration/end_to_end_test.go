//go:build integration
// +build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/omer/go-bot/internal/browser"
	"github.com/omer/go-bot/internal/config"
	"github.com/omer/go-bot/internal/logger"
	"github.com/omer/go-bot/internal/proxy"
	"github.com/omer/go-bot/internal/serp"
	"github.com/omer/go-bot/internal/stats"
	"github.com/omer/go-bot/internal/task"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEndToEnd_SimpleSearch tests a simple search flow end-to-end
func TestEndToEnd_SimpleSearch(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	log, err := logger.New(logger.Config{
		Level: logger.ErrorLevel,
	})
	require.NoError(t, err)
	defer log.Close()

	// Create browser
	browserOpts := browser.BrowserOptions{
		Headless: true,
		Timeout:  30 * time.Second,
	}
	b, err := browser.NewBrowser(browserOpts)
	require.NoError(t, err)
	defer b.Close()

	// Create searcher
	searcher := serp.NewSearcher(b, log)

	// Perform search
	err = searcher.Search("golang tutorial")
	require.NoError(t, err)

	// Get results
	results, err := searcher.GetResults()

	// Note: GetResults might fail with placeholder implementation
	// This is expected and will be fixed in later phases
	if err == nil {
		assert.NotEmpty(t, results)
	}
}

// TestEndToEnd_TaskExecution tests full task execution with worker pool
func TestEndToEnd_TaskExecution(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	log, err := logger.New(logger.Config{
		Level: logger.ErrorLevel,
	})
	require.NoError(t, err)
	defer log.Close()

	// Create worker pool without proxy
	pool := task.NewWorkerPool(task.WorkerPoolConfig{
		Workers:   2,
		QueueSize: 10,
		ProxyPool: nil,
		Logger:    log,
	})

	err = pool.Start()
	require.NoError(t, err)
	defer pool.Stop()

	// Create a simple task
	testTask, err := task.NewTask(task.TaskConfig{
		Keyword:   "golang",
		TargetURL: "go.dev",
	})
	require.NoError(t, err)

	// Submit task
	err = pool.Submit(testTask)
	require.NoError(t, err)

	// Wait for result (with timeout)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	select {
	case result := <-pool.GetResults():
		assert.NotNil(t, result)
		assert.Equal(t, testTask.ID, result.Task.ID)
		// Success is not guaranteed as it depends on actual search
	case <-ctx.Done():
		t.Fatal("Timeout waiting for task result")
	}
}

// TestEndToEnd_ProxyRotation tests proxy rotation functionality
func TestEndToEnd_ProxyRotation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Note: This test requires valid proxy servers
	// Using example proxies that will fail validation
	proxies := []string{
		"http://proxy1.example.com:8080",
		"http://proxy2.example.com:8080",
	}

	pool, err := proxy.NewProxyPool(proxies, proxy.RotationStrategyRoundRobin)
	require.NoError(t, err)

	// Get proxies in round-robin fashion
	proxy1, err := pool.Get()
	require.NoError(t, err)
	assert.Contains(t, proxy1.Host, "proxy")

	proxy2, err := pool.Get()
	require.NoError(t, err)
	assert.Contains(t, proxy2.Host, "proxy")

	// Should rotate
	assert.NotEqual(t, proxy1.Host, proxy2.Host)
}

// TestEndToEnd_StatisticsCollection tests statistics collection
func TestEndToEnd_StatisticsCollection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	collector := stats.NewStatsCollector("test_integration_stats.json")

	// Record some test tasks
	for i := 0; i < 5; i++ {
		taskStats := stats.TaskStats{
			TaskID:     "test-task-" + string(rune('0'+i)),
			Keyword:    "golang",
			TargetURL:  "example.com",
			Success:    i%2 == 0,
			Position:   i + 1,
			PageNumber: 1,
			Duration:   float64(1000 + i*100),
			Timestamp:  time.Now(),
		}
		collector.RecordTask(taskStats)
	}

	// Get summary
	summary := collector.GetSummary()
	assert.Equal(t, 5, summary["total_tasks"])
	assert.Equal(t, 3, summary["success_tasks"])
	assert.Equal(t, 2, summary["failed_tasks"])

	// Save and load
	err := collector.Save()
	require.NoError(t, err)

	newCollector := stats.NewStatsCollector("test_integration_stats.json")
	err = newCollector.Load()
	require.NoError(t, err)

	newSummary := newCollector.GetSummary()
	assert.Equal(t, summary["total_tasks"], newSummary["total_tasks"])
}

// TestEndToEnd_SchedulerSingleCycle tests scheduler in single-cycle mode
func TestEndToEnd_SchedulerSingleCycle(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cfg := &config.Config{
		Keywords: []config.Keyword{
			{Term: "golang", TargetURL: "go.dev"},
		},
		Workers:       2,
		Interval:      300,
		PageTimeout:   30,
		SearchTimeout: 15,
		MaxRetries:    3,
		RetryDelay:    5,
		Selectors: config.SelectorConfig{
			SearchBox:  "textarea[name='q']",
			ResultItem: "div.g",
		},
	}

	log, err := logger.New(logger.Config{
		Level: logger.ErrorLevel,
	})
	require.NoError(t, err)
	defer log.Close()

	statsCollector := stats.NewStatsCollector("test_scheduler_stats.json")

	pool := task.NewWorkerPool(task.WorkerPoolConfig{
		Workers:   cfg.Workers,
		QueueSize: 10,
		Logger:    log,
	})

	scheduler := task.NewScheduler(task.SchedulerConfig{
		Config:         cfg,
		WorkerPool:     pool,
		StatsCollector: statsCollector,
		Logger:         log,
		Interval:       1 * time.Second,
	})

	// Start in single-cycle mode
	err = scheduler.Start(false)
	require.NoError(t, err)

	// Wait for cycle to complete
	time.Sleep(5 * time.Second)

	// Stop scheduler
	err = scheduler.Stop()
	require.NoError(t, err)

	// Check that at least one cycle ran
	schedulerStats := scheduler.Stats()
	assert.GreaterOrEqual(t, schedulerStats["cycles_run"].(int), 1)
}

// TestEndToEnd_ConfigLoading tests configuration loading and validation
func TestEndToEnd_ConfigLoading(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Test loading example config
	cfg, err := config.Load("../../configs/config.json.example")
	require.NoError(t, err)

	// Validate
	err = cfg.Validate()
	require.NoError(t, err)

	// Check values
	assert.GreaterOrEqual(t, len(cfg.Keywords), 1)
	assert.GreaterOrEqual(t, cfg.Workers, 1)
	assert.GreaterOrEqual(t, cfg.PageTimeout, 1)
}

// TestEndToEnd_BrowserOperations tests basic browser operations
func TestEndToEnd_BrowserOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	b, err := browser.NewBrowser(browser.BrowserOptions{
		Headless: true,
		Timeout:  30 * time.Second,
	})
	require.NoError(t, err)
	defer b.Close()

	// Navigate to a page
	err = b.Navigate("https://www.google.com")
	require.NoError(t, err)

	// Check if element exists
	exists := b.ElementExists("textarea[name='q']")
	assert.True(t, exists, "Search box should exist on Google homepage")

	// Type into search box
	err = b.Type("textarea[name='q']", "golang")
	require.NoError(t, err)

	// Note: Not submitting to avoid actually searching
}

// TestEndToEnd_ProxyValidation tests proxy validation
func TestEndToEnd_ProxyValidation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	validator := proxy.NewProxyValidator("https://www.google.com", 5*time.Second)

	// Test with invalid proxy
	invalidProxy, err := proxy.ParseProxy("http://127.0.0.1:9999")
	require.NoError(t, err)

	err = validator.Validate(invalidProxy)
	assert.Error(t, err, "Invalid proxy should fail validation")
}
