package main

import (
	"testing"
)

// TestUnitario: Verifica que los structs funcionen bien
func TestSentenceModel(t *testing.T) {
	s := Sentence{English: "Hello", Spanish: "Hola"}
	if s.English != "Hello" {
		t.Errorf("Esperaba 'Hello', obtuve %s", s.English)
	}
}