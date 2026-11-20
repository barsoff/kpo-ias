package domain

import "fmt"

type Score struct {
	points int
}

func (s Score) Points() int {
	return s.points
}

func newScore(points, maxScore int) (Score, error) {
	if points < 0 || points > maxScore {
		return Score{}, fmt.Errorf("балл %d вне диапазона от 0 до %d", points, maxScore)
	}
	return Score{points: points}, nil
}

type Evaluation struct {
	StudentID   string
	CriterionID string
	Score       Score
}

func (g *Gradebook) newEvaluation(studentID, criterionID string, points int) (Evaluation, error) {
	if _, exists := g.students[studentID]; !exists {
		return Evaluation{}, fmt.Errorf("студент %q не найден", studentID)
	}
	criterion, exists := g.criteria[criterionID]
	if !exists {
		return Evaluation{}, fmt.Errorf("критерий %q не найден в рубрикаторе", criterionID)
	}
	score, err := newScore(points, criterion.MaxScore)
	if err != nil {
		return Evaluation{}, fmt.Errorf("критерий %q: %w", criterionID, err)
	}
	return Evaluation{StudentID: studentID, CriterionID: criterionID, Score: score}, nil
}
