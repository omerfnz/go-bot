// Package stats provides statistics collection and reporting functionality.
// It tracks task execution results, rankings, and performance metrics.
package stats

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// TaskStats represents statistics for a single task execution
type TaskStats struct {
	TaskID     string    `json:"task_id"`
	Keyword    string    `json:"keyword"`
	TargetURL  string    `json:"target_url"`
	Success    bool      `json:"success"`
	Position   int       `json:"position"`    // Position where target was found (0 if not found)
	PageNumber int       `json:"page_number"` // Page number where target was found
	Duration   float64   `json:"duration_ms"` // Duration in milliseconds
	ProxyUsed  string    `json:"proxy_used"`  // Proxy URL used
	Error      string    `json:"error"`       // Error message if failed
	Timestamp  time.Time `json:"timestamp"`   // When the task was executed
}

// KeywordStats represents aggregated statistics for a keyword
type KeywordStats struct {
	Keyword       string    `json:"keyword"`
	TargetURL     string    `json:"target_url"`
	TotalAttempts int       `json:"total_attempts"`
	SuccessCount  int       `json:"success_count"`
	FailureCount  int       `json:"failure_count"`
	AvgPosition   float64   `json:"avg_position"`
	AvgDuration   float64   `json:"avg_duration_ms"`
	LastSeen      time.Time `json:"last_seen"`
	BestPosition  int       `json:"best_position"`  // Best (lowest) position seen
	WorstPosition int       `json:"worst_position"` // Worst (highest) position seen
}

// Statistics represents the complete statistics collection
type Statistics struct {
	StartTime    time.Time               `json:"start_time"`
	LastUpdate   time.Time               `json:"last_update"`
	TotalTasks   int                     `json:"total_tasks"`
	SuccessTasks int                     `json:"success_tasks"`
	FailedTasks  int                     `json:"failed_tasks"`
	TaskHistory  []TaskStats             `json:"task_history"`
	KeywordStats map[string]KeywordStats `json:"keyword_stats"` // key: keyword-targeturl
}

// StatsCollector manages statistics collection
type StatsCollector struct {
	stats    *Statistics
	filePath string
	mu       sync.RWMutex
}

// NewStatsCollector creates a new statistics collector
//
// Example:
//
//	collector := NewStatsCollector("data/stats.json")
func NewStatsCollector(filePath string) *StatsCollector {
	return &StatsCollector{
		stats: &Statistics{
			StartTime:    time.Now(),
			LastUpdate:   time.Now(),
			TaskHistory:  make([]TaskStats, 0),
			KeywordStats: make(map[string]KeywordStats),
		},
		filePath: filePath,
	}
}

// RecordTask records a task execution result
//
// Example:
//
//	collector.RecordTask(TaskStats{
//	    Keyword:   "golang",
//	    TargetURL: "example.com",
//	    Success:   true,
//	    Position:  3,
//	})
func (sc *StatsCollector) RecordTask(taskStats TaskStats) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	// Update timestamp if not set
	if taskStats.Timestamp.IsZero() {
		taskStats.Timestamp = time.Now()
	}

	// Add to task history
	sc.stats.TaskHistory = append(sc.stats.TaskHistory, taskStats)

	// Update overall stats
	sc.stats.TotalTasks++
	if taskStats.Success {
		sc.stats.SuccessTasks++
	} else {
		sc.stats.FailedTasks++
	}
	sc.stats.LastUpdate = time.Now()

	// Update keyword stats
	if taskStats.Keyword != "" && taskStats.TargetURL != "" {
		sc.updateKeywordStats(taskStats)
	}
}

// updateKeywordStats updates aggregated keyword statistics
func (sc *StatsCollector) updateKeywordStats(taskStats TaskStats) {
	key := fmt.Sprintf("%s-%s", taskStats.Keyword, taskStats.TargetURL)

	kwStats, exists := sc.stats.KeywordStats[key]
	if !exists {
		kwStats = KeywordStats{
			Keyword:       taskStats.Keyword,
			TargetURL:     taskStats.TargetURL,
			BestPosition:  999999,
			WorstPosition: 0,
		}
	}

	// Update counts
	kwStats.TotalAttempts++
	if taskStats.Success {
		kwStats.SuccessCount++
	} else {
		kwStats.FailureCount++
	}

	// Update positions (only for successful tasks)
	if taskStats.Success && taskStats.Position > 0 {
		if taskStats.Position < kwStats.BestPosition {
			kwStats.BestPosition = taskStats.Position
		}
		if taskStats.Position > kwStats.WorstPosition {
			kwStats.WorstPosition = taskStats.Position
		}

		// Calculate average position
		totalPositions := float64(kwStats.AvgPosition) * float64(kwStats.SuccessCount-1)
		kwStats.AvgPosition = (totalPositions + float64(taskStats.Position)) / float64(kwStats.SuccessCount)
	}

	// Calculate average duration
	totalDuration := kwStats.AvgDuration * float64(kwStats.TotalAttempts-1)
	kwStats.AvgDuration = (totalDuration + taskStats.Duration) / float64(kwStats.TotalAttempts)

	kwStats.LastSeen = taskStats.Timestamp

	sc.stats.KeywordStats[key] = kwStats
}

