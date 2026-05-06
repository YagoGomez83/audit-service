package handlers_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/YagoGomez83/audit-service/internal/handlers"
	"github.com/YagoGomez83/audit-service/internal/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestCreateLog_MissingAction_Returns400 evalúa que el servidor rechace
// payloads incompletos sin siquiera tocar la base de datos.
func TestCreateLog_MissingAction_Returns400(t *testing.T) {
	// 1. Arrange (Preparación)

	// Instanciamos nuestro repositorio falso.
	mockRepo := new(mocks.AuditRepoMock)

	// Instanciamos el Handler, inyectándole el repositorio falso
	handler := handlers.NewAuditHandler(mockRepo)

	// Simulamos el JSON de entrada (sin el campo "action")
	bodyJSON := []byte(`{"entity_type": "auth_system"}`)

	// Creamos una petición HTTP ficticia
	req := httptest.NewRequest(http.MethodPost, "/api/v1/audit", bytes.NewBuffer(bodyJSON))
	req.Header.Set("Content-Type", "application/json")

	// Creamos un "Grabador" (Recorder) que simulará ser el cliente
	// para capturar la respuesta del servidor.
	rr := httptest.NewRecorder()

	// 2. Act (Ejecución)

	// Llamamos al método directamente, como si el Router HTTP lo hubiera hecho.
	handler.CreateLog(rr, req)

	// 3. Assert (Verificación)

	// Validamos que el código de estado sea HTTP 400
	assert.Equal(t, http.StatusBadRequest, rr.Code)

	// Validamos que el mensaje de error sea el correcto
	expectedResponse := `{"error":"action is required"}` + "\n"
	assert.Equal(t, expectedResponse, rr.Body.String())

	// Verificamos que el repositorio falso NUNCA fue llamado.
	// Esto es vital en seguridad: si el payload es inválido,
	// la base de datos no debe ser tocada para ahorrar recursos.
	mockRepo.AssertNotCalled(t, "Create")
}

// TestCreateLog_ValidPayload_Returns201 evalúa que un JSON válido
// sea procesado correctamente y llame al repositorio simulado.
func TestCreateLog_ValidPayload_Returns201(t *testing.T) {
	// 1. Arrange
	mockRepo := new(mocks.AuditRepoMock)
	handler := handlers.NewAuditHandler(mockRepo)

	// Un payload JSON perfectamente válido
	bodyJSON := []byte(`{
		"action": "DATA_EXPORT",
		"entity_type": "financial_records"
	}`)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/audit", bytes.NewBuffer(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	// LA MAGIA DEL MOCK: Programamos el comportamiento.
	// Le decimos a Testify: "Cuando el handler llame al método 'Create'
	// con CUALQUIER contexto y CUALQUIER puntero a AuditLog,
	// debes devolver 'nil' (simulando que no hubo error en la base de datos)".
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.AuditLog")).Return(nil)

	// 2. Act
	handler.CreateLog(rr, req)

	// 3. Assert

	// Validamos el código de estado exitoso
	assert.Equal(t, http.StatusCreated, rr.Code)

	// Verificamos que el repositorio SÍ fue llamado exactamente una vez.
	// Si el handler tuviera un bug y no llamara a la BD, el test fallaría aquí.
	mockRepo.AssertCalled(t, "Create", mock.Anything, mock.AnythingOfType("*domain.AuditLog"))
	mockRepo.AssertNumberOfCalls(t, "Create", 1)
}
