package files

import (
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"strconv"

	"hw2/internal/application"
)

type CSVExporter struct{}

func (CSVExporter) Export(path string, report application.FinalReport) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if errors.Is(err, os.ErrExist) {
		return fmt.Errorf("файл %q уже существует, выберите другое имя", path)
	}
	if err != nil {
		return fmt.Errorf("создать CSV: %w", err)
	}

	rows := [][]string{
		{"student_id", "student_name", "result_percent"},
		{report.Student.ID, report.Student.Name, strconv.FormatFloat(report.Result.Percent, 'f', 2, 64)},
	}
	writeErr := csv.NewWriter(file).WriteAll(rows)
	closeErr := file.Close()
	if err := errors.Join(writeErr, closeErr); err != nil {
		os.Remove(path)
		return fmt.Errorf("записать CSV: %w", err)
	}
	return nil
}
