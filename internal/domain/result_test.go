package domain_test

import (
	"errors"
	"math"
	"testing"

	"hw2/internal/domain"
)

func TestCalculateResultWithPartialAndCompleteRubric(t *testing.T) {
	tests := []struct {
		name          string
		hasProject    bool
		projectPoints int
		strategy      domain.ResultStrategy
		wantPercent   float64
		wantEvaluated int
		wantError     error
	}{
		{name: "current result with a missing evaluation", strategy: domain.CurrentResult{}, wantPercent: 80, wantEvaluated: 1},
		{name: "final result requires the missing evaluation", strategy: domain.FinalResult{}, wantEvaluated: 1, wantError: domain.ErrIncompleteRubric},
		{name: "current result with a complete rubric", hasProject: true, projectPoints: 15, strategy: domain.CurrentResult{}, wantPercent: 78, wantEvaluated: 2},
		{name: "final result with a complete rubric", hasProject: true, projectPoints: 15, strategy: domain.FinalResult{}, wantPercent: 78, wantEvaluated: 2},
		{name: "recorded zero counts as complete", hasProject: true, strategy: domain.FinalResult{}, wantPercent: 48, wantEvaluated: 2},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gradebook := newTestGradebook(t)
			if err := gradebook.SetEvaluation(studentID, "labs", 8); err != nil {
				t.Fatal(err)
			}
			if test.hasProject {
				if err := gradebook.SetEvaluation(studentID, "project", test.projectPoints); err != nil {
					t.Fatal(err)
				}
			}

			result, err := domain.CalculateResult(gradebook, studentID, test.strategy)
			if !errors.Is(err, test.wantError) {
				t.Fatalf("CalculateResult error = %v, want %v", err, test.wantError)
			}
			if result.EvaluatedCriteria != test.wantEvaluated || result.TotalCriteria != 2 {
				t.Errorf("progress = %d/%d, want %d/2", result.EvaluatedCriteria, result.TotalCriteria, test.wantEvaluated)
			}
			if test.wantError == nil && !(math.Abs(result.Percent-test.wantPercent) < 1e-9) {
				t.Errorf("percent = %v, want %v", result.Percent, test.wantPercent)
			}
		})
	}
}

func TestEvaluationChangesAffectOnlySelectedStudent(t *testing.T) {
	gradebook := newTestGradebook(t)
	const secondID = "student-2"
	if err := gradebook.AddStudent(secondID, "Анна Иванова"); err != nil {
		t.Fatal(err)
	}
	if err := gradebook.AddEvaluation(studentID, "labs", 8); err != nil {
		t.Fatal(err)
	}
	if err := gradebook.AddEvaluation(secondID, "labs", 9); err != nil {
		t.Fatal(err)
	}
	checkSecondStudent := func() {
		t.Helper()
		evaluations, err := gradebook.Evaluations(secondID)
		if err != nil || len(evaluations) != 1 {
			t.Fatalf("second student's evaluations = %+v, %v", evaluations, err)
		}
		evaluation := evaluations[0]
		if evaluation.StudentID != secondID || evaluation.CriterionID != "labs" || evaluation.Score.Points() != 9 {
			t.Fatalf("unexpected evaluation for second student: %+v", evaluation)
		}
		result, err := domain.CalculateResult(gradebook, secondID, domain.CurrentResult{})
		if err != nil || result.Percent != 90 || result.EvaluatedCriteria != 1 || result.TotalCriteria != 2 {
			t.Fatalf("second student's result = %+v, %v, want 90 percent and progress 1/2", result, err)
		}
	}
	if err := gradebook.CorrectEvaluation(studentID, "labs", 5); err != nil {
		t.Fatal(err)
	}
	result, err := domain.CalculateResult(gradebook, studentID, domain.CurrentResult{})
	if err != nil || result.Percent != 50 {
		t.Fatalf("result after correction = %+v, %v, want 50 percent", result, err)
	}
	checkSecondStudent()

	if err := gradebook.CancelEvaluation(studentID, "labs"); err != nil {
		t.Fatal(err)
	}
	result, err = domain.CalculateResult(gradebook, studentID, domain.CurrentResult{})
	if !errors.Is(err, domain.ErrNoEvaluations) || result.EvaluatedCriteria != 0 || result.TotalCriteria != 2 {
		t.Fatalf("result after cancellation = %+v, %v, want no result and progress 0/2", result, err)
	}
	checkSecondStudent()
}

func TestCalculateResultWithEmptyRubric(t *testing.T) {
	gradebook := &domain.Gradebook{}
	if err := gradebook.AddStudent(studentID, "Анна Иванова"); err != nil {
		t.Fatal(err)
	}

	for _, strategy := range []domain.ResultStrategy{domain.CurrentResult{}, domain.FinalResult{}} {
		result, err := domain.CalculateResult(gradebook, studentID, strategy)
		if !errors.Is(err, domain.ErrEmptyRubric) || result != (domain.Result{}) {
			t.Errorf("result with %T = %+v, %v, want empty rubric", strategy, result, err)
		}
	}
}
