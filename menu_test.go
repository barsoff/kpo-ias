package main

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestMenuAndCommandsShareGradebook(t *testing.T) {
	args := []string{"-data", filepath.Join(t.TempDir(), "book.json")}
	var output bytes.Buffer
	input := strings.NewReader("99\n2\ns1\nАнна Иванова\n2\ns1\nДубликат\n0\n")
	if err := run(args, input, &output); err != nil {
		t.Fatal(err)
	}
	for _, message := range []string{"Введите номер", "Имя студента", "уже существует"} {
		if !strings.Contains(output.String(), message) {
			t.Fatalf("missing message %q", message)
		}
	}
	output.Reset()
	if err := run(args, strings.NewReader("2\ns2\n"), &output); err != nil {
		t.Fatal(err)
	}
	output.Reset()
	if err := run(append(args, "students"), strings.NewReader(""), &output); err != nil {
		t.Fatal(err)
	}
	if output.String() != "s1 — Анна Иванова\n" {
		t.Fatalf("unexpected saved students: %q", output.String())
	}
}
