package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"text/template"
	"time"

	"go.uber.org/mock/gomock"
)

// getExampleMarkdown returns example Markdown content with dates relative to current date using Go templates.
func getExampleMarkdown() []byte {
	type markdownData struct {
		Yesterday string
		Today     string
		Tomorrow  string
	}

	const markdownTmpl = `---
a: 1
b: x
  - b 2
  - [ ] b 3
---

# Test

- Item _cool_ 1
- [ ] Due yesterday 📅 {{.Yesterday}}
- [ ] Due today     📅 {{.Today}}
- [ ] Due tomorrow  📅 {{.Tomorrow}}
- [ ] Scheduled yesterday ⏳ {{.Yesterday}}
- [ ] Scheduled today     ⏳ {{.Today}}
- [ ] Scheduled tomorrow  ⏳ {{.Tomorrow}}
- [ ] Start yesterday 🛫 {{.Yesterday}}
- [ ] Start today     🛫 {{.Today}}
- [ ] Start tomorrow  🛫 {{.Tomorrow}}
- [ ] Due tomorrow Start yesterday 🛫 {{.Yesterday}} 📅 {{.Tomorrow}}
- [ ] Due tomorrow Start today     🛫 {{.Today}} 📅 {{.Tomorrow}}
- [ ] Due tomorrow Start tomorrow  🛫 {{.Tomorrow}} 📅 {{.Tomorrow}}
- [ ] Scheduled tomorrow Start yesterday 🛫 {{.Yesterday}} ⏳ {{.Tomorrow}}
- [ ] Scheduled tomorrow Start today     🛫 {{.Today}} ⏳ {{.Tomorrow}}
- [ ] Scheduled tomorrow Start tomorrow  🛫 {{.Tomorrow}} ⏳ {{.Tomorrow}}
- [ ] Recurring 🔁 every month ⏳ {{.Today}} 📅 {{.Tomorrow}}
- [X] Task
  - [ ] Subtask
- [x] Large _cool_ real task 🆔 jps5k3 #tag ⛔ peg74d,gg3xkn ⏬ 🔁 every day ➕ 2024-10-15 🛫 2024-10-15 ⏳ 2024-10-15 📅 2024-10-15 ❌ 2024-10-15 ✅ 2024-10-15 ^some-id
- [ ] Task _2 📅2024-10-19 📅2024-10-15
  🔼⏬_ #tag ^e5bebf

  Second paragraph ⏬
  - [ ] Second line.`

	today := time.Now()
	data := markdownData{
		Yesterday: today.AddDate(0, 0, -1).Format(time.DateOnly),
		Today:     today.Format(time.DateOnly),
		Tomorrow:  today.AddDate(0, 0, 1).Format(time.DateOnly),
	}

	tmpl, err := template.New("markdown").Parse(markdownTmpl)
	if err != nil {
		panic(err)
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		panic(err)
	}

	return buf.Bytes()
}

