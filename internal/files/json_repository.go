package files

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"hw2/internal/domain"
)

type JSONRepository struct {
	path string
}

func NewJSONRepository(path string) *JSONRepository {
	return &JSONRepository{path: path}
}

type jsonStudent struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type jsonCriterion struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	MaxScore int    `json:"max_score"`
	Weight   int    `json:"weight"`
}

type jsonEvaluation struct {
	StudentID   string `json:"student_id"`
	CriterionID string `json:"criterion_id"`
	Points      *int   `json:"points"`
}

type jsonGradebook struct {
	Students    []jsonStudent    `json:"students"`
	Criteria    []jsonCriterion  `json:"criteria"`
	Evaluations []jsonEvaluation `json:"evaluations"`
}

func (r *JSONRepository) Load() (*domain.Gradebook, error) {
	content, err := os.ReadFile(r.path)
	if errors.Is(err, os.ErrNotExist) {
		return &domain.Gradebook{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("прочитать ведомость: %w", err)
	}

	var data jsonGradebook
	if err := json.Unmarshal(content, &data); err != nil {
		return nil, fmt.Errorf("прочитать JSON ведомости: %w", err)
	}
	if data.Students == nil || data.Criteria == nil || data.Evaluations == nil {
		return nil, fmt.Errorf("в JSON ведомости нужны списки students, criteria и evaluations")
	}

	gradebook := &domain.Gradebook{}
	for i, student := range data.Students {
		if err := gradebook.AddStudent(student.ID, student.Name); err != nil {
			return nil, fmt.Errorf("студент в строке списка %d: %w", i+1, err)
		}
	}
	for i, criterion := range data.Criteria {
		if err := gradebook.AddCriterion(criterion.ID, criterion.Name, criterion.MaxScore, criterion.Weight); err != nil {
			return nil, fmt.Errorf("критерий в строке списка %d: %w", i+1, err)
		}
	}
	for i, evaluation := range data.Evaluations {
		if evaluation.Points == nil {
			return nil, fmt.Errorf("оценка %d: отсутствует поле points", i+1)
		}
		if err := gradebook.AddEvaluation(evaluation.StudentID, evaluation.CriterionID, *evaluation.Points); err != nil {
			return nil, fmt.Errorf("оценка %d: %w", i+1, err)
		}
	}
	return gradebook, nil
}

func (r *JSONRepository) Update(change func(*domain.Gradebook) error) error {
	lockPath := r.path + ".lock"
	lock, err := os.OpenFile(lockPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if errors.Is(err, os.ErrExist) {
		return fmt.Errorf("ведомость занята другой командой, повторите попытку")
	}
	if err != nil {
		return fmt.Errorf("подготовить запись ведомости: %w", err)
	}
	defer func() {
		lock.Close()
		os.Remove(lockPath)
	}()

	gradebook, err := r.Load()
	if err != nil {
		return err
	}
	if err := change(gradebook); err != nil {
		return err
	}
	if err := r.save(gradebook); err != nil {
		return fmt.Errorf("сохранить ведомость: %w", err)
	}
	return nil
}

func (r *JSONRepository) save(gradebook *domain.Gradebook) error {
	data := jsonGradebook{
		Students:    []jsonStudent{},
		Criteria:    []jsonCriterion{},
		Evaluations: []jsonEvaluation{},
	}
	for _, student := range gradebook.Students() {
		data.Students = append(data.Students, jsonStudent{ID: student.ID, Name: student.Name})
		evaluations, err := gradebook.Evaluations(student.ID)
		if err != nil {
			return err
		}
		for _, evaluation := range evaluations {
			points := evaluation.Score.Points()
			data.Evaluations = append(data.Evaluations, jsonEvaluation{
				StudentID: evaluation.StudentID, CriterionID: evaluation.CriterionID, Points: &points,
			})
		}
	}
	for _, criterion := range gradebook.Criteria() {
		data.Criteria = append(data.Criteria, jsonCriterion{
			ID: criterion.ID, Name: criterion.Name, MaxScore: criterion.MaxScore, Weight: criterion.Weight,
		})
	}

	file, err := os.CreateTemp(filepath.Dir(r.path), ".gradebook-*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	writeErr := encoder.Encode(data)
	closeErr := file.Close()
	if err := errors.Join(writeErr, closeErr); err != nil {
		return err
	}
	return os.Rename(file.Name(), r.path)
}
