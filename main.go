// Package main provides a tool for filtering and displaying Markdown tasks based on their status and dates.
package main

import (
	"flag"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	obsidian "github.com/powerman/goldmark-obsidian"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

// TODO: Поддержка флагов -email.
// TODO: Вывод с указанием из какого файла эти задачи (если есть хоть одна из этого файла).
// TODO: Отправка вывода на email либо stdout (если -email не задан).

const (
	// estimatedFileSize is a rough estimate of average file size in bytes for pre-allocation.
	estimatedFileSize = 1024 // 1 KB
)

// isMarkdownFile checks if the given path has a .md extension.
func isMarkdownFile(path string) bool {
	return strings.HasSuffix(strings.ToLower(path), ".md")
}

// findMarkdownFiles collects all .md files from a directory recursively.
func findMarkdownFiles(dir string) ([]string, error) {
	var files []string
	err := filepath.Walk(dir, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !fi.IsDir() && isMarkdownFile(fi.Name()) {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

// readFiles reads and concatenates contents of all provided files and directories.
// For directories, it recursively reads all files with .md extension.
func readFiles(paths []string) ([]byte, error) {
	var mdFiles []string
	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			return nil, err
		}

		if info.IsDir() {
			files, err := findMarkdownFiles(path)
			if err != nil {
				return nil, err
			}
			mdFiles = append(mdFiles, files...)
		} else if isMarkdownFile(path) {
			mdFiles = append(mdFiles, path)
		}
	}

	data := make([]byte, 0, len(mdFiles)*estimatedFileSize) // Pre-allocate with rough estimate per file
	for _, filename := range mdFiles {
		//nolint:gosec // Reading files provided as command-line arguments is the intended behavior.
		fileData, err := os.ReadFile(filename)
		if err != nil {
			return nil, err
		}
		data = append(data, fileData...)
		// Add newline between files to ensure proper parsing
		data = append(data, '\n')
	}
	return data, nil
}

func main() {
	fromDay := flag.Int("from-day", 0, "Start day relative to today (-1 for yesterday, 0 for today)")
	toDay := flag.Int("to-day", 1, "End day relative to today (1 for tomorrow)")
	flag.Parse()

	var data []byte
	var err error
	if flag.NArg() > 0 {
		data, err = readFiles(flag.Args())
	} else {
		data, err = io.ReadAll(os.Stdin)
	}
	if err != nil {
		log.Fatal(err)
	}

	err = run(*fromDay, *toDay, data, os.Stdout)
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
