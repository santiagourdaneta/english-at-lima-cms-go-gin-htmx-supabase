# --- FASE 1: Compilación ---
FROM golang:1.25-alpine AS builder

# Instalamos git por si alguna dependencia lo necesita
RUN apk add --no-cache git

# Definimos el directorio de trabajo
WORKDIR /app

# Copiamos los archivos de módulos primero para aprovechar el caché de Docker
COPY go.mod go.sum ./
RUN go mod download

# Copiamos el resto del código fuente
COPY . .

# Compilamos el binario con optimizaciones de producción (sin debug info)
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o english_admin .

# --- FASE 2: Ejecución ---
FROM alpine:latest  

# Instalamos certificados CA (necesarios para conectar con Supabase/SSL)
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copiamos el binario desde la fase de compilación
COPY --from=builder /app/english_admin .

# Copiamos la carpeta de templates (esencial para Gin)
COPY --from=builder /app/templates ./templates

# Exponemos el puerto que usa tu servidor
EXPOSE 8080

# Comando para arrancar la aplicación
CMD ["./english_admin"]