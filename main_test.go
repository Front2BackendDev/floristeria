package main

import (
	"testing"
)

func TestCamelToSnake(t *testing.T) {
	tests := map[string]string{
		"nombreRemitente":     "nombre_remitente",
		"fechaCreacion":       "fecha_creacion",
		"itemsFlores":         "items_flores",
		"numeroWhatsAppUsado": "numero_whats_app_usado",
		"estado":              "estado",
	}

	for in, expected := range tests {
		got := camelToSnake(in)
		if got != expected {
			t.Errorf("camelToSnake(%q) = %q; want %q", in, got, expected)
		}
	}
}

func TestResolveColumn(t *testing.T) {
	cols := map[string]string{
		"nombre_remitente": "nombre_remitente",
		"estado":           "estado",
		"fechaCreacion":    "fechaCreacion",
	}

	if resolveColumn(cols, "nombreRemitente") != "nombre_remitente" {
		t.Errorf("failed resolving snake_case column")
	}
	if resolveColumn(cols, "fechaCreacion") != "fechaCreacion" {
		t.Errorf("failed resolving camelCase column")
	}
	if resolveColumn(cols, "inexistente") != "" {
		t.Errorf("expected empty string for non-existent column")
	}
}

func TestFindIDColumn(t *testing.T) {
	cols1 := map[string]string{"id": "id", "nombre": "nombre"}
	if findIDColumn(cols1) != "id" {
		t.Errorf("expected 'id'")
	}

	cols2 := map[string]string{"pedido_id": "pedido_id", "nombre": "nombre"}
	if findIDColumn(cols2) != "pedido_id" {
		t.Errorf("expected 'pedido_id'")
	}

	cols3 := map[string]string{"id_pedido": "id_pedido", "nombre": "nombre"}
	if findIDColumn(cols3) != "id_pedido" {
		t.Errorf("expected 'id_pedido'")
	}
}

func TestFirstValue(t *testing.T) {
	m := map[string]any{
		"id":        nil,
		"pedido_id": 123,
	}
	val := firstValue(m, "id", "pedido_id")
	if val != "123" {
		t.Errorf("expected '123', got %q", val)
	}
}
