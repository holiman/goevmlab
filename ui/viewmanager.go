// Copyright 2019 Martin Holst Swende
// This file is part of the goevmlab library.
//
// The library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the goevmlab library. If not, see <http://www.gnu.org/licenses/>.

// Package ui contains some tools to provide a nice terminal-based user interface
package ui

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/holiman/goevmlab/evms"
	"github.com/holiman/goevmlab/ops"
	"github.com/holiman/goevmlab/traces"
	"github.com/rivo/tview"
)

const (
	headingCol = tcell.ColorYellow
)

type Config struct {
	HasChunking bool
}

type viewManager struct {
	trace *traces.Traces

	traceView *tview.Table
	stackView *tview.Table
	memView   *tview.Table
	opView    *tview.Form
	root      *tview.Grid

	config *Config
}

func NewDiffviewManager(traces []*traces.Traces) {
	root := tview.NewGrid().
		SetRows(10, 0, 10, 10).
		//SetColumns(10,0,0,0).
		SetBorders(false)

	var managers []*viewManager
	for i, trace := range traces {
		opView := tview.NewForm()
		opView.SetTitle("Op").SetBorder(true)
		ops := tview.NewTable()
		ops.SetTitle("Operations").SetBorder(true)
		stack := tview.NewTable()
		stack.SetTitle("Stack").SetBorder(true)
		mem := tview.NewTable()
		mem.SetTitle("Memory").SetBorder(true)
		root.
			AddItem(opView, 0, i, 1, 1, 0, 50, false).
			AddItem(ops, 1, i, 1, 1, 0, 50, false).
			AddItem(stack, 2, i, 1, 1, 0, 50, false).
			AddItem(mem, 3, i, 1, 1, 0, 50, false)
		mgr := viewManager{
			trace:     trace,
			traceView: ops,
			opView:    opView,
			stackView: stack,
			memView:   mem,
			root:      root,
		}
		mgr.init(trace)
		managers = append(managers, &mgr)
	}
	// Multiplex input to all managers
	root.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		for _, manager := range managers {
			h := manager.traceView.InputHandler()
			h(event, nil)
		}
		return nil
	})
	// Display diffs
	ref := managers[0].trace
	for _, manager := range managers[1:] {
		ops := manager.trace.Ops
		for i, step := range ref.Ops {
			if i >= len(ops) {
				break
			}
			other := ops[i]
			if !step.Equals(other) {
				manager.traceView.GetCell(i+1, 0).SetBackgroundColor(tcell.ColorRed)
			}
		}
	}

	if err := tview.NewApplication().SetRoot(root, true).Run(); err != nil {
		panic(err)
	}
}

// NewViewManager create a viewmanager for the single-trace view
func NewViewManager(trace *traces.Traces, cfg *Config) {
	app := tview.NewApplication()
	root := tview.NewFlex().SetDirection(tview.FlexRow)

	ops := tview.NewTable()
	ops.SetTitle("Operations").SetBorder(true)
	opView := tview.NewForm()
	opView.SetTitle("Op").SetBorder(true)
	stack := tview.NewTable()
	stack.SetTitle("Stack").SetBorder(true)
	mem := tview.NewTable()
	mem.SetTitle("Memory").SetBorder(true)
	searchField := tview.NewInputField().SetPlaceholder("Press '/' for opcode search, and 'n' for next. Press 'm' to toggle mem/stack layout. ")
	var doSearch = func() {
		query := searchField.GetText()
		cur, _ := ops.GetSelection()
		query = strings.TrimPrefix(query, "/")
		tl, idx := trace.Search(query, cur)
		if tl != nil {
			ops.Select(idx+1, 0)
		}
	}
	searchField.SetLabel("Search>  ").SetDoneFunc(func(key tcell.Key) {
		doSearch()
		app.SetFocus(ops)
	})

	mgr := viewManager{
		trace:     trace,
		traceView: ops,
		opView:    opView,
		stackView: stack,
		memView:   mem,
		config:    cfg,
	}

	mgr.init(trace)

	// Create flex row for upper view.
	upper := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(ops, 0, 3, true).
		AddItem(opView, 0, 2, false)

	// Create flex row for lower view.
	direction := tview.FlexRow
	lower := tview.NewFlex().SetDirection(direction).
		AddItem(stack, 0, 1, false).
		AddItem(mem, 0, 1, false)

	bottom := tview.NewFlex().SetDirection(tview.FlexRow).AddItem(searchField, 0, 1, false)
	root = root.AddItem(upper, 0, 1, true).AddItem(lower, 0, 1, false).AddItem(bottom, 1, 1, false)

	capture := func(event *tcell.EventKey) *tcell.EventKey {
		// Toggle flow direction when 'm' key is pressed.
		switch event.Rune() {
		case rune('m'):
			direction = (direction + 1) % 2
			lower.SetDirection(direction)
		case rune('/'):
			searchField.SetText("")
			app.SetFocus(searchField)
		case rune('n'):
			doSearch()
		}
		return event
	}
	app.SetRoot(root, true).SetInputCapture(capture)
	if err := app.Run(); err != nil {
		panic(err)
	}
}

