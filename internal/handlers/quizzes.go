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

func NewQuizForm(c *gin.Context) {
	c.HTML(http.StatusOK, "new-quiz.html", nil)
}

func ValidateQuiz(question string, options []string, correct string) error {
	if len(strings.TrimSpace(question)) < 10 {
		return fmt.Errorf("la pregunta es demasiado corta")
	}
	if len(options) < 3 {
		return fmt.Errorf("se requieren al menos 3 opciones")
	}
	for _, opt := range options {
		if strings.TrimSpace(opt) == "" {
			return fmt.Errorf("las opciones no pueden estar vacías")
		}
	}
	if correct == "" {
		return fmt.Errorf("debe marcar una respuesta como correcta")
	}
	return nil
}

func SaveQuiz(c *gin.Context) {
	// 1. Captura y Sanitizado
	question := Sanitize(c.PostForm("question"))
	options := []string{
		Sanitize(c.PostForm("opt1")),
		Sanitize(c.PostForm("opt2")),
		Sanitize(c.PostForm("opt3")),
	}
	correct := Sanitize(c.PostForm("correct"))

	// 2. Validación de lógica de negocio
	if err := ValidateQuiz(question, options, correct); err != nil {
		SendToast(c, err.Error(), "error")
		return
	}

	// 3. Guardado
	if err := repository.InsertQuiz(question, options, correct); err != nil {
		SendToast(c, "Error al crear el Quiz", "error")
		return
	}

	SendToast(c, "Quiz creado con éxito", "success")
	c.Header("HX-Trigger", "refreshList")
}

func GetQuizzes(c *gin.Context) {
	resp, err := repository.CallSupabase("GET", "quizzes", nil, "select=*&order=id.desc")
	if err != nil {
		c.String(http.StatusInternalServerError, "Error de conexión")
		return
	}

	if resp != nil {
		defer resp.Body.Close() // Cerramos el body para evitar fugas de memoria

		var quizzes []models.Quiz
		// Validamos el Decode para evitar procesar basura si la respuesta falla
		if err := json.NewDecoder(resp.Body).Decode(&quizzes); err != nil {
			c.String(http.StatusInternalServerError, "Error al procesar datos")
			return
		}

		c.HTML(http.StatusOK, "quizzes.html", gin.H{"Quizzes": quizzes})
		return
	}

	c.String(http.StatusNotFound, "No se encontraron datos")
}

func UpdateQuiz(c *gin.Context) {
	id := c.Param("id")

	// 1. Captura y Sanitizado
	question := Sanitize(c.PostForm("question"))
	options := []string{
		Sanitize(c.PostForm("opt1")),
		Sanitize(c.PostForm("opt1")),
		Sanitize(c.PostForm("opt1")),
	}
	correct := Sanitize(c.PostForm("correct"))

	// 2. Validación de la Aduana
	if err := ValidateQuiz(question, options, correct); err != nil {
		SendToast(c, err.Error(), "error")
		return
	}

	// 3. Persistencia
	err := repository.UpdateQuiz(id, question, options, correct)
	if err != nil {
		SendToast(c, "Error al actualizar el Quiz en Supabase", "error")
		return
	}

	SendToast(c, "Quiz actualizado correctamente", "success")
	c.Header("HX-Trigger", "refreshList")
}

func DeleteQuiz(c *gin.Context) {
	filter := "id=eq." + c.Param("id")
	resp, err := repository.CallSupabase("DELETE", "quizzes", nil, filter)
	if err == nil && resp != nil {
		resp.Body.Close()
	}
	c.Status(http.StatusOK)
}

func ExportQuizzesCSV(c *gin.Context) {
	var data []models.Quiz
	resp, err := repository.CallSupabase("GET", "quizzes", nil, "select=*")
	if err != nil || resp == nil {
		c.String(http.StatusInternalServerError, "Error al obtener datos para exportar")
		return
	}
	defer resp.Body.Close()

	// Decodificamos ignorando el error con _ o manejándolo
	_ = json.NewDecoder(resp.Body).Decode(&data)

	c.Header("Content-Disposition", "attachment; filename=quizzes_backup.csv")
	c.Header("Content-Type", "text/csv")

	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()

	// Todas las escrituras deben ser validadas o silenciadas
	_ = writer.Write([]string{"Pregunta", "Opción 1", "Opción 2", "Opción 3", "Correcta"})

	for _, q := range data {
		_ = writer.Write([]string{
			q.Question,
			q.Opt1,
			q.Opt2,
			q.Opt3,
			q.Correct,
		})
	}
}
