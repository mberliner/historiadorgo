package jira

import (
	"strings"
)

// ADFDocument representa un documento en formato Atlassian Document Format
type ADFDocument struct {
	Type    string       `json:"type"`
	Version int          `json:"version"`
	Content []ADFContent `json:"content"`
}

// ADFContent representa el contenido de un documento ADF
type ADFContent struct {
	Type    string    `json:"type"`
	Content []ADFText `json:"content,omitempty"`
	Text    string    `json:"text,omitempty"`
	Marks   []ADFMark `json:"marks,omitempty"`
}

// ADFText representa texto dentro del contenido ADF
type ADFText struct {
	Type  string    `json:"type"`
	Text  string    `json:"text"`
	Marks []ADFMark `json:"marks,omitempty"`
}

// ADFMark representa marcado de texto (bold, italic, etc.)
type ADFMark struct {
	Type string `json:"type"`
}

// NewADFDocument crea un nuevo documento ADF vacío
func NewADFDocument() *ADFDocument {
	return &ADFDocument{
		Type:    "doc",
		Version: 1,
		Content: []ADFContent{},
	}
}

// AddParagraph agrega un párrafo al documento ADF
func (doc *ADFDocument) AddParagraph(text string) {
	paragraph := ADFContent{
		Type: "paragraph",
		Content: []ADFText{
			{
				Type: "text",
				Text: text,
			},
		},
	}
	doc.Content = append(doc.Content, paragraph)
}

// AddBulletList agrega una lista con bullets al documento ADF
func (doc *ADFDocument) AddBulletList(items []string) {
	for _, item := range items {
		if strings.TrimSpace(item) != "" {
			doc.AddParagraph("• " + strings.TrimSpace(item))
		}
	}
}

// CreateAcceptanceCriteriaADF crea un documento ADF para criterios de aceptación
func CreateAcceptanceCriteriaADF(criteriaText string) *ADFDocument {
	doc := NewADFDocument()

	if criteriaText == "" {
		return doc
	}

	// Dividir criterios por ';' o por líneas
	criteria := splitCriteria(criteriaText)

	if len(criteria) == 1 {
		// Un solo criterio - agregar como párrafo simple
		doc.AddParagraph(criteria[0])
	} else {
		// Múltiples criterios - agregar como lista con bullets
		doc.AddBulletList(criteria)
	}

	return doc
}

// CreateDescriptionWithCriteriaADF crea un documento ADF para descripción con criterios incluidos
func CreateDescriptionWithCriteriaADF(description, criteriaText string) *ADFDocument {
	doc := NewADFDocument()

	// Agregar descripción principal
	if description != "" {
		doc.AddParagraph(description)
	}

	// Agregar criterios de aceptación si existen
	if criteriaText != "" {
		// Agregar separador
		doc.AddParagraph("")
		doc.AddParagraph("--- Criterios de Aceptación ---")

		criteria := splitCriteria(criteriaText)

		if len(criteria) == 1 {
			doc.AddParagraph(criteria[0])
		} else {
			doc.AddBulletList(criteria)
		}
	}

	return doc
}

// CreateDescriptionADF crea un documento ADF simple para la descripción
func CreateDescriptionADF(description string) *ADFDocument {
	doc := NewADFDocument()

	if description != "" {
		doc.AddParagraph(description)
	}

	return doc
}

// splitCriteria divide el texto de criterios por ';' o por líneas
func splitCriteria(criteriaText string) []string {
	var criteria []string

	// Primero intentar dividir por ';'
	if strings.Contains(criteriaText, ";") {
		parts := strings.Split(criteriaText, ";")
		for _, part := range parts {
			if trimmed := strings.TrimSpace(part); trimmed != "" {
				criteria = append(criteria, trimmed)
			}
		}
	} else {
		// Si no hay ';', dividir por líneas
		lines := strings.Split(criteriaText, "\n")
		for _, line := range lines {
			if trimmed := strings.TrimSpace(line); trimmed != "" {
				criteria = append(criteria, trimmed)
			}
		}
	}

	// Si no se encontraron múltiples criterios, usar el texto completo
	if len(criteria) == 0 {
		criteria = append(criteria, strings.TrimSpace(criteriaText))
	}

	return criteria
}
