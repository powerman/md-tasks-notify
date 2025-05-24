// Package main provides a tool for filtering and displaying Markdown tasks based on their status and dates.
package main

import (
	"bytes"
	"log"
	"os"
	"text/template"
	"time"

	obsidian "github.com/powerman/goldmark-obsidian"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
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

// TODO: Поддержка флагов -email, -from-day, -to-day.
// TODO: Чтение файлов из os.Args (либо stdin, если файлов нет).
// TODO: Вывод с указанием из какого файла эти задачи (если есть хоть одна из этого файла).
// TODO: Отправка вывода на email либо stdout (если -email не задан).
func main() {
	source := getExampleMarkdown()

	md := goldmark.New(
		goldmark.WithExtensions(
			obsidian.NewPlugTasks(),
			obsidian.NewObsidian(),
		),
		goldmark.WithRendererOptions(renderer.WithNodeRenderers(
			// Prio <500 needed to overwrite extension.GFM rendering to HTML.
			util.Prioritized(NewActualTasksRenderer(1, 1), 0),
		)),
	)
	err := md.Convert(source, os.Stdout)
	if err != nil {
		log.Fatal(err)
	}
}
