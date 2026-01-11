package handlers

import (
	"encoding/json"
	"english-at-lima-cms/internal/middleware"
	"english-at-lima-cms/internal/models"
	"english-at-lima-cms/internal/repository"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func GlobalSearch(c *gin.Context) {
	query := strings.TrimSpace(c.Query("search"))
	if len(query) < 2 {
		c.Status(http.StatusNoContent)
		return
	}

	var sentences []models.Sentence
	var quizzes []models.Quiz
	var resources []models.Resource
	var wg sync.WaitGroup
	wg.Add(3)

	search := func(table string, target interface{}, filter string) {
		defer wg.Done()
		resp, err := repository.CallSupabase("GET", table, nil, filter)
		if err == nil && resp != nil {
			defer resp.Body.Close()
			_ = json.NewDecoder(resp.Body).Decode(target)
		}
	}

	go search("sentences", &sentences, fmt.Sprintf("or=(english.ilike.*%s*,spanish.ilike.*%s*)", query, query))
	go search("quizzes", &quizzes, fmt.Sprintf("question.ilike.*%s*", query))
	go search("resources", &resources, fmt.Sprintf("title.ilike.*%s*", query))

	wg.Wait()

	if len(sentences)+len(quizzes)+len(resources) == 0 {
		c.Writer.Header().Set("HX-Trigger", fmt.Sprintf(`{"showToast": {"message": "No se encontró nada para '%s'", "type": "error"}}`, query))
	}

	c.HTML(http.StatusOK, "search-results.html", gin.H{
		"Sentences": sentences, "Quizzes": quizzes, "Resources": resources, "Query": query,
	})
}

func GetStats(c *gin.Context) {
	var wg sync.WaitGroup
	wg.Add(3)

	counts := gin.H{"sentences": 0, "quizzes": 0, "resources": 0}
	var mu sync.Mutex // Usamos un mutex local para escribir en el mapa counts de forma segura

	getTableCount := func(table string, key string) {
		defer wg.Done()
		filter := "select=id&limit=1"
		resp, err := repository.CallSupabase("GET", table, nil, filter)

		if err == nil && resp != nil {
			defer resp.Body.Close()

			rangeHeader := resp.Header.Get("Content-Range")
			if rangeHeader != "" {
				parts := strings.Split(rangeHeader, "/")
				if len(parts) > 1 {
					mu.Lock()
					counts[key] = parts[1]
					mu.Unlock()
				}
			}
		}
	}

	go getTableCount("sentences", "sentences")
	go getTableCount("quizzes", "quizzes")
	go getTableCount("resources", "resources")

	wg.Wait()
	c.HTML(http.StatusOK, "stats-panel.html", counts)
}

func GetAdminDashboard(c *gin.Context) {
	session := sessions.Default(c)
	email := session.Get("user_id")

	c.HTML(http.StatusOK, "admin.html", gin.H{
		"UserEmail": email,
	})
}

func GetAuditLogs(c *gin.Context) {
	logs, _ := repository.GetAuditLogs()
	c.HTML(http.StatusOK, "audit_logs.html", gin.H{
		"logs": logs,
	})
}

func BanIPHandler(c *gin.Context) {
	ipToBan := c.Param("ip")

	err := repository.BanIP(ipToBan, "Actividad maliciosa detectada")
	if err != nil {
		SendToast(c, "Error al banear", "error")
		return
	}

	middleware.AddToBlacklist(ipToBan)

	SendToast(c, "IP bloqueada con éxito", "success")
	c.Status(http.StatusOK)
}
