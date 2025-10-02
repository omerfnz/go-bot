package stats

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ===== StatsCollector creation tests =====

func TestNewStatsCollector(t *testing.T) {
	collector := NewStatsCollector("test/stats.json")

	assert.NotNil(t, collector)
	assert.NotNil(t, collector.stats)
	assert.Equal(t, "test/stats.json", collector.filePath)
	assert.Equal(t, 0, collector.stats.TotalTasks)
	assert.NotNil(t, collector.stats.TaskHistory)
	assert.NotNil(t, collector.stats.KeywordStats)
}

// ===== RecordTask tests =====

func TestRecordTask_Success(t *testing.T) {
	collector := NewStatsCollector("test/stats.json")

	taskStats := TaskStats{
		TaskID:     "task-1",
		Keyword:    "golang",
		TargetURL:  "example.com",
		Success:    true,
		Position:   3,
		PageNumber: 1,
		Duration:   1500.5,
		ProxyUsed:  "http://proxy:8080",
	}

	collector.RecordTask(taskStats)

	stats := collector.GetStats()
	assert.Equal(t, 1, stats.TotalTasks)
	assert.Equal(t, 1, stats.SuccessTasks)
	assert.Equal(t, 0, stats.FailedTasks)
	assert.Len(t, stats.TaskHistory, 1)
	assert.Equal(t, "task-1", stats.TaskHistory[0].TaskID)
}

func TestRecordTask_Failure(t *testing.T) {
	collector := NewStatsCollector("test/stats.json")

	taskStats := TaskStats{
		TaskID:    "task-2",
		Keyword:   "golang",
		TargetURL: "example.com",
		Success:   false,
		Duration:  500.0,
		Error:     "target not found",
	}

	collector.RecordTask(taskStats)

	stats := collector.GetStats()
	assert.Equal(t, 1, stats.TotalTasks)
	assert.Equal(t, 0, stats.SuccessTasks)
	assert.Equal(t, 1, stats.FailedTasks)
	assert.Equal(t, "target not found", stats.TaskHistory[0].Error)
}

func TestRecordTask_AutoTimestamp(t *testing.T) {
	collector := NewStatsCollector("test/stats.json")

	taskStats := TaskStats{
		TaskID:    "task-3",
		Keyword:   "test",
		TargetURL: "example.com",
		Success:   true,
		// No timestamp set
	}

	collector.RecordTask(taskStats)

	stats := collector.GetStats()
	assert.False(t, stats.TaskHistory[0].Timestamp.IsZero())
}

func TestRecordTask_MultipleTasks(t *testing.T) {
	collector := NewStatsCollector("test/stats.json")

	for i := 0; i < 5; i++ {
		taskStats := TaskStats{
			TaskID:    "task-" + string(rune(i)),
			Keyword:   "golang",
			TargetURL: "example.com",
			Success:   i%2 == 0, // Alternate success/failure
			Position:  i + 1,
		}
		collector.RecordTask(taskStats)
	}

	stats := collector.GetStats()
	assert.Equal(t, 5, stats.TotalTasks)
	assert.Equal(t, 3, stats.SuccessTasks)
	assert.Equal(t, 2, stats.FailedTasks)
	assert.Len(t, stats.TaskHistory, 5)
}

// ===== KeywordStats tests =====

func TestKeywordStats_SingleKeyword(t *testing.T) {
	collector := NewStatsCollector("test/stats.json")

	// Record multiple tasks for same keyword
	for i := 1; i <= 3; i++ {
		taskStats := TaskStats{
			Keyword:   "golang",
			TargetURL: "example.com",
			Success:   true,
			Position:  i * 2, // 2, 4, 6
			Duration:  float64(i * 100),
		}
		collector.RecordTask(taskStats)
	}

	kwStats, exists := collector.GetKeywordStats("golang", "example.com")
	require.True(t, exists)

	assert.Equal(t, "golang", kwStats.Keyword)
	assert.Equal(t, "example.com", kwStats.TargetURL)
	assert.Equal(t, 3, kwStats.TotalAttempts)
	assert.Equal(t, 3, kwStats.SuccessCount)
	assert.Equal(t, 0, kwStats.FailureCount)
	assert.Equal(t, 2, kwStats.BestPosition)  // Best is lowest
	assert.Equal(t, 6, kwStats.WorstPosition) // Worst is highest
	assert.InDelta(t, 4.0, kwStats.AvgPosition, 0.1)
	assert.InDelta(t, 200.0, kwStats.AvgDuration, 0.1)
}

