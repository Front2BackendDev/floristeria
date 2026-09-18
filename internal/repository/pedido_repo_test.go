package repository

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
		got := CamelToSnake(in)
		if got != expected {
			t.Errorf("CamelToSnake(%q) = %q; want %q", in, got, expected)
		}
	}
}

func TestResolveColumn(t *testing.T) {
	cols := map[string]string{
		"nombre_remitente": "nombre_remitente",
		"estado":           "estado",
		"fechaCreacion":    "fechaCreacion",
	}

	if ResolveColumn(cols, "nombreRemitente") != "nombre_remitente" {
		t.Errorf("falló al resolver columna en snake_case")
	}
	if ResolveColumn(cols, "fechaCreacion") != "fechaCreacion" {
		t.Errorf("falló al resolver columna en camelCase")
	}
	if ResolveColumn(cols, "inexistente") != "" {
		t.Errorf("se esperaba cadena vacía para columna inexistente")
	}
}

func TestFindIDColumn(t *testing.T) {
	cols1 := map[string]string{"id": "id", "nombre": "nombre"}
	if FindIDColumn(cols1) != "id" {
		t.Errorf("se esperaba 'id'")
	}

	cols2 := map[string]string{"pedido_id": "pedido_id", "nombre": "nombre"}
	if FindIDColumn(cols2) != "pedido_id" {
		t.Errorf("se esperaba 'pedido_id'")
	}

	cols3 := map[string]string{"id_pedido": "id_pedido", "nombre": "nombre"}
	if FindIDColumn(cols3) != "id_pedido" {
		t.Errorf("se esperaba 'id_pedido'")
	}
}

func TestFirstValue(t *testing.T) {
	m := map[string]any{
		"id":        nil,
		"pedido_id": 123,
	}
	val := FirstValue(m, "id", "pedido_id")
	if val != "123" {
		t.Errorf("se esperaba '123', se obtuvo %q", val)
	}
}
