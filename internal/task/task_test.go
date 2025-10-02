package task

import (
	"errors"
	"testing"
	"time"

	"github.com/omer/go-bot/internal/logger"
	"github.com/omer/go-bot/internal/proxy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ===== Task creation tests =====

func TestNewTask_Success(t *testing.T) {
	config := TaskConfig{
		Keyword:   "golang tutorial",
		TargetURL: "example.com",
		ProxyURL:  "http://proxy:8080",
	}

	task, err := NewTask(config)
	require.NoError(t, err)
	assert.NotNil(t, task)
	assert.NotEmpty(t, task.ID)
	assert.Equal(t, "golang tutorial", task.Keyword)
	assert.Equal(t, "example.com", task.TargetURL)
	assert.Equal(t, "http://proxy:8080", task.ProxyURL)
	assert.Equal(t, TaskStatusPending, task.Status)
	assert.Equal(t, TaskTypeSearch, task.Type)
	assert.False(t, task.CreatedAt.IsZero())
	assert.Nil(t, task.StartedAt)
	assert.Nil(t, task.CompletedAt)
}

func TestNewTask_WithMetadata(t *testing.T) {
	metadata := map[string]interface{}{
		"campaign": "test",
		"priority": 1,
	}

	config := TaskConfig{
		Keyword:   "test",
		TargetURL: "example.com",
		Metadata:  metadata,
	}

	task, err := NewTask(config)
	require.NoError(t, err)
	assert.Equal(t, metadata, task.Metadata)
}

func TestNewTask_MissingKeyword(t *testing.T) {
	config := TaskConfig{
		TargetURL: "example.com",
	}

	task, err := NewTask(config)
	assert.Error(t, err)
	assert.Nil(t, task)
	assert.Contains(t, err.Error(), "keyword is required")
}

func TestNewTask_MissingTargetURL(t *testing.T) {
	config := TaskConfig{
		Keyword: "golang",
	}

	task, err := NewTask(config)
	assert.Error(t, err)
	assert.Nil(t, task)
	assert.Contains(t, err.Error(), "target URL is required")
}

// ===== Task state management tests =====

func TestTask_MarkRunning(t *testing.T) {
	task, _ := NewTask(TaskConfig{
		Keyword:   "test",
		TargetURL: "example.com",
	})

	assert.Nil(t, task.StartedAt)
	assert.Equal(t, TaskStatusPending, task.Status)

	task.MarkRunning()

	assert.NotNil(t, task.StartedAt)
	assert.Equal(t, TaskStatusRunning, task.Status)
}

func TestTask_MarkCompleted(t *testing.T) {
	task, _ := NewTask(TaskConfig{
		Keyword:   "test",
		TargetURL: "example.com",
	})

	task.MarkRunning()
	time.Sleep(10 * time.Millisecond)
	task.MarkCompleted()

	assert.NotNil(t, task.CompletedAt)
	assert.Equal(t, TaskStatusCompleted, task.Status)
	assert.True(t, task.IsCompleted())
}

func TestTask_MarkFailed(t *testing.T) {
	task, _ := NewTask(TaskConfig{
		Keyword:   "test",
		TargetURL: "example.com",
	})

	task.MarkRunning()
	task.MarkFailed()

	assert.NotNil(t, task.CompletedAt)
	assert.Equal(t, TaskStatusFailed, task.Status)
	assert.True(t, task.IsCompleted())
}

func TestTask_Duration(t *testing.T) {
	task, _ := NewTask(TaskConfig{
		Keyword:   "test",
		TargetURL: "example.com",
	})

	// Duration should be 0 before starting
	assert.Equal(t, time.Duration(0), task.Duration())

	task.MarkRunning()

	// Duration should still be 0 if not completed
	assert.Equal(t, time.Duration(0), task.Duration())

	time.Sleep(50 * time.Millisecond)
	task.MarkCompleted()

	// Duration should be greater than 50ms
	duration := task.Duration()
	assert.Greater(t, duration, 50*time.Millisecond)
	assert.Less(t, duration, 200*time.Millisecond)
}

func TestTask_IsCompleted(t *testing.T) {
	task, _ := NewTask(TaskConfig{
		Keyword:   "test",
		TargetURL: "example.com",
	})

	assert.False(t, task.IsCompleted())

	task.MarkRunning()
	assert.False(t, task.IsCompleted())

	task.MarkCompleted()
	assert.True(t, task.IsCompleted())
}

// ===== TaskResult tests =====

func TestNewTaskResult_Success(t *testing.T) {
	task, _ := NewTask(TaskConfig{
		Keyword:   "test",
		TargetURL: "example.com",
	})
	task.MarkRunning()
	time.Sleep(10 * time.Millisecond)
	task.MarkCompleted()

	result := NewTaskResult(task, true, nil)

	assert.NotNil(t, result)
	assert.Equal(t, task, result.Task)
	assert.True(t, result.Success)
	assert.Nil(t, result.Error)
	assert.Greater(t, result.Duration, time.Duration(0))
}

func TestNewTaskResult_Failure(t *testing.T) {
	task, _ := NewTask(TaskConfig{
		Keyword:   "test",
		TargetURL: "example.com",
	})
	task.MarkRunning()
	task.MarkFailed()

	testErr := errors.New("test error")
	result := NewTaskResult(task, false, testErr)

	assert.NotNil(t, result)
	assert.False(t, result.Success)
	assert.Equal(t, testErr, result.Error)
}

func TestNewTaskResult_RunningTask(t *testing.T) {
	task, _ := NewTask(TaskConfig{
		Keyword:   "test",
		TargetURL: "example.com",
	})
	task.MarkRunning()
	time.Sleep(10 * time.Millisecond)

	// Create result while task is still running
	result := NewTaskResult(task, false, nil)

	// Duration should be calculated from start time
	assert.Greater(t, result.Duration, time.Duration(0))
}

// ===== WorkerPool creation tests =====

func TestNewWorkerPool_Defaults(t *testing.T) {
	log, err := logger.New(logger.Config{
		Level:      logger.InfoLevel,
		EnableFile: false,
	})
	require.NoError(t, err)

	pool := NewWorkerPool(WorkerPoolConfig{
		Workers: 0, // Should default to 1
		Logger:  log,
	})

	assert.NotNil(t, pool)
	assert.Equal(t, 1, pool.workers)
	assert.NotNil(t, pool.taskQueue)
	assert.NotNil(t, pool.resultQueue)
	assert.False(t, pool.IsRunning())
}

func TestNewWorkerPool_CustomConfig(t *testing.T) {
	log, err := logger.New(logger.Config{
		Level:      logger.InfoLevel,
		EnableFile: false,
	})
	require.NoError(t, err)

	pool := NewWorkerPool(WorkerPoolConfig{
		Workers:   5,
		QueueSize: 100,
		Logger:    log,
	})

	assert.Equal(t, 5, pool.workers)
	assert.Equal(t, 100, cap(pool.taskQueue))
	assert.False(t, pool.IsRunning())
}

// ===== WorkerPool lifecycle tests =====

func TestWorkerPool_StartStop(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping worker pool test in short mode")
	}

	log, err := logger.New(logger.Config{
		Level:      logger.InfoLevel,
		EnableFile: false,
	})
	require.NoError(t, err)

	pool := NewWorkerPool(WorkerPoolConfig{
		Workers: 2,
		Logger:  log,
	})

	// Start pool
	err = pool.Start()
	assert.NoError(t, err)
	assert.True(t, pool.IsRunning())

	// Try to start again (should error)
	err = pool.Start()
	assert.Error(t, err)

	// Stop pool
	err = pool.Stop()
	assert.NoError(t, err)
	assert.False(t, pool.IsRunning())

	// Try to stop again (should error)
	err = pool.Stop()
	assert.Error(t, err)
}

