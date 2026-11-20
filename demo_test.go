package main

import (
	"encoding/csv"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestDemo(t *testing.T) {
	dir := t.TempDir()
	program := filepath.Join(dir, "pos.exe")
	dataPath := filepath.Join(dir, "gradebook.json")
	csvPath := filepath.Join(dir, "result.csv")

	build := exec.Command("go", "build", "-buildvcs=false", "-o", program, ".")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, output)
	}
	runCommand := func(args ...string) (string, error) {
		commandArgs := append([]string{"-data", dataPath}, args...)
		output, err := exec.Command(program, commandArgs...).CombinedOutput()
		t.Logf("%s\n%s", strings.Join(args, " "), output)
		return string(output), err
	}
	runOK := func(args ...string) string {
		t.Helper()
		output, err := runCommand(args...)
		if err != nil {
			t.Fatalf("command failed: %v", err)
		}
		return output
	}
	checkResult := func(percent, progress string) {
		t.Helper()
		output := runOK("result", "s1")
		if !strings.Contains(output, percent+"/100") || !strings.Contains(output, progress+")") {
			t.Fatalf("want result %s/100 and progress %s, got %q", percent, progress, output)
		}
	}

	runOK("add-student", "s1", "Анна Иванова")
	runOK("add-criterion", "labs", "Лабораторные", "10", "60")
	runOK("add-criterion", "project", "Проект", "20", "40")
	runOK("students")
	runOK("criteria")
	runOK("add-evaluation", "s1", "labs", "8")
	checkResult("80.00", "1/2")

	output, err := runCommand("export", "s1", csvPath)
	if err == nil || !strings.Contains(output, "нужны оценки по всем критериям") {
		t.Fatalf("expected incomplete rubric error, got %v: %s", err, output)
	}
	if _, err := os.Stat(csvPath); !os.IsNotExist(err) {
		t.Fatalf("the incomplete result must leave no CSV file, stat error: %v", err)
	}

	runOK("add-evaluation", "s1", "project", "15")
	runOK("evaluations", "s1")
	checkResult("78.00", "2/2")
	runOK("correct-evaluation", "s1", "project", "0")
	checkResult("48.00", "2/2")
	runOK("cancel-evaluation", "s1", "project")
	checkResult("80.00", "1/2")
	runOK("add-evaluation", "s1", "project", "15")
	t.Log("Чтение сохранённой ведомости в новом процессе")
	checkResult("78.00", "2/2")
	runOK("export", "s1", csvPath)

	content, err := os.ReadFile(csvPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Содержимое CSV:\n%s", content)
	rows, err := csv.NewReader(strings.NewReader(string(content))).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 || strings.Join(rows[0], ",") != "student_id,student_name,result_percent" ||
		strings.Join(rows[1], ",") != "s1,Анна Иванова,78.00" {
		t.Fatalf("unexpected CSV rows: %q", rows)
	}
}
