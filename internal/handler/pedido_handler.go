package handler

import (
	"encoding/json"
	"net/http"
	"strings"
)

// Health comprueba el estado del servicio y la conexión a la base de datos.
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	if err := h.service.HealthCheck(r.Context()); err != nil {
		WriteError(w, http.StatusServiceUnavailable, "base de datos no disponible: "+err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{
		"status":   "ok",
		"database": "connected",
	})
}

// Pedidos atiende peticiones de listado y creación masiva.
func (h *Handler) Pedidos(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		pedidos, err := h.service.ListarPedidos(r.Context())
		if err != nil {
			WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}
		WriteJSON(w, http.StatusOK, pedidos)

	case http.MethodPost:
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			WriteError(w, http.StatusBadRequest, "formato JSON inválido: "+err.Error())
			return
		}

		creado, err := h.service.CrearPedido(r.Context(), payload)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}
		WriteJSON(w, http.StatusCreated, creado)

	default:
		w.Header().Set("Allow", "GET, POST")
		WriteError(w, http.StatusMethodNotAllowed, "método no permitido")
	}
}

// PedidoDetalle atiende operaciones individuales por ID (PATCH/PUT y DELETE).
func (h *Handler) PedidoDetalle(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/pedidos/")
	id = strings.TrimSpace(id)
	if id == "" {
		http.NotFound(w, r)
		return
	}

	switch r.Method {
	case http.MethodGet:
		pedido, err := h.service.ObtenerPedido(r.Context(), id)
		if err != nil {
			WriteError(w, http.StatusNotFound, err.Error())
			return
		}
		WriteJSON(w, http.StatusOK, pedido)

	case http.MethodPatch, http.MethodPut:
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			WriteError(w, http.StatusBadRequest, "formato JSON inválido: "+err.Error())
			return
		}

		// Si viene el campo "estado", actualizamos estado directamente
		if nuevoEstado, ok := payload["estado"].(string); ok && len(payload) == 1 {
			actualizado, err := h.service.ActualizarEstado(r.Context(), id, nuevoEstado)
			if err != nil {
				WriteError(w, http.StatusBadRequest, err.Error())
				return
			}
			WriteJSON(w, http.StatusOK, actualizado)
			return
		}

		// Si vienen otros campos
		actualizado, err := h.service.ActualizarEstado(r.Context(), id, payload["estado"].(string))
		if err != nil {
			WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
		WriteJSON(w, http.StatusOK, actualizado)

	case http.MethodDelete:
		if err := h.service.EliminarPedido(r.Context(), id); err != nil {
			WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}
		w.WriteHeader(http.StatusNoContent)

	default:
		w.Header().Set("Allow", "GET, PATCH, PUT, DELETE")
		WriteError(w, http.StatusMethodNotAllowed, "método no permitido")
	}
}
