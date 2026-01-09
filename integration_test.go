package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestPublicPageContent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	
	router.LoadHTMLGlob("templates/*")
	
	router.GET("/public", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "English At Lima",
		})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/public", nil)
	router.ServeHTTP(w, req)

	// Verificaciones E2E
	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "English At Lima", "El HTML debería contener el título")
}