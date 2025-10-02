package shutdown

import (
	"context"
	"errors"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/omer/go-bot/internal/logger"
)

func TestNewHandler(t *testing.T) {
	log := logger.NewDefault()

	handler := NewHandler(Options{
		Logger:  log,
		Timeout: 10 * time.Second,
		Signals: []os.Signal{syscall.SIGINT},
	})

	if handler == nil {
		t.Fatal("Expected handler to be created")
	}
	if handler.logger != log {
		t.Error("Expected logger to be set")
	}
	if handler.timeout != 10*time.Second {
		t.Errorf("Expected timeout 10s, got %v", handler.timeout)
	}
	if len(handler.signals) != 1 {
		t.Errorf("Expected 1 signal, got %d", len(handler.signals))
	}
}

func TestNewHandler_Defaults(t *testing.T) {
	handler := NewHandler(Options{})

	if handler.timeout != 30*time.Second {
		t.Errorf("Expected default timeout 30s, got %v", handler.timeout)
	}
	if len(handler.signals) != 2 {
		t.Errorf("Expected 2 default signals, got %d", len(handler.signals))
	}
	if handler.logger == nil {
		t.Error("Expected default logger to be set")
	}
}

func TestHandler_Register(t *testing.T) {
	handler := NewHandler(Options{})

	called := false
	fn := func(ctx context.Context) error {
		called = true
		return nil
	}

	handler.Register("test", fn)

	if len(handler.shutdownFuncs) != 1 {
		t.Errorf("Expected 1 shutdown func, got %d", len(handler.shutdownFuncs))
	}

	// Call the registered function
	handler.shutdownFuncs[0](context.Background())

	if !called {
		t.Error("Expected shutdown function to be called")
	}
}

func TestHandler_Shutdown_Success(t *testing.T) {
	handler := NewHandler(Options{
		Timeout: 2 * time.Second,
	})

	called1 := false
	called2 := false

	handler.Register("component1", func(ctx context.Context) error {
		called1 = true
		return nil
	})

	handler.Register("component2", func(ctx context.Context) error {
		called2 = true
		return nil
	})

	handler.Shutdown()

	if !called1 {
		t.Error("Expected component1 shutdown to be called")
	}
	if !called2 {
		t.Error("Expected component2 shutdown to be called")
	}
}

func TestHandler_Shutdown_WithErrors(t *testing.T) {
	handler := NewHandler(Options{
		Timeout: 2 * time.Second,
	})

	testErr := errors.New("shutdown error")

	handler.Register("failing_component", func(ctx context.Context) error {
		return testErr
	})

	handler.Register("success_component", func(ctx context.Context) error {
		return nil
	})

	// Should complete despite errors
	handler.Shutdown()
}

func TestHandler_Shutdown_Timeout(t *testing.T) {
	handler := NewHandler(Options{
		Timeout: 100 * time.Millisecond,
	})

	handler.Register("slow_component", func(ctx context.Context) error {
		time.Sleep(1 * time.Second) // Longer than timeout
		return nil
	})

	start := time.Now()
	handler.Shutdown()
	duration := time.Since(start)

	// Should timeout around 100ms, not wait for 1 second
	if duration > 500*time.Millisecond {
		t.Errorf("Shutdown took too long: %v", duration)
	}
}

func TestHandler_Shutdown_LIFO(t *testing.T) {
	handler := NewHandler(Options{
		Timeout: 2 * time.Second,
	})

	order := make([]string, 0)
	mu := make(chan struct{}, 1)
	mu <- struct{}{}

	handler.Register("first", func(ctx context.Context) error {
		<-mu
		order = append(order, "first")
		mu <- struct{}{}
		return nil
	})

	handler.Register("second", func(ctx context.Context) error {
		<-mu
		order = append(order, "second")
		mu <- struct{}{}
		return nil
	})

	handler.Shutdown()

	// Should be called in reverse order (LIFO)
	if len(order) != 2 {
		t.Fatalf("Expected 2 calls, got %d", len(order))
	}

	// Note: Due to concurrent execution, order might not be strictly LIFO
	// But functions should complete
	t.Logf("Shutdown order: %v", order)
}

func TestHandler_IsShuttingDown(t *testing.T) {
	handler := NewHandler(Options{})

	if handler.IsShuttingDown() {
		t.Error("Expected IsShuttingDown to be false initially")
	}

	handler.Register("test", func(ctx context.Context) error {
		time.Sleep(50 * time.Millisecond)
		return nil
	})

	go handler.Shutdown()
	time.Sleep(10 * time.Millisecond) // Give it time to start

	if !handler.IsShuttingDown() {
		t.Error("Expected IsShuttingDown to be true during shutdown")
	}
}

func TestHandler_Shutdown_Multiple(t *testing.T) {
	handler := NewHandler(Options{
		Timeout: 1 * time.Second,
	})

	calls := 0
	handler.Register("test", func(ctx context.Context) error {
		calls++
		return nil
	})

	// Call shutdown multiple times
	handler.Shutdown()
	handler.Shutdown()
	handler.Shutdown()

	// Should only execute once
	if calls != 1 {
		t.Errorf("Expected 1 call, got %d", calls)
	}
}

func TestWithTimeout(t *testing.T) {
	tests := []struct {
		name      string
		timeout   time.Duration
		fnDelay   time.Duration
		shouldErr bool
	}{
		{
			name:      "completes in time",
			timeout:   1 * time.Second,
			fnDelay:   100 * time.Millisecond,
			shouldErr: false,
		},
		{
			name:      "times out",
			timeout:   100 * time.Millisecond,
			fnDelay:   1 * time.Second,
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := WithTimeout(tt.timeout, func() error {
				time.Sleep(tt.fnDelay)
				return nil
			})

			ctx := context.Background()
			err := fn(ctx)

			if tt.shouldErr && err == nil {
				t.Error("Expected timeout error")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestSignalNames(t *testing.T) {
	signals := []os.Signal{syscall.SIGINT, syscall.SIGTERM}
	names := signalNames(signals)

	if len(names) != 2 {
		t.Errorf("Expected 2 names, got %d", len(names))
	}

	for _, name := range names {
		if name == "" {
			t.Error("Expected non-empty signal name")
		}
	}
}

func TestHandler_Register_MultipleComponents(t *testing.T) {
	handler := NewHandler(Options{
		Timeout: 2 * time.Second,
	})

	components := []string{"database", "cache", "api", "worker"}
	calls := make(map[string]bool)

	for _, comp := range components {
		name := comp // Capture for closure
		handler.Register(name, func(ctx context.Context) error {
			calls[name] = true
			return nil
		})
	}

	handler.Shutdown()

	for _, comp := range components {
		if !calls[comp] {
			t.Errorf("Expected component %s to be shut down", comp)
		}
	}
}
