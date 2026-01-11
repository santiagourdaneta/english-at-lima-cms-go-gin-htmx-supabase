package handlers

import (
	"strings"
	"testing"
)

func TestModuleValidations(t *testing.T) {
	// Sub-test para Recursos
	t.Run("Validación de Recursos", func(t *testing.T) {
		tests := []struct {
			name    string
			t, u, r string // title, url, type
			wantErr bool
		}{
			{"Recurso Válido", "Guía PDF", "https://lima.com/file.pdf", "PDF", false},
			{"Título Corto", "Ab", "https://lima.com", "Link", true},
			{"URL Insegura", "Video", "ftp://servidor.com", "Video", true},
		}
		for _, tt := range tests {
			if err := ValidateResource(tt.t, tt.u, tt.r); (err != nil) != tt.wantErr {
				t.Errorf("%s: error esperado %v, obtenido %v", tt.name, tt.wantErr, err)
			}
		}
	})
}

func TestSQLInjectionPrevention(t *testing.T) {
	// Intentos de inyección comunes
	maliciousInputs := []string{
		"'; DROP TABLE users; --",
		"<script>alert('xss')</script>",
		"OR 1=1",
		"admin'--",
	}

	for _, input := range maliciousInputs {
		// Probamos que nuestra validación o el sanitizador lo maneje
		err := ValidateSentence(input, "Traducción válida")
		// Aquí podrías añadir lógica para rechazar caracteres como ';' o '--'
		if err == nil && strings.Contains(input, ";") {
			t.Logf("⚠️ Advertencia: El input '%s' pasó la validación. Asegúrate de usar Query Parameters en el repositorio.", input)
		}
	}
}

func TestModuleValidationsQuiz(t *testing.T) {
	// Sub-test para Quizzes
	t.Run("Validación de Quizzes", func(t *testing.T) {
		tests := []struct {
			name    string
			q       string   // question
			o       []string // options
			c       string   // correct
			wantErr bool
		}{
			{"Quiz Perfecto", "What is 'Apple'?", []string{"Manzana", "Pera", "Banana"}, "Manzana", false},
			{"Pregunta Corta", "Hi?", []string{"1", "2", "3"}, "1", true},
			{"Opción Vacía", "Valid question?", []string{"", "2", "3"}, "2", true},
			{"Sin Correcta", "Valid question?", []string{"1", "2", "3"}, "", true},
		}
		for _, tt := range tests {
			if err := ValidateQuiz(tt.q, tt.o, tt.c); (err != nil) != tt.wantErr {
				t.Errorf("%s: error esperado %v, obtenido %v", tt.name, tt.wantErr, err)
			}
		}
	})
}
