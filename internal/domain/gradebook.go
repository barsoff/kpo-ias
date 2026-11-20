package domain

import (
	"fmt"
	"sort"
	"strings"
)

type Student struct {
	ID   string
	Name string
}

type Criterion struct {
	ID       string
	Name     string
	MaxScore int
	Weight   int
}

type evaluationKey struct {
	studentID   string
	criterionID string
}

type Gradebook struct {
	students    map[string]Student
	criteria    map[string]Criterion
	evaluations map[evaluationKey]Evaluation
}

func (g *Gradebook) AddStudent(id, name string) error {
	id, name = strings.TrimSpace(id), strings.TrimSpace(name)
	if id == "" || name == "" {
		return fmt.Errorf("ID и имя студента обязательны")
	}
	if _, exists := g.students[id]; exists {
		return fmt.Errorf("студент с ID %q уже существует", id)
	}
	if g.students == nil {
		g.students = make(map[string]Student)
	}
	g.students[id] = Student{ID: id, Name: name}
	return nil
}

func (g *Gradebook) AddCriterion(id, name string, maxScore, weight int) error {
	id, name = strings.TrimSpace(id), strings.TrimSpace(name)
	if id == "" || name == "" {
		return fmt.Errorf("ID и название критерия обязательны")
	}
	if maxScore <= 0 || weight <= 0 {
		return fmt.Errorf("максимальный балл и вес критерия должны быть положительными")
	}
	if _, exists := g.criteria[id]; exists {
		return fmt.Errorf("критерий с ID %q уже существует", id)
	}
	if g.criteria == nil {
		g.criteria = make(map[string]Criterion)
	}
	g.criteria[id] = Criterion{ID: id, Name: name, MaxScore: maxScore, Weight: weight}
	return nil
}

func (g *Gradebook) Students() []Student {
	students := make([]Student, 0, len(g.students))
	for _, student := range g.students {
		students = append(students, student)
	}
	sort.Slice(students, func(i, j int) bool { return students[i].ID < students[j].ID })
	return students
}

func (g *Gradebook) Student(id string) (Student, bool) {
	student, exists := g.students[id]
	return student, exists
}

func (g *Gradebook) Criteria() []Criterion {
	criteria := make([]Criterion, 0, len(g.criteria))
	for _, criterion := range g.criteria {
		criteria = append(criteria, criterion)
	}
	sort.Slice(criteria, func(i, j int) bool { return criteria[i].ID < criteria[j].ID })
	return criteria
}

func (g *Gradebook) AddEvaluation(studentID, criterionID string, points int) error {
	key := evaluationKey{studentID: studentID, criterionID: criterionID}
	if _, exists := g.evaluations[key]; exists {
		return fmt.Errorf("у студента %q уже есть оценка по критерию %q", studentID, criterionID)
	}
	return g.SetEvaluation(studentID, criterionID, points)
}

func (g *Gradebook) CorrectEvaluation(studentID, criterionID string, points int) error {
	key := evaluationKey{studentID: studentID, criterionID: criterionID}
	if _, exists := g.evaluations[key]; !exists {
		return fmt.Errorf("у студента %q нет оценки по критерию %q", studentID, criterionID)
	}
	return g.SetEvaluation(studentID, criterionID, points)
}

func (g *Gradebook) SetEvaluation(studentID, criterionID string, points int) error {
	evaluation, err := g.newEvaluation(studentID, criterionID, points)
	if err != nil {
		return err
	}
	if g.evaluations == nil {
		g.evaluations = make(map[evaluationKey]Evaluation)
	}
	key := evaluationKey{studentID: studentID, criterionID: criterionID}
	g.evaluations[key] = evaluation
	return nil
}

func (g *Gradebook) CancelEvaluation(studentID, criterionID string) error {
	key := evaluationKey{studentID: studentID, criterionID: criterionID}
	if _, exists := g.evaluations[key]; !exists {
		return fmt.Errorf("у студента %q нет оценки по критерию %q", studentID, criterionID)
	}
	delete(g.evaluations, key)
	return nil
}

func (g *Gradebook) Evaluations(studentID string) ([]Evaluation, error) {
	if _, exists := g.students[studentID]; !exists {
		return nil, fmt.Errorf("студент %q не найден", studentID)
	}
	evaluations := []Evaluation{}
	for key, evaluation := range g.evaluations {
		if key.studentID == studentID {
			evaluations = append(evaluations, evaluation)
		}
	}
	sort.Slice(evaluations, func(i, j int) bool {
		return evaluations[i].CriterionID < evaluations[j].CriterionID
	})
	return evaluations, nil
}
