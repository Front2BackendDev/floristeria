package repository

import (
	"testing"
)

func TestFirstValue(t *testing.T) {
	m := map[string]any{
		"id":          nil,
		"firebase_id": "0DBiiUv29HBhVu9TqArM",
	}
	val := FirstValue(m, "id", "firebase_id")
	if val != "0DBiiUv29HBhVu9TqArM" {
		t.Errorf("se esperaba '0DBiiUv29HBhVu9TqArM', se obtuvo %q", val)
	}
}

func TestGenerateUniqueID(t *testing.T) {
	id1 := GenerateUniqueID()
	id2 := GenerateUniqueID()

	if len(id1) < 10 {
		t.Errorf("id1 demasiado corto: %s", id1)
	}
	if id1 == id2 {
		t.Errorf("se esperaba IDs únicos, se obtuvo idéntico: %s", id1)
	}
}

func TestParseNumeric(t *testing.T) {
	if parseNumeric(10) != 10 {
		t.Errorf("falló parse int")
	}
	if parseNumeric("15.5") != 15.5 {
		t.Errorf("falló parse string decimal")
	}
	if parseNumeric(nil) != 0 {
		t.Errorf("falló parse nil")
	}
}

func TestParseStringArray(t *testing.T) {
	arr := []any{"Rojo", "Blanco"}
	parsed := parseStringArray(arr)
	if len(parsed) != 2 || parsed[0] != "Rojo" || parsed[1] != "Blanco" {
		t.Errorf("falló parseStringArray: %v", parsed)
	}
}
