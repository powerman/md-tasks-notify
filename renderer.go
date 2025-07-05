package main

import (
	"fmt"
	"time"

	mathjax "github.com/litao91/goldmark-mathjax"
	obsast "github.com/powerman/goldmark-obsidian/ast"
	"github.com/yuin/goldmark/ast"
	extast "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
	"go.abhg.dev/goldmark/hashtag"
	"go.abhg.dev/goldmark/mermaid"
	"go.abhg.dev/goldmark/wikilink"
)

// FilteredTasksRenderer implement renderer.NodeRenderer object.
type FilteredTasksRenderer struct {
	StatusType            map[obsast.PlugTasksStatusType]bool
	DueAfter              time.Time
	DueBefore             time.Time
	ScheduledAfter        time.Time
	ScheduledBefore       time.Time
	StartBefore           time.Time
	RequireDueOrScheduled bool
}

// NewActualTasksRenderer returns FilteredTasksRenderer configured to filter tasks:
//   - not done
//   - due or scheduled between dayFrom and dayTo (inclusive)
//   - without start date or start before today (inclusive)
//
// Value 0 for dayFrom and dayTo means today, 1 means tomorrow, -1 means yesterday, etc.
func NewActualTasksRenderer(dayFrom, dayTo int) renderer.NodeRenderer {
	if dayFrom > dayTo {
		panic(fmt.Sprintf("dayFrom %d must be <= dayTo %d", dayFrom, dayTo))
	}
	const day = 24 * time.Hour
	now := time.Now()
	now = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	return &FilteredTasksRenderer{
		StatusType: map[obsast.PlugTasksStatusType]bool{
			obsast.PlugTasksStatusTypeTODO:       true,
			obsast.PlugTasksStatusTypeInProgress: true,
		},
		DueAfter:              now.Add(time.Duration(dayFrom) * day),
		DueBefore:             now.Add(time.Duration(dayTo) * day),
		ScheduledAfter:        now.Add(time.Duration(dayFrom) * day),
		ScheduledBefore:       now.Add(time.Duration(dayTo) * day),
		StartBefore:           now,
		RequireDueOrScheduled: true,
	}
}

// RegisterFuncs add AST objects to Renderer.
//
//nolint:funlen // This function naturally needs many statements to register all AST node types
func (r *FilteredTasksRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	// Register render functions for all known kinds to make sure no other functions
	// (registered with lower priority) will be called and output some HTML.
	reg.Register(ast.KindAutoLink, nil)
	reg.Register(ast.KindBlockquote, nil)
	reg.Register(ast.KindCodeBlock, nil)
	reg.Register(ast.KindCodeSpan, nil)
	reg.Register(ast.KindDocument, nil)
	reg.Register(ast.KindEmphasis, nil)
	reg.Register(ast.KindFencedCodeBlock, nil)
	reg.Register(ast.KindHTMLBlock, nil)
	reg.Register(ast.KindHeading, nil)
	reg.Register(ast.KindImage, nil)
	reg.Register(ast.KindLink, nil)
	reg.Register(ast.KindList, nil)
	reg.Register(ast.KindListItem, r.listItem)
	reg.Register(ast.KindParagraph, nil)
	reg.Register(ast.KindRawHTML, nil)
	reg.Register(ast.KindString, nil)
	reg.Register(ast.KindText, nil)
	reg.Register(ast.KindTextBlock, nil)
	reg.Register(ast.KindThematicBreak, nil)

	reg.Register(extast.KindDefinitionDescription, nil)
	reg.Register(extast.KindDefinitionList, nil)
	reg.Register(extast.KindDefinitionTerm, nil)
	reg.Register(extast.KindFootnote, nil)
	reg.Register(extast.KindFootnoteBacklink, nil)
	reg.Register(extast.KindFootnoteLink, nil)
	reg.Register(extast.KindFootnoteList, nil)
	reg.Register(extast.KindStrikethrough, nil)
	reg.Register(extast.KindTable, nil)
	reg.Register(extast.KindTableCell, nil)
	reg.Register(extast.KindTableHeader, nil)
	reg.Register(extast.KindTableRow, nil)
	reg.Register(extast.KindTaskCheckBox, nil)

	reg.Register(hashtag.Kind, nil)

	reg.Register(wikilink.Kind, nil)

	reg.Register(mermaid.Kind, nil)
	reg.Register(mermaid.ScriptKind, nil)

	reg.Register(mathjax.KindMathBlock, nil)
	reg.Register(mathjax.KindInlineMath, nil)

	reg.Register(obsast.KindBlockID, nil)

	reg.Register(obsast.KindPlugTasksStatus, nil)
	reg.Register(obsast.KindPlugTasksPrio, nil)
	reg.Register(obsast.KindPlugTasksID, nil)
	reg.Register(obsast.KindPlugTasksDependsOn, nil)
	reg.Register(obsast.KindPlugTasksDue, nil)
	reg.Register(obsast.KindPlugTasksScheduled, nil)
	reg.Register(obsast.KindPlugTasksStart, nil)
	reg.Register(obsast.KindPlugTasksCreated, nil)
	reg.Register(obsast.KindPlugTasksDone, nil)
	reg.Register(obsast.KindPlugTasksCancelled, nil)
	reg.Register(obsast.KindPlugTasksRecurring, nil)
	reg.Register(obsast.KindPlugTasksOnCompletion, nil)
}