func TestKeywordStats_MultipleKeywords(t *testing.T) {
	collector := NewStatsCollector("test/stats.json")

	// Record tasks for different keywords
	keywords := []string{"golang", "python", "rust"}
	for _, keyword := range keywords {
		taskStats := TaskStats{
			Keyword:   keyword,
			TargetURL: "example.com",
			Success:   true,
			Position:  5,
		}
		collector.RecordTask(taskStats)
	}

	stats := collector.GetStats()
	assert.Len(t, stats.KeywordStats, 3)

	for _, keyword := range keywords {
		kwStats, exists := collector.GetKeywordStats(keyword, "example.com")
		assert.True(t, exists)
		assert.Equal(t, keyword, kwStats.Keyword)
	}
}

func TestKeywordStats_NonExistent(t *testing.T) {
	collector := NewStatsCollector("test/stats.json")

	_, exists := collector.GetKeywordStats("nonexistent", "example.com")
	assert.False(t, exists)
}

// ===== GetSummary tests =====

func TestGetSummary_Empty(t *testing.T) {
	collector := NewStatsCollector("test/stats.json")

	summary := collector.GetSummary()

	assert.Equal(t, 0, summary["total_tasks"])
	assert.Equal(t, 0, summary["success_tasks"])
	assert.Equal(t, 0, summary["failed_tasks"])
	assert.Equal(t, "0.00%", summary["success_rate"])
	assert.Equal(t, 0, summary["unique_keywords"])
}

func TestGetSummary_WithData(t *testing.T) {
	collector := NewStatsCollector("test/stats.json")

	// Add some tasks
	for i := 0; i < 10; i++ {
		taskStats := TaskStats{
			Keyword:   "golang",
			TargetURL: "example.com",
			Success:   i < 7, // 7 success, 3 failures
		}
		collector.RecordTask(taskStats)
	}

	summary := collector.GetSummary()

	assert.Equal(t, 10, summary["total_tasks"])
	assert.Equal(t, 7, summary["success_tasks"])
	assert.Equal(t, 3, summary["failed_tasks"])
	assert.Equal(t, "70.00%", summary["success_rate"])
	assert.Equal(t, 1, summary["unique_keywords"])
}

// ===== Save/Load tests =====

func TestSaveLoad_Success(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "stats.json")
	collector := NewStatsCollector(filePath)

	// Add some data
	taskStats := TaskStats{
		TaskID:     "task-1",
		Keyword:    "golang",
		TargetURL:  "example.com",
		Success:    true,
		Position:   3,
		PageNumber: 1,
		Duration:   1500.5,
	}
	collector.RecordTask(taskStats)

	// Save
	err := collector.Save()
	require.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(filePath)
	assert.NoError(t, err)

	// Load into new collector
	collector2 := NewStatsCollector(filePath)
	err = collector2.Load()
	require.NoError(t, err)

	// Verify data
	stats := collector2.GetStats()
	assert.Equal(t, 1, stats.TotalTasks)
	assert.Equal(t, 1, stats.SuccessTasks)
	assert.Len(t, stats.TaskHistory, 1)
	assert.Equal(t, "task-1", stats.TaskHistory[0].TaskID)
}

func TestLoad_NonExistentFile(t *testing.T) {
	collector := NewStatsCollector("nonexistent/stats.json")

	err := collector.Load()
	assert.NoError(t, err) // Should not error for non-existent file
}

func TestSave_CreatesDirectory(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "subdir", "stats.json")
	collector := NewStatsCollector(filePath)

	err := collector.Save()
	require.NoError(t, err)

	// Verify directory was created
	_, err = os.Stat(filepath.Dir(filePath))
	assert.NoError(t, err)
}

// ===== Reset tests =====

