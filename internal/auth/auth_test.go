package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Sachin0Amit/new/internal/mesh"
	"github.com/stretchr/testify/assert"
)

func TestSecurityHardening(t *testing.T) {
	// Mock Mesh for testing
	m, _ := mesh.NewKnowledgeMesh(t.TempDir())
	svc := NewJWTService("test_secret", m)

	t.Run("Token Tampering", func(t *testing.T) {
		token, _ := svc.GenerateAccessToken("user1", "node1", "user")
		
		// Modify last character to invalidate signature
		tampered := token[:len(token)-1] + "X"
		
		_, err := svc.ValidateAccessToken(tampered)
		assert.Error(t, err)
	})

	t.Run("Token Expiry", func(t *testing.T) {
		// This is hard to test without mocking time, but we can verify 
		// the claims have the correct expiry.
		token, _ := svc.GenerateAccessToken("user1", "node1", "user")
		claims, _ := svc.ValidateAccessToken(token)
		
		assert.WithinDuration(t, time.Now().Add(15*time.Minute), claims.ExpiresAt.Time, 5*time.Second)
	})

	t.Run("CSRF Protection", func(t *testing.T) {
		middleware := CSRFMiddleware()
		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		// 1. Missing Header
		req := httptest.NewRequest("POST", "/api/data", nil)
		req.AddCookie(&http.Cookie{Name: "csrf_token", Value: "valid_token"})
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusForbidden, rr.Code)

		// 2. Matching Header
		req = httptest.NewRequest("POST", "/api/data", nil)
		req.AddCookie(&http.Cookie{Name: "csrf_token", Value: "valid_token"})
		req.Header.Set("X-CSRF-Token", "valid_token")
		rr = httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("Rate Limiting", func(t *testing.T) {
		middleware := RateLimitMiddleware()
		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		// Send 11 requests from same IP (limit is 10)
		for i := 0; i < 11; i++ {
			req := httptest.NewRequest("POST", "/login", nil)
			req.RemoteAddr = "1.2.3.4"
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)
			
			if i < 10 {
				assert.Equal(t, http.StatusOK, rr.Code)
			} else {
				assert.Equal(t, http.StatusTooManyRequests, rr.Code)
			}
		}
	})
}
