package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/YagoGomez83/audit-service/internal/domain"
)

// AuditHandler agrupa las funciones HTTP relacionadas con la auditoría.
type AuditHandler struct {
	// Inyectamos la INTERFAZ, no la implementación concreta de PostgreSQL.
	// Esto nos permite testear el handler aislando la base de datos.
	repo domain.AuditRepository
}

// NewAuditHandler es el constructor. Recibe el repositorio y devuelve el handler.
func NewAuditHandler(repo domain.AuditRepository) *AuditHandler {
	return &AuditHandler{
		repo: repo,
	}
}

// writeJSON escribe una respuesta JSON con el código de estado indicado.
func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

// writeError escribe una respuesta de error en formato JSON.
func writeError(w http.ResponseWriter, message string, code int) {
	writeJSON(w, code, map[string]string{"error": message})
}

// CreateLog crea un nuevo registro de auditoría.
//
// @Summary      Crear un log de auditoría
// @Description  Registra una nueva acción en el sistema. El campo "action" es obligatorio.
// @Tags         audit
// @Accept       json
// @Produce      json
// @Param        log  body      domain.AuditLog  true  "Datos del log a crear"
// @Success      201  {object}  domain.AuditLog
// @Failure      400  {object}  map[string]string  "Error de validación (ej. falta action)"
// @Failure      500  {object}  map[string]string  "Error interno del servidor"
// @Router       /audit [post]
func (h *AuditHandler) CreateLog(w http.ResponseWriter, r *http.Request) {
	// Seguridad: limitar el cuerpo a 1 MB para prevenir ataques DoS por payload enorme.
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields() // Rechaza campos JSON no definidos en el struct.

	var entry domain.AuditLog
	if err := dec.Decode(&entry); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Validar campo obligatorio.
	if entry.Action == "" {
		writeError(w, "action is required", http.StatusBadRequest)
		return
	}

	if err := h.repo.Create(r.Context(), &entry); err != nil {
		log.Printf("ERROR [CreateLog] repo.Create: %v", err)
		writeError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, entry)
}

// GetLogs devuelve una lista paginada de logs de auditoría.
//
// @Summary      Obtener logs de auditoría
// @Description  Recupera un listado paginado de registros, ordenados del más reciente al más antiguo.
// @Tags         audit
// @Produce      json
// @Param        limit   query     int  false  "Cantidad máxima de registros (default 10, max 100)"
// @Param        offset  query     int  false  "Cantidad de registros a saltar (default 0)"
// @Success      200  {array}   domain.AuditLog
// @Failure      500  {object}  map[string]string  "Error interno del servidor"
// @Router       /audit [get]
func (h *AuditHandler) GetLogs(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	// Valores por defecto seguros (Secure by Default).
	limit := 10
	offset := 0

	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
		// Hard-cap: evita que alguien mande ?limit=99999999 y genere un DoS.
		if l > 100 {
			limit = 100
		} else {
			limit = l
		}
	}

	if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
		offset = o
	}

	logs, err := h.repo.GetLogs(r.Context(), limit, offset)
	if err != nil {
		log.Printf("ERROR [GetLogs] repo.GetLogs: %v", err)
		writeError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// Garantiza que el JSON devuelva `[]` en lugar de `null` cuando no hay registros.
	if logs == nil {
		logs = []*domain.AuditLog{}
	}

	writeJSON(w, http.StatusOK, logs)
}
