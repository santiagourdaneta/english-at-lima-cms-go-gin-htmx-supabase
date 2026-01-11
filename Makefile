# Instalar el linter si no existe
install-linter:
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.62.2

	# Variables
BINARY_NAME=english-at-lima-cms

# Cargar variables desde el archivo .env
ifneq ("$(wildcard .env)","")
    include .env
    export
endif

# Limpiar archivos temporales y binarios
clean:
	@echo "🧹 Limpiando binarios antiguos..."
	@rm -f $(BINARY_NAME)
	@go clean

# El comando Check es nuestro filtro de calidad
check:
	@echo "🛡️  EL GUARDIÁN: Iniciando inspección profunda..."
	@go mod tidy
	@go fmt ./...
	@go vet ./...
	@go test ./internal/handlers/... -v
	@golangci-lint run
	@echo "✨ SISTEMA IMPENETRABLE: Todo el código cumple con los estándares élite."

# Este comando lo ejecutas DESPUÉS de tu git push
notify:
	@echo "🔔 Notificando a Render para actualizar el servicio..."
	@curl -s -X GET "$(RENDER_DEPLOY_HOOK)?clear_cache=1" > /dev/null
	@echo "🚀 Despliegue en marcha con limpieza de caché en Render."
