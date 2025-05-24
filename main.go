// Package main provides a tool for filtering and displaying Markdown tasks based on their status and dates.
package main

import (
	"io"
	"log"
	"os"

	obsidian "github.com/powerman/goldmark-obsidian"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

// TODO: Поддержка флагов -email, -from-day, -to-day.
// TODO: Чтение файлов из os.Args (либо stdin, если файлов нет).
// TODO: Вывод с указанием из какого файла эти задачи (если есть хоть одна из этого файла).
// TODO: Отправка вывода на email либо stdout (если -email не задан).
func main() {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}
	err = run(0, 1, data, os.Stdout)
	if err != nil {
		log.Fatal(err)
	}
}

func run(dayFrom int, dayTo int, source []byte, w io.Writer) error {
	md := goldmark.New(
		goldmark.WithExtensions(
			obsidian.NewPlugTasks(),
			obsidian.NewObsidian(),
		),
		goldmark.WithRendererOptions(renderer.WithNodeRenderers(
			// Prio <500 needed to overwrite extension.GFM rendering to HTML.
			util.Prioritized(NewActualTasksRenderer(dayFrom, dayTo), 0),
		)),
	)
	err := md.Convert(source, w)
	return err
}
