package task

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/omer/go-bot/internal/browser"
	"github.com/omer/go-bot/internal/logger"
	"github.com/omer/go-bot/internal/proxy"
	"github.com/omer/go-bot/internal/serp"
)

// TaskExecutor is a function type that executes a task
type TaskExecutor func(task *Task) *TaskResult

// WorkerPool manages a pool of workers for concurrent task execution
type WorkerPool struct {
	workers      int                // Number of worker goroutines
	taskQueue    chan *Task         // Channel for incoming tasks
	resultQueue  chan *TaskResult   // Channel for task results
	wg           sync.WaitGroup     // WaitGroup for worker synchronization
	ctx          context.Context    // Context for cancellation
	cancel       context.CancelFunc // Cancel function
	proxyPool    *proxy.ProxyPool   // Proxy pool
	logger       *logger.Logger     // Logger
	executor     TaskExecutor       // Custom task executor (for testing)
	running      bool               // Whether the pool is running
	mu           sync.RWMutex       // Mutex for concurrent access
	tasksStarted int                // Number of tasks started
	tasksDone    int                // Number of tasks completed
}

// WorkerPoolConfig holds configuration for creating a worker pool
type WorkerPoolConfig struct {
	Workers   int              // Number of worker goroutines
	QueueSize int              // Size of task queue (0 for unbuffered)
	ProxyPool *proxy.ProxyPool // Proxy pool for rotation
	Logger    *logger.Logger   // Logger instance
	Executor  TaskExecutor     // Optional custom executor (for testing)
}

// NewWorkerPool creates a new worker pool
//
// Example:
//
//	pool := NewWorkerPool(WorkerPoolConfig{
//	    Workers:   5,
//	    QueueSize: 100,
//	    Logger:    log,
//	})
func NewWorkerPool(config WorkerPoolConfig) *WorkerPool {
	// Set defaults
	if config.Workers <= 0 {
		config.Workers = 1
	}
	if config.QueueSize < 0 {
		config.QueueSize = 0
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &WorkerPool{
		workers:     config.Workers,
		taskQueue:   make(chan *Task, config.QueueSize),
		resultQueue: make(chan *TaskResult, config.Workers),
		ctx:         ctx,
		cancel:      cancel,
		proxyPool:   config.ProxyPool,
		logger:      config.Logger,
		executor:    config.Executor,
		running:     false,
	}
}

// Start starts the worker pool
// It spawns worker goroutines and starts processing tasks
func (wp *WorkerPool) Start() error {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	if wp.running {
		return fmt.Errorf("worker pool is already running")
	}

	wp.logger.Info("Starting worker pool", map[string]interface{}{
		"workers": wp.workers,
	})

	// Start worker goroutines
	for i := 0; i < wp.workers; i++ {
		wp.wg.Add(1)
		go wp.worker(i)
	}

	wp.running = true
	return nil
}

// Stop stops the worker pool gracefully
// It waits for all workers to finish their current tasks
func (wp *WorkerPool) Stop() error {
	wp.mu.Lock()
	if !wp.running {
		wp.mu.Unlock()
		return fmt.Errorf("worker pool is not running")
	}
	wp.mu.Unlock()

	wp.logger.Info("Stopping worker pool", nil)

	// Close task queue to signal workers to stop
	close(wp.taskQueue)

	// Wait for all workers to finish
	wp.wg.Wait()

	// Close result queue
	close(wp.resultQueue)

	// Cancel context
	wp.cancel()

	wp.mu.Lock()
	wp.running = false
	wp.mu.Unlock()

	wp.logger.Info("Worker pool stopped", map[string]interface{}{
		"tasks_started": wp.tasksStarted,
		"tasks_done":    wp.tasksDone,
	})

	return nil
}

// Submit submits a task to the worker pool
// Returns error if the pool is not running or context is cancelled
func (wp *WorkerPool) Submit(task *Task) error {
	wp.mu.RLock()
	defer wp.mu.RUnlock()

	if !wp.running {
		return fmt.Errorf("worker pool is not running")
	}

	select {
	case <-wp.ctx.Done():
		return fmt.Errorf("worker pool is shutting down")
	case wp.taskQueue <- task:
		wp.logger.Debug("Task submitted", map[string]interface{}{
			"task_id": task.ID,
			"keyword": task.Keyword,
		})
		return nil
	}
}

// GetResults returns the result channel
// Results can be read from this channel as tasks complete
func (wp *WorkerPool) GetResults() <-chan *TaskResult {
	return wp.resultQueue
}

// IsRunning returns whether the worker pool is currently running
func (wp *WorkerPool) IsRunning() bool {
	wp.mu.RLock()
	defer wp.mu.RUnlock()
	return wp.running
}

// Stats returns current worker pool statistics
func (wp *WorkerPool) Stats() map[string]interface{} {
	wp.mu.RLock()
	defer wp.mu.RUnlock()

	return map[string]interface{}{
		"workers":       wp.workers,
		"running":       wp.running,
		"tasks_started": wp.tasksStarted,
		"tasks_done":    wp.tasksDone,
		"queue_length":  len(wp.taskQueue),
	}
}

// worker is the main worker goroutine function
func (wp *WorkerPool) worker(id int) {
	defer wp.wg.Done()

	wp.logger.Debug("Worker started", map[string]interface{}{
		"worker_id": id,
	})

	for {
		select {
		case <-wp.ctx.Done():
			// Context cancelled, stop worker
			wp.logger.Debug("Worker stopping (context done)", map[string]interface{}{
				"worker_id": id,
			})
			return

		case task, ok := <-wp.taskQueue:
			if !ok {
				// Task queue closed, stop worker
				wp.logger.Debug("Worker stopping (queue closed)", map[string]interface{}{
					"worker_id": id,
				})
				return
			}

			// Execute task
			wp.mu.Lock()
			wp.tasksStarted++
			wp.mu.Unlock()

			wp.logger.Info("Worker executing task", map[string]interface{}{
				"worker_id": id,
				"task_id":   task.ID,
				"keyword":   task.Keyword,
			})

			var result *TaskResult
			if wp.executor != nil {
				// Use custom executor (for testing)
				result = wp.executor(task)
			} else {
				// Use default executor
				result = wp.executeTask(task)
			}

			wp.mu.Lock()
			wp.tasksDone++
			wp.mu.Unlock()

			// Send result
			select {
			case wp.resultQueue <- result:
				wp.logger.Debug("Task result sent", map[string]interface{}{
					"task_id": task.ID,
					"success": result.Success,
				})
			case <-wp.ctx.Done():
				wp.logger.Warn("Failed to send result (context done)", map[string]interface{}{
					"task_id": task.ID,
				})
				return
			}
		}
	}
}

// executeTask executes a single task
func (wp *WorkerPool) executeTask(task *Task) *TaskResult {
	task.MarkRunning()

	// Get proxy if pool is available
	var taskProxy *proxy.Proxy
	var proxySuccess bool = false
	if wp.proxyPool != nil {
		var err error
		taskProxy, err = wp.proxyPool.Get()
		if err == nil && taskProxy != nil {
			defer func() {
				wp.proxyPool.Release(taskProxy, proxySuccess)
			}()

			if task.ProxyURL == "" {
				task.ProxyURL = taskProxy.String()
			}
		}
	}

	// Create browser
	browserOpts := browser.BrowserOptions{
		Headless: true,
		Proxy:    taskProxy,
		Timeout:  60 * time.Second,
	}

	b, err := browser.NewBrowser(browserOpts)
	if err != nil {
		task.MarkFailed()
		return NewTaskResult(task, false, fmt.Errorf("failed to create browser: %w", err))
	}
	defer b.Close()

	// Create searcher
	searcher := serp.NewSearcher(b, wp.logger)

	// Perform search
	err = searcher.Search(task.Keyword)
	if err != nil {
		task.MarkFailed()
		return NewTaskResult(task, false, fmt.Errorf("search failed: %w", err))
	}

	// Find target
	result, err := searcher.FindTarget(task.TargetURL)
	if err != nil {
		task.MarkFailed()
		return NewTaskResult(task, false, fmt.Errorf("target not found: %w", err))
	}

	// Click target
	err = searcher.ClickTargetResult(task.TargetURL)
	if err != nil {
		task.MarkFailed()
		return NewTaskResult(task, false, fmt.Errorf("failed to click target: %w", err))
	}

	// Wait on target page
	time.Sleep(5 * time.Second)

	task.MarkCompleted()
	proxySuccess = true // Mark proxy as successful

	taskResult := NewTaskResult(task, true, nil)
	taskResult.Position = result.Position
	taskResult.PageNumber = 1 // TODO: Get actual page number
	taskResult.Message = fmt.Sprintf("Found and clicked target at position %d", result.Position)

	return taskResult
}
