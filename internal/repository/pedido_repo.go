package repository

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

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
	// Consulta combinada: recupera datos_originales y columnas actualizadas
	q := `
		SELECT 
			firebase_id,
			COALESCE(estado, datos_originales->>'estado', 'pendiente') AS estado,
			COALESCE(fecha_entrega_texto, datos_originales->>'fechaEntrega', '') AS fecha_entrega_texto,
			COALESCE(hora_entrega_texto, datos_originales->>'horaEntrega', '') AS hora_entrega_texto,
			COALESCE(fecha_creacion_texto, datos_originales->>'fechaCreacion', '') AS fecha_creacion_texto,
			COALESCE(monto_total_texto, datos_originales->>'montoTotal', '0') AS monto_total_texto,
			COALESCE(monto_anticipo_texto, datos_originales->>'montoAnticipo', '0') AS monto_anticipo_texto,
			datos_originales
		FROM pedidos
		ORDER BY COALESCE(fecha_creacion_texto, datos_originales->>'fechaCreacion') DESC NULLS LAST;
	`
	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("error listando pedidos: %w", err)
	}
	defer rows.Close()

	out := make([]map[string]any, 0)
	for rows.Next() {
		var firebaseID, estado, fechaEntrega, horaEntrega, fechaCreacion, montoTotal, montoAnticipo string
		var datosOriginalesRaw []byte

		if err := rows.Scan(
			&firebaseID,
			&estado,
			&fechaEntrega,
			&horaEntrega,
			&fechaCreacion,
			&montoTotal,
			&montoAnticipo,
			&datosOriginalesRaw,
		); err != nil {
			return nil, err
		}

		item := make(map[string]any)
		if len(datosOriginalesRaw) > 0 {
			_ = json.Unmarshal(datosOriginalesRaw, &item)
		}

		// Asegurar sincronización de campos clave con la base de datos
		item["id"] = firebaseID
		item["firebase_id"] = firebaseID
		item["estado"] = estado
		if item["fechaEntrega"] == nil || item["fechaEntrega"] == "" {
			item["fechaEntrega"] = fechaEntrega
		}
		if item["horaEntrega"] == nil || item["horaEntrega"] == "" {
			item["horaEntrega"] = horaEntrega
		}
		if item["fechaCreacion"] == nil || item["fechaCreacion"] == "" {
			item["fechaCreacion"] = fechaCreacion
		}
		if item["montoTotal"] == nil || item["montoTotal"] == "" {
			item["montoTotal"] = montoTotal
		}
		if item["montoAnticipo"] == nil || item["montoAnticipo"] == "" {
			item["montoAnticipo"] = montoAnticipo
		}

		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Cargar y asociar flores de items_flores
	if err := r.attachFlowerItems(ctx, out); err != nil {
		log.Printf("repositorio: advertencia asociando items_flores: %v", err)
	}

	return out, nil
}

func (r *postgresPedidoRepo) GetByID(ctx context.Context, id string) (map[string]any, error) {
	q := `
		SELECT 
			firebase_id,
			COALESCE(estado, datos_originales->>'estado', 'pendiente') AS estado,
			COALESCE(fecha_entrega_texto, datos_originales->>'fechaEntrega', '') AS fecha_entrega_texto,
			COALESCE(hora_entrega_texto, datos_originales->>'horaEntrega', '') AS hora_entrega_texto,
			COALESCE(fecha_creacion_texto, datos_originales->>'fechaCreacion', '') AS fecha_creacion_texto,
			COALESCE(monto_total_texto, datos_originales->>'montoTotal', '0') AS monto_total_texto,
			COALESCE(monto_anticipo_texto, datos_originales->>'montoAnticipo', '0') AS monto_anticipo_texto,
			datos_originales
		FROM pedidos
		WHERE firebase_id = $1;
	`
	var firebaseID, estado, fechaEntrega, horaEntrega, fechaCreacion, montoTotal, montoAnticipo string
	var datosOriginalesRaw []byte

	err := r.db.QueryRow(ctx, q, id).Scan(
		&firebaseID,
		&estado,
		&fechaEntrega,
		&horaEntrega,
		&fechaCreacion,
		&montoTotal,
		&montoAnticipo,
		&datosOriginalesRaw,
	)
	if err != nil {
		return nil, fmt.Errorf("pedido %s no encontrado: %w", id, err)
	}

	item := make(map[string]any)
	if len(datosOriginalesRaw) > 0 {
		_ = json.Unmarshal(datosOriginalesRaw, &item)
	}

	item["id"] = firebaseID
	item["firebase_id"] = firebaseID
	item["estado"] = estado
	if item["fechaEntrega"] == nil || item["fechaEntrega"] == "" {
		item["fechaEntrega"] = fechaEntrega
	}
	if item["horaEntrega"] == nil || item["horaEntrega"] == "" {
		item["horaEntrega"] = horaEntrega
	}
	if item["fechaCreacion"] == nil || item["fechaCreacion"] == "" {
		item["fechaCreacion"] = fechaCreacion
	}
	if item["montoTotal"] == nil || item["montoTotal"] == "" {
		item["montoTotal"] = montoTotal
	}
	if item["montoAnticipo"] == nil || item["montoAnticipo"] == "" {
		item["montoAnticipo"] = montoAnticipo
	}

	orders := []map[string]any{item}
	_ = r.attachFlowerItems(ctx, orders)
	return orders[0], nil
}

