package task

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/omer/go-bot/internal/config"
	"github.com/omer/go-bot/internal/logger"
	"github.com/omer/go-bot/internal/stats"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper to create a test config
func createTestConfig() *config.Config {
	return &config.Config{
		Keywords: []config.Keyword{
			{Term: "golang", TargetURL: "example.com"},
			{Term: "go programming", TargetURL: "example.org"},
		},
		Workers:       2,
		Interval:      300,
		PageTimeout:   30,
		SearchTimeout: 15,
		MaxRetries:    3,
		RetryDelay:    5,
		Selectors: config.SelectorConfig{
			SearchBox:  "input[name='q']",
			ResultItem: "div.result",
		},
	}
}

func TestNewScheduler(t *testing.T) {
	cfg := createTestConfig()
	log, _ := logger.New(logger.Config{Level: logger.InfoLevel})

	pool := NewWorkerPool(WorkerPoolConfig{
		Workers:   2,
		QueueSize: 10,
		Logger:    log,
	})

	scheduler := NewScheduler(SchedulerConfig{
		Config:     cfg,
		WorkerPool: pool,
		Logger:     log,
		Interval:   5 * time.Minute,
	})

	assert.NotNil(t, scheduler)
	assert.Equal(t, 5*time.Minute, scheduler.interval)
	assert.False(t, scheduler.IsRunning())
	assert.Equal(t, 0, scheduler.cyclesRun)
}

func TestNewScheduler_DefaultInterval(t *testing.T) {
	cfg := createTestConfig()
	log, _ := logger.New(logger.Config{Level: logger.InfoLevel})

	pool := NewWorkerPool(WorkerPoolConfig{
		Workers: 2,
		Logger:  log,
	})

	scheduler := NewScheduler(SchedulerConfig{
		Config:     cfg,
		WorkerPool: pool,
		Logger:     log,
		Interval:   0, // Should use default
	})

	assert.Equal(t, 5*time.Minute, scheduler.interval)
}

func TestScheduler_StartStop(t *testing.T) {
	cfg := createTestConfig()
	log, _ := logger.New(logger.Config{Level: logger.ErrorLevel})

	// Mock executor that completes immediately
	mockExecutor := func(task *Task) *TaskResult {
		task.MarkRunning()
		time.Sleep(10 * time.Millisecond)
		task.MarkCompleted()
		return NewTaskResult(task, true, nil)
	}

	pool := NewWorkerPool(WorkerPoolConfig{
		Workers:   2,
		QueueSize: 10,
		Logger:    log,
		Executor:  mockExecutor,
	})

	scheduler := NewScheduler(SchedulerConfig{
		Config:     cfg,
		WorkerPool: pool,
		Logger:     log,
		Interval:   100 * time.Millisecond,
	})

	// Start scheduler in single-cycle mode
	err := scheduler.Start(false)
	require.NoError(t, err)
	assert.True(t, scheduler.IsRunning())

	// Wait for cycle to complete
	time.Sleep(200 * time.Millisecond)

	// Stop scheduler
	err = scheduler.Stop()
	require.NoError(t, err)
	assert.False(t, scheduler.IsRunning())
}

