-- migrations/000001_init_audit_schema.down.sql
-- La convención "down" REVIERTE el cambio (va hacia atrás).
-- Debe ser el inverso exacto del .up.sql, en orden inverso.

DROP TABLE IF EXISTS audit_logs;
