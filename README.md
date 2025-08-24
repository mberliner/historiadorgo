# Historiador Go

Aplicación CLI para crear historias de usuario en Jira desde archivos Excel/CSV con gestión automática de subtareas y Features. Implementación en Go con arquitectura hexagonal.

## 🚀 Inicio Rápido

**Requisitos**: Go 1.21+

1. **Instalar dependencias:**
   ```bash
   go mod tidy
   ```

2. **Configurar credenciales:**
   ```bash
   cp .env.example .env
   # Editar .env con tus credenciales de Jira
   ```

3. **Compilar:**
   ```bash
   go build -o historiador cmd/main.go
   ```

4. **Probar conexión:**
   ```bash
   ./historiador test-connection
   ```

5. **Procesar archivos:**
   ```bash
   # Modo prueba
   ./historiador process -p TU_PROYECTO --dry-run
   
   # Procesamiento real
   ./historiador process -p TU_PROYECTO
   ```

## ✨ Características

- ✅ **Procesamiento automático** de archivos CSV/Excel
- ✅ **Creación automática de Features** desde descripciones
- ✅ **Subtareas automáticas** con validación avanzada
- ✅ **Prevención de duplicados** con normalización inteligente
- ✅ **Modo dry-run** para pruebas seguras
- ✅ **Rollback opcional** si fallan subtareas
- ✅ **Reportes detallados** de procesamiento

## 📋 Formato de Archivo

### Columnas Requeridas
- `titulo`: Título de la historia
- `descripcion`: Descripción detallada  
- `criterio_aceptacion`: Criterios separados por `;`

### Columnas Opcionales
- `subtareas`: Lista separada por `;` o salto de línea
- `parent`: Key existente (`PROJ-123`) **O** descripción para crear Feature

## 🔧 Comandos

```bash
# Procesar todos los archivos en entrada/
./historiador process -p PROYECTO

# Procesar archivo específico
./historiador process -f archivo.csv -p PROYECTO

# Validar archivo sin crear issues
./historiador validate -f archivo.csv

# Diagnosticar configuración
./historiador diagnose -p PROYECTO

# Modo dry-run
./historiador process -p PROYECTO --dry-run
```

## ⚙️ Configuración (.env)

```env
JIRA_URL=https://empresa.atlassian.net
JIRA_EMAIL=email@empresa.com
JIRA_API_TOKEN=tu-token-aqui
PROJECT_KEY=PROJ
SUBTASK_ISSUE_TYPE=Subtarea
FEATURE_ISSUE_TYPE=Feature
ACCEPTANCE_CRITERIA_FIELD=customfield_10001
ROLLBACK_ON_SUBTASK_FAILURE=false
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

**Arquitectura**: Hexagonal (Ports & Adapters) con Clean Architecture
**Portado desde**: Proyecto Python historiador con misma funcionalidad