func TestScheduler_StartAlreadyRunning(t *testing.T) {
	cfg := createTestConfig()
	log, _ := logger.New(logger.Config{Level: logger.ErrorLevel})

	pool := NewWorkerPool(WorkerPoolConfig{
		Workers: 2,
		Logger:  log,
		Executor: func(task *Task) *TaskResult {
			time.Sleep(100 * time.Millisecond)
			task.MarkCompleted()
			return NewTaskResult(task, true, nil)
		},
	})

	scheduler := NewScheduler(SchedulerConfig{
		Config:     cfg,
		WorkerPool: pool,
		Logger:     log,
		Interval:   1 * time.Second,
	})

	err := scheduler.Start(false)
	require.NoError(t, err)

	// Try to start again
	err = scheduler.Start(false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already running")

	scheduler.Stop()
}

func TestScheduler_StopNotRunning(t *testing.T) {
	cfg := createTestConfig()
	log, _ := logger.New(logger.Config{Level: logger.ErrorLevel})

	pool := NewWorkerPool(WorkerPoolConfig{
		Workers: 2,
		Logger:  log,
	})

	scheduler := NewScheduler(SchedulerConfig{
		Config:     cfg,
		WorkerPool: pool,
		Logger:     log,
		Interval:   1 * time.Second,
	})

	err := scheduler.Stop()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not running")
}

func TestScheduler_SingleCycle(t *testing.T) {
	cfg := createTestConfig()
	log, _ := logger.New(logger.Config{Level: logger.ErrorLevel})

	tasksExecuted := 0
	mockExecutor := func(task *Task) *TaskResult {
		tasksExecuted++
		task.MarkRunning()
		task.MarkCompleted()
		return NewTaskResult(task, true, nil)
	}

	pool := NewWorkerPool(WorkerPoolConfig{
		Workers:   2,
		QueueSize: 10,
		Logger:    log,
		Executor:  mockExecutor,
	})

	scheduler := NewScheduler(SchedulerConfig{
		Config:     cfg,
		WorkerPool: pool,
		Logger:     log,
		Interval:   1 * time.Second,
	})

	err := scheduler.Start(false) // Single cycle
	require.NoError(t, err)

	// Wait for cycle to complete
	time.Sleep(200 * time.Millisecond)

	stats := scheduler.Stats()
	assert.Equal(t, 1, stats["cycles_run"])
	assert.Equal(t, 2, tasksExecuted) // 2 keywords in config
}

func TestScheduler_ContinuousMode(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping continuous mode test in short mode")
	}

	cfg := createTestConfig()
	log, _ := logger.New(logger.Config{Level: logger.ErrorLevel})

	mockExecutor := func(task *Task) *TaskResult {
		task.MarkRunning()
		task.MarkCompleted()
		return NewTaskResult(task, true, nil)
	}

	pool := NewWorkerPool(WorkerPoolConfig{
		Workers:   2,
		QueueSize: 10,
		Logger:    log,
		Executor:  mockExecutor,
	})

	scheduler := NewScheduler(SchedulerConfig{
		Config:     cfg,
		WorkerPool: pool,
		Logger:     log,
		Interval:   100 * time.Millisecond,
	})

	err := scheduler.Start(true) // Continuous mode
	require.NoError(t, err)

	// Wait for multiple cycles
	time.Sleep(350 * time.Millisecond)

	stats := scheduler.Stats()
	cyclesRun := stats["cycles_run"].(int)

	// Should have run at least 2 cycles
	assert.GreaterOrEqual(t, cyclesRun, 2)

	// Stop scheduler
	err = scheduler.Stop()
	require.NoError(t, err)
}

func TestScheduler_WithStatsCollector(t *testing.T) {
	cfg := createTestConfig()
	log, _ := logger.New(logger.Config{Level: logger.ErrorLevel})

	statsCollector := stats.NewStatsCollector("test_stats.json")

	mockExecutor := func(task *Task) *TaskResult {
		task.MarkRunning()
		time.Sleep(10 * time.Millisecond)
		task.MarkCompleted()
		result := NewTaskResult(task, true, nil)
		result.Position = 3
		result.PageNumber = 1
		return result
	}

	pool := NewWorkerPool(WorkerPoolConfig{
		Workers:   2,
		QueueSize: 10,
		Logger:    log,
		Executor:  mockExecutor,
	})

	scheduler := NewScheduler(SchedulerConfig{
		Config:         cfg,
		WorkerPool:     pool,
		StatsCollector: statsCollector,
		Logger:         log,
		Interval:       1 * time.Second,
	})

	err := scheduler.Start(false)
	require.NoError(t, err)

	// Wait for cycle to complete
	time.Sleep(200 * time.Millisecond)

	scheduler.Stop()

	// Check stats were recorded
	summary := statsCollector.GetSummary()
	assert.Greater(t, summary["total_tasks"].(int), 0)
}

func TestScheduler_TaskFailure(t *testing.T) {
	cfg := createTestConfig()
	log, _ := logger.New(logger.Config{Level: logger.ErrorLevel})

	mockExecutor := func(task *Task) *TaskResult {
		task.MarkRunning()
		task.MarkFailed()
		return NewTaskResult(task, false, errors.New("simulated failure"))
	}

	pool := NewWorkerPool(WorkerPoolConfig{
		Workers:   2,
		QueueSize: 10,
		Logger:    log,
		Executor:  mockExecutor,
	})

	scheduler := NewScheduler(SchedulerConfig{
		Config:     cfg,
		WorkerPool: pool,
		Logger:     log,
		Interval:   1 * time.Second,
	})

	err := scheduler.Start(false)
	require.NoError(t, err)

	// Wait for cycle to complete
	time.Sleep(200 * time.Millisecond)

	// Scheduler should complete even with failures
	stats := scheduler.Stats()
	assert.Equal(t, 1, stats["cycles_run"])

	scheduler.Stop()
}

