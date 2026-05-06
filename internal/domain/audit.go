package domain

import (
	"context"
	"encoding/json"
	"time"
)

// AuditLog representa un registro de auditoría en el sistema.
// Cada campo mapea directamente a una columna de la tabla audit_logs.
type AuditLog struct {
	ID            string          `json:"id"`                       // string para soportar UUID de PostgreSQL
	UserID        *string         `json:"user_id,omitempty"`        // puntero: puede ser NULL (acción sin usuario)
	Action        string          `json:"action"`                   // obligatorio, nunca NULL
	EntityType    *string         `json:"entity_type,omitempty"`    // puntero: entidad opcional
	EntityID      *string         `json:"entity_id,omitempty"`      // puntero: id de entidad opcional
	IPAddress     *string         `json:"ip_address,omitempty"`     // puntero: puede no haber IP
	UserAgent     *string         `json:"user_agent,omitempty"`     // puntero: puede no haber agente
	ActionDetails json.RawMessage `json:"action_details,omitempty" swaggertype:"object"` // bytes crudos: evita parsear JSON innecesariamente
	CreatedAt     time.Time       `json:"created_at"`
}

// AuditRepository define las operaciones que nuestra aplicación necesita
// para persistir y leer los registros de auditoría.
type AuditRepository interface {
	// Create inserta un nuevo registro en la base de datos.
	// Recibe un context para manejo de timeouts y un puntero al registro.
	Create(ctx context.Context, log *AuditLog) error

	// GetLogs recupera un listado paginado de logs ordenados del más reciente al más antiguo.
	GetLogs(ctx context.Context, limit int, offset int) ([]*AuditLog, error)
}
