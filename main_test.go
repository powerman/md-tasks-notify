package main

import (
	"bytes"
	"strings"
	"testing"
	"text/template"
	"time"
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
			err := run(tt.fromDate, tt.toDate, tt.input, &buf)
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
