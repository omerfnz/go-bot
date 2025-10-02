// Package task provides task management and worker pool functionality.
// It enables concurrent execution of search tasks with configurable workers.
package task

import (
	"fmt"
	"time"
)

// TaskType represents the type of task to execute
type TaskType string

const (
	// TaskTypeSearch represents a search task
	TaskTypeSearch TaskType = "search"
	// TaskTypeClick represents a click task
	TaskTypeClick TaskType = "click"
)

// TaskStatus represents the current status of a task
type TaskStatus string

const (
	// TaskStatusPending means the task is waiting to be executed
	TaskStatusPending TaskStatus = "pending"
	// TaskStatusRunning means the task is currently being executed
	TaskStatusRunning TaskStatus = "running"
	// TaskStatusCompleted means the task completed successfully
	TaskStatusCompleted TaskStatus = "completed"
	// TaskStatusFailed means the task failed
	TaskStatusFailed TaskStatus = "failed"
)

// Task represents a single task to be executed
type Task struct {
	ID          string                 // Unique task ID
	Type        TaskType               // Type of task (search, click, etc.)
	Keyword     string                 // Search keyword
	TargetURL   string                 // Target URL to find and click
	ProxyURL    string                 // Proxy URL to use (optional)
	Status      TaskStatus             // Current task status
	CreatedAt   time.Time              // Task creation time
	StartedAt   *time.Time             // Task start time (nil if not started)
	CompletedAt *time.Time             // Task completion time (nil if not completed)
	Metadata    map[string]interface{} // Additional metadata
}

// TaskResult represents the result of an executed task
type TaskResult struct {
	Task       *Task         // Reference to the original task
	Success    bool          // Whether the task succeeded
	Error      error         // Error if task failed
	Position   int           // Position where target was found (0 if not found)
	PageNumber int           // Page number where target was found
	Duration   time.Duration // Task execution duration
	Message    string        // Additional message or details
}

// TaskConfig holds configuration for creating a new task
type TaskConfig struct {
	Keyword   string                 // Required: Search keyword
	TargetURL string                 // Required: Target URL
	ProxyURL  string                 // Optional: Proxy URL
	Metadata  map[string]interface{} // Optional: Additional metadata
}

// NewTask creates a new task with the given configuration
//
// Example:
//
//	task := NewTask(TaskConfig{
//	    Keyword:   "golang tutorial",
//	    TargetURL: "example.com",
//	})
func NewTask(config TaskConfig) (*Task, error) {
	// Validate required fields
	if config.Keyword == "" {
		return nil, fmt.Errorf("keyword is required")
	}
	if config.TargetURL == "" {
		return nil, fmt.Errorf("target URL is required")
	}

	// Generate task ID
	taskID := generateTaskID()

	task := &Task{
		ID:        taskID,
		Type:      TaskTypeSearch,
		Keyword:   config.Keyword,
		TargetURL: config.TargetURL,
		ProxyURL:  config.ProxyURL,
		Status:    TaskStatusPending,
		CreatedAt: time.Now(),
		Metadata:  config.Metadata,
	}

	return task, nil
}

// MarkRunning marks the task as running
func (t *Task) MarkRunning() {
	t.Status = TaskStatusRunning
	now := time.Now()
	t.StartedAt = &now
}

// MarkCompleted marks the task as completed
func (t *Task) MarkCompleted() {
	t.Status = TaskStatusCompleted
	now := time.Now()
	t.CompletedAt = &now
}

// MarkFailed marks the task as failed
func (t *Task) MarkFailed() {
	t.Status = TaskStatusFailed
	now := time.Now()
	t.CompletedAt = &now
}

// Duration returns the task execution duration
// Returns 0 if task hasn't started or completed
func (t *Task) Duration() time.Duration {
	if t.StartedAt == nil || t.CompletedAt == nil {
		return 0
	}
	return t.CompletedAt.Sub(*t.StartedAt)
}

// IsCompleted returns true if the task is in a terminal state (completed or failed)
func (t *Task) IsCompleted() bool {
	return t.Status == TaskStatusCompleted || t.Status == TaskStatusFailed
}

// NewTaskResult creates a new TaskResult
func NewTaskResult(task *Task, success bool, err error) *TaskResult {
	duration := task.Duration()
	if !task.IsCompleted() && task.StartedAt != nil {
		duration = time.Since(*task.StartedAt)
	}

	return &TaskResult{
		Task:     task,
		Success:  success,
		Error:    err,
		Duration: duration,
	}
}

// generateTaskID generates a unique task ID
func generateTaskID() string {
	// Simple ID generation based on timestamp
	// In production, you might want to use UUID or a more robust solution
	return fmt.Sprintf("task-%d", time.Now().UnixNano())
}
