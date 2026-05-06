package mocks

import (
	"context"

	"github.com/YagoGomez83/audit-service/internal/domain"
	"github.com/stretchr/testify/mock"
)

// AuditRepoMock es una versión falsa de la base de datos para pruebas.
// Hereda la magia de mock.Mock de Testify.
type AuditRepoMock struct {
	mock.Mock
}

// Create simula la inserción. En lugar de ir a Postgres, simplemente
// registra que fue llamado y devuelve lo que nosotros le programemos en el test.
func (m *AuditRepoMock) Create(ctx context.Context, log *domain.AuditLog) error {
	// Le decimos a Testify: "Registra que llamaron a este método con estos argumentos"
	args := m.Called(ctx, log)

	// Retorna el error que hayamos configurado en nuestro test (puede ser nil)
	return args.Error(0)
}

// GetLogs simula la recuperación de logs.
func (m *AuditRepoMock) GetLogs(ctx context.Context, limit int, offset int) ([]*domain.AuditLog, error) {
	args := m.Called(ctx, limit, offset)

	// Si el primer argumento (índice 0) que programamos devolver es nil, manejamos el cast
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	// Casteamos el resultado al tipo correcto ([]*domain.AuditLog)
	return args.Get(0).([]*domain.AuditLog), args.Error(1)
}
