-- =============================================================================
-- Migración Flyway: V1__initial_schema.sql
-- Base de datos: PostgreSQL (Railway)
-- Descripción: Esquema inicial para Karen's Floristik
-- =============================================================================

-- 1. Tabla de Pedidos
CREATE TABLE IF NOT EXISTS pedidos (
    firebase_id           TEXT PRIMARY KEY,
    barrio                TEXT,
    celular_destinatario  TEXT,
    celular_remitente     TEXT,
    ciudad                TEXT,
    descripcion_pedido    TEXT,
    direccion             TEXT,
    estado                TEXT,
    fecha_creacion_texto  TEXT,
    fecha_entrega_texto   TEXT,
    hora_entrega_texto    TEXT,
    medio_pago            TEXT,
    mensaje_tarjeta       TEXT,
    metodo_pago_anticipo  TEXT,
    monto_anticipo_texto  TEXT,
    monto_total_texto     TEXT,
    nombre_destinatario   TEXT,
    nombre_remitente      TEXT,
    numero_whatsapp_usado TEXT,
    pais                  TEXT,
    punto_referencia      TEXT,
    datos_originales      JSONB NOT NULL,
    migrado_en            TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);

-- Índices recomendados para optimización de consultas
CREATE INDEX IF NOT EXISTS idx_pedidos_fecha_entrega ON pedidos (fecha_entrega_texto);
CREATE INDEX IF NOT EXISTS idx_pedidos_estado ON pedidos (estado);
CREATE INDEX IF NOT EXISTS idx_pedidos_fecha_creacion ON pedidos (fecha_creacion_texto DESC);

-- 2. Secuencia para items_flores
CREATE SEQUENCE IF NOT EXISTS items_flores_id_seq;

-- 3. Tabla de Items de Flores (relacionada con pedidos)
CREATE TABLE IF NOT EXISTS items_flores (
    id                 BIGINT PRIMARY KEY DEFAULT nextval('items_flores_id_seq'::regclass),
    pedido_firebase_id TEXT NOT NULL REFERENCES pedidos(firebase_id) ON DELETE CASCADE,
    posicion           INTEGER NOT NULL,
    cantidad           NUMERIC,
    colores            TEXT[] NOT NULL DEFAULT '{}'::text[],
    emoji              TEXT,
    tipo               TEXT,
    datos_originales   JSONB NOT NULL,
    CONSTRAINT items_flores_pedido_firebase_id_posicion_key UNIQUE (pedido_firebase_id, posicion)
);

-- Índice para acelerar la búsqueda de flores por pedido
CREATE INDEX IF NOT EXISTS idx_items_flores_pedido ON items_flores (pedido_firebase_id);
