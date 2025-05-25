package main

import (
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

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

// readMarkdownFilesFromFS reads markdown files from a filesystem.
func readMarkdownFilesFromFS(fsys fs.FS, paths []string) (map[string][]byte, error) {
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
		return make(map[string][]byte), nil
	}

	result := make(map[string][]byte)
	for _, filename := range mdFiles {
		fileData, err := fs.ReadFile(fsys, filename)
		if err != nil {
			return nil, err
		}
		result[filename] = fileData
	}

	return result, nil
}

// readMarkdownFiles reads markdown files from paths.
func readMarkdownFiles(paths []string) (map[string][]byte, error) {
	relPaths := make([]string, len(paths))
	for i, path := range paths {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return nil, err
		}
		// Remove leading "/" to make path relative to root for fs.FS.
		relPaths[i] = strings.TrimPrefix(absPath, "/")
	}
	filesData, err := readMarkdownFilesFromFS(os.DirFS("/"), relPaths)
	if err != nil {
		return nil, err
	}

	// Convert relative paths back to absolute
	result := make(map[string][]byte, len(filesData))
	for filename, data := range filesData {
		result["/"+filename] = data
	}
	return result, nil
}

// readMarkdownFilesOrStdin reads markdown files from paths or stdin.
func readMarkdownFilesOrStdin(paths []string) (map[string][]byte, error) {
	if len(paths) > 0 {
		return readMarkdownFiles(paths)
	}

	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return nil, fmt.Errorf("read stdin: %w", err)
	}
	return map[string][]byte{"": data}, nil
}
