package task

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/omer/go-bot/internal/config"
	"github.com/omer/go-bot/internal/logger"
	"github.com/omer/go-bot/internal/stats"
)

// Scheduler manages continuous task execution with configurable intervals
type Scheduler struct {
	config         *config.Config
	workerPool     *WorkerPool
	statsCollector *stats.StatsCollector
	logger         *logger.Logger
	interval       time.Duration
	running        bool
	mu             sync.RWMutex
	ctx            context.Context
	cancel         context.CancelFunc
	wg             sync.WaitGroup
	cyclesRun      int
	lastCycleAt    time.Time
}

// SchedulerConfig holds configuration for creating a scheduler
type SchedulerConfig struct {
	Config         *config.Config        // Application config
	WorkerPool     *WorkerPool           // Worker pool for task execution
	StatsCollector *stats.StatsCollector // Stats collector
	Logger         *logger.Logger        // Logger instance
	Interval       time.Duration         // Interval between cycles (0 = run once)
}

// NewScheduler creates a new scheduler instance
//
// Example:
//
//	scheduler := NewScheduler(SchedulerConfig{
//	    Config:     cfg,
//	    WorkerPool: pool,
//	    Logger:     log,
//	    Interval:   5 * time.Minute,
//	})
func NewScheduler(config SchedulerConfig) *Scheduler {
	if config.Interval == 0 {
		config.Interval = 5 * time.Minute
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Scheduler{
		config:         config.Config,
		workerPool:     config.WorkerPool,
		statsCollector: config.StatsCollector,
		logger:         config.Logger,
		interval:       config.Interval,
		running:        false,
		ctx:            ctx,
		cancel:         cancel,
		cyclesRun:      0,
	}
}

// Start starts the scheduler
// If continuous is true, it runs in an infinite loop with interval delays
// If continuous is false, it runs once and stops
func (s *Scheduler) Start(continuous bool) error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return fmt.Errorf("scheduler is already running")
	}
	s.running = true
	s.mu.Unlock()

	s.logger.Info("Starting scheduler", map[string]interface{}{
		"continuous": continuous,
		"interval":   s.interval,
		"keywords":   len(s.config.Keywords),
	})

	// Start the scheduler loop
	s.wg.Add(1)
	go s.run(continuous)

	return nil
}

// Stop stops the scheduler gracefully
func (s *Scheduler) Stop() error {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return fmt.Errorf("scheduler is not running")
	}
	s.mu.Unlock()

	s.logger.Info("Stopping scheduler", map[string]interface{}{
		"cycles_run": s.cyclesRun,
	})

	// Cancel context
	s.cancel()

	// Wait for scheduler loop to finish
	s.wg.Wait()

	s.mu.Lock()
	s.running = false
	s.mu.Unlock()

	s.logger.Info("Scheduler stopped", nil)
	return nil
}

// IsRunning returns whether the scheduler is currently running
func (s *Scheduler) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// Stats returns current scheduler statistics
func (s *Scheduler) Stats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return map[string]interface{}{
		"running":    s.running,
		"cycles_run": s.cyclesRun,
		"last_cycle": s.lastCycleAt,
		"interval":   s.interval,
	}
}

// run is the main scheduler loop
func (s *Scheduler) run(continuous bool) {
	defer s.wg.Done()

	for {
		// Check if context is cancelled
		select {
		case <-s.ctx.Done():
			s.logger.Info("Scheduler loop stopping (context cancelled)", nil)
			return
		default:
		}

		// Run one cycle
		s.logger.Info("Starting scheduler cycle", map[string]interface{}{
			"cycle": s.cyclesRun + 1,
		})

		err := s.runCycle()
		if err != nil {
			s.logger.Error("Scheduler cycle failed", map[string]interface{}{
				"error": err,
				"cycle": s.cyclesRun + 1,
			})
		} else {
			s.mu.Lock()
			s.cyclesRun++
			s.lastCycleAt = time.Now()
			s.mu.Unlock()

			s.logger.Info("Scheduler cycle completed", map[string]interface{}{
				"cycle": s.cyclesRun,
			})
		}

		// If not continuous, stop after one cycle
		if !continuous {
			s.logger.Info("Scheduler running in single-cycle mode, stopping", nil)
			return
		}

		// Wait for interval before next cycle
		s.logger.Info("Waiting before next cycle", map[string]interface{}{
			"interval": s.interval,
		})

		select {
		case <-s.ctx.Done():
			s.logger.Info("Scheduler loop stopping (context cancelled during wait)", nil)
			return
		case <-time.After(s.interval):
			// Continue to next cycle
		}
	}
}

