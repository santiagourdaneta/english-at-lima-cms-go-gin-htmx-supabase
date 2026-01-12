package main

import (
	"english-at-lima-cms/internal/handlers"

	"english-at-lima-cms/internal/middleware"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"net/http"
	"os"
	"strings"
)

func main() {

	// Sincronizar IPs baneadas antes de aceptar peticiones
	middleware.LoadBlacklist()
	middleware.StartBlacklistCleaner() // Inicia el cronómetro de limpieza

	_ = godotenv.Load()
	r := setupRouter()

	// El IPBlocker debe ser el PRIMER middleware de todos
	r.Use(middleware.IPBlocker())
	r.Use(middleware.GlobalSecurityInspector()) // El que revisa XSS

	// LIMITAR TODAS LAS PETICIONES A 2MB
	// Si alguien intenta enviar más que esto (como un texto infinito),
	// el servidor le cierra la puerta en la cara automáticamente.
	r.Use(MaxAllowedSize(2 << 20))

	// 1. Configurar el almacenamiento de la sesión (usa una clave secreta)
	store := cookie.NewStore([]byte(os.Getenv("SESSION_SECRET")))
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   3600 * 8, // 8 horas
		HttpOnly: true,
		Secure:   true, // OBLIGATORIO para Render (HTTPS)
		SameSite: http.SameSiteLaxMode,
	})
	r.Use(sessions.Sessions("mysession", store))

	// 2. Aplicar el middleware de sesiones a TODO el servidor
	r.Use(sessions.Sessions("mysession", store))

	r.LoadHTMLGlob("templates/*.html")
	r.Static("/static", "./static")

	_ = r.Run(":8080")
}

func setupRouter() *gin.Engine {
	r := gin.Default()

	// Rutas públicas
	r.GET("/login", handlers.ShowLogin)
	r.POST("/login", middleware.RateLimiter(), handlers.Login)
	r.GET("/logout", handlers.Logout) // Logout general

	r.GET("/", func(c *gin.Context) {
    c.JSON(200, gin.H{"message": "Servidor funcionando"})
})

	// Grupo Admin PROTEGIDO
	admin := r.Group("/admin")
	admin.Use(middleware.AuthRequired())
	{
		admin.GET("/", func(c *gin.Context) {
			c.HTML(http.StatusOK, "admin.html", nil)
		})

		admin.GET("/logout", handlers.Logout)

		admin.GET("/logs", handlers.GetAuditLogs)

		// --- MÓDULO FRASES ---
		admin.GET("/sentences", handlers.GetSentences)
		admin.GET("/sentences/new", handlers.NewSentenceForm)
		admin.POST("/sentences/save", handlers.SaveSentence)
		admin.POST("/sentences/update/:id", handlers.UpdateSentence)
		admin.DELETE("/sentences/:id", handlers.DeleteSentence)

		// --- MÓDULO RECURSOS ---
		admin.GET("/resources", handlers.GetResources)
		admin.GET("/resources/new", handlers.NewResourceForm)
		admin.POST("/resources/save", handlers.SaveResource)
		admin.POST("/resources/update/:id", handlers.UpdateResource)
		admin.DELETE("/resources/:id", handlers.DeleteResource)

		// --- MÓDULO QUIZZES ---
		admin.GET("/quizzes", handlers.GetQuizzes)
		admin.GET("/quizzes/new", handlers.NewQuizForm)
		admin.POST("/quizzes/save", handlers.SaveQuiz)
		admin.POST("/quizzes/update/:id", handlers.UpdateQuiz)
		admin.DELETE("/quizzes/:id", handlers.DeleteQuiz)

		// Búsqueda y Stats
		admin.GET("/search", handlers.GlobalSearch)
		admin.GET("/stats", handlers.GetStats)

		// EXPORTAR A CSV
		admin.GET("/sentences/export", handlers.ExportSentencesCSV)
		admin.GET("/quizzes/export", handlers.ExportQuizzesCSV)
	}

	return r
}

func MaxAllowedSize(limit int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, limit)
		c.Next()

		// Si al procesar la petición hubo un error por tamaño
		if len(c.Errors) > 0 {
			for _, e := range c.Errors {
				if strings.Contains(e.Error(), "request body too large") {
					c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "El envío supera el límite permitido"})
					return
				}
			}
		}
	}
}
