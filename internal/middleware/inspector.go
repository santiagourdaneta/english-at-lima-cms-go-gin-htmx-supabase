package middleware

import (
	"bytes"
	"english-at-lima-cms/internal/security"
	"github.com/gin-gonic/gin"
	"io"
	"strings"
)

func GlobalSecurityInspector() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "POST" || c.Request.Method == "PATCH" {
			bodyBytes, _ := io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

			bodyString := string(bodyBytes)

			// 1. Protección de hardware (Payload Limit)
			if len(bodyString) > 2*1024*1024 {
				security.LogIntrusion(c, "MASSIVE_PAYLOAD", "Tamaño excedido")
				c.AbortWithStatusJSON(413, gin.H{"error": "Payload demasiado grande"})
				return
			}

			// 2. Protección XSS
			if strings.Contains(bodyString, "<script") || strings.Contains(bodyString, "javascript:") {
				security.LogIntrusion(c, "XSS_ATTEMPT", "Payload sospechoso")
			}
		}
		c.Next()
	}
}
