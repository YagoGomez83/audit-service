package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/YagoGomez83/audit-service/internal/config"
	"github.com/YagoGomez83/audit-service/internal/handlers"
	"github.com/YagoGomez83/audit-service/internal/middleware"
	"github.com/YagoGomez83/audit-service/internal/repository"
	"github.com/YagoGomez83/audit-service/pkg/database"
	httpSwagger "github.com/swaggo/http-swagger"

	// Importación anónima: su efecto secundario (init) registra
	// los docs generados por swag en el paquete de Swagger.
	_ "github.com/YagoGomez83/audit-service/docs"
)

// @title           Audit Service API
// @version         1.0
// @description     Microservicio DevSecOps para registro inmutable de auditoría.
// @contact.name    YagoGomez83
// @license.name    MIT
// @host            localhost:8080
// @BasePath        /api/v1
// @schemes         http

// main es el punto de entrada de nuestra aplicación.
func main() {
	log.Println("Iniciando Microservicio de Auditoría...")

	cfg := config.LoadConfig()
	log.Printf("Servidor configurado en el puerto: %s", cfg.ServerPort)

	ctx := context.Background()

	pool, err := database.NewPostgresPool(ctx, cfg)
	if err != nil {
		log.Fatalf("No se pudo inicializar el pool de base de datos: %v", err)
	}
	defer pool.Close()

	// Instanciamos el repositorio pasándole el pool de conexiones.
	// auditRepo cumple la interfaz domain.AuditRepository; el resto de la app
	// solo conocerá esa interfaz, nunca el struct concreto de postgres.
	auditRepo := repository.NewPostgresAuditRepository(pool)
	auditHandler := handlers.NewAuditHandler(auditRepo)

	// 4. Configurar el Enrutador (Multiplexor HTTP)
	mux := http.NewServeMux()

	// Patrón de enrutamiento nativo de Go 1.22+
	mux.HandleFunc("POST /api/v1/audit", auditHandler.CreateLog)
	mux.HandleFunc("GET /api/v1/audit", auditHandler.GetLogs)

	// Swagger UI — disponible en http://localhost:8080/swagger/index.html
	// http-swagger actúa como handler que sirve los archivos estáticos de la UI
	// y apunta a los docs generados en la carpeta /docs/.
	mux.HandleFunc("GET /swagger/", httpSwagger.WrapHandler)

	// 5. Configurar el Servidor HTTP con blindaje DevSecOps y Observabilidad
	//
	// Envolvemos nuestro Mux completo con el middleware RequestLogger.
	// Ahora, "loggedMux" es el nuevo punto de entrada de la aplicación.
	// Toda petición (incluyendo Swagger) pasará por el guardia de logs.
	loggedMux := middleware.RequestLogger(mux)

	server := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      loggedMux,         // Inyectamos el mux envuelto
		ReadTimeout:  5 * time.Second,   // Tiempo máximo para leer la petición del cliente
		WriteTimeout: 10 * time.Second,  // Tiempo máximo para responderle al cliente
		IdleTimeout:  120 * time.Second, // Tiempo máximo para mantener conexiones Keep-Alive
	}

	// 6. Levantar el servidor
	log.Printf("Servidor escuchando en http://localhost:%s", cfg.ServerPort)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Error crítico en el servidor HTTP: %v", err)
	}
}
