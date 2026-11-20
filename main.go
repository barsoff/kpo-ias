package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"go.uber.org/dig"

	"hw2/internal/application"
	"hw2/internal/cli"
	"hw2/internal/files"
)

func main() {
	if err := run(os.Args[1:], os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "Ошибка:", err)
		os.Exit(1)
	}
}

func run(args []string, input io.Reader, output io.Writer) error {
	flags := flag.NewFlagSet("hw2", flag.ContinueOnError)
	flags.SetOutput(output)
	path := flags.String("data", "gradebook.json", "путь к файлу ведомости")
	flags.Usage = func() {
		fmt.Fprintln(output, cli.Usage)
		flags.PrintDefaults()
	}
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}

	container := dig.New()
	constructors := []any{
		func() application.GradebookRepository { return files.NewJSONRepository(*path) },
		func() application.ResultExporter { return files.CSVExporter{} },
		func() io.Writer { return output },
		application.NewFacade,
		cli.New,
	}
	for _, constructor := range constructors {
		if err := container.Provide(constructor); err != nil {
			return fmt.Errorf("зарегистрировать зависимость: %w", err)
		}
	}
	return container.Invoke(func(console *cli.Console) error {
		if flags.NArg() == 0 {
			return console.Menu(input)
		}
		return console.Run(flags.Args())
	})
}
