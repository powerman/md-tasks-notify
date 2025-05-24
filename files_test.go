package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestIsMarkdownFile(t *testing.T) {
	tests := []struct {
		name string
		path string
		want bool
	}{
		{"Lowercase md", "file.md", true},
		{"Uppercase MD", "file.MD", true},
		{"Mixed case Md", "file.Md", true},
		{"Not markdown", "file.txt", false},
		{"No extension", "file", false},
		{"Directory with md", "dir.md/file", false},
		{"Path with spaces", "my file.md", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isMarkdownFile(tt.path); got != tt.want {
				t.Errorf("isMarkdownFile(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestReadFiles(t *testing.T) {
	// Create a temporary directory for test files.
	tmpDir, err := os.MkdirTemp("", "md-tasks-notify-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test files and directories.
	files := map[string]string{
		"test1.md":         "# Test 1\n- [ ] Task 1",
		"test2.md":         "# Test 2\n- [ ] Task 2",
		"subdir/test3.md":  "# Test 3\n- [ ] Task 3",
		"test4.txt":        "Not a markdown file",
		"subdir/test5.txt": "Not a markdown file",
		"empty.md":         "",
		"subdir2/empty.md": "",
	}

	for path, content := range files {
		fullPath := filepath.Join(tmpDir, filepath.FromSlash(path))
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	tests := []struct {
		name        string
		paths       []string
		wantErr     bool
		wantContent []byte
	}{
		{
			name:        "Single file",
			paths:       []string{filepath.Join(tmpDir, "test1.md")},
			wantContent: []byte("# Test 1\n- [ ] Task 1"),
		},
		{
			name:        "Multiple files",
			paths:       []string{filepath.Join(tmpDir, "test1.md"), filepath.Join(tmpDir, "test2.md")},
			wantContent: []byte("# Test 1\n- [ ] Task 1\n# Test 2\n- [ ] Task 2"),
		},
		{
			name:        "Directory",
			paths:       []string{filepath.Join(tmpDir, "subdir")},
			wantContent: []byte("# Test 3\n- [ ] Task 3"),
		},
		{
			name:        "Mixed files and directories",
			paths:       []string{filepath.Join(tmpDir, "test1.md"), filepath.Join(tmpDir, "subdir")},
			wantContent: []byte("# Test 1\n- [ ] Task 1\n# Test 3\n- [ ] Task 3"),
		},
		{
			name:    "Non-existent file",
			paths:   []string{filepath.Join(tmpDir, "nonexistent.md")},
			wantErr: true,
		},
		{
			name:        "Empty markdown file",
			paths:       []string{filepath.Join(tmpDir, "empty.md")},
			wantContent: []byte(""),
		},
		{
			name:        "Empty directory",
			paths:       []string{filepath.Join(tmpDir, "subdir2")},
			wantContent: nil,
		},
		{
			name:        "Non-markdown file",
			paths:       []string{filepath.Join(tmpDir, "test4.txt")},
			wantContent: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := readFiles(tt.paths)
			if (err != nil) != tt.wantErr {
				t.Errorf("readFiles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			if !bytes.Equal(got, tt.wantContent) {
				t.Errorf("readFiles() = %q, want %q", got, tt.wantContent)
			}
		})
	}
}
