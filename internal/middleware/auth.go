package middleware

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		user := session.Get("user_id")

		if user == nil {
			// Si no hay usuario en sesión, lo mandamos fuera
			c.Redirect(http.StatusSeeOther, "/login")
			c.Abort() // <--- Detiene el resto de la ejecución
			return
		}
		c.Next()
	}
}
