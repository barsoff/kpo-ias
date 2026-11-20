package domain

import (
	"errors"
	"fmt"
)

var (
	ErrEmptyRubric      = errors.New("в рубрикаторе пока нет критериев")
	ErrNoEvaluations    = errors.New("у студента пока нет оценок")
	ErrIncompleteRubric = errors.New("для окончательного итога нужны оценки по всем критериям")
)

type Result struct {
	Percent           float64
	EvaluatedCriteria int
	TotalCriteria     int
}

type ResultStrategy interface {
	checkCompleteness(evaluated, total int) error
}

type CurrentResult struct{}

func (CurrentResult) checkCompleteness(evaluated, _ int) error {
	if evaluated == 0 {
		return ErrNoEvaluations
	}
	return nil
}

type FinalResult struct{}

func (FinalResult) checkCompleteness(evaluated, total int) error {
	if evaluated < total {
		return ErrIncompleteRubric
	}
	return nil
}

func CalculateResult(g *Gradebook, studentID string, strategy ResultStrategy) (Result, error) {
	if _, exists := g.students[studentID]; !exists {
		return Result{}, fmt.Errorf("студент %q не найден", studentID)
	}
	criteria := g.Criteria()
	result := Result{TotalCriteria: len(criteria)}
	if result.TotalCriteria == 0 {
		return result, ErrEmptyRubric
	}

	var weightedScore, evaluatedWeight float64
	for _, criterion := range criteria {
		key := evaluationKey{studentID: studentID, criterionID: criterion.ID}
		evaluation, exists := g.evaluations[key]
		if !exists {
			continue
		}
		weight := float64(criterion.Weight)
		weightedScore += float64(evaluation.Score.Points()) / float64(criterion.MaxScore) * weight
		evaluatedWeight += weight
		result.EvaluatedCriteria++
	}

	if strategy == nil {
		return result, errors.New("укажите режим расчёта итога")
	}
	if err := strategy.checkCompleteness(result.EvaluatedCriteria, result.TotalCriteria); err != nil {
		return result, err
	}
	result.Percent = weightedScore / evaluatedWeight * 100
	return result, nil
}
