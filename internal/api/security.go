package api

import (
	"net"
	"net/http"
	"sync"
	"time"
)

// RateLimiter implements token bucket rate limiting
type RateLimiter struct {
	limits map[string]*TokenBucket
	mu     sync.RWMutex
	rps    int
	burst  int
}

// TokenBucket holds the state for a single rate limit bucket
type TokenBucket struct {
	tokens    float64
	lastRefil time.Time
	mu        sync.RWMutex
}

// NewRateLimiter creates a new rate limiter (requests per minute)
func NewRateLimiter(requestsPerMinute, burst int) *RateLimiter {
	return &RateLimiter{
		limits: make(map[string]*TokenBucket),
		rps:    requestsPerMinute / 60,
		burst:  burst,
	}
}

// Allow checks if a request is allowed
func (rl *RateLimiter) Allow(identifier string) bool {
	rl.mu.Lock()
	bucket, exists := rl.limits[identifier]
	if !exists {
		bucket = &TokenBucket{
			tokens:    float64(rl.burst),
			lastRefil: time.Now(),
		}
		rl.limits[identifier] = bucket
	}
	rl.mu.Unlock()

	return bucket.Allow(rl.rps)
}

// Allow checks if a request can proceed
func (tb *TokenBucket) Allow(rps int) bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastRefil).Seconds()
	tb.tokens += elapsed * float64(rps)
	tb.lastRefil = now

	if tb.tokens >= 1 {
		tb.tokens--
		return true
	}
	return false
}

// Middleware returns an HTTP middleware for rate limiting
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIP := getClientIP(r)

		if !rl.Allow(clientIP) {
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// getClientIP extracts the client IP from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		return forwarded
	}

	// Fall back to RemoteAddr
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ip
}

// CapabilityEnforcer enforces policies on tool usage
type CapabilityEnforcer struct {
	policies map[string]Policy
	mu       sync.RWMutex
}

// Policy defines what a tool can do
type Policy struct {
	ToolName    string      `json:"tool_name"`
	Enabled     bool        `json:"enabled"`
	RateLimit   int         `json:"rate_limit"`
	Timeout     int         `json:"timeout_seconds"`
	Constraints map[string]string `json:"constraints"`
	AllowList   []string    `json:"allow_list"` // Whitelisted values
	DenyList    []string    `json:"deny_list"`  // Blacklisted values
}

// NewCapabilityEnforcer creates a new enforcer
func NewCapabilityEnforcer() *CapabilityEnforcer {
	return &CapabilityEnforcer{
		policies: make(map[string]Policy),
	}
}

// RegisterPolicy registers a tool policy
func (ce *CapabilityEnforcer) RegisterPolicy(policy Policy) {
	ce.mu.Lock()
	defer ce.mu.Unlock()
	ce.policies[policy.ToolName] = policy
}

// CanExecute checks if a tool can be executed
func (ce *CapabilityEnforcer) CanExecute(toolName string, args interface{}) (bool, string) {
	ce.mu.RLock()
	defer ce.mu.RUnlock()

	policy, exists := ce.policies[toolName]
	if !exists {
		return false, "tool not registered"
	}

	if !policy.Enabled {
		return false, "tool is disabled"
	}

	// Check constraints (simplified - would need proper argument parsing)
	return true, ""
}

// AuditEntry logs tool execution
type AuditEntry struct {
	Timestamp   time.Time              `json:"timestamp"`
	Actor       string                 `json:"actor"`
	Tool        string                 `json:"tool"`
	Arguments   interface{}            `json:"arguments"`
	Result      interface{}            `json:"result"`
	Error       string                 `json:"error,omitempty"`
	Duration    time.Duration          `json:"duration"`
	Allowed     bool                   `json:"allowed"`
	Signature   string                 `json:"signature"` // ED25519 signature
}

// AuditTrail logs all tool executions
type AuditTrail struct {
	entries []AuditEntry
	mu      sync.RWMutex
}

// NewAuditTrail creates a new audit trail
func NewAuditTrail() *AuditTrail {
	return &AuditTrail{
		entries: make([]AuditEntry, 0),
	}
}

// Log logs an entry
func (at *AuditTrail) Log(entry AuditEntry) {
	at.mu.Lock()
	defer at.mu.Unlock()
	entry.Timestamp = time.Now()
	at.entries = append(at.entries, entry)
}

// GetEntries returns all entries
func (at *AuditTrail) GetEntries() []AuditEntry {
	at.mu.RLock()
	defer at.mu.RUnlock()
	entries := make([]AuditEntry, len(at.entries))
	copy(entries, at.entries)
	return entries
}

// GetEntriesSince returns entries since a given time
func (at *AuditTrail) GetEntriesSince(since time.Time) []AuditEntry {
	at.mu.RLock()
	defer at.mu.RUnlock()

	var entries []AuditEntry
	for _, entry := range at.entries {
		if entry.Timestamp.After(since) {
			entries = append(entries, entry)
		}
	}
	return entries
}
