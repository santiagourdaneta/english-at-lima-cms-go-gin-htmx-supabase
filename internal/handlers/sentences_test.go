package handlers

import "testing"

func TestSentenceLogic(t *testing.T) {
	// Tabla de pruebas: Entradas vs Resultados esperados
	tests := []struct {
		name        string
		text        string
		translation string
		wantErr     bool
	}{
		{"Frase válida", "Hello world", "Hola mundo", false},
		{"Frase muy corta", "Hi", "Hola", true},
		{"Traducción vacía", "Valid sentence", "", true},
		{"Ataque de desbordamiento", string(make([]byte, 1000)), "Error esperado", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSentence(tt.text, tt.translation)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSentence() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
