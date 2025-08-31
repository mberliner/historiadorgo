package fixtures

import (
	"historiadorgo/internal/domain/entities"
	"historiadorgo/internal/infrastructure/config"
)

// GetSampleUserStories returns a set of sample user stories for testing
func GetSampleUserStories() []*entities.UserStory {
	return []*entities.UserStory{
		entities.NewUserStory(
			"Login de usuario",
			"Como usuario quiero poder autenticarme en el sistema",
			"Usuario puede loguearse con credenciales válidas;Error visible con credenciales inválidas",
			"Crear formulario de login;Validar credenciales;Mostrar errores",
			"DEMO-100",
		),
		entities.NewUserStory(
			"Dashboard principal",
			"Como usuario autenticado quiero ver mi dashboard",
			"Dashboard carga en menos de 3 segundos;Datos están actualizados",
			"Crear componente dashboard;Conectar con API;Optimizar performance",
			"Sistema de Gestión de Usuarios",
		),
		entities.NewUserStory(
			"Historia sin subtareas",
			"Una historia simple sin subtareas",
			"Criterio simple",
			"",
			"",
		),
		entities.NewUserStory(
			"Historia con subtareas inválidas",
			"Historia con subtareas problemáticas",
			"Debe manejar subtareas inválidas",
			"Subtarea válida;;Subtarea muy larga que excede los 255 caracteres permitidos en el sistema de gestión de proyectos y por lo tanto debería ser considerada como inválida y no procesada correctamente por el validador de longitud de texto máximo permitido en los campos de subtareas del sistema",
			"PROJ-200",
		),
	}
}

// GetSampleUserStoriesWithMixedParents returns stories with different parent types
func GetSampleUserStoriesWithMixedParents() []*entities.UserStory {
	return []*entities.UserStory{
		entities.NewUserStory(
			"Historia con key existente",
			"Historia vinculada a issue existente",
			"Debe vincularse correctamente",
			"Subtarea 1;Subtarea 2",
			"PROJ-123",
		),
		entities.NewUserStory(
			"Historia para crear feature",
			"Historia que debe crear un feature automáticamente",
			"Feature se crea automáticamente",
			"Implementar feature;Documentar feature",
			"Nuevo Sistema de Reportes",
		),
		entities.NewUserStory(
			"Historia sin parent",
			"Historia independiente",
			"Funciona sin parent",
			"",
			"",
		),
	}
}

// GetEmptyUserStories returns an empty slice for testing edge cases
func GetEmptyUserStories() []*entities.UserStory {
	return []*entities.UserStory{}
}

// GetSingleUserStory returns a single story for simple tests
func GetSingleUserStory() []*entities.UserStory {
	return []*entities.UserStory{
		entities.NewUserStory(
			"Historia única",
			"Una sola historia para pruebas simples",
			"Debe procesarse correctamente",
			"Única subtarea",
			"",
		),
	}
}

// ValidUserStory1 returns a valid user story for testing
func ValidUserStory1() *entities.UserStory {
	story := entities.NewUserStory(
		"Login de usuario",
		"Como usuario quiero poder autenticarme en el sistema",
		"Usuario puede loguearse con credenciales válidas",
		"Crear formulario;Validar credenciales",
		"",
	)
	story.Row = 1
	return story
}

// UserStoryWithParent returns a user story with a parent feature
func UserStoryWithParent() *entities.UserStory {
	story := entities.NewUserStory(
		"Historia con feature",
		"Historia que requiere un feature padre",
		"Debe vincularse al feature correctamente",
		"Subtarea 1;Subtarea 2",
		"Nuevo Sistema de Reportes",
	)
	story.Row = 1
	return story
}

// UserStoryWithSubtasks returns a user story with subtasks
func UserStoryWithSubtasks() *entities.UserStory {
	story := entities.NewUserStory(
		"Historia con subtareas",
		"Historia que tiene múltiples subtareas",
		"Todas las subtareas deben crearse",
		"Primera subtarea;Segunda subtarea;Tercera subtarea",
		"",
	)
	story.Row = 1
	return story
}

// ValidUserStory2 returns another valid user story for testing
func ValidUserStory2() *entities.UserStory {
	story := entities.NewUserStory(
		"Dashboard principal",
		"Como usuario autenticado quiero ver mi dashboard",
		"Dashboard carga en menos de 3 segundos",
		"Crear dashboard;Conectar API",
		"PROJ-100",
	)
	story.Row = 2
	return story
}

// SuccessProcessResult returns a successful process result
func SuccessProcessResult() *entities.ProcessResult {
	result := entities.NewProcessResult(1)
	result.Success = true
	result.IssueKey = "PROJ-123"
	result.IssueURL = "https://test.atlassian.net/browse/PROJ-123"
	return result
}

// ErrorProcessResult returns a failed process result
func ErrorProcessResult() *entities.ProcessResult {
	result := entities.NewProcessResult(1)
	result.Success = false
	result.ErrorMessage = "Failed to create issue"
	return result
}

// NewBatchResult returns a sample batch result for testing
func NewBatchResult() *entities.BatchResult {
	result := entities.NewBatchResult("test-file.csv", 2, false)
	result.AddResult(SuccessProcessResult())
	result.Finish()
	return result
}

// ValidationResultForTest represents a test validation result to avoid import cycles
type ValidationResultForTest struct {
	TotalStories    int
	WithSubtasks    int
	TotalSubtasks   int
	WithParent      int
	InvalidSubtasks int
	Preview         string
}

// NewValidationResult returns a sample validation result for testing
func NewValidationResult() *ValidationResultForTest {
	return &ValidationResultForTest{
		TotalStories:    2,
		WithSubtasks:    1,
		TotalSubtasks:   3,
		WithParent:      1,
		InvalidSubtasks: 0,
		Preview:         "Sample preview",
	}
}

// TestConfig returns a test configuration
var TestConfig = &config.Config{
	JiraURL:                  "https://test.atlassian.net",
	JiraEmail:                "test@example.com",
	JiraAPIToken:             "test-token",
	ProjectKey:               "TEST",
	DefaultIssueType:         "Story",
	SubtaskIssueType:         "Sub-task",
	FeatureIssueType:         "Epic",
	InputDirectory:           "entrada",
	ProcessedDirectory:       "procesados",
	LogsDirectory:            "logs",
	RollbackOnSubtaskFailure: true,
	FeatureRequiredFields:    "summary,description",
}
