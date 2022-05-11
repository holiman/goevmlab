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

package traces

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const (
	headingCol = tcell.ColorYellow
)

// asciish returns '.' for non-printable characters
func asciish(b byte) byte {
	if b >= 20 && b < 0x7f {
		return b
	}
	return '.'
}

// ShowHex displays a classic hexdump using the table.
// if prevdata is non-nil, all diffs between data and prevdata are highlighted
func ShowHex(data, prevData []byte, table *tview.Table) {
	var (
		ascii    []byte
		showDiff = prevData != nil
	)
	table.Clear()
	setCell := func(row, col int, text string, align int) {
		cell := tview.NewTableCell(text).
			SetTextColor(headingCol).
			SetAlign(align)
		table.SetCell(row, col, cell)
	}

	// Headings
	table.SetFixed(1, 0).SetSelectable(false, false)
	for i := 0; i <= 0xf; i++ {
		setCell(0, i+1, fmt.Sprintf("%02x", i), tview.AlignRight)
	}
	setCell(0, 0x11, "[ -- ascii -- ]", tview.AlignCenter)
	setCell(1, 0, fmt.Sprintf("%08x", 0), tview.AlignLeft)

	for i, b := range data {
		var (
			row       = (i / 0x10) + 1
			col       = i%0x10 + 1
			cellStyle = tcell.StyleDefault
		)

		if showDiff {
			if i >= len(prevData) || prevData[i] != b {
				cellStyle = cellStyle.Reverse(true)
			}
		}
		// Print out position
		if col == 1 {
			setCell(row, 0, fmt.Sprintf("%08x", i), tview.AlignLeft)
			ascii = []byte{}
		}
		table.SetCell(row, col, tview.NewTableCell(fmt.Sprintf("%02x", b)).SetStyle(cellStyle))
		ascii = append(ascii, asciish(b))
		setCell(row, 0x11, string(ascii), tview.AlignLeft)
		if i > 1024*1024 {
			setCell(row, 0, "... too large to display", tview.AlignRight)
			break
		}
	}
	table.ScrollToBeginning()

}
