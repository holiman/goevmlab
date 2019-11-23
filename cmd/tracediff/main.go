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

package main

import (
	"flag"
	"fmt"
	"github.com/holiman/goevmlab/traces"
	"github.com/holiman/goevmlab/ui"
	"os"
	"strconv"
)

func init() {
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage:", os.Args[0], "filename1 filename2")
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr, `
Reads the given trace-files, and displays the traces side by side in a nice CLI user interface`)
	}
}

func main() {

	testTraces := []string{
		"./testdata/geth.00005448.trace",
		"./testdata/parity.00005448.trace",
	}

	flag.Parse()
	if flag.NArg() != 2 {
		fmt.Printf("Expected two arguments\n")
		flag.Usage()
		os.Exit(1)
	}
	f1, f2 := flag.Arg(0), flag.Arg(1)
	// Some debugging help here
	if n, err := strconv.Atoi(f1); err == nil {
		if n < len(testTraces) {
			f1 = testTraces[n]
		}
	}
	if n, err := strconv.Atoi(f2); err == nil {
		if n < len(testTraces) {
			f2 = testTraces[n]
		}
	}
	trace1, err := traces.ReadFile(f1)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	trace2, err := traces.ReadFile(f2)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	ui.NewDiffviewManager([]*traces.Traces{
		trace1, trace2,
	})
}
