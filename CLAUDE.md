# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Descripción del Proyecto

Historiador Go es una aplicación CLI para crear historias de usuario en Jira desde archivos Excel/CSV con gestión automática de features y subtareas. Está implementado en Go usando arquitectura hexagonal y el patrón ports & adapters.

## Comandos de Desarrollo

### Construcción y Ejecución
```bash
# Compilar la aplicación
make build
# o
go build -o bin/historiador cmd/main.go

# Ejecutar directamente
make run ARGS="test-connection"
# o
./bin/historiador test-connection

# Instalar en el sistema
make install
```

### Testing y Calidad de Código
```bash
# Ejecutar todos los tests
make test
# o 
go test -v -race -cover ./...

# Formatear y validar código
make fmt      # Formatear código con go fmt
make vet      # Análisis estático con go vet
make lint     # Linting completo (fmt + vet)

# Limpiar dependencias
make mod-tidy
# o
go mod tidy
```

### Comandos de la Aplicación
```bash
# Probar conexión con Jira
make test-connection

# Validar archivos sin crear issues
make validate FILE=archivo.csv [PROJECT=PROJ-KEY]

# Procesar archivos (requiere PROJECT)
make process PROJECT=PROJ-KEY [FILE=archivo.csv] [DRY_RUN=true]

# Diagnosticar configuración de features
make diagnose PROJECT=PROJ-KEY

# Configurar proyecto (crea directorios, .env)
make setup
```

## Arquitectura

El proyecto sigue **Arquitectura Hexagonal** (Ports & Adapters) con principios de Clean Architecture:

### Estructura de Capas
- **Dominio** (`internal/domain/`): Entidades de negocio centrales e interfaces de repositorios
  - `entities/`: Objetos de negocio centrales (UserStory, BatchResult, FeatureResult, ProcessResult)
  - `repositories/`: Interfaces de puertos (FileRepository, JiraRepository, FeatureManager)

- **Aplicación** (`internal/application/`): Lógica de negocio y casos de uso
  - `usecases/`: Operaciones de negocio (ProcessFiles, ValidateFile, TestConnection, DiagnoseFeatures)
  - `ports/`: Interfaces de servicios (FileService, JiraService)

- **Infraestructura** (`internal/infrastructure/`): Adaptadores externos
  - `jira/`: Cliente API de Jira y gestión de features
  - `filesystem/`: Procesamiento de archivos (CSV/Excel)
  - `config/`: Configuración de entorno
  - `logger/`: Logging estructurado

- **Presentación** (`internal/presentation/`): Interfaz de usuario
  - `cli/`: Definiciones de comandos Cobra y configuración de aplicación
  - `formatters/`: Formateo de salida para diferentes tipos de resultados

### Patrones de Diseño Clave
- **Inyección de Dependencias**: Todas las dependencias inyectadas vía constructores
- **Patrón Repository**: Acceso a datos abstracto a través de interfaces
- **Patrón Use Case**: Cada operación de negocio es un caso de uso separado
- **Patrón Command**: Los comandos CLI encapsulan operaciones

### Flujo de Datos
1. Comandos CLI (`presentation/cli`) validan entrada y llaman casos de uso
2. Casos de uso (`application/usecases`) orquestan lógica de negocio
3. Repositorios (`domain/repositories`) definen contratos de acceso a datos
4. Adaptadores de infraestructura implementan interfaces de repositorio
5. Los resultados fluyen de vuelta a través de formateadores para salida al usuario

## Configuración

La aplicación usa variables de entorno cargadas desde `.env`:
- Credenciales Jira: `JIRA_URL`, `JIRA_EMAIL`, `JIRA_API_TOKEN`
- Configuración de proyecto: `PROJECT_KEY`, campos `*_ISSUE_TYPE`
- Directorios: `INPUT_DIRECTORY`, `LOGS_DIRECTORY`, `PROCESSED_DIRECTORY`
- Comportamiento de features: `ROLLBACK_ON_SUBTASK_FAILURE`, `FEATURE_REQUIRED_FIELDS`

## Procesamiento de Archivos

Los archivos de entrada en el directorio `entrada/` se procesan para crear:
- Historias de usuario con título, descripción, criterios de aceptación
- Subtareas opcionales (separadas por punto y coma o salto de línea)
- Feature padre opcional (clave existente o descripción para nueva feature)

Los resultados se guardan en `procesados/` con timestamps y logs detallados en `logs/`.

## Estructura de Testing

Los tests siguen las mismas capas arquitectónicas:
- Tests unitarios junto a archivos fuente (`*_test.go`)
- Mocks en `tests/mocks/` para dependencias externas
- Fixtures de test en `tests/fixtures/`
- Estructura de tests de integración en `tests/unit/`

## Agentes Personalizados

Los siguientes "agentes" están definidos como secuencias de comandos especializadas que Claude debe ejecutar cuando el usuario las solicite:

### qa-agent
**Comando**: `qa-agent`
**Propósito**: Agente de calidad completa para validación de código

**Secuencia de ejecución**:
1. Ejecutar `make lint` (fmt + vet)
2. Ejecutar `make test` con cobertura completa
3. Verificar que la cobertura sea ≥80% (excluyendo mocks/fixtures de `/tests/`)
4. Si hay errores de lint/format, repararlos automáticamente
5. Si hay tests fallando, analizar y sugerir correcciones
6. Re-ejecutar hasta que todos los checks pasen
7. Generar reporte final con estadísticas de cobertura y calidad

