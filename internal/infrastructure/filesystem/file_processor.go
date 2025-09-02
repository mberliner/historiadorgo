package filesystem

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"historiadorgo/internal/domain/entities"

	"github.com/go-playground/validator/v10"
	"github.com/gocarina/gocsv"
	"github.com/xuri/excelize/v2"
)

const (
	csvExtension  = ".csv"
	xlsxExtension = ".xlsx"
	xlsExtension  = ".xls"
)

type FileProcessor struct {
	validator    *validator.Validate
	processedDir string
}

type CSVRecord struct {
	Titulo             string `csv:"titulo"`
	Descripcion        string `csv:"descripcion"`
	Subtareas          string `csv:"subtareas"`
	CriterioAceptacion string `csv:"criterio_aceptacion"`
	Parent             string `csv:"parent"`
}

func NewFileProcessor(processedDir string) *FileProcessor {
	return &FileProcessor{
		validator:    validator.New(),
		processedDir: processedDir,
	}
}

func (fp *FileProcessor) ReadFile(ctx context.Context, filePath string) ([]*entities.UserStory, error) {
	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case csvExtension:
		return fp.readCSV(filePath)
	case xlsxExtension, xlsExtension:
		return fp.readExcel(filePath)
	default:
		return nil, fmt.Errorf("unsupported file format: %s", ext)
	}
}

func (fp *FileProcessor) ValidateFile(ctx context.Context, filePath string) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", filePath)
	}

	ext := strings.ToLower(filepath.Ext(filePath))
	if ext != csvExtension && ext != xlsxExtension && ext != xlsExtension {
		return fmt.Errorf("unsupported file format: %s. Supported formats: %s, %s, %s", ext, csvExtension, xlsxExtension, xlsExtension)
	}

	stories, err := fp.ReadFile(ctx, filePath)
	if err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	if len(stories) == 0 {
		return fmt.Errorf("file contains no valid stories")
	}

	for i, story := range stories {
		if err := fp.validator.Struct(story); err != nil {
			return fmt.Errorf("validation error in row %d: %w", i+2, err)
		}
	}

	return nil
}

func (fp *FileProcessor) MoveToProcessed(ctx context.Context, filePath string) error {
	if err := os.MkdirAll(fp.processedDir, 0755); err != nil {
		return fmt.Errorf("error creating processed directory: %w", err)
	}

	fileName := filepath.Base(filePath)
	destPath := filepath.Join(fp.processedDir, fileName)

	return os.Rename(filePath, destPath)
}

func (fp *FileProcessor) GetPendingFiles(ctx context.Context, inputDir string) ([]string, error) {
	var files []string

	err := filepath.Walk(inputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			ext := strings.ToLower(filepath.Ext(path))
			if ext == ".csv" || ext == ".xlsx" || ext == ".xls" {
				files = append(files, path)
			}
		}

		return nil
	})

	return files, err
}

func (fp *FileProcessor) readCSV(filePath string) ([]*entities.UserStory, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error opening CSV file: %w", err)
	}
	defer file.Close()

	var records []*CSVRecord
	if err := gocsv.UnmarshalFile(file, &records); err != nil {
		return nil, fmt.Errorf("error parsing CSV: %w", err)
	}

	var stories []*entities.UserStory
	for _, record := range records {
		if record.Titulo == "" || record.Descripcion == "" || record.CriterioAceptacion == "" {
			continue
		}

		story := entities.NewUserStory(
			record.Titulo,
			record.Descripcion,
			record.CriterioAceptacion,
			record.Subtareas,
			record.Parent,
		)
		stories = append(stories, story)
	}

	return stories, nil
}

func (fp *FileProcessor) readExcel(filePath string) ([]*entities.UserStory, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error opening Excel file: %w", err)
	}
	defer f.Close()

	sheetName := f.GetSheetName(0)
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("error reading Excel rows: %w", err)
	}

	if len(rows) < 2 {
		return nil, fmt.Errorf("Excel file must have at least a header row and one data row")
	}

	header := rows[0]
	columnMap := fp.mapColumns(header)

	var stories []*entities.UserStory
	for i, row := range rows[1:] {
		if len(row) == 0 {
			continue
		}

		record := fp.parseExcelRow(row, columnMap)
		if record.Titulo == "" || record.Descripcion == "" || record.CriterioAceptacion == "" {
			continue
		}

		story := entities.NewUserStory(
			record.Titulo,
			record.Descripcion,
			record.CriterioAceptacion,
			record.Subtareas,
			record.Parent,
		)

		if err := fp.validator.Struct(story); err != nil {
			return nil, fmt.Errorf("validation error in row %d: %w", i+2, err)
		}

		stories = append(stories, story)
	}

	return stories, nil
}

func (fp *FileProcessor) mapColumns(header []string) map[string]int {
	columnMap := make(map[string]int)

	for i, col := range header {
		switch strings.ToLower(strings.TrimSpace(col)) {
		case "titulo":
			columnMap["titulo"] = i
		case "descripcion":
			columnMap["descripcion"] = i
		case "subtareas":
			columnMap["subtareas"] = i
		case "criterio_aceptacion":
			columnMap["criterio_aceptacion"] = i
		case "parent":
			columnMap["parent"] = i
		}
	}

	return columnMap
}

func (fp *FileProcessor) parseExcelRow(row []string, columnMap map[string]int) *CSVRecord {
	record := &CSVRecord{}

	if idx, exists := columnMap["titulo"]; exists && idx < len(row) {
		record.Titulo = strings.TrimSpace(row[idx])
	}
	if idx, exists := columnMap["descripcion"]; exists && idx < len(row) {
		record.Descripcion = strings.TrimSpace(row[idx])
	}
	if idx, exists := columnMap["subtareas"]; exists && idx < len(row) {
		record.Subtareas = strings.TrimSpace(row[idx])
	}
	if idx, exists := columnMap["criterio_aceptacion"]; exists && idx < len(row) {
		record.CriterioAceptacion = strings.TrimSpace(row[idx])
	}
	if idx, exists := columnMap["parent"]; exists && idx < len(row) {
		record.Parent = strings.TrimSpace(row[idx])
	}

	return record
}
