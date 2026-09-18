package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"floristeria/internal/service"
)

// Handler agrupa las dependencias necesarias para responder a las rutas HTTP.
type Handler struct {
	service   *service.PedidoService
	staticDir string
}

// NewHandler crea una nueva instancia del Handler.
func NewHandler(svc *service.PedidoService, staticDir string) *Handler {
	return &Handler{
		service:   svc,
		staticDir: staticDir,
	}
}

// Routes registra y retorna el multiplexor HTTP con todas las rutas y middlewares.
func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()

	// Rutas de API REST
	mux.HandleFunc("/api/health", h.Health)
	mux.HandleFunc("/api/pedidos", h.Pedidos)
	mux.HandleFunc("/api/pedidos/", h.PedidoDetalle)

	// Servidor de archivos estáticos (Frontend en ./web)
	fileServer := http.FileServer(http.Dir(h.staticDir))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Verificar si el archivo solicitado existe
		path := filepath.Join(h.staticDir, filepath.Clean(r.URL.Path))
		if _, err := os.Stat(path); os.IsNotExist(err) {
			// Fallback a index.html para soportar navegación web
			http.ServeFile(w, r, filepath.Join(h.staticDir, "index.html"))
			return
		}
		fileServer.ServeHTTP(w, r)
	})

	return LoggingMiddleware(mux)
}

// LoggingMiddleware imprime cada petición entrante en consola.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func WriteJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func WriteError(w http.ResponseWriter, status int, message string) {
	WriteJSON(w, status, map[string]string{"error": message})
}
