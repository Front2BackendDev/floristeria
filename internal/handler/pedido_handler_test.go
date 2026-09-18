package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"floristeria/internal/service"
)

type mockRepo struct{}

func (m *mockRepo) List(ctx context.Context) ([]map[string]any, error) {
	return []map[string]any{{"id": "1", "estado": "pendiente"}}, nil
}
func (m *mockRepo) GetByID(ctx context.Context, id string) (map[string]any, error) {
	return map[string]any{"id": id, "estado": "pendiente"}, nil
}
func (m *mockRepo) Create(ctx context.Context, payload map[string]any) (map[string]any, error) {
	return payload, nil
}
func (m *mockRepo) Update(ctx context.Context, id string, payload map[string]any) (map[string]any, error) {
	return payload, nil
}
func (m *mockRepo) Delete(ctx context.Context, id string) error {
	return nil
}
func (m *mockRepo) Ping(ctx context.Context) error {
	return nil
}

func TestHandler_Health(t *testing.T) {
	svc := service.NewPedidoService(&mockRepo{})
	h := NewHandler(svc, "../../web")

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rec := httptest.NewRecorder()

	h.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("se esperaba status 200, se obtuvo %d", rec.Code)
	}
}

func TestHandler_PedidosList(t *testing.T) {
	svc := service.NewPedidoService(&mockRepo{})
	h := NewHandler(svc, "../../web")

	req := httptest.NewRequest(http.MethodGet, "/api/pedidos", nil)
	rec := httptest.NewRecorder()

	h.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("se esperaba status 200, se obtuvo %d", rec.Code)
	}
}
