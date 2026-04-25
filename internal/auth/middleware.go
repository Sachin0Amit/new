package auth

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type contextKey string
const ClaimsKey contextKey = "user_claims"

// AuthMiddleware extracts and validates Bearer token.
func AuthMiddleware(svc *JWTService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			token := strings.TrimPrefix(authHeader, "Bearer ")
			claims, err := svc.ValidateAccessToken(token)
			if err != nil {
				http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), ClaimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RoleMiddleware enforces RBAC.
func RoleMiddleware(requiredRole string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value(ClaimsKey).(*Claims)
			if !ok || (claims.Role != requiredRole && claims.Role != "sovereign") {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RateLimitMiddleware protects sensitive endpoints (login).
type IPStats struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

func RateLimitMiddleware() func(http.Handler) http.Handler {
	var stats sync.Map
	
	// Cleanup goroutine
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			stats.Range(func(key, value interface{}) bool {
				if time.Since(value.(*IPStats).lastSeen) > 5*time.Minute {
					stats.Delete(key)
				}
				return true
			})
		}
	}()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr // Production should use X-Forwarded-For
			
			val, _ := stats.LoadOrStore(ip, &IPStats{
				limiter:  rate.NewLimiter(rate.Every(time.Minute/10), 10), // 10 per minute
				lastSeen: time.Now(),
			})
			
			s := val.(*IPStats)
			s.lastSeen = time.Now()
			
			if !s.limiter.Allow() {
				http.Error(w, "Too many login attempts", http.StatusTooManyRequests)
				return
			}
			
			next.ServeHTTP(w, r)
		})
	}
}

// CSRFMiddleware validates X-CSRF-Token header against a cookie.
func CSRFMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "GET" || r.Method == "HEAD" || r.Method == "OPTIONS" {
				next.ServeHTTP(w, r)
				return
			}

			csrfCookie, err := r.Cookie("csrf_token")
			if err != nil {
				http.Error(w, "Missing CSRF cookie", http.StatusForbidden)
				return
			}

			csrfHeader := r.Header.Get("X-CSRF-Token")
			if csrfHeader == "" || csrfHeader != csrfCookie.Value {
				http.Error(w, "Invalid CSRF token", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