func (r *postgresPedidoRepo) Create(ctx context.Context, payload map[string]any) (map[string]any, error) {
	// 1. Generar ID único para firebase_id
	id := FirstValue(payload, "id", "firebase_id")
	if id == "" {
		id = GenerateUniqueID()
	}
	payload["id"] = id
	payload["firebase_id"] = id

	// 2. Extraer campos
	barrio := FirstValue(payload, "barrio")
	celularDest := FirstValue(payload, "celularDestinatario", "celular_destinatario")
	celularRem := FirstValue(payload, "celularRemitente", "celular_remitente")
	ciudad := FirstValue(payload, "ciudad")
	descripcion := FirstValue(payload, "descripcionPedido", "descripcion_pedido")
	direccion := FirstValue(payload, "direccion")
	estado := FirstValue(payload, "estado")
	if estado == "" {
		estado = "pendiente"
	}
	fechaCreacion := FirstValue(payload, "fechaCreacion", "fecha_creacion")
	if fechaCreacion == "" {
		fechaCreacion = time.Now().UTC().Format(time.RFC3339)
	}
	fechaEntrega := FirstValue(payload, "fechaEntrega", "fecha_entrega")
	horaEntrega := FirstValue(payload, "horaEntrega", "hora_entrega")
	medioPago := FirstValue(payload, "medioPago", "medio_pago")
	mensajeTarjeta := FirstValue(payload, "mensajeTarjeta", "mensaje_tarjeta")
	metodoAnticipo := FirstValue(payload, "metodoPagoAnticipo", "metodo_pago_anticipo")
	montoAnticipo := FirstValue(payload, "montoAnticipo", "monto_anticipo")
	montoTotal := FirstValue(payload, "montoTotal", "monto_total")
	nombreDest := FirstValue(payload, "nombreDestinatario", "nombre_destinatario")
	nombreRem := FirstValue(payload, "nombreRemitente", "nombre_remitente")
	numeroWhatsApp := FirstValue(payload, "numeroWhatsAppUsado", "numero_whatsapp_usado")
	pais := FirstValue(payload, "pais")
	puntoRef := FirstValue(payload, "puntoReferencia", "punto_referencia")

	// 3. Serializar datos_originales
	datosOriginales, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("error serializando datos_originales: %w", err)
	}

	// 4. Inserción en pedidos
	insertQ := `
		INSERT INTO pedidos (
			firebase_id, barrio, celular_destinatario, celular_remitente,
			ciudad, descripcion_pedido, direccion, estado,
			fecha_creacion_texto, fecha_entrega_texto, hora_entrega_texto,
			medio_pago, mensaje_tarjeta, metodo_pago_anticipo,
			monto_anticipo_texto, monto_total_texto, nombre_destinatario,
			nombre_remitente, numero_whatsapp_usado, pais,
			punto_referencia, datos_originales, migrado_en
		) VALUES (
			$1, $2, $3, $4,
			$5, $6, $7, $8,
			$9, $10, $11,
			$12, $13, $14,
			$15, $16, $17,
			$18, $19, $20,
			$21, $22, now()
		);
	`
	_, err = r.db.Exec(ctx, insertQ,
		id, barrio, celularDest, celularRem,
		ciudad, descripcion, direccion, estado,
		fechaCreacion, fechaEntrega, horaEntrega,
		medioPago, mensajeTarjeta, metodoAnticipo,
		montoAnticipo, montoTotal, nombreDest,
		nombreRem, numeroWhatsApp, pais,
		puntoRef, datosOriginales,
	)
	if err != nil {
		return nil, fmt.Errorf("error insertando pedido en postgres: %w", err)
	}

	// 5. Inserción en items_flores
	if flowers, ok := payload["itemsFlores"].([]any); ok && len(flowers) > 0 {
		for i, raw := range flowers {
			flower, ok := raw.(map[string]any)
			if !ok {
				continue
			}

			cant := parseNumeric(flower["cantidad"])
			tipo := FirstValue(flower, "tipo")
			emoji := FirstValue(flower, "emoji")
			colores := parseStringArray(flower["colores"])

			flowerJSON, _ := json.Marshal(flower)

			flowerInsertQ := `
				INSERT INTO items_flores (
					pedido_firebase_id, posicion, cantidad,
					colores, emoji, tipo, datos_originales
				) VALUES ($1, $2, $3, $4, $5, $6, $7);
			`
			_, err := r.db.Exec(ctx, flowerInsertQ, id, i, cant, colores, emoji, tipo, flowerJSON)
			if err != nil {
				log.Printf("repositorio: advertencia insertando flor #%d: %v", i, err)
			}
		}
	}

	return payload, nil
}

