package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"floristeria/internal/handler"
	"floristeria/internal/repository"
	"floristeria/internal/service"
)

func main() {
	ctx := context.Background()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("variable de entorno DATABASE_URL es requerida")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// 1. Capa de Persistencia (PostgreSQL)
	pool, err := repository.NewPostgresPool(ctx, dsn)
	if err != nil {
		log.Fatalf("error inicializando base de datos: %v", err)
	}
	defer pool.Close()

	repo := repository.NewPedidoRepository(pool)

	// 2. Capa de Servicio (Lógica de negocio)
	svc := service.NewPedidoService(repo)

	// 3. Capa de Handler / Rutas HTTP (sirve frontend desde 'web')
	h := handler.NewHandler(svc, "web")

	// 4. Iniciar servidor HTTP
	log.Printf("floristeria escuchando en :%s", port)
	if err := http.ListenAndServe(":"+port, h.Routes()); err != nil {
		log.Fatalf("servidor detenido con error: %v", err)
	}
}
