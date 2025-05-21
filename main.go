package main

import (
	"log"
	"os"

	obsidian "github.com/powerman/goldmark-obsidian"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

var source = []byte(`---
a: 1
b: x
  - b 2
  - [ ] b 3
---

# Test

- Item _cool_ 1
- [ ] Due yesterday 📅 2024-10-18
- [ ] Due today     📅 2024-10-19
- [ ] Due tomorrow  📅 2024-11-02
- [ ] Scheduled yesterday ⏳ 2024-10-18
- [ ] Scheduled today     ⏳ 2024-10-19
- [ ] Scheduled tomorrow  ⏳ 2024-10-20
- [ ] Start yesterday 🛫 2024-10-18
- [ ] Start today     🛫 2024-10-19
- [ ] Start tomorrow  🛫 2024-10-20
- [ ] Due tomorrow Start yesterday 🛫 2024-10-18 📅 2024-10-20
- [ ] Due tomorrow Start today     🛫 2024-10-19 📅 2024-10-20
- [ ] Due tomorrow Start tomorrow  🛫 2024-10-20 📅 2024-10-20
- [ ] Scheduled tomorrow Start yesterday 🛫 2024-10-18 ⏳ 2024-10-20
- [ ] Scheduled tomorrow Start today     🛫 2024-10-19 ⏳ 2024-10-20
- [ ] Scheduled tomorrow Start tomorrow  🛫 2024-10-20 ⏳ 2024-10-20
- [X] Task
  - [ ] Subtask
- [x] Large _cool_ real task 🆔 jps5k3 #tag ⛔ peg74d,gg3xkn ⏬ 🔁 every day ➕ 2024-10-15 🛫 2024-10-15 ⏳ 2024-10-15 📅 2024-10-15 ❌ 2024-10-15 ✅ 2024-10-15 ^some-id
- [ ] Task _2 📅2024-10-19 📅2024-10-15
  🔼⏬_ #tag ^e5bebf

  Second paragraph ⏬
  - [ ] Second line.
	`)

// TODO: Поддержка флагов -email, -from-day, -to-day.
// TODO: Чтение файлов из os.Args (либо stdin, если файлов нет).
// TODO: Вывод с указанием из какого файла эти задачи (если есть хоть одна из этого файла).
// TODO: Отправка вывода на email либо stdout (если -email не задан).
func main() {
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