// GetStats returns a copy of current statistics
func (sc *StatsCollector) GetStats() Statistics {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	// Create a deep copy to avoid race conditions
	statsCopy := Statistics{
		StartTime:    sc.stats.StartTime,
		LastUpdate:   sc.stats.LastUpdate,
		TotalTasks:   sc.stats.TotalTasks,
		SuccessTasks: sc.stats.SuccessTasks,
		FailedTasks:  sc.stats.FailedTasks,
		TaskHistory:  make([]TaskStats, len(sc.stats.TaskHistory)),
		KeywordStats: make(map[string]KeywordStats),
	}

	copy(statsCopy.TaskHistory, sc.stats.TaskHistory)
	for k, v := range sc.stats.KeywordStats {
		statsCopy.KeywordStats[k] = v
	}

	return statsCopy
}

// GetSummary returns a summary of statistics
func (sc *StatsCollector) GetSummary() map[string]interface{} {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	successRate := 0.0
	if sc.stats.TotalTasks > 0 {
		successRate = float64(sc.stats.SuccessTasks) / float64(sc.stats.TotalTasks) * 100
	}

	return map[string]interface{}{
		"total_tasks":     sc.stats.TotalTasks,
		"success_tasks":   sc.stats.SuccessTasks,
		"failed_tasks":    sc.stats.FailedTasks,
		"success_rate":    fmt.Sprintf("%.2f%%", successRate),
		"start_time":      sc.stats.StartTime,
		"last_update":     sc.stats.LastUpdate,
		"unique_keywords": len(sc.stats.KeywordStats),
	}
}

// Save saves statistics to a JSON file
func (sc *StatsCollector) Save() error {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	// Create directory if it doesn't exist
	dir := filepath.Dir(sc.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Marshal statistics to JSON
	data, err := json.MarshalIndent(sc.stats, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal stats: %w", err)
	}

	// Write to file
	if err := os.WriteFile(sc.filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write stats file: %w", err)
	}

	return nil
}

// Load loads statistics from a JSON file
func (sc *StatsCollector) Load() error {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	// Read file
	data, err := os.ReadFile(sc.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, not an error
			return nil
		}
		return fmt.Errorf("failed to read stats file: %w", err)
	}

	// Unmarshal JSON
	var stats Statistics
	if err := json.Unmarshal(data, &stats); err != nil {
		return fmt.Errorf("failed to unmarshal stats: %w", err)
	}

	// Initialize maps if nil
	if stats.TaskHistory == nil {
		stats.TaskHistory = make([]TaskStats, 0)
	}
	if stats.KeywordStats == nil {
		stats.KeywordStats = make(map[string]KeywordStats)
	}

	sc.stats = &stats

	return nil
}

// Reset resets all statistics
func (sc *StatsCollector) Reset() {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	sc.stats = &Statistics{
		StartTime:    time.Now(),
		LastUpdate:   time.Now(),
		TaskHistory:  make([]TaskStats, 0),
		KeywordStats: make(map[string]KeywordStats),
	}
}

// GetKeywordStats returns statistics for a specific keyword-target combination
func (sc *StatsCollector) GetKeywordStats(keyword, targetURL string) (KeywordStats, bool) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	key := fmt.Sprintf("%s-%s", keyword, targetURL)
	stats, exists := sc.stats.KeywordStats[key]
	return stats, exists
}

// GetRecentTasks returns the N most recent tasks
func (sc *StatsCollector) GetRecentTasks(n int) []TaskStats {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if n <= 0 || n > len(sc.stats.TaskHistory) {
		n = len(sc.stats.TaskHistory)
	}

	// Return last N tasks
	start := len(sc.stats.TaskHistory) - n
	tasks := make([]TaskStats, n)
	copy(tasks, sc.stats.TaskHistory[start:])

	return tasks
}
