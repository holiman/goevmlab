package traces

import (
	"fmt"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

const (
	headingCol = tcell.ColorYellow
)

func ShowHex(data []byte, table *tview.Table) {
	table.Clear()
	// Headings
	table.SetFixed(1, 0).SetSelectable(false, false)
	for i := 0; i <= 0xf; i++ {
		table.SetCell(0, i+1,
			tview.NewTableCell(fmt.Sprintf("%02x", i)).
				SetTextColor(headingCol).
				SetAlign(tview.AlignRight))
	}
	table.SetCell(0,0x11, tview.NewTableCell("[ -- ascii -- ]").SetTextColor(headingCol).
		SetAlign(tview.AlignCenter))
	var ascii []byte
	for i, b := range data {
		row := (i / 0x10) + 1
		col := i%0x10 + 1
		// Print out position
		if i%0x10 == 0 {
			table.SetCell(row, 0,
				tview.NewTableCell(fmt.Sprintf("%08x", i)).
					SetTextColor(headingCol).SetAlign(tview.AlignLeft))
			ascii = []byte{}
		}
		table.SetCell(row, col, tview.NewTableCell(fmt.Sprintf("%02x", b)))

		//txt = lambda c: chr(c) if 0x20 <= c < 0x7F else "."
		if b >= 20 && b < 0x7f{
			ascii = append(ascii, b)
		}else{
			ascii = append(ascii, '.')
		}
		table.SetCell(row, 0x11, tview.NewTableCell(string(ascii)).
			SetTextColor(headingCol).SetAlign(tview.AlignLeft) )
	}

	table.ScrollToBeginning()

}
