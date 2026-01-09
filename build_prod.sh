#!/bin/bash

echo "🏗️  Iniciando proceso de Producción para English At Lima..."

# 1. Limpieza preventiva
echo "🧹 Limpiando archivos temporales y logs viejos..."
rm -f english_admin_prod
rm -f server.log
touch server.log

# 2. Compilación de Alto Rendimiento
# -s: omite la tabla de símbolos (reduce tamaño)
# -w: omite la información de depuración DWARF
echo "⚙️  Compilando binario optimizado para la nube..."
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o english_admin_prod .

if [ $? -eq 0 ]; then
    echo "✅ ¡Éxito! Binario 'english_admin_prod' listo para subir."
    echo "📦 Tamaño del archivo reducido para carga ultra-rápida."
else
    echo "❌ Error en la compilación de producción."
    exit 1
fi