func (r *postgresPedidoRepo) Update(ctx context.Context, id string, payload map[string]any) (map[string]any, error) {
	nuevoEstado, hasEstado := payload["estado"].(string)
	if !hasEstado || strings.TrimSpace(nuevoEstado) == "" {
		return nil, fmt.Errorf("actualización requiere campo 'estado'")
	}

	q := `
		UPDATE pedidos 
		SET estado = $1, 
		    datos_originales = jsonb_set(COALESCE(datos_originales, '{}'::jsonb), '{estado}', to_jsonb($1::text))
		WHERE firebase_id = $2
		RETURNING firebase_id, estado, datos_originales;
	`
	var firebaseID, estado string
	var datosOriginalesRaw []byte

	err := r.db.QueryRow(ctx, q, nuevoEstado, id).Scan(&firebaseID, &estado, &datosOriginalesRaw)
	if err != nil {
		return nil, fmt.Errorf("error actualizando estado de pedido %s: %w", id, err)
	}

	item := make(map[string]any)
	if len(datosOriginalesRaw) > 0 {
		_ = json.Unmarshal(datosOriginalesRaw, &item)
	}
	item["id"] = firebaseID
	item["firebase_id"] = firebaseID
	item["estado"] = estado

	return item, nil
}

func (r *postgresPedidoRepo) Delete(ctx context.Context, id string) error {
	// 1. Eliminar items de flores asociados
	_, _ = r.db.Exec(ctx, `DELETE FROM items_flores WHERE pedido_firebase_id = $1`, id)

	// 2. Eliminar pedido
	cmdTag, err := r.db.Exec(ctx, `DELETE FROM pedidos WHERE firebase_id = $1`, id)
	if err != nil {
		return fmt.Errorf("error eliminando pedido %s: %w", id, err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("pedido %s no encontrado", id)
	}
	return nil
}

func (r *postgresPedidoRepo) attachFlowerItems(ctx context.Context, orders []map[string]any) error {
	q := `
		SELECT 
			pedido_firebase_id,
			posicion,
			COALESCE(cantidad, 0) AS cantidad,
			colores,
			COALESCE(emoji, '🌸') AS emoji,
			COALESCE(tipo, '') AS tipo,
			datos_originales
		FROM items_flores
		ORDER BY posicion ASC;
	`
	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return err
	}
	defer rows.Close()

	byOrder := make(map[string][]map[string]any)
	for rows.Next() {
		var pedidoFirebaseID string
		var posicion int
		var cantidad float64
		var colores []string
		var emoji, tipo string
		var datosOriginalesRaw []byte

		if err := rows.Scan(
			&pedidoFirebaseID,
			&posicion,
			&cantidad,
			&colores,
			&emoji,
			&tipo,
			&datosOriginalesRaw,
		); err != nil {
			return err
		}

		item := make(map[string]any)
		if len(datosOriginalesRaw) > 0 {
			_ = json.Unmarshal(datosOriginalesRaw, &item)
		}

		// Normalizar claves si no estaban en datos_originales
		if item["tipo"] == nil || item["tipo"] == "" {
			item["tipo"] = tipo
		}
		if item["emoji"] == nil || item["emoji"] == "" {
			item["emoji"] = emoji
		}
		if item["cantidad"] == nil {
			item["cantidad"] = cantidad
		}
		if item["colores"] == nil {
			item["colores"] = colores
		}

		byOrder[pedidoFirebaseID] = append(byOrder[pedidoFirebaseID], item)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for _, order := range orders {
		id := FirstValue(order, "id", "firebase_id")
		if items, ok := byOrder[id]; ok && len(items) > 0 {
			order["itemsFlores"] = items
		}
	}

	return nil
}

// GenerateUniqueID genera un identificador único alfanumérico aleatorio (20 caracteres)
func GenerateUniqueID() string {
	bytes := make([]byte, 10)
	if _, err := rand.Read(bytes); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes)
}

func FirstValue(values map[string]any, keys ...string) string {
	for _, key := range keys {
		if val, ok := values[key]; ok && val != nil {
			s := strings.TrimSpace(fmt.Sprint(val))
			if s != "" {
				return s
			}
		}
	}
	return ""
}

func parseNumeric(v any) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case string:
		var f float64
		_, _ = fmt.Sscanf(val, "%f", &f)
		return f
	default:
		return 0
	}
}

func parseStringArray(v any) []string {
	res := make([]string, 0)
	switch arr := v.(type) {
	case []string:
		return arr
	case []any:
		for _, item := range arr {
			if s := strings.TrimSpace(fmt.Sprint(item)); s != "" {
				res = append(res, s)
			}
		}
	}
	return res
}
