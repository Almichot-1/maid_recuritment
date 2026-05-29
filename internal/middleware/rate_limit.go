package middleware

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"maid-recruitment-tracking/pkg/utils"
)

type rateLimitEntry struct {
	count   int
	resetAt time.Time
}

type fixedWindowRateLimiter struct {
	name    string
	limit   int
	window  time.Duration
	entries map[string]rateLimitEntry
	mu      sync.Mutex
}

func NewIPRateLimitMiddleware(name string, limit int, window time.Duration) func(http.Handler) http.Handler {
	limiter := &fixedWindowRateLimiter{
		name:    strings.TrimSpace(name),
		limit:   limit,
		window:  window,
		entries: make(map[string]rateLimitEntry),
	}

	if limiter.limit <= 0 {
		limiter.limit = 10
	}
	if limiter.window <= 0 {
		limiter.window = time.Minute
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			allowed, retryAfter := limiter.allow(rateLimitKey(r, limiter.name))
			if !allowed {
				w.Header().Set("Retry-After", retryAfter.String())
				_ = utils.WriteJSON(w, http.StatusTooManyRequests, map[string]string{"error": "too many requests"})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func (l *fixedWindowRateLimiter) allow(key string) (bool, time.Duration) {
	now := time.Now().UTC()

	l.mu.Lock()
	defer l.mu.Unlock()

	for entryKey, entry := range l.entries {
		if now.After(entry.resetAt) {
			delete(l.entries, entryKey)
		}
	}

	entry, ok := l.entries[key]
	if !ok || now.After(entry.resetAt) {
		l.entries[key] = rateLimitEntry{
			count:   1,
			resetAt: now.Add(l.window),
		}
		return true, 0
	}

	if entry.count >= l.limit {
		return false, time.Until(entry.resetAt).Round(time.Second)
	}

	entry.count++
	l.entries[key] = entry
	return true, 0
}

func rateLimitKey(r *http.Request, prefix string) string {
	return prefix + ":" + requestIP(r)
}

func requestIP(r *http.Request) string {
	if r == nil {
		return "unknown"
	}

	if forwarded := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); forwarded != "" {
		parts := strings.Split(forwarded, ",")
		return strings.TrimSpace(parts[0])
	}

	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err == nil && strings.TrimSpace(host) != "" {
		return strings.TrimSpace(host)
	}

	if trimmed := strings.TrimSpace(r.RemoteAddr); trimmed != "" {
		return trimmed
	}

	return "unknown"
}
