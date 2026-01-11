package handlers

import (
	"english-at-lima-cms/internal/repository"
	"fmt"
	"github.com/gin-gonic/gin"
	"html"
	"net/http"
	"regexp"
	"strings"
)

func Sanitize(input string) string {
	reControl := regexp.MustCompile(`[\x00-\x1F\x7F]`)
	clean := reControl.ReplaceAllString(input, "")
	const htmlRegex = `<[^>]*>`
	reHTML := regexp.MustCompile(htmlRegex)
	clean = reHTML.ReplaceAllString(clean, "")
	clean = html.UnescapeString(clean)
	return strings.TrimSpace(clean)
}

func LogIntrusion(c *gin.Context, eventType string, data string) {
	ip := c.ClientIP()
	fmt.Printf("🚨 [SEGURIDAD] Tipo: %s | IP: %s\n", eventType, ip)
	go func() {
		_ = repository.InsertAuditLog(ip, eventType, data)
	}()
}

func SendToast(c *gin.Context, message string, msgType string) {
	headerValue := fmt.Sprintf(`{"showToast": {"message": "%s", "type": "%s"}}`, message, msgType)
	c.Header("HX-Trigger", headerValue)
	c.Status(http.StatusUnprocessableEntity)
}
