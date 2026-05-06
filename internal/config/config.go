package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// AppConfig agrupa toda la configuración de la aplicación.
// Usamos un Struct para tener la configuración fuertemente tipada.
type AppConfig struct {
	ServerPort string
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
}

// LoadConfig lee el archivo .env y retorna la estructura poblada.
func LoadConfig() *AppConfig {
	// Intentamos cargar el archivo .env
	// Si falla, no detenemos la app (por si estamos en Docker donde
	// las variables se inyectan sin archivo .env)
	err := godotenv.Load()
	if err != nil {
		log.Println("Aviso: No se encontró archivo .env, usando variables de entorno del sistema")
	}

	// Retornamos un puntero a AppConfig con los valores obtenidos
	return &AppConfig{
		ServerPort: getEnvOrDefault("SERVER_PORT", "8080"),
		DBHost:     getEnvOrDefault("DB_HOST", "localhost"),
		DBPort:     getEnvOrDefault("DB_PORT", "5432"),
		DBUser:     getEnvOrDefault("DB_USER", "postgres"),
		DBPassword: getEnvOrDefault("DB_PASSWORD", "postgres"),
		DBName:     getEnvOrDefault("DB_NAME", "postgres"),
	}
}

// getEnvOrDefault es una función auxiliar (privada, porque empieza con minúscula).
// Busca una variable; si no existe o está vacía, devuelve el valor fallback.
func getEnvOrDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
