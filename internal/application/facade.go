package application

import "hw2/internal/domain"

type GradebookRepository interface {
	Load() (*domain.Gradebook, error)
	Update(change func(*domain.Gradebook) error) error
}

type FinalReport struct {
	Student domain.Student
	Result  domain.Result
}

type ResultExporter interface {
	Export(path string, report FinalReport) error
}

type Facade struct {
	repository GradebookRepository
	exporter   ResultExporter
}

func NewFacade(repository GradebookRepository, exporter ResultExporter) *Facade {
	return &Facade{repository: repository, exporter: exporter}
}

func (f *Facade) AddStudent(id, name string) error {
	return f.repository.Update(func(gradebook *domain.Gradebook) error {
		return gradebook.AddStudent(id, name)
	})
}

func (f *Facade) Students() ([]domain.Student, error) {
	gradebook, err := f.repository.Load()
	if err != nil {
		return nil, err
	}
	return gradebook.Students(), nil
}

func (f *Facade) AddCriterion(id, name string, maxScore, weight int) error {
	return f.repository.Update(func(gradebook *domain.Gradebook) error {
		return gradebook.AddCriterion(id, name, maxScore, weight)
	})
}

func (f *Facade) Criteria() ([]domain.Criterion, error) {
	gradebook, err := f.repository.Load()
	if err != nil {
		return nil, err
	}
	return gradebook.Criteria(), nil
}

func (f *Facade) AddEvaluation(studentID, criterionID string, points int) error {
	return f.repository.Update(func(gradebook *domain.Gradebook) error {
		return gradebook.AddEvaluation(studentID, criterionID, points)
	})
}

func (f *Facade) CorrectEvaluation(studentID, criterionID string, points int) error {
	return f.repository.Update(func(gradebook *domain.Gradebook) error {
		return gradebook.CorrectEvaluation(studentID, criterionID, points)
	})
}

func (f *Facade) CancelEvaluation(studentID, criterionID string) error {
	return f.repository.Update(func(gradebook *domain.Gradebook) error {
		return gradebook.CancelEvaluation(studentID, criterionID)
	})
}

func (f *Facade) Evaluations(studentID string) ([]domain.Evaluation, error) {
	gradebook, err := f.repository.Load()
	if err != nil {
		return nil, err
	}
	return gradebook.Evaluations(studentID)
}

func (f *Facade) CurrentResult(studentID string) (domain.Result, error) {
	gradebook, err := f.repository.Load()
	if err != nil {
		return domain.Result{}, err
	}
	return domain.CalculateResult(gradebook, studentID, domain.CurrentResult{})
}

func (f *Facade) ExportFinalResult(studentID, path string) error {
	gradebook, err := f.repository.Load()
	if err != nil {
		return err
	}
	result, err := domain.CalculateResult(gradebook, studentID, domain.FinalResult{})
	if err != nil {
		return err
	}
	student, _ := gradebook.Student(studentID)
	return f.exporter.Export(path, FinalReport{Student: student, Result: result})
}