func TestScheduler_ContextCancellation(t *testing.T) {
	cfg := createTestConfig()
	log, _ := logger.New(logger.Config{Level: logger.ErrorLevel})

	mockExecutor := func(task *Task) *TaskResult {
		time.Sleep(50 * time.Millisecond)
		task.MarkCompleted()
		return NewTaskResult(task, true, nil)
	}

	pool := NewWorkerPool(WorkerPoolConfig{
		Workers:   2,
		QueueSize: 10,
		Logger:    log,
		Executor:  mockExecutor,
	})

	scheduler := NewScheduler(SchedulerConfig{
		Config:     cfg,
		WorkerPool: pool,
		Logger:     log,
		Interval:   5 * time.Second,
	})

	err := scheduler.Start(true)
	require.NoError(t, err)

	// Stop properly
	time.Sleep(150 * time.Millisecond)
	err = scheduler.Stop()
	require.NoError(t, err)

	// Should not be running after stop
	assert.False(t, scheduler.IsRunning())
}

func TestRetryWithBackoff_Success(t *testing.T) {
	ctx := context.Background()
	attempts := 0

	fn := func() error {
		attempts++
		if attempts < 3 {
			return errors.New("temporary error")
		}
		return nil
	}

	err := RetryWithBackoff(ctx, 5, 10*time.Millisecond, fn)
	assert.NoError(t, err)
	assert.Equal(t, 3, attempts)
}

func TestRetryWithBackoff_MaxRetriesExceeded(t *testing.T) {
	ctx := context.Background()
	attempts := 0

	fn := func() error {
		attempts++
		return errors.New("persistent error")
	}

	err := RetryWithBackoff(ctx, 3, 5*time.Millisecond, fn)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "operation failed after 3 attempts")
	assert.Equal(t, 3, attempts)
}

func TestRetryWithBackoff_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	attempts := 0

	fn := func() error {
		attempts++
		return errors.New("error")
	}

	// Cancel context after short time
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	err := RetryWithBackoff(ctx, 10, 20*time.Millisecond, fn)
	assert.Error(t, err)
	// Should stop early due to context cancellation
	assert.Less(t, attempts, 10)
}

func TestRetryWithBackoff_ImmediateSuccess(t *testing.T) {
	ctx := context.Background()
	attempts := 0

	fn := func() error {
		attempts++
		return nil
	}

	err := RetryWithBackoff(ctx, 5, 10*time.Millisecond, fn)
	assert.NoError(t, err)
	assert.Equal(t, 1, attempts)
}

func TestRetryWithBackoff_BackoffCapping(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping backoff capping test in short mode (takes too long)")
	}

	ctx := context.Background()
	attempts := 0
	maxAttempts := 5

	fn := func() error {
		attempts++
		return errors.New("error")
	}

	start := time.Now()
	// Use smaller initial delay for testing
	err := RetryWithBackoff(ctx, maxAttempts, 100*time.Millisecond, fn)
	elapsed := time.Since(start)

	assert.Error(t, err)
	assert.Equal(t, maxAttempts, attempts)

	// With backoff: 100ms, 200ms, 400ms, 800ms = ~1.5s total
	// Allow some buffer for test execution
	assert.Less(t, elapsed, 3*time.Second)
}

func TestRunWithPanicRecovery_NoPanic(t *testing.T) {
	log, _ := logger.New(logger.Config{Level: logger.ErrorLevel})

	executed := false
	fn := func() {
		executed = true
	}

	RunWithPanicRecovery(fn, log)
	assert.True(t, executed)
}

func TestRunWithPanicRecovery_WithPanic(t *testing.T) {
	log, _ := logger.New(logger.Config{Level: logger.ErrorLevel})

	fn := func() {
		panic("test panic")
	}

	// Should not panic, should recover
	assert.NotPanics(t, func() {
		RunWithPanicRecovery(fn, log)
	})
}

func TestScheduler_Stats(t *testing.T) {
	cfg := createTestConfig()
	log, _ := logger.New(logger.Config{Level: logger.ErrorLevel})

	pool := NewWorkerPool(WorkerPoolConfig{
		Workers: 2,
		Logger:  log,
	})

	scheduler := NewScheduler(SchedulerConfig{
		Config:     cfg,
		WorkerPool: pool,
		Logger:     log,
		Interval:   1 * time.Minute,
	})

	stats := scheduler.Stats()
	assert.False(t, stats["running"].(bool))
	assert.Equal(t, 0, stats["cycles_run"].(int))
	assert.Equal(t, 1*time.Minute, stats["interval"].(time.Duration))
}

func TestScheduler_EmptyKeywords(t *testing.T) {
	cfg := createTestConfig()
	cfg.Keywords = []config.Keyword{} // Empty keywords
	log, _ := logger.New(logger.Config{Level: logger.ErrorLevel})

	pool := NewWorkerPool(WorkerPoolConfig{
		Workers: 2,
		Logger:  log,
	})

	scheduler := NewScheduler(SchedulerConfig{
		Config:     cfg,
		WorkerPool: pool,
		Logger:     log,
		Interval:   1 * time.Second,
	})

	err := scheduler.Start(false)
	require.NoError(t, err)

	// Wait for cycle attempt
	time.Sleep(100 * time.Millisecond)

	// Should handle gracefully (cycle will fail but scheduler should continue)
	scheduler.Stop()
}

func TestScheduler_MultipleStarts(t *testing.T) {
	cfg := createTestConfig()
	log, _ := logger.New(logger.Config{Level: logger.ErrorLevel})

	pool := NewWorkerPool(WorkerPoolConfig{
		Workers:   2,
		QueueSize: 10,
		Logger:    log,
		Executor: func(task *Task) *TaskResult {
			time.Sleep(10 * time.Millisecond)
			task.MarkCompleted()
			return NewTaskResult(task, true, nil)
		},
	})

	scheduler := NewScheduler(SchedulerConfig{
		Config:     cfg,
		WorkerPool: pool,
		Logger:     log,
		Interval:   5 * time.Second,
	})

	// First start
	err := scheduler.Start(false)
	require.NoError(t, err)

	// Try to start again while running
	err = scheduler.Start(false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already running")

	time.Sleep(100 * time.Millisecond)
	scheduler.Stop()
}

func TestScheduler_LongInterval(t *testing.T) {
	cfg := createTestConfig()
	log, _ := logger.New(logger.Config{Level: logger.ErrorLevel})

	pool := NewWorkerPool(WorkerPoolConfig{
		Workers:   2,
		QueueSize: 10,
		Logger:    log,
		Executor: func(task *Task) *TaskResult {
			task.MarkCompleted()
			return NewTaskResult(task, true, nil)
		},
	})

	scheduler := NewScheduler(SchedulerConfig{
		Config:     cfg,
		WorkerPool: pool,
		Logger:     log,
		Interval:   10 * time.Minute, // Long interval
	})

	err := scheduler.Start(false)
	require.NoError(t, err)

	// Wait for first cycle
	time.Sleep(100 * time.Millisecond)

	// Should complete one cycle
	stats := scheduler.Stats()
	assert.Equal(t, 1, stats["cycles_run"])

	scheduler.Stop()
}

func TestScheduler_WorkerPoolNotRunning(t *testing.T) {
	cfg := createTestConfig()
	log, _ := logger.New(logger.Config{Level: logger.ErrorLevel})

	tasksExecuted := 0
	pool := NewWorkerPool(WorkerPoolConfig{
		Workers:   2,
		QueueSize: 10,
		Logger:    log,
		Executor: func(task *Task) *TaskResult {
			tasksExecuted++
			task.MarkCompleted()
			return NewTaskResult(task, true, nil)
		},
	})
	// Don't start the worker pool - scheduler should start it

	scheduler := NewScheduler(SchedulerConfig{
		Config:     cfg,
		WorkerPool: pool,
		Logger:     log,
		Interval:   1 * time.Second,
	})

	err := scheduler.Start(false)
	require.NoError(t, err)

	// Wait for cycle to complete
	time.Sleep(200 * time.Millisecond)

	// Tasks should have been executed
	assert.Greater(t, tasksExecuted, 0)

	scheduler.Stop()
}

func TestRetryWithBackoff_ContextTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	attempts := 0
	fn := func() error {
		attempts++
		time.Sleep(50 * time.Millisecond)
		return errors.New("error")
	}

	err := RetryWithBackoff(ctx, 10, 30*time.Millisecond, fn)
	assert.Error(t, err)
	// Should stop early due to context timeout
	assert.Less(t, attempts, 10)
}

func TestScheduler_ResultCollection(t *testing.T) {
	cfg := createTestConfig()
	log, _ := logger.New(logger.Config{Level: logger.ErrorLevel})

	pool := NewWorkerPool(WorkerPoolConfig{
		Workers:   2,
		QueueSize: 10,
		Logger:    log,
		Executor: func(task *Task) *TaskResult {
			task.MarkRunning()
			task.MarkCompleted()
			result := NewTaskResult(task, true, nil)
			result.Position = 5
			result.PageNumber = 1
			return result
		},
	})

	scheduler := NewScheduler(SchedulerConfig{
		Config:     cfg,
		WorkerPool: pool,
		Logger:     log,
		Interval:   1 * time.Second,
	})

	err := scheduler.Start(false)
	require.NoError(t, err)

	// Wait for cycle
	time.Sleep(200 * time.Millisecond)

	stats := scheduler.Stats()
	assert.Equal(t, 1, stats["cycles_run"])

	scheduler.Stop()
}