func setHeadings(headings []string, table *tview.Table) {
	table.SetFixed(1, 0).SetSelectable(false, false)
	for col, title := range headings {
		table.SetCell(0, col,
			tview.NewTableCell(strings.ToUpper(title)).
				SetTextColor(headingCol).
				SetAlign(tview.AlignRight))
	}
}

func (mgr *viewManager) onStepSelected(line *traces.TraceLine) {
	mgr.opView.Clear(true)
	mgr.stackView.Clear()
	if line == nil {
		return
	}
	// Update the detailed opview
	{
		add := func(label, data string) {
			field := tview.NewInputField().
				SetLabel(label).
				SetText(data)
			mgr.opView.AddFormItem(field)
		}

		var headers []string
		if evms.IgnoreEOF {
			headers = []string{"pc", "section", "opcode", "opName", "gasCost", "gas", "memSize", "addr", "functionDepth"}
		} else {
			headers = []string{"pc", "opcode", "opName", "gasCost", "gas", "memSize", "addr"}
		}
		for _, l := range headers {
			add(l, line.Get(l))
		}
		// Add the call stack info
		cs := line.CallStack()
		for i := len(cs) - 1; i >= 0; i-- {
			info := cs[i]
			add(fmt.Sprintf("call %d ", i), info.String())

		}
		op := ops.OpCode(line.Op())
		add("Pops", strings.Join(op.Pops(), ","))
		add("Pushes", strings.Join(op.Pushes(), ","))
	}
	{ // Update the stack view
		setHeadings([]string{"pos", "                                                            data", "desc      "}, mgr.stackView)

		op := ops.OpCode(line.Op())
		popDescriptors := op.Pops()

		for i, item := range line.Stack() {
			mgr.stackView.SetCell(i+1, 0, tview.NewTableCell(fmt.Sprintf("%02d", i)))
			mgr.stackView.SetCell(i+1, 1, tview.NewTableCell(fmt.Sprintf("%64s", item.Hex())))
			if i < len(popDescriptors) {
				mgr.stackView.SetCell(i+1, 2, tview.NewTableCell(popDescriptors[i]))
			}
		}
		mgr.stackView.ScrollToBeginning()
	}
	{ // Update the mem view
		var prevMem []byte
		if prevOp := mgr.trace.Get(int(line.Step()) - 1); prevOp != nil {
			prevMem = prevOp.Memory()
		}
		traces.ShowHex(line.Memory(), prevMem, mgr.memView)
	}
}

func (mgr *viewManager) init(trace *traces.Traces) {

	{ // The detailed opview
		mgr.opView.SetFieldBackgroundColor(tcell.ColorGray)
		mgr.opView.SetItemPadding(0)

	}

	{ // The operations table
		table := mgr.traceView
		headings := []string{"step", "pc", "section", "opName", "opCode",
			"gas", "gasCost", "depth", "functionDepth", "refund"}

		if mgr.config != nil && mgr.config.HasChunking {
			headings = append(headings, "chunk")
		}

		table.SetSelectable(true, false).
			SetSelectedFunc(func(row int, column int) {
				table.GetCell(row, column).SetTextColor(tcell.ColorRed)
			}).
			SetSelectionChangedFunc(func(row, col int) {
				// don't update for headings
				if row < 1 {
					return
				}
				mgr.onStepSelected(trace.Get(row - 1))
			}).
			Select(1, 1).SetFixed(1, 1)

		// Headings
		for col, title := range headings {
			table.SetCell(0, col,
				tview.NewTableCell(strings.ToUpper(title)).
					SetTextColor(headingCol).
					SetAlign(tview.AlignCenter).
					SetSelectable(false))
		}
		// Ops table body
		for i, elem := range trace.Ops {
			if elem == nil {
				break
			}
			row := i + 1
			for col, title := range headings {
				data := elem.Get(title)
				table.SetCell(row, col, tview.NewTableCell(data))
			}
		}
	}
	{ // Stack
		setHeadings([]string{"pos", "                            data", "desc"}, mgr.stackView)
	}

}
