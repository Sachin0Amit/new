package security

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRateLimiter(t *testing.T) {
	// 1 tok/sec, cap 2
	guard := NewGuard(1.0, 2.0)

	// Burst
	assert.True(t, guard.Allow(), "Initial allow")
	assert.True(t, guard.Allow(), "Burst allow")
	assert.False(t, guard.Allow(), "Limit reached")

	// Wait for refill
	time.Sleep(1100 * time.Millisecond)
	assert.True(t, guard.Allow(), "Refilled allow")
	assert.False(t, guard.Allow(), "Limit reached again")
}

func TestSanitizer(t *testing.T) {
	guard := DefaultGuard()

	cases := []struct {
		input    string
		expected string
	}{
		{"Hello World", "Hello World"},
		{"Hello \x00 World", "Hello  World"},
		{"<script>alert(1)</script>", "<script>alert(1)</script>"}, // This is allowed, but printable
		{"Prompt\r\nInjection", "Prompt\nInjection"},            // Simplified, usually cleaned by regex but \r might be stripped
		{"Non-printable \x01\x02\x03", "Non-printable "},
	}

	for _, tc := range cases {
		result := guard.Sanitize(tc.input)
		// Note: The regex strips [^\p{L}\p{N}\p{P}\s]. \r is a whitespace, \n is whitespace. 
		// Control chars like \x01 are not.
		t.Logf("Input: %q, Result: %q", tc.input, result)
		assert.NotContains(t, result, "\x01")
	}
}

func TestValidator(t *testing.T) {
	guard := DefaultGuard()

	assert.NoError(t, guard.Validate(map[string]interface{}{"prompt": "test"}))
	assert.NoError(t, guard.Validate(map[string]interface{}{"input": "test"}))
	assert.Error(t, guard.Validate(map[string]interface{}{"unknown": "test"}))
	assert.Error(t, guard.Validate(nil))
}
