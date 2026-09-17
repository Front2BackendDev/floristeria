package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type server struct{ db *pgxpool.Pool }

func main() {
	ctx := context.Background()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is required")
	}
	db, err := pgxpool.New(ctx, dsn)
	if err != nil { log.Fatal(err) }
	defer db.Close()
	if err := db.Ping(ctx); err != nil { log.Fatal(err) }

	s := &server{db: db}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", s.health)
	mux.HandleFunc("/api/pedidos", s.pedidos)
	mux.HandleFunc("/api/pedidos/", s.pedido)
	mux.Handle("/", http.FileServer(http.Dir("karen")))

	port := os.Getenv("PORT")
	if port == "" { port = "8080" }
	log.Printf("floristeria escuchando en :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, logging(mux)))
}

func (s *server) health(w http.ResponseWriter, r *http.Request) {
	if err := s.db.Ping(r.Context()); err != nil { writeError(w, http.StatusServiceUnavailable, err); return }
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *server) pedidos(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		rows, err := s.db.Query(r.Context(), `SELECT row_to_json(p) FROM pedidos p ORDER BY COALESCE(to_jsonb(p)->>'fechaCreacion', to_jsonb(p)->>'fecha_creacion') DESC NULLS LAST`)
		if err != nil {
			// Some schemas use a different date column; keep reads useful in that case.
			rows, err = s.db.Query(r.Context(), `SELECT row_to_json(p) FROM pedidos p`)
		}
		if err != nil { writeError(w, http.StatusInternalServerError, err); return }
		defer rows.Close()
		out := make([]map[string]any, 0)
		for rows.Next() {
			var raw []byte
			if err := rows.Scan(&raw); err != nil { writeError(w, 500, err); return }
			var item map[string]any
			if err := json.Unmarshal(raw, &item); err != nil { writeError(w, 500, err); return }
			out = append(out, item)
		}
		if err := rows.Err(); err != nil { writeError(w, 500, err); return }
		if err := s.attachFlowerItems(r.Context(), out); err != nil { log.Printf("no se pudieron asociar items_flores: %v", err) }
		writeJSON(w, http.StatusOK, out)
	case http.MethodPost:
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil { writeError(w, 400, err); return }
		item, err := s.insert(r.Context(), "pedidos", payload)
		if err != nil { writeError(w, 500, err); return }
		writeJSON(w, http.StatusCreated, item)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *server) pedido(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/pedidos/")
	if id == "" { http.NotFound(w, r); return }
	switch r.Method {
	case http.MethodPatch, http.MethodPut:
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil { writeError(w, 400, err); return }
		cols, err := s.columns(r.Context(), "pedidos")
		if err != nil { writeError(w, 500, err); return }
		sets, args := make([]string, 0), []any{id}
		for key, value := range payload {
			col := resolveColumn(cols, key)
			if col == "" || col == "id" { continue }
			sets = append(sets, fmt.Sprintf(`"%s"=$%d`, col, len(args)+1)); args = append(args, value)
		}
		if len(sets) == 0 { writeError(w, 400, fmt.Errorf("no editable fields")); return }
		args = append(args, id)
		q := fmt.Sprintf(`UPDATE pedidos SET %s WHERE "id"=$%d RETURNING row_to_json(pedidos)`, strings.Join(sets, ","), len(args))
		var raw []byte
		if err := s.db.QueryRow(r.Context(), q, args...).Scan(&raw); err != nil { writeError(w, 404, err); return }
		var out map[string]any; _ = json.Unmarshal(raw, &out); writeJSON(w, 200, out)
	case http.MethodDelete:
		_, err := s.db.Exec(r.Context(), `DELETE FROM pedidos WHERE "id"=$1`, id)
		if err != nil { writeError(w, 500, err); return }
		w.WriteHeader(http.StatusNoContent)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *server) insert(ctx context.Context, table string, payload map[string]any) (map[string]any, error) {
	cols, err := s.columns(ctx, table); if err != nil { return nil, err }
	names, marks, args := make([]string, 0), make([]string, 0), make([]any, 0)
	for key, value := range payload {
		col := resolveColumn(cols, key); if col == "" || key == "id" { continue }
		names = append(names, `"`+col+`"`); args = append(args, value); marks = append(marks, fmt.Sprintf("$%d", len(args)))
	}
	if len(names) == 0 { return nil, fmt.Errorf("no compatible columns for %s", table) }
	q := fmt.Sprintf(`INSERT INTO "%s" (%s) VALUES (%s) RETURNING row_to_json("%s")`, table, strings.Join(names, ","), strings.Join(marks, ","), table)
	var raw []byte
	if err := s.db.QueryRow(ctx, q, args...).Scan(&raw); err != nil { return nil, err }
	var out map[string]any; return out, json.Unmarshal(raw, &out)
}

// Conserva la forma itemsFlores que esperaba el frontend cuando la migración
// mantiene esos registros en una tabla relacional separada.
func (s *server) attachFlowerItems(ctx context.Context, orders []map[string]any) error {
	var exists bool
	if err := s.db.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema='public' AND table_name='items_flores')`).Scan(&exists); err != nil || !exists { return nil }
	rows, err := s.db.Query(ctx, `SELECT row_to_json(f) FROM items_flores f`); if err != nil { return err }; defer rows.Close()
	byOrder := map[string][]map[string]any{}
	for rows.Next() {
		var raw []byte; if err := rows.Scan(&raw); err != nil { return err }
		var item map[string]any; if err := json.Unmarshal(raw, &item); err != nil { return err }
		fk := firstValue(item, "pedido_id", "pedidoId", "id_pedido", "pedido")
		if fk != "" { byOrder[fk] = append(byOrder[fk], item) }
	}
	for _, order := range orders { id := firstValue(order, "id", "pedido_id"); if items := byOrder[id]; id != "" && len(items) > 0 { order["itemsFlores"] = items } }
	return rows.Err()
}

func firstValue(values map[string]any, keys ...string) string {
	for _, key := range keys { if value, ok := values[key]; ok && value != nil && fmt.Sprint(value) != "" { return fmt.Sprint(value) } }
	return ""
}

func (s *server) columns(ctx context.Context, table string) (map[string]string, error) {
	rows, err := s.db.Query(ctx, `SELECT column_name FROM information_schema.columns WHERE table_schema='public' AND table_name=$1`, table)
	if err != nil { return nil, err }; defer rows.Close()
	cols := map[string]string{}
	for rows.Next() { var name string; if err := rows.Scan(&name); err != nil { return nil, err }; cols[name] = name }
	return cols, rows.Err()
}

func resolveColumn(cols map[string]string, key string) string {
	if _, ok := cols[key]; ok { return key }
	snake := camelToSnake(key); if _, ok := cols[snake]; ok { return snake }
	return ""
}
func camelToSnake(s string) string { var b strings.Builder; for i, r := range s { if r >= 'A' && r <= 'Z' { if i > 0 { b.WriteByte('_') }; b.WriteByte(byte(r-'A'+'a')) } else { b.WriteRune(r) } }; return b.String() }
func writeJSON(w http.ResponseWriter, status int, value any) { w.Header().Set("Content-Type", "application/json"); w.WriteHeader(status); _ = json.NewEncoder(w).Encode(value) }
func writeError(w http.ResponseWriter, status int, err error) { writeJSON(w, status, map[string]string{"error": err.Error()}) }
func logging(next http.Handler) http.Handler { return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { log.Printf("%s %s", r.Method, r.URL.Path); next.ServeHTTP(w, r) }) }
