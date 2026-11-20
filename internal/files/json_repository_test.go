package files_test

import (
	"os"
	"path/filepath"
	"testing"

	"hw2/internal/domain"
	"hw2/internal/files"
)

const bookWithLab = `{
  "students": [{"id":"s1","name":"Анна"}],
  "criteria": [{"id":"labs","name":"Лабораторные","max_score":10,"weight":60}],
  "evaluations": [`

func TestInvalidJSONPreventsUpdate(t *testing.T) {
	cases := []struct {
		name    string
		content string
	}{
		{"truncated file", bookWithLab},
		{"missing points", bookWithLab + `{"student_id":"s1","criterion_id":"labs"}]}`},
		{"out of range", bookWithLab + `{"student_id":"s1","criterion_id":"labs","points":12}]}`},
		{"unknown criterion", bookWithLab + `{"student_id":"s1","criterion_id":"exam","points":8}]}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "gradebook.json")
			if err := os.WriteFile(path, []byte(tc.content), 0600); err != nil {
				t.Fatal(err)
			}
			repo := files.NewJSONRepository(path)
			err := repo.Update(func(book *domain.Gradebook) error {
				t.Error("update started with an invalid gradebook")
				return nil
			})
			if err == nil {
				t.Fatal("invalid gradebook accepted")
			}
			content, err := os.ReadFile(path)
			if err != nil || string(content) != tc.content {
				t.Fatalf("original file changed or unreadable: %q, %v", content, err)
			}
		})
	}
}

func TestRejectedCorrectionPreservesSavedGradebook(t *testing.T) {
	path := filepath.Join(t.TempDir(), "gradebook.json")
	original := bookWithLab + `{"student_id":"s1","criterion_id":"labs","points":8}]}`
	if err := os.WriteFile(path, []byte(original), 0600); err != nil {
		t.Fatal(err)
	}
	repo := files.NewJSONRepository(path)
	err := repo.Update(func(book *domain.Gradebook) error {
		return book.CorrectEvaluation("s1", "labs", 12)
	})
	if err == nil {
		t.Fatal("out-of-range correction accepted")
	}
	content, err := os.ReadFile(path)
	if err != nil || string(content) != original {
		t.Fatalf("original file changed or unreadable: %q, %v", content, err)
	}
	err = repo.Update(func(book *domain.Gradebook) error {
		return book.CorrectEvaluation("s1", "labs", 0)
	})
	if err != nil {
		t.Fatal(err)
	}
	book, err := files.NewJSONRepository(path).Load()
	if err != nil {
		t.Fatal(err)
	}
	result, err := domain.CalculateResult(book, "s1", domain.FinalResult{})
	if err != nil || result.Percent != 0 || result.EvaluatedCriteria != 1 {
		t.Fatalf("reloaded zero grade: result=%+v, err=%v", result, err)
	}
}
