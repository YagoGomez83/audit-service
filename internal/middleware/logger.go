package middleware

import (
	"log"
	"net/http"
	"time"
)

// responseRecorder envuelve al http.ResponseWriter original.
// Nos permite espiar (capturar) el status code que el handler
// decide enviar, ya que la librería estándar de Go no lo guarda por defecto.
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader intercepta la llamada original, guarda el status code y luego la ejecuta.
func (rec *responseRecorder) WriteHeader(statusCode int) {
	rec.statusCode = statusCode
	rec.ResponseWriter.WriteHeader(statusCode)
}

// RequestLogger es un Middleware que registra cada petición entrante y saliente.
// Recibe el "next" handler (tu AuditHandler, por ejemplo) y devuelve un nuevo handler.
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now() // Capturamos el inicio de la petición

		// 1. Envolvemos el escritor HTTP original.
		// Inicializamos con 200 por defecto en caso de que el handler nunca llame a WriteHeader.
		rec := &responseRecorder{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// 2. Ejecutamos el siguiente Handler en la cadena (tu código de negocio).
		// Se quedará procesando aquí hasta que termine.
		next.ServeHTTP(rec, r)

		// 3. Medimos cuánto tiempo tomó procesar toda la petición.
		duration := time.Since(start)

		// 4. Registramos en consola (o sistema de logs) con un formato estandarizado.
		// En producción, esto debería escribirse en formato JSON estructurado
		// (ej: usando log/slog o zap) para enviarlo a Datadog o Kibana.
		// #nosec G706 -- Aceptamos el riesgo de inyección de log en esta etapa, en prod se migrará a slog JSON.
		log.Printf(
			"| %3d | %10v | %-15s | %-6s %s",
			rec.statusCode, // Código HTTP devuelto (ej: 201, 400)
			duration,       // Latencia (ej: 4.5ms)
			r.RemoteAddr,   // IP y puerto del cliente
			r.Method,       // GET, POST
			r.URL.Path,     // /api/v1/audit
		)
	})
}