func TestReset(t *testing.T) {
	collector := NewStatsCollector("test/stats.json")

	// Add some data
	for i := 0; i < 5; i++ {
		taskStats := TaskStats{
			Keyword:   "golang",
			TargetURL: "example.com",
			Success:   true,
		}
		collector.RecordTask(taskStats)
	}

	assert.Equal(t, 5, collector.stats.TotalTasks)

	// Reset
	collector.Reset()

	stats := collector.GetStats()
	assert.Equal(t, 0, stats.TotalTasks)
	assert.Equal(t, 0, stats.SuccessTasks)
	assert.Equal(t, 0, stats.FailedTasks)
	assert.Len(t, stats.TaskHistory, 0)
	assert.Len(t, stats.KeywordStats, 0)
}

// ===== GetRecentTasks tests =====

func TestGetRecentTasks_Empty(t *testing.T) {
	collector := NewStatsCollector("test/stats.json")

	recent := collector.GetRecentTasks(5)
	assert.Len(t, recent, 0)
}

func TestGetRecentTasks_LessThanN(t *testing.T) {
	collector := NewStatsCollector("test/stats.json")

	// Add 3 tasks
	for i := 0; i < 3; i++ {
		taskStats := TaskStats{
			TaskID:    "task-" + string(rune(i)),
			Keyword:   "golang",
			TargetURL: "example.com",
			Success:   true,
		}
		collector.RecordTask(taskStats)
	}

	recent := collector.GetRecentTasks(5) // Request 5 but only 3 exist
	assert.Len(t, recent, 3)
}

func TestGetRecentTasks_ExactlyN(t *testing.T) {
	collector := NewStatsCollector("test/stats.json")

	// Add 5 tasks
	for i := 0; i < 5; i++ {
		taskStats := TaskStats{
			TaskID:    "task-" + string(rune(i)),
			Keyword:   "golang",
			TargetURL: "example.com",
			Success:   true,
		}
		collector.RecordTask(taskStats)
		time.Sleep(1 * time.Millisecond) // Ensure different timestamps
	}

	recent := collector.GetRecentTasks(3)
	assert.Len(t, recent, 3)

	// Should return last 3 tasks
	assert.Equal(t, "task-"+string(rune(2)), recent[0].TaskID)
	assert.Equal(t, "task-"+string(rune(3)), recent[1].TaskID)
	assert.Equal(t, "task-"+string(rune(4)), recent[2].TaskID)
}

func TestGetRecentTasks_Zero(t *testing.T) {
	collector := NewStatsCollector("test/stats.json")

	// Add tasks
	for i := 0; i < 3; i++ {
		taskStats := TaskStats{
			TaskID:  "task-" + string(rune(i)),
			Success: true,
		}
		collector.RecordTask(taskStats)
	}

	recent := collector.GetRecentTasks(0) // Request 0
	assert.Len(t, recent, 3)              // Should return all
}

// ===== Concurrency tests =====

func TestConcurrentRecordTask(t *testing.T) {
	collector := NewStatsCollector("test/stats.json")

	const numGoroutines = 10
	const tasksPerGoroutine = 5

	done := make(chan bool)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			for j := 0; j < tasksPerGoroutine; j++ {
				taskStats := TaskStats{
					Keyword:   "golang",
					TargetURL: "example.com",
					Success:   true,
				}
				collector.RecordTask(taskStats)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	stats := collector.GetStats()
	assert.Equal(t, numGoroutines*tasksPerGoroutine, stats.TotalTasks)
}

// ===== Edge cases =====

func TestTaskStats_ZeroPosition(t *testing.T) {
	collector := NewStatsCollector("test/stats.json")

	taskStats := TaskStats{
		Keyword:   "golang",
		TargetURL: "example.com",
		Success:   true,
		Position:  0, // Zero position should not affect BestPosition
	}
	collector.RecordTask(taskStats)

	kwStats, exists := collector.GetKeywordStats("golang", "example.com")
	require.True(t, exists)

	// BestPosition should remain at initial value
	assert.Equal(t, 999999, kwStats.BestPosition)
}

func TestTaskStats_EmptyKeyword(t *testing.T) {
	collector := NewStatsCollector("test/stats.json")

	taskStats := TaskStats{
		Keyword:   "", // Empty keyword
		TargetURL: "example.com",
		Success:   true,
	}
	collector.RecordTask(taskStats)

	stats := collector.GetStats()
	assert.Equal(t, 1, stats.TotalTasks)
	assert.Len(t, stats.KeywordStats, 0) // Should not create keyword stats
}
