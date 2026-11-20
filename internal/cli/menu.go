package cli

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type menuItem struct {
	command string
	title   string
	fields  []string
}

var menuItems = []menuItem{
	{"students", "Список студентов", nil},
	{"add-student", "Добавить студента", []string{"Новый ID студента", "Имя студента"}},
	{"criteria", "Посмотреть рубрикатор", nil},
	{"add-criterion", "Добавить критерий", []string{"Новый ID критерия", "Название критерия", "Максимальный балл", "Вес критерия"}},
	{"evaluations", "Посмотреть оценки студента", []string{"ID студента"}},
	{"add-evaluation", "Выставить оценку", []string{"ID студента", "ID критерия", "Балл"}},
	{"correct-evaluation", "Исправить оценку", []string{"ID студента", "ID критерия", "Новый балл"}},
	{"cancel-evaluation", "Отменить оценку", []string{"ID студента", "ID критерия"}},
	{"result", "Рассчитать текущий итог", []string{"ID студента"}},
	{"export", "Сохранить окончательный итог в CSV", []string{"ID студента", "Путь к новому CSV-файлу"}},
}

func (c *Console) Menu(input io.Reader) error {
	scanner := bufio.NewScanner(input)
	var menu strings.Builder
	fmt.Fprintln(&menu, "ПОС — ведомость курса")
	for i, item := range menuItems {
		fmt.Fprintf(&menu, "%2d. %s\n", i+1, item.title)
	}
	fmt.Fprintln(&menu, " 0. Выход")
	for {
		if _, err := fmt.Fprint(c.output, menu.String(), "Выберите номер: "); err != nil {
			return err
		}
		if !scanner.Scan() {
			return scanner.Err()
		}
		number, err := strconv.Atoi(strings.TrimSpace(scanner.Text()))
		if err != nil || number < 0 || number > len(menuItems) {
			if _, err := fmt.Fprintf(c.output, "Введите номер от 0 до %d.\n", len(menuItems)); err != nil {
				return err
			}
			continue
		}
		if number == 0 {
			return nil
		}
		item := menuItems[number-1]
		args := []string{item.command}
		for _, field := range item.fields {
			if _, err := fmt.Fprintf(c.output, "%s: ", field); err != nil {
				return err
			}
			if !scanner.Scan() {
				return scanner.Err()
			}
			args = append(args, strings.TrimSpace(scanner.Text()))
		}
		if err := c.Run(args); err != nil {
			if _, writeErr := fmt.Fprintln(c.output, "Ошибка:", err); writeErr != nil {
				return writeErr
			}
		}
	}
}
