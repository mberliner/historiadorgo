package entities

import (
	"strings"
)

type UserStory struct {
	Titulo             string   `json:"titulo" validate:"required,min=1,max=255"`
	Descripcion        string   `json:"descripcion" validate:"required,min=1"`
	CriterioAceptacion string   `json:"criterio_aceptacion" validate:"required,min=1"`
	Subtareas          []string `json:"subtareas,omitempty"`
	Parent             string   `json:"parent,omitempty"`
	Row                int      `json:"row,omitempty"`
}

func NewUserStory(titulo, descripcion, criterioAceptacion string, subtareasRaw, parent string) *UserStory {
	story := &UserStory{
		Titulo:             titulo,
		Descripcion:        descripcion,
		CriterioAceptacion: criterioAceptacion,
		Parent:             parent,
	}

	story.parseSubtareas(subtareasRaw)
	return story
}

func (us *UserStory) parseSubtareas(subtareasRaw string) {
	if subtareasRaw == "" {
		return
	}

	var tasks []string

	for _, part := range strings.Split(subtareasRaw, ";") {
		for _, task := range strings.Split(part, "\n") {
			if trimmed := strings.TrimSpace(task); trimmed != "" {
				tasks = append(tasks, trimmed)
			}
		}
	}

	if len(tasks) > 0 {
		us.Subtareas = tasks
	}
}

func (us *UserStory) HasSubtareas() bool {
	return len(us.Subtareas) > 0
}

func (us *UserStory) HasParent() bool {
	return us.Parent != ""
}

func (us *UserStory) GetValidSubtareas() []string {
	var valid []string
	for _, subtarea := range us.Subtareas {
		if len(subtarea) > 0 && len(subtarea) <= 255 {
			valid = append(valid, subtarea)
		}
	}
	return valid
}
