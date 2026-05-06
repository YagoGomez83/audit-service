package repository

import (
"context"

"github.com/YagoGomez83/audit-service/internal/domain"
"github.com/jackc/pgx/v5/pgxpool"
)

// postgresAuditRepo es la implementacion concreta de domain.AuditRepository.
// Es privado (minuscula) para forzar el uso del constructor.
type postgresAuditRepo struct {
db *pgxpool.Pool
}

// NewPostgresAuditRepository construye el repositorio.
// Devuelve la INTERFAZ, no el struct concreto, para que los niveles superiores
// no puedan acceder directamente al campo db.
func NewPostgresAuditRepository(db *pgxpool.Pool) domain.AuditRepository {
return &postgresAuditRepo{db: db}
}

// Create inserta un nuevo AuditLog en la tabla audit_logs usando una consulta
// parametrizada ($1, $2, ...) que previene SQL Injection de forma estructural.
// La clausula RETURNING recupera id y created_at generados por PostgreSQL
// en el mismo viaje de red, sin necesidad de una segunda query.
func (r *postgresAuditRepo) Create(ctx context.Context, log *domain.AuditLog) error {
const query = `
INSERT INTO audit_logs (user_id, action, entity_type, entity_id, ip_address, user_agent, action_details)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, created_at`

return r.db.QueryRow(ctx, query,
log.UserID,
log.Action,
log.EntityType,
log.EntityID,
log.IPAddress,
log.UserAgent,
log.ActionDetails,
).Scan(&log.ID, &log.CreatedAt)
}

// GetLogs recupera los registros ordenados del más reciente al más antiguo.
func (r *postgresAuditRepo) GetLogs(ctx context.Context, limit int, offset int) ([]*domain.AuditLog, error) {
	const query = `
		SELECT id, user_id, action, entity_type, entity_id, ip_address, user_agent, action_details, created_at
		FROM audit_logs
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	// db.Query devuelve un cursor (rows) para múltiples filas.
	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	// Garantiza que la conexión se devuelva al pool cuando termine la función.
	defer rows.Close()

	var logs []*domain.AuditLog

	for rows.Next() {
		var log domain.AuditLog

		err := rows.Scan(
			&log.ID,
			&log.UserID,
			&log.Action,
			&log.EntityType,
			&log.EntityID,
			&log.IPAddress,
			&log.UserAgent,
			&log.ActionDetails,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		logs = append(logs, &log)
	}

	// rows.Next() también puede terminar por error de red a mitad de la transmisión.
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return logs, nil
}
