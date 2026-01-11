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

// Renderiza el formulario
func NewSentenceForm(c *gin.Context) {
	c.HTML(http.StatusOK, "new-sentence.html", nil)
}

// ValidateSentence comprueba la integridad de la frase (Separada para Testeo)
func ValidateSentence(english, spanish string) error {
	if len(strings.TrimSpace(english)) < 5 {
		return fmt.Errorf("la frase es demasiado corta")
	}
	if len(english) > 500 {
		return fmt.Errorf("la frase excede el límite de 500 caracteres")
	}
	if strings.TrimSpace(spanish) == "" {
		return fmt.Errorf("la traducción no puede estar vacía")
	}
	return nil
}

// Procesa el guardado
func SaveSentence(c *gin.Context) {

	var s models.Sentence

	// PASO 1: Auto-Sanitizado (Magia automática)
	s.Spanish = Sanitize(c.PostForm("spanish"))
	s.English = Sanitize(c.PostForm("english"))

	// PASO 2: Validación (Sobre el texto ya limpio)
	if err := ValidateSentence(s.English, s.Spanish); err != nil {
		SendToast(c, err.Error(), "error")
		return
	}

	resp, err := repository.CallSupabase("POST", "sentences", s, "")
	if err == nil && resp != nil {
		defer resp.Body.Close()
	}
	c.Redirect(http.StatusSeeOther, "/admin/sentences")
}

func GetSentences(c *gin.Context) {
	resp, _ := repository.CallSupabase("GET", "sentences", nil, "select=*&order=id.desc")
	if resp != nil {
		defer resp.Body.Close() // <--- SOLUCIONA bodyclose
		var data []models.Sentence
		_ = json.NewDecoder(resp.Body).Decode(&data)
		c.HTML(http.StatusOK, "sentences.html", gin.H{"Sentences": data})
	}
}

func UpdateSentence(c *gin.Context) {
	var s models.Sentence
	id := c.Param("id")
	s.Spanish = Sanitize(c.PostForm("spanish"))
	s.English = Sanitize(c.PostForm("english"))

	// LA ADUANA: Validación robusta
	if err := ValidateSentence(s.English, s.Spanish); err != nil {
		SendToast(c, err.Error(), "error")
		return
	}

	// Si pasa, actualizamos en el repositorio
	err := repository.UpdateSentence(id, s.English, s.Spanish)
	if err != nil {
		SendToast(c, "Error al actualizar en la base de datos", "error")
		return
	}

	SendToast(c, "Frase actualizada correctamente", "success")
	c.Header("HX-Trigger", "refreshList") // Dispara recarga en el front
}

func DeleteSentence(c *gin.Context) {
	filter := "id=eq." + c.Param("id")
	resp, err := repository.CallSupabase("DELETE", "sentences", nil, filter)
	if err == nil && resp != nil {
		defer resp.Body.Close()
	}
	c.Status(http.StatusOK)
}

func ExportSentencesCSV(c *gin.Context) {
	var data []models.Sentence
	resp, err := repository.CallSupabase("GET", "sentences", nil, "select=*&order=id.asc")
	if err != nil || resp == nil {
		c.String(http.StatusInternalServerError, "Error al obtener frases")
		return
	}
	defer resp.Body.Close()

	_ = json.NewDecoder(resp.Body).Decode(&data)

	c.Header("Content-Disposition", "attachment; filename=sentences.csv")
	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()

	_ = writer.Write([]string{"ID", "English", "Spanish"})
	for _, s := range data {
		_ = writer.Write([]string{fmt.Sprintf("%d", s.ID), s.English, s.Spanish})
	}
}
