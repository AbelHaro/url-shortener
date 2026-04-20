package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type RateLimitMiddleware struct {
	limit  int
	window time.Duration
	debug  bool

	mu      sync.Mutex
	clients map[string]*rateLimitState
}

type rateLimitState struct {
	count   int
	resetAt time.Time
}

func NewRateLimitMiddleware(limit int, window time.Duration, debug bool) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		limit:   limit,
		window:  window,
		debug:   debug,
		clients: make(map[string]*rateLimitState),
	}
}

func (m *RateLimitMiddleware) Limit() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip rate limiting in debug mode (tests)
		if m.debug {
			c.Next()
			return
		}

		clientIP := c.ClientIP()
		now := time.Now()

		m.mu.Lock()
		state, ok := m.clients[clientIP]
		if !ok || now.After(state.resetAt) {
			m.clients[clientIP] = &rateLimitState{
				count:   1,
				resetAt: now.Add(m.window),
			}
			m.mu.Unlock()
			c.Next()
			return
		}

		if state.count >= m.limit {
			resetAfter := time.Until(state.resetAt)
			m.mu.Unlock()
			if resetAfter > 0 {
				c.Header("Retry-After", formatRetryAfter(resetAfter))
			}
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			return
		}

		state.count++
		m.mu.Unlock()
		c.Next()
	}
}

func formatRetryAfter(d time.Duration) string {
	seconds := int64(d / time.Second)
	if d%time.Second != 0 {
		seconds++
	}
	if seconds < 1 {
		seconds = 1
	}
	return strconv.FormatInt(seconds, 10)
}
