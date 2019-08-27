package main

import (
	"fmt"
	"github.com/gdamore/tcell"
	"github.com/holiman/goevmlab/ops"
	"github.com/holiman/goevmlab/traces"
	"github.com/rivo/tview"
	"os"
	"strings"
)

const (
	headingCol = tcell.ColorYellow
)

type viewManager struct {
	traceView *tview.Table
	stackView *tview.Table
	memView   *tview.Table
	opView    *tview.Form
}

func NewViewManager(trace *traces.Traces) *viewManager {

	newPrimitive := func(text string) tview.Primitive {
		return tview.NewTextView().
			SetTextAlign(tview.AlignCenter).
			SetText(text)
	}

	ops := tview.NewTable()
	ops.SetTitle("Operations").SetBorder(true)

	opView := tview.NewForm()
	opView.SetTitle("Op").SetBorder(true)

	stack := tview.NewTable()
	stack.SetTitle("Stack").SetBorder(true)

	mem := tview.NewTable()
	mem.SetTitle("Memory").SetBorder(true)

	mgr := viewManager{
		traceView: ops,
		opView:    opView,
		stackView: stack,
		memView:   mem,
	}

	mgr.init(trace)

	//focus := 0
	//focusOrder := [ops, opView, stack, mem]

	grid := tview.NewGrid().
		SetRows(3, 0, 15, 3).
		SetColumns(0, 80).
		SetBorders(true).
		AddItem(newPrimitive("Header"), 0, 0, 1, 2, 0, 0, false).
		AddItem(newPrimitive("Footer"), 3, 0, 1, 2, 0, 0, false)

	// Layout for screens narrower than 100 cells (menu and side bar are hidden).
	//grid.AddItem(menu, 0, 0, 0, 0, 0, 0, false).
	//	AddItem(main, 1, 0, 1, 3, 0, 0, false).
	//	AddItem(sideBar, 0, 0, 0, 0, 0, 0, false)

	// Layout for screens wider than 100 cells.
	grid.
		AddItem(opView, 1, 1, 1, 1, 0, 50, false).
		AddItem(stack, 2, 0, 1, 1, 0, 50, false).
		AddItem(mem, 2, 1, 1, 1, 0, 50, false).
		AddItem(ops, 1, 0, 1, 1, 0, 50, true)

	//grid.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
	//	if event.Key() == tcell.KeyRight{
	//		grid.
	//		focus ++
	//		focus %= len(focusOrder)
	//		ops.GetFocusable()
	//	}
	//})

	if err := tview.NewApplication().SetRoot(grid, true).Run(); err != nil {
		panic(err)
	}
	return &mgr
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

		for _, l := range []string{"pc", "opcode", "opName", "gasCost", "gas", "memSize", "addr"} {
			add(l, line.Get(l))
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
			mgr.stackView.SetCell(i+1, 1, tview.NewTableCell(fmt.Sprintf("%64x", item)))
			if i < len(popDescriptors) {
				mgr.stackView.SetCell(i+1, 2, tview.NewTableCell(popDescriptors[i]))
			}
		}
	}
	{ // Update the mem view
		traces.ShowHex(line.Memory(), mgr.memView)
	}
}

func (mgr *viewManager) init(trace *traces.Traces) {

	{ // The detailed opview
		mgr.opView.SetFieldBackgroundColor(tcell.ColorGray)
		mgr.opView.SetItemPadding(0)

	}

	{ // The operations table
		table := mgr.traceView
		headings := []string{"step", "pc", "opName", "opCode",
			"gas", "gasCost", "depth", "refund"}

		table.SetSelectable(true, false).
			SetSelectedFunc(func(row int, column int) {
				table.GetCell(row, column).SetTextColor(tcell.ColorRed)
			}).
			SetSelectionChangedFunc(func(row, col int) {
				mgr.onStepSelected(trace.Get(row - 1))
			}).
			Select(1, 1).SetFixed(1, 1)

		// Headings
		for col, title := range headings {
			table.SetCell(0, col,
				tview.NewTableCell(strings.ToUpper(title)).
					SetTextColor(headingCol).
					SetAlign(tview.AlignCenter))
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
			if i == 100 {
				break
			}
		}
	}
	{ // Stack
		setHeadings([]string{"pos", "                            data", "desc"}, mgr.stackView)
	}

}

func main() {

	//trace, err := traces.ReadTrace("/home/user/go/src/github.com/holiman/goevmlab/testdata/traces/geth_nomemory.jsonl")
	//trace, err := traces.ReadTrace("/home/user/go/src/github.com/holiman/goevmlab/testdata/traces/geth_memory.jsonl")
	trace, err := traces.ReadTrace("/home/user/go/src/github.com/holiman/goevmlab/testdata/traces/geth_traceTransaction.json")
	if err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}

	NewViewManager(trace)

}
