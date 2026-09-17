package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type server struct {
	db *pgxpool.Pool
}

func main() {
	ctx := context.Background()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is required")
	}

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		log.Fatalf("error configurando conexión PostgreSQL: %v", err)
	}
	config.MaxConns = 15
	config.MinConns = 2
	config.MaxConnLifetime = 30 * time.Minute
	config.MaxConnIdleTime = 5 * time.Minute

	db, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		log.Fatalf("error conectando a PostgreSQL: %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		log.Fatalf("error haciendo ping a PostgreSQL: %v", err)
	}

	s := &server{db: db}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", s.health)
	mux.HandleFunc("/api/pedidos", s.pedidos)
	mux.HandleFunc("/api/pedidos/", s.pedido)
	mux.Handle("/", http.FileServer(http.Dir("karen")))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("floristeria escuchando en :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, logging(mux)))
}

func (s *server) health(w http.ResponseWriter, r *http.Request) {
	if err := s.db.Ping(r.Context()); err != nil {
		writeError(w, http.StatusServiceUnavailable, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"status":   "ok",
		"database": "connected",
	})
}

func (s *server) pedidos(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		rows, err := s.db.Query(r.Context(), `SELECT row_to_json(p) FROM pedidos p ORDER BY COALESCE(to_jsonb(p)->>'fechaCreacion', to_jsonb(p)->>'fecha_creacion') DESC NULLS LAST`)
		if err != nil {
			rows, err = s.db.Query(r.Context(), `SELECT row_to_json(p) FROM pedidos p`)
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		defer rows.Close()

		out := make([]map[string]any, 0)
		for rows.Next() {
			var raw []byte
			if err := rows.Scan(&raw); err != nil {
				writeError(w, http.StatusInternalServerError, err)
				return
			}
			var item map[string]any
			if err := json.Unmarshal(raw, &item); err != nil {
				writeError(w, http.StatusInternalServerError, err)
				return
			}
			if item["id"] == nil {
				item["id"] = firstValue(item, "id", "pedido_id", "id_pedido")
			}
			out = append(out, item)
		}
		if err := rows.Err(); err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}

		if err := s.attachFlowerItems(r.Context(), out); err != nil {
			log.Printf("no se pudieron asociar items_flores: %v", err)
		}
		writeJSON(w, http.StatusOK, out)

	case http.MethodPost:
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		item, err := s.insert(r.Context(), "pedidos", payload)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		if flowers, ok := payload["itemsFlores"].([]any); ok {
			if err := s.persistFlowerItems(r.Context(), item, flowers); err != nil {
				log.Printf("no se pudieron guardar items_flores: %v", err)
			}
			if item["itemsFlores"] == nil {
				item["itemsFlores"] = flowers
			}
		}
		writeJSON(w, http.StatusCreated, item)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *server) pedido(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/pedidos/")
	if id == "" {
		http.NotFound(w, r)
		return
	}

	cols, err := s.columns(r.Context(), "pedidos")
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	idCol := findIDColumn(cols)

	switch r.Method {
	case http.MethodPatch, http.MethodPut:
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}

		sets, args := make([]string, 0), make([]any, 0)
		for key, value := range payload {
			col := resolveColumn(cols, key)
			if col == "" || col == idCol {
				continue
			}
			args = append(args, value)
			sets = append(sets, fmt.Sprintf(`"%s"=$%d`, col, len(args)))
		}
		if len(sets) == 0 {
			writeError(w, http.StatusBadRequest, fmt.Errorf("no editable fields"))
			return
		}

		args = append(args, id)
		q := fmt.Sprintf(`UPDATE pedidos SET %s WHERE "%s"::text=$%d RETURNING row_to_json(pedidos)`,
			strings.Join(sets, ", "), idCol, len(args))

		var raw []byte
		if err := s.db.QueryRow(r.Context(), q, args...).Scan(&raw); err != nil {
			writeError(w, http.StatusNotFound, err)
			return
		}
		var out map[string]any
		_ = json.Unmarshal(raw, &out)
		if out["id"] == nil {
			out["id"] = firstValue(out, "id", "pedido_id", "id_pedido")
		}
		writeJSON(w, http.StatusOK, out)

	case http.MethodDelete:
		// Eliminar detalles de items_flores primero por integridad referencial
		flowerCols, err := s.columns(r.Context(), "items_flores")
		if err == nil && len(flowerCols) > 0 {
			fk := resolveColumn(flowerCols, "pedido_id")
			if fk == "" {
				fk = resolveColumn(flowerCols, "pedidoId")
			}
			if fk == "" {
				fk = resolveColumn(flowerCols, "id_pedido")
			}
			if fk != "" {
				_, _ = s.db.Exec(r.Context(), fmt.Sprintf(`DELETE FROM items_flores WHERE "%s"::text=$1`, fk), id)
			}
		}

		_, err = s.db.Exec(r.Context(), fmt.Sprintf(`DELETE FROM pedidos WHERE "%s"::text=$1`, idCol), id)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *server) insert(ctx context.Context, table string, payload map[string]any) (map[string]any, error) {
	cols, err := s.columns(ctx, table)
	if err != nil {
		return nil, err
	}
	idCol := findIDColumn(cols)
	names, marks, args := make([]string, 0), make([]string, 0), make([]any, 0)

	for key, value := range payload {
		col := resolveColumn(cols, key)
		if col == "" || (col == idCol && value == nil) {
			continue
		}

		var finalVal any = value
		switch v := value.(type) {
		case []any, map[string]any:
			if b, err := json.Marshal(v); err == nil {
				finalVal = string(b)
			}
		}

		args = append(args, finalVal)
		names = append(names, `"`+col+`"`)
		marks = append(marks, fmt.Sprintf("$%d", len(args)))
	}

	if len(names) == 0 {
		return nil, fmt.Errorf("no compatible columns for %s", table)
	}

	q := fmt.Sprintf(`INSERT INTO "%s" (%s) VALUES (%s) RETURNING row_to_json("%s")`,
		table, strings.Join(names, ", "), strings.Join(marks, ", "), table)

	var raw []byte
	if err := s.db.QueryRow(ctx, q, args...).Scan(&raw); err != nil {
		return nil, err
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	if out["id"] == nil {
		out["id"] = firstValue(out, "id", "pedido_id", "id_pedido")
	}
	return out, nil
}

func (s *server) attachFlowerItems(ctx context.Context, orders []map[string]any) error {
	var exists bool
	err := s.db.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema='public' AND table_name='items_flores')`).Scan(&exists)
	if err != nil || !exists {
		return nil
	}
	rows, err := s.db.Query(ctx, `SELECT row_to_json(f) FROM items_flores f`)
	if err != nil {
		return err
	}
	defer rows.Close()

	byOrder := map[string][]map[string]any{}
	for rows.Next() {
		var raw []byte
		if err := rows.Scan(&raw); err != nil {
			return err
		}
		var item map[string]any
		if err := json.Unmarshal(raw, &item); err != nil {
			continue
		}
		fk := firstValue(item, "pedido_id", "pedidoId", "id_pedido", "pedido")
		if fk != "" {
			if coloresStr, ok := item["colores"].(string); ok && strings.HasPrefix(coloresStr, "[") {
				var coloresArr []any
				if err := json.Unmarshal([]byte(coloresStr), &coloresArr); err == nil {
					item["colores"] = coloresArr
				}
			}
			byOrder[fk] = append(byOrder[fk], item)
		}
	}

	for _, order := range orders {
		id := firstValue(order, "id", "pedido_id", "id_pedido")
		if items, ok := byOrder[id]; ok && len(items) > 0 {
			order["itemsFlores"] = items
		}
	}
	return rows.Err()
}

func (s *server) persistFlowerItems(ctx context.Context, order map[string]any, flowers []any) error {
	cols, err := s.columns(ctx, "items_flores")
	if err != nil || len(cols) == 0 {
		return err
	}
	orderID := firstValue(order, "id", "pedido_id", "id_pedido")
	if orderID == "" {
		return nil
	}
	foreignKey := resolveColumn(cols, "pedido_id")
	if foreignKey == "" {
		foreignKey = resolveColumn(cols, "pedidoId")
	}
	if foreignKey == "" {
		foreignKey = resolveColumn(cols, "id_pedido")
	}
	if foreignKey == "" {
		return nil
	}

	for _, raw := range flowers {
		flower, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		payload := map[string]any{foreignKey: orderID}
		for key, value := range flower {
			payload[key] = value
		}
		if _, err := s.insert(ctx, "items_flores", payload); err != nil {
			log.Printf("error guardando item_flor: %v", err)
		}
	}
	return nil
}

func (s *server) columns(ctx context.Context, table string) (map[string]string, error) {
	rows, err := s.db.Query(ctx, `SELECT column_name FROM information_schema.columns WHERE table_schema='public' AND table_name=$1`, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols := map[string]string{}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		cols[name] = name
	}
	return cols, rows.Err()
}

func findIDColumn(cols map[string]string) string {
	for _, candidate := range []string{"id", "pedido_id", "id_pedido", "pedidoId"} {
		if c, ok := cols[candidate]; ok {
			return c
		}
		snake := camelToSnake(candidate)
		if c, ok := cols[snake]; ok {
			return c
		}
	}
	return "id"
}

func resolveColumn(cols map[string]string, key string) string {
	if _, ok := cols[key]; ok {
		return key
	}
	snake := camelToSnake(key)
	if _, ok := cols[snake]; ok {
		return snake
	}
	return ""
}

func camelToSnake(s string) string {
	var b strings.Builder
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				b.WriteByte('_')
			}
			b.WriteByte(byte(r - 'A' + 'a'))
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func firstValue(values map[string]any, keys ...string) string {
	for _, key := range keys {
		if value, ok := values[key]; ok && value != nil && fmt.Sprint(value) != "" {
			return fmt.Sprint(value)
		}
	}
	return ""
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
