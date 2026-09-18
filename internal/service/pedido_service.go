package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"floristeria/internal/domain"
	"floristeria/internal/repository"
)

// PedidoService orquesta la lógica de negocio de los pedidos.
type PedidoService struct {
	repo repository.PedidoRepository
}

// NewPedidoService instancia un nuevo servicio inyectándole su repositorio.
func NewPedidoService(repo repository.PedidoRepository) *PedidoService {
	return &PedidoService{repo: repo}
}

func (s *PedidoService) HealthCheck(ctx context.Context) error {
	return s.repo.Ping(ctx)
}

func (s *PedidoService) ListarPedidos(ctx context.Context) ([]map[string]any, error) {
	return s.repo.List(ctx)
}

func (s *PedidoService) ObtenerPedido(ctx context.Context, id string) (map[string]any, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("id de pedido requerido")
	}
	return s.repo.GetByID(ctx, id)
}

func (s *PedidoService) CrearPedido(ctx context.Context, payload map[string]any) (map[string]any, error) {
	if len(payload) == 0 {
		return nil, fmt.Errorf("los datos del pedido no pueden estar vacíos")
	}

	// Validar y aplicar valores por defecto
	if val, ok := payload["estado"].(string); !ok || strings.TrimSpace(val) == "" {
		payload["estado"] = domain.EstadoPendiente
	}

	if val, ok := payload["fechaCreacion"].(string); !ok || strings.TrimSpace(val) == "" {
		payload["fechaCreacion"] = time.Now().UTC().Format(time.RFC3339)
	}

	return s.repo.Create(ctx, payload)
}

func (s *PedidoService) ActualizarEstado(ctx context.Context, id string, nuevoEstado string) (map[string]any, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("id de pedido requerido")
	}
	if strings.TrimSpace(nuevoEstado) == "" {
		return nil, fmt.Errorf("el nuevo estado no puede estar vacío")
	}

	payload := map[string]any{
		"estado": nuevoEstado,
	}
	return s.repo.Update(ctx, id, payload)
}

func (s *PedidoService) EliminarPedido(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("id de pedido requerido")
	}
	return s.repo.Delete(ctx, id)
}
