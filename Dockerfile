# --- FASE 1: Compilación ---
# Usamos una versión específica y estable para evitar sorpresas
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

# Aprovechar el caché de capas de Docker para dependencias
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Compilación estática optimizada
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o english_admin ./server/main.go

# --- FASE 2: Ejecución (Runtime) ---
FROM alpine:3.21

# Instalar dependencias mínimas para HTTPS y Timezones
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copiar el binario
COPY --from=builder /app/english_admin .

# --- IMPORTANTE: Copiar todos los assets necesarios ---
# Gin necesita templates y static para renderizar la web
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/static ./static

# Crear carpeta de uploads con permisos (necesaria para tus recursos PDF)
RUN mkdir ./uploads

EXPOSE 8080

# Usar variables de entorno por defecto (pueden ser sobrescritas)
ENV GIN_MODE=release

CMD ["./english_admin"]