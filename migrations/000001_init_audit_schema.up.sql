-- migrations/000001_init_audit_schema.up.sql
-- Esta migración crea la tabla principal del microservicio.
-- La convención "up" significa que APLICA el cambio (va hacia adelante).

-- Activamos la extensión que permite generar UUIDs dentro de PostgreSQL.
-- gen_random_uuid() es la función que usaremos como valor por defecto del ID.
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS audit_logs (
    -- UUID como PK: global, no adivina el siguiente ID, ideal para sistemas distribuidos.
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Campos opcionales (NULL permitido): no toda acción tiene usuario o entidad.
    user_id       TEXT        NULL,
    entity_type   TEXT        NULL,
    entity_id     TEXT        NULL,
    ip_address    TEXT        NULL,  -- TEXT almacena la IP como string (sin validación de formato en BD)
    user_agent    TEXT        NULL,

    -- Campos obligatorios.
    action        TEXT        NOT NULL,

    -- JSONB almacena JSON binario: más rápido para buscar y filtrar que JSON plano.
    action_details JSONB      NULL,

    -- Timestamp con zona horaria. DEFAULT manejado por PostgreSQL, no por la app.
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Índice para acelerar consultas por usuario (caso de uso más común).
CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id    ON audit_logs (user_id);

-- Índice para acelerar consultas por entidad (ej: "dame todos los cambios del producto X").
CREATE INDEX IF NOT EXISTS idx_audit_logs_entity     ON audit_logs (entity_type, entity_id);

-- Índice para ordenar por fecha (paginación, dashboards).
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs (created_at DESC);
