package main

import (
	"testing"
)

type Sentence struct {
	English string
	Spanish string
}

// TestUnitario: Verifica que los structs funcionen bien
func TestSentenceModel(t *testing.T) {

	s := Sentence{Spanish: "Hola"}
	if s.Spanish != "Hola" {
		t.Errorf("Esperaba 'Hola', obtuve %s", s.Spanish)
	}
}