func TestWorkerPool_Submit(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping worker pool test in short mode")
	}

	log, err := logger.New(logger.Config{
		Level:      logger.InfoLevel,
		EnableFile: false,
	})
	require.NoError(t, err)

	pool := NewWorkerPool(WorkerPoolConfig{
		Workers:   1,
		QueueSize: 10,
		Logger:    log,
	})

	// Submit before starting should fail
	task, _ := NewTask(TaskConfig{
		Keyword:   "test",
		TargetURL: "example.com",
	})
	err = pool.Submit(task)
	assert.Error(t, err)

	// Start pool
	err = pool.Start()
	require.NoError(t, err)
	defer pool.Stop()

	// Submit should succeed
	err = pool.Submit(task)
	assert.NoError(t, err)
}

func TestWorkerPool_ExecuteTask(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping worker pool test in short mode")
	}

	log, err := logger.New(logger.Config{
		Level:      logger.DebugLevel,
		EnableFile: false,
	})
	require.NoError(t, err)

	// Create a mock executor
	mockExecutor := func(task *Task) *TaskResult {
		task.MarkRunning()
		time.Sleep(10 * time.Millisecond)
		task.MarkCompleted()
		return NewTaskResult(task, true, nil)
	}

	pool := NewWorkerPool(WorkerPoolConfig{
		Workers:  1,
		Logger:   log,
		Executor: mockExecutor,
	})

	err = pool.Start()
	require.NoError(t, err)
	defer pool.Stop()

	// Submit task
	task, _ := NewTask(TaskConfig{
		Keyword:   "test",
		TargetURL: "example.com",
	})
	err = pool.Submit(task)
	require.NoError(t, err)

	// Wait for result
	select {
	case result := <-pool.GetResults():
		assert.NotNil(t, result)
		assert.True(t, result.Success)
		assert.Equal(t, task.ID, result.Task.ID)
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for result")
	}
}

