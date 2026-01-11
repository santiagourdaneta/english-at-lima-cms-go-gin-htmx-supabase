package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	// Cambiamos mu por rateLimitMu para evitar colisiones
	rateLimitMu sync.Mutex
	clients     = make(map[string]time.Time)
)

const (
	limitDuration = 1 * time.Second
)

func RateLimiter() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		rateLimitMu.Lock()
		lastSeen, exists := clients[ip]
		rateLimitMu.Unlock()

		if exists && time.Since(lastSeen) < limitDuration {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "Demasiadas peticiones. Por favor, espera un segundo.",
			})
			return
		}

		rateLimitMu.Lock()
		clients[ip] = time.Now()
		rateLimitMu.Unlock()

		c.Next()
	}
}
