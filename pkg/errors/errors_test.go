package errors

import (
	"fmt"
	"testing"
	"errors"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	inner := errors.New("inner error")
	err := New(CodeInternal, "Operation failed", inner)
	
	assert.NotNil(t, err)
	assert.Equal(t, CodeInternal, err.Code)
	assert.Equal(t, "Operation failed", err.Message)
	assert.Equal(t, fmt.Sprintf("[%s] %s: %v", CodeInternal, "Operation failed", inner), err.Error())
}

func TestWrap(t *testing.T) {
	baseErr := errors.New("base error")
	err := Wrap(CodeNotFound, baseErr, "Resource missing")
	
	assert.NotNil(t, err)
	assert.Equal(t, CodeNotFound, err.Code)
	assert.ErrorIs(t, err, baseErr) // Test Unwrap support
	assert.Equal(t, "[NOT_FOUND] Resource missing: base error", err.Error())
}

func TestIsCode(t *testing.T) {
	err := New(CodeInvalidArgument, "invalid input", nil)
	assert.True(t, IsCode(err, CodeInvalidArgument))
	assert.False(t, IsCode(err, CodeInternal))
	
	// Test normal error
	assert.False(t, IsCode(errors.New("plain error"), CodeInvalidArgument))
	
	// Test wrapped sovereign error
	wrapped := fmt.Errorf("wrapped: %w", err)
	assert.True(t, IsCode(wrapped, CodeInvalidArgument))
}

func TestWrapNil(t *testing.T) {
	err := Wrap(CodeInternal, nil, "Op")
	assert.Nil(t, err)
}
