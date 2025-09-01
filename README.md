# Historiador Go

Aplicación CLI para crear historias de usuario en Jira desde archivos Excel/CSV con gestión automática de subtareas y Features.

## 📋 Uso del Ejecutable

### Comando Base (sin parámetros)
Procesa todos los archivos CSV/Excel del directorio `entrada/` usando configuración del archivo `.env`:
```bash
historiador
```
Si no existe `.env`, inicia configuración interactiva automáticamente.

### Comandos Específicos

#### `test-connection`
Verifica conectividad con Jira usando credenciales configuradas:
```bash
historiador test-connection
```

#### `process`
Procesa archivos CSV/Excel para crear historias de usuario en Jira:
```bash
# Procesar todos los archivos en entrada/
historiador process -p PROYECTO

# Procesar archivo específico
historiador process -f archivo.csv -p PROYECTO

# Modo dry-run (simula sin crear issues)
historiador process -p PROYECTO --dry-run

# Combinar opciones
historiador process -f archivo.csv -p PROYECTO --dry-run
```

#### `validate`
Valida formato de archivos sin conectar a Jira:
```bash
# Validar archivo específico
historiador validate -f archivo.csv

# Validar con validaciones específicas de proyecto
historiador validate -f archivo.csv -p PROYECTO
```

#### `diagnose`
Diagnostica configuración de Features en el proyecto Jira:
```bash
historiador diagnose -p PROYECTO
```

### Parámetros Globales
- `-p, --project`: Clave del proyecto Jira (ej: PROJ)
- `-f, --file`: Archivo específico a procesar
- `--dry-run`: Modo simulación (no crea issues)
- `--log-level`: Nivel de logging (DEBUG, INFO, WARN, ERROR)
- `-b, --batch-size`: Tamaño del lote de procesamiento (default: 10)
- `-h, --help`: Ayuda del comando

### Configuración Automática
Al ejecutar por primera vez sin archivo `.env`, se inicia configuración interactiva que:
- Solicita credenciales de Jira (URL, email, token)
- Detecta automáticamente tipos de issues del proyecto
- Configura campos personalizados disponibles
- Crea estructura de directorios necesaria

## 📋 Formato de Archivo

### Columnas Requeridas
- `titulo`: Título de la historia de usuario
- `descripcion`: Descripción detallada de la funcionalidad
- `criterio_aceptacion`: Criterios de aceptación separados por `;`

### Columnas Opcionales
- `subtareas`: Lista de subtareas separadas por `;` o salto de línea
- `parent`: Key de Feature existente (`PROJ-123`) **O** descripción para crear nueva Feature

### Ejemplo de Archivo CSV
```csv
titulo,descripcion,criterio_aceptacion,subtareas,parent
Login de usuario,Permitir autenticación de usuarios,Usuario puede ingresar credenciales; Sistema valida datos; Redirige al dashboard,Crear formulario; Validar backend; Manejar errores,Gestión de Usuarios
```

## ✨ Características

- ✅ **Configuración automática interactiva** al primer uso
- ✅ **Procesamiento automático** de archivos CSV/Excel
- ✅ **Creación automática de Features** desde descripciones
- ✅ **Subtareas automáticas** con validación avanzada
- ✅ **Prevención de duplicados** con normalización inteligente
- ✅ **Modo dry-run** para pruebas seguras
- ✅ **Rollback opcional** si fallan subtareas
- ✅ **Reportes detallados** de procesamiento
- ✅ **Detección automática** de campos personalizados de Jira

## ⚙️ Variables de Configuración (.env)

Las siguientes variables se configuran automáticamente durante la instalación inicial:

```env
# Conexión Jira
JIRA_URL=https://tuempresa.atlassian.net
JIRA_EMAIL=tu-email@empresa.com
JIRA_API_TOKEN=tu-token-api

# Proyecto
PROJECT_KEY=PROJ
SUBTASK_ISSUE_TYPE=Subtarea
FEATURE_ISSUE_TYPE=Feature
ACCEPTANCE_CRITERIA_FIELD=customfield_10001

# Comportamiento
ROLLBACK_ON_SUBTASK_FAILURE=false
FEATURE_REQUIRED_FIELDS=summary,description

# Directorios
INPUT_DIRECTORY=entrada
LOGS_DIRECTORY=logs
PROCESSED_DIRECTORY=procesados
```

## 📁 Estructura del Proyecto

```
historiadorgo/
├── cmd/                           # Punto de entrada
│   └── main.go
├── internal/
│   ├── domain/                    # Capa de dominio
│   │   ├── entities/              # Entidades de negocio
│   │   └── repositories/          # Interfaces de repositorios
│   ├── application/               # Capa de aplicación
│   │   ├── usecases/              # Casos de uso
│   │   └── ports/                 # Puertos/Interfaces
│   ├── infrastructure/            # Capa de infraestructura
│   │   ├── config/                # Configuración
│   │   ├── jira/                  # Adaptador Jira
│   │   └── filesystem/            # Adaptador archivos
│   └── presentation/              # Capa de presentación
│       ├── cli/                   # Comandos CLI
│       └── formatters/            # Formateo de salida
├── entrada/                       # Archivos a procesar
├── procesados/                    # Archivos procesados
├── logs/                          # Logs de ejecución
├── go.mod
├── go.sum
└── README.md
```

---

## 🔨 Desarrollo y Compilación

### Requisitos
- Go 1.21 o superior

### Compilar desde Código Fuente

1. **Configurar proyecto:**
   ```bash
   make setup  # Crea directorios y .env de ejemplo
   ```

2. **Compilar:**
   ```bash
   make build
   # o manualmente:
   go build -o historiador cmd/main.go
   ```

3. **Instalar en sistema:**
   ```bash
   make install  # Instala en /usr/local/bin/
   ```

### Comandos de Desarrollo

```bash
# Testing y calidad
make test              # Ejecutar todos los tests
make lint              # Formatear y validar código
make fmt               # Solo formatear código
make vet               # Solo análisis estático

# Construcción multiplataforma
make build-all         # Compilar para Linux, Windows, macOS

# Comandos de aplicación con make
make test-connection   # Probar conexión Jira
make validate FILE=archivo.csv
make process PROJECT=PROJ-KEY [FILE=archivo.csv] [DRY_RUN=true]
make diagnose PROJECT=PROJ-KEY
```

**Arquitectura**: Hexagonal (Ports & Adapters) con Clean Architecture  
**Documentación técnica**: Ver `CLAUDE.md` para detalles de desarrollo