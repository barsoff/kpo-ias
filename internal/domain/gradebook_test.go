package domain_test

import (
	"testing"

	"hw2/internal/domain"
)

const studentID = "student-1"

func newTestGradebook(t *testing.T) *domain.Gradebook {
	t.Helper()
	gradebook := &domain.Gradebook{}
	if err := gradebook.AddStudent(studentID, "Анна Иванова"); err != nil {
		t.Fatal(err)
	}
	if err := gradebook.AddCriterion("labs", "Лабораторные", 10, 60); err != nil {
		t.Fatal(err)
	}
	if err := gradebook.AddCriterion("project", "Проект", 20, 40); err != nil {
		t.Fatal(err)
	}
	return gradebook
}

func TestSetEvaluationValidatesStudentCriterionAndScore(t *testing.T) {
	tests := []struct {
		name        string
		studentID   string
		criterionID string
		points      int
		wantError   bool
	}{
		{name: "zero", studentID: studentID, criterionID: "labs", points: 0},
		{name: "criterion maximum", studentID: studentID, criterionID: "labs", points: 10},
		{name: "negative points", studentID: studentID, criterionID: "labs", points: -1, wantError: true},
		{name: "above maximum", studentID: studentID, criterionID: "labs", points: 11, wantError: true},
		{name: "unknown criterion", studentID: studentID, criterionID: "exam", points: 8, wantError: true},
		{name: "unknown student", studentID: "student-2", criterionID: "labs", points: 8, wantError: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gradebook := newTestGradebook(t)
			err := gradebook.SetEvaluation(test.studentID, test.criterionID, test.points)
			if (err != nil) != test.wantError {
				t.Fatalf("SetEvaluation error = %v, want error: %t", err, test.wantError)
			}

			evaluations, err := gradebook.Evaluations(studentID)
			if err != nil {
				t.Fatal(err)
			}
			if test.wantError {
				if len(evaluations) != 0 {
					t.Fatalf("rejected evaluation was recorded: %+v", evaluations)
				}
				return
			}
			if len(evaluations) != 1 || evaluations[0].Score.Points() != test.points {
				t.Fatalf("evaluations = %+v, want one evaluation with %d points", evaluations, test.points)
			}
		})
	}
}

func TestCorrectionAndCancellation(t *testing.T) {
	gradebook := newTestGradebook(t)
	if err := gradebook.SetEvaluation(studentID, "labs", 8); err != nil {
		t.Fatal(err)
	}
	if err := gradebook.SetEvaluation(studentID, "labs", 7); err != nil {
		t.Fatal(err)
	}
	if err := gradebook.SetEvaluation(studentID, "labs", 12); err == nil {
		t.Fatal("out-of-range correction succeeded")
	}
	evaluations, err := gradebook.Evaluations(studentID)
	if err != nil {
		t.Fatal(err)
	}
	if len(evaluations) != 1 || evaluations[0].Score.Points() != 7 {
		t.Fatalf("evaluations = %+v, want one evaluation with 7 points", evaluations)
	}

	if err := gradebook.CancelEvaluation(studentID, "labs"); err != nil {
		t.Fatal(err)
	}
	evaluations, err = gradebook.Evaluations(studentID)
	if err != nil || len(evaluations) != 0 {
		t.Fatalf("Evaluations after cancellation = %+v, %v, want empty list", evaluations, err)
	}
}

func TestGradebookRejectsDuplicateIDsAndInvalidCriterion(t *testing.T) {
	gradebook := newTestGradebook(t)

	if err := gradebook.AddStudent(studentID, "Другое имя"); err == nil {
		t.Error("student with duplicate ID was accepted")
	}
	if err := gradebook.AddCriterion("labs", "Другое название", 15, 50); err == nil {
		t.Error("criterion with duplicate ID was accepted")
	}
	if err := gradebook.AddCriterion("exam", "Экзамен", 0, 10); err == nil {
		t.Error("criterion with a zero maximum was accepted")
	}
	if err := gradebook.AddCriterion("exam", "Экзамен", 10, 0); err == nil {
		t.Error("criterion with a zero weight was accepted")
	}
}
