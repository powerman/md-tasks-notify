package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"
)

// isMarkdownFile checks if the given path has a .md extension.
func isMarkdownFile(path string) bool {
	return strings.EqualFold(filepath.Ext(path), ".md")
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
			if len(files) == 0 {
				log.Printf("Warning: Directory %q contains no Markdown files.", path)
			}
			mdFiles = append(mdFiles, files...)
		} else if isMarkdownFile(path) {
			mdFiles = append(mdFiles, path)
		} else {
			log.Printf("Warning: Skipping %q as it is not a Markdown file.", path)
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
