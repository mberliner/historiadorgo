# Makefile para el proyecto Historiador Go

# Variables
BINARY_NAME=historiador
MAIN_PATH=cmd/main.go
BUILD_DIR=bin
DIST_DIR=dist

# Comandos de desarrollo
.PHONY: help build run clean test mod-tidy fmt vet lint

help: ## Mostrar este mensaje de ayuda
	@echo "Comandos disponibles:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## Compilar la aplicación
	@echo "🔨 Compilando $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "✅ Compilación exitosa: $(BUILD_DIR)/$(BINARY_NAME)"

run: build ## Compilar y ejecutar
	@echo "🚀 Ejecutando $(BINARY_NAME)..."
	./$(BUILD_DIR)/$(BINARY_NAME) $(ARGS)

clean: ## Limpiar archivos generados
	@echo "🧹 Limpiando archivos generados..."
	rm -rf $(BUILD_DIR) $(DIST_DIR)
	go clean

test: ## Ejecutar tests
	@echo "🧪 Ejecutando tests..."
	go test -v -race -cover ./...

mod-tidy: ## Limpiar dependencias
	@echo "📦 Limpiando dependencias..."
	go mod tidy

fmt: ## Formatear código
	@echo "🎨 Formateando código..."
	go fmt ./...

vet: ## Analizar código estático
	@echo "🔍 Analizando código..."
	go vet ./...

lint: fmt vet ## Ejecutar linting completo
	@echo "✨ Linting completado"

# Comandos de construcción
.PHONY: build-linux build-windows build-darwin build-all

build-linux: ## Compilar para Linux
	@echo "🐧 Compilando para Linux..."
	@mkdir -p $(DIST_DIR)
	GOOS=linux GOARCH=amd64 go build -o $(DIST_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)

build-windows: ## Compilar para Windows
	@echo "🪟 Compilando para Windows..."
	@mkdir -p $(DIST_DIR)
	GOOS=windows GOARCH=amd64 go build -o $(DIST_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)

build-darwin: ## Compilar para macOS
	@echo "🍎 Compilando para macOS..."
	@mkdir -p $(DIST_DIR)
	GOOS=darwin GOARCH=amd64 go build -o $(DIST_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)

build-all: build-linux build-windows build-darwin ## Compilar para todas las plataformas
	@echo "✅ Compilación multiplataforma completada"

# Comandos de desarrollo específicos del proyecto
.PHONY: test-connection validate process diagnose

test-connection: build ## Probar conexión con Jira
	./$(BUILD_DIR)/$(BINARY_NAME) test-connection

validate: build ## Validar archivo (requiere FILE)
	@if [ -z "$(FILE)" ]; then echo "❌ Usar: make validate FILE=archivo.csv"; exit 1; fi
	./$(BUILD_DIR)/$(BINARY_NAME) validate -f $(FILE) $(if $(PROJECT),-p $(PROJECT))

process: build ## Procesar archivos (requiere PROJECT)
	@if [ -z "$(PROJECT)" ]; then echo "❌ Usar: make process PROJECT=PROJ-KEY [FILE=archivo.csv] [DRY_RUN=true]"; exit 1; fi
	./$(BUILD_DIR)/$(BINARY_NAME) process -p $(PROJECT) $(if $(FILE),-f $(FILE)) $(if $(DRY_RUN),--dry-run)

diagnose: build ## Diagnosticar configuración de Features (requiere PROJECT)
	@if [ -z "$(PROJECT)" ]; then echo "❌ Usar: make diagnose PROJECT=PROJ-KEY"; exit 1; fi
	./$(BUILD_DIR)/$(BINARY_NAME) diagnose -p $(PROJECT)

# Configuración inicial
.PHONY: setup

setup: mod-tidy ## Configuración inicial del proyecto
	@echo "⚙️  Configurando proyecto..."
	@mkdir -p entrada procesados logs
	@if [ ! -f .env ]; then cp .env.example .env; echo "📝 Archivo .env creado. Configura tus credenciales."; fi
	@echo "✅ Configuración inicial completada"

# Instalación
.PHONY: install

install: build ## Instalar binario en el sistema
	@echo "📦 Instalando $(BINARY_NAME)..."
	sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "✅ $(BINARY_NAME) instalado en /usr/local/bin/"