// runCycle executes one full cycle of tasks
func (s *Scheduler) runCycle() error {
	// Start worker pool if not running
	if !s.workerPool.IsRunning() {
		err := s.workerPool.Start()
		if err != nil {
			return fmt.Errorf("failed to start worker pool: %w", err)
		}
	}

	// Create tasks for all keywords
	tasks := make([]*Task, 0, len(s.config.Keywords))
	for _, kw := range s.config.Keywords {
		task, err := NewTask(TaskConfig{
			Keyword:   kw.Term,
			TargetURL: kw.TargetURL,
		})
		if err != nil {
			s.logger.Error("Failed to create task", map[string]interface{}{
				"error":   err,
				"keyword": kw.Term,
			})
			continue
		}
		tasks = append(tasks, task)
	}

	if len(tasks) == 0 {
		return fmt.Errorf("no tasks created")
	}

	// Submit all tasks
	s.logger.Info("Submitting tasks", map[string]interface{}{
		"count": len(tasks),
	})

	for _, task := range tasks {
		err := s.workerPool.Submit(task)
		if err != nil {
			s.logger.Error("Failed to submit task", map[string]interface{}{
				"error":   err,
				"task_id": task.ID,
			})
		}
	}

	// Collect results
	resultsCollected := 0
	for resultsCollected < len(tasks) {
		select {
		case <-s.ctx.Done():
			return fmt.Errorf("context cancelled while collecting results")
		case result := <-s.workerPool.GetResults():
			resultsCollected++

			s.logger.Info("Task completed", map[string]interface{}{
				"task_id":  result.Task.ID,
				"keyword":  result.Task.Keyword,
				"success":  result.Success,
				"duration": result.Duration,
				"position": result.Position,
			})

			// Record stats if collector is available
			if s.statsCollector != nil {
				taskStats := stats.TaskStats{
					TaskID:     result.Task.ID,
					Keyword:    result.Task.Keyword,
					TargetURL:  result.Task.TargetURL,
					Success:    result.Success,
					Position:   result.Position,
					PageNumber: result.PageNumber,
					Duration:   float64(result.Duration.Milliseconds()),
					Timestamp:  time.Now(),
				}

				if result.Error != nil {
					taskStats.Error = result.Error.Error()
				}

				s.statsCollector.RecordTask(taskStats)
			}
		}
	}

	s.logger.Info("All tasks completed", map[string]interface{}{
		"total":     len(tasks),
		"collected": resultsCollected,
	})

	return nil
}

// RetryWithBackoff executes a function with exponential backoff retry logic
func RetryWithBackoff(ctx context.Context, maxRetries int, initialDelay time.Duration, fn func() error) error {
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		// Check context
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Try function
		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		// Don't wait after last attempt
		if attempt == maxRetries-1 {
			break
		}

		// Calculate backoff delay: initialDelay * 2^attempt
		backoff := time.Duration(float64(initialDelay) * math.Pow(2, float64(attempt)))

		// Cap backoff at 5 minutes
		maxBackoff := 5 * time.Minute
		if backoff > maxBackoff {
			backoff = maxBackoff
		}

		// Wait before retry
		timer := time.NewTimer(backoff)
		select {
		case <-timer.C:
			// Continue to next attempt
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		}
	}

	return fmt.Errorf("operation failed after %d attempts: %w", maxRetries, lastErr)
}

// RunWithPanicRecovery wraps a function with panic recovery
func RunWithPanicRecovery(fn func(), logger *logger.Logger) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("Panic recovered", map[string]interface{}{
				"panic": r,
			})
		}
	}()

	fn()
}
