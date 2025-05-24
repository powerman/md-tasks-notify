// Package main provides a tool for filtering and displaying Markdown tasks based on their status and dates.
//
// TODO: Вывод с указанием из какого файла эти задачи (если есть хоть одна из этого файла).
package main

import (
	"bytes"
	"flag"
	"io"
	"log"
	"os"

	obsidian "github.com/powerman/goldmark-obsidian"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

const emailSubject = "Actual tasks"

func main() {
	fromDay := flag.Int("from-day", 0, "Start day relative to today (-1 for yesterday, 0 for today)")
	toDay := flag.Int("to-day", 1, "End day relative to today (1 for tomorrow)")
	emailTo := flag.String("email", "", "Send output to this email address instead of stdout")
	flag.Parse()

	var data []byte
	var err error
	if flag.NArg() > 0 {
		data, err = ReadFiles(flag.Args())
	} else {
		data, err = io.ReadAll(os.Stdin)
	}
	if err != nil {
		log.Fatal(err)
	}

	var buf bytes.Buffer
	if err := run(*fromDay, *toDay, data, &buf); err != nil {
		log.Fatal(err)
	}

	if *emailTo == "" {
		_, err = io.Copy(os.Stdout, &buf)
	} else {
		err = NewEmail(nil).Send(*emailTo, emailSubject, &buf)
	}
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
