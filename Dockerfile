# =============================================================================
# Stage 1: Builder
# Compilamos el binario en un entorno con todas las herramientas de Go.
# Este stage NO llega a producción.
# =============================================================================
FROM golang:1.26-alpine AS builder

# ca-certificates: necesario para que go mod download funcione con HTTPS
RUN apk add --no-cache ca-certificates git

WORKDIR /app

# Copiamos los archivos de dependencias primero (optimización de caché Docker).
# Si go.mod y go.sum no cambian, Docker reutiliza esta capa sin re-descargar módulos.
COPY go.mod go.sum ./
RUN go mod download

# Copiamos el resto del código fuente
COPY . .

# Compilamos el binario con flags de seguridad y optimización:
# CGO_ENABLED=0  → binario 100% estático (no depende de librerías C del SO)
# GOOS=linux     → compilación cruzada para Linux (necesario si compilas en Mac/Windows)
# -ldflags="-w -s" → elimina tabla de depuración y símbolos (binario más pequeño)
# -trimpath      → elimina rutas del sistema de archivos local del binario (seguridad)
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -trimpath \
    -o /app/audit-service \
    ./cmd/api/

# =============================================================================
# Stage 2: Final
# Imagen mínima de producción. Solo contiene el binario compilado.
# ~15 MB totales vs ~300 MB del stage builder.
# =============================================================================
FROM alpine:3.21

# ca-certificates: para peticiones TLS/HTTPS salientes desde la app
# tzdata: para manejo correcto de zonas horarias en timestamps
RUN apk add --no-cache ca-certificates tzdata

# DevSecOps: nunca corras contenedores como root.
# Creamos un usuario y grupo sin privilegios dedicados a la app.
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

# Copiamos ÚNICAMENTE el binario compilado desde el stage anterior.
# Nada del código fuente, go.mod, ni herramientas de Go llegan aquí.
COPY --from=builder /app/audit-service .

# Activamos el usuario sin privilegios antes de exponer el puerto
USER appuser

# Puerto documentado (no abre el puerto, eso lo hace docker-compose/k8s)
EXPOSE 8080

ENTRYPOINT ["./audit-service"]
