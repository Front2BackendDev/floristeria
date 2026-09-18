package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PedidoRepository define el contrato para la persistencia de pedidos.
type PedidoRepository interface {
	List(ctx context.Context) ([]map[string]any, error)
	GetByID(ctx context.Context, id string) (map[string]any, error)
	Create(ctx context.Context, payload map[string]any) (map[string]any, error)
	Update(ctx context.Context, id string, payload map[string]any) (map[string]any, error)
	Delete(ctx context.Context, id string) error
	Ping(ctx context.Context) error
}

type postgresPedidoRepo struct {
	db *pgxpool.Pool
}

// NewPedidoRepository crea una nueva instancia del repositorio PostgreSQL.
func NewPedidoRepository(db *pgxpool.Pool) PedidoRepository {
	return &postgresPedidoRepo{db: db}
}

func (r *postgresPedidoRepo) Ping(ctx context.Context) error {
	return r.db.Ping(ctx)
}

func (r *postgresPedidoRepo) List(ctx context.Context) ([]map[string]any, error) {
	rows, err := r.db.Query(ctx, `SELECT row_to_json(p) FROM pedidos p ORDER BY COALESCE(to_jsonb(p)->>'fechaCreacion', to_jsonb(p)->>'fecha_creacion') DESC NULLS LAST`)
	if err != nil {
		rows, err = r.db.Query(ctx, `SELECT row_to_json(p) FROM pedidos p`)
	}
	if err != nil {
		return nil, fmt.Errorf("error listando pedidos: %w", err)
	}
	defer rows.Close()

	out := make([]map[string]any, 0)
	for rows.Next() {
		var raw []byte
		if err := rows.Scan(&raw); err != nil {
			return nil, err
		}
		var item map[string]any
		if err := json.Unmarshal(raw, &item); err != nil {
			return nil, err
		}
		if item["id"] == nil {
			item["id"] = FirstValue(item, "id", "pedido_id", "id_pedido")
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if err := r.attachFlowerItems(ctx, out); err != nil {
		log.Printf("repositorio: no se pudieron asociar items_flores: %v", err)
	}
	return out, nil
}

func (r *postgresPedidoRepo) GetByID(ctx context.Context, id string) (map[string]any, error) {
	cols, err := r.columns(ctx, "pedidos")
	if err != nil {
		return nil, err
	}
	idCol := FindIDColumn(cols)

	q := fmt.Sprintf(`SELECT row_to_json(p) FROM pedidos p WHERE "%s"::text = $1`, idCol)
	var raw []byte
	if err := r.db.QueryRow(ctx, q, id).Scan(&raw); err != nil {
		return nil, fmt.Errorf("pedido con id %s no encontrado: %w", id, err)
	}

	var item map[string]any
	if err := json.Unmarshal(raw, &item); err != nil {
		return nil, err
	}
	if item["id"] == nil {
		item["id"] = FirstValue(item, "id", "pedido_id", "id_pedido")
	}

	orders := []map[string]any{item}
	_ = r.attachFlowerItems(ctx, orders)
	return orders[0], nil
}

func (r *postgresPedidoRepo) Create(ctx context.Context, payload map[string]any) (map[string]any, error) {
	item, err := r.insert(ctx, "pedidos", payload)
	if err != nil {
		return nil, fmt.Errorf("error insertando pedido: %w", err)
	}

	if flowers, ok := payload["itemsFlores"].([]any); ok {
		if err := r.persistFlowerItems(ctx, item, flowers); err != nil {
			log.Printf("repositorio: no se pudieron guardar items_flores: %v", err)
		}
		if item["itemsFlores"] == nil {
			item["itemsFlores"] = flowers
		}
	}
	return item, nil
}

func (r *postgresPedidoRepo) Update(ctx context.Context, id string, payload map[string]any) (map[string]any, error) {
	cols, err := r.columns(ctx, "pedidos")
	if err != nil {
		return nil, err
	}
	idCol := FindIDColumn(cols)

	sets, args := make([]string, 0), make([]any, 0)
	for key, value := range payload {
		col := ResolveColumn(cols, key)
		if col == "" || col == idCol {
			continue
		}
		args = append(args, value)
		sets = append(sets, fmt.Sprintf(`"%s"=$%d`, col, len(args)))
	}
	if len(sets) == 0 {
		return nil, fmt.Errorf("no hay campos válidos para actualizar")
	}

	args = append(args, id)
	q := fmt.Sprintf(`UPDATE pedidos SET %s WHERE "%s"::text=$%d RETURNING row_to_json(pedidos)`,
		strings.Join(sets, ", "), idCol, len(args))

	var raw []byte
	if err := r.db.QueryRow(ctx, q, args...).Scan(&raw); err != nil {
		return nil, fmt.Errorf("error actualizando pedido %s: %w", id, err)
	}

	var out map[string]any
	_ = json.Unmarshal(raw, &out)
	if out["id"] == nil {
		out["id"] = FirstValue(out, "id", "pedido_id", "id_pedido")
	}
	return out, nil
}

func (r *postgresPedidoRepo) Delete(ctx context.Context, id string) error {
	cols, err := r.columns(ctx, "pedidos")
	if err != nil {
		return err
	}
	idCol := FindIDColumn(cols)

	// Eliminar detalle de items_flores primero para evitar violaciones de clave foránea
	flowerCols, err := r.columns(ctx, "items_flores")
	if err == nil && len(flowerCols) > 0 {
		fk := ResolveColumn(flowerCols, "pedido_id")
		if fk == "" {
			fk = ResolveColumn(flowerCols, "pedidoId")
		}
		if fk == "" {
			fk = ResolveColumn(flowerCols, "id_pedido")
		}
		if fk != "" {
			_, _ = r.db.Exec(ctx, fmt.Sprintf(`DELETE FROM items_flores WHERE "%s"::text=$1`, fk), id)
		}
	}

	_, err = r.db.Exec(ctx, fmt.Sprintf(`DELETE FROM pedidos WHERE "%s"::text=$1`, idCol), id)
	return err
}

func (r *postgresPedidoRepo) insert(ctx context.Context, table string, payload map[string]any) (map[string]any, error) {
	cols, err := r.columns(ctx, table)
	if err != nil {
		return nil, err
	}
	idCol := FindIDColumn(cols)
	names, marks, args := make([]string, 0), make([]string, 0), make([]any, 0)

	for key, value := range payload {
		col := ResolveColumn(cols, key)
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
		return nil, fmt.Errorf("no hay columnas compatibles para %s", table)
	}

	q := fmt.Sprintf(`INSERT INTO "%s" (%s) VALUES (%s) RETURNING row_to_json("%s")`,
		table, strings.Join(names, ", "), strings.Join(marks, ", "), table)

	var raw []byte
	if err := r.db.QueryRow(ctx, q, args...).Scan(&raw); err != nil {
		return nil, err
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	if out["id"] == nil {
		out["id"] = FirstValue(out, "id", "pedido_id", "id_pedido")
	}
	return out, nil
}

func (r *postgresPedidoRepo) attachFlowerItems(ctx context.Context, orders []map[string]any) error {
	var exists bool
	err := r.db.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema='public' AND table_name='items_flores')`).Scan(&exists)
	if err != nil || !exists {
		return nil
	}
	rows, err := r.db.Query(ctx, `SELECT row_to_json(f) FROM items_flores f`)
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
		fk := FirstValue(item, "pedido_id", "pedidoId", "id_pedido", "pedido")
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
		id := FirstValue(order, "id", "pedido_id", "id_pedido")
		if items, ok := byOrder[id]; ok && len(items) > 0 {
			order["itemsFlores"] = items
		}
	}
	return rows.Err()
}

func (r *postgresPedidoRepo) persistFlowerItems(ctx context.Context, order map[string]any, flowers []any) error {
	cols, err := r.columns(ctx, "items_flores")
	if err != nil || len(cols) == 0 {
		return err
	}
	orderID := FirstValue(order, "id", "pedido_id", "id_pedido")
	if orderID == "" {
		return nil
	}
	foreignKey := ResolveColumn(cols, "pedido_id")
	if foreignKey == "" {
		foreignKey = ResolveColumn(cols, "pedidoId")
	}
	if foreignKey == "" {
		foreignKey = ResolveColumn(cols, "id_pedido")
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
		if _, err := r.insert(ctx, "items_flores", payload); err != nil {
			log.Printf("repositorio: error guardando item_flor: %v", err)
		}
	}
	return nil
}

func (r *postgresPedidoRepo) columns(ctx context.Context, table string) (map[string]string, error) {
	rows, err := r.db.Query(ctx, `SELECT column_name FROM information_schema.columns WHERE table_schema='public' AND table_name=$1`, table)
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

// Helpers exportados para pruebas y uso en capas
func FindIDColumn(cols map[string]string) string {
	for _, candidate := range []string{"id", "pedido_id", "id_pedido", "pedidoId"} {
		if c, ok := cols[candidate]; ok {
			return c
		}
		snake := CamelToSnake(candidate)
		if c, ok := cols[snake]; ok {
			return c
		}
	}
	return "id"
}

func ResolveColumn(cols map[string]string, key string) string {
	if _, ok := cols[key]; ok {
		return key
	}
	snake := CamelToSnake(key)
	if _, ok := cols[snake]; ok {
		return snake
	}
	return ""
}

func CamelToSnake(s string) string {
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

func FirstValue(values map[string]any, keys ...string) string {
	for _, key := range keys {
		if value, ok := values[key]; ok && value != nil && fmt.Sprint(value) != "" {
			return fmt.Sprint(value)
		}
	}
	return ""
}
