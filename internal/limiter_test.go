package internal_test

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/arizard/gomments/internal"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClientIPRateLimiterMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("allows requests under limit", func(t *testing.T) {
		middleware := internal.NewClientIPRateLimiterMiddleware(10) // 10 RPS
		router := gin.New()
		router.Use(middleware)
		router.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		// init the rate limiter, because it needs to fill the bucket
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		time.Sleep(500 * time.Millisecond)

		// Make 5 requests quickly - should all succeed
		for i := 0; i < 5; i++ {
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = "192.168.1.1:12345"
			w := httptest.NewRecorder()

			start := time.Now()
			router.ServeHTTP(w, req)
			duration := time.Since(start)

			assert.Equal(t, http.StatusOK, w.Code)
			// Should not be significantly delayed (rate limiter working)
			assert.Less(t, duration, 50*time.Millisecond, "iteration: %d", i)
		}
	})

	t.Run("rate limits per IP address", func(t *testing.T) {
		middleware := internal.NewClientIPRateLimiterMiddleware(2) // 2 RPS for easier testing
		router := gin.New()
		router.Use(middleware)
		router.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		// init the limiter to fill the leaky bucket
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		time.Sleep(1000 * time.Millisecond)

		// First IP - make 3 requests rapidly
		ip1Delays := make([]time.Duration, 3)
		for i := 0; i < 3; i++ {
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = "192.168.1.1:12345"
			w := httptest.NewRecorder()

			start := time.Now()
			router.ServeHTTP(w, req)
			ip1Delays[i] = time.Since(start)

			assert.Equal(t, http.StatusOK, w.Code)
		}

		// First 2 requests should be fast, 3rd should be delayed
		assert.Less(t, ip1Delays[0], 50*time.Millisecond)
		assert.Less(t, ip1Delays[1], 50*time.Millisecond)
		assert.Greater(t, ip1Delays[2], 400*time.Millisecond) // ~500ms delay for 2 RPS
	})

	t.Run("isolates rate limits between different IPs", func(t *testing.T) {
		middleware := internal.NewClientIPRateLimiterMiddleware(2) // 2 RPS
		router := gin.New()
		router.Use(middleware)
		router.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		// Exhaust IP1's rate limit
		for i := 0; i < 2; i++ {
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = "192.168.1.1:12345"
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
		}

		// IP2 should still be fast (not affected by IP1's limit)
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.2:12345"
		w := httptest.NewRecorder()

		start := time.Now()
		router.ServeHTTP(w, req)
		duration := time.Since(start)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Less(t, duration, 50*time.Millisecond)
	})

	t.Run("handles concurrent requests safely", func(t *testing.T) {
		middleware := internal.NewClientIPRateLimiterMiddleware(100) // High limit to avoid delays
		router := gin.New()
		router.Use(middleware)
		router.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		const numGoroutines = 50
		const requestsPerGoroutine = 10

		var wg sync.WaitGroup
		successCount := int64(0)
		var mu sync.Mutex

		// Launch many goroutines making concurrent requests
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(goroutineID int) {
				defer wg.Done()

				for j := 0; j < requestsPerGoroutine; j++ {
					req := httptest.NewRequest("GET", "/test", nil)
					req.RemoteAddr = "192.168.1.100:12345" // Same IP for all
					w := httptest.NewRecorder()

					router.ServeHTTP(w, req)

					if w.Code == http.StatusOK {
						mu.Lock()
						successCount++
						mu.Unlock()
					}
				}
			}(i)
		}

		wg.Wait()

		// All requests should succeed (no panics or race conditions)
		require.Equal(t, int64(numGoroutines*requestsPerGoroutine), successCount)
	})
}

