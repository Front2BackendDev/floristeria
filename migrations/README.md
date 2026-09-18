# Migraciones de Base de Datos (Flyway / PostgreSQL)

Este directorio contiene las definiciones DDL versionadas de la base de datos de Karen's Floristik en PostgreSQL (Railway).

## Convención de Versiones Flyway

Los archivos siguen la nomenclatura estándar de Flyway:
`V<Version>__<Descripcion>.sql`

- **`V1__initial_schema.sql`**: Esquema inicial migrado desde Firebase hacia PostgreSQL:
  - Tabla `pedidos`: contiene los campos de cada orden con clave primaria `firebase_id` y el documento completo en `datos_originales` (`jsonb`).
  - Tabla `items_flores`: contiene las flores y cantidades configuradas por el cliente, vinculadas por `pedido_firebase_id`.

## Modelo de Datos

```
+--------------------------------------------------------+
| pedidos                                                |
+--------------------------------------------------------+
| firebase_id (PK, TEXT)                                 |
| nombre_remitente, celular_remitente                    |
| nombre_destinatario, celular_destinatario              |
| fecha_creacion_texto, fecha_entrega_texto, hora_entrega|
| direccion, barrio, ciudad, pais, punto_referencia      |
| descripcion_pedido, mensaje_tarjeta                    |
| monto_total_texto, monto_anticipo_texto                |
| medio_pago, metodo_pago_anticipo, estado               |
| numero_whatsapp_usado                                  |
| datos_originales (JSONB NOT NULL)                      |
| migrado_en (TIMESTAMPTZ DEFAULT now())                 |
+--------------------------------------------------------+
                           | 1
                           |
                           | N
+--------------------------------------------------------+
| items_flores                                           |
+--------------------------------------------------------+
| id (PK, BIGINT)                                        |
| pedido_firebase_id (FK -> pedidos.firebase_id)         |
| posicion (INTEGER)                                     |
| cantidad (NUMERIC)                                     |
| colores (TEXT[])                                       |
| emoji (TEXT)                                           |
| tipo (TEXT)                                            |
| datos_originales (JSONB NOT NULL)                      |
+--------------------------------------------------------+
```

## Aplicación manual (psql)
```bash
psql "$DATABASE_URL" -f migrations/V1__initial_schema.sql
```
