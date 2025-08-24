package internal

import (
	"log"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/ratelimit"
)

func NewClientIPRateLimiterMiddleware(rps int) gin.HandlerFunc {
	limiters := map[string]ratelimit.Limiter{}
	mu := sync.RWMutex{}

	// periodically flush the limiter map
	ticker := time.NewTicker(1 * time.Minute)
	go func() {
		for {
			<-ticker.C
			mu.Lock()
			limiters = map[string]ratelimit.Limiter{}
			mu.Unlock()
		}
	}()

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
