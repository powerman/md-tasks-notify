package main

import (
	"bytes"
	"testing"
	"testing/fstest"
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
	// Create a virtual filesystem for testing.
	fsys := fstest.MapFS{
		"test1.md":         &fstest.MapFile{Data: []byte("# Test 1\n- [ ] Task 1")},
		"test2.md":         &fstest.MapFile{Data: []byte("# Test 2\n- [ ] Task 2")},
		"subdir/test3.md":  &fstest.MapFile{Data: []byte("# Test 3\n- [ ] Task 3")},
		"test4.txt":        &fstest.MapFile{Data: []byte("Not a markdown file")},
		"subdir/test5.txt": &fstest.MapFile{Data: []byte("Not a markdown file")},
		"empty.md":         &fstest.MapFile{Data: []byte("")},
		"subdir2/empty.md": &fstest.MapFile{Data: []byte("")},
	}

	tests := []struct {
		name    string
		paths   []string
		wantErr bool
		want    map[string][]byte
	}{
		{
			name:  "Single file",
			paths: []string{"test1.md"},
			want:  map[string][]byte{"test1.md": []byte("# Test 1\n- [ ] Task 1")},
		},
		{
			name:  "Multiple files",
			paths: []string{"test1.md", "test2.md"},
			want: map[string][]byte{
				"test1.md": []byte("# Test 1\n- [ ] Task 1"),
				"test2.md": []byte("# Test 2\n- [ ] Task 2"),
			},
		},
		{
			name:  "Directory",
			paths: []string{"subdir"},
			want: map[string][]byte{
				"subdir/test3.md": []byte("# Test 3\n- [ ] Task 3"),
			},
		},
		{
			name:  "Mixed files and directories",
			paths: []string{"test1.md", "subdir"},
			want: map[string][]byte{
				"test1.md":        []byte("# Test 1\n- [ ] Task 1"),
				"subdir/test3.md": []byte("# Test 3\n- [ ] Task 3"),
			},
		},
		{
			name:    "Non-existent file",
			paths:   []string{"nonexistent.md"},
			wantErr: true,
		},
		{
			name:  "Empty markdown file",
			paths: []string{"empty.md"},
			want:  map[string][]byte{"empty.md": []byte("")},
		},
		{
			name:  "Empty directory",
			paths: []string{"subdir2"},
			want:  map[string][]byte{"subdir2/empty.md": []byte("")},
		},
		{
			name:  "Non-markdown file",
			paths: []string{"test4.txt"},
			want:  make(map[string][]byte),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := readMarkdownFilesFromFS(fsys, tt.paths)
			if (err != nil) != tt.wantErr {
				t.Errorf("readFiles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			if len(got) != len(tt.want) {
				t.Errorf("readFiles() returned %d files, want %d", len(got), len(tt.want))
				return
			}
			for k, v := range tt.want {
				if !bytes.Equal(got[k], v) {
					t.Errorf("readFiles()[%q] = %q, want %q", k, got[k], v)
				}
			}
		})
	}
}
