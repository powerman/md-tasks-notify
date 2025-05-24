package main

import (
	"bytes"
	"io/fs"
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
func findMarkdownFiles(fsys fs.FS, dir string) ([]string, error) {
	var files []string
	err := fs.WalkDir(fsys, dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && isMarkdownFile(d.Name()) {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

// readFiles reads and concatenates contents of all provided files and directories.
// For directories, it recursively reads all files with .md extension.
func readFiles(fsys fs.FS, paths []string) ([]byte, error) {
	var mdFiles []string
	for _, path := range paths {
		info, err := fs.Stat(fsys, path)
		if err != nil {
			return nil, err
		}

		switch {
		case info.IsDir():
			files, err := findMarkdownFiles(fsys, path)
			if err != nil {
				return nil, err
			}
			if len(files) == 0 {
				log.Printf("Warning: Directory %q contains no Markdown files.", path)
			}
			mdFiles = append(mdFiles, files...)
		case isMarkdownFile(path):
			mdFiles = append(mdFiles, path)
		default:
			log.Printf("Warning: Skipping %q as it is not a Markdown file.", path)
		}
	}

	if len(mdFiles) == 0 {
		return nil, nil
	}

	var allData [][]byte //nolint:prealloc // Premature optimization.
	for _, filename := range mdFiles {
		fileData, err := fs.ReadFile(fsys, filename)
		if err != nil {
			return nil, err
		}
		allData = append(allData, fileData)
	}

	return bytes.Join(allData, []byte{'\n'}), nil
}

// ReadFiles is a convenience wrapper around readFiles that uses the OS filesystem.
func ReadFiles(paths []string) ([]byte, error) {
	relPaths := make([]string, len(paths))
	for i, path := range paths {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return nil, err
		}
		// Remove leading "/" to make path relative to root for fs.FS.
		relPaths[i] = strings.TrimPrefix(absPath, "/")
	}
	return readFiles(os.DirFS("/"), relPaths)
}
