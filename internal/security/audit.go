package security

import (
	"english-at-lima-cms/internal/repository"
	"fmt"
	"github.com/gin-gonic/gin"
)

func LogIntrusion(c *gin.Context, eventType string, data string) {
	ip := c.ClientIP()
	fmt.Printf("🚨 [SEGURIDAD] %s | IP: %s\n", eventType, ip)

	// Registro en segundo plano para no saturar tu RAM
	go func() {
		_ = repository.InsertAuditLog(ip, eventType, data)
	}()
}
