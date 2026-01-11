package middleware

import (
	"english-at-lima-cms/internal/repository"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"sync"
	"time"
)

var (
	blacklistedIPs = make(map[string]bool)
	gatekeeperMu   sync.RWMutex
	loginFailures  = make(map[string]int)
	loginMu        sync.Mutex
)

func Gatekeeper() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if isBanned(ip) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "IP bloqueada por seguridad"})
			return
		}
		c.Next()
	}
}

func isBanned(ip string) bool {
	loginMu.Lock()
	defer loginMu.Unlock()

	if blacklistedIPs[ip] {
		return true
	}
	// LA VERSIÓN CORRECTA:
	return loginFailures[ip] >= 5
}

func LoadBlacklist() {
	ips, err := repository.FetchAllBannedIPs()
	if err != nil {
		return
	}
	gatekeeperMu.Lock()
	defer gatekeeperMu.Unlock()
	for _, ip := range ips {
		blacklistedIPs[ip] = true
	}
}

func AddToBlacklist(ip string) {
	gatekeeperMu.Lock()
	defer gatekeeperMu.Unlock()
	blacklistedIPs[ip] = true
}

func IPBlocker() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		gatekeeperMu.RLock()
		isBanned := blacklistedIPs[clientIP]
		gatekeeperMu.RUnlock()

		if isBanned {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Acceso denegado. IP en lista negra.",
			})
			return
		}
		c.Next()
	}
}

func StartBlacklistCleaner() {
	const unaSemana = 10080 * time.Minute
	ticker := time.NewTicker(unaSemana)
	go func() {
		for range ticker.C {
			fmt.Println("🧹 Limpieza semanal de lista negra...")
			LoadBlacklist()
		}
	}()
}

// RecordLoginFailure cuenta los fallos y devuelve TRUE si debe ser bloqueado
func RecordLoginFailure(ip string) bool {
	loginMu.Lock()
	defer loginMu.Unlock()

	loginFailures[ip]++

	// Si llega a 5 fallos, lo bloqueamos
	return loginFailures[ip] >= 5
}

// ResetLoginAttempts limpia el contador cuando el usuario acierta la clave
func ResetLoginAttempts(ip string) {
	loginMu.Lock()
	defer loginMu.Unlock()
	delete(loginFailures, ip)
}
