package errors

import (
	"errors"
	"fmt"
)

// Code defines machine-readable application error codes.
type Code string

const (
	CodeInternal        Code = "INTERNAL_ERROR"
	CodeNotFound        Code = "NOT_FOUND"
	CodeInference       Code = "INFERENCE_FAILED"
	CodeStorage         Code = "STORAGE_ERROR"
	CodeValidation      Code = "VALIDATION_ERROR"
	CodeInvalidArgument Code = "INVALID_ARGUMENT"
)

// SovereignError is the unified error type for the system.
type SovereignError struct {
	Code    Code   `json:"code"`
	Message string `json:"message"`
	Err     error  `json:"-"`
}

func (e *SovereignError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *SovereignError) Unwrap() error {
	return e.Err
}

// New creates a new SovereignError.
func New(code Code, message string, err error) *SovereignError {
	return &SovereignError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// Wrap wraps an existing error into a SovereignError.
func Wrap(code Code, err error, message string) *SovereignError {
	if err == nil {
		return nil
	}
	return &SovereignError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// IsCode checks if an error is a SovereignError with a specific code.
func IsCode(err error, code Code) bool {
	var se *SovereignError
	if errors.As(err, &se) {
		return se.Code == code
	}
	return false
}
