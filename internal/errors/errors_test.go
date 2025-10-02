package errors

import (
	"errors"
	"fmt"
	"testing"
)

func TestNew(t *testing.T) {
	err := New(ErrorTypeProxy, "proxy connection failed")

	if err.Type != ErrorTypeProxy {
		t.Errorf("Expected type %v, got %v", ErrorTypeProxy, err.Type)
	}
	if err.Message != "proxy connection failed" {
		t.Errorf("Expected message 'proxy connection failed', got '%s'", err.Message)
	}
	if err.Err != nil {
		t.Errorf("Expected no underlying error, got %v", err.Err)
	}
}

func TestWrap(t *testing.T) {
	originalErr := fmt.Errorf("network timeout")
	wrappedErr := Wrap(originalErr, ErrorTypeNetwork, "failed to connect")

	if wrappedErr.Type != ErrorTypeNetwork {
		t.Errorf("Expected type %v, got %v", ErrorTypeNetwork, wrappedErr.Type)
	}
	if wrappedErr.Err != originalErr {
		t.Errorf("Expected wrapped error to contain original error")
	}
}

func TestAppError_Error(t *testing.T) {
	tests := []struct {
		name     string
		appErr   *AppError
		expected string
	}{
		{
			name: "error without underlying error",
			appErr: &AppError{
				Type:    ErrorTypeProxy,
				Message: "proxy failed",
			},
			expected: "[proxy] proxy failed",
		},
		{
			name: "error with underlying error",
			appErr: &AppError{
				Type:    ErrorTypeTimeout,
				Message: "operation timed out",
				Err:     fmt.Errorf("context deadline exceeded"),
			},
			expected: "[timeout] operation timed out: context deadline exceeded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.appErr.Error()
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestAppError_Unwrap(t *testing.T) {
	originalErr := fmt.Errorf("original error")
	wrappedErr := Wrap(originalErr, ErrorTypeNetwork, "wrapped")

	unwrapped := wrappedErr.Unwrap()
	if unwrapped != originalErr {
		t.Errorf("Expected unwrapped error to be original error")
	}
}

func TestAppError_WithContext(t *testing.T) {
	err := New(ErrorTypeProxy, "proxy error")
	err.WithContext("proxy_url", "http://proxy.example.com")
	err.WithContext("attempt", 3)

	if err.Context["proxy_url"] != "http://proxy.example.com" {
		t.Errorf("Expected context proxy_url to be set")
	}
	if err.Context["attempt"] != 3 {
		t.Errorf("Expected context attempt to be 3")
	}
}

func TestIs(t *testing.T) {
	proxyErr := New(ErrorTypeProxy, "proxy failed")

	if !Is(proxyErr, ErrorTypeProxy) {
		t.Error("Expected Is to return true for matching type")
	}
	if Is(proxyErr, ErrorTypeBrowser) {
		t.Error("Expected Is to return false for non-matching type")
	}

	// Test with wrapped error
	wrappedErr := fmt.Errorf("wrapped: %w", proxyErr)
	if !Is(wrappedErr, ErrorTypeProxy) {
		t.Error("Expected Is to work with wrapped errors")
	}
}

func TestIsRetryable(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		retryable bool
	}{
		{
			name:      "timeout error",
			err:       New(ErrorTypeTimeout, "timeout"),
			retryable: true,
		},
		{
			name:      "network error",
			err:       New(ErrorTypeNetwork, "network failed"),
			retryable: true,
		},
		{
			name:      "proxy error",
			err:       New(ErrorTypeProxy, "proxy failed"),
			retryable: true,
		},
		{
			name:      "captcha error",
			err:       New(ErrorTypeCaptcha, "captcha detected"),
			retryable: false,
		},
		{
			name:      "browser error",
			err:       New(ErrorTypeBrowser, "browser crashed"),
			retryable: false,
		},
		{
			name:      "standard error",
			err:       fmt.Errorf("standard error"),
			retryable: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsRetryable(tt.err)
			if result != tt.retryable {
				t.Errorf("Expected retryable=%v, got %v", tt.retryable, result)
			}
		})
	}
}

func TestGetType(t *testing.T) {
	err := New(ErrorTypeProxy, "proxy error")
	errType := GetType(err)

	if errType != ErrorTypeProxy {
		t.Errorf("Expected type %v, got %v", ErrorTypeProxy, errType)
	}

	// Test with standard error
	stdErr := fmt.Errorf("standard error")
	errType = GetType(stdErr)
	if errType != ErrorTypeUnknown {
		t.Errorf("Expected type %v for standard error, got %v", ErrorTypeUnknown, errType)
	}
}

func TestErrorType_String(t *testing.T) {
	tests := []struct {
		errType  ErrorType
		expected string
	}{
		{ErrorTypeProxy, "proxy"},
		{ErrorTypeBrowser, "browser"},
		{ErrorTypeSelector, "selector"},
		{ErrorTypeTimeout, "timeout"},
		{ErrorTypeCaptcha, "captcha"},
		{ErrorTypeNetwork, "network"},
		{ErrorTypeConfig, "config"},
		{ErrorTypeValidation, "validation"},
		{ErrorTypeUnknown, "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.errType.String()
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestCommonErrorConstructors(t *testing.T) {
	originalErr := fmt.Errorf("original")

	tests := []struct {
		name     string
		errFunc  func() *AppError
		expected ErrorType
	}{
		{
			name:     "NewProxyError",
			errFunc:  func() *AppError { return NewProxyError("test", originalErr) },
			expected: ErrorTypeProxy,
		},
		{
			name:     "NewBrowserError",
			errFunc:  func() *AppError { return NewBrowserError("test", originalErr) },
			expected: ErrorTypeBrowser,
		},
		{
			name:     "NewSelectorError",
			errFunc:  func() *AppError { return NewSelectorError("test", originalErr) },
			expected: ErrorTypeSelector,
		},
		{
			name:     "NewTimeoutError",
			errFunc:  func() *AppError { return NewTimeoutError("test", originalErr) },
			expected: ErrorTypeTimeout,
		},
		{
			name:     "NewCaptchaError",
			errFunc:  func() *AppError { return NewCaptchaError("test") },
			expected: ErrorTypeCaptcha,
		},
		{
			name:     "NewNetworkError",
			errFunc:  func() *AppError { return NewNetworkError("test", originalErr) },
			expected: ErrorTypeNetwork,
		},
		{
			name:     "NewConfigError",
			errFunc:  func() *AppError { return NewConfigError("test", originalErr) },
			expected: ErrorTypeConfig,
		},
		{
			name:     "NewValidationError",
			errFunc:  func() *AppError { return NewValidationError("test") },
			expected: ErrorTypeValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.errFunc()
			if err.Type != tt.expected {
				t.Errorf("Expected type %v, got %v", tt.expected, err.Type)
			}
		})
	}
}

func TestAppError_ErrorsAs(t *testing.T) {
	originalErr := New(ErrorTypeProxy, "proxy failed")
	wrappedErr := fmt.Errorf("wrapped: %w", originalErr)

	var appErr *AppError
	if !errors.As(wrappedErr, &appErr) {
		t.Error("Expected errors.As to unwrap AppError")
	}
	if appErr.Type != ErrorTypeProxy {
		t.Errorf("Expected type %v, got %v", ErrorTypeProxy, appErr.Type)
	}
}

func TestAppError_WithContext_Chaining(t *testing.T) {
	err := New(ErrorTypeNetwork, "network error").
		WithContext("url", "http://example.com").
		WithContext("retry", 3).
		WithContext("timeout", "30s")

	if len(err.Context) != 3 {
		t.Errorf("Expected 3 context entries, got %d", len(err.Context))
	}
}
