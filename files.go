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
	result := make(map[string][]byte)
	for _, path := range paths {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return nil, err
		}
		info, err := os.Stat(absPath)
		if err != nil {
			return nil, err
		}

		// Normalize: for directories, append "." to make them look like files.
		targetPath := absPath
		if info.IsDir() {
			targetPath = filepath.Join(absPath, ".")
		}

		// Now we can handle both files and directories uniformly.
		dir := filepath.Dir(targetPath)
		baseName := filepath.Base(targetPath)
		filesData, err := readMarkdownFilesFromFS(os.DirFS(dir), []string{baseName})
		if err != nil {
			return nil, err
		}

		// Convert relative paths to absolute.
		for filename, data := range filesData {
			result[filepath.Join(dir, filename)] = data
		}
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