func TestWorkerPool_MultipleTasks(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping worker pool test in short mode")
	}

	log, err := logger.New(logger.Config{
		Level:      logger.InfoLevel,
		EnableFile: false,
	})
	require.NoError(t, err)

	// Create a mock executor
	executedTasks := 0
	mockExecutor := func(task *Task) *TaskResult {
		task.MarkRunning()
		time.Sleep(5 * time.Millisecond)
		task.MarkCompleted()
		executedTasks++
		return NewTaskResult(task, true, nil)
	}

	pool := NewWorkerPool(WorkerPoolConfig{
		Workers:  2,
		Logger:   log,
		Executor: mockExecutor,
	})

	err = pool.Start()
	require.NoError(t, err)

	// Submit multiple tasks
	numTasks := 5
	for i := 0; i < numTasks; i++ {
		task, _ := NewTask(TaskConfig{
			Keyword:   "test",
			TargetURL: "example.com",
		})
		err = pool.Submit(task)
		require.NoError(t, err)
	}

	// Collect results
	results := make([]*TaskResult, 0, numTasks)
	for i := 0; i < numTasks; i++ {
		select {
		case result := <-pool.GetResults():
			results = append(results, result)
		case <-time.After(2 * time.Second):
			t.Fatalf("Timeout waiting for result %d/%d", i+1, numTasks)
		}
	}

	assert.Len(t, results, numTasks)
	assert.Equal(t, numTasks, executedTasks)

	err = pool.Stop()
	assert.NoError(t, err)
}

func TestWorkerPool_Stats(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping worker pool test in short mode")
	}

	log, err := logger.New(logger.Config{
		Level:      logger.InfoLevel,
		EnableFile: false,
	})
	require.NoError(t, err)

	mockExecutor := func(task *Task) *TaskResult {
		task.MarkRunning()
		task.MarkCompleted()
		return NewTaskResult(task, true, nil)
	}

	pool := NewWorkerPool(WorkerPoolConfig{
		Workers:  2,
		Logger:   log,
		Executor: mockExecutor,
	})

	stats := pool.Stats()
	assert.Equal(t, 2, stats["workers"])
	assert.False(t, stats["running"].(bool))

	err = pool.Start()
	require.NoError(t, err)
	defer pool.Stop()

	stats = pool.Stats()
	assert.True(t, stats["running"].(bool))
	assert.Equal(t, 0, stats["tasks_started"])
	assert.Equal(t, 0, stats["tasks_done"])
}

func TestWorkerPool_WithProxyPool(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping worker pool test in short mode")
	}

	log, err := logger.New(logger.Config{
		Level:      logger.InfoLevel,
		EnableFile: false,
	})
	require.NoError(t, err)

	// Create proxy pool
	proxyURLs := []string{"http://proxy1:8080", "http://proxy2:8080"}
	proxyPool, err := proxy.NewProxyPool(proxyURLs, proxy.RotationStrategyRoundRobin)
	require.NoError(t, err)

	mockExecutor := func(task *Task) *TaskResult {
		task.MarkRunning()
		task.MarkCompleted()
		return NewTaskResult(task, true, nil)
	}

	pool := NewWorkerPool(WorkerPoolConfig{
		Workers:   1,
		Logger:    log,
		ProxyPool: proxyPool,
		Executor:  mockExecutor,
	})

	err = pool.Start()
	require.NoError(t, err)
	defer pool.Stop()

	task, _ := NewTask(TaskConfig{
		Keyword:   "test",
		TargetURL: "example.com",
	})
	err = pool.Submit(task)
	require.NoError(t, err)

	// Wait for result
	select {
	case result := <-pool.GetResults():
		assert.True(t, result.Success)
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout")
	}
}

// ===== generateTaskID test =====

func TestGenerateTaskID(t *testing.T) {
	id1 := generateTaskID()
	time.Sleep(1 * time.Millisecond) // Ensure different timestamp
	id2 := generateTaskID()

	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.NotEqual(t, id1, id2) // IDs should be unique
	assert.Contains(t, id1, "task-")
}
