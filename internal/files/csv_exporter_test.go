package files_test

import (
	"bytes"
	"encoding/csv"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"hw2/internal/application"
	"hw2/internal/domain"
	"hw2/internal/files"
)

func TestExportPreservesFirstReport(t *testing.T) {
	path := filepath.Join(t.TempDir(), "result.csv")
	report := application.FinalReport{
		Student: domain.Student{ID: "s1", Name: `Анна, группа "Б"`},
		Result:  domain.Result{Percent: 78, EvaluatedCriteria: 2, TotalCriteria: 2},
	}
	exporter := files.CSVExporter{}
	if err := exporter.Export(path, report); err != nil {
		t.Fatal(err)
	}
	first, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	rows, err := csv.NewReader(bytes.NewReader(first)).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 || !slices.Equal(rows[1], []string{"s1", report.Student.Name, "78.00"}) {
		t.Fatalf("unexpected CSV: %q", rows)
	}
	report.Result.Percent = 48
	if err := exporter.Export(path, report); err == nil {
		t.Fatal("second export to the same path succeeded")
	}
	after, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(first, after) {
		t.Fatalf("first report changed or unreadable: %q, %v", after, err)
	}
}
