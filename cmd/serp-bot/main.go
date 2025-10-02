// Package main provides the CLI entry point for the SERP bot application.
package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/omer/go-bot/internal/config"
	"github.com/omer/go-bot/internal/logger"
	"github.com/omer/go-bot/internal/proxy"
	"github.com/omer/go-bot/internal/stats"
	"github.com/omer/go-bot/internal/task"
	"github.com/spf13/cobra"
)

var (
	// Version information
	version   = "1.0.0"
	buildTime = "unknown"

	// Command flags
	configFile  string
	logLevel    string
	headless    bool
	workers     int
	enableStats bool
)

func main() {
	rootCmd := &cobra.Command{
		Use:     "serp-bot",
		Short:   "SERP Bot - Automated search engine ranking bot",
		Version: fmt.Sprintf("%s (built: %s)", version, buildTime),
		Long: `SERP Bot is an automated tool for searching keywords on search engines,
finding target URLs, and clicking on them to improve search rankings.

Features:
  - Automated Google search
  - Target URL finding and clicking
  - Proxy rotation support
  - Concurrent task execution
  - Statistics tracking
  - Human-like behavior`,
	}

	// Start command
	startCmd := &cobra.Command{
		Use:   "start",
		Short: "Start the SERP bot",
		Long:  "Start the SERP bot with the specified configuration",
		RunE:  runStart,
	}

	// Add flags
	startCmd.Flags().StringVarP(&configFile, "config", "c", "configs/config.json", "Path to configuration file")
	startCmd.Flags().StringVarP(&logLevel, "log-level", "l", "", "Log level (debug, info, warn, error)")
	startCmd.Flags().BoolVar(&headless, "headless", true, "Run browser in headless mode")
	startCmd.Flags().IntVarP(&workers, "workers", "w", 0, "Number of worker goroutines (0 = use config value)")
	startCmd.Flags().BoolVar(&enableStats, "stats", true, "Enable statistics collection")

	// Stats command
	statsCmd := &cobra.Command{
		Use:   "stats",
		Short: "Show statistics",
		Long:  "Show collected statistics from previous runs",
		RunE:  runStats,
	}
	statsCmd.Flags().IntVarP(&workers, "recent", "n", 10, "Number of recent tasks to show")

	// Add commands
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(statsCmd)

	// Execute
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// runStart executes the start command
func runStart(cmd *cobra.Command, args []string) error {
	fmt.Printf("🚀 SERP Bot v%s\n\n", version)

	// Load configuration
	fmt.Printf("📋 Loading configuration from: %s\n", configFile)
	cfg, err := config.Load(configFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Override with flags
	if logLevel != "" {
		cfg.LogLevel = logLevel
	}
	if workers > 0 {
		cfg.Workers = workers
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Initialize logger
	logConfig := logger.Config{
		Level:      logger.LogLevel(cfg.LogLevel),
		LogFile:    cfg.LogFile,
		EnableFile: cfg.LogFile != "",
	}
	log, err := logger.New(logConfig)
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	defer log.Close()

	log.Info("SERP Bot starting", map[string]interface{}{
		"version": version,
		"workers": cfg.Workers,
	})

	// Initialize proxy pool
	var proxyPool *proxy.ProxyPool
	if len(cfg.Proxies) > 0 {
		log.Info("Initializing proxy pool", map[string]interface{}{
			"proxies": len(cfg.Proxies),
		})
		proxyPool, err = proxy.NewProxyPool(cfg.Proxies, proxy.RotationStrategyRoundRobin)
		if err != nil {
			return fmt.Errorf("failed to initialize proxy pool: %w", err)
		}
	} else {
		log.Warn("No proxies configured - running without proxy", nil)
	}

	// Initialize statistics collector
	var statsCollector *stats.StatsCollector
	if enableStats {
		statsCollector = stats.NewStatsCollector("data/stats.json")
		if err := statsCollector.Load(); err != nil {
			log.Warn("Failed to load previous stats", map[string]interface{}{
				"error": err,
			})
		}
	}

	// Initialize worker pool
	workerPool := task.NewWorkerPool(task.WorkerPoolConfig{
		Workers:   cfg.Workers,
		QueueSize: cfg.Workers * 2,
		ProxyPool: proxyPool,
		Logger:    log,
	})

	// Start worker pool
	if err := workerPool.Start(); err != nil {
		return fmt.Errorf("failed to start worker pool: %w", err)
	}

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Result processor goroutine
	go func() {
		for result := range workerPool.GetResults() {
			if result.Success {
				log.Info("Task completed successfully", map[string]interface{}{
					"task_id":  result.Task.ID,
					"keyword":  result.Task.Keyword,
					"position": result.Position,
					"duration": fmt.Sprintf("%.2fs", result.Duration.Seconds()),
				})
			} else {
				log.Error("Task failed", map[string]interface{}{
					"task_id": result.Task.ID,
					"keyword": result.Task.Keyword,
					"error":   result.Error,
				})
			}

			// Record statistics
			if enableStats && statsCollector != nil {
				taskStats := stats.TaskStats{
					TaskID:     result.Task.ID,
					Keyword:    result.Task.Keyword,
					TargetURL:  result.Task.TargetURL,
					Success:    result.Success,
					Position:   result.Position,
					PageNumber: result.PageNumber,
					Duration:   float64(result.Duration.Milliseconds()),
					ProxyUsed:  result.Task.ProxyURL,
					Timestamp:  time.Now(),
				}
				if result.Error != nil {
					taskStats.Error = result.Error.Error()
				}
				statsCollector.RecordTask(taskStats)
			}
		}
	}()

	// Submit tasks
	log.Info("Submitting tasks", map[string]interface{}{
		"keywords": len(cfg.Keywords),
	})

	for _, keyword := range cfg.Keywords {
		t, err := task.NewTask(task.TaskConfig{
			Keyword:   keyword.Term,
			TargetURL: keyword.TargetURL,
		})
		if err != nil {
			log.Error("Failed to create task", map[string]interface{}{
				"error": err,
			})
			continue
		}

		if err := workerPool.Submit(t); err != nil {
			log.Error("Failed to submit task", map[string]interface{}{
				"error": err,
			})
		}
	}

	fmt.Printf("\n✅ %d tasks submitted\n", len(cfg.Keywords))
	fmt.Println("⏳ Waiting for tasks to complete... (Ctrl+C to stop)")

	// Wait for shutdown signal
	<-sigChan
	fmt.Println("\n\n🛑 Shutdown signal received...")

	// Stop worker pool
	log.Info("Stopping worker pool", nil)
	if err := workerPool.Stop(); err != nil {
		log.Error("Error stopping worker pool", map[string]interface{}{
			"error": err,
		})
	}

	// Save statistics
	if enableStats && statsCollector != nil {
		log.Info("Saving statistics", nil)
		if err := statsCollector.Save(); err != nil {
			log.Error("Failed to save statistics", map[string]interface{}{
				"error": err,
			})
		} else {
			summary := statsCollector.GetSummary()
			fmt.Println("\n📊 Statistics:")
			fmt.Printf("  Total tasks: %d\n", summary["total_tasks"])
			fmt.Printf("  Success: %d\n", summary["success_tasks"])
			fmt.Printf("  Failed: %d\n", summary["failed_tasks"])
			fmt.Printf("  Success rate: %s\n", summary["success_rate"])
		}
	}

	fmt.Println("\n✨ Shutdown complete. Goodbye!")
	return nil
}

// runStats executes the stats command
func runStats(cmd *cobra.Command, args []string) error {
	recentCount, _ := cmd.Flags().GetInt("recent")

	collector := stats.NewStatsCollector("data/stats.json")
	if err := collector.Load(); err != nil {
		return fmt.Errorf("failed to load stats: %w", err)
	}

	summary := collector.GetSummary()
	fmt.Println("📊 SERP Bot Statistics")
	fmt.Println("═══════════════════════")
	fmt.Printf("Total tasks: %d\n", summary["total_tasks"])
	fmt.Printf("Success: %d\n", summary["success_tasks"])
	fmt.Printf("Failed: %d\n", summary["failed_tasks"])
	fmt.Printf("Success rate: %s\n", summary["success_rate"])
	fmt.Printf("Unique keywords: %d\n", summary["unique_keywords"])
	fmt.Printf("Start time: %v\n", summary["start_time"])
	fmt.Printf("Last update: %v\n", summary["last_update"])

	// Show recent tasks
	recentTasks := collector.GetRecentTasks(recentCount)
	if len(recentTasks) > 0 {
		fmt.Printf("\n📝 Recent %d tasks:\n", len(recentTasks))
		fmt.Println("─────────────────────────────")
		for i, t := range recentTasks {
			status := "✅"
			if !t.Success {
				status = "❌"
			}
			fmt.Printf("%d. %s [%s] %s -> Position: %d (%.2fs)\n",
				i+1, status, t.Keyword, t.TargetURL, t.Position, t.Duration/1000)
		}
	}

	return nil
}