**Criterios de éxito**: 
- Todos los tests pasan
- Cobertura ≥80%
- Sin errores de lint/format
- Código compilable con `make build`

### build-agent
**Comando**: `build-agent`
**Propósito**: Agente de construcción y validación de binarios

**Secuencia de ejecución**:
1. Ejecutar `qa-agent` primero (prerequisito)
2. Ejecutar `make build` para compilar binario
3. Verificar que el binario se crea correctamente en `bin/historiador`
4. Probar comando básico: `./bin/historiador --help`
5. Ejecutar `make test-connection` para validar funcionalidad básica
6. Validar que el binario puede instalarse: `make install`
7. Generar reporte de build con tamaño del binario y dependencias

**Criterios de éxito**:
- qa-agent pasa completamente
- Binario generado sin errores
- Comandos básicos funcionan
- Instalación exitosa

### release-agent [VERSION]
**Comando**: `release-agent v1.2.3`
**Propósito**: Agente de preparación completa de release

**Secuencia de ejecución**:
1. Ejecutar `build-agent` (incluye qa-agent)
2. Verificar que la rama actual esté limpia (git status)
3. Actualizar versión en archivos relevantes si existen
4. Ejecutar suite completa de tests de aplicación:
   - `make test-connection`
   - `make validate` con archivos de test
   - `make diagnose` con proyecto de test
5. Crear tag git: `git tag -a {VERSION} -m "Release {VERSION}"`
6. Generar o actualizar CHANGELOG con cambios desde último tag
7. Crear commit de release si hay cambios de versión
8. Mostrar resumen final para revisión antes de push

**Criterios de éxito**:
- build-agent pasa completamente
- Todos los comandos de aplicación funcionan
- Tag git creado correctamente
- CHANGELOG actualizado

### coverage-agent [THRESHOLD]
**Comando**: `coverage-agent 85`
**Propósito**: Agente especializado en análisis de cobertura de tests

**Secuencia de ejecución**:
1. Ejecutar `go test -v -race -cover ./... -coverprofile=coverage_clean.out -coverpkg=./internal/...` (excluye mocks/fixtures)
2. Generar reporte HTML de cobertura: `go tool cover -html=coverage_clean.out -o coverage_clean.html`
3. Analizar archivos con cobertura insuficiente (solo código de producción)
4. Identificar funciones/métodos sin tests en lógica de negocio
5. Sugerir tests específicos para áreas no cubiertas
6. Si el threshold no se alcanza, crear TODOs específicos para mejorarlo
7. Generar reporte detallado con métricas por paquete

**Nota sobre exclusiones**:
- Los mocks (`tests/mocks/`) se excluyen porque son código de testing, no de producción
- Los fixtures (`tests/fixtures/`) se excluyen porque son datos de prueba helpers
- Los CLI commands pueden tener cobertura baja ya que requieren tests de integración/E2E
- Solo se mide cobertura de código en `./internal/...` (lógica de negocio)

**Criterios de éxito**:
- Cobertura total ≥ threshold especificado (default: 80%)
- Reporte HTML generado
- Identificación clara de gaps de testing

### fix-agent [TIPO]
**Comando**: `fix-agent lint` o `fix-agent tests` o `fix-agent all`
**Propósito**: Agente de reparación automática de problemas comunes

**Secuencia de ejecución**:
Para `lint`:
1. Ejecutar `make fmt` para formatear código
2. Ejecutar `make vet` y analizar warnings
3. Corregir problemas comunes de Go (variables no usadas, imports redundantes, etc.)
4. Re-ejecutar hasta que no haya warnings

Para `tests`:
1. Ejecutar tests y capturar fallos
2. Analizar errores de compilación y runtime
3. Sugerir correcciones específicas para cada fallo
4. Aplicar fixes automáticos para problemas comunes

Para `all`:
1. Ejecutar ambos secuencialmente
2. Asegurar que todas las correcciones son compatibles

**Criterios de éxito**:
- Sin errores de lint/format
- Tests compilables y ejecutables
- Código funcionalmente equivalente

## Uso de Agentes

Para usar cualquier agente, simplemente escribir su nombre como comando:

```bash
# Ejemplos de uso
qa-agent
build-agent  
release-agent v1.5.0
coverage-agent 85
fix-agent all
```

Los agentes son ejecutados secuencialmente y reportan progreso usando el sistema TodoWrite para tracking de tareas.

## Consideraciones de Cobertura de Tests

### Archivos Excluidos de Cobertura
Los siguientes tipos de archivos se excluyen del cálculo de cobertura por las razones indicadas:

- **`tests/mocks/`**: Código de testing auxiliar, no lógica de producción
- **`tests/fixtures/`**: Datos de prueba y helpers, no requieren cobertura
- **`cmd/main.go`**: Función main simple, se testea via tests de integración
- **CLI Commands**: Se testean mejor con tests E2E que unitarios

### Umbrales de Cobertura Recomendados
- **Código de negocio** (`internal/domain/`, `internal/application/`): **≥95%**
- **Infraestructura** (`internal/infrastructure/`): **≥85%**
- **Presentación** (`internal/presentation/formatters/`): **≥90%**
- **Total del proyecto** (excluyendo mocks/fixtures): **≥80%**

### Comando de Cobertura Estándar
```bash
# Cobertura correcta excluyendo archivos de testing
go test -v -race -cover ./... -coverprofile=coverage.out -coverpkg=./internal/...
go tool cover -html=coverage.out -o coverage.html
```