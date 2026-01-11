package handlers

import (
	"encoding/csv"
	"encoding/json"
	"english-at-lima-cms/internal/models"
	"english-at-lima-cms/internal/repository"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func NewResourceForm(c *gin.Context) {
	c.HTML(http.StatusOK, "new-resource.html", nil)
}

func ValidateResource(title, url, resType string) error {
	title = strings.TrimSpace(title)
	if len(title) < 3 || len(title) > 100 {
		return fmt.Errorf("título debe tener entre 3 y 100 caracteres")
	}
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return fmt.Errorf("URL debe ser válida y segura (http/https)")
	}
	if resType == "" {
		return fmt.Errorf("debe seleccionar un tipo de recurso")
	}
	return nil
}

func SaveResource(c *gin.Context) {
	// 1. Auto-Sanitizado
	title := Sanitize(c.PostForm("title"))
	url := strings.TrimSpace(c.PostForm("url")) // Las URLs no se sanean igual, solo se limpian espacios
	resType := Sanitize(c.PostForm("type"))

	// 2. Validación Robusta
	if err := ValidateResource(title, url, resType); err != nil {
		SendToast(c, err.Error(), "error")
		return
	}

	// 3. Persistencia en Supabase
	if err := repository.InsertResource(title, url, resType); err != nil {
		SendToast(c, "Error al guardar en la base de datos", "error")
		return
	}

	SendToast(c, "Recurso guardado exitosamente", "success")
	c.Header("HX-Trigger", "refreshList") // HTMX recarga la lista automáticamente
}

// Handler para ACTUALIZAR
func UpdateResource(c *gin.Context) {
	id := c.Param("id")
	title := strings.TrimSpace(c.PostForm("title"))
	url := strings.TrimSpace(c.PostForm("url"))
	resType := c.PostForm("type")

	// LA ADUANA: Validación robusta
	if err := ValidateResource(title, url, resType); err != nil {
		SendToast(c, err.Error(), "error")
		return
	}

	err := repository.UpdateResource(id, title, url, resType)
	if err != nil {
		SendToast(c, "Error al actualizar el recurso", "error")
		return
	}

	SendToast(c, "Recurso actualizado con éxito", "success")
	c.Header("HX-Trigger", "refreshList")
}

func GetResources(c *gin.Context) {
	resp, err := repository.CallSupabase("GET", "resources", nil, "select=*&order=title.asc")
	if err != nil || resp == nil {
		c.String(http.StatusInternalServerError, "Error de conexión")
		return
	}
	defer resp.Body.Close()

	var data []models.Resource
	_ = json.NewDecoder(resp.Body).Decode(&data) // <--- SOLUCIONA errcheck
	c.HTML(http.StatusOK, "resources.html", gin.H{"Resources": data})
}

func DeleteResource(c *gin.Context) {
	filter := "id=eq." + c.Param("id")
	resp, err := repository.CallSupabase("DELETE", "resources", nil, filter)
	if err == nil && resp != nil {
		defer resp.Body.Close()
	}
	c.Status(http.StatusOK)
}

func ExportResourcesCSV(c *gin.Context) {
	var data []models.Resource
	resp, err := repository.CallSupabase("GET", "resources", nil, "select=*&order=title.asc")
	if err != nil || resp == nil {
		c.String(http.StatusInternalServerError, "Error al obtener recursos")
		return
	}
	defer resp.Body.Close()

	_ = json.NewDecoder(resp.Body).Decode(&data)

	c.Header("Content-Disposition", "attachment; filename=resources.csv")
	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()

	_ = writer.Write([]string{"ID", "Titulo", "Tipo", "URL"})
	for _, r := range data {
		_ = writer.Write([]string{fmt.Sprint(r.ID), r.Title, r.Type, r.URL})
	}
}