// listItem renders a list item node, but only if it matches the configured filters.
// The entering parameter indicates whether we're entering (true) or leaving (false) the node.
//
//nolint:revive // entering parameter is required by goldmark AST walker interface
func (r *FilteredTasksRenderer) listItem(w util.BufWriter, source []byte, node ast.Node, entering bool) (
	ast.WalkStatus, error,
) {
	n := node.(*ast.ListItem)
	if !entering {
		return ast.WalkContinue, nil
	}
	if n.FirstChild() == nil {
		return ast.WalkContinue, nil
	}

	var task struct {
		StatusType obsast.PlugTasksStatusType
		Due        time.Time
		Scheduled  time.Time
		Start      time.Time
	}
	err := ast.Walk(n.FirstChild(), func(n ast.Node, _ bool) (ast.WalkStatus, error) {
		switch n := n.(type) {
		case *ast.List:
			return ast.WalkSkipChildren, nil
		case *obsast.PlugTasksStatus:
			task.StatusType = n.StatusType
		case *obsast.PlugTasksDue:
			task.Due = n.Date
		case *obsast.PlugTasksScheduled:
			task.Scheduled = n.Date
		case *obsast.PlugTasksStart:
			task.Start = n.Date
		}
		return ast.WalkContinue, nil
	})
	if err != nil {
		return 0, err
	}

	var (
		hasDue          = !task.Due.IsZero()
		hasScheduled    = !task.Scheduled.IsZero()
		dueInTime       = isBetween(task.Due, r.DueBefore, r.DueAfter)
		scheduledInTime = isBetween(task.Scheduled, r.ScheduledBefore, r.ScheduledAfter)
		started         = isBetween(task.Start, r.StartBefore, time.Time{})
	)
	if r.StatusType[task.StatusType] &&
		started &&
		((hasDue && hasScheduled && (dueInTime || scheduledInTime)) ||
			(dueInTime && scheduledInTime)) &&
		(!r.RequireDueOrScheduled || hasDue || hasScheduled) {
		seg := n.FirstChild().Lines().At(0)
		_ = w.WriteByte('-')
		_ = w.WriteByte(' ')
		_, _ = w.Write(seg.Value(source))
		_ = w.WriteByte('\n')
	}

	return ast.WalkContinue, nil
}

// All checks are optional, interval is inclusive at both sides.
func isBetween(t, before, after time.Time) bool {
	return t.IsZero() || ((before.IsZero() || t.Compare(before) <= 0) &&
		(after.IsZero() || t.Compare(after) >= 0))
}
