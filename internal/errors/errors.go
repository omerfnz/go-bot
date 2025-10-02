// Package errors provides custom error types and error handling utilities
package errors

import (
	"errors"
	"fmt"
)

// ErrorType represents the category of error
type ErrorType int

const (
	// ErrorTypeUnknown represents an unknown error type
	ErrorTypeUnknown ErrorType = iota

	// ErrorTypeProxy represents proxy-related errors
	ErrorTypeProxy

	// ErrorTypeBrowser represents browser automation errors
	ErrorTypeBrowser

	// ErrorTypeSelector represents CSS selector errors
	ErrorTypeSelector

	// ErrorTypeTimeout represents timeout errors
	ErrorTypeTimeout

	// ErrorTypeCaptcha represents CAPTCHA detection errors
	ErrorTypeCaptcha

	// ErrorTypeNetwork represents network-related errors
	ErrorTypeNetwork

	// ErrorTypeConfig represents configuration errors
	ErrorTypeConfig

	// ErrorTypeValidation represents validation errors
	ErrorTypeValidation
)

// String returns the string representation of ErrorType
func (e ErrorType) String() string {
	switch e {
	case ErrorTypeProxy:
		return "proxy"
	case ErrorTypeBrowser:
		return "browser"
	case ErrorTypeSelector:
		return "selector"
	case ErrorTypeTimeout:
		return "timeout"
	case ErrorTypeCaptcha:
		return "captcha"
	case ErrorTypeNetwork:
		return "network"
	case ErrorTypeConfig:
		return "config"
	case ErrorTypeValidation:
		return "validation"
	default:
		return "unknown"
	}
}

// AppError represents an application-specific error with type and context
type AppError struct {
	Type    ErrorType              // Error type/category
	Message string                 // Error message
	Err     error                  // Underlying error (if any)
	Context map[string]interface{} // Additional context
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Type, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Type, e.Message)
}

// Unwrap returns the underlying error for errors.Is/As compatibility
func (e *AppError) Unwrap() error {
	return e.Err
}

// WithContext adds context information to the error
func (e *AppError) WithContext(key string, value interface{}) *AppError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// New creates a new AppError
func New(errType ErrorType, message string) *AppError {
	return &AppError{
		Type:    errType,
		Message: message,
		Context: make(map[string]interface{}),
	}
}

// Wrap wraps an existing error with type and message
func Wrap(err error, errType ErrorType, message string) *AppError {
	return &AppError{
		Type:    errType,
		Message: message,
		Err:     err,
		Context: make(map[string]interface{}),
	}
}

// Is checks if the error is of a specific type
func Is(err error, errType ErrorType) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Type == errType
	}
	return false
}

// IsRetryable determines if an error is retryable
func IsRetryable(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		switch appErr.Type {
		case ErrorTypeTimeout, ErrorTypeNetwork, ErrorTypeProxy:
			return true
		case ErrorTypeCaptcha:
			return false // CAPTCHA requires manual intervention
		default:
			return false
		}
	}
	return false
}

// GetType extracts the ErrorType from an error
func GetType(err error) ErrorType {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Type
	}
	return ErrorTypeUnknown
}

// Common error constructors

// NewProxyError creates a new proxy error
func NewProxyError(message string, err error) *AppError {
	return Wrap(err, ErrorTypeProxy, message)
}

// NewBrowserError creates a new browser error
func NewBrowserError(message string, err error) *AppError {
	return Wrap(err, ErrorTypeBrowser, message)
}

// NewSelectorError creates a new selector error
func NewSelectorError(message string, err error) *AppError {
	return Wrap(err, ErrorTypeSelector, message)
}

// NewTimeoutError creates a new timeout error
func NewTimeoutError(message string, err error) *AppError {
	return Wrap(err, ErrorTypeTimeout, message)
}

// NewCaptchaError creates a new CAPTCHA error
func NewCaptchaError(message string) *AppError {
	return New(ErrorTypeCaptcha, message)
}

// NewNetworkError creates a new network error
func NewNetworkError(message string, err error) *AppError {
	return Wrap(err, ErrorTypeNetwork, message)
}

// NewConfigError creates a new config error
func NewConfigError(message string, err error) *AppError {
	return Wrap(err, ErrorTypeConfig, message)
}

// NewValidationError creates a new validation error
func NewValidationError(message string) *AppError {
	return New(ErrorTypeValidation, message)
}