func TestRun(t *testing.T) {
	tests := []struct {
		name     string
		fromDate int
		toDate   int
		input    []byte
		wantErr  bool
		contains []string // Strings that should be present in output
		excludes []string // Strings that should not be present in output
	}{
		{
			name:     "Today and tomorrow",
			fromDate: 0,
			toDate:   1,
			input:    getExampleMarkdown(),
			contains: []string{
				"Due today",                          // Should contain tasks due today
				"Due tomorrow",                       // Should contain tasks due tomorrow
				"Scheduled tomorrow",                 // Should contain tasks scheduled for tomorrow
				"Due tomorrow Start yesterday",       // Should contain tasks starting before today with due date tomorrow
				"Due tomorrow Start today",           // Should contain tasks starting today with due date tomorrow
				"Scheduled tomorrow Start yesterday", // Should contain tasks starting before today with scheduled date tomorrow
				"Scheduled tomorrow Start today",     // Should contain tasks starting today with scheduled date tomorrow
				"Recurring",
			},
			excludes: []string{
				"---",                               // Should not contain frontmatter
				"# Test",                            // Should not contain headers
				"Item _cool_ 1",                     // Should not contain non-task items
				"Due yesterday",                     // Should not contain past tasks
				"[X] Task",                          // Should not contain completed tasks
				"[x] Large _cool_ real task",        // Should not contain completed tasks (lowercase x)
				"Due tomorrow Start tomorrow",       // Should not contain tasks starting tomorrow
				"Scheduled tomorrow Start tomorrow", // Should not contain tasks starting tomorrow
				"Task _2",                           // Should not contain tasks with dates far in future
			},
			wantErr: false,
		},
		{
			name:     "Only today",
			fromDate: 0,
			toDate:   0,
			input:    getExampleMarkdown(),
			contains: []string{
				"Due today",       // Should contain tasks due today
				"Scheduled today", // Should contain tasks scheduled for today
				"Recurring",
			},
			excludes: []string{
				"Due tomorrow",  // Should not contain future tasks
				"Due yesterday", // Should not contain past tasks
				"[X] Task",      // Should not contain completed tasks
			},
			wantErr: false,
		},
		{
			name:     "Yesterday to tomorrow",
			fromDate: -1,
			toDate:   1,
			input:    getExampleMarkdown(),
			contains: []string{
				"Due yesterday",   // Should contain tasks due yesterday
				"Due today",       // Should contain tasks due today
				"Due tomorrow",    // Should contain tasks due tomorrow
				"Scheduled today", // Should contain tasks scheduled for today
				"Recurring",
			},
			excludes: []string{
				"[X] Task", // Should not contain completed tasks
				"Task _2",  // Should not contain tasks with dates far in future
			},
			wantErr: false,
		},
		{
			name:     "Empty input",
			fromDate: 0,
			toDate:   1,
			input:    []byte{},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := filterActualTasks(tt.fromDate, tt.toDate, tt.input, &buf)
			if (err != nil) != tt.wantErr {
				t.Errorf("run() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			output := buf.String()
			for _, want := range tt.contains {
				if !strings.Contains(output, want) {
					t.Errorf("run() output should contain %q but doesn't\nOutput:\n%s", want, output)
				}
			}
			for _, exclude := range tt.excludes {
				if strings.Contains(output, exclude) {
					t.Errorf("run() output should not contain %q but does\nOutput:\n%s", exclude, output)
				}
			}
		})
	}
}

func TestFilterMarkdownFiles(t *testing.T) {
	tests := []struct {
		name    string
		files   map[string][]byte
		fromDay int
		toDay   int
		want    map[string][]byte
		wantErr bool
	}{
		{
			name: "Single file with tasks",
			files: map[string][]byte{
				"test.md": getExampleMarkdown(),
			},
			fromDay: 0,
			toDay:   1,
			wantErr: false,
		},
		{
			name: "Multiple files",
			files: map[string][]byte{
				"test1.md": getExampleMarkdown(),
				"test2.md": []byte("- [ ] Due today 📅 " + time.Now().Format(time.DateOnly)),
				"empty.md": []byte("# No tasks here"),
			},
			fromDay: 0,
			toDay:   1,
			wantErr: false,
		},
		{
			name:    "Empty input",
			files:   make(map[string][]byte),
			fromDay: 0,
			toDay:   1,
			wantErr: false,
			want:    make(map[string][]byte),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := filterMarkdownFiles(tt.files, &tt.fromDay, &tt.toDay)
			if (err != nil) != tt.wantErr {
				t.Errorf("filterMarkdownFiles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.want != nil && len(got) != len(tt.want) {
				t.Errorf("filterMarkdownFiles() got = %v tasks, want %v tasks", len(got), len(tt.want))
			}

			// Check that non-empty files contain some content
			if len(tt.files) > 0 && len(got) == 0 {
				t.Error("filterMarkdownFiles() returned empty result for non-empty input")
			}
		})
	}
}

func TestFormatTasks(t *testing.T) {
	tests := []struct {
		name     string
		tasks    map[string][]byte
		contains []string
		excludes []string
	}{
		{
			name: "Single file",
			tasks: map[string][]byte{
				"test.md": []byte("- [ ] Task 1\n- [ ] Task 2"),
			},
			contains: []string{
				"test.md:",
				"Task 1",
				"Task 2",
			},
		},
		{
			name: "Multiple files",
			tasks: map[string][]byte{
				"test1.md": []byte("- [ ] Task 1"),
				"test2.md": []byte("- [ ] Task 2"),
			},
			contains: []string{
				"test1.md:",
				"test2.md:",
				"Task 1",
				"Task 2",
			},
		},
		{
			name:  "Empty input",
			tasks: make(map[string][]byte),
		},
		{
			name: "Empty file content",
			tasks: map[string][]byte{
				"test.md": []byte(""),
			},
			contains: []string{"test.md:"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := formatTasks(tt.tasks)
			output := buf.String()

			for _, want := range tt.contains {
				if !strings.Contains(output, want) {
					t.Errorf("formatTasks() output should contain %q but doesn't\nOutput:\n%s", want, output)
				}
			}
			for _, exclude := range tt.excludes {
				if strings.Contains(output, exclude) {
					t.Errorf("formatTasks() output should not contain %q but does\nOutput:\n%s", exclude, output)
				}
			}
		})
	}
}

func TestNoEmailWhenEmpty(t *testing.T) {
	// Setup mock
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSMTP := NewMockSMTPSender(ctrl)

	// Create temporary file without tasks
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "no_tasks.md")
	err := os.WriteFile(tempFile, []byte("# Just a header\n\nSome text"), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	// Test parameters
	fromDay := 0
	toDay := 1
	emailTo := "test@example.com"

	// Create config with mock
	emailCfg := &EmailConfig{
		Host:     "localhost",
		Port:     25,
		From:     "test@example.com",
		SendMail: mockSMTP.SendMail,
	}

	// No email should be sent
	mockSMTP.EXPECT().SendMail(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

	// Run the test
	var stdout bytes.Buffer
	err = run(&fromDay, &toDay, &emailTo, emailCfg, &stdout, []string{tempFile})
	if err != nil {
		t.Errorf("run() unexpected error = %v", err)
	}
	// No content should be sent
	if stdout.Len() > 0 {
		t.Errorf("run() should not produce output for empty tasks, got %v", stdout.String())
	}
}

func TestEmailWithTasks(t *testing.T) {
	// Setup mock
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSMTP := NewMockSMTPSender(ctrl)

	// Create test file with known task
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "tasks.md")
	taskDate := time.Now().Format(time.DateOnly)
	taskContent := "- [ ] Test task 📅 " + taskDate + "\n"
	err := os.WriteFile(tempFile, []byte(taskContent), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	// Test parameters
	fromDay := 0
	toDay := 1
	emailTo := "test@example.com"

	// Create config with mock
	emailCfg := &EmailConfig{
		Host:     "localhost",
		Port:     25,
		From:     "test@example.com",
		SendMail: mockSMTP.SendMail,
	}

	// Email should be sent once
	mockSMTP.EXPECT().SendMail(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(nil)

	// Run the test
	var stdout bytes.Buffer
	err = run(&fromDay, &toDay, &emailTo, emailCfg, &stdout, []string{tempFile})
	if err != nil {
		t.Errorf("run() unexpected error = %v", err)
	}
	// Email should be sent, but no stdout content
	if stdout.Len() > 0 {
		t.Errorf("run() should not write to stdout when using email, got: %s", stdout.String())
	}
}

func TestRun_Integration(t *testing.T) {
	tests := []struct {
		name      string
		fromDay   int
		toDay     int
		emailTo   string
		paths     []string
		wantErr   bool
		wantPanic bool
	}{
		{
			name:    "No paths provided",
			fromDay: 0,
			toDay:   1,
			paths:   []string{},
		},
		{
			name:    "Invalid path",
			fromDay: 0,
			toDay:   1,
			paths:   []string{"nonexistent.md"},
			wantErr: true,
		},
		{
			name:      "Invalid date range",
			fromDay:   2,
			toDay:     1,
			paths:     []string{},
			wantPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout bytes.Buffer
			if tt.wantPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Error("Expected panic but got none")
					}
				}()
			}
			err := run(&tt.fromDay, &tt.toDay, &tt.emailTo, nil, &stdout, tt.paths)
			if !tt.wantPanic {
				if (err != nil) != tt.wantErr {
					t.Errorf("run() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
		})
	}
}
