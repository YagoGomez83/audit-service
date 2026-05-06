# 🛡️ Audit Service

![Go](https://img.shields.io/badge/Go-1.26-00ADD8?style=flat&logo=go&logoColor=white)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-pgx%2Fv5-336791?style=flat&logo=postgresql&logoColor=white)
![Docker](https://img.shields.io/badge/Docker-Multi--Stage-2496ED?style=flat&logo=docker&logoColor=white)
![GitHub Actions](https://img.shields.io/badge/CI%2FCD-GitHub%20Actions-2088FF?style=flat&logo=github-actions&logoColor=white)
![Swagger](https://img.shields.io/badge/API-Swagger%20UI-85EA2D?style=flat&logo=swagger&logoColor=black)
![License](https://img.shields.io/badge/License-MIT-green?style=flat)

Microservicio DevSecOps de **registro inmutable de auditoría**, construido en Go con PostgreSQL. Diseñado para capturar cualquier acción significativa del sistema (login, cambios de configuración, accesos a datos sensibles) y exponerla a través de una API REST segura, documentada y observable.

---

## 🏗️ Arquitectura & Principios DevSecOps

Este proyecto aplica las siguientes prácticas de ingeniería de forma explícita:

| Principio | Implementación |
|---|---|
| **Clean Architecture** | Separación estricta en capas: `domain` → `repository` → `handlers`. La capa HTTP nunca conoce los detalles de PostgreSQL. |
| **Repository Pattern** | La interfaz `domain.AuditRepository` desacopla la lógica de negocio del motor de base de datos. Sustituible sin tocar los handlers. |
| **Inyección de Dependencias** | `NewAuditHandler(repo)` recibe la interfaz, no el struct concreto. Permite inyectar mocks en testing. |
| **Connection Pooling** | `pgxpool` con `MaxConns=15`, `MinConns=2`, `MaxConnLifetime=1h`. Previene saturación de conexiones en picos de carga. |
| **Mitigación de SQLi** | Todas las queries usan parámetros posicionales (`$1`, `$2`). Ningún valor de usuario se concatena directamente en SQL. |
| **Observabilidad** | Middleware propio (`RequestLogger`) registra latencia, IP del cliente, método HTTP y código de respuesta en cada petición. |
| **DoS Prevention** | `http.MaxBytesReader` limita el body de entrada a 1 MB en el endpoint de creación. |
| **Imagen mínima** | Dockerfile multi-stage: imagen final basada en `alpine:3.21` (~15 MB vs ~300 MB del builder). Binario compilado con `CGO_ENABLED=0 -trimpath -ldflags="-w -s"`. |
| **Shift-Left Security** | Pipeline CI/CD ejecuta `gosec` (SAST) y `golangci-lint` antes de construir la imagen Docker. |

---

## 📋 Prerrequisitos

- [Docker Desktop](https://www.docker.com/products/docker-desktop/) (incluye Docker Compose)
- [Go 1.26+](https://go.dev/dl/) *(solo necesario para desarrollo local sin Docker)*

---

## 🚀 Setup Rápido

Clona el repositorio e inicia todos los servicios con un solo comando:

```bash
git clone https://github.com/YagoGomez83/audit-service.git
cd audit-service
docker compose up --build
```

Docker Compose levantará:
1. **PostgreSQL 16** con el esquema inicializado automáticamente.
2. **audit-service** escuchando en `http://localhost:8080`.

Para detener y limpiar los volúmenes:

```bash
docker compose down -v
```

### Variables de Entorno

Configura un archivo `.env` en la raíz del proyecto (o usa los valores por defecto del `docker-compose.yml`):

```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=audit_user
DB_PASSWORD=audit_password
DB_NAME=audit_db
SERVER_PORT=8080
```

---

## 🧪 Testing

Los Unit Tests están diseñados para correr de forma **completamente aislada**, sin necesidad de levantar la base de datos ni contenedores adicionales.

Esto es posible gracias al **Repository Pattern con Mocks** (`testify/mock`): el test inyecta un repositorio falso (`AuditRepoMock`) en el handler, verificando únicamente la lógica HTTP sin dependencias externas.

```bash
# Ejecutar todos los tests
go test ./...

# Con reporte de cobertura
go test -v -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Casos de Test Cubiertos

| Test | Descripción | Espera |
|---|---|---|
| `TestCreateLog_MissingAction_Returns400` | Payload sin campo `action` | `HTTP 400` sin llamar al repositorio |
| `TestCreateLog_ValidPayload_Returns201` | Payload correcto | `HTTP 201` con el log creado |
| `TestGetLogs_ReturnsLogs_200` | Listado paginado | `HTTP 200` con el array de logs |

---

## 📖 Documentación de API (Swagger)

Con el servicio corriendo, accede a la interfaz interactiva:

```
http://localhost:8080/swagger/index.html
```

### Endpoints Principales

| Método | Ruta | Descripción |
|---|---|---|
| `POST` | `/api/v1/audit` | Registra un nuevo evento de auditoría |
| `GET` | `/api/v1/audit?limit=10&offset=0` | Obtiene logs paginados (más reciente primero) |

#### Ejemplo de Request

```bash
curl -X POST http://localhost:8080/api/v1/audit \
  -H "Content-Type: application/json" \
  -d '{
    "action": "USER_LOGIN",
    "user_id": "usr-123",
    "entity_type": "auth_system",
    "ip_address": "192.168.1.100",
    "action_details": {"method": "password", "success": true}
  }'
```

---

## 📁 Estructura del Proyecto

```
audit-service/
├── cmd/
│   └── api/
│       └── main.go          # Punto de entrada: wiring de dependencias y arranque del servidor
├── internal/
│   ├── config/
│   │   └── config.go        # Carga de configuración desde variables de entorno
│   ├── domain/
│   │   └── audit.go         # Entidad AuditLog + interfaz AuditRepository (corazón del negocio)
│   ├── handlers/
│   │   ├── audit_handler.go # Lógica HTTP: validación, serialización, respuestas
│   │   └── audit_handler_test.go # Unit tests con mocks (sin DB)
│   ├── middleware/
│   │   └── logger.go        # Middleware de observabilidad: latencia, IP, status code
│   ├── mocks/
│   │   └── audit_repo_mock.go # Mock generado con testify para testing aislado
│   └── repository/
│       └── postgres_audit_repo.go # Implementación concreta de AuditRepository con pgx
├── migrations/
│   ├── 000001_init_audit_schema.up.sql   # Esquema inicial de la tabla audit_logs
│   └── 000001_init_audit_schema.down.sql # Rollback del esquema
├── pkg/
│   └── database/
│       └── postgres.go      # Inicialización del pool de conexiones pgxpool
├── docs/                    # Documentación Swagger auto-generada por swag
├── .github/
│   └── workflows/
│       └── ci.yml           # Pipeline DevSecOps: SAST, lint, tests, Docker build
├── Dockerfile               # Multi-stage build: builder (golang:1.26) → final (alpine:3.21)
├── docker-compose.yml       # Orquestación local: PostgreSQL + audit-service
└── go.mod                   # Módulo: github.com/YagoGomez83/audit-service
```

---

## ⚙️ CI/CD Pipeline (GitHub Actions)

El pipeline implementa **Shift-Left Security** con tres jobs independientes:

```
push / pull_request
        │
        ├──► [A] security-and-lint  ──► gosec (SAST) + golangci-lint
        │
        ├──► [B] test               ──► go test -coverprofile (sin DB)
        │
        └──► [C] docker-build       ──► docker build (solo si A y B pasan)
```

El job `docker-build` tiene `needs: [security-and-lint, test]`, implementando el principio **Fail-Fast**: si `gosec` detecta una vulnerabilidad o un test falla, la construcción de la imagen se cancela automáticamente.

---

## 📄 Licencia

MIT © [YagoGomez83](https://github.com/YagoGomez83)
