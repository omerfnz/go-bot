// Package shutdown provides graceful shutdown functionality
package shutdown

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/omer/go-bot/internal/logger"
)

// Handler manages graceful shutdown of the application
type Handler struct {
	logger        *logger.Logger
	shutdownFuncs []ShutdownFunc
	timeout       time.Duration
	signals       []os.Signal
	mu            sync.Mutex
	shuttingDown  bool
}

// ShutdownFunc is a function that performs cleanup during shutdown
type ShutdownFunc func(ctx context.Context) error

// Options configures the shutdown handler
type Options struct {
	Logger  *logger.Logger // Logger for shutdown messages
	Timeout time.Duration  // Maximum time to wait for cleanup (default: 30s)
	Signals []os.Signal    // Signals to listen for (default: SIGINT, SIGTERM)
}

// NewHandler creates a new shutdown handler
func NewHandler(opts Options) *Handler {
	if opts.Timeout == 0 {
		opts.Timeout = 30 * time.Second
	}

	if len(opts.Signals) == 0 {
		opts.Signals = []os.Signal{syscall.SIGINT, syscall.SIGTERM}
	}

	if opts.Logger == nil {
		opts.Logger = logger.NewDefault()
	}

	return &Handler{
		logger:        opts.Logger,
		shutdownFuncs: make([]ShutdownFunc, 0),
		timeout:       opts.Timeout,
		signals:       opts.Signals,
		shuttingDown:  false,
	}
}

// Register registers a cleanup function to be called during shutdown
// Functions are called in reverse order of registration (LIFO)
func (h *Handler) Register(name string, fn ShutdownFunc) {
	h.mu.Lock()
	defer h.mu.Unlock()

	wrappedFn := func(ctx context.Context) error {
		h.logger.WithField("component", name).Info("Shutting down component")
		start := time.Now()

		err := fn(ctx)

		duration := time.Since(start)
		if err != nil {
			h.logger.WithFields(map[string]interface{}{
				"component": name,
				"duration":  duration,
				"error":     err,
			}).Error("Component shutdown failed")
		} else {
			h.logger.WithFields(map[string]interface{}{
				"component": name,
				"duration":  duration,
			}).Info("Component shutdown completed")
		}

		return err
	}

	h.shutdownFuncs = append(h.shutdownFuncs, wrappedFn)
}

// Listen starts listening for shutdown signals
// Blocks until a signal is received
func (h *Handler) Listen() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, h.signals...)

	h.logger.Info("Shutdown handler listening for signals", map[string]interface{}{
		"signals": signalNames(h.signals),
		"timeout": h.timeout,
	})

	sig := <-sigChan
	h.logger.WithField("signal", sig).Info("Shutdown signal received")

	h.Shutdown()
}

// Shutdown performs graceful shutdown
func (h *Handler) Shutdown() {
	h.mu.Lock()
	if h.shuttingDown {
		h.mu.Unlock()
		return
	}
	h.shuttingDown = true
	h.mu.Unlock()

	h.logger.Info("Starting graceful shutdown", map[string]interface{}{
		"timeout":    h.timeout,
		"components": len(h.shutdownFuncs),
	})

	ctx, cancel := context.WithTimeout(context.Background(), h.timeout)
	defer cancel()

	// Call shutdown functions in reverse order (LIFO)
	var wg sync.WaitGroup
	errors := make([]error, 0)
	errorMu := sync.Mutex{}

	for i := len(h.shutdownFuncs) - 1; i >= 0; i-- {
		wg.Add(1)
		go func(fn ShutdownFunc) {
			defer wg.Done()
			if err := fn(ctx); err != nil {
				errorMu.Lock()
				errors = append(errors, err)
				errorMu.Unlock()
			}
		}(h.shutdownFuncs[i])
	}

	// Wait for all shutdowns to complete or timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		if len(errors) > 0 {
			h.logger.Warn("Graceful shutdown completed with errors", map[string]interface{}{
				"error_count": len(errors),
			})
		} else {
			h.logger.Info("Graceful shutdown completed successfully")
		}
	case <-ctx.Done():
		h.logger.Error("Graceful shutdown timed out", map[string]interface{}{
			"timeout": h.timeout,
		})
	}
}

// IsShuttingDown returns true if shutdown is in progress
func (h *Handler) IsShuttingDown() bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.shuttingDown
}

// signalNames returns human-readable names for signals
func signalNames(signals []os.Signal) []string {
	names := make([]string, len(signals))
	for i, sig := range signals {
		names[i] = fmt.Sprintf("%v", sig)
	}
	return names
}

// WithTimeout wraps a function with timeout
func WithTimeout(timeout time.Duration, fn func() error) ShutdownFunc {
	return func(ctx context.Context) error {
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		errChan := make(chan error, 1)
		go func() {
			errChan <- fn()
		}()

		select {
		case err := <-errChan:
			return err
		case <-ctx.Done():
			return fmt.Errorf("shutdown function timed out: %w", ctx.Err())
		}
	}
}
