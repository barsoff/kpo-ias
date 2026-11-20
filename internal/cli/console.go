package cli

import (
	"errors"
	"fmt"
	"io"
	"strconv"

	"hw2/internal/application"
	"hw2/internal/domain"
)

const Usage = `ПОС — ведомость одного курса
Меню на русском: go run .
Отдельная команда: go run . [-data JSON_PATH] COMMAND

Команды:
  add-student STUDENT_ID "NAME"
  students
  add-criterion CRITERION_ID "TITLE" MAX_POINTS WEIGHT
  criteria
  add-evaluation STUDENT_ID CRITERION_ID POINTS
  evaluations STUDENT_ID
  correct-evaluation STUDENT_ID CRITERION_ID POINTS
  cancel-evaluation STUDENT_ID CRITERION_ID
  result STUDENT_ID
  export STUDENT_ID CSV_PATH

Пример: go run . add-evaluation s1 labs 8`

type Console struct {
	facade *application.Facade
	output io.Writer
}

func New(facade *application.Facade, output io.Writer) *Console {
	return &Console{facade: facade, output: output}
}

func (c *Console) Run(args []string) error {
	if len(args) == 0 || args[0] == "help" {
		_, err := fmt.Fprintln(c.output, Usage)
		return err
	}
	command, values := args[0], args[1:]
	var selected *menuItem
	for i := range menuItems {
		if menuItems[i].command == command {
			selected = &menuItems[i]
			break
		}
	}
	if selected == nil {
		return fmt.Errorf("неизвестная команда %q. Список команд: help", command)
	}
	if len(values) != len(selected.fields) {
		return fmt.Errorf("команда %q ожидает аргументов: %d. Список команд: help", command, len(selected.fields))
	}

	var err error
	switch command {
	case "add-student":
		err = c.facade.AddStudent(values[0], values[1])
	case "students":
		return c.printStudents()
	case "add-criterion":
		maxScore, parseErr := strconv.Atoi(values[2])
		if parseErr != nil {
			return fmt.Errorf("максимальный балл должен быть целым числом")
		}
		weight, parseErr := strconv.Atoi(values[3])
		if parseErr != nil {
			return fmt.Errorf("вес должен быть целым числом")
		}
		err = c.facade.AddCriterion(values[0], values[1], maxScore, weight)
	case "criteria":
		return c.printCriteria()
	case "add-evaluation", "correct-evaluation":
		points, parseErr := strconv.Atoi(values[2])
		if parseErr != nil {
			return fmt.Errorf("балл должен быть целым числом")
		}
		if command == "add-evaluation" {
			err = c.facade.AddEvaluation(values[0], values[1], points)
		} else {
			err = c.facade.CorrectEvaluation(values[0], values[1], points)
		}
	case "cancel-evaluation":
		err = c.facade.CancelEvaluation(values[0], values[1])
	case "evaluations":
		return c.printEvaluations(values[0])
	case "result":
		return c.printResult(values[0])
	case "export":
		if err := c.facade.ExportFinalResult(values[0], values[1]); err != nil {
			return err
		}
		_, err := fmt.Fprintf(c.output, "Итог экспортирован в %s\n", values[1])
		return err
	}
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(c.output, "Готово.")
	return err
}

func (c *Console) printStudents() error {
	students, err := c.facade.Students()
	if err != nil {
		return err
	}
	if len(students) == 0 {
		_, err := fmt.Fprintln(c.output, "Студентов пока нет.")
		return err
	}
	for _, student := range students {
		if _, err := fmt.Fprintf(c.output, "%s — %s\n", student.ID, student.Name); err != nil {
			return err
		}
	}
	return nil
}

func (c *Console) printCriteria() error {
	criteria, err := c.facade.Criteria()
	if err != nil {
		return err
	}
	if len(criteria) == 0 {
		_, err := fmt.Fprintln(c.output, "Критериев пока нет.")
		return err
	}
	for _, criterion := range criteria {
		if _, err := fmt.Fprintf(c.output, "%s — %s (максимум %d, вес %d)\n",
			criterion.ID, criterion.Name, criterion.MaxScore, criterion.Weight); err != nil {
			return err
		}
	}
	return nil
}

func (c *Console) printEvaluations(studentID string) error {
	evaluations, err := c.facade.Evaluations(studentID)
	if err != nil {
		return err
	}
	if len(evaluations) == 0 {
		_, err := fmt.Fprintln(c.output, "У студента пока нет оценок.")
		return err
	}
	for _, evaluation := range evaluations {
		if _, err := fmt.Fprintf(c.output, "%s — %d\n", evaluation.CriterionID, evaluation.Score.Points()); err != nil {
			return err
		}
	}
	return nil
}

func (c *Console) printResult(studentID string) error {
	result, err := c.facade.CurrentResult(studentID)
	if errors.Is(err, domain.ErrEmptyRubric) || errors.Is(err, domain.ErrNoEvaluations) {
		_, writeErr := fmt.Fprintf(c.output, "Итог недоступен (оценено %d/%d): %s\n",
			result.EvaluatedCriteria, result.TotalCriteria, err)
		return writeErr
	}
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(c.output, "Текущий итог: %.2f/100 (оценено %d/%d)\n",
		result.Percent, result.EvaluatedCriteria, result.TotalCriteria)
	return err
}
