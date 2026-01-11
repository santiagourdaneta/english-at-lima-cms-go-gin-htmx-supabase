package handlers

import (
	"english-at-lima-cms/internal/repository"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// ShowLogin renderiza la página de acceso
func ShowLogin(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", nil)
}

// Login maneja la autenticación triple-check
func Login(c *gin.Context) {
	email := c.PostForm("email")
	password := c.PostForm("password")

	token, err := repository.AuthenticateUser(email, password)
	if err != nil {
		LogIntrusion(c, "LOGIN_FAIL", email)
		SendToast(c, "Credenciales inválidas", "error")
		return
	}

	session := sessions.Default(c)
	session.Set("user_id", email)
	session.Set("token", token)

	if err := session.Save(); err != nil {
		SendToast(c, "Error de sesión", "error")
		return
	}

	c.Header("HX-Redirect", "/admin")
	c.Status(http.StatusOK)
}

// Logout destruye la sesión de forma segura
func Logout(c *gin.Context) {
	// 1. Primero obtenemos la sesión
	session := sessions.Default(c)

	// 2. Limpiamos los datos
	session.Clear()

	// 3. Guardamos los cambios y verificamos el error
	if err := session.Save(); err != nil {
		// Si falla el guardado, al menos notificamos
		SendToast(c, "Error al cerrar sesión", "error")
		return
	}

	// 4. Forzamos la expiración de la cookie
	c.SetCookie("mysession", "", -1, "/", "", false, true)

	// 5. Redirigimos al login
	c.Redirect(http.StatusSeeOther, "/login")
}
