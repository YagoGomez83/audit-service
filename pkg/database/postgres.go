package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/YagoGomez83/audit-service/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPostgresPool inicializa y retorna un pool de conexiones a PostgreSQL.
func NewPostgresPool(ctx context.Context, cfg *config.AppConfig) (*pgxpool.Pool, error) {
	// 1. Construimos el Data Source Name (DSN) o cadena de conexión
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)

	// 2. Parseamos la configuración básica
	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		// %w "envuelve" (wraps) el error original. Esto permite rastrearlo después.
		return nil, fmt.Errorf("error parseando la configuración de la DB: %w", err)
	}

	// 3. DevSecOps: Afinación del Pool para resiliencia
	// Evitamos que un pico de tráfico sature las conexiones máximas de Postgres
	poolConfig.MaxConns = 15                      // Máximo de conexiones simultáneas
	poolConfig.MinConns = 2                       // Conexiones mínimas siempre activas
	poolConfig.MaxConnLifetime = time.Hour        // Tiempo máximo de vida de una conexión (evita fugas de memoria)
	poolConfig.MaxConnIdleTime = 30 * time.Minute // Si una conexión no se usa en 30m, se cierra

	// 4. Creamos el Pool de conexiones
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("error creando el pool de conexiones: %w", err)
	}

	// 5. Verificación real (Ping)
	// NewWithConfig a veces no falla si la DB está apagada, solo crea el pool.
	// El Ping asegura que realmente podemos hablar con la base de datos.
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("la base de datos no responde al ping: %w", err)
	}

	log.Println("Conexión establecida exitosamente con el pool de PostgreSQL")

	// Retornamos el pool configurado y un error 'nil' (todo salió bien)
	return pool, nil
}
