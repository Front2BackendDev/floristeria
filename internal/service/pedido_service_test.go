package service

import (
	"context"
	"testing"

	"floristeria/internal/domain"
)

type mockRepo struct {
	createdPayload map[string]any
	updatedPayload map[string]any
	deletedID      string
}

func (m *mockRepo) List(ctx context.Context) ([]map[string]any, error) {
	return []map[string]any{{"id": "1", "estado": "pendiente"}}, nil
}
func (m *mockRepo) GetByID(ctx context.Context, id string) (map[string]any, error) {
	return map[string]any{"id": id, "estado": "pendiente"}, nil
}
func (m *mockRepo) Create(ctx context.Context, payload map[string]any) (map[string]any, error) {
	m.createdPayload = payload
	return payload, nil
}
func (m *mockRepo) Update(ctx context.Context, id string, payload map[string]any) (map[string]any, error) {
	m.updatedPayload = payload
	return payload, nil
}
func (m *mockRepo) Delete(ctx context.Context, id string) error {
	m.deletedID = id
	return nil
}
func (m *mockRepo) Ping(ctx context.Context) error {
	return nil
}

func TestPedidoService_CrearPedido(t *testing.T) {
	repo := &mockRepo{}
	svc := NewPedidoService(repo)

	payload := map[string]any{
		"nombreRemitente": "Carlos",
	}

	result, err := svc.CrearPedido(context.Background(), payload)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}

	if result["estado"] != domain.EstadoPendiente {
		t.Errorf("se esperaba estado %q, se obtuvo %q", domain.EstadoPendiente, result["estado"])
	}

	if result["fechaCreacion"] == nil || result["fechaCreacion"] == "" {
		t.Errorf("se esperaba que fechaCreacion fuera generada")
	}
}

func TestPedidoService_Validaciones(t *testing.T) {
	repo := &mockRepo{}
	svc := NewPedidoService(repo)

	// Id vacío en actualizar
	_, err := svc.ActualizarEstado(context.Background(), "", "completado")
	if err == nil {
		t.Errorf("se esperaba error con id vacío")
	}

	// Estado vacío en actualizar
	_, err = svc.ActualizarEstado(context.Background(), "123", "")
	if err == nil {
		t.Errorf("se esperaba error con estado vacío")
	}

	// Id vacío en eliminar
	err = svc.EliminarPedido(context.Background(), "")
	if err == nil {
		t.Errorf("se esperaba error con id vacío en eliminar")
	}
}
