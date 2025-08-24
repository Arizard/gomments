package internal

import (
	"log"
	"sync"

	"github.com/gin-gonic/gin"
	"go.uber.org/ratelimit"
)

func NewClientIPRateLimiterMiddleware(rps int) gin.HandlerFunc {
	limiters := map[string]ratelimit.Limiter{}
	mu := sync.RWMutex{}

	return func(c *gin.Context) {
		ip := c.ClientIP()

		log.Printf("client ip: %s", ip)

		mu.RLock()
		l, rOK := limiters[ip]
		mu.RUnlock()

		if !rOK {
			mu.Lock()
			if wL, wOK := limiters[ip]; wOK {
				l = wL
			} else {
				l = ratelimit.New(rps)
				limiters[ip] = l
			}
			mu.Unlock()
		}

		l.Take()
	}
}
