// Package main provides a tool for filtering and displaying Markdown tasks based on their status and dates.
package main

import (
	"bytes"
	"flag"
	"fmt"
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
	log.SetFlags(0)

	fromDay := flag.Int("from-day", 0, "Start day relative to today (-1 for yesterday, 0 for today)")
	toDay := flag.Int("to-day", 1, "End day relative to today (1 for tomorrow)")
	emailTo := flag.String("email", "", "Send output to this email address instead of stdout")
	flag.Parse()
	if *fromDay > *toDay {
		log.Fatalln("Error: from-day must be less than or equal to to-day")
	}

	err := run(fromDay, toDay, emailTo, os.Stdout, flag.Args())
	if err != nil {
		log.Fatalln("Failed to", err)
	}
}

// run is testable part of main function.
func run(fromDay *int, toDay *int, emailTo *string, stdout io.Writer, paths []string) error {
	files, err := readMarkdownFilesOrStdin(paths)
	if err != nil {
		return err
	}

	tasks, err := filterMarkdownFiles(files, fromDay, toDay)
	if err != nil {
		return err
	}

	buf := formatTasks(tasks)

	if *emailTo == "" {
		_, err = io.Copy(stdout, &buf)
	} else {
		err = NewEmail(nil).Send(*emailTo, emailSubject, &buf)
	}
	return err
}

// filterMarkdownFiles processes each file and returns a map of filenames to their filtered task content.
func filterMarkdownFiles(files map[string][]byte, fromDay *int, toDay *int) (map[string][]byte, error) {
	tasks := make(map[string][]byte)
	for filename, data := range files {
		var buf bytes.Buffer
		if err := filterActualTasks(*fromDay, *toDay, data, &buf); err != nil {
			return nil, fmt.Errorf("filter tasks: %w", err)
		}
		if buf.Len() > 0 {
			tasks[filename] = buf.Bytes()
		}
	}
	return tasks, nil
}

// filterActualTasks filters the actual tasks from the markdown data.
func filterActualTasks(dayFrom int, dayTo int, markdownData []byte, filteredTasks io.Writer) error {
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
	err := md.Convert(markdownData, filteredTasks)
	return err
}

// formatTasks takes filtered tasks and formats them with filenames into a single buffer.
func formatTasks(tasks map[string][]byte) bytes.Buffer {
	var buf bytes.Buffer
	first := true
	for filename, taskData := range tasks {
		if !first {
			buf.WriteByte('\n')
		}
		if filename != "" {
			buf.WriteString(filename)
			buf.WriteString(":\n")
		}
		buf.Write(taskData)
		first = false
	}
	return buf
}
