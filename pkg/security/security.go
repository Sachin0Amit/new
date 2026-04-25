package security

import (
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/Sachin0Amit/new/pkg/errors"
)

// SecurityManager defines the core defensive operations for the system.
type SecurityManager interface {
	Allow() bool
	Sanitize(input string) string
	Validate(payload map[string]interface{}) error
}

// SovereignGuard implements SecurityManager with local-first defensive logic.
type SovereignGuard struct {
	mu            sync.Mutex
	tokens        float64
	capacity      float64
	rate          float64
	lastRefill    time.Time
	sanitizerRegex *regexp.Regexp
}

// NewGuard creates a new security guard with a specific rate (tokens/sec) and burst capacity.
func NewGuard(rate, capacity float64) *SovereignGuard {
	return &SovereignGuard{
		tokens:        capacity,
		capacity:      capacity,
		rate:          rate,
		lastRefill:    time.Now(),
		sanitizerRegex: regexp.MustCompile(`[^\p{L}\p{N}\p{P}\s]`), // Strip non-printable/control chars
	}
}

// Allow implements a thread-safe Token Bucket rate limiter.
func (g *SovereignGuard) Allow() bool {
	g.mu.Lock()
	defer g.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(g.lastRefill).Seconds()
	g.tokens += elapsed * g.rate
	if g.tokens > g.capacity {
		g.tokens = g.capacity
	}
	g.lastRefill = now

	if g.tokens >= 1.0 {
		g.tokens -= 1.0
		return true
	}

	return false
}

// Sanitize removes potentially malicious control characters and unusual Unicode sequences.
func (g *SovereignGuard) Sanitize(input string) string {
	// 1. Basic trimming
	input = strings.TrimSpace(input)
	
	// 2. Filter out control characters and non-printable sequences
	input = g.sanitizerRegex.ReplaceAllString(input, "")
	
	// 3. Limit total length to prevent resource exhaustion (e.g., 64KB)
	if len(input) > 64*1024 {
		input = input[:64*1024]
	}
	
	return input
}

// Validate ensures the payload contains required fields and adheres to safety constraints.
func (g *SovereignGuard) Validate(payload map[string]interface{}) error {
	if payload == nil {
		return errors.New(errors.CodeValidation, "payload cannot be nil", nil)
	}

	// Example: Ensure "prompt" or "input" exists
	if _, ok := payload["prompt"]; !ok {
		if _, ok := payload["input"]; !ok {
			return errors.New(errors.CodeValidation, "payload must contain 'prompt' or 'input' key", nil)
		}
	}

	return nil
}

// Helper for rapid setup
func DefaultGuard() *SovereignGuard {
	return NewGuard(5.0, 10.0) // 5 requests/sec, 10 burst
}
