# 1. Etapa de construcción (Build)
FROM golang:1.24-alpine AS builder
WORKDIR /app
# Copiar archivos de dependencias
COPY go.mod go.sum ./
RUN go mod download
# Copiar todo el código (incluyendo la carpeta server y templates)
COPY . .
# Compilar el binario desde la subcarpeta server
RUN go build -o main server/main.go

# 2. Etapa de ejecución (Runtime)
FROM alpine:latest
WORKDIR /root/
# MUY IMPORTANTE: Copiar el binario Y la carpeta de templates del builder
COPY --from=builder /app/main .
COPY --from=builder /app/templates ./templates

# Exponer el puerto (Render usa el puerto que definas o 8080 por defecto)
EXPOSE 8080

# Ejecutar el binario
CMD ["./